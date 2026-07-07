package main

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const photoUsage = `tg photo --to <chat_id> --file <path|url|file_id> [--caption "..."] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>] [--spoiler]`

func cmdPhoto(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	booleanFlags = append(booleanFlags, "spoiler")
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, photoUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg photo — send a photo via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>  photo file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"         caption text (optional)`)
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println("  --spoiler                  mark photo with spoiler animation (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg photo --to 123456789 --file ./cat.jpg --caption "hi"`)
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
		output.Error("--file is required", photoUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, photoUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}
	spoiler, exitCode := boolFlag(values, "spoiler")
	if exitCode != 0 {
		return exitCode
	}
	if spoiler {
		extra["has_spoiler"] = "true"
	}

	file := telegram.ResolveInputFile(values["file"])
	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendPhoto", "photo", file, common, extra, nil)
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
