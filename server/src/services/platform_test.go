package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/codegouvaor/code/server/src/config"
	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
)

// ── Secret box ───────────────────────────────────────────────────────────────

func TestSecretBoxRoundTrip(t *testing.T) {
	t.Parallel()

	box, err := newSecretBox("test-key")
	if err != nil {
		t.Fatalf("newSecretBox: %v", err)
	}
	sealed, err := box.Seal("ghp_secret")
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	if strings.Contains(sealed, "ghp_secret") {
		t.Fatal("the sealed value leaks the clear text secret")
	}
	opened, err := box.Open(sealed)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened != "ghp_secret" {
		t.Fatalf("unexpected opened value: %q", opened)
	}
	if empty, err := box.Seal(""); err != nil || empty != "" {
		t.Fatalf("empty secrets must stay empty: %q %v", empty, err)
	}
}

func TestSecretBoxRejectsTamperedPayload(t *testing.T) {
	t.Parallel()

	box, _ := newSecretBox("test-key")
	sealed, _ := box.Seal("token")
	if _, err := box.Open(sealed[:len(sealed)-2] + "aa"); err == nil {
		t.Fatal("expected a tampered payload to be rejected")
	}
	if _, err := box.Open("clear-text"); err == nil {
		t.Fatal("expected an unsupported payload format to be rejected")
	}
}

func TestSecretBoxRequiresKey(t *testing.T) {
	t.Parallel()

	if _, err := newSecretBox("   "); err == nil {
		t.Fatal("expected a missing key to be rejected")
	}
}

// ── Authorization ────────────────────────────────────────────────────────────

func newTestAuthorizer() (*Authorizer, *fakeProjectRepo, *fakeProjectMemberRepo, *fakeOrgRepo, *fakeOrgMemberRepo) {
	projects := newFakeProjectRepo()
	members := newFakeProjectMemberRepo()
	orgs := newFakeOrgRepo()
	orgMembers := newFakeOrgMemberRepo()
	authorizer := NewAuthorizer(projects, members, orgs, orgMembers, newFakeOrgTeamMemberRepo())
	return authorizer, projects, members, orgs, orgMembers
}

func TestAuthorizationRoleMatrix(t *testing.T) {
	t.Parallel()

	authorizer, projects, members, _, _ := newTestAuthorizer()
	ctx := context.Background()
	project := projects.add(models.Project{
		Reference:  "state/roadmap",
		Namespace:  "state",
		Slug:       "roadmap",
		OwnerID:    "owner-1",
		Visibility: "public",
	})

	cases := []struct {
		name     string
		userID   string
		member   string
		action   Action
		expected bool
	}{
		{name: "owner can delete", userID: "owner-1", action: ActionProjectDelete, expected: true},
		{name: "anonymous cannot write", userID: "", action: ActionProjectWrite, expected: false},
		{name: "anonymous can read public", userID: "", action: ActionProjectRead, expected: true},
		{name: "member can write", userID: "member-1", member: "member", action: ActionProjectWrite, expected: true},
		{name: "member cannot manage members", userID: "member-1", member: "member", action: ActionMemberManage, expected: false},
		{name: "maintainer can sync", userID: "maintainer-1", member: "maintainer", action: ActionRepositorySync, expected: true},
		{name: "admin can manage members", userID: "admin-1", member: "admin", action: ActionMemberManage, expected: true},
		{name: "admin cannot delete", userID: "admin-1", member: "admin", action: ActionProjectDelete, expected: false},
	}
	for _, tc := range cases {
		if tc.member != "" {
			members.add(models.ProjectMember{ProjectID: project.ID, UserID: tc.userID, Role: tc.member})
		}
		access, err := authorizer.ProjectAccess(ctx, tc.userID, &project)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if got := access.Can(tc.action); got != tc.expected {
			t.Fatalf("%s: expected %v, got %v (role %s)", tc.name, tc.expected, got, access.Role)
		}
	}
}

func TestAuthorizationDeniesPrivateProjectForOutsider(t *testing.T) {
	t.Parallel()

	authorizer, projects, _, _, _ := newTestAuthorizer()
	project := projects.add(models.Project{
		Reference: "state/secret", Namespace: "state", Slug: "secret", OwnerID: "owner-1", Visibility: "private",
	})
	_, err := authorizer.AuthorizeProject(context.Background(), interfaces.Principal{UserID: "intruder"}, &project, ActionProjectRead)
	if !errors.Is(err, utils.ErrUnauthorized) && !errors.Is(err, utils.ErrForbidden) {
		t.Fatalf("expected the access to be denied, got %v", err)
	}
}

func TestOrganizationOwnershipGrantsProjectAdmin(t *testing.T) {
	t.Parallel()

	authorizer, projects, _, orgs, orgMembers := newTestAuthorizer()
	ctx := context.Background()
	organization := orgs.add(models.Organization{Slug: "codegouvaor", Name: "Code", OwnerID: "owner-1", Visibility: "public"})
	organizationID := organization.ID
	project := projects.add(models.Project{
		Reference:      "codegouvaor/react-ads",
		Namespace:      "codegouvaor",
		Slug:           "react-ads",
		OwnerID:        "owner-1",
		OrganizationID: &organizationID,
		Visibility:     "private",
	})
	// An organization admin inherits administrative rights on its projects.
	orgMembers.add(models.OrganizationMember{OrganizationID: organization.ID, UserID: "admin-2", Role: "admin"})
	access, err := authorizer.ProjectAccess(ctx, "admin-2", &project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !access.Can(ActionMemberManage) {
		t.Fatalf("expected the organization admin to manage project members, got role %s", access.Role)
	}
	if access.Can(ActionProjectDelete) {
		t.Fatal("only the owner may delete the project")
	}
}

// ── Projects ─────────────────────────────────────────────────────────────────

func newTestProjectService() (*ProjectService, *fakeProjectRepo, *fakeUserRepo, *fakeOrgRepo, *fakeBindingRepo) {
	projects := newFakeProjectRepo()
	users := newFakeUserRepo()
	orgs := newFakeOrgRepo()
	bindings := newFakeBindingRepo()
	authorizer := NewAuthorizer(projects, newFakeProjectMemberRepo(), orgs, newFakeOrgMemberRepo(), newFakeOrgTeamMemberRepo())
	registry := providers.NewRegistry()
	service := NewProjectService(
		projects, newFakeProjectMemberRepo(), newFakeProjectAssetRepo(), newFakeProjectStarRepo(),
		newFakeProjectWatchRepo(), bindings, orgs, newFakeOrgMemberRepo(), users, registry, nil, authorizer,
	)
	return service, projects, users, orgs, bindings
}

func TestProjectCreationBuildsStableReference(t *testing.T) {
	t.Parallel()

	service, _, users, _, _ := newTestProjectService()
	username := "codegouvaor"
	users.add(models.User{Common: models.Common{ID: "user-1"}, Email: "a@example.gouv.aor", DisplayName: "Code", Username: &username})

	project, err := service.Create(context.Background(), interfaces.Principal{UserID: "user-1"}, CreateProjectInput{
		Name: "React Ads", Visibility: "public",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Reference != "codegouvaor/react-ads" {
		t.Fatalf("unexpected reference: %q", project.Reference)
	}
	view, err := service.View(context.Background(), interfaces.Principal{UserID: "user-1"}, project.Reference)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if view.ViewerPermission != string(RoleOwner) {
		t.Fatalf("unexpected viewer permission: %q", view.ViewerPermission)
	}
	if view.Capabilities[providers.CapabilityDocumentation] != string(providers.CapabilityNative) {
		t.Fatalf("expected Code-native capabilities without a binding: %+v", view.Capabilities)
	}
}

func TestProjectRequiresUsername(t *testing.T) {
	t.Parallel()

	service, _, users, _, _ := newTestProjectService()
	users.add(models.User{Common: models.Common{ID: "user-1"}, Email: "a@example.gouv.aor", DisplayName: "Code"})
	if _, err := service.Create(context.Background(), interfaces.Principal{UserID: "user-1"}, CreateProjectInput{Name: "Docs"}); !errors.Is(err, utils.ErrUsernameInvalid) {
		t.Fatalf("expected a username error, got %v", err)
	}
}

func TestProjectVisibilityHidesPrivateProjects(t *testing.T) {
	t.Parallel()

	service, projects, users, _, _ := newTestProjectService()
	owner := "owner-name"
	users.add(models.User{Common: models.Common{ID: "user-1"}, Email: "a@example.gouv.aor", DisplayName: "Owner", Username: &owner})
	project := projects.add(models.Project{
		Reference: "owner-name/secret", Namespace: "owner-name", Slug: "secret", OwnerID: "user-1", Visibility: "private",
	})
	views, _, err := service.List(context.Background(), interfaces.Principal{UserID: "someone-else"}, interfaces.ProjectFilter{Visibility: "private"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("expected the private project to be hidden, got %+v", views)
	}
	if _, err := service.Get(context.Background(), interfaces.Principal{UserID: "someone-else"}, project.ID); err == nil {
		t.Fatal("expected the private project to be inaccessible")
	}
}

func TestProjectStarToggles(t *testing.T) {
	t.Parallel()

	service, projects, _, _, _ := newTestProjectService()
	project := projects.add(models.Project{
		Reference: "state/roadmap", Namespace: "state", Slug: "roadmap", OwnerID: "user-1", Visibility: "public",
	})
	principal := interfaces.Principal{UserID: "user-2"}
	view, err := service.Star(context.Background(), principal, project.ID)
	if err != nil {
		t.Fatalf("star: %v", err)
	}
	if !view.Starred || view.Stars != 1 {
		t.Fatalf("unexpected star state: %+v", view)
	}
	view, err = service.Unstar(context.Background(), principal, project.ID)
	if err != nil {
		t.Fatalf("unstar: %v", err)
	}
	if view.Starred || view.Stars != 0 {
		t.Fatalf("unexpected star state after unstar: %+v", view)
	}
}

func TestProjectAssetKindsAreValidated(t *testing.T) {
	t.Parallel()

	service, projects, _, _, _ := newTestProjectService()
	project := projects.add(models.Project{
		Reference: "state/docs", Namespace: "state", Slug: "docs", OwnerID: "user-1", Visibility: "public",
	})
	principal := interfaces.Principal{UserID: "user-1"}
	if _, err := service.CreateAsset(context.Background(), principal, project.ID, CreateAssetInput{Kind: "packages", Name: "nope"}); !errors.Is(err, utils.ErrValidationFailed) {
		t.Fatalf("expected an invalid kind error, got %v", err)
	}
	asset, err := service.CreateAsset(context.Background(), principal, project.ID, CreateAssetInput{Kind: "api", Name: "Public API"})
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}
	if asset.Slug != "public-api" {
		t.Fatalf("unexpected slug: %q", asset.Slug)
	}
	if _, err := service.CreateAsset(context.Background(), principal, project.ID, CreateAssetInput{Kind: "api", Name: "Public API"}); err == nil {
		t.Fatal("expected a duplicate slug to be rejected")
	}
}

// ── Jobs ─────────────────────────────────────────────────────────────────────

func TestJobEnqueueIsIdempotent(t *testing.T) {
	t.Parallel()

	repository := newFakeJobRepo()
	service := NewJobService(repository, nil, nil)
	ctx := context.Background()
	first, err := service.Submit(ctx, EnqueueInput{Kind: models.SyncJobRepositorySync, IdempotencyKey: "key-1"})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	second, err := service.Submit(ctx, EnqueueInput{Kind: models.SyncJobRepositorySync, IdempotencyKey: "key-1"})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected the same job, got %s and %s", first.ID, second.ID)
	}
}

func TestJobRetriesRetryableFailures(t *testing.T) {
	t.Parallel()

	repository := newFakeJobRepo()
	service := NewJobService(repository, nil, nil)
	service.backoffBase = time.Millisecond
	ctx := context.Background()
	service.Register(models.SyncJobRepositorySync, func(context.Context, *models.SyncJob) error {
		return &utils.AppError{Status: http.StatusServiceUnavailable, Code: "UPSTREAM", Message: "later"}
	})
	if _, err := service.Submit(ctx, EnqueueInput{Kind: models.SyncJobRepositorySync, IdempotencyKey: "retry"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	processed, err := service.ProcessOnce(ctx)
	if err != nil || !processed {
		t.Fatalf("expected a processed job, got %v %v", processed, err)
	}
	jobs, err := repository.ListByStatus(ctx, models.SyncJobFailed, 10)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("expected a failed job scheduled for retry, got %v %v", jobs, err)
	}
	if jobs[0].RunAfter.Before(time.Now().UTC()) {
		t.Fatal("expected the retry to be scheduled in the future")
	}
}

func TestJobStopsRetryingPermanentFailures(t *testing.T) {
	t.Parallel()

	repository := newFakeJobRepo()
	service := NewJobService(repository, nil, nil)
	ctx := context.Background()
	service.Register(models.SyncJobRepositorySync, func(context.Context, *models.SyncJob) error {
		return utils.ErrForbidden
	})
	if _, err := service.Submit(ctx, EnqueueInput{Kind: models.SyncJobRepositorySync, IdempotencyKey: "permanent"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := service.ProcessOnce(ctx); err != nil {
		t.Fatalf("process: %v", err)
	}
	dead, err := repository.ListByStatus(ctx, models.SyncJobDead, 10)
	if err != nil || len(dead) != 1 {
		t.Fatalf("expected a dead job, got %v %v", dead, err)
	}
}

func TestJobStatsCountsStatuses(t *testing.T) {
	t.Parallel()

	repository := newFakeJobRepo()
	service := NewJobService(repository, nil, nil)
	if _, err := service.Submit(context.Background(), EnqueueInput{Kind: models.SyncJobRepositorySync, IdempotencyKey: "s-1"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	stats, err := service.Stats(context.Background())
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats[models.SyncJobQueued] != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

// ── Synchronisation ──────────────────────────────────────────────────────────

type stubAdapter struct {
	repository    *providers.Repository
	repositoryErr error
}

func (a stubAdapter) Name() string { return "stub" }

func (a stubAdapter) Descriptor() providers.Descriptor {
	return providers.Descriptor{Name: "stub", Capabilities: map[string]providers.Capability{
		providers.CapabilityRepositories: providers.CapabilityIntegrated,
	}}
}

func (a stubAdapter) Repository(context.Context, providers.RepositoryRef) (*providers.Repository, error) {
	if a.repositoryErr != nil {
		return nil, a.repositoryErr
	}
	return a.repository, nil
}

type syncHarness struct {
	service       *SyncService
	projects      *fakeProjectRepo
	bindings      *fakeBindingRepo
	resources     *fakeResourceRepo
	cursors       *fakeCursorRepo
	subscriptions *fakeSubscriptionRepo
	jobs          *JobService
	connections   *ProviderConnectionService
}

func newSyncHarness(t *testing.T, adapter providers.Adapter) *syncHarness {
	t.Helper()
	projects := newFakeProjectRepo()
	projects.add(models.Project{
		Common:    models.Common{ID: "project-1"},
		Reference: "codegouvaor/react-ads", Namespace: "codegouvaor",
		Slug: "react-ads", OwnerID: "owner-1", Visibility: "public",
	})
	bindings := newFakeBindingRepo()
	resources := newFakeResourceRepo()
	cursors := newFakeCursorRepo()
	subscriptions := newFakeSubscriptionRepo()
	jobs := NewJobService(newFakeJobRepo(), nil, nil)
	registry := providers.NewRegistry()
	registry.Register(providers.Descriptor{
		Name: "giteria",
		Capabilities: map[string]providers.Capability{
			providers.CapabilityRepositories: providers.CapabilityIntegrated,
		},
	}, func(providers.Options) (providers.Adapter, error) {
		return adapter, nil
	})
	connections, err := NewProviderConnectionService(
		newFakeConnectionRepo(), registry,
		config.ProvidersConfig{EncryptionKey: "test-key", Giteria: config.ProviderEndpointConfig{BaseURL: "https://giteria.test/api/v1", Token: "webhook-secret"}},
		nil,
	)
	if err != nil {
		t.Fatalf("connections: %v", err)
	}
	authorizer := NewAuthorizer(newFakeProjectRepo(), newFakeProjectMemberRepo(), newFakeOrgRepo(), newFakeOrgMemberRepo(), newFakeOrgTeamMemberRepo())
	service := NewSyncService(
		projects, bindings, resources, cursors, subscriptions, newFakeUserRepo(),
		connections, nil, jobs, authorizer, nil,
	)
	service.RegisterHandlers()
	return &syncHarness{
		service: service, projects: projects, bindings: bindings, resources: resources, cursors: cursors,
		subscriptions: subscriptions, jobs: jobs, connections: connections,
	}
}

func (h *syncHarness) binding() models.RepositoryBinding {
	binding := h.bindings.add(models.RepositoryBinding{
		ProjectID:     "project-1",
		Provider:      models.ProviderGiteria,
		ExternalID:    "99",
		ExternalOwner: "codegouvaor",
		ExternalRepo:  "react-ads",
		SyncEnabled:   true,
		SyncStatus:    models.BindingSyncSynced,
	})
	return binding
}

func signGitHubBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	harness := newSyncHarness(t, stubAdapter{})
	binding := harness.binding()
	headers := http.Header{}
	headers.Set("X-GitHub-Event", "issues")
	headers.Set("X-Hub-Signature-256", signGitHubBody("wrong-secret", []byte(`{}`)))

	_, err := harness.service.HandleWebhook(context.Background(), models.ProviderGiteria, binding.ID, headers, []byte(`{}`))
	if !errors.Is(err, utils.ErrWebhookSignatureInvalid) {
		t.Fatalf("expected a signature error, got %v", err)
	}
}

func TestWebhookIsIdempotentAndOrdersEvents(t *testing.T) {
	t.Parallel()

	harness := newSyncHarness(t, stubAdapter{})
	binding := harness.binding()
	ctx := context.Background()
	recent := time.Now().UTC()

	first := issueWebhook(t, recent)
	headers := http.Header{}
	headers.Set("X-GitHub-Event", "issues")
	headers.Set("X-GitHub-Delivery", "delivery-1")
	headers.Set("X-Hub-Signature-256", signGitHubBody("webhook-secret", first))

	outcome, err := harness.service.HandleWebhook(ctx, models.ProviderGiteria, binding.ID, headers, first)
	if err != nil {
		t.Fatalf("first delivery: %v", err)
	}
	if !outcome.Accepted || outcome.Duplicate || outcome.Event != "issue.opened" {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}

	// Replaying the same delivery must never create a second job.
	replay, err := harness.service.HandleWebhook(ctx, models.ProviderGiteria, binding.ID, headers, first)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if !replay.Duplicate {
		t.Fatalf("expected the replay to be detected as a duplicate: %+v", replay)
	}

	// An older event is applied idempotently but flagged as stale.
	staleBody := issueWebhook(t, recent.Add(-2*time.Hour))
	staleHeaders := http.Header{}
	staleHeaders.Set("X-GitHub-Event", "issues")
	staleHeaders.Set("X-GitHub-Delivery", "delivery-0")
	staleHeaders.Set("X-Hub-Signature-256", signGitHubBody("webhook-secret", staleBody))
	stale, err := harness.service.HandleWebhook(ctx, models.ProviderGiteria, binding.ID, staleHeaders, staleBody)
	if err != nil {
		t.Fatalf("stale delivery: %v", err)
	}
	if !stale.Stale {
		t.Fatalf("expected the older event to be flagged stale: %+v", stale)
	}

	// The cursor must not move backwards.
	cursor, err := harness.cursors.Get(ctx, binding.ID, models.ExternalResourceIssue)
	if err != nil {
		t.Fatalf("cursor: %v", err)
	}
	if cursor.LastEventAt == nil || cursor.LastEventAt.Before(recent.Add(-time.Minute)) {
		t.Fatalf("cursor moved backwards: %+v", cursor)
	}
	resource, err := harness.resources.GetByBindingAndExternal(ctx, binding.ID, models.ExternalResourceIssue, "5001")
	if err != nil {
		t.Fatalf("read model not updated: %v", err)
	}
	if resource.Title == "" {
		t.Fatalf("unexpected read model entry: %+v", resource)
	}
}

func TestWebhookDetachesDeletedRepository(t *testing.T) {
	t.Parallel()

	harness := newSyncHarness(t, stubAdapter{})
	binding := harness.binding()
	body := []byte(`{"action":"deleted","repository":{"id":99,"full_name":"codegouvaor/react-ads","updated_at":"2026-01-01T00:00:00Z"}}`)
	headers := http.Header{}
	headers.Set("X-GitHub-Event", "repository")
	headers.Set("X-GitHub-Delivery", "delivery-deleted")
	headers.Set("X-Hub-Signature-256", signGitHubBody("webhook-secret", body))

	outcome, err := harness.service.HandleWebhook(context.Background(), models.ProviderGiteria, binding.ID, headers, body)
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if outcome.Event != "repository.deleted" {
		t.Fatalf("unexpected event: %q", outcome.Event)
	}
	updated, err := harness.bindings.GetByID(context.Background(), binding.ID)
	if err != nil {
		t.Fatalf("binding: %v", err)
	}
	if updated.SyncStatus != models.BindingSyncDetached || updated.DetachedAt == nil {
		t.Fatalf("expected the binding to be detached: %+v", updated)
	}
}

func TestSyncBindingDetachesWhenRepositoryDisappears(t *testing.T) {
	t.Parallel()

	harness := newSyncHarness(t, stubAdapter{repositoryErr: providers.NewError(providers.ErrorNotFound, "giteria", "repository", "gone", nil)})
	binding := harness.binding()
	if _, err := harness.service.SyncBinding(context.Background(), binding.ID); err != nil {
		t.Fatalf("sync: %v", err)
	}
	updated, _ := harness.bindings.GetByID(context.Background(), binding.ID)
	if updated.SyncStatus != models.BindingSyncDetached {
		t.Fatalf("expected a detached binding, got %q", updated.SyncStatus)
	}
}

func TestSyncBindingSurfacesRateLimitAndRetries(t *testing.T) {
	t.Parallel()

	rateErr := providers.NewError(providers.ErrorRateLimited, "giteria", "repository", "slow down", nil)
	harness := newSyncHarness(t, stubAdapter{repositoryErr: rateErr})
	binding := harness.binding()
	projectID := binding.ProjectID
	bindingID := binding.ID
	if _, err := harness.jobs.Submit(context.Background(), EnqueueInput{
		Kind: models.SyncJobRepositorySync, IdempotencyKey: "sync-1", ProjectID: &projectID, BindingID: &bindingID,
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := harness.jobs.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("process: %v", err)
	}
	failed, err := harness.jobs.List(context.Background(), models.SyncJobFailed, 10)
	if err != nil || len(failed) != 1 {
		t.Fatalf("expected the rate limit to be retried, got %v %v", failed, err)
	}
	updated, _ := harness.bindings.GetByID(context.Background(), binding.ID)
	if updated.SyncStatus != models.BindingSyncError {
		t.Fatalf("expected the binding to record the error, got %q", updated.SyncStatus)
	}
}

func TestSyncBindingUpdatesMetadata(t *testing.T) {
	t.Parallel()

	updatedAt := time.Now().UTC().Add(-time.Hour)
	harness := newSyncHarness(t, stubAdapter{repository: &providers.Repository{
		Provider:      models.ProviderGiteria,
		ID:            "99",
		Name:          "react-ads",
		FullName:      "codegouvaor/react-ads",
		URL:           "https://giteria.test/codegouvaor/react-ads",
		DefaultBranch: "main",
		UpdatedAt:     updatedAt,
	}})
	binding := harness.binding()
	synced, err := harness.service.SyncBinding(context.Background(), binding.ID)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if synced.SyncStatus != models.BindingSyncSynced || synced.DefaultBranch != "main" {
		t.Fatalf("unexpected binding: %+v", synced)
	}
	if synced.LastSyncedAt == nil {
		t.Fatal("expected lastSyncedAt to be recorded")
	}
}

func issueWebhook(t *testing.T, updatedAt time.Time) []byte {
	t.Helper()
	payload := map[string]any{
		"action": "opened",
		"issue": map[string]any{
			"id":         5001,
			"number":     12,
			"title":      "Un bug",
			"state":      "open",
			"html_url":   "https://giteria.test/codegouvaor/react-ads/issues/12",
			"updated_at": updatedAt.Format(time.RFC3339),
			"user":       map[string]any{"login": "camille"},
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return encoded
}

// ── Provider connections ─────────────────────────────────────────────────────

func TestProviderConnectionStoresEncryptedTokens(t *testing.T) {
	t.Parallel()

	repository := newFakeConnectionRepo()
	registry := providers.NewRegistry()
	registry.Register(providers.Descriptor{Name: "github"}, func(providers.Options) (providers.Adapter, error) {
		return stubAdapter{}, nil
	})
	service, err := NewProviderConnectionService(repository, registry, config.ProvidersConfig{EncryptionKey: "test-key"}, nil)
	if err != nil {
		t.Fatalf("service: %v", err)
	}
	connection, err := service.Upsert(context.Background(), "user-1", ConnectionInput{
		Provider: "github", ProviderAccountID: "42", AccountLogin: "camille",
		AccessToken: "ghp_plain", Scopes: "read:user user:email",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if strings.Contains(connection.AccessTokenEnc, "ghp_plain") {
		t.Fatal("the access token was stored in clear text")
	}
	token, err := service.AccessToken(connection)
	if err != nil {
		t.Fatalf("access token: %v", err)
	}
	if token != "ghp_plain" {
		t.Fatalf("unexpected token: %q", token)
	}
	views, err := service.List(context.Background(), "user-1")
	if err != nil || len(views) != 1 {
		t.Fatalf("list: %v %v", views, err)
	}
	if len(views[0].Scopes) != 2 {
		t.Fatalf("unexpected scopes: %+v", views[0].Scopes)
	}
}

func TestProviderConnectionCannotBeStolen(t *testing.T) {
	t.Parallel()

	repository := newFakeConnectionRepo()
	registry := providers.NewRegistry()
	registry.Register(providers.Descriptor{Name: "github"}, func(providers.Options) (providers.Adapter, error) {
		return stubAdapter{}, nil
	})
	service, _ := NewProviderConnectionService(repository, registry, config.ProvidersConfig{EncryptionKey: "test-key"}, nil)
	ctx := context.Background()
	if _, err := service.Upsert(ctx, "user-1", ConnectionInput{Provider: "github", ProviderAccountID: "42", AccessToken: "a"}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if _, err := service.Upsert(ctx, "user-2", ConnectionInput{Provider: "github", ProviderAccountID: "42", AccessToken: "b"}); !errors.Is(err, utils.ErrProviderConnectionExists) {
		t.Fatalf("expected the second account to be rejected, got %v", err)
	}
}

func TestDisconnectWipesTokens(t *testing.T) {
	t.Parallel()

	repository := newFakeConnectionRepo()
	registry := providers.NewRegistry()
	registry.Register(providers.Descriptor{Name: "github"}, func(providers.Options) (providers.Adapter, error) {
		return stubAdapter{}, nil
	})
	service, _ := NewProviderConnectionService(repository, registry, config.ProvidersConfig{EncryptionKey: "test-key"}, nil)
	ctx := context.Background()
	if _, err := service.Upsert(ctx, "user-1", ConnectionInput{Provider: "github", ProviderAccountID: "42", AccessToken: "a"}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := service.Disconnect(ctx, "user-1", "github"); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	connection, err := repository.GetByProviderAccount(ctx, "github", "42")
	if err != nil {
		t.Fatalf("connection: %v", err)
	}
	if connection.Status != models.ConnectionRevoked || connection.AccessTokenEnc != "" || connection.RefreshTokenEnc != "" {
		t.Fatalf("expected the credentials to be wiped: %+v", connection)
	}
}

// ── Owners ───────────────────────────────────────────────────────────────────

func TestOwnerResolutionSharedNamespace(t *testing.T) {
	t.Parallel()

	users := newFakeUserRepo()
	orgs := newFakeOrgRepo()
	orgMembers := newFakeOrgMemberRepo()
	projects := newFakeProjectRepo()
	authorizer := NewAuthorizer(projects, newFakeProjectMemberRepo(), orgs, orgMembers, newFakeOrgTeamMemberRepo())
	service := NewOwnerService(
		users, orgs, orgMembers, newFakeOrgTeamRepo(), projects, newFakeBindingRepo(), newFakeProjectStarRepo(), authorizer,
	)

	username := "codegouvaor"
	users.add(models.User{Common: models.Common{ID: "user-1"}, Email: "a@example.gouv.aor", DisplayName: "Code", Username: &username})
	owner, err := service.Resolve(context.Background(), "", "codegouvaor")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if owner.Type != "user" {
		t.Fatalf("expected a user owner, got %q", owner.Type)
	}

	// An organization cannot reuse a username.
	available, err := service.IsNameAvailable(context.Background(), "codegouvaor")
	if err != nil || available {
		t.Fatalf("expected the name to be taken: %v %v", available, err)
	}

	orgs.add(models.Organization{Slug: "etat", Name: "État", OwnerID: "user-2", Visibility: "public"})
	organizationOwner, err := service.Resolve(context.Background(), "", "etat")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if organizationOwner.Type != "organization" || !organizationOwner.Capabilities.Teams {
		t.Fatalf("unexpected organization owner: %+v", organizationOwner)
	}
}

func TestOwnerNameValidation(t *testing.T) {
	t.Parallel()

	invalid := []string{"", "a", "-leading", "trailing-", "Upper Case", "two--hyphens", "api"}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Fatalf("expected %q to be rejected", name)
		}
	}
	for _, name := range []string{"codegouvaor", "react-ads", "code-2"} {
		if err := ValidateName(name); err != nil {
			t.Fatalf("expected %q to be valid: %v", name, err)
		}
	}
}

func TestClaimUsernameRejectsOrganizationSlug(t *testing.T) {
	t.Parallel()

	users := newFakeUserRepo()
	orgs := newFakeOrgRepo()
	orgMembers := newFakeOrgMemberRepo()
	projects := newFakeProjectRepo()
	authorizer := NewAuthorizer(projects, newFakeProjectMemberRepo(), orgs, orgMembers, newFakeOrgTeamMemberRepo())
	service := NewOwnerService(
		users, orgs, orgMembers, newFakeOrgTeamRepo(), projects, newFakeBindingRepo(), newFakeProjectStarRepo(), authorizer,
	)
	users.add(models.User{Common: models.Common{ID: "user-1"}, Email: "a@example.gouv.aor", DisplayName: "Code"})
	orgs.add(models.Organization{Slug: "etat", Name: "État", OwnerID: "user-2", Visibility: "public"})

	if _, err := service.ClaimUsername(context.Background(), "user-1", "etat"); !errors.Is(err, utils.ErrUsernameTaken) {
		t.Fatalf("expected the slug to be taken, got %v", err)
	}
	user, err := service.ClaimUsername(context.Background(), "user-1", "CodeGouvaor")
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if user.Username == nil || *user.Username != "codegouvaor" {
		t.Fatalf("unexpected username: %+v", user.Username)
	}
}
