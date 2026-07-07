package app

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const locationUsage = `tg location --to <chat_id> --lat <latitude> --long <longitude> [--live-period <seconds>] [--accuracy <meters>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func CmdLocation(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "lat", "long", "live-period", "accuracy")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, locationUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg location — send a location via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>              destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --lat <latitude>            latitude as a decimal number (required)")
		fmt.Println("  --long <longitude>          longitude as a decimal number (required)")
		fmt.Println("  --live-period <seconds>     live location period in seconds (optional)")
		fmt.Println("  --accuracy <meters>         horizontal accuracy in meters (optional)")
		fmt.Println("  --parse-mode <mode>         Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>             reply to a message id (optional)")
		fmt.Println("  --silent                    disable notification (optional)")
		fmt.Println("  --protect                   protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>               message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg location --to 123456789 --lat 40.7580 --long -73.9855`)
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

	if values["lat"] == "" {
		output.Error("--lat is required", locationUsage)
		return 2
	}
	if values["long"] == "" {
		output.Error("--long is required", locationUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, locationUsage)
	if exitCode != 0 {
		return exitCode
	}

	lat, exitCode := floatFlag(values, "lat", locationUsage)
	if exitCode != 0 {
		return exitCode
	}
	long, exitCode := floatFlag(values, "long", locationUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{
		"latitude":  strconv.FormatFloat(lat, 'f', -1, 64),
		"longitude": strconv.FormatFloat(long, 'f', -1, 64),
	}

	livePeriod, exitCode := intFlag(values, "live-period", locationUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["live-period"] != "" {
		extra["live_period"] = strconv.Itoa(livePeriod)
	}

	accuracy, exitCode := floatFlag(values, "accuracy", locationUsage)
	if exitCode != 0 {
		return exitCode
	}
	if values["accuracy"] != "" {
		extra["horizontal_accuracy"] = strconv.FormatFloat(accuracy, 'f', -1, 64)
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendParams(client, "sendLocation", "location", common, extra, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
