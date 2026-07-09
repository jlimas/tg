package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
	"github.com/jlimas/tg/internal/telegram"
)

const listenUsage = `tg listen`

// listenPollTimeout is how long each getUpdates long-poll call waits for a
// new update before Telegram returns empty. listenHardCap is the total time
// tg listen will keep polling (in listenPollTimeout-sized chunks) before
// giving up and exiting.
const (
	listenPollTimeout = 25 * time.Second
	listenHardCap     = 5 * time.Minute
)

// updateFetcher is the subset of telegram.Client that listenForMessages
// needs, so tests can substitute a fake without hitting the network.
type updateFetcher interface {
	GetUpdates(ctx context.Context, offset int64, timeoutSeconds int) ([]telegram.Update, error)
}

func CmdListen(args []string) int {
	values, _, err := cliflags.ParseWith(args, cliflags.Spec{})
	if err != nil {
		return flagError(err, listenUsage)
	}
	if values["help"] == "true" {
		fmt.Println("tg listen — wait for the next message sent to the configured default chat")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  (none — listens on config's default_chat_id)")
		fmt.Println()
		fmt.Println("notes:")
		fmt.Println("  polls for up to 5m; messages from any chat other than")
		fmt.Println("  default_chat_id are silently discarded, not queued for a later call")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println("  tg listen")
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
	if cfg.DefaultChatID == "" {
		output.Error("no default_chat_id configured", `tg config set --default-chat-id <id>`)
		return 2
	}

	client := telegram.NewClient(cfg.BotToken)
	ctx, cancel := context.WithTimeout(context.Background(), listenHardCap)
	defer cancel()

	messages, err := listenForMessages(ctx, client, cfg.DefaultChatID)
	if err != nil {
		output.Error(fmt.Sprintf("listening for messages: %v", err), "")
		return 1
	}

	if len(messages) == 0 {
		output.Line("listen: no messages received from %s within %s", cfg.DefaultChatID, listenHardCap)
		return 0
	}

	fmt.Printf("messages[%d]{message_id,from,reply_to,text}:\n", len(messages))
	for _, m := range messages {
		fmt.Printf("  %d,%s,%s,%s\n", m.MessageID, toonField(fromDisplay(m)), replyToID(m), toonField(m.Text))
	}
	output.Help(`tg text --to <chat_id> --message "..."`)
	return 0
}

// listenForMessages long-polls fetcher via getUpdates in listenPollTimeout
// chunks until it sees at least one message from chatID, or ctx is done.
// Updates from other chats are acknowledged (advancing the offset) but
// otherwise discarded.
func listenForMessages(ctx context.Context, fetcher updateFetcher, chatID string) ([]telegram.IncomingMessage, error) {
	var offset int64

	// Drain any backlog first so only messages received after listen starts
	// are surfaced.
	backlog, err := fetcher.GetUpdates(ctx, offset, 0)
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil
		}
		return nil, err
	}
	offset = nextOffset(offset, backlog)

	for {
		if ctx.Err() != nil {
			return nil, nil
		}

		updates, err := fetcher.GetUpdates(ctx, offset, int(listenPollTimeout.Seconds()))
		if err != nil {
			if ctx.Err() != nil {
				return nil, nil
			}
			return nil, err
		}
		offset = nextOffset(offset, updates)

		var matched []telegram.IncomingMessage
		for _, u := range updates {
			if u.Message == nil {
				continue
			}
			if strconv.FormatInt(u.Message.Chat.ID, 10) != chatID {
				continue
			}
			matched = append(matched, *u.Message)
		}
		if len(matched) > 0 {
			return matched, nil
		}
	}
}

func nextOffset(current int64, updates []telegram.Update) int64 {
	next := current
	for _, u := range updates {
		if u.UpdateID+1 > next {
			next = u.UpdateID + 1
		}
	}
	return next
}

func fromDisplay(m telegram.IncomingMessage) string {
	if m.From.Username != "" {
		return "@" + m.From.Username
	}
	return m.From.FirstName
}

// replyToID returns the message_id m is replying to, or "" if it isn't a
// reply — lets callers correlate an inbound reply with the outbound
// message_id printed when it was sent.
func replyToID(m telegram.IncomingMessage) string {
	if m.ReplyToMessage == nil {
		return ""
	}
	return strconv.Itoa(m.ReplyToMessage.MessageID)
}

// toonField quotes s if it contains a character that would otherwise be
// ambiguous in TOON's comma-delimited row format.
func toonField(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
