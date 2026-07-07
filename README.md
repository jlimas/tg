# tg

A fast, zero-dependency Telegram CLI, written in Go. One static binary,
no Python/Node runtime, no external libraries at runtime — install it and
send a message in two commands.

```sh
$ tg text --to 123456789 --message "hello from tg"
sent: message 142 to 123456789
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/jlimas/tg/main/install.sh | sh
```

This downloads the latest release binary for your OS/arch into
`~/.local/bin`. Override the destination with `INSTALL_DIR=/some/path`, or
pin a version with `VERSION=v0.1.0`.

Alternatively, build from source (requires Go 1.25+):

```sh
git clone https://github.com/jlimas/tg
cd tg
just install   # or: go build -o ~/.local/bin/tg .
```

## Setup

Create a bot with [@BotFather](https://t.me/BotFather) to get a token, then:

```sh
tg config set --bot-token "123456:AAExample-Token" --default-chat-id 987654321
```

`--default-chat-id` is optional — if set, `tg text` doesn't need `--to`.
Config is stored at `~/.config/tg/config.toml`.

Bots can only message users who have started a chat with them, or
groups/channels they've been added to — send `/start` to your bot first to
get it to respond, and use [@userinfobot](https://t.me/userinfobot) or
similar to find your numeric chat id.

## Usage

```sh
tg                                          # status and next steps
tg config show                              # view current config (token masked)
tg text --to 123456789 --message "hello"
tg photo --to 123456789 --file ./cat.jpg --caption "hi"
tg document --to 123456789 --file ./report.pdf --caption "Q3"
tg video --to 123456789 --file ./clip.mp4 --supports-streaming
tg audio --to 123456789 --file ./song.mp3 --performer "X" --title "Y"
tg voice --to 123456789 --file ./note.ogg
tg text --message "hello"                   # uses default_chat_id
tg text --to 123456789 --message "*bold*" --parse-mode Markdown
tg --help                                   # full command reference
tg <command> --help                         # per-command flags and examples
```

Exit codes are meaningful for scripting: `0` on success, `1` on a runtime
error (e.g. the Telegram API rejected the request), `2` on a usage error
(missing/unknown flags).

## Design

`tg`'s output follows [AXI](https://agentskills.io) conventions — compact,
structured, predictable exit codes, no interactive prompts, no unknown
flags silently ignored. That makes it just as easy to drive from a shell
script or CI job as it is to read at a terminal, and it's why the CLI
composes well as a tool call for LLM agents scripting Telegram from the
shell.

## Development

```sh
just --list    # see all available recipes
just check     # gofmt, vet, build
just run text --to 123456789 --message "test"
```

Releases are cut by pushing a `vX.Y.Z` tag; GitHub Actions builds
cross-platform binaries with [goreleaser](https://goreleaser.com) and
publishes them to GitHub Releases.

## Roadmap

MVP covers sending text messages. Natural next steps: photo/file uploads,
reading incoming messages (`getUpdates`), and multi-bot/account config
profiles.

## License

[MIT](LICENSE)
