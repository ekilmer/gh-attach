package cmd

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSessionTransferCommandRequiresSSHTarget(t *testing.T) {
	cmd := NewCmdSession()
	cmd.SetArgs([]string{"transfer"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); err == nil {
		t.Fatal("transfer without --ssh unexpectedly succeeded")
	}
	for _, target := range []string{"-bad", "host name", "host\ncommand"} {
		if err := validateSSHTarget(target); err == nil {
			t.Fatalf("accepted invalid SSH target %q", target)
		}
	}
}

func TestTransferSessionInstallsProtectedFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SSH shell fixture requires Unix")
	}
	dir := t.TempDir()
	sshPath := filepath.Join(dir, "ssh")
	fakeSSH := `#!/bin/sh
test "$1" = "-T" || exit 2
test "$2" = "test-host" || exit 3
exec sh -c "$3"
`
	if err := os.WriteFile(sshPath, []byte(fakeSSH), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "remote-config"))

	path, err := transferSession(t.Context(), "test-host", "secret-cookie", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "remote-config", "gh", "attach-session") {
		t.Fatalf("remote path = %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "secret-cookie" {
		t.Fatalf("stored unexpected cookie content")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("token file mode = %o, want 600", info.Mode().Perm())
	}
	if strings.Contains(path, "secret-cookie") {
		t.Fatal("token leaked into command output")
	}
}
