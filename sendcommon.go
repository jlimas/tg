package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

var commonFlagNames = []string{"to", "parse-mode", "reply-to", "silent", "protect", "thread"}

var validParseModes = map[string]bool{
	"":           true,
	"Markdown":   true,
	"MarkdownV2": true,
	"HTML":       true,
}

func resolveChatID(values map[string]string, cfg config.Config) (string, int) {
	chatID := values["to"]
	if chatID == "" {
		chatID = cfg.DefaultChatID
	}
	if chatID == "" {
		output.Error("--to is required (no default_chat_id configured)", "tg config set --default-chat-id <id>")
		return "", 2
	}
	return chatID, 0
}

func commonParamsFrom(values map[string]string, chatID string, usage string) (telegram.CommonParams, int) {
	parseMode := values["parse-mode"]
	if !validParseModes[parseMode] {
		output.Error(fmt.Sprintf("invalid --parse-mode %q", parseMode), "valid values: Markdown, MarkdownV2, HTML")
		return telegram.CommonParams{}, 2
	}

	replyTo, exitCode := intFlag(values, "reply-to", usage)
	if exitCode != 0 {
		return telegram.CommonParams{}, exitCode
	}
	threadID, exitCode := intFlag(values, "thread", usage)
	if exitCode != 0 {
		return telegram.CommonParams{}, exitCode
	}

	// cliflags does not support valueless boolean flags yet, so common send
	// booleans intentionally require an explicit value such as --silent=true.
	silent, exitCode := boolFlag(values, "silent")
	if exitCode != 0 {
		return telegram.CommonParams{}, exitCode
	}
	protect, exitCode := boolFlag(values, "protect")
	if exitCode != 0 {
		return telegram.CommonParams{}, exitCode
	}

	return telegram.CommonParams{
		ChatID:              chatID,
		ParseMode:           parseMode,
		ReplyToMessageID:    replyTo,
		DisableNotification: silent,
		ProtectContent:      protect,
		MessageThreadID:     threadID,
	}, 0
}

func intFlag(values map[string]string, name string, usage string) (int, int) {
	value, ok := values[name]
	if !ok {
		return 0, 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		output.Error(fmt.Sprintf("invalid --%s %q", name, value), usage)
		return 0, 2
	}
	return parsed, 0
}

func boolFlag(values map[string]string, name string) (bool, int) {
	value, ok := values[name]
	if !ok {
		return false, 0
	}
	switch strings.ToLower(value) {
	case "true":
		return true, 0
	case "false":
		return false, 0
	default:
		output.Error(fmt.Sprintf("invalid --%s %q", name, value), fmt.Sprintf("use --%s=true or --%s=false", name, name))
		return false, 2
	}
}
