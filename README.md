<div align="center">

# gh-attach

[![version](https://badgen.net/github/release/ekilmer/gh-attach?label=version)](https://github.com/ekilmer/gh-attach/releases)
[![license](https://badgen.net/github/license/ekilmer/gh-attach?color=green)](./LICENSE)
[![downloads](https://img.shields.io/github/downloads/ekilmer/gh-attach/total?color=green)](https://github.com/ekilmer/gh-attach/releases)
[![skills.sh](https://skills.sh/b/ekilmer/gh-attach)](https://skills.sh/ekilmer/gh-attach)

A GitHub CLI extension that uploads and downloads GitHub attachments.

<a href="docs/assets/gh-attach-demo.webp">
  <img src="docs/assets/gh-attach-demo.webp" alt="gh-attach demo" width="800" />
</a>

</div>

This fork started from [sudosubin/gh-attach](https://github.com/sudosubin/gh-attach) and includes additional modifications. The original project uses the MIT License, which requires copies or substantial portions to include its copyright and permission notice. This fork retains that notice in [LICENSE](./LICENSE).

## Quick Start

```sh
gh extension install ekilmer/gh-attach
gh attach ./image.png -R owner/repo
```

## Installation

Requires [GitHub CLI](https://github.com/cli/cli#installation). Browser-cookie mode also requires `gh auth login`. Session-token mode does not when the repository ID is available from GitHub's page.

```sh
gh extension install ekilmer/gh-attach
gh attach ./image.png -R owner/repo
```

### Install from a local checkout

From the repository root, build the executable before installing the local extension:

```sh
go build -o gh-attach ./cmd/gh-attach
gh extension install .
gh attach --help
```

For local installs, `gh extension install .` links to this checkout; it does not build the Go executable. After source changes, rerun the `go build` command. If the extension is already installed, building the executable is enough; you do not need to reinstall it.

## Usage

```sh
$ gh attach ./image.png -R owner/repo
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000

$ gh attach upload ./image.png -R owner/repo
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000

$ gh attach ./image.png ./report.pdf -R owner/repo
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000
https://github.com/user-attachments/files/123/report.pdf

$ gh attach ./image.png ./report.pdf --markdown
![image.png](https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000)
[report.pdf](https://github.com/user-attachments/files/123/report.pdf)

$ gh attach ./image.png -R owner/repo --browser chrome --profile Default
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000

$ GH_ATTACH_SESSION_TOKEN="$USER_SESSION" gh attach ./image.png -R owner/repo
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000

$ gh attach ./image.png --json id,href,name
[
  {
    "href": "https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000",
    "id": 123,
    "name": "image.png"
  }
]

$ gh attach ./image.png --json href --jq '.[].href'
https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000

$ gh attach ./image.png --json href,name --template '{{range .}}{{.name}} -> {{.href}}{{"\n"}}{{end}}'
image.png -> https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000
```

Up to two files are uploaded concurrently, and per-file failures do not stop the remaining uploads. With `--json`, results are always returned as an array.

### Download

```sh
$ gh attach download https://github.com/user-attachments/assets/550e8400-e29b-41d4-a716-446655440000 -O image.png

$ gh attach download https://github.com/user-attachments/files/123/report.pdf -O -
```

Downloads use an explicit session token or browser selection first. Otherwise they use the active `gh` authentication token and retry with the matching browser cookies after an authorization failure.

## Use with AI agents

`gh-attach` ships as an [agent skill](https://agentskills.io), so AI coding agents can attach screenshots or files to a PR or issue, embed them as Markdown, and download `user-attachments` URLs from a natural-language request like *"attach this screenshot to the PR"*.

```sh
npx skills add ekilmer/gh-attach
```

The [Agent Skills standard](https://agentskills.io/clients) is supported by Claude Code, OpenAI Codex, Cursor, GitHub Copilot, and more.

## Options

### Upload options

- `-R, --repo <[HOST/]OWNER/REPO>`: Target repository. Auto detection is available from current repository.
- `--browser <name>`: Browser to read cookies from (`auto|arc|atlas|brave|chrome|chromium|comet|dia|edge|firefox|floorp|helium|librewolf|opera|safari|vivaldi|waterfox|whale|zen`).
- `--profile <name>`: Browser profile name. For Firefox-family multi-account containers, append `:<container-name>` or `:id=<container-id>` to pin a specific container (e.g. `default:Work`, `default:id=2`).
- `--cookie-store-path <path>`: Explicit cookie DB file path.
- `--session-token <value>`: Bare `user_session` cookie value. Prefer the `GH_ATTACH_SESSION_TOKEN` environment variable to keep this account credential out of command history and process arguments. Explicit browser options override the environment variable.
- `--markdown`: Output Markdown references.
- `--json <fields>`: Output JSON with selected fields.
- `-q, --jq <expression>`: Apply jq filter to JSON output (requires `--json`).
- `-t, --template <go-template>`: Format JSON output using Go template (requires `--json`).
- `-v, --verbose`: Print cookie source resolution logs to stderr.
- `-h, --help`: Show help.

### Download options

- `-O, --output <file>`: File to write to. Use `-` for standard output.
- `--clobber`: Overwrite an existing file.
- `--browser`, `--profile`, `--cookie-store-path`, `--session-token`: Select the same authentication sources as upload.
- `-v, --verbose`: Print cookie source resolution logs to stderr.
- `-h, --help`: Show help.

## Configuration

You can use a config file to register frequently used browser settings without having to pass them as command line arguments each time.

### SSH and headless uploads

Run this command on the computer where you are signed in to GitHub in a browser, using the SSH host name you normally connect to:

```sh
gh attach session transfer --ssh user@server
```

The command selects the browser session matching your local `gh` login, sends its `user_session` cookie over SSH standard input, and reports the cookie's expiration when the browser provides it. Use `--browser` and `--profile` if you need a particular browser profile. The cookie value is never printed or passed as a command argument. The remote copy goes to `${XDG_CONFIG_HOME:-~/.config}/gh/attach-session` with file mode `600` on Unix.

Once transferred, agents on the SSH host can upload without further authentication flags:

```sh
gh attach ./example.zip -R owner/repo
```

Uploads use the default token file automatically when no `--session-token`, `GH_ATTACH_SESSION_TOKEN`, or explicit browser option is supplied. To use a different path, move the transferred file there and set `session_token_file: /absolute/path` in `attach.yml`. A flag or environment token takes precedence. Downloads continue to try the active `gh` token first; use `GH_ATTACH_SESSION_TOKEN` explicitly if a download needs the browser session.

The cookie logs in as your GitHub account. [GitHub lists its lifetime as two weeks](https://docs.github.com/en/site-policy/privacy-policies/github-cookies#cookies), so refresh the file when the session expires. Do not place the cookie value in `attach.yml` or share it with an agent in a prompt.

The config file is loaded from `${XDG_CONFIG_HOME:-~/.config}/gh/attach.yml`.

**Example**

```yaml
browsers:
  - browser: chrome
    profile: Default
  - browser: firefox
    profile: default-release
  - browser: safari
```

**Schema**

- `session_token_file`: Absolute path to a private file containing a bare `user_session` cookie (Optional, overrides the default token file for uploads).
- `browser`: Browser to read cookies from (Required, one of `auto|arc|atlas|brave|chrome|chromium|comet|dia|edge|firefox|floorp|helium|librewolf|opera|safari|vivaldi|waterfox|whale|zen`)
- `profile`: Browser profile name/path (Optional, name or path)
- `cookie_store_path`: Explicit cookie DB file path (Optional)

## Supported Browsers

- Chromium family (Arc, Atlas, Brave, Chrome, Chromium, Comet, Dia, Edge, Helium, Opera, Vivaldi, Whale)
- Firefox family (Firefox, Floorp, LibreWolf, Waterfox, Zen)
- Safari

## How It Works

- It first resolves the target repository (`owner/repo`).
- With browser-cookie mode, it resolves the current GitHub login via the `gh` API and selects a browser session whose [`dotcom_user`](https://docs.github.com/en/site-policy/privacy-policies/github-cookies#cookies) matches it.
- With `--session-token` or `GH_ATTACH_SESSION_TOKEN`, it uses the supplied bare `user_session` value without reading a browser.
- Using that session cookie, it requests GitHub upload policies (`/upload/policies/assets`) and uploads each file binary.
- It finalizes each user-attachments asset and prints the results as URLs or formatted output via `--json`.
- Downloads prefer an explicit authentication source, then the active `gh` token, then matching browser cookies.

## Development

```sh
go test ./...
go build ./...
```

## License

MIT, see [LICENSE](./LICENSE).
