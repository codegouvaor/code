package providers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(baseURL string, maxAttempts int) *Client {
	return NewClient(ClientConfig{
		Provider:    "test",
		BaseURL:     baseURL,
		Token:       "token",
		MaxAttempts: maxAttempts,
		BaseDelay:   time.Millisecond,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	})
}

func TestClientDecodesSuccessfulPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("unexpected authorization header: %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("unexpected page: %q", got)
		}
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "4999")
		w.Header().Set("Link", `<https://api.example.com/items?page=3>; rel="next"`)
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	defer server.Close()

	var payload struct {
		Value string `json:"value"`
	}
	response, err := testClient(server.URL, 1).Do(context.Background(), Request{
		Path:  "/items",
		Query: map[string][]string{"page": {"2"}},
	}, &payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Value != "ok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if response.RateLimit.Limit != 5000 || response.RateLimit.Remaining != 4999 {
		t.Fatalf("rate limit not parsed: %+v", response.RateLimit)
	}
	if !response.Page.HasMore || response.Page.NextPage != 3 {
		t.Fatalf("pagination not parsed: %+v", response.Page)
	}
}

func TestClientMapsNotFoundAndUnauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
		}
	}))
	defer server.Close()

	client := testClient(server.URL, 1)
	_, err := client.Do(context.Background(), Request{Path: "/missing"}, nil)
	if !IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
	_, err = client.Do(context.Background(), Request{Path: "/private"}, nil)
	var providerErr *Error
	if !errors.As(err, &providerErr) || providerErr.Kind != ErrorUnauthorized {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if providerErr.Message != "Bad credentials" {
		t.Fatalf("provider message not extracted: %q", providerErr.Message)
	}
}

func TestClientRetriesServerErrors(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	var payload map[string]bool
	if _, err := testClient(server.URL, 3).Do(context.Background(), Request{Path: "/flaky"}, &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestClientHonoursRateLimitHint(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", strconv.Itoa(30))
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer server.Close()

	_, err := testClient(server.URL, 1).Do(context.Background(), Request{Path: "/issues"}, nil)
	delay, ok := RetryDelay(err)
	if !ok {
		t.Fatalf("expected a rate limit error, got %v", err)
	}
	if delay < 30*time.Second {
		t.Fatalf("expected the Retry-After hint, got %s", delay)
	}
}

func TestClientReportsMissingBaseURL(t *testing.T) {
	t.Parallel()

	client := NewClient(ClientConfig{Provider: "github"})
	_, err := client.Do(context.Background(), Request{Path: "/repos"}, nil)
	var providerErr *Error
	if !errors.As(err, &providerErr) || providerErr.Kind != ErrorUnavailable {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestAsAppErrorMapsProviderErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		kind ErrorKind
		code string
	}{
		{ErrorNotFound, "PROVIDER_RESOURCE_NOT_FOUND"},
		{ErrorUnauthorized, "PROVIDER_UNAUTHORIZED"},
		{ErrorRateLimited, "PROVIDER_RATE_LIMITED"},
		{ErrorUnavailable, "PROVIDER_UNAVAILABLE"},
		{ErrorUnsupported, "PROVIDER_CAPABILITY_UNAVAILABLE"},
	}
	for _, tc := range cases {
		err := &Error{Kind: tc.kind, Provider: "github", Message: "boom"}
		if got := AsAppError(err).Code; got != tc.code {
			t.Fatalf("kind %s: expected %s, got %s", tc.kind, tc.code, got)
		}
	}
}
