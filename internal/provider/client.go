package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context/ctxhttp"
)

type client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	userAgent  string
}

type option func(c *client)

func withHTTPClient(httpClient *http.Client) option {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

func withUserAgent(userAgent string) option {
	return func(c *client) {
		c.userAgent = userAgent
	}
}

func newClient(baseURL, token string, opts ...option) (*client, error) {
	c := client{
		baseURL:    baseURL,
		token:      token,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c, nil
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
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.baseURL, path), body)
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
	return ctxhttp.Do(ctx, c.httpClient, req)
}
