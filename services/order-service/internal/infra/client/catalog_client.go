package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CatalogHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCatalogHTTPClient(baseURL string) *CatalogHTTPClient {
	return &CatalogHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type deductRequest struct {
	Quantity int `json:"quantity"`
}

func (c *CatalogHTTPClient) DeductStock(ctx context.Context, productID string, quantity int) error {
	url := fmt.Sprintf("%s/products/%s/deduct-stock", c.baseURL, productID)

	body, err := json.Marshal(deductRequest{Quantity: quantity})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("falha ao comunicar com catalog-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog-service retornou status %d", resp.StatusCode)
	}

	return nil
}
