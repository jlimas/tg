package app

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const diceUsage = `tg dice --to <chat_id> [--emoji 🎲|🎯|🏀|⚽|🎳|🎰] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

var validDiceEmoji = map[string]bool{
	"🎲": true,
	"🎯": true,
	"🏀": true,
	"⚽": true,
	"🎳": true,
	"🎰": true,
}

func CmdDice(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "emoji")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, diceUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg dice — send a dice animation via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>        destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --emoji <emoji>       🎲, 🎯, 🏀, ⚽, 🎳, or 🎰 (optional; defaults to 🎲)")
		fmt.Println("  --parse-mode <mode>   Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>       reply to a message id (optional)")
		fmt.Println("  --silent              disable notification (optional)")
		fmt.Println("  --protect             protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>         message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("examples:")
		fmt.Println(`  tg dice --to 123456789`)
		fmt.Println(`  tg dice --to 123456789 --emoji "🎯"`)
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

	common, exitCode := commonParamsFrom(values, to, diceUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{}
	if values["emoji"] != "" {
		if !validDiceEmoji[values["emoji"]] {
			output.Error(fmt.Sprintf("invalid --emoji %q", values["emoji"]), "valid values: 🎲 🎯 🏀 ⚽ 🎳 🎰")
			return 2
		}
		extra["emoji"] = values["emoji"]
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendParams(client, "sendDice", "dice", common, extra, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
