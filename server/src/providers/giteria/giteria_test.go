package giteria

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegouvaor/code/server/src/providers"
)

func newAdapter(t *testing.T, handler http.Handler) (*Adapter, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	adapter, err := Factory(providers.Options{
		BaseURL:     server.URL,
		Token:       "giteria-token",
		MaxAttempts: 1,
	})
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	giteriaAdapter, ok := adapter.(*Adapter)
	if !ok {
		t.Fatalf("unexpected adapter type %T", adapter)
	}
	return giteriaAdapter, server
}

func TestRepositoryMapsGiteaPayload(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/codegouvaor/react-ads" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"id": 99,
			"name": "react-ads",
			"full_name": "codegouvaor/react-ads",
			"html_url": "https://giteria.gouv.aor/codegouvaor/react-ads",
			"default_branch": "main",
			"private": false,
			"stars_count": 8,
			"open_issues_count": 2,
			"owner": {"login": "codegouvaor", "type": "Organization"}
		}`))
	}))
	defer server.Close()

	repository, err := adapter.Repository(context.Background(), providers.RepositoryRef{Owner: "codegouvaor", Repo: "react-ads"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.ID != "99" || repository.Visibility != "public" {
		t.Fatalf("unexpected repository: %+v", repository)
	}
	if repository.OwnerType != "organization" {
		t.Fatalf("unexpected owner type: %q", repository.OwnerType)
	}
}

func TestBlobDecodesBase64Content(t *testing.T) {
	t.Parallel()

	encoded := base64.StdEncoding.EncodeToString([]byte("bonjour"))
	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"name": "README.md",
			"path": "README.md",
			"sha": "abc",
			"type": "file",
			"size": 7,
			"encoding": "base64",
			"content": "` + encoded + `"
		}`))
	}))
	defer server.Close()

	blob, err := adapter.Blob(context.Background(), providers.BlobQuery{
		Ref: providers.RepositoryRef{Owner: "codegouvaor", Repo: "react-ads"}, Path: "README.md",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := DecodeContent(blob.Content, blob.Encoding)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded != "bonjour" {
		t.Fatalf("unexpected content: %q", decoded)
	}
}

func TestIssuesIgnorePullRequests(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("type"); got != "issues" {
			t.Errorf("expected type=issues, got %q", got)
		}
		_, _ = w.Write([]byte(`[
			{"id":1,"number":1,"title":"Ticket","state":"open","html_url":"u1"},
			{"id":2,"number":2,"title":"PR","state":"open","html_url":"u2","pull_request":{"merged":false}}
		]`))
	}))
	defer server.Close()

	issues, err := adapter.Issues(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "codegouvaor", Repo: "react-ads"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].Title != "Ticket" {
		t.Fatalf("expected a single issue, got %+v", issues)
	}
}

func TestBranchesUseDefaultBranchFromRepository(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/codegouvaor/react-ads":
			_, _ = w.Write([]byte(`{"id":1,"name":"react-ads","default_branch":"main"}`))
		case "/repos/codegouvaor/react-ads/branches":
			_, _ = w.Write([]byte(`[{"name":"main","protected":true,"commit":{"id":"sha1"}}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	branches, err := adapter.Branches(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "codegouvaor", Repo: "react-ads"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 || !branches[0].Default {
		t.Fatalf("default branch not set: %+v", branches)
	}
}

func TestDescriptorDeclaresNativeDocumentation(t *testing.T) {
	t.Parallel()

	if got := Descriptor.Capability(providers.CapabilityDocumentation); got != providers.CapabilityHybrid {
		t.Fatalf("expected hybrid documentation, got %s", got)
	}
	if got := Descriptor.Capability(providers.CapabilityReviews); got != providers.CapabilityIntegrated {
		t.Fatalf("expected integrated reviews, got %s", got)
	}
}
