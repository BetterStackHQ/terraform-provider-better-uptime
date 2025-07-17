package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func rateLimitRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	if resp.StatusCode == 429 {
		return true, nil
	}

	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

type client struct {
	baseURL     string
	token       string
	retryClient *retryablehttp.Client
	userAgent   string
}

type ClientConfig struct {
	BaseURL      string
	Token        string
	UserAgent    string
	HTTPClient   *http.Client
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

func newClient(config ClientConfig) (*client, error) {
	// Set reasonable bounds for max retries
	if config.RetryMax < 0 || config.RetryMax > 10 {
		config.RetryMax = 10
	}
	// Set default wait times
	if config.RetryWaitMin == 0 {
		config.RetryWaitMin = 1 * time.Second
	}
	if config.RetryWaitMax == 0 {
		config.RetryWaitMax = 30 * time.Second
	}

	// Create retry client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = config.RetryMax
	retryClient.RetryWaitMin = config.RetryWaitMin
	retryClient.RetryWaitMax = config.RetryWaitMax
	retryClient.CheckRetry = rateLimitRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	// Use custom HTTP client if provided
	if config.HTTPClient != nil {
		retryClient.HTTPClient = config.HTTPClient
	}

	// Disable default logging
	retryClient.RequestLogHook = nil
	retryClient.ResponseLogHook = nil
	retryClient.ErrorHandler = nil

	return &client{
		baseURL:     config.BaseURL,
		token:       config.Token,
		retryClient: retryClient,
		userAgent:   config.UserAgent,
	}, nil
}

func (c *client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *client) Patch(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPatch, path, body)
}

func (c *client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func (c *client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(method, fmt.Sprintf("%s%s", c.baseURL, path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if method == http.MethodPost || method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.retryClient.Do(req.WithContext(ctx))
}
