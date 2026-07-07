package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const pollUsage = `tg poll --to <chat_id> --question <question> --option <text> --option <text> [--option <text> ...] [--anonymous] [--quiz --correct-option <index>] [--multiple] [--explanation <text>] [--parse-mode Markdown|MarkdownV2|HTML] [--reply-to <message_id>] [--silent] [--protect] [--thread <message_thread_id>]`

type inputPollOption struct {
	Text string `json:"text"`
}

func cmdPoll(args []string) int {
	allowedFlags := append([]string{}, commonAllowedFlagNames...)
	allowedFlags = append(allowedFlags, "question", "correct-option", "explanation")
	booleanFlags := append([]string{}, commonBooleanFlagNames...)
	booleanFlags = append(booleanFlags, "anonymous", "quiz", "multiple")
	values, multi, err := cliflags.ParseWith(args, cliflags.Spec{
		Allowed:    allowedFlags,
		Repeatable: []string{"option"},
		Boolean:    booleanFlags,
	})
	if err != nil {
		return flagError(err, pollUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg poll — send a poll via the configured bot")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --to <chat_id>             destination chat id (defaults to config's default_chat_id)")
		fmt.Println("  --question <question>      poll question (required)")
		fmt.Println("  --option <text>            poll option text (required; repeat 2-10 times)")
		fmt.Println("  --anonymous                send as an anonymous poll (optional; Telegram defaults true)")
		fmt.Println("  --quiz                     send as a quiz poll (optional)")
		fmt.Println("  --correct-option <index>   correct option index, starting at 0 (required with --quiz)")
		fmt.Println("  --multiple                 allow multiple answers (optional)")
		fmt.Println("  --explanation <text>       quiz explanation text (optional)")
		fmt.Println("  --parse-mode <mode>        Markdown, MarkdownV2, or HTML (optional)")
		fmt.Println("  --reply-to <id>            reply to a message id (optional)")
		fmt.Println("  --silent                   disable notification (optional)")
		fmt.Println("  --protect                  protect content from forwarding/saving (optional)")
		fmt.Println("  --thread <id>              message thread id for forum topics (optional)")
		fmt.Println()
		fmt.Println("examples:")
		fmt.Println(`  tg poll --to 123456789 --question "Lunch?" --option Pizza --option Tacos`)
		fmt.Println(`  tg poll --to 123456789 --question "2+2?" --quiz --correct-option 1 --option 3 --option 4`)
		return 0
	}

	if values["question"] == "" {
		output.Error("--question is required", pollUsage)
		return 2
	}
	options := multi["option"]
	if len(options) < 2 || len(options) > 10 {
		output.Error(fmt.Sprintf("--option must be given 2-10 times (got %d)", len(options)), pollUsage)
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

	common, exitCode := commonParamsFrom(values, to, pollUsage)
	if exitCode != 0 {
		return exitCode
	}

	quiz, exitCode := boolFlag(values, "quiz")
	if exitCode != 0 {
		return exitCode
	}
	if quiz && values["correct-option"] == "" {
		output.Error("--quiz requires --correct-option", pollUsage)
		return 2
	}

	optionsJSON, err := buildPollOptionsJSON(options)
	if err != nil {
		output.Error(fmt.Sprintf("encoding poll options: %v", err), "")
		return 1
	}
	extra := map[string]string{
		"question": values["question"],
		"options":  optionsJSON,
	}

	if _, ok := values["anonymous"]; ok {
		anonymous, exitCode := boolFlag(values, "anonymous")
		if exitCode != 0 {
			return exitCode
		}
		extra["is_anonymous"] = strconv.FormatBool(anonymous)
	}
	if quiz {
		idx, exitCode := intFlag(values, "correct-option", pollUsage)
		if exitCode != 0 {
			return exitCode
		}
		if idx < 0 || idx >= len(options) {
			output.Error(fmt.Sprintf("--correct-option %d is out of range (0-%d for %d options)", idx, len(options)-1, len(options)), pollUsage)
			return 2
		}
		extra["type"] = "quiz"
		extra["correct_option_id"] = strconv.Itoa(idx)
	}
	multiple, exitCode := boolFlag(values, "multiple")
	if exitCode != 0 {
		return exitCode
	}
	if multiple {
		extra["allows_multiple_answers"] = "true"
	}
	if values["explanation"] != "" {
		extra["explanation"] = values["explanation"]
	}

	client := telegram.NewClient(cfg.BotToken)
	msg, exitCode := sendParams(client, "sendPoll", "poll", common, extra, "check --to and that the bot can message this chat")
	if exitCode != 0 {
		return exitCode
	}

	output.Line("sent: message %d to %s", msg.MessageID, to)
	return 0
}

func buildPollOptionsJSON(options []string) (string, error) {
	pollOptions := make([]inputPollOption, len(options))
	for i, option := range options {
		pollOptions[i] = inputPollOption{Text: option}
	}
	optionsJSON, err := json.Marshal(pollOptions)
	if err != nil {
		return "", err
	}
	return string(optionsJSON), nil
}
