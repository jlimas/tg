package telegram

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveInputFileLocalPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(path, []byte("photo"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	file := ResolveInputFile(path)
	if file.Kind != InputFileLocalUpload {
		t.Fatalf("Kind = %q, want %q", file.Kind, InputFileLocalUpload)
	}
	if file.Value != path {
		t.Fatalf("Value = %q, want %q", file.Value, path)
	}
}

func TestResolveInputFileRemoteURL(t *testing.T) {
	const value = "https://example.com/a.jpg"

	file := ResolveInputFile(value)
	if file.Kind != InputFileRemoteURL {
		t.Fatalf("Kind = %q, want %q", file.Kind, InputFileRemoteURL)
	}
	if file.Value != value {
		t.Fatalf("Value = %q, want %q", file.Value, value)
	}
}

func TestResolveInputFileFileID(t *testing.T) {
	const value = "AgACAgIAAx0FakeFileID"

	file := ResolveInputFile(value)
	if file.Kind != InputFileFileID {
		t.Fatalf("Kind = %q, want %q", file.Kind, InputFileFileID)
	}
	if file.Value != value {
		t.Fatalf("Value = %q, want %q", file.Value, value)
	}
}

func TestResolveInputFileStatWinsOverURLPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows paths cannot contain the colon needed for this URL-shaped relative path")
	}

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "https:", "example.com"), 0o700); err != nil {
		t.Fatalf("make URL-shaped directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "https:", "example.com", "a.jpg"), []byte("photo"), 0o600); err != nil {
		t.Fatalf("write URL-shaped temp file: %v", err)
	}
	t.Chdir(tmp)

	const value = "https://example.com/a.jpg"
	file := ResolveInputFile(value)
	if file.Kind != InputFileLocalUpload {
		t.Fatalf("Kind = %q, want %q", file.Kind, InputFileLocalUpload)
	}
	if file.Value != value {
		t.Fatalf("Value = %q, want %q", file.Value, value)
	}
}

func TestSendMessagePreservesRawQueryPost(t *testing.T) {
	var gotReq *http.Request
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotReq = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"ok": true,
				"result": {
					"message_id": 42,
					"date": 1234567890,
					"chat": {"id": 123}
				}
			}`)),
			Header: make(http.Header),
		}, nil
	})

	msg, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID:    "123",
		Text:      "hello world",
		ParseMode: "HTML",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if msg.MessageID != 42 {
		t.Fatalf("MessageID = %d, want 42", msg.MessageID)
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.Method != http.MethodPost {
		t.Fatalf("Method = %q, want %q", gotReq.Method, http.MethodPost)
	}
	if gotReq.URL.String() != "https://api.telegram.org/botTOKEN/sendMessage?chat_id=123&parse_mode=HTML&text=hello+world" {
		t.Fatalf("URL = %q", gotReq.URL.String())
	}
	if gotReq.Body != nil {
		t.Fatal("Body is non-nil, want nil")
	}
	if gotReq.Header.Get("Content-Type") != "" {
		t.Fatalf("Content-Type = %q, want empty", gotReq.Header.Get("Content-Type"))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
