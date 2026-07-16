package payproviders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type HTTPClient struct {
	httpClient *http.Client
	secretKey  string
	baseURL    string
	name       Name
	log        *slog.Logger
}

type ClientConfig struct {
	Name       Name
	BaseURL    string
	SecretKey  string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

func NewHTTPClient(cfg ClientConfig) *HTTPClient {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPClient{
		httpClient: client,
		secretKey:  cfg.SecretKey,
		baseURL:    cfg.BaseURL,
		name:       cfg.Name,
		log:        log,
	}
}

func (c *HTTPClient) Get(ctx context.Context, endpoint string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, endpoint, nil, "")
}

func (c *HTTPClient) Post(ctx context.Context, endpoint string, payload any) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return c.doRequest(ctx, http.MethodPost, endpoint, jsonData, "application/json")
}

func (c *HTTPClient) doRequest(ctx context.Context, method, endpoint string, body []byte, contentType string) ([]byte, error) {
	url := c.baseURL + endpoint

	c.log.Debug("payment request",
		"provider", c.name,
		"method", method,
		"url", url,
	)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 200 OK and 201 Created are both treated as success.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%v API error: status %d, body: %s", c.name, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}