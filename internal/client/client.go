package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultURL = "https://console.cloud.mtn.ng"

// DefaultMaxRetries is the number of additional attempts made for retryable
// failures when Config.MaxRetries is not set.
const DefaultMaxRetries = 3

type Config struct {
	URL      string
	Token    string
	Username string
	Password string
	Timeout  time.Duration
	Insecure bool
	// MaxRetries is the number of retries for transient failures (429/5xx,
	// network errors). When zero, DefaultMaxRetries is used; use a negative
	// value to disable retries.
	MaxRetries int
}

type Client struct {
	baseURL    string
	apiURL     string
	token      string
	username   string
	password   string
	httpClient *http.Client
	maxRetries int
}

func New(config Config) (*Client, error) {
	baseURL := strings.TrimRight(config.URL, "/")
	if baseURL == "" {
		baseURL = DefaultURL
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	maxRetries := config.MaxRetries
	switch {
	case maxRetries == 0:
		maxRetries = DefaultMaxRetries
	case maxRetries < 0:
		maxRetries = 0
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if config.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &Client{
		baseURL:  baseURL,
		apiURL:   baseURL + "/api",
		token:    config.Token,
		username: config.Username,
		password: config.Password,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		maxRetries: maxRetries,
	}, nil
}

func (c *Client) Authenticate(ctx context.Context) error {
	if c.token != "" {
		return nil
	}
	if c.username == "" || c.password == "" {
		return &APIError{Message: "provide either token or username/password"}
	}

	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)

	u, err := url.Parse(c.baseURL + "/oauth/token")
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("grant_type", "password")
	q.Set("scope", "write")
	q.Set("client_id", "morph-cli")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := decodeResponse(resp, &body); err != nil {
		return err
	}
	token, _ := body["access_token"].(string)
	if token == "" {
		return &APIError{StatusCode: resp.StatusCode, Message: "authentication response did not include access_token", Body: body}
	}
	c.token = token
	return nil
}

func (c *Client) get(ctx context.Context, path string, query map[string]string, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}

func (c *Client) post(ctx context.Context, path string, payload any, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, payload, out)
}

func (c *Client) put(ctx context.Context, path string, payload any, out any) error {
	return c.do(ctx, http.MethodPut, path, nil, payload, out)
}

func (c *Client) delete(ctx context.Context, path string, query map[string]string) error {
	var out map[string]any
	return c.do(ctx, http.MethodDelete, path, query, nil, &out)
}

func (c *Client) do(ctx context.Context, method, path string, query map[string]string, payload any, out any) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	u, err := url.Parse(c.apiURL + "/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return err
	}
	q := u.Query()
	for key, value := range query {
		if value != "" {
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()

	var payloadBytes []byte
	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	var retryAfter time.Duration
	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffDelay(attempt, retryAfter)):
			}
		}

		var body io.Reader
		if payloadBytes != nil {
			body = bytes.NewReader(payloadBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "terraform-provider-mtncloud")
		req.Header.Set("Authorization", "Bearer "+c.token)
		if payloadBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network/transport error: retry only idempotent reads.
			if attempt < c.maxRetries && method == http.MethodGet {
				continue
			}
			return err
		}

		if attempt < c.maxRetries && shouldRetry(method, resp.StatusCode) {
			retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			_ = resp.Body.Close()
			continue
		}

		defer resp.Body.Close()
		return decodeResponse(resp, out)
	}
}

// shouldRetry reports whether a response status warrants a retry. 429 is always
// retried (the request was not processed); 5xx and transport errors are retried
// only for idempotent GETs to avoid duplicate writes.
func shouldRetry(method string, statusCode int) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode >= 500 && statusCode <= 599 && method == http.MethodGet {
		return true
	}
	return false
}

// backoffDelay returns an exponential backoff with jitter, honoring a server
// Retry-After hint when present. Capped at 5s.
func backoffDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	const base = 300 * time.Millisecond
	const max = 5 * time.Second
	delay := base << (attempt - 1) // attempt starts at 1 for the first retry
	if delay > max || delay <= 0 {
		delay = max
	}
	// Full jitter in [delay/2, delay].
	jitter := delay/2 + time.Duration(rand.Int63n(int64(delay/2)+1))
	return jitter
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(header)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return 0
}

func decodeResponse(resp *http.Response, out any) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var body map[string]any
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			body = map[string]any{"raw": string(bodyBytes)}
		}
	} else {
		body = map[string]any{}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    extractMessage(body),
			Body:       body,
		}
	}

	if out == nil {
		return nil
	}
	if len(bodyBytes) == 0 {
		return nil
	}
	return json.Unmarshal(bodyBytes, out)
}

func extractMessage(body map[string]any) string {
	for _, key := range []string{"message", "msg", "error", "error_description"} {
		if value, ok := body[key].(string); ok && value != "" {
			return value
		}
	}
	if errorsValue, ok := body["errors"]; ok {
		return fmt.Sprint(errorsValue)
	}
	if len(body) > 0 {
		return fmt.Sprint(body)
	}
	return "unknown error"
}
