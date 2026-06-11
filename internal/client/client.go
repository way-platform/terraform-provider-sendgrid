package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Client is a thin wrapper around the SendGrid v3 API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the default SendGrid API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient injects a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New creates a SendGrid API client.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: "https://api.sendgrid.com",
		httpClient: &http.Client{
			Transport: &retryTransport{
				maxRetries: 3,
				next:       http.DefaultTransport,
			},
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Client) do(
	ctx context.Context,
	method, path string,
	body io.Reader,
	contentType string,
) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, reqBody, respBody any) error {
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		body = bytes.NewReader(b)
	}

	resp, err := c.do(ctx, method, path, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	if respBody != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// APIError represents an error response from the SendGrid API.
type APIError struct {
	StatusCode int
	Errors     []apiErrorDetail `json:"errors"`
}

type apiErrorDetail struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Help    string `json:"help"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("sendgrid api error (status %d): %s", e.StatusCode, e.Errors[0].Message)
	}
	return fmt.Sprintf("sendgrid api error (status %d)", e.StatusCode)
}

func parseErrorResponse(resp *http.Response) error {
	apiErr := &APIError{StatusCode: resp.StatusCode}
	if err := json.NewDecoder(resp.Body).Decode(apiErr); err != nil {
		return &APIError{StatusCode: resp.StatusCode}
	}
	return apiErr
}

// IsNotFound returns true if the error is a 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
