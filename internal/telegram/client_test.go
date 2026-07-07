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

func TestSendUsesFormRequestWithCommonAndExtraParams(t *testing.T) {
	var gotReq *http.Request
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotReq = req
		return okMessageResponse(), nil
	})

	msg, err := client.Send(context.Background(), "sendLocation", CommonParams{
		ChatID:              "123",
		DisableNotification: true,
		MessageThreadID:     654,
	}, map[string]string{
		"latitude":  "40.758",
		"longitude": "-73.9855",
		"chat_id":   "999",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
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
	if gotReq.URL.Path != "/botTOKEN/sendLocation" {
		t.Fatalf("URL path = %q, want /botTOKEN/sendLocation", gotReq.URL.Path)
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
		"disable_notification": "true",
		"message_thread_id":    "654",
		"latitude":             "40.758",
		"longitude":            "-73.9855",
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
			}, nil)
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
	}, nil)
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

func TestSendMediaLocalUploadIncludesExtraFiles(t *testing.T) {
	dir := t.TempDir()
	documentPath := filepath.Join(dir, "report.pdf")
	if err := os.WriteFile(documentPath, []byte("document bytes"), 0o600); err != nil {
		t.Fatalf("write document file: %v", err)
	}
	thumbnailPath := filepath.Join(dir, "cover.jpg")
	if err := os.WriteFile(thumbnailPath, []byte("thumbnail bytes"), 0o600); err != nil {
		t.Fatalf("write thumbnail file: %v", err)
	}

	var gotForm *multipart.Form
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
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

	_, err := client.SendMedia(context.Background(), "sendDocument", "document", InputFile{
		Kind:  InputFileLocalUpload,
		Value: documentPath,
	}, CommonParams{
		ChatID: "123",
	}, map[string]string{
		"caption": "Q3",
	}, map[string]InputFile{
		"thumbnail": {
			Kind:  InputFileLocalUpload,
			Value: thumbnailPath,
		},
	})
	if err != nil {
		t.Fatalf("SendMedia returned error: %v", err)
	}
	if gotForm == nil {
		t.Fatal("no multipart form captured")
	}
	defer gotForm.RemoveAll()

	if got := gotForm.Value["caption"]; len(got) != 1 || got[0] != "Q3" {
		t.Fatalf("caption = %#v, want [Q3]", got)
	}
	documentFiles := gotForm.File["document"]
	if len(documentFiles) != 1 {
		t.Fatalf("len(documentFiles) = %d, want 1", len(documentFiles))
	}
	if documentFiles[0].Filename != "report.pdf" {
		t.Fatalf("document filename = %q, want report.pdf", documentFiles[0].Filename)
	}
	if got := documentFiles[0].Header.Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("document Content-Type = %q, want application/octet-stream", got)
	}
	if contents := readUploadedFile(t, documentFiles[0]); contents != "document bytes" {
		t.Fatalf("document contents = %q, want document bytes", contents)
	}

	thumbnailFiles := gotForm.File["thumbnail"]
	if len(thumbnailFiles) != 1 {
		t.Fatalf("len(thumbnailFiles) = %d, want 1", len(thumbnailFiles))
	}
	if thumbnailFiles[0].Filename != "cover.jpg" {
		t.Fatalf("thumbnail filename = %q, want cover.jpg", thumbnailFiles[0].Filename)
	}
	if got := thumbnailFiles[0].Header.Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("thumbnail Content-Type = %q, want application/octet-stream", got)
	}
	if contents := readUploadedFile(t, thumbnailFiles[0]); contents != "thumbnail bytes" {
		t.Fatalf("thumbnail contents = %q, want thumbnail bytes", contents)
	}
}

func TestSendMediaLocalUploadUsesInputFileName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "generated.tmp")
	if err := os.WriteFile(path, []byte("document bytes"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var gotForm *multipart.Form
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
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

	_, err := client.SendMedia(context.Background(), "sendDocument", "document", InputFile{
		Kind:     InputFileLocalUpload,
		Value:    path,
		FileName: "report Q3.pdf",
	}, CommonParams{
		ChatID: "123",
	}, nil, nil)
	if err != nil {
		t.Fatalf("SendMedia returned error: %v", err)
	}
	if gotForm == nil {
		t.Fatal("no multipart form captured")
	}
	defer gotForm.RemoveAll()

	files := gotForm.File["document"]
	if len(files) != 1 {
		t.Fatalf("len(files) = %d, want 1", len(files))
	}
	if files[0].Filename != "report Q3.pdf" {
		t.Fatalf("filename = %q, want report Q3.pdf", files[0].Filename)
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

func readUploadedFile(t *testing.T, file *multipart.FileHeader) string {
	t.Helper()

	uploaded, err := file.Open()
	if err != nil {
		t.Fatalf("open uploaded file: %v", err)
	}
	defer uploaded.Close()

	contents, err := io.ReadAll(uploaded)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	return string(contents)
}
