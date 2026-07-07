// Package telegram is a minimal client for the Telegram Bot API
// (https://core.telegram.org/bots/api), covering only the calls tg needs.
package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

// SendMessageParams are the inputs for SendMessage. ParseMode may be empty,
// "Markdown", "MarkdownV2", or "HTML".
type SendMessageParams struct {
	ChatID    string
	Text      string
	ParseMode string
}

// SendMessage posts a text message to a chat and returns the sent message.
func (c *Client) SendMessage(ctx context.Context, p SendMessageParams) (*Message, error) {
	form := url.Values{}
	form.Set("chat_id", p.ChatID)
	form.Set("text", p.Text)
	if p.ParseMode != "" {
		form.Set("parse_mode", p.ParseMode)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", apiBase, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = form.Encode()

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
