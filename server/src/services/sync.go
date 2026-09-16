package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
)

// SyncService connects Code projects to external forges, imports their
// resources into the local read model and processes webhook deliveries. Every
// operation is idempotent and replayable: the same delivery can be received
// twice, webhooks can arrive out of order and a forge can disappear without
// corrupting the platform state.
type SyncService struct {
	projects      interfaces.ProjectRepository
	bindings      interfaces.RepositoryBindingRepository
	resources     interfaces.ExternalResourceRepository
	cursors       interfaces.SyncCursorRepository
	subscriptions interfaces.WebhookSubscriptionRepository
	users         interfaces.UserRepository
	connections   *ProviderConnectionService
	registry      *providers.Registry
	events        interfaces.EventBus
	jobs          *JobService
	auth          *Authorizer
	logger        *slog.Logger
}

// NewSyncService builds the synchronisation service.
func NewSyncService(
	projects interfaces.ProjectRepository,
	bindings interfaces.RepositoryBindingRepository,
	resources interfaces.ExternalResourceRepository,
	cursors interfaces.SyncCursorRepository,
	subscriptions interfaces.WebhookSubscriptionRepository,
	users interfaces.UserRepository,
	connections *ProviderConnectionService,
	events interfaces.EventBus,
	jobs *JobService,
	auth *Authorizer,
	logger *slog.Logger,
) *SyncService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SyncService{
		projects:      projects,
		bindings:      bindings,
		resources:     resources,
		cursors:       cursors,
		subscriptions: subscriptions,
		users:         users,
		connections:   connections,
		registry:      connections.Registry(),
		events:        events,
		jobs:          jobs,
		auth:          auth,
		logger:        logger,
	}
}

// RegisterHandlers wires the job kinds into the worker.
func (s *SyncService) RegisterHandlers() {
	if s.jobs == nil {
		return
	}
	s.jobs.Register(models.SyncJobRepositorySync, s.handleRepositorySync)
	s.jobs.Register(models.SyncJobRepositoryImport, s.handleRepositoryImport)
	s.jobs.Register(models.SyncJobWebhookProcess, s.handleWebhookProcess)
	s.jobs.Register(models.SyncJobReconciliation, s.handleReconciliation)
}

// ── Bindings ─────────────────────────────────────────────────────────────────

// ConnectBindingInput attaches an external repository to a Code project.
type ConnectBindingInput struct {
	Provider string
	Owner    string
	Repo     string
	// ConnectionID optionally pins the credential used for synchronisation.
	ConnectionID string
	IsPrimary    bool
}

// ConnectBinding verifies the repository on the provider then records the
// binding and schedules the initial import.
func (s *SyncService) ConnectBinding(ctx context.Context, principal interfaces.Principal, projectRef string, input ConnectBindingInput) (*models.RepositoryBinding, error) {
	project, err := s.resolveProject(ctx, projectRef)
	if err != nil {
		return nil, err
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if !s.registry.Has(provider) {
		return nil, utils.ErrProviderNotSupported
	}
	if input.Owner == "" || input.Repo == "" {
		return nil, utils.ErrValidationFailed
	}

	adapter, err := s.adapterFor(ctx, provider, principal.UserID, input.ConnectionID)
	if err != nil {
		return nil, err
	}
	reader, err := providers.RepositoryOf(adapter)
	if err != nil {
		return nil, err
	}
	ref := providers.RepositoryRef{Owner: input.Owner, Repo: strings.TrimSuffix(input.Repo, ".git")}
	remote, err := reader.Repository(ctx, ref)
	if err != nil {
		return nil, providers.AsAppError(err)
	}

	if existing, existingErr := s.bindings.GetByProjectProviderAndExternalID(ctx, project.ID, provider, remote.ID); existingErr == nil && existing != nil && existing.ID != "" {
		return nil, utils.ErrRepositoryBindingExists
	}

	now := time.Now().UTC()
	binding := &models.RepositoryBinding{
		Common:           models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID:        project.ID,
		Provider:         provider,
		ExternalID:       remote.ID,
		ExternalURL:      remote.URL,
		ExternalOwner:    remote.OwnerLogin,
		ExternalRepo:     remote.Name,
		DefaultBranch:    remote.DefaultBranch,
		IsPrimary:        input.IsPrimary,
		SyncEnabled:      true,
		SyncStatus:       models.BindingSyncPending,
		ExternalArchived: remote.Archived,
	}
	if binding.ExternalOwner == "" {
		binding.ExternalOwner = input.Owner
	}
	if binding.ExternalRepo == "" {
		binding.ExternalRepo = input.Repo
	}
	if updated := remote.UpdatedAt; !updated.IsZero() {
		binding.ExternalUpdated = &updated
	}
	if input.ConnectionID != "" {
		binding.ConnectionID = &input.ConnectionID
	}

	existing, listErr := s.bindings.ListByProject(ctx, project.ID)
	if listErr != nil {
		return nil, listErr
	}
	if len(existing) == 0 {
		binding.IsPrimary = true
	}
	if err := s.bindings.Create(ctx, binding); err != nil {
		return nil, err
	}

	// The Code project now mirrors a real repository: adopt its topics when the
	// project has none and remember which provider answers for it.
	if len(decodeStringSlice(project.Topics)) == 0 && len(remote.Topics) > 0 {
		project.Topics = encodeStringSlice(remote.Topics)
	}
	project.DefaultProvider = provider
	project.UpdatedAt = now
	if err := s.projects.Update(ctx, project); err != nil {
		s.logger.Warn("project update after binding failed", "project_id", project.ID, "error", err)
	}

	s.publish(ctx, "repository.connected", principal.UserID, map[string]any{
		"projectId": project.ID,
		"bindingId": binding.ID,
		"provider":  provider,
		"repo":      binding.ExternalOwner + "/" + binding.ExternalRepo,
	})

	if s.jobs != nil {
		projectID := project.ID
		bindingID := binding.ID
		if _, err := s.jobs.Submit(ctx, EnqueueInput{
			Kind:           models.SyncJobRepositoryImport,
			IdempotencyKey: fmt.Sprintf("import:%s", binding.ID),
			ProjectID:      &projectID,
			BindingID:      &bindingID,
			Payload:        map[string]any{"provider": provider},
		}); err != nil {
			s.logger.Warn("initial import not scheduled", "binding_id", binding.ID, "error", err)
		}
	}
	return binding, nil
}

// SyncBinding refreshes the metadata of a binding right away. It is used both
// by the manual sync endpoint and by the background job.
func (s *SyncService) SyncBinding(ctx context.Context, bindingID string) (*models.RepositoryBinding, error) {
	binding, err := s.bindings.GetByID(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	project, err := s.projects.GetByID(ctx, binding.ProjectID)
	if err != nil {
		return nil, err
	}
	adapter, err := s.adapterFor(ctx, binding.Provider, project.OwnerID, derefString(binding.ConnectionID))
	if err != nil {
		return nil, err
	}
	reader, err := providers.RepositoryOf(adapter)
	if err != nil {
		return nil, err
	}
	remote, err := reader.Repository(ctx, providers.RepositoryRef{Owner: binding.ExternalOwner, Repo: binding.ExternalRepo})
	if err != nil {
		// The repository disappeared from the forge: detach instead of
		// retrying forever.
		if providers.IsNotFound(err) {
			return s.detach(ctx, binding, "the repository no longer exists on the provider")
		}
		s.markSyncError(ctx, binding, err)
		return nil, providers.AsAppError(err)
	}

	now := time.Now().UTC()
	binding.ExternalID = remote.ID
	binding.ExternalURL = remote.URL
	binding.DefaultBranch = remote.DefaultBranch
	binding.ExternalArchived = remote.Archived
	binding.SyncStatus = models.BindingSyncSynced
	binding.LastSyncError = ""
	binding.LastSyncedAt = &now
	binding.UpdatedAt = now
	if updated := remote.UpdatedAt; !updated.IsZero() {
		binding.ExternalUpdated = &updated
	}
	if err := s.bindings.Update(ctx, binding); err != nil {
		return nil, err
	}
	if err := s.advanceCursor(ctx, binding, "repository", "", remote.UpdatedAt, remote.UpdatedAt, ""); err != nil {
		s.logger.Warn("cursor update failed", "binding_id", binding.ID, "error", err)
	}
	s.publish(ctx, "repository.synced", "", map[string]any{
		"projectId": binding.ProjectID,
		"bindingId": binding.ID,
		"provider":  binding.Provider,
		"status":    binding.SyncStatus,
	})
	return binding, nil
}

// DisconnectBinding detaches a repository from a project.
func (s *SyncService) DisconnectBinding(ctx context.Context, principal interfaces.Principal, projectRef, bindingID string) error {
	project, err := s.resolveProject(ctx, projectRef)
	if err != nil {
		return err
	}
	binding, err := s.bindings.GetByID(ctx, bindingID)
	if err != nil {
		return err
	}
	if binding.ProjectID != project.ID {
		return utils.ErrRepositoryBindingNotFound
	}
	if _, err := s.requireRepositorySync(ctx, principal, project); err != nil {
		return err
	}
	if _, err := s.detach(ctx, binding, "detached by "+principal.UserID); err != nil {
		return err
	}
	return nil
}

// ListBindings returns the bindings of a project.
func (s *SyncService) ListBindings(ctx context.Context, principal interfaces.Principal, projectRef string) ([]models.RepositoryBinding, error) {
	project, err := s.resolveProject(ctx, projectRef)
	if err != nil {
		return nil, err
	}
	if _, err := s.authzProject(ctx, principal, project, ActionProjectRead); err != nil {
		return nil, err
	}
	return s.bindings.ListByProject(ctx, project.ID)
}

// Reconcile scans the bindings that have not been synchronised for a while and
// schedules one job per stale binding.
func (s *SyncService) Reconcile(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	staleBefore := time.Now().UTC().Add(-time.Hour)
	bindings, err := s.bindings.ListSyncable(ctx, staleBefore, limit)
	if err != nil {
		return 0, err
	}
	scheduled := 0
	for index := range bindings {
		binding := bindings[index]
		if s.jobs == nil {
			break
		}
		projectID := binding.ProjectID
		bindingID := binding.ID
		bucket := time.Now().UTC().Truncate(30 * time.Minute).Format("20060102T1504")
		if _, err := s.jobs.Submit(ctx, EnqueueInput{
			Kind:           models.SyncJobRepositorySync,
			IdempotencyKey: fmt.Sprintf("reconcile:%s:%s", binding.ID, bucket),
			ProjectID:      &projectID,
			BindingID:      &bindingID,
		}); err != nil {
			return scheduled, err
		}
		scheduled++
	}
	return scheduled, nil
}

// ── Webhooks ─────────────────────────────────────────────────────────────────

// WebhookOutcome describes what happened to a delivery.
type WebhookOutcome struct {
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
	Stale     bool   `json:"stale"`
	Event     string `json:"event"`
	JobID     string `json:"jobId,omitempty"`
	BindingID string `json:"bindingId"`
	ProjectID string `json:"projectId"`
}

// HandleWebhook verifies, deduplicates and enqueues a provider delivery. The
// HTTP handler stays thin: everything that matters happens here and in the
// job worker.
func (s *SyncService) HandleWebhook(ctx context.Context, provider, bindingRef string, headers http.Header, body []byte) (*WebhookOutcome, error) {
	name := strings.ToLower(strings.TrimSpace(provider))
	if !s.registry.Has(name) {
		return nil, utils.ErrProviderNotSupported
	}
	binding, err := s.bindings.GetByID(ctx, bindingRef)
	if err != nil {
		return nil, utils.ErrRepositoryBindingNotFound
	}
	if binding.Provider != name {
		return nil, utils.ErrProviderNotSupported
	}

	deliveryID := deliveryIDOf(name, headers)
	if err := s.verifyWebhook(ctx, binding, name, headers, body); err != nil {
		return nil, err
	}

	event, eventAt, resource, err := s.decodeWebhook(name, binding, headers, body)
	if err != nil {
		return nil, err
	}

	// Duplicated delivery: the same event must never be applied twice.
	if deliveryID != "" && s.isDuplicate(ctx, binding, deliveryID, resource) {
		s.logger.Info("duplicate webhook ignored", "provider", name, "binding_id", binding.ID, "delivery_id", deliveryID)
		return &WebhookOutcome{Accepted: true, Duplicate: true, Event: event, BindingID: binding.ID, ProjectID: binding.ProjectID}, nil
	}

	stale := s.isStale(ctx, binding, resource, eventAt)

	if resource != nil {
		if err := s.resources.Upsert(ctx, resource); err != nil {
			return nil, err
		}
	}
	// The delivery identifier is always recorded on the "webhook" cursor so
	// that a replayed delivery is detected whatever its payload is. The
	// resource cursor additionally tracks the freshness of the read model.
	if err := s.advanceCursor(ctx, binding, "webhook", deliveryID, eventAt, eventAt, ""); err != nil {
		s.logger.Warn("webhook cursor update failed", "binding_id", binding.ID, "error", err)
	}
	if resource != nil {
		if err := s.advanceCursor(ctx, binding, resourceKind(resource), deliveryID, eventAt, eventAt, ""); err != nil {
			s.logger.Warn("resource cursor update failed", "binding_id", binding.ID, "error", err)
		}
	}
	s.recordDelivery(ctx, binding, deliveryID, resource, eventAt)

	// "repository deleted" is terminal: detach instead of scheduling a sync.
	if strings.HasPrefix(event, "repository.deleted") {
		if _, err := s.detach(ctx, binding, "the repository was deleted on the provider"); err != nil {
			return nil, err
		}
		return &WebhookOutcome{Accepted: true, Event: event, BindingID: binding.ID, ProjectID: binding.ProjectID}, nil
	}

	outcome := &WebhookOutcome{Accepted: true, Event: event, Stale: stale, BindingID: binding.ID, ProjectID: binding.ProjectID}
	if s.jobs != nil {
		key := fmt.Sprintf("webhook:%s:%s", binding.ID, deliveryID)
		if deliveryID == "" {
			key = fmt.Sprintf("webhook:%s:%s:%d", binding.ID, event, time.Now().UTC().UnixNano())
		}
		projectID := binding.ProjectID
		bindingID := binding.ID
		job, err := s.jobs.Submit(ctx, EnqueueInput{
			Kind:           models.SyncJobWebhookProcess,
			IdempotencyKey: key,
			ProjectID:      &projectID,
			BindingID:      &bindingID,
			Payload: map[string]any{
				"provider": name,
				"event":    event,
				"stale":    stale,
			},
		})
		if err != nil {
			return nil, err
		}
		outcome.JobID = job.ID
	}

	s.publish(ctx, event, "", map[string]any{
		"projectId": binding.ProjectID,
		"bindingId": binding.ID,
		"provider":  name,
		"stale":     stale,
	})
	return outcome, nil
}

// verifyWebhook checks the signature/token of a delivery against the secret
// recorded for the binding. An unknown or missing secret is a hard failure: a
// spoofed webhook could otherwise rewrite the read model.
func (s *SyncService) verifyWebhook(ctx context.Context, binding *models.RepositoryBinding, provider string, headers http.Header, body []byte) error {
	secret := ""
	if subscription, err := s.subscriptions.GetByBindingAndExternal(ctx, binding.ID, binding.ExternalID); err == nil && subscription != nil && subscription.SecretEnc != "" {
		if s.connections.box != nil {
			if decoded, decodeErr := s.connections.box.Open(subscription.SecretEnc); decodeErr == nil {
				secret = decoded
			}
		}
	}
	if secret == "" {
		secret = s.connections.serviceToken(provider)
	}
	if secret == "" {
		return utils.ErrWebhookSignatureInvalid
	}
	switch provider {
	case models.ProviderGitHub, models.ProviderGiteria:
		signature := strings.TrimSpace(headers.Get("X-Hub-Signature-256"))
		if signature == "" {
			signature = strings.TrimSpace(headers.Get("X-Gitea-Signature"))
		}
		if signature == "" {
			return utils.ErrWebhookSignatureInvalid
		}
		signature = strings.TrimPrefix(signature, "sha256=")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if subtle.ConstantTimeCompare([]byte(strings.ToLower(signature)), []byte(expected)) != 1 {
			return utils.ErrWebhookSignatureInvalid
		}
	case models.ProviderGitLab:
		token := headers.Get("X-Gitlab-Token")
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
			return utils.ErrWebhookSignatureInvalid
		}
	default:
		return utils.ErrWebhookSignatureInvalid
	}
	return nil
}

func deliveryIDOf(provider string, headers http.Header) string {
	switch provider {
	case models.ProviderGitHub:
		return headers.Get("X-GitHub-Delivery")
	case models.ProviderGitLab:
		return headers.Get("X-Gitlab-Event-UUID")
	case models.ProviderGiteria:
		if value := headers.Get("X-Gitea-Delivery"); value != "" {
			return value
		}
		return headers.Get("X-GitHub-Delivery")
	default:
		return ""
	}
}

// decodeWebhook normalises a delivery into a Code event, a timestamp and the
// optional read-model entry it carries.
func (s *SyncService) decodeWebhook(provider string, binding *models.RepositoryBinding, headers http.Header, body []byte) (string, time.Time, *models.ExternalResource, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", time.Time{}, nil, utils.ErrWebhookPayloadInvalid
	}
	eventHeader := ""
	switch provider {
	case models.ProviderGitHub:
		eventHeader = headers.Get("X-GitHub-Event")
	case models.ProviderGitLab:
		eventHeader = headers.Get("X-Gitlab-Event")
	case models.ProviderGiteria:
		// Gitea/Giteria send X-Gitea-Event on recent versions and keep the
		// GitHub-compatible header on older ones.
		eventHeader = headers.Get("X-Gitea-Event")
		if eventHeader == "" {
			eventHeader = headers.Get("X-GitHub-Event")
		}
	}
	event := normalizeEvent(provider, eventHeader, payload)
	resource := s.resourceFromWebhook(provider, event, payload, binding)
	// Payloads often only carry the timestamp of the nested object (an issue, a
	// merge request, a release): fall back to it so out-of-order deliveries can
	// still be detected.
	eventAt := extractTime(payload)
	if eventAt.IsZero() && resource != nil && resource.ExternalUpdatedAt != nil {
		eventAt = *resource.ExternalUpdatedAt
	}
	return event, eventAt, resource, nil
}

// resourceFromWebhook extracts the read-model entry of an issue, review or
// release event. Other event kinds only trigger a synchronisation.
func (s *SyncService) resourceFromWebhook(provider, event string, payload map[string]any, binding *models.RepositoryBinding) *models.ExternalResource {
	kind := ""
	switch {
	case strings.HasPrefix(event, "issue."):
		kind = models.ExternalResourceIssue
	case strings.HasPrefix(event, "review."):
		kind = models.ExternalResourceReview
	case strings.HasPrefix(event, "release."):
		kind = models.ExternalResourceRelease
	default:
		return nil
	}

	object := map[string]any(nil)
	for _, key := range []string{"issue", "pull_request", "release", "object_attributes"} {
		if candidate, ok := payload[key].(map[string]any); ok {
			object = candidate
			break
		}
	}
	if object == nil {
		return nil
	}

	externalID := externalIDOf(object)
	if externalID == "" {
		return nil
	}
	now := time.Now().UTC()
	resource := &models.ExternalResource{
		Common:      models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		BindingID:   binding.ID,
		ProjectID:   binding.ProjectID,
		Provider:    provider,
		Kind:        kind,
		ExternalID:  externalID,
		Number:      intOf(object, "number", "iid"),
		Title:       firstNonEmpty(stringOf(object["title"]), stringOf(object["name"]), stringOf(object["tag_name"])),
		State:       stateOf(object),
		URL:         firstNonEmpty(stringOf(object["html_url"]), stringOf(object["url"])),
		AuthorLogin: authorLoginOf(object),
		SyncedAt:    now,
	}
	if updatedAt := extractTime(object); !updatedAt.IsZero() {
		resource.ExternalUpdatedAt = &updatedAt
	} else if !extractTime(payload).IsZero() {
		value := extractTime(payload)
		resource.ExternalUpdatedAt = &value
	}
	return resource
}

func externalIDOf(object map[string]any) string {
	if id := stringOf(object["id"]); id != "" {
		return id
	}
	if id, ok := object["id"].(float64); ok {
		return fmt.Sprintf("%.0f", id)
	}
	if tag := stringOf(object["tag_name"]); tag != "" {
		return tag
	}
	return ""
}

func intOf(object map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := object[key].(type) {
		case float64:
			return int(value)
		case int:
			return value
		}
	}
	return 0
}

func stateOf(object map[string]any) string {
	if merged, ok := object["merged"].(bool); ok && merged {
		return "merged"
	}
	return firstNonEmpty(stringOf(object["state"]), stringOf(object["action"]), "updated")
}

func authorLoginOf(object map[string]any) string {
	if user, ok := object["user"].(map[string]any); ok {
		if login := stringOf(user["login"]); login != "" {
			return login
		}
	}
	if author, ok := object["author"].(map[string]any); ok {
		if login := firstNonEmpty(stringOf(author["login"]), stringOf(author["username"])); login != "" {
			return login
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// ── Jobs ─────────────────────────────────────────────────────────────────────

func (s *SyncService) handleRepositorySync(ctx context.Context, job *models.SyncJob) error {
	if job.BindingID == nil || *job.BindingID == "" {
		return utils.ErrValidationFailed
	}
	_, err := s.SyncBinding(ctx, *job.BindingID)
	return err
}

func (s *SyncService) handleRepositoryImport(ctx context.Context, job *models.SyncJob) error {
	if job.BindingID == nil || *job.BindingID == "" {
		return utils.ErrValidationFailed
	}
	binding, err := s.SyncBinding(ctx, *job.BindingID)
	if err != nil {
		return err
	}
	imported, err := s.ImportResources(ctx, binding)
	if err != nil {
		return err
	}
	s.logger.Info("repository import finished", "binding_id", binding.ID, "resources", imported)
	return nil
}

func (s *SyncService) handleWebhookProcess(ctx context.Context, job *models.SyncJob) error {
	if job.BindingID == nil || *job.BindingID == "" {
		return utils.ErrValidationFailed
	}
	payload := decodeMap(job.Payload)
	if stale, ok := payload["stale"].(bool); ok && stale {
		// Out-of-order delivery: the read model is already up to date, nothing
		// else to do besides acknowledging the job.
		s.logger.Info("stale webhook acknowledged", "job_id", job.ID, "binding_id", *job.BindingID)
		return nil
	}
	_, err := s.SyncBinding(ctx, *job.BindingID)
	return err
}

func (s *SyncService) handleReconciliation(ctx context.Context, job *models.SyncJob) error {
	scheduled, err := s.Reconcile(ctx, 25)
	if err != nil {
		return err
	}
	s.logger.Info("reconciliation finished", "scheduled", scheduled)
	return nil
}

// ImportResources pulls issues, reviews and releases into the read model.
func (s *SyncService) ImportResources(ctx context.Context, binding *models.RepositoryBinding) (int, error) {
	project, err := s.projects.GetByID(ctx, binding.ProjectID)
	if err != nil {
		return 0, err
	}
	adapter, err := s.adapterFor(ctx, binding.Provider, project.OwnerID, derefString(binding.ConnectionID))
	if err != nil {
		return 0, err
	}
	ref := providers.RepositoryRef{Owner: binding.ExternalOwner, Repo: binding.ExternalRepo}
	imported := 0

	if capability := s.registry.Capabilities(binding.Provider)[providers.CapabilityIssues]; capability != providers.CapabilityUnavailable {
		if reader, readerErr := providers.IssuesOf(adapter); readerErr == nil {
			issues, listErr := reader.Issues(ctx, providers.ListQuery{Ref: ref, State: "all", Limit: 50})
			if listErr != nil {
				return imported, providers.AsAppError(listErr)
			}
			for _, issue := range issues {
				if err := s.upsertResource(ctx, binding, models.ExternalResourceIssue, issue.ID, issue.Number, issue.Title, issue.State, issue.URL, issue.AuthorLogin, issue.UpdatedAt); err != nil {
					return imported, err
				}
				imported++
			}
		}
	}
	if capability := s.registry.Capabilities(binding.Provider)[providers.CapabilityReviews]; capability != providers.CapabilityUnavailable {
		if reader, readerErr := providers.ReviewsOf(adapter); readerErr == nil {
			reviews, listErr := reader.Reviews(ctx, providers.ListQuery{Ref: ref, State: "all", Limit: 50})
			if listErr != nil {
				return imported, providers.AsAppError(listErr)
			}
			for _, review := range reviews {
				state := review.State
				if review.Merged {
					state = "merged"
				}
				if err := s.upsertResource(ctx, binding, models.ExternalResourceReview, review.ID, review.Number, review.Title, state, review.URL, review.AuthorLogin, review.UpdatedAt); err != nil {
					return imported, err
				}
				imported++
			}
		}
	}
	if capability := s.registry.Capabilities(binding.Provider)[providers.CapabilityReleases]; capability != providers.CapabilityUnavailable {
		if reader, readerErr := providers.ReleasesOf(adapter); readerErr == nil {
			releases, listErr := reader.Releases(ctx, providers.ListQuery{Ref: ref, Limit: 50})
			if listErr != nil {
				return imported, providers.AsAppError(listErr)
			}
			for _, release := range releases {
				id := release.ID
				if id == "" {
					id = release.TagName
				}
				if err := s.upsertResource(ctx, binding, models.ExternalResourceRelease, id, 0, release.Name, "published", release.URL, release.AuthorLogin, release.PublishedAt); err != nil {
					return imported, err
				}
				imported++
			}
		}
	}
	return imported, nil
}

func (s *SyncService) upsertResource(
	ctx context.Context,
	binding *models.RepositoryBinding,
	kind, externalID string,
	number int,
	title, state, url, author string,
	externalUpdatedAt time.Time,
) error {
	now := time.Now().UTC()
	resource := &models.ExternalResource{
		Common:      models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		BindingID:   binding.ID,
		ProjectID:   binding.ProjectID,
		Provider:    binding.Provider,
		Kind:        kind,
		ExternalID:  externalID,
		Number:      number,
		Title:       title,
		State:       state,
		URL:         url,
		AuthorLogin: author,
		SyncedAt:    now,
	}
	if !externalUpdatedAt.IsZero() {
		resource.ExternalUpdatedAt = &externalUpdatedAt
	}
	return s.resources.Upsert(ctx, resource)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (s *SyncService) detach(ctx context.Context, binding *models.RepositoryBinding, reason string) (*models.RepositoryBinding, error) {
	now := time.Now().UTC()
	binding.SyncStatus = models.BindingSyncDetached
	binding.SyncEnabled = false
	binding.LastSyncError = reason
	binding.DetachedAt = &now
	binding.UpdatedAt = now
	if err := s.bindings.Update(ctx, binding); err != nil {
		return nil, err
	}
	if err := s.resources.DeleteByBinding(ctx, binding.ID); err != nil {
		s.logger.Warn("read model cleanup failed", "binding_id", binding.ID, "error", err)
	}
	s.publish(ctx, "repository.detached", "", map[string]any{
		"projectId": binding.ProjectID,
		"bindingId": binding.ID,
		"provider":  binding.Provider,
		"reason":    reason,
	})
	return binding, nil
}

func (s *SyncService) markSyncError(ctx context.Context, binding *models.RepositoryBinding, cause error) {
	now := time.Now().UTC()
	binding.SyncStatus = models.BindingSyncError
	binding.LastSyncError = cause.Error()
	binding.UpdatedAt = now
	if err := s.bindings.Update(ctx, binding); err != nil {
		s.logger.Warn("sync error not persisted", "binding_id", binding.ID, "error", err)
	}
}

func (s *SyncService) advanceCursor(ctx context.Context, binding *models.RepositoryBinding, resource, eventID string, eventAt time.Time, cursorAt time.Time, checksum string) error {
	if resource == "" {
		resource = "repository"
	}
	cursor, err := s.cursors.Get(ctx, binding.ID, resource)
	if err != nil {
		cursor = &models.SyncCursor{
			Common:    models.Common{ID: utils.NewID(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			BindingID: binding.ID,
			Provider:  binding.Provider,
			Resource:  resource,
		}
	}
	now := time.Now().UTC()
	// Cursors only move forward: an out-of-order event must never rewind the
	// synchronisation position.
	if cursor.LastEventAt == nil || eventAt.IsZero() || !eventAt.Before(*cursor.LastEventAt) {
		if !eventAt.IsZero() {
			value := eventAt.UTC()
			cursor.LastEventAt = &value
		}
		if eventID != "" {
			cursor.LastEventID = eventID
		}
		if !cursorAt.IsZero() {
			cursor.Cursor = cursorAt.UTC().Format(time.RFC3339Nano)
		}
		if checksum != "" {
			cursor.Checksum = checksum
		}
	}
	cursor.LastEventKind = resource
	cursor.UpdatedAt = now
	return s.cursors.Upsert(ctx, cursor)
}

func (s *SyncService) isDuplicate(ctx context.Context, binding *models.RepositoryBinding, deliveryID string, resource *models.ExternalResource) bool {
	subscription, err := s.subscriptions.GetByBindingAndExternal(ctx, binding.ID, binding.ExternalID)
	if err == nil && subscription != nil && subscription.LastDeliveryID == deliveryID {
		return true
	}
	for _, name := range []string{"webhook", resourceKind(resource)} {
		if name == "" {
			continue
		}
		cursor, err := s.cursors.Get(ctx, binding.ID, name)
		if err == nil && cursor.LastEventID == deliveryID {
			return true
		}
	}
	return false
}

func (s *SyncService) isStale(ctx context.Context, binding *models.RepositoryBinding, resource *models.ExternalResource, eventAt time.Time) bool {
	if resource == nil || eventAt.IsZero() {
		return false
	}
	cursor, err := s.cursors.Get(ctx, binding.ID, resourceKind(resource))
	if err != nil || cursor.LastEventAt == nil {
		return false
	}
	return eventAt.Before(*cursor.LastEventAt)
}

func (s *SyncService) recordDelivery(ctx context.Context, binding *models.RepositoryBinding, deliveryID string, resource *models.ExternalResource, eventAt time.Time) {
	subscription, err := s.subscriptions.GetByBindingAndExternal(ctx, binding.ID, binding.ExternalID)
	if err != nil || subscription == nil {
		return
	}
	now := time.Now().UTC()
	subscription.LastDeliveryID = deliveryID
	subscription.LastReceivedAt = &now
	subscription.UpdatedAt = now
	if err := s.subscriptions.Update(ctx, subscription); err != nil {
		s.logger.Warn("webhook subscription not updated", "binding_id", binding.ID, "error", err)
	}
}

func (s *SyncService) adapterFor(ctx context.Context, provider, ownerUserID, connectionID string) (providers.Adapter, error) {
	if connectionID != "" {
		connection, err := s.connections.connections.GetByID(ctx, connectionID)
		if err != nil {
			return nil, utils.ErrProviderConnectionNotFound
		}
		if connection.Provider != provider {
			return nil, utils.ErrOAuthProviderMismatch
		}
		return s.connections.AdapterForConnection(connection)
	}
	return s.connections.Adapter(ctx, ownerUserID, provider)
}

func (s *SyncService) resolveProject(ctx context.Context, ref string) (*models.Project, error) {
	project, err := s.projects.GetByID(ctx, ref)
	if err == nil && project.ID != "" {
		return project, nil
	}
	return s.projects.GetByReference(ctx, ref)
}

func (s *SyncService) requireRepositorySync(ctx context.Context, principal interfaces.Principal, project *models.Project) (ProjectAccess, error) {
	return s.authzProject(ctx, principal, project, ActionRepositorySync)
}

func (s *SyncService) authzProject(ctx context.Context, principal interfaces.Principal, project *models.Project, action Action) (ProjectAccess, error) {
	if s.auth == nil {
		return ProjectAccess{}, utils.ErrForbidden
	}
	return s.auth.AuthorizeProject(ctx, principal, project, action)
}

func (s *SyncService) publish(ctx context.Context, eventType, actorID string, payload map[string]any) {
	if s.events == nil {
		return
	}
	_ = s.events.Publish(ctx, interfaces.Event{
		ID:        utils.NewID(),
		Topic:     eventType,
		Type:      eventType,
		ActorID:   actorID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Payload:   payload,
	})
}

// resourceKind returns the cursor resource name of a read-model entry.
func resourceKind(resource *models.ExternalResource) string {
	if resource == nil {
		return ""
	}
	return resource.Kind
}

// normalizeEvent maps provider-specific event names to Code event names.
func normalizeEvent(provider, header string, payload map[string]any) string {
	action := strings.ToLower(stringOf(payload["action"]))
	switch provider {
	case models.ProviderGitHub, models.ProviderGiteria:
		switch strings.ToLower(header) {
		case "issues":
			return "issue." + action
		case "pull_request":
			return "review." + action
		case "release":
			return "release." + action
		case "push":
			return "repository.pushed"
		case "repository":
			return "repository." + action
		default:
			if action != "" {
				return strings.ToLower(header) + "." + action
			}
			return strings.ToLower(header)
		}
	case models.ProviderGitLab:
		kind := strings.ToLower(stringOf(payload["object_kind"]))
		switch kind {
		case "issue":
			return "issue." + gitlabAction(action)
		case "merge_request":
			return "review." + gitlabAction(action)
		case "push":
			return "repository.pushed"
		case "release":
			return "release.created"
		default:
			return kind + "." + gitlabAction(action)
		}
	default:
		return strings.ToLower(header)
	}
}

func gitlabAction(action string) string {
	switch action {
	case "open":
		return "created"
	case "update":
		return "updated"
	case "close":
		return "closed"
	case "merge":
		return "merged"
	case "reopen":
		return "reopened"
	case "approved":
		return "approved"
	default:
		return action
	}
}

// extractTime finds the most relevant timestamp in a webhook payload.
func extractTime(payload map[string]any) time.Time {
	for _, key := range []string{"updated_at", "pushed_at", "created_at", "timestamp"} {
		if value := stringOf(payload[key]); value != "" {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return parsed.UTC()
			}
			if parsed, err := time.Parse("2006-01-02T15:04:05Z0700", value); err == nil {
				return parsed.UTC()
			}
		}
	}
	if resource, ok := payload["object_attributes"].(map[string]any); ok {
		if value := stringOf(resource["updated_at"]); value != "" {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return parsed.UTC()
			}
		}
	}
	return time.Time{}
}

func stringOf(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
