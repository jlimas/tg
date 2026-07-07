package main

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const contactUsage = `tg contact --to <chat_id> --phone <phone_number> --first-name <first_name> [--last-name <last_name>] [--vcard <vcard>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

func cmdContact(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "phone", "first-name", "last-name", "vcard")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed: allowedFlags,
		Boolean: booleanFlags,
	})
	if err != nil {
		return flagError(err, contactUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg contact — send a contact via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>              destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --phone <phone_number>      contact phone number (required)")
		fmt.Println("  --first-name <first_name>   contact first name (required)")
		fmt.Println("  --last-name <last_name>     contact last name (optional)")
		fmt.Println("  --vcard <vcard>             raw vCard string (optional)")
		fmt.Println("  --parse-mode <mode>         Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>             reply to a message id (optional)")
		fmt.Println("  --silent                    disable notification (optional)")
		fmt.Println("  --protect                   protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>               message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg contact --to 123456789 --phone "+15551234567" --first-name "Ada"`)
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

	if values["phone"] == "" {
		output.Error("--phone is required", contactUsage)
		return 2
	}
	if values["first-name"] == "" {
		output.Error("--first-name is required", contactUsage)
		return 2
	}

	common, exitCode := commonParamsFrom(values, to, contactUsage)
	if exitCode != 0 {
		return exitCode
	}

	extra := map[string]string{
		"phone_number": values["phone"],
		"first_name":   values["first-name"],
	}
	if values["last-name"] != "" {
		extra["last_name"] = values["last-name"]
	}
	if values["vcard"] != "" {
		extra["vcard"] = values["vcard"]
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendParams(client, "sendContact", "contact", common, extra, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}
