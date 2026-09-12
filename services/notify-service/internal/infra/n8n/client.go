package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

func NewClient(baseURL, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type webhookPayload struct {
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func (c *Client) DispatchWebhook(ctx context.Context, eventType string, payload interface{}) error {
	if c.baseURL == "" {
		return nil // n8n desativado localmente
	}

	url := fmt.Sprintf("%s/webhook/commerce", c.baseURL)

	body, err := json.Marshal(webhookPayload{
		Event:     eventType,
		Data:      payload,
		Timestamp: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.secret != "" {
		req.Header.Set("X-Webhook-Secret", c.secret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar com webhook n8n: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("n8n retornou status de erro: %d", resp.StatusCode)
	}

	return nil
}
