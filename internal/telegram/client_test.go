package telegram

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

func TestSendMessageIncludesCommonParams(t *testing.T) {
	var gotQuery url.Values
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotQuery = req.URL.Query()
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

	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID:              "123",
		Text:                "hello",
		ReplyToMessageID:    321,
		DisableNotification: true,
		ProtectContent:      true,
		MessageThreadID:     654,
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	want := map[string]string{
		"chat_id":              "123",
		"text":                 "hello",
		"reply_parameters":     `{"message_id":321}`,
		"disable_notification": "true",
		"protect_content":      "true",
		"message_thread_id":    "654",
	}
	for name, value := range want {
		if gotQuery.Get(name) != value {
			t.Fatalf("%s = %q, want %q", name, gotQuery.Get(name), value)
		}
	}
}

func TestSendMediaRemoteURLAndFileIDUseForm(t *testing.T) {
	tests := []struct {
		name string
		file InputFile
	}{
		{
			name: "remote URL",
			file: InputFile{Kind: InputFileRemoteURL, Value: "https://example.com/cat.jpg"},
		},
		{
			name: "file ID",
			file: InputFile{Kind: InputFileFileID, Value: "AgACAgIAAx0FakeFileID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq *http.Request
			client := NewClient("TOKEN")
			client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotReq = req
				return okMessageResponse(), nil
			})

			msg, err := client.SendMedia(context.Background(), "sendPhoto", "photo", tt.file, CommonParams{
				ChatID:              "123",
				ParseMode:           "HTML",
				DisableNotification: true,
				MessageThreadID:     654,
			}, map[string]string{
				"caption":     "hi",
				"has_spoiler": "true",
				"chat_id":     "999",
			})
			if err != nil {
				t.Fatalf("SendMedia returned error: %v", err)
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
			if gotReq.URL.Path != "/botTOKEN/sendPhoto" {
				t.Fatalf("URL path = %q, want /botTOKEN/sendPhoto", gotReq.URL.Path)
			}
			if gotReq.Body != nil {
				t.Fatal("Body is non-nil, want nil")
			}
			if gotReq.Header.Get("Content-Type") != "" {
				t.Fatalf("Content-Type = %q, want empty", gotReq.Header.Get("Content-Type"))
			}

			gotQuery := gotReq.URL.Query()
			want := map[string]string{
				"chat_id":              "999",
				"parse_mode":           "HTML",
				"disable_notification": "true",
				"message_thread_id":    "654",
				"caption":              "hi",
				"has_spoiler":          "true",
				"photo":                tt.file.Value,
			}
			for name, value := range want {
				if gotQuery.Get(name) != value {
					t.Fatalf("%s = %q, want %q", name, gotQuery.Get(name), value)
				}
			}
		})
	}
}

func TestSendMediaLocalUploadUsesMultipart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cat.jpg")
	if err := os.WriteFile(path, []byte("photo bytes"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var gotReq *http.Request
	var gotForm *multipart.Form
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotReq = req
		form, err := req.MultipartReader()
		if err != nil {
			t.Fatalf("MultipartReader returned error: %v", err)
		}
		gotForm, err = form.ReadForm(1024 * 1024)
		if err != nil {
			t.Fatalf("ReadForm returned error: %v", err)
		}
		return okMessageResponse(), nil
	})

	_, err := client.SendMedia(context.Background(), "sendPhoto", "photo", InputFile{
		Kind:  InputFileLocalUpload,
		Value: path,
	}, CommonParams{
		ChatID:    "123",
		ParseMode: "Markdown",
	}, map[string]string{
		"caption":     "hi",
		"has_spoiler": "true",
	})
	if err != nil {
		t.Fatalf("SendMedia returned error: %v", err)
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.URL.RawQuery != "" {
		t.Fatalf("RawQuery = %q, want empty", gotReq.URL.RawQuery)
	}
	if !strings.HasPrefix(gotReq.Header.Get("Content-Type"), "multipart/form-data;") {
		t.Fatalf("Content-Type = %q, want multipart/form-data", gotReq.Header.Get("Content-Type"))
	}
	if gotForm == nil {
		t.Fatal("no multipart form captured")
	}
	defer gotForm.RemoveAll()

	wantValues := map[string]string{
		"chat_id":     "123",
		"parse_mode":  "Markdown",
		"caption":     "hi",
		"has_spoiler": "true",
	}
	for name, value := range wantValues {
		got := gotForm.Value[name]
		if len(got) != 1 || got[0] != value {
			t.Fatalf("%s = %#v, want [%q]", name, got, value)
		}
	}

	files := gotForm.File["photo"]
	if len(files) != 1 {
		t.Fatalf("len(files) = %d, want 1", len(files))
	}
	if files[0].Filename != "cat.jpg" {
		t.Fatalf("filename = %q, want cat.jpg", files[0].Filename)
	}
	uploaded, err := files[0].Open()
	if err != nil {
		t.Fatalf("open uploaded file: %v", err)
	}
	defer uploaded.Close()
	contents, err := io.ReadAll(uploaded)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(contents) != "photo bytes" {
		t.Fatalf("uploaded contents = %q, want photo bytes", contents)
	}
}

func okMessageResponse() *http.Response {
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
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
