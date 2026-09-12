package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OrderHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOrderHTTPClient(baseURL string) *OrderHTTPClient {
	return &OrderHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (c *OrderHTTPClient) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	url := fmt.Sprintf("%s/orders/%s/status", c.baseURL, orderID)

	body, err := json.Marshal(updateStatusRequest{Status: status})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Log e não trava o fluxo caso o serviço de pedidos esteja ocupado (consistência eventual)
		return fmt.Errorf("aviso: falha ao atualizar status no order-service: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
