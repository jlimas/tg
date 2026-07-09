package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jlimas/tg/internal/telegram"
)

func TestCmdListenRequiresBotToken(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	exitCode := CmdListen(nil)
	if exitCode != 2 {
		t.Fatalf("CmdListen exitCode = %d, want 2", exitCode)
	}
}

func TestCmdListenRequiresDefaultChatID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	configDir := filepath.Join(home, ".config", "tg")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("make config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("bot_token = \"123:ABC\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	output := captureStdout(t, func() int {
		exitCode := CmdListen(nil)
		if exitCode != 2 {
			t.Fatalf("CmdListen exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})
	if !strings.Contains(output, "no default_chat_id configured") {
		t.Fatalf("CmdListen output = %q, want default_chat_id error", output)
	}
}

func TestCmdListenRejectsUnknownFlag(t *testing.T) {
	setTestConfigHome(t)

	exitCode := CmdListen([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("CmdListen exitCode = %d, want 2", exitCode)
	}
}

func TestCmdListenHelp(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		return CmdListen([]string{"--help"})
	})
	if !strings.Contains(output, "tg listen") {
		t.Fatalf("CmdListen --help output = %q, want usage text", output)
	}
}

// fakeFetcher replays a canned sequence of GetUpdates responses, one per
// call, and records the offsets it was called with.
type fakeFetcher struct {
	responses [][]telegram.Update
	offsets   []int64
	call      int
}

func (f *fakeFetcher) GetUpdates(ctx context.Context, offset int64, timeoutSeconds int) ([]telegram.Update, error) {
	f.offsets = append(f.offsets, offset)
	if f.call >= len(f.responses) {
		return nil, nil
	}
	updates := f.responses[f.call]
	f.call++
	return updates, nil
}

func messageUpdate(updateID int64, chatID int64, text string) telegram.Update {
	u := telegram.Update{UpdateID: updateID}
	u.Message = &telegram.IncomingMessage{MessageID: int(updateID), Text: text}
	u.Message.Chat.ID = chatID
	return u
}

func TestListenForMessagesReturnsMatchAndAdvancesOffset(t *testing.T) {
	fetcher := &fakeFetcher{
		responses: [][]telegram.Update{
			{}, // backlog drain: nothing pending
			{messageUpdate(10, 999, "from another chat")},
			{messageUpdate(11, 123, "hello")},
		},
	}

	messages, err := listenForMessages(context.Background(), fetcher, "123")
	if err != nil {
		t.Fatalf("listenForMessages returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Text != "hello" {
		t.Fatalf("Text = %q, want %q", messages[0].Text, "hello")
	}

	wantOffsets := []int64{0, 0, 11}
	if len(fetcher.offsets) != len(wantOffsets) {
		t.Fatalf("offsets = %v, want %v", fetcher.offsets, wantOffsets)
	}
	for i, want := range wantOffsets {
		if fetcher.offsets[i] != want {
			t.Fatalf("offsets[%d] = %d, want %d", i, fetcher.offsets[i], want)
		}
	}
}

func TestListenForMessagesDiscardsOtherChatsUntilHardCap(t *testing.T) {
	fetcher := &fakeFetcher{
		responses: [][]telegram.Update{
			{}, // backlog drain
			{messageUpdate(1, 999, "not for us")},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	messages, err := listenForMessages(ctx, fetcher, "123")
	if err != nil {
		t.Fatalf("listenForMessages returned error: %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("messages = %#v, want none", messages)
	}
}

func TestNextOffsetTracksHighestUpdateID(t *testing.T) {
	updates := []telegram.Update{
		messageUpdate(5, 1, "a"),
		messageUpdate(9, 1, "b"),
		messageUpdate(7, 1, "c"),
	}
	if got := nextOffset(0, updates); got != 10 {
		t.Fatalf("nextOffset = %d, want 10", got)
	}
	if got := nextOffset(20, updates); got != 20 {
		t.Fatalf("nextOffset = %d, want 20 (unchanged when already ahead)", got)
	}
}

func TestFromDisplayPrefersUsername(t *testing.T) {
	m := telegram.IncomingMessage{}
	m.From.FirstName = "Ada"
	if got := fromDisplay(m); got != "Ada" {
		t.Fatalf("fromDisplay = %q, want %q", got, "Ada")
	}
	m.From.Username = "ada"
	if got := fromDisplay(m); got != "@ada" {
		t.Fatalf("fromDisplay = %q, want %q", got, "@ada")
	}
}

func TestReplyToIDReturnsEmptyWhenNotAReply(t *testing.T) {
	m := telegram.IncomingMessage{}
	if got := replyToID(m); got != "" {
		t.Fatalf("replyToID = %q, want empty", got)
	}
}

func TestReplyToIDReturnsRepliedMessageID(t *testing.T) {
	m := telegram.IncomingMessage{}
	m.ReplyToMessage = &struct {
		MessageID int `json:"message_id"`
	}{MessageID: 142}
	if got := replyToID(m); got != "142" {
		t.Fatalf("replyToID = %q, want %q", got, "142")
	}
}

func TestToonFieldQuotesSpecialChars(t *testing.T) {
	if got := toonField("plain"); got != "plain" {
		t.Fatalf("toonField(plain) = %q, want %q", got, "plain")
	}
	if got := toonField(`has, comma`); got != `"has, comma"` {
		t.Fatalf("toonField(comma) = %q", got)
	}
	if got := toonField(`has "quote"`); got != `"has ""quote"""` {
		t.Fatalf("toonField(quote) = %q", got)
	}
}
