package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/ekilmer/gh-attach/internal/app"
	"github.com/ekilmer/gh-attach/internal/cookies"
	"github.com/spf13/cobra"
)

type SessionTransferOptions struct {
	SSH             string
	Browser         string
	Profile         string
	CookieStorePath string
	Verbose         bool
}

func NewCmdSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage a browser session for headless uploads",
	}
	cmd.AddCommand(newCmdSessionTransfer(nil))
	return cmd
}

func newCmdSessionTransfer(runF func(*SessionTransferOptions) error) *cobra.Command {
	opts := &SessionTransferOptions{}
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Send your local GitHub browser session to an SSH host",
		Long:  "Read the GitHub login cookie from a local browser and send it over SSH to the remote gh-attach token file.",
		Example: `  $ gh attach session transfer --ssh user@server
  $ gh attach session transfer --ssh server --browser firefox --profile default-release`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateSSHTarget(opts.SSH); err != nil {
				return err
			}
			if runF != nil {
				return runF(opts)
			}
			return sessionTransferRun(cmd.Context(), opts, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&opts.SSH, "ssh", "", "SSH host or user@host to receive the session")
	cmd.Flags().StringVar(&opts.Browser, "browser", "", "Local browser to read cookies from ("+cookies.BrowserChoices()+")")
	cmd.Flags().StringVar(&opts.Profile, "profile", "", "Local browser profile name")
	cmd.Flags().StringVar(&opts.CookieStorePath, "cookie-store-path", "", "Local browser cookie store file path")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose cookie source resolution")
	_ = cmd.MarkFlagRequired("ssh")
	return cmd
}

func validateSSHTarget(target string) error {
	if target == "" || strings.HasPrefix(target, "-") || strings.ContainsAny(target, " \t\r\n") {
		return errors.New("--ssh must be a host or user@host")
	}
	return nil
}

func sessionTransferRun(ctx context.Context, opts *SessionTransferOptions, stdout, stderr io.Writer) error {
	token, expires, err := app.NewService(stderr).BrowserSessionToken(ctx, "github.com", cookies.ResolveInput{
		Browser:         opts.Browser,
		Profile:         opts.Profile,
		CookieStorePath: opts.CookieStorePath,
	}, opts.Verbose)
	if err != nil {
		return err
	}
	remotePath, err := transferSession(ctx, opts.SSH, token, stderr)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "GitHub session stored at %s:%s\n", opts.SSH, remotePath)
	if expires.IsZero() {
		fmt.Fprintln(stdout, "Cookie expiration: unavailable from this browser")
	} else {
		fmt.Fprintf(stdout, "Cookie expiration: %s\n", expires.UTC().Format(time.RFC3339))
	}
	return nil
}

const remoteSessionInstall = `set -eu
dir="${XDG_CONFIG_HOME:-$HOME/.config}/gh"
umask 077
mkdir -p "$dir"
tmp="$(mktemp "$dir/.attach-session.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
cat > "$tmp"
chmod 600 "$tmp"
mv -f "$tmp" "$dir/attach-session"
printf 'GH_ATTACH_SESSION_PATH=%s\n' "$dir/attach-session"`

func transferSession(ctx context.Context, target, token string, stderr io.Writer) (string, error) {
	if err := validateSSHTarget(target); err != nil {
		return "", err
	}
	command := exec.CommandContext(ctx, "ssh", "-T", target, remoteSessionInstall)
	command.Stdin = strings.NewReader(token)
	command.Stderr = stderr
	var output bytes.Buffer
	command.Stdout = &output
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("SSH session transfer to %s: %w", target, err)
	}
	for _, line := range strings.Split(output.String(), "\n") {
		if path, ok := strings.CutPrefix(line, "GH_ATTACH_SESSION_PATH="); ok && path != "" {
			return path, nil
		}
	}
	return "", errors.New("SSH transfer completed without reporting the remote file path")
}
