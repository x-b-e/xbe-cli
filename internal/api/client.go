package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xbe-inc/xbe-cli/internal/version"
)

const defaultBaseURL = "https://server.x-b-e.com"

// Client is a minimal HTTP client for the XBE API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   strings.TrimSpace(token),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs a GET request to the given path with query params.
func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, int, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base url: %w", err)
	}

	path = "/" + strings.TrimLeft(path, "/")
	base.Path = strings.TrimRight(base.Path, "/") + path
	if query != nil {
		base.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("User-Agent", "xbe-cli/"+version.String())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return body, resp.StatusCode, fmt.Errorf("request failed: %s", resp.Status)
	}

	return body, resp.StatusCode, nil
}
