---
name: gh-attach
description: Uploads a local file (screenshot, image, PDF, zip, video) to GitHub user-attachments, downloads GitHub user-attachments, and embeds local files in a PR, issue, or comment. Use when asked to "attach a screenshot to the PR", "add an image to the issue", "embed before/after screenshots", "attach this file", or "download this GitHub attachment". Powered by `gh-attach`.
license: MIT
---

# gh-attach

`gh attach` uploads a file to GitHub's internal user-attachments endpoint (no public API exists) and prints the URL, which GitHub auto-renders (image/video/file) wherever it's pasted. The URL inherits the repo's visibility, so private-repo uploads stay private.

## Prerequisites

```sh
gh extension list | grep -q 'gh attach' || gh extension install ekilmer/gh-attach
```

Uploads use a GitHub browser session cookie, not the `gh` token. By default, `gh` must be authenticated so `gh-attach` can select the matching browser account. If the wrong account is selected, add `--browser <name> --profile <name>`. For a headless SSH host, run `gh attach session transfer --ssh user@host` on the browser computer once; the remote upload command automatically reads the protected token file. `session_token_file` in `${XDG_CONFIG_HOME:-~/.config}/gh/attach.yml` overrides that file path. Treat the cookie as a full account credential and never print its value.

## When a saved session expires

A browser session can expire or be revoked. If an SSH upload fails with an authentication error, tell the user that the saved session may need refreshing. Ask them to sign in to GitHub in a browser on their own computer and run this command there, using their SSH host and any `--browser` or `--profile` options they needed for the first transfer:

```sh
gh attach session transfer --ssh user@host
```

If `session_token_file` points elsewhere, move the refreshed file to that configured path on the SSH host and keep it private (mode `600` on Unix). Retry the upload after the transfer. Never ask the user to paste the cookie into chat or print it for troubleshooting.

## Steps

**1. Upload**: Use an absolute quoted path. `-R` is optional inside a repository. For GHES, use `-R host/owner/repo`. The command prints the URL on one line. GitHub auto-renders it (image/video/file), so use it as-is:

```sh
URL=$(gh attach "$FILE" -R <owner>/<repo>)
```

**2. Embed** (always `--body-file -`, e.g. `gh pr comment/edit`, `gh issue comment/edit`):

```sh
printf '## Screenshots\n\n%s\n' "$URL" | gh pr comment <pr> -R <owner>/<repo> --body-file -
```

**3. Download**: Specify the destination explicitly. Private attachments use the active `gh` token, with browser cookies as an authorization fallback:

```sh
gh attach download "$URL" -O "$FILE"
```

## Notes

- Private repo: URL renders only for authorized viewers. An anonymous fetch is expected to return 404 or 403.
- Sizing: embed `<img width="800" src="$URL">` instead of the bare URL.
- GitHub Cloud and GHES decide which file extensions and content types they accept.
