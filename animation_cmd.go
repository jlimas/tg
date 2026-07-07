package main

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const animationUsage = `tg animation --to <chat_id> --file <path|url|file_id> [--caption "..."] [--duration <seconds>] [--width <px>] [--height <px>] [--thumbnail <path|url|file_id>] [--spoiler] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func cmdAnimation(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "file", "caption", "duration", "width", "height", "thumbnail")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	booleanFlags = append(booleanFlags, "spoiler")
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, animationUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg animation — send an animation via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>                  destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --file <path|url|file_id>       animation file, URL, or Telegram file id (required)")
		fmt.Println(`  --caption "<text>"              caption text (optional)`)
		fmt.Println("  --duration <seconds>            animation duration in seconds (optional)")
		fmt.Println("  --width <px>                    animation width (optional)")
		fmt.Println("  --height <px>                   animation height (optional)")
		fmt.Println("  --thumbnail <path|url|file_id>  thumbnail file, URL, or Telegram file id (optional)")
		fmt.Println("  --spoiler                       mark animation with spoiler animation (optional)")
		fmt.Println("  --parse-mode <mode>             Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>                 reply to a message id (optional)")
		fmt.Println("  --silent                        disable notification (optional)")
		fmt.Println("  --protect                       protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>                   message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg animation --to 123456789 --file ./loop.gif --caption "launch"`)
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
		output.Error("--file is required", animationUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, animationUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["caption"] != "" {
		extra["caption"] = values["caption"]
	}
	duration, exitCode := intFlag(values, "duration", animationUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["duration"] != "" {
		extra["duration"] = strconv.Itoa(duration)
	}
	width, exitCode := intFlag(values, "width", animationUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["width"] != "" {
		extra["width"] = strconv.Itoa(width)
	}
	height, exitCode := intFlag(values, "height", animationUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["height"] != "" {
		extra["height"] = strconv.Itoa(height)
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
	msg, exitCode := sendMedia(client, "sendAnimation", "animation", file, common, extra, extraFiles, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
