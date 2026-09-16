package github

import (
	"context"
	"errors"
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
		Token:       "test-token",
		MaxAttempts: 1,
	})
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	githubAdapter, ok := adapter.(*Adapter)
	if !ok {
		t.Fatalf("unexpected adapter type %T", adapter)
	}
	return githubAdapter, server
}

func TestRepositoryMapsGitHubPayload(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/state/roadmap" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("unexpected authorization: %q", got)
		}
		_, _ = w.Write([]byte(`{
			"id": 42,
			"name": "roadmap",
			"full_name": "state/roadmap",
			"description": "Roadmap",
			"html_url": "https://github.com/state/roadmap",
			"clone_url": "https://github.com/state/roadmap.git",
			"default_branch": "main",
			"visibility": "public",
			"topics": ["gouvernement"],
			"stargazers_count": 12,
			"forks_count": 3,
			"open_issues_count": 5,
			"owner": {"login": "state", "type": "Organization"}
		}`))
	}))
	defer server.Close()

	repository, err := adapter.Repository(context.Background(), providers.RepositoryRef{Owner: "state", Repo: "roadmap"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.ID != "42" || repository.DefaultBranch != "main" {
		t.Fatalf("unexpected repository: %+v", repository)
	}
	if repository.OwnerType != "organization" || repository.OwnerLogin != "state" {
		t.Fatalf("unexpected owner: %+v", repository)
	}
	if len(repository.Topics) != 1 || repository.Topics[0] != "gouvernement" {
		t.Fatalf("unexpected topics: %+v", repository.Topics)
	}
}

func TestIssuesSkipsPullRequests(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/state/roadmap/issues":
			_, _ = w.Write([]byte(`[
				{"id":1,"number":1,"title":"Bug","state":"open","html_url":"u1","labels":[{"name":"bug"}]},
				{"id":2,"number":2,"title":"PR","state":"open","html_url":"u2","pull_request":{"url":"p"}}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	issues, err := adapter.Issues(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "state", Repo: "roadmap"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].Title != "Bug" {
		t.Fatalf("expected a single issue, got %+v", issues)
	}
	if len(issues[0].Labels) != 1 || issues[0].Labels[0] != "bug" {
		t.Fatalf("labels not mapped: %+v", issues[0].Labels)
	}
}

func TestBranchesMarksDefaultBranch(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/state/roadmap":
			_, _ = w.Write([]byte(`{"id":1,"name":"roadmap","default_branch":"main"}`))
		case "/repos/state/roadmap/branches":
			_, _ = w.Write([]byte(`[
				{"name":"main","protected":true,"commit":{"sha":"abc"}},
				{"name":"next","protected":false,"commit":{"sha":"def"}}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	branches, err := adapter.Branches(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "state", Repo: "roadmap"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 2 || !branches[0].Default || branches[1].Default {
		t.Fatalf("default branch not detected: %+v", branches)
	}
}

func TestTreeRejectsUnknownRepository(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer server.Close()

	_, err := adapter.Tree(context.Background(), providers.TreeQuery{
		Ref: providers.RepositoryRef{Owner: "state", Repo: "missing"},
	})
	var providerErr *providers.Error
	if !errors.As(err, &providerErr) || providerErr.Kind != providers.ErrorNotFound {
		t.Fatalf("expected a not found error, got %v", err)
	}
}

func TestDescriptorDeclaresCapabilities(t *testing.T) {
	t.Parallel()

	if got := Descriptor.Capability(providers.CapabilityReviews); got != providers.CapabilityIntegrated {
		t.Fatalf("expected integrated reviews, got %s", got)
	}
	if got := Descriptor.Capability(providers.CapabilityPackages); got != providers.CapabilityUnavailable {
		t.Fatalf("expected unavailable packages, got %s", got)
	}
}

func TestFactoryUsesDefaultBaseURL(t *testing.T) {
	t.Parallel()

	adapter, err := Factory(providers.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	githubAdapter := adapter.(*Adapter)
	if githubAdapter.client.BaseURL() != defaultBaseURL {
		t.Fatalf("expected the default base URL, got %q", githubAdapter.client.BaseURL())
	}
}
