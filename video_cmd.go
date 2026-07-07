package main

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const videoUsage = `tg video --to <chat_id> --file <path|url|file_id> [--caption "..."] [--duration <seconds>] [--width <px>] [--height <px>] [--thumbnail <path|url|file_id>] [--supports-streaming] [--spoiler] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func cmdVideo(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption", "duration", "width", "height", "thumbnail")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	booleanFlags = append(booleanFlags, "supports-streaming", "spoiler")
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, videoUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg video — send a video via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>                  destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>       video file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"              caption text (optional)`)
		fmt.Println("  --duration <seconds>            video duration in seconds (optional)")
		fmt.Println("  --width <px>                    video width (optional)")
		fmt.Println("  --height <px>                   video height (optional)")
		fmt.Println("  --thumbnail <path|url|file_id>  thumbnail file, URL, or Telegram file id (optional)")
		fmt.Println("  --supports-streaming            allow streaming playback (optional)")
		fmt.Println("  --spoiler                       mark video with spoiler animation (optional)")
		fmt.Println("  --parse-mode <mode>             Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>                 reply to a message id (optional)")
		fmt.Println("  --silent                        disable notification (optional)")
		fmt.Println("  --protect                       protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>                   message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg video --to 123456789 --file ./clip.mp4 --caption "launch" --supports-streaming`)
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
		output.Error("--file is required", videoUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, videoUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}
	duration, exitCode := intFlag(values, "duration", videoUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["duration"] != "" {
		extra["duration"] = strconv.Itoa(duration)
	}
	width, exitCode := intFlag(values, "width", videoUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["width"] != "" {
		extra["width"] = strconv.Itoa(width)
	}
	height, exitCode := intFlag(values, "height", videoUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["height"] != "" {
		extra["height"] = strconv.Itoa(height)
	}
	supportsStreaming, exitCode := boolFlag(values, "supports-streaming")
	if exitCode != 0 {
		return exitCode
	}
	if supportsStreaming {
		extra["supports_streaming"] = "true"
	}
	spoiler, exitCode := boolFlag(values, "spoiler")
	if exitCode != 0 {
		return exitCode
	}
	if spoiler {
		extra["has_spoiler"] = "true"
	}

	file := telegram.ResolveInputFile(values["file"])
	var extraFiles map[string]telegram.InputFile
	if values["thumbnail"] != "" {
		extraFiles = map[string]telegram.InputFile{
			"thumbnail": telegram.ResolveInputFile(values["thumbnail"]),
		}
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendMedia(client, "sendVideo", "video", file, common, extra, extraFiles)
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
