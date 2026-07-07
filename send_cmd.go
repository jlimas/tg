package main

import (
	"context"
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

var validParseModes = map[string]bool{
	"":           true,
	"Markdown":   true,
	"MarkdownV2": true,
	"HTML":       true,
}

const sendUsage = `tg send --to <chat_id> --text "..." [--parse-mode Markdown|MarkdownV2|HTML]`

func cmdSend(args []string) int {
	values, err := cliflags.Parse(args, []string{"to", "text", "parse-mode"})
	if err != nil {
		return flagError(err, sendUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg send — send a text message via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>        destination chat id (defaults to config's default_chat_id)")
		fmt.Println(`  --text "<message>"    message text (required)`)
		fmt.Println("  --parse-mode <mode>   Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg send --to 123456789 --text "hello from tg"`)
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

	to := values["to"]
	if to == "" {
		to = cfg.DefaultChatID
	}
	if to == "" {
		output.Error("--to is required (no default_chat_id configured)", sendUsage)
		return 2
	}

	text := values["text"]
	if text == "" {
		output.Error("--text is required", sendUsage)
		return 2
	}

	parseMode := values["parse-mode"]
	if !validParseModes[parseMode] {
		output.Error(fmt.Sprintf("invalid --parse-mode %q", parseMode), "valid values: Markdown, MarkdownV2, HTML")
		return 2
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, err := client.SendMessage(context.Background(), telegram.SendMessageParams{
		ChatID:    to,
		Text:      text,
		ParseMode: parseMode,
	})
	if err != nil {
		output.Error(fmt.Sprintf("sending message: %v", err), "check --to and that the bot can message this chat")
		return 1
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
