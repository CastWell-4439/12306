package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrHoldNotFound = errors.New("inventory hold not found")

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type TryHoldInput struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
	Qty          int    `json:"qty"`
	Capacity     int    `json:"capacity"`
}

type ConfirmInput struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
}

type ReleaseInput struct {
	PartitionKey string `json:"partition_key"`
	HoldID       string `json:"hold_id"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) TryHold(ctx context.Context, in TryHoldInput) error {
	return c.post(ctx, "/inventory/try-hold", in)
}

func (c *Client) ConfirmHold(ctx context.Context, in ConfirmInput) error {
	return c.post(ctx, "/inventory/confirm-hold", in)
}

func (c *Client) ReleaseHold(ctx context.Context, in ReleaseInput) error {
	return c.post(ctx, "/inventory/release-hold", in)
}

func (c *Client) post(ctx context.Context, path string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := extractErrorMessage(body)
	if resp.StatusCode == http.StatusNotFound && strings.Contains(strings.ToLower(msg), "hold not found") {
		return ErrHoldNotFound
	}
	if msg == "" {
		msg = string(body)
	}
	return fmt.Errorf("inventory %s failed: status=%d err=%s", path, resp.StatusCode, strings.TrimSpace(msg))
}

func extractErrorMessage(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	msg, _ := payload["error"].(string)
	return msg
}
