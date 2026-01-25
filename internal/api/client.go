package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xbe-inc/xbe-cli/internal/telemetry"
	"github.com/xbe-inc/xbe-cli/internal/version"
)

const defaultBaseURL = "https://server.x-b-e.com"

// Package-level telemetry provider for HTTP instrumentation
var telemetryProvider *telemetry.Provider

// SetTelemetryProvider sets the telemetry provider for HTTP instrumentation.
// This should be called before creating any clients.
func SetTelemetryProvider(tp *telemetry.Provider) {
	telemetryProvider = tp
}

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

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Wrap transport with telemetry instrumentation if available
	if telemetryProvider != nil {
		httpClient.Transport = telemetryProvider.HTTPTransport(http.DefaultTransport)
	}

	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      strings.TrimSpace(token),
		HTTPClient: httpClient,
	}
}

// Get performs a GET request to the given path with query params.
func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, int, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base url: %w", err)
	}

	if query == nil {
		query = url.Values{}
	}
	ApplySparseFieldOverrides(ctx, path, query)

	path = "/" + strings.TrimLeft(path, "/")
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = query.Encode()

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

// Post performs a POST request to the given path with a JSON body.
func (c *Client) Post(ctx context.Context, path string, jsonBody []byte) ([]byte, int, error) {
	return c.doWithBody(ctx, http.MethodPost, path, jsonBody)
}

// PostWithQuery performs a POST request to the given path with query params and a JSON body.
func (c *Client) PostWithQuery(ctx context.Context, path string, query url.Values, jsonBody []byte) ([]byte, int, error) {
	return c.doWithBodyAndQuery(ctx, http.MethodPost, path, query, jsonBody)
}

// Patch performs a PATCH request to the given path with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, jsonBody []byte) ([]byte, int, error) {
	return c.doWithBody(ctx, http.MethodPatch, path, jsonBody)
}

// Delete performs a DELETE request to the given path.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, int, error) {
	return c.doWithBody(ctx, http.MethodDelete, path, nil)
}

func (c *Client) doWithBody(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	return c.doWithBodyAndQuery(ctx, method, path, nil, body)
}

func (c *Client) doWithBodyAndQuery(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, int, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid base url: %w", err)
	}

	path = "/" + strings.TrimLeft(path, "/")
	base.Path = strings.TrimRight(base.Path, "/") + path
	if query != nil {
		base.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, base.String(), bodyReader)
	if err != nil {
		return nil, 0, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}
	req.Header.Set("User-Agent", "xbe-cli/"+version.String())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return respBody, resp.StatusCode, fmt.Errorf("request failed: %s", resp.Status)
	}

	return respBody, resp.StatusCode, nil
}
