package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultUserAgent    = "code-platform/1.0 (+https://code.gouv.aor)"
	defaultMaxAttempts  = 3
	defaultBaseDelay    = 250 * time.Millisecond
	defaultHTTPTimeout  = 15 * time.Second
	maxErrorBodySnippet = 2048
)

// RateLimit is the normalised rate-limit state returned by the providers.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset,omitzero"`
}

// Exhausted reports whether the quota is currently consumed.
func (r RateLimit) Exhausted() bool { return r.Limit > 0 && r.Remaining <= 0 }

// Page is the normalised pagination state of a provider listing.
type Page struct {
	NextPage int    `json:"nextPage,omitempty"`
	NextURL  string `json:"nextUrl,omitempty"`
	HasMore  bool   `json:"hasMore"`
}

// Request describes one provider call.
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Body    any
	Accept  string
	Headers map[string]string
}

// Response is the metadata attached to a successful provider call.
type Response struct {
	Status    int
	Page      Page
	RateLimit RateLimit
	Header    http.Header
}

// ClientConfig configures a Client.
type ClientConfig struct {
	Provider   string
	BaseURL    string
	Token      string
	UserAgent  string
	Timeout    time.Duration
	HTTPClient *http.Client
	// MaxAttempts counts the initial call; <= 0 falls back to the default.
	MaxAttempts int
	// BaseDelay is the first retry delay; it doubles on each attempt.
	BaseDelay time.Duration
	// Sleep is injected by tests to make retry timing deterministic.
	Sleep func(ctx context.Context, delay time.Duration) error
}

// Client is the shared HTTP transport of every adapter. It normalises base
// URLs, authentication, pagination, rate limits, retries and error mapping so
// that adapters only have to translate payloads.
type Client struct {
	provider    string
	baseURL     string
	token       string
	userAgent   string
	httpClient  *http.Client
	maxAttempts int
	baseDelay   time.Duration
	sleep       func(ctx context.Context, delay time.Duration) error
}

// NewClient builds a provider client from a configuration.
func NewClient(cfg ClientConfig) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultHTTPTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	attempts := cfg.MaxAttempts
	if attempts <= 0 {
		attempts = defaultMaxAttempts
	}
	baseDelay := cfg.BaseDelay
	if baseDelay <= 0 {
		baseDelay = defaultBaseDelay
	}
	sleep := cfg.Sleep
	if sleep == nil {
		sleep = sleepWithContext
	}
	return &Client{
		provider:    cfg.Provider,
		baseURL:     strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		token:       cfg.Token,
		userAgent:   userAgent,
		httpClient:  httpClient,
		maxAttempts: attempts,
		baseDelay:   baseDelay,
		sleep:       sleep,
	}
}

// BaseURL exposes the normalised base URL (useful for adapters building links).
func (c *Client) BaseURL() string { return c.baseURL }

// Do performs a provider call and decodes the JSON payload into out.
func (c *Client) Do(ctx context.Context, request Request, out any) (*Response, error) {
	body, response, err := c.do(ctx, request)
	if err != nil {
		return response, err
	}
	if out == nil || len(bytes.TrimSpace(body)) == 0 {
		return response, nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return response, &Error{
			Kind:      ErrorUnknown,
			Provider:  c.provider,
			Operation: request.operation(),
			Message:   "the provider returned a payload that could not be decoded",
			Status:    response.Status,
			Cause:     err,
		}
	}
	return response, nil
}

// DoRaw performs a provider call and returns the raw response body.
func (c *Client) DoRaw(ctx context.Context, request Request) ([]byte, *Response, error) {
	return c.do(ctx, request)
}

func (c *Client) do(ctx context.Context, request Request) ([]byte, *Response, error) {
	if c.baseURL == "" {
		return nil, nil, &Error{
			Kind:      ErrorUnavailable,
			Provider:  c.provider,
			Operation: request.operation(),
			Message:   "the provider base URL is not configured",
		}
	}

	endpoint, err := c.endpoint(request)
	if err != nil {
		return nil, nil, &Error{
			Kind:      ErrorInvalid,
			Provider:  c.provider,
			Operation: request.operation(),
			Message:   "the provider request path is invalid",
			Cause:     err,
		}
	}

	var payload []byte
	if request.Body != nil {
		payload, err = json.Marshal(request.Body)
		if err != nil {
			return nil, nil, &Error{
				Kind:      ErrorInvalid,
				Provider:  c.provider,
				Operation: request.operation(),
				Message:   "the provider request body could not be encoded",
				Cause:     err,
			}
		}
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		httpRequest, buildErr := c.buildRequest(ctx, request, endpoint, payload)
		if buildErr != nil {
			return nil, nil, buildErr
		}

		response, doErr := c.httpClient.Do(httpRequest)
		if doErr != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, nil, &Error{
					Kind:      ErrorUnavailable,
					Provider:  c.provider,
					Operation: request.operation(),
					Message:   "the provider request was cancelled",
					Cause:     ctxErr,
				}
			}
			lastErr = &Error{
				Kind:      ErrorUnavailable,
				Provider:  c.provider,
				Operation: request.operation(),
				Message:   "the provider could not be reached",
				Cause:     doErr,
			}
			if attempt < c.maxAttempts {
				if sleepErr := c.sleep(ctx, c.backoff(attempt, 0)); sleepErr != nil {
					return nil, nil, lastErr
				}
				continue
			}
			return nil, nil, lastErr
		}

		body, readErr := io.ReadAll(io.LimitReader(response.Body, 8<<20))
		_ = response.Body.Close()
		if readErr != nil {
			return nil, nil, &Error{
				Kind:      ErrorUnavailable,
				Provider:  c.provider,
				Operation: request.operation(),
				Message:   "the provider response could not be read",
				Cause:     readErr,
			}
		}

		meta := &Response{
			Status:    response.StatusCode,
			Header:    response.Header,
			Page:      parsePage(response.Header),
			RateLimit: parseRateLimit(response.Header),
		}

		switch {
		case response.StatusCode >= 200 && response.StatusCode < 300:
			return body, meta, nil
		case response.StatusCode == http.StatusNotFound:
			return body, meta, c.httpError(ErrorNotFound, request, response.StatusCode, body)
		case response.StatusCode == http.StatusUnauthorized:
			return body, meta, c.httpError(ErrorUnauthorized, request, response.StatusCode, body)
		case response.StatusCode == http.StatusForbidden:
			if meta.RateLimit.Exhausted() {
				rateErr := c.httpError(ErrorRateLimited, request, response.StatusCode, body)
				rateErr.RetryAfter = retryAfter(response.Header, meta.RateLimit)
				return body, meta, rateErr
			}
			// 403 also means "token without the required scope": this is an
			// authorization problem, and retrying would never help.
			return body, meta, c.httpError(ErrorUnauthorized, request, response.StatusCode, body)
		case response.StatusCode == http.StatusTooManyRequests:
			rateErr := c.httpError(ErrorRateLimited, request, response.StatusCode, body)
			rateErr.RetryAfter = retryAfter(response.Header, meta.RateLimit)
			if attempt < c.maxAttempts {
				if sleepErr := c.sleep(ctx, c.backoff(attempt, rateErr.RetryAfter)); sleepErr != nil {
					return body, meta, rateErr
				}
				continue
			}
			return body, meta, rateErr
		case response.StatusCode >= 500:
			lastErr = c.httpError(ErrorUnavailable, request, response.StatusCode, body)
			if attempt < c.maxAttempts {
				if sleepErr := c.sleep(ctx, c.backoff(attempt, 0)); sleepErr != nil {
					return body, meta, lastErr
				}
				continue
			}
			return body, meta, lastErr
		case response.StatusCode == http.StatusConflict:
			return body, meta, c.httpError(ErrorConflict, request, response.StatusCode, body)
		default:
			return body, meta, c.httpError(ErrorInvalid, request, response.StatusCode, body)
		}
	}
	if lastErr == nil {
		lastErr = &Error{
			Kind:      ErrorUnavailable,
			Provider:  c.provider,
			Operation: request.operation(),
			Message:   "the provider call failed",
		}
	}
	return nil, nil, lastErr
}

func (c *Client) buildRequest(ctx context.Context, request Request, endpoint string, payload []byte) (*http.Request, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	method := request.Method
	if method == "" {
		method = http.MethodGet
	}
	httpRequest, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, &Error{
			Kind:      ErrorInvalid,
			Provider:  c.provider,
			Operation: request.operation(),
			Message:   "the provider request could not be built",
			Cause:     err,
		}
	}
	accept := request.Accept
	if accept == "" {
		accept = "application/json"
	}
	httpRequest.Header.Set("Accept", accept)
	httpRequest.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+c.token)
	}
	if payload != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	return httpRequest, nil
}

func (c *Client) endpoint(request Request) (string, error) {
	path := request.Path
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	endpoint := c.baseURL + path
	if len(request.Query) > 0 {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return "", err
		}
		query := parsed.Query()
		for key, values := range request.Query {
			for _, value := range values {
				if value == "" {
					continue
				}
				query.Add(key, value)
			}
		}
		parsed.RawQuery = query.Encode()
		endpoint = parsed.String()
	}
	return endpoint, nil
}

func (c *Client) httpError(kind ErrorKind, request Request, status int, body []byte) *Error {
	return &Error{
		Kind:      kind,
		Provider:  c.provider,
		Operation: request.operation(),
		Message:   providerMessage(body),
		Status:    status,
	}
}

func (c *Client) backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	delay := c.baseDelay << (attempt - 1)
	if delay > 5*time.Second {
		delay = 5 * time.Second
	}
	// Deterministic jitter keeps concurrent workers from retrying in lockstep.
	jitter := time.Duration(rand.Int63n(int64(delay/2 + 1)))
	return delay + jitter
}

func (r Request) operation() string {
	if r.Path == "" {
		return strings.ToLower(r.Method)
	}
	return strings.ToLower(r.Method + " " + r.Path)
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// providerMessage extracts a human-readable message from a provider error
// payload, falling back to a truncated snippet of the raw body.
func providerMessage(body []byte) string {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return "the provider rejected the request"
	}
	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err == nil {
		for _, key := range []string{"message", "error_description", "error", "detail", "title"} {
			value, ok := payload[key]
			if !ok {
				continue
			}
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return typed
				}
			case map[string]any:
				if message, ok := typed["message"].(string); ok && message != "" {
					return message
				}
			case []any:
				if len(typed) > 0 {
					if message, ok := typed[0].(string); ok && message != "" {
						return message
					}
				}
			}
		}
	}
	snippet := string(trimmed)
	if len(snippet) > maxErrorBodySnippet {
		snippet = snippet[:maxErrorBodySnippet]
	}
	return snippet
}

func parseRateLimit(header http.Header) RateLimit {
	limit := firstInt(header, "X-RateLimit-Limit", "RateLimit-Limit")
	remaining := firstInt(header, "X-RateLimit-Remaining", "RateLimit-Remaining")
	reset := firstInt64(header, "X-RateLimit-Reset", "RateLimit-Reset")
	out := RateLimit{Limit: limit, Remaining: remaining}
	switch {
	case reset > 0 && reset < 1_000_000_000:
		// RateLimit-Reset (RFC draft) is expressed in seconds from now.
		out.Reset = time.Now().UTC().Add(time.Duration(reset) * time.Second)
	case reset > 0:
		out.Reset = time.Unix(reset, 0).UTC()
	}
	if limit == 0 && remaining == 0 {
		return RateLimit{}
	}
	return out
}

func retryAfter(header http.Header, rateLimit RateLimit) time.Duration {
	if value := strings.TrimSpace(header.Get("Retry-After")); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
		if when, err := http.ParseTime(value); err == nil {
			delay := time.Until(when)
			if delay > 0 {
				return delay
			}
		}
	}
	if !rateLimit.Reset.IsZero() {
		if delay := time.Until(rateLimit.Reset); delay > 0 {
			return delay
		}
	}
	return time.Minute
}

// parsePage understands the pagination conventions of the supported forges:
// GitHub/Giteria use RFC 5988 Link headers, GitLab uses X-Next-Page.
func parsePage(header http.Header) Page {
	page := Page{}
	if next := parseLinkHeader(header.Get("Link")); next != "" {
		page.NextURL = next
		page.HasMore = true
		page.NextPage = pageFromURL(next)
	}
	if value := strings.TrimSpace(header.Get("X-Next-Page")); value != "" {
		if next, err := strconv.Atoi(value); err == nil && next > 0 {
			page.NextPage = next
			page.HasMore = true
		}
	}
	if strings.EqualFold(strings.TrimSpace(header.Get("X-HasMore")), "true") {
		page.HasMore = true
	}
	return page
}

func parseLinkHeader(value string) string {
	if value == "" {
		return ""
	}
	for _, part := range strings.Split(value, ",") {
		segments := strings.Split(strings.TrimSpace(part), ";")
		if len(segments) < 2 {
			continue
		}
		urlPart := strings.Trim(strings.TrimSpace(segments[0]), "<>")
		for _, attribute := range segments[1:] {
			if strings.TrimSpace(attribute) == `rel="next"` {
				return urlPart
			}
		}
	}
	return ""
}

func pageFromURL(raw string) int {
	parsed, err := url.Parse(raw)
	if err != nil {
		return 0
	}
	value := parsed.Query().Get("page")
	page, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return page
}

func firstInt(header http.Header, keys ...string) int {
	for _, key := range keys {
		value := strings.TrimSpace(header.Get(key))
		if value == "" {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func firstInt64(header http.Header, keys ...string) int64 {
	for _, key := range keys {
		value := strings.TrimSpace(header.Get(key))
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

// IsNotFound reports whether an error is a provider 404.
func IsNotFound(err error) bool {
	var providerErr *Error
	return errors.As(err, &providerErr) && providerErr.Kind == ErrorNotFound
}

// IsUnsupported reports whether an error is a missing capability.
func IsUnsupported(err error) bool {
	var providerErr *Error
	return errors.As(err, &providerErr) && providerErr.Kind == ErrorUnsupported
}

// IsRetryable reports whether the error may be retried later.
func IsRetryable(err error) bool {
	var providerErr *Error
	return errors.As(err, &providerErr) && providerErr.Retryable()
}

// RetryDelay exposes the retry hint of a provider error.
func RetryDelay(err error) (time.Duration, bool) {
	var providerErr *Error
	if !errors.As(err, &providerErr) {
		return 0, false
	}
	if providerErr.Kind != ErrorRateLimited {
		return 0, false
	}
	if providerErr.RetryAfter <= 0 {
		return time.Minute, true
	}
	return providerErr.RetryAfter, true
}
