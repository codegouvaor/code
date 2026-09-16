package gitlab

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
		Token:       "gl-token",
		MaxAttempts: 1,
	})
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	gitlabAdapter, ok := adapter.(*Adapter)
	if !ok {
		t.Fatalf("unexpected adapter type %T", adapter)
	}
	return gitlabAdapter, server
}

func TestRepositoryEncodesProjectPath(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/projects/etat%2Ffeuille-de-route" && r.URL.Path != "/projects/etat/feuille-de-route" {
			t.Errorf("unexpected path: %s", r.URL.EscapedPath())
		}
		_, _ = w.Write([]byte(`{
			"id": 7,
			"name": "feuille-de-route",
			"path_with_namespace": "etat/feuille-de-route",
			"web_url": "https://gitlab.com/etat/feuille-de-route",
			"default_branch": "main",
			"visibility": "public",
			"namespace": {"full_path": "etat", "kind": "group"}
		}`))
	}))
	defer server.Close()

	repository, err := adapter.Repository(context.Background(), providers.RepositoryRef{Owner: "etat", Repo: "feuille-de-route"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.ID != "7" || repository.OwnerType != "organization" {
		t.Fatalf("unexpected repository: %+v", repository)
	}
}

func TestReviewsMarkMergedRequests(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{
			"id": 11,
			"iid": 4,
			"title": "Ameliore la recherche",
			"state": "merged",
			"web_url": "https://gitlab.com/mr/4",
			"source_branch": "feature",
			"target_branch": "main",
			"author": {"username": "camille"},
			"merged_at": "2026-01-02T03:04:05Z"
		}]`))
	}))
	defer server.Close()

	reviews, err := adapter.Reviews(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "etat", Repo: "service"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reviews) != 1 || !reviews[0].Merged {
		t.Fatalf("merged flag not set: %+v", reviews)
	}
	if reviews[0].MergedAt.IsZero() {
		t.Fatal("mergedAt not parsed")
	}
}

func TestReleasesMapTags(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{
			"tag_name": "v1.2.0",
			"name": "v1.2.0",
			"description": "Notes",
			"created_at": "2026-02-02T00:00:00Z",
			"released_at": "2026-02-03T00:00:00Z",
			"_links": {"self_url": "https://gitlab.com/releases/v1.2.0"}
		}]`))
	}))
	defer server.Close()

	releases, err := adapter.Releases(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "etat", Repo: "service"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(releases) != 1 || releases[0].TagName != "v1.2.0" {
		t.Fatalf("unexpected releases: %+v", releases)
	}
	if releases[0].PublishedAt.IsZero() {
		t.Fatal("publishedAt not parsed")
	}
}

func TestAuthfailuresAreUnauthorized(t *testing.T) {
	t.Parallel()

	adapter, server := newAdapter(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"401 Unauthorized"}`))
	}))
	defer server.Close()

	_, err := adapter.Issues(context.Background(), providers.ListQuery{
		Ref: providers.RepositoryRef{Owner: "etat", Repo: "prive"},
	})
	var providerErr *providers.Error
	if !errors.As(err, &providerErr) || providerErr.Kind != providers.ErrorUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestDescriptorDeclaresPackages(t *testing.T) {
	t.Parallel()

	if got := Descriptor.Capability(providers.CapabilityPackages); got != providers.CapabilityIntegrated {
		t.Fatalf("expected integrated packages, got %s", got)
	}
}
