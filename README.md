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

`--default-chat-id` is optional — if set, `--to` becomes optional too, on
every sending command (`text`, `photo`, `document`, `video`, `media-group`,
`paid-media`, `video-note`, `animation`, `sticker`, `audio`, `voice`,
`location`, `venue`, `contact`, `dice`, `poll`). Pass `--to` on any command
to send to a different chat without changing your config. Config is stored
at `~/.config/tg/config.toml`.

Bots can only message users who have started a chat with them, or
groups/channels they've been added to — send `/start` to your bot first to
get it to respond, and use [@userinfobot](https://t.me/userinfobot) or
similar to find your numeric chat id.

## Usage

```sh
tg                                             # status and next steps
tg config show                                # view current config (token masked)

# [--to CHAT_ID] is optional on every command below if --default-chat-id
# is configured (see Setup); pass it to target a different chat instead.
tg text [--to 123456789] --message "hello"
tg photo [--to 123456789] --file ./cat.jpg --caption "hi"
tg document [--to 123456789] --file ./report.pdf --caption "Q3"
tg video [--to 123456789] --file ./clip.mp4 --supports-streaming
tg media-group [--to 123456789] --file a.jpg --file b.jpg
tg paid-media [--to 123456789] --star-count 50 --file preview.jpg
tg video-note [--to 123456789] --file ./round.mp4
tg animation [--to 123456789] --file ./loop.gif --caption "launch"
tg sticker [--to 123456789] --file ./sticker.webp
tg audio [--to 123456789] --file ./song.mp3 --performer "X" --title "Y"
tg voice [--to 123456789] --file ./note.ogg
tg location [--to 123456789] --lat 40.7580 --long -73.9855
tg venue [--to 123456789] --lat 40.75 --long -73.98 --title "MSG" --address "4 Penn Plaza"
tg contact [--to 123456789] --phone "+15551234567" --first-name "Ada"
tg dice [--to 123456789] --emoji "🎯"
tg poll [--to 123456789] --question "Lunch?" --option Pizza --option Tacos

tg text --message "hello"                    # no --to: sends to default_chat_id
tg text --to 123456789 --message "*bold*" --parse-mode Markdown
tg --help                                    # full command reference
tg <command> --help                          # per-command flags and examples
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

## Status

All intended send-message features are implemented; the project is now in
maintenance mode.

## License

[MIT](LICENSE)
