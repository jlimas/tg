package main

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const audioUsage = `tg audio --to <chat_id> --file <path|url|file_id> [--caption "..."] [--duration <seconds>] [--performer <name>] [--title <title>] [--thumbnail <path|url|file_id>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func cmdAudio(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption", "duration", "performer", "title", "thumbnail")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, audioUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg audio — send an audio file via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>                  destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>       audio file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"              caption text (optional)`)
		fmt.Println("  --duration <seconds>            audio duration in seconds (optional)")
		fmt.Println("  --performer <name>              performer name (optional)")
		fmt.Println("  --title <title>                 track title (optional)")
		fmt.Println("  --thumbnail <path|url|file_id>  thumbnail file, URL, or Telegram file id (optional)")
		fmt.Println("  --parse-mode <mode>             Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>                 reply to a message id (optional)")
		fmt.Println("  --silent                        disable notification (optional)")
		fmt.Println("  --protect                       protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>                   message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg audio --to 123456789 --file ./song.mp3 --performer "X" --title "Y"`)
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
		output.Error("--file is required", audioUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, audioUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}
	duration, exitCode := intFlag(values, "duration", audioUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["duration"] != "" {
		extra["duration"] = strconv.Itoa(duration)
	}
	if values["performer"] != "" {
		extra["performer"] = values["performer"]
	}
	if values["title"] != "" {
		extra["title"] = values["title"]
	}

	file := telegram.ResolveInputFile(values["file"])
	var extraFiles map[string]telegram.InputFile
	if values["thumbnail"] != "" {
		extraFiles = map[string]telegram.InputFile{
			"thumbnail": telegram.ResolveInputFile(values["thumbnail"]),
		}
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendAudio", "audio", file, common, extra, extraFiles)
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
