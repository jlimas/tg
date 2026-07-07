package app

import (
	"context"
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const paidMediaUsage = `tg paid-media --to <chat_id> --star-count <stars> --file <path|url|file_id> [--file <path|url|file_id> ...] [--type photo|video] [--caption "..."] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

var validPaidMediaTypes = map[string]bool{
	"photo": true,
	"video": true,
}

func CmdPaidMedia(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "star-count", "type", "caption")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, multi, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed:    allowedFlags,
		Repeatable: []string{"file"},
		Boolean:    booleanFlags,
	})
	if err != nil {
		return flagError(err, paidMediaUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg paid-media - send paid media via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --star-count <stars>       number of Telegram Stars required to unlock the media (required)")
		fmt.Println("  --file <path|url|file_id>  paid media file, URL, or Telegram file id (required; repeatable)")
		fmt.Println("  --type <type>              photo or video (default: photo)")
		fmt.Println(`  --caption "<text>"         caption text for the paid media message (optional)`)
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg paid-media --to 123456789 --star-count 50 --file preview.jpg`)
		return 0
	}

	if values["star-count"] == "" {
		output.Error("--star-count is required", paidMediaUsage)
		return 2
	}
	starCount, exitCode := intFlag(values, "star-count", paidMediaUsage)
	if exitCode != 0 {
		return exitCode
	}
	if starCount <= 0 {
		output.Error("--star-count must be a positive integer", paidMediaUsage)
		return 2
	}

	fileValues := multi["file"]
	if len(fileValues) < 1 {
		output.Error("--file must be given at least once", paidMediaUsage)
		return 2
	}

	mediaType := values["type"]
	if mediaType == "" {
		mediaType = "photo"
	}
	if !validPaidMediaTypes[mediaType] {
		output.Error(fmt.Sprintf("invalid --type %q", mediaType), "valid values: photo, video")
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

	common, exitCode := commonParamsFrom(values, to, paidMediaUsage)
	if exitCode != 0 {
		return exitCode
	}

	files := make([]telegram.InputFile, len(fileValues))
	for i, value := range fileValues {
		files[i] = telegram.ResolveInputFile(value)
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, err := client.SendPaidMedia(context.Background(), common, starCount, mediaType, files, extra)
	if err != nil {
		output.Error(fmt.Sprintf("sending paid media: %v", err), "check --to and that the bot can message this chat")
		return 1
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
