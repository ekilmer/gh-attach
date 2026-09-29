package cmd

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSessionTokenFileAuthSource(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(sessionTokenEnv, "placeholder")
	if err := os.Unsetenv(sessionTokenEnv); err != nil {
		t.Fatal(err)
	}
	configDir := filepath.Join(dir, "gh")
	if err := os.Mkdir(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	tokenPath := filepath.Join(configDir, "session")
	if err := os.WriteFile(tokenPath, []byte("file-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "attach.yml"), []byte("session_token_file: "+tokenPath+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
		new  func(func(string)) *cobra.Command
		want string
	}{
		{
			name: "upload",
			args: []string{"example.zip"},
			new: func(record func(string)) *cobra.Command {
				return NewCmdAttach(func(opts *AttachOptions) error { record(opts.SessionToken); return nil })
			},
			want: "file-token",
		},
		{
			name: "download keeps bearer authentication",
			args: []string{"https://github.com/user-attachments/files/1/example.zip", "-O", "example.zip"},
			new: func(record func(string)) *cobra.Command {
				return NewCmdDownload(func(opts *DownloadOptions) error { record(opts.SessionToken); return nil })
			},
		},
		{
			name: "explicit browser",
			args: []string{"example.zip", "--browser", "firefox"},
			new: func(record func(string)) *cobra.Command {
				return NewCmdAttach(func(opts *AttachOptions) error { record(opts.SessionToken); return nil })
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got string
			cmd := test.new(func(token string) { got = token })
			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("SessionToken = %q, want %q", got, test.want)
			}
		})
	}

	t.Setenv(sessionTokenEnv, "environment-token")
	var got string
	cmd := NewCmdAttach(func(opts *AttachOptions) error { got = opts.SessionToken; return nil })
	cmd.SetArgs([]string{"example.zip"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got != "environment-token" {
		t.Fatalf("SessionToken = %q, want environment-token", got)
	}
}

func TestDefaultSessionTokenFileAuthSource(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv(sessionTokenEnv, "placeholder")
	if err := os.Unsetenv(sessionTokenEnv); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "gh", "attach-session")
	if err := os.Mkdir(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("default-file-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got string
	cmd := NewCmdAttach(func(opts *AttachOptions) error { got = opts.SessionToken; return nil })
	cmd.SetArgs([]string{"example.zip"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got != "default-file-token" {
		t.Fatalf("SessionToken = %q, want default-file-token", got)
	}
}

func TestReadSessionTokenFileRejectsUnprotectedOrInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readSessionTokenFile("relative/session"); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("relative path error = %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := readSessionTokenFile(path); err == nil || !strings.Contains(err.Error(), "chmod 600") {
			t.Fatalf("unprotected file error = %v", err)
		}
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readSessionTokenFile(path); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("empty file error = %v", err)
	}
}

func TestNewCmdAttach_AcceptsMultipleFiles(t *testing.T) {
	var got []string
	cmd := NewCmdAttach(func(opts *AttachOptions) error {
		got = append(got, opts.FilePaths...)
		return nil
	})
	cmd.SetArgs([]string{"image.png", "report.pdf"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if want := []string{"image.png", "report.pdf"}; !slices.Equal(got, want) {
		t.Fatalf("FilePaths = %v, want %v", got, want)
	}
}

func TestNewCmdAttach_Markdown(t *testing.T) {
	var markdown bool
	cmd := NewCmdAttach(func(opts *AttachOptions) error {
		markdown = opts.Markdown
		return nil
	})
	cmd.SetArgs([]string{"image.png", "--markdown"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !markdown {
		t.Fatal("Markdown = false, want true")
	}
}

func TestNewCmdAttach_RejectsMarkdownWithJSON(t *testing.T) {
	called := false
	cmd := NewCmdAttach(func(_ *AttachOptions) error {
		called = true
		return nil
	})
	cmd.SetArgs([]string{"image.png", "--markdown", "--json", "href"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "markdown json") {
		t.Fatalf("Execute() error = %v, want mutually exclusive flags error", err)
	}
	if called {
		t.Fatal("upload ran despite incompatible output flags")
	}
}

func TestNewCmdAttach_SessionTokenPrecedence(t *testing.T) {
	t.Setenv(sessionTokenEnv, "environment-token")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "environment", args: []string{"image.png"}, want: "environment-token"},
		{name: "flag", args: []string{"image.png", "--session-token", "flag-token"}, want: "flag-token"},
		{name: "explicit browser", args: []string{"image.png", "--browser", "firefox"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got string
			cmd := NewCmdAttach(func(opts *AttachOptions) error {
				got = opts.SessionToken
				return nil
			})
			cmd.SetArgs(test.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("SessionToken = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewCmdAttach_RejectsSessionTokenWithBrowserOptions(t *testing.T) {
	for _, flag := range []string{"--browser", "--profile", "--cookie-store-path"} {
		t.Run(flag, func(t *testing.T) {
			cmd := NewCmdAttach(nil)
			cmd.SetArgs([]string{"image.png", "--session-token", "secret", flag, "value"})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), "session-token") {
				t.Fatalf("Execute() error = %v, want mutually exclusive flags error", err)
			}
		})
	}
}

func TestNewCmdAttach_AllowsBrowserWithProfile(t *testing.T) {
	cmd := NewCmdAttach(func(_ *AttachOptions) error { return nil })
	cmd.SetArgs([]string{"image.png", "--browser", "firefox", "--profile", "default"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewCmdAttach_RejectsEmptySessionToken(t *testing.T) {
	called := false
	cmd := NewCmdAttach(func(_ *AttachOptions) error {
		called = true
		return nil
	})
	cmd.SetArgs([]string{"image.png", "--session-token="})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "session token is empty") {
		t.Fatalf("Execute() error = %v, want empty token error", err)
	}
	if called {
		t.Fatal("upload ran with an empty session token")
	}
}

func TestNewCmdAttach_RejectsEmptySessionTokenEnvironment(t *testing.T) {
	t.Setenv(sessionTokenEnv, " ")
	cmd := NewCmdAttach(func(_ *AttachOptions) error {
		t.Fatal("upload ran with an empty session token")
		return nil
	})
	cmd.SetArgs([]string{"image.png"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "session token is empty") {
		t.Fatalf("Execute() error = %v, want empty token error", err)
	}
}
