package main

import (
	"context"
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const mediaGroupUsage = `tg media-group --to <chat_id> --file <path|url|file_id> --file <path|url|file_id> [--type photo|video|audio|document] [--caption "..."] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

var validMediaGroupTypes = map[string]bool{
	"photo":    true,
	"video":    true,
	"audio":    true,
	"document": true,
}

func cmdMediaGroup(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "type", "caption")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, multi, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed:    allowedFlags,
		Repeatable: []string{"file"},
		Boolean:    booleanFlags,
	})
	if err != nil {
		return flagError(err, mediaGroupUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg media-group - send a media album via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>  media file, URL, or Telegram file id (required; repeat 2-10 times)")
		fmt.Println("  --type <type>              photo, video, audio, or document (default: photo)")
		fmt.Println(`  --caption "<text>"         album caption on the first item only (optional)`)
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg media-group --to 123456789 --file a.jpg --file b.jpg`)
		return 0
	}

	fileValues := multi["file"]
	if len(fileValues) < 2 || len(fileValues) > 10 {
		output.Error(fmt.Sprintf("--file must be given 2-10 times (got %d)", len(fileValues)), mediaGroupUsage)
		return 2
	}

	mediaType := values["type"]
	if mediaType == "" {
		mediaType = "photo"
	}
	if !validMediaGroupTypes[mediaType] {
		output.Error(fmt.Sprintf("invalid --type %q", mediaType), "valid values: photo, video, audio, document")
		return 2
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

	common, exitCode := commonParamsFrom(values, to, mediaGroupUsage)
	if exitCode != 0 {
		return exitCode
	}

	files := make([]telegram.InputFile, len(fileValues))
	for i, value := range fileValues {
		files[i] = telegram.ResolveInputFile(value)
	}

	client := telegram.NewClient(cfg.BotToken)
	msgs, err := client.SendMediaGroup(context.Background(), common, mediaType, values["caption"], files)
	if err != nil {
		output.Error(fmt.Sprintf("sending media group: %v", err), "check --to and that the bot can message this chat")
		return 1
	}

	output.Line("sent: %d messages to %s", len(msgs), to)
	return 0
}
