package main

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const stickerUsage = `tg sticker --to <chat_id> --file <path|url|file_id> [--emoji <emoji>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func cmdSticker(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "emoji")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, stickerUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg sticker — send a sticker via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>  sticker file, URL, or Telegram file id — file_id is the common case (required)")
		fmt.Println("  --emoji <emoji>            associated emoji for a just-uploaded sticker (optional)")
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg sticker --to 123456789 --file ./sticker.webp`)
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		output.Error(fmt.Sprintf("reading config: %v", err), "")
		return 1
	}
	if cfg.BotToken == "" {
		output.Error("no bot token configured", `tg config set --bot-token "<token>"`)
		return 2
	}

	to, exitCode := resolveChatID(values, cfg)
	if exitCode != 0 {
		return exitCode
	}

	if values["file"] == "" {
		output.Error("--file is required", stickerUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, stickerUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["emoji"] != "" {
		extra["emoji"] = values["emoji"]
	}

	file := telegram.ResolveInputFile(values["file"])
	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendSticker", "sticker", file, common, extra, nil, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
