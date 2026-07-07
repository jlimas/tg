package app

import (
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const venueUsage = `tg venue --to <chat_id> --lat <latitude> --long <longitude> --title <title> --address <address> [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func CmdVenue(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "lat", "long", "title", "address")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, venueUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg venue — send a venue via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>              destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --lat <latitude>            latitude as a decimal number (required)")
		fmt.Println("  --long <longitude>          longitude as a decimal number (required)")
		fmt.Println("  --title <title>             venue title (required)")
		fmt.Println("  --address <address>         venue address (required)")
		fmt.Println("  --parse-mode <mode>         Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>             reply to a message id (optional)")
		fmt.Println("  --silent                    disable notification (optional)")
		fmt.Println("  --protect                   protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>               message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg venue --to 123456789 --lat 40.75 --long -73.98 --title "MSG" --address "4 Penn Plaza"`)
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
		output.Error("--lat is required", venueUsage)
		return 2
	}
	if values["long"] == "" {
		output.Error("--long is required", venueUsage)
		return 2
	}
	if values["title"] == "" {
		output.Error("--title is required", venueUsage)
		return 2
	}
	if values["address"] == "" {
		output.Error("--address is required", venueUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, venueUsage)
	if exitCode != 0 {
		return exitCode
	}

	lat, exitCode := floatFlag(values, "lat", venueUsage)
	if exitCode != 0 {
		return exitCode
	}
	long, exitCode := floatFlag(values, "long", venueUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{
		"latitude":  strconv.FormatFloat(lat, 'f', -1, 64),
		"longitude": strconv.FormatFloat(long, 'f', -1, 64),
		"title":     values["title"],
		"address":   values["address"],
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendParams(client, "sendVenue", "venue", common, extra, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
