package app

import (
	"context"
	"fmt"

	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

func sendParams(client *telegram.Client, method string, label string, common telegram.CommonParams, extra map[string]string, hint string) (*telegram.Message, int) {
	msg, err := client.Send(context.Background(), method, common, extra)
	if err != nil {
		output.Error(fmt.Sprintf("sending %s: %v", label, err), hint)
		return nil, 1
	}
	return msg, 0
}
