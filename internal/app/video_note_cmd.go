package app

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const videoNoteUsage = `tg video-note --to <chat_id> --file <path|file_id> [--duration <seconds>] [--length <px>] [--thumbnail <path|url|file_id>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func CmdVideoNote(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "duration", "length", "thumbnail")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, videoNoteUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg video-note — send a video note via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>                  destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|file_id>           video note file or Telegram file id; URLs are not supported (required)")
		fmt.Println("  --duration <seconds>            video note duration in seconds (optional)")
		fmt.Println("  --length <px>                   video note side diameter in pixels (optional)")
		fmt.Println("  --thumbnail <path|url|file_id>  thumbnail file, URL, or Telegram file id (optional)")
		fmt.Println("  --parse-mode <mode>             Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>                 reply to a message id (optional)")
		fmt.Println("  --silent                        disable notification (optional)")
		fmt.Println("  --protect                       protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>                   message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg video-note --to 123456789 --file ./round.mp4`)
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
		output.Error("--file is required", videoNoteUsage)
		return 2
	}

	file := telegram.ResolveInputFile(values["file"])
	if file.Kind == telegram.InputFileRemoteURL {
		output.Error("--file must be a local path or file_id (video notes don't support URLs)", "download the file locally or use a previously uploaded file_id")
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, videoNoteUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	duration, exitCode := intFlag(values, "duration", videoNoteUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["duration"] != "" {
		extra["duration"] = strconv.Itoa(duration)
	}
	length, exitCode := intFlag(values, "length", videoNoteUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["length"] != "" {
		extra["length"] = strconv.Itoa(length)
	}

	var extraFiles map[string]telegram.InputFile
	if values["thumbnail"] != "" {
		extraFiles = map[string]telegram.InputFile{
			"thumbnail": telegram.ResolveInputFile(values["thumbnail"]),
		}
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendVideoNote", "video_note", file, common, extra, extraFiles, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
