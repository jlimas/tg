package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

func sendMedia(client *telegram.Client, method string, field string, file telegram.InputFile, common telegram.CommonParams, extra map[string]string, extraFiles map[string]telegram.InputFile, hint string) (*telegram.Message, int) {
	msg, err := client.SendMedia(context.Background(), method, field, file, common, extra, extraFiles)
	if err != nil {
		output.Error(fmt.Sprintf("sending %s: %v", strings.ReplaceAll(field, "_", " "), err), hint)
		return nil, 1
	}
	return msg, 0
}
