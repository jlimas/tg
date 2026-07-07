# tg

A small AXI-style CLI for sending Telegram messages through a bot.

## Install

```sh
go build -o ~/.local/bin/tg .
```

## Setup

Create a bot with [@BotFather](https://t.me/BotFather) to get a token, then:

```sh
tg config set --bot-token "123456:AAExample-Token" --default-chat-id 987654321
```

`default_chat_id` is optional — if set, `tg send` doesn't need `--to`.

Config is stored at `~/.config/tg/config.toml`.

## Usage

```sh
tg                                          # show status and next steps
tg send --to 123456789 --text "hello"
tg config show
tg --help
```

Bots can only message users who have started a chat with them, or
groups/channels they've been added to — send a `/start` to your bot first.
