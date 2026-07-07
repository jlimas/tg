package main

import (
	"context"
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const textUsage = `tg text --to <chat_id> --message "..." [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent=true] [--protect=true] [--thread <message_thread_id>]`

func cmdText(args []string) int {
	allowedFlags := append([]string{}, commonFlagNames...)
	allowedFlags = append(allowedFlags, "message")
	values, err := cliflags.Parse(args, allowedFlags)
	if err != nil {
		return flagError(err, textUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg text — send a text message via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>        destination chat id (defaults to config's default_chat_id)")
		fmt.Println(`  --message "<text>"    message text (required)`)
		fmt.Println("  --parse-mode <mode>   Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>       reply to a message id (optional)")
		fmt.Println("  --silent=true         disable notification (optional; explicit value required)")
		fmt.Println("  --protect=true        protect content from forwarding/saving (optional; explicit value required)")
		fmt.Println("  --thread <id>         message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg text --to 123456789 --message "hello from tg"`)
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

	text := values["message"]
	if text == "" {
		output.Error("--message is required", textUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, textUsage)
	if exitCode != 0 {
		return exitCode
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, err := client.SendMessage(context.Background(), telegram.SendMessageParams{
		ChatID:              common.ChatID,
		Text:                text,
		ParseMode:           common.ParseMode,
		ReplyToMessageID:    common.ReplyToMessageID,
		DisableNotification: common.DisableNotification,
		ProtectContent:      common.ProtectContent,
		MessageThreadID:     common.MessageThreadID,
	})
	if err != nil {
		output.Error(fmt.Sprintf("sending message: %v", err), "check --to and that the bot can message this chat")
		return 1
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
