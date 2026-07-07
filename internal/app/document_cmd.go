package app

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const documentUsage = `tg document --to <chat_id> --file <path|url|file_id> [--caption "..."] [--thumbnail <path|url|file_id>] [--file-name <name>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func CmdDocument(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption", "thumbnail", "file-name")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, documentUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg document — send a document via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>                  destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>       document file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"              caption text (optional)`)
		fmt.Println("  --thumbnail <path|url|file_id>  thumbnail file, URL, or Telegram file id (optional)")
		fmt.Println("  --file-name <name>              displayed filename for local uploads (optional)")
		fmt.Println("  --parse-mode <mode>             Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>                 reply to a message id (optional)")
		fmt.Println("  --silent                        disable notification (optional)")
		fmt.Println("  --protect                       protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>                   message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg document --to 123456789 --file ./report.pdf --caption "Q3"`)
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
		output.Error("--file is required", documentUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, documentUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}

	file := telegram.ResolveInputFile(values["file"])
	if values["file-name"] != "" {
		file.FileName = values["file-name"]
	}

	var extraFiles map[string]telegram.InputFile
	if values["thumbnail"] != "" {
		extraFiles = map[string]telegram.InputFile{
			"thumbnail": telegram.ResolveInputFile(values["thumbnail"]),
		}
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendDocument", "document", file, common, extra, extraFiles, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
