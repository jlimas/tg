// Package telegram is a minimal client for the Telegram Bot API
// (https://core.telegram.org/bots/api), covering only the calls tg needs.
package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.telegram.org"

// Client talks to the Telegram Bot API using a bot token.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient builds a Client for the given bot token.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Message is the subset of Telegram's Message object tg cares about.
type Message struct {
	MessageID int   `json:"message_id"`
	Date      int64 `json:"date"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

// APIError is returned when Telegram responds with ok:false. Code and
// Description come directly from Telegram's response.
type APIError struct {
	Code        int
	Description string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegram API error %d: %s", e.Code, e.Description)
}

type apiResponse[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

// InputFileKind describes how a Telegram media argument should be sent.
type InputFileKind string

const (
	InputFileLocalUpload InputFileKind = "local_upload"
	InputFileRemoteURL   InputFileKind = "remote_url"
	InputFileFileID      InputFileKind = "file_id"
)

// InputFile is a Telegram media argument resolved to a local upload, remote
// URL, or file_id-like string.
type InputFile struct {
	Kind     InputFileKind
	Value    string
	FileName string
}

// ResolveInputFile classifies s using Telegram's InputFile conventions.
func ResolveInputFile(s string) InputFile {
	if _, err := os.Stat(s); err == nil {
		return InputFile{Kind: InputFileLocalUpload, Value: s}
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return InputFile{Kind: InputFileRemoteURL, Value: s}
	}
	return InputFile{Kind: InputFileFileID, Value: s}
}

// CommonParams are fields shared by most Telegram send methods.
type CommonParams struct {
	ChatID              string
	ParseMode           string
	ReplyToMessageID    int
	DisableNotification bool
	ProtectContent      bool
	MessageThreadID     int
}

func (p CommonParams) toParams() map[string]string {
	params := map[string]string{
		"chat_id": p.ChatID,
	}
	if p.ParseMode != "" {
		params["parse_mode"] = p.ParseMode
	}
	if p.ReplyToMessageID != 0 {
		replyParameters, _ := json.Marshal(struct {
			MessageID int `json:"message_id"`
		}{
			MessageID: p.ReplyToMessageID,
		})
		params["reply_parameters"] = string(replyParameters)
	}
	if p.DisableNotification {
		params["disable_notification"] = "true"
	}
	if p.ProtectContent {
		params["protect_content"] = "true"
	}
	if p.MessageThreadID != 0 {
		params["message_thread_id"] = strconv.Itoa(p.MessageThreadID)
	}
	return params
}

// SendMessageParams are the inputs for SendMessage. ParseMode may be empty,
// "Markdown", "MarkdownV2", or "HTML".
type SendMessageParams struct {
	ChatID              string
	Text                string
	ParseMode           string
	ReplyToMessageID    int
	DisableNotification bool
	ProtectContent      bool
	MessageThreadID     int
}

// SendMessage posts a text message to a chat and returns the sent message.
func (c *Client) SendMessage(ctx context.Context, p SendMessageParams) (*Message, error) {
	params := CommonParams{
		ChatID:              p.ChatID,
		ParseMode:           p.ParseMode,
		ReplyToMessageID:    p.ReplyToMessageID,
		DisableNotification: p.DisableNotification,
		ProtectContent:      p.ProtectContent,
		MessageThreadID:     p.MessageThreadID,
	}.toParams()
	params["text"] = p.Text

	return c.call(ctx, "sendMessage", params, nil)
}

// Send posts a non-media message using method for the verb-specific Telegram
// Bot API call.
func (c *Client) Send(ctx context.Context, method string, common CommonParams, extra map[string]string) (*Message, error) {
	params := common.toParams()
	for name, value := range extra {
		params[name] = value
	}
	return c.call(ctx, method, params, nil)
}

// SendMedia posts a media message using method and field for the verb-specific
// Telegram Bot API call.
func (c *Client) SendMedia(ctx context.Context, method string, field string, file InputFile, common CommonParams, extra map[string]string, extraFiles map[string]InputFile) (*Message, error) {
	params := common.toParams()
	for name, value := range extra {
		params[name] = value
	}
	files := map[string]InputFile{
		field: file,
	}
	for name, file := range extraFiles {
		files[name] = file
	}
	return c.call(ctx, method, params, files)
}

func (c *Client) call(ctx context.Context, method string, params map[string]string, files map[string]InputFile) (*Message, error) {
	formParams := make(map[string]string, len(params)+len(files))
	for name, value := range params {
		formParams[name] = value
	}

	localFiles := make(map[string]InputFile)
	for name, file := range files {
		if file.Kind == InputFileLocalUpload {
			localFiles[name] = file
			continue
		}
		formParams[name] = file.Value
	}

	endpoint := fmt.Sprintf("%s/bot%s/%s", apiBase, c.token, method)
	if len(localFiles) == 0 {
		return c.callForm(ctx, endpoint, formParams)
	}
	return c.callMultipart(ctx, endpoint, formParams, localFiles)
}

func (c *Client) callForm(ctx context.Context, endpoint string, params map[string]string) (*Message, error) {
	form := url.Values{}
	for name, value := range params {
		form.Set(name, value)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = form.Encode()

	return c.do(req)
}

func (c *Client) callMultipart(ctx context.Context, endpoint string, params map[string]string, files map[string]InputFile) (*Message, error) {
	bodyReader, bodyWriter := io.Pipe()
	writer := multipart.NewWriter(bodyWriter)
	go func() {
		var err error
		defer func() {
			if closeErr := writer.Close(); err == nil {
				err = closeErr
			}
			_ = bodyWriter.CloseWithError(err)
		}()

		for name, value := range params {
			if err = writer.WriteField(name, value); err != nil {
				return
			}
		}
		for name, file := range files {
			if err = addInputFilePart(writer, name, file); err != nil {
				return
			}
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bodyReader)
	if err != nil {
		_ = bodyReader.Close()
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return c.do(req)
}

func addInputFilePart(writer *multipart.Writer, fieldName string, file InputFile) error {
	f, err := os.Open(file.Value)
	if err != nil {
		return err
	}
	defer f.Close()

	fileName := file.FileName
	if fileName == "" {
		fileName = filepath.Base(file.Value)
	}
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	return err
}

func (c *Client) do(req *http.Request) (*Message, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("contacting telegram: %w", err)
	}
	defer resp.Body.Close()

	var parsed apiResponse[Message]
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decoding telegram response: %w", err)
	}
	if !parsed.OK {
		return nil, &APIError{Code: parsed.ErrorCode, Description: parsed.Description}
	}
	return &parsed.Result, nil
}
