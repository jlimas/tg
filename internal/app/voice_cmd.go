package app

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const voiceUsage = `tg voice --to <chat_id> --file <path|url|file_id> [--caption "..."] [--duration <seconds>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func CmdVoice(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption", "duration")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, voiceUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg voice — send a voice message via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>  voice file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"         caption text (optional)`)
		fmt.Println("  --duration <seconds>       voice duration in seconds (optional)")
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg voice --to 123456789 --file ./note.ogg --caption "hi"`)
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
		output.Error("--file is required", voiceUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, voiceUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}
	duration, exitCode := intFlag(values, "duration", voiceUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["duration"] != "" {
		extra["duration"] = strconv.Itoa(duration)
	}

	file := telegram.ResolveInputFile(values["file"])
	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendVoice", "voice", file, common, extra, nil, "voice requires an OGG/OPUS file — try tg audio or tg document for other audio formats")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
