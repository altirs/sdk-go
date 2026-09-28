package altirs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://altirs.ai/api"
	defaultTimeout = 30 * time.Second
)

// Client communicates with the Altirs (Guardrails-as-a-Service) API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets the HTTP client timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.http.Timeout = d
	}
}

// WithHTTPClient replaces the default http.Client entirely.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.http = hc
	}
}

// NewClient creates a Altirs API client. apiKey is required.
// Use functional options to override defaults:
//
//	client := altirs.NewClient("grd_...", altirs.WithBaseURL("http://localhost:3000"))
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Check runs guardrail checks on a single input/response pair.
func (c *Client) Check(ctx context.Context, payload CheckPayload) (*CheckResponse, error) {
	var resp CheckResponse
	if err := c.do(ctx, "/v1/guardrail/check", payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Batch runs guardrail checks on multiple input/response pairs in one request.
// Maximum 20 items per batch.
func (c *Client) Batch(ctx context.Context, items []CheckPayload) (*BatchResponse, error) {
	body := BatchPayload{Items: items}
	var resp BatchResponse
	if err := c.do(ctx, "/v1/guardrail/batch", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// IsSafe is a convenience method that returns true when the input passes all
// guardrail checks (i.e. result.Safe == true).
func (c *Client) IsSafe(ctx context.Context, input string) (bool, error) {
	res, err := c.Check(ctx, CheckPayload{Input: &input})
	if err != nil {
		return false, err
	}
	return res.Safe, nil
}

// MaskPII is a convenience method that returns the input text with PII redacted.
// If no masking was applied, the original text is returned.
func (c *Client) MaskPII(ctx context.Context, text string) (string, error) {
	mask := true
	res, err := c.Check(ctx, CheckPayload{
		Input:  &text,
		Checks: []CheckName{CheckPII},
		Options: &CheckOptions{
			PII: &PiiOptions{Mask: &mask},
		},
	})
	if err != nil {
		return "", err
	}
	if res.MaskedInput != nil {
		return *res.MaskedInput, nil
	}
	return text, nil
}

// do executes a POST request and decodes the JSON response into dst.
func (c *Client) do(ctx context.Context, path string, body any, dst any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("altirs: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("altirs: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("X-SDK-Version", "go/"+sdkVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return &AltirsError{
				Message:   "request cancelled or timed out",
				Status:    408,
				RequestID: "timeout",
			}
		}
		return &AltirsError{
			Message:   err.Error(),
			Status:    0,
			RequestID: "unknown",
		}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("altirs: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(data, &errResp); jsonErr != nil {
			return &AltirsError{
				Message:   fmt.Sprintf("request failed with status %d", resp.StatusCode),
				Status:    resp.StatusCode,
				RequestID: "unknown",
			}
		}
		msg := errResp.Error
		if msg == "" {
			msg = fmt.Sprintf("request failed with status %d", resp.StatusCode)
		}
		rid := errResp.RequestID
		if rid == "" {
			rid = "unknown"
		}
		return &AltirsError{
			Message:   msg,
			Status:    resp.StatusCode,
			RequestID: rid,
		}
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("altirs: failed to decode response: %w", err)
	}
	return nil
}
