package telegram

import (
	"context"
	"encoding/json"
	"errors"
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

func TestSendMediaGroupLocalUploadsUseMultipartAttachNames(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(firstPath, []byte("first photo"), 0o600); err != nil {
		t.Fatalf("write first photo: %v", err)
	}
	secondPath := filepath.Join(dir, "b.jpg")
	if err := os.WriteFile(secondPath, []byte("second photo"), 0o600); err != nil {
		t.Fatalf("write second photo: %v", err)
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
		return okMessagesResponse(), nil
	})

	msgs, err := client.SendMediaGroup(context.Background(), CommonParams{
		ChatID:    "123",
		ParseMode: "HTML",
	}, "photo", "album caption", []InputFile{
		{Kind: InputFileLocalUpload, Value: firstPath},
		{Kind: InputFileLocalUpload, Value: secondPath},
	})
	if err != nil {
		t.Fatalf("SendMediaGroup returned error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len(msgs) = %d, want 2", len(msgs))
	}
	if msgs[0].MessageID != 100 || msgs[1].MessageID != 101 {
		t.Fatalf("message ids = [%d %d], want [100 101]", msgs[0].MessageID, msgs[1].MessageID)
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.URL.Path != "/botTOKEN/sendMediaGroup" {
		t.Fatalf("URL path = %q, want /botTOKEN/sendMediaGroup", gotReq.URL.Path)
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
		"chat_id":    "123",
		"parse_mode": "HTML",
	}
	for name, value := range wantValues {
		got := gotForm.Value[name]
		if len(got) != 1 || got[0] != value {
			t.Fatalf("%s = %#v, want [%q]", name, got, value)
		}
	}
	media := parseMediaGroupJSON(t, gotForm.Value["media"])
	wantMedia := []inputMediaItem{
		{Type: "photo", Media: "attach://file0", Caption: "album caption"},
		{Type: "photo", Media: "attach://file1"},
	}
	assertMediaGroupItems(t, media, wantMedia)

	firstFiles := gotForm.File["file0"]
	if len(firstFiles) != 1 {
		t.Fatalf("len(file0) = %d, want 1", len(firstFiles))
	}
	if firstFiles[0].Filename != "a.jpg" {
		t.Fatalf("file0 filename = %q, want a.jpg", firstFiles[0].Filename)
	}
	if contents := readUploadedFile(t, firstFiles[0]); contents != "first photo" {
		t.Fatalf("file0 contents = %q, want first photo", contents)
	}

	secondFiles := gotForm.File["file1"]
	if len(secondFiles) != 1 {
		t.Fatalf("len(file1) = %d, want 1", len(secondFiles))
	}
	if secondFiles[0].Filename != "b.jpg" {
		t.Fatalf("file1 filename = %q, want b.jpg", secondFiles[0].Filename)
	}
	if contents := readUploadedFile(t, secondFiles[0]); contents != "second photo" {
		t.Fatalf("file1 contents = %q, want second photo", contents)
	}
}

func TestSendMediaGroupNonLocalFilesUseForm(t *testing.T) {
	var gotReq *http.Request
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotReq = req
		return okMessagesResponse(), nil
	})

	msgs, err := client.SendMediaGroup(context.Background(), CommonParams{
		ChatID:              "123",
		DisableNotification: true,
	}, "document", "docs", []InputFile{
		{Kind: InputFileRemoteURL, Value: "https://example.com/a.pdf"},
		{Kind: InputFileFileID, Value: "BQACAgIAAx0FakeFileID"},
	})
	if err != nil {
		t.Fatalf("SendMediaGroup returned error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len(msgs) = %d, want 2", len(msgs))
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.Method != http.MethodPost {
		t.Fatalf("Method = %q, want %q", gotReq.Method, http.MethodPost)
	}
	if gotReq.URL.Path != "/botTOKEN/sendMediaGroup" {
		t.Fatalf("URL path = %q, want /botTOKEN/sendMediaGroup", gotReq.URL.Path)
	}
	if gotReq.Body != nil {
		t.Fatal("Body is non-nil, want nil")
	}
	if gotReq.Header.Get("Content-Type") != "" {
		t.Fatalf("Content-Type = %q, want empty", gotReq.Header.Get("Content-Type"))
	}

	gotQuery := gotReq.URL.Query()
	if gotQuery.Get("chat_id") != "123" {
		t.Fatalf("chat_id = %q, want 123", gotQuery.Get("chat_id"))
	}
	if gotQuery.Get("disable_notification") != "true" {
		t.Fatalf("disable_notification = %q, want true", gotQuery.Get("disable_notification"))
	}
	if gotQuery.Get("file0") != "" || gotQuery.Get("file1") != "" {
		t.Fatalf("file query params = [%q %q], want none", gotQuery.Get("file0"), gotQuery.Get("file1"))
	}
	media := parseMediaGroupJSON(t, []string{gotQuery.Get("media")})
	wantMedia := []inputMediaItem{
		{Type: "document", Media: "https://example.com/a.pdf", Caption: "docs"},
		{Type: "document", Media: "BQACAgIAAx0FakeFileID"},
	}
	assertMediaGroupItems(t, media, wantMedia)
}

func TestSendMediaGroupMixedFilesUploadOnlyLocalItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "b.jpg")
	if err := os.WriteFile(path, []byte("local photo"), 0o600); err != nil {
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
		return okMessagesResponse(), nil
	})

	msgs, err := client.SendMediaGroup(context.Background(), CommonParams{
		ChatID: "123",
	}, "photo", "mixed album", []InputFile{
		{Kind: InputFileRemoteURL, Value: "https://example.com/a.jpg"},
		{Kind: InputFileLocalUpload, Value: path},
		{Kind: InputFileFileID, Value: "AgACAgIAAx0FakeFileID"},
	})
	if err != nil {
		t.Fatalf("SendMediaGroup returned error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len(msgs) = %d, want 2", len(msgs))
	}
	if gotForm == nil {
		t.Fatal("no multipart form captured")
	}
	defer gotForm.RemoveAll()

	media := parseMediaGroupJSON(t, gotForm.Value["media"])
	wantMedia := []inputMediaItem{
		{Type: "photo", Media: "https://example.com/a.jpg", Caption: "mixed album"},
		{Type: "photo", Media: "attach://file1"},
		{Type: "photo", Media: "AgACAgIAAx0FakeFileID"},
	}
	assertMediaGroupItems(t, media, wantMedia)

	if len(gotForm.File) != 1 {
		t.Fatalf("multipart file fields = %#v, want only file1", gotForm.File)
	}
	if _, ok := gotForm.File["file0"]; ok {
		t.Fatal("file0 was uploaded, want remote URL to skip multipart")
	}
	if _, ok := gotForm.File["file2"]; ok {
		t.Fatal("file2 was uploaded, want file_id to skip multipart")
	}
	files := gotForm.File["file1"]
	if len(files) != 1 {
		t.Fatalf("len(file1) = %d, want 1", len(files))
	}
	if files[0].Filename != "b.jpg" {
		t.Fatalf("file1 filename = %q, want b.jpg", files[0].Filename)
	}
	if contents := readUploadedFile(t, files[0]); contents != "local photo" {
		t.Fatalf("file1 contents = %q, want local photo", contents)
	}
}

func TestSendPaidMediaLocalUploadUsesMultipartAndDecodesSingleMessage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preview.jpg")
	if err := os.WriteFile(path, []byte("paid photo"), 0o600); err != nil {
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

	msg, err := client.SendPaidMedia(context.Background(), CommonParams{
		ChatID:    "123",
		ParseMode: "HTML",
	}, 50, "photo", []InputFile{
		{Kind: InputFileLocalUpload, Value: path},
	}, map[string]string{
		"caption": "paid caption",
	})
	if err != nil {
		t.Fatalf("SendPaidMedia returned error: %v", err)
	}
	if msg.MessageID != 42 {
		t.Fatalf("MessageID = %d, want 42", msg.MessageID)
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.URL.Path != "/botTOKEN/sendPaidMedia" {
		t.Fatalf("URL path = %q, want /botTOKEN/sendPaidMedia", gotReq.URL.Path)
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
		"chat_id":    "123",
		"parse_mode": "HTML",
		"star_count": "50",
		"caption":    "paid caption",
	}
	for name, value := range wantValues {
		got := gotForm.Value[name]
		if len(got) != 1 || got[0] != value {
			t.Fatalf("%s = %#v, want [%q]", name, got, value)
		}
	}
	media := parseMediaGroupJSON(t, gotForm.Value["media"])
	wantMedia := []inputMediaItem{
		{Type: "photo", Media: "attach://file0"},
	}
	assertMediaGroupItems(t, media, wantMedia)

	files := gotForm.File["file0"]
	if len(files) != 1 {
		t.Fatalf("len(file0) = %d, want 1", len(files))
	}
	if files[0].Filename != "preview.jpg" {
		t.Fatalf("file0 filename = %q, want preview.jpg", files[0].Filename)
	}
	if contents := readUploadedFile(t, files[0]); contents != "paid photo" {
		t.Fatalf("file0 contents = %q, want paid photo", contents)
	}
}

func TestGetUpdatesSendsOffsetAndTimeout(t *testing.T) {
	var gotReq *http.Request
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotReq = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok": true, "result": []}`)),
			Header:     make(http.Header),
		}, nil
	})

	updates, err := client.GetUpdates(context.Background(), 55, 25)
	if err != nil {
		t.Fatalf("GetUpdates returned error: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("updates = %#v, want empty", updates)
	}
	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.URL.Path != "/botTOKEN/getUpdates" {
		t.Fatalf("URL path = %q, want /botTOKEN/getUpdates", gotReq.URL.Path)
	}
	query := gotReq.URL.Query()
	if query.Get("offset") != "55" {
		t.Fatalf("offset = %q, want 55", query.Get("offset"))
	}
	if query.Get("timeout") != "25" {
		t.Fatalf("timeout = %q, want 25", query.Get("timeout"))
	}
}

func TestGetUpdatesParsesMessages(t *testing.T) {
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"ok": true,
				"result": [
					{
						"update_id": 900,
						"message": {
							"message_id": 7,
							"date": 1234567890,
							"text": "hi there",
							"chat": {"id": 123},
							"from": {"username": "ada", "first_name": "Ada"}
						}
					}
				]
			}`)),
			Header: make(http.Header),
		}, nil
	})

	updates, err := client.GetUpdates(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("GetUpdates returned error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("len(updates) = %d, want 1", len(updates))
	}
	if updates[0].UpdateID != 900 {
		t.Fatalf("UpdateID = %d, want 900", updates[0].UpdateID)
	}
	if updates[0].Message == nil {
		t.Fatal("Message = nil, want set")
	}
	if updates[0].Message.Text != "hi there" {
		t.Fatalf("Text = %q, want %q", updates[0].Message.Text, "hi there")
	}
	if updates[0].Message.Chat.ID != 123 {
		t.Fatalf("Chat.ID = %d, want 123", updates[0].Message.Chat.ID)
	}
	if updates[0].Message.From.Username != "ada" {
		t.Fatalf("From.Username = %q, want %q", updates[0].Message.From.Username, "ada")
	}
}

func TestGetUpdatesPropagatesAPIError(t *testing.T) {
	client := NewClient("TOKEN")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok": false, "error_code": 409, "description": "conflict"}`)),
			Header:     make(http.Header),
		}, nil
	})

	_, err := client.GetUpdates(context.Background(), 0, 0)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %v, want *APIError", err, err)
	}
	if apiErr.Code != 409 {
		t.Fatalf("Code = %d, want 409", apiErr.Code)
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

func okMessagesResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
			"ok": true,
			"result": [
				{
					"message_id": 100,
					"date": 1234567890,
					"chat": {"id": 123}
				},
				{
					"message_id": 101,
					"date": 1234567891,
					"chat": {"id": 123}
				}
			]
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

func parseMediaGroupJSON(t *testing.T, values []string) []inputMediaItem {
	t.Helper()

	if len(values) != 1 {
		t.Fatalf("media values = %#v, want exactly one", values)
	}
	var media []inputMediaItem
	if err := json.Unmarshal([]byte(values[0]), &media); err != nil {
		t.Fatalf("unmarshal media JSON %q: %v", values[0], err)
	}
	return media
}

func assertMediaGroupItems(t *testing.T, got []inputMediaItem, want []inputMediaItem) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(media) = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("media[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
