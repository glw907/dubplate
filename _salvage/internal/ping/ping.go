// Package ping is a provider-agnostic client for a dead-man's-switch
// monitoring service: given a check's URL, it reports success or failure.
// It carries no slug or provider knowledge; that lives in
// [internal/config].
package ping

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client pings a monitoring URL, retrying transient failures within
// Timeout per attempt.
type Client struct {
	Timeout time.Duration
	Retries int
}

// NewClient returns a Client with the shell layer's curl budget: 3
// retries, a 10-second timeout per attempt.
func NewClient() *Client {
	return &Client{Timeout: 10 * time.Second, Retries: 3}
}

// Success reports a successful check run by GETting url. It is a no-op
// returning nil when url is empty, the no-Worker adopter case.
func (c *Client) Success(ctx context.Context, url string) error {
	if url == "" {
		return nil
	}
	return c.do(ctx, http.MethodGet, url)
}

// Fail reports a failed check run by POSTing to url with "/fail"
// appended (after stripping any trailing slash from url). It is a no-op
// returning nil when url is empty.
func (c *Client) Fail(ctx context.Context, url string) error {
	if url == "" {
		return nil
	}
	return c.do(ctx, http.MethodPost, strings.TrimSuffix(url, "/")+"/fail")
}

// do sends method to url, retrying up to c.Retries additional times on a
// request error or a >=400 status. It gives up and returns the last
// error once attempts are exhausted.
func (c *Client) do(ctx context.Context, method, url string) error {
	client := &http.Client{Timeout: c.Timeout}

	var lastErr error
	attempts := c.Retries + 1
	for range attempts {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("build %s request to %s: %w", method, url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		// Drain before closing so the connection can be reused by a
		// retry or a later ping.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 400 {
			return nil
		}
		lastErr = fmt.Errorf("ping %s: status %s", url, resp.Status)
	}
	return fmt.Errorf("ping %s failed after %d attempts: %w", url, attempts, lastErr)
}
