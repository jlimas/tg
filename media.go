package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

func sendMedia(client *telegram.Client, method string, field string, file telegram.InputFile, common telegram.CommonParams, extra map[string]string, extraFiles map[string]telegram.InputFile) (*telegram.Message, int) {
	msg, err := client.SendMedia(context.Background(), method, field, file, common, extra, extraFiles)
	if err != nil {
		output.Error(fmt.Sprintf("sending %s: %v", strings.ReplaceAll(field, "_", " "), err), "check --to and that the bot can message this chat")
		return nil, 1
	}
	return msg, 0
}
