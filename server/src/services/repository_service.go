package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	redisclient "github.com/codegouvaor/code/server/internal/redis"
	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
)

// ResolvedRepository is a Code project plus the repository binding that serves
// it on a provider. Handlers and services always start from this value: the
// frontend addresses "owner/repo" and the backend decides which forge answers.
type ResolvedRepository struct {
	Project    *models.Project
	Binding    *models.RepositoryBinding
	Provider   string
	Descriptor providers.Descriptor
	Access     ProjectAccess
}

// RepositoryService is the read facade over the provider layer. It resolves a
// Code reference, picks the right adapter, enforces capabilities and caches
// provider reads so that the platform stays responsive and rate-limit friendly.
type RepositoryService struct {
	projects    interfaces.ProjectRepository
	bindings    interfaces.RepositoryBindingRepository
	orgs        interfaces.OrganizationRepository
	users       interfaces.UserRepository
	resources   interfaces.ExternalResourceRepository
	connections *ProviderConnectionService
	registry    *providers.Registry
	auth        *Authorizer
	cache       redisclient.Cache
	cachePrefix string
	cacheTTL    time.Duration
	events      interfaces.EventBus
	logger      *slog.Logger
}

// NewRepositoryService builds the repository service.
func NewRepositoryService(
	projects interfaces.ProjectRepository,
	bindings interfaces.RepositoryBindingRepository,
	orgs interfaces.OrganizationRepository,
	users interfaces.UserRepository,
	resources interfaces.ExternalResourceRepository,
	connections *ProviderConnectionService,
	auth *Authorizer,
	cache redisclient.Cache,
	cachePrefix string,
	cacheTTL time.Duration,
	events interfaces.EventBus,
	logger *slog.Logger,
) *RepositoryService {
	if cacheTTL <= 0 {
		cacheTTL = 2 * time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RepositoryService{
		projects:    projects,
		bindings:    bindings,
		orgs:        orgs,
		users:       users,
		resources:   resources,
		connections: connections,
		registry:    connections.Registry(),
		auth:        auth,
		cache:       cache,
		cachePrefix: cachePrefix,
		cacheTTL:    cacheTTL,
		events:      events,
		logger:      logger,
	}
}

// Resolve turns an owner/repo pair into a project, its binding and its access
// level. It falls back to the binding's external coordinates so that a
// repository can be explored before a Code project exists for it.
func (s *RepositoryService) Resolve(ctx context.Context, principal interfaces.Principal, owner, repo string) (*ResolvedRepository, error) {
	owner = strings.ToLower(strings.TrimSpace(owner))
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, utils.ErrProjectNotFound
	}

	project, err := s.projects.GetByReference(ctx, owner+"/"+repo)
	if err != nil && utils.AsAppError(err).Code != "PROJECT_NOT_FOUND" {
		return nil, err
	}
	if project == nil || project.ID == "" {
		// Not a Code reference: look for a bound repository instead.
		binding, bindingErr := s.findBindingByExternal(ctx, owner, repo)
		if bindingErr != nil {
			return nil, utils.ErrProjectNotFound
		}
		project, err = s.projects.GetByID(ctx, binding.ProjectID)
		if err != nil {
			return nil, err
		}
	}

	access, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectRead)
	if err != nil {
		return nil, err
	}

	resolved := &ResolvedRepository{Project: project, Access: access}
	bindings, err := s.bindings.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	if len(bindings) > 0 {
		primary := bindings[0]
		for _, binding := range bindings {
			if binding.IsPrimary && binding.ExternalRepo == repo {
				primary = binding
				break
			}
			if binding.ExternalRepo == repo && binding.ExternalOwner == owner {
				primary = binding
			}
		}
		binding := primary
		resolved.Binding = &binding
		resolved.Provider = binding.Provider
		if descriptor, ok := s.registry.Descriptor(binding.Provider); ok {
			resolved.Descriptor = descriptor
		}
	}
	return resolved, nil
}

// ResolveProject resolves a project-scoped repository reference.
func (s *RepositoryService) ResolveProject(ctx context.Context, principal interfaces.Principal, projectRef string) (*ResolvedRepository, error) {
	project, err := s.projects.GetByID(ctx, projectRef)
	if err != nil {
		if project, err = s.projects.GetByReference(ctx, projectRef); err != nil {
			return nil, err
		}
	}
	access, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectRead)
	if err != nil {
		return nil, err
	}
	resolved := &ResolvedRepository{Project: project, Access: access}
	bindings, err := s.bindings.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	if len(bindings) > 0 {
		primary := bindings[0]
		for _, binding := range bindings {
			if binding.IsPrimary {
				primary = binding
				break
			}
		}
		binding := primary
		resolved.Binding = &binding
		resolved.Provider = binding.Provider
		if descriptor, ok := s.registry.Descriptor(binding.Provider); ok {
			resolved.Descriptor = descriptor
		}
	}
	return resolved, nil
}

// Capabilities returns the capability map that applies to a resolved
// repository. A project without a bound repository reports every provider
// feature as unavailable but keeps the Code-native ones.
func (s *RepositoryService) Capabilities(resolved *ResolvedRepository) map[string]string {
	if resolved == nil || resolved.Provider == "" {
		return codeOnlyCapabilities()
	}
	return stringifyCapabilities(s.registry.Capabilities(resolved.Provider))
}

// Descriptor returns the provider descriptor of a resolved repository.
func (s *RepositoryService) Descriptor(resolved *ResolvedRepository) map[string]any {
	if resolved == nil || resolved.Provider == "" {
		return map[string]any{"name": "", "capabilities": s.Capabilities(resolved)}
	}
	return map[string]any{
		"name":             resolved.Descriptor.Name,
		"displayName":      resolved.Descriptor.DisplayName,
		"websiteUrl":       resolved.Descriptor.WebsiteURL,
		"documentationUrl": resolved.Descriptor.DocumentationURL,
		"capabilities":     s.Capabilities(resolved),
	}
}

// ── Reads ────────────────────────────────────────────────────────────────────

// Repository returns the provider metadata of a resolved repository.
func (s *RepositoryService) Repository(ctx context.Context, resolved *ResolvedRepository) (*providers.Repository, error) {
	if resolved.Binding == nil {
		return nil, utils.ErrRepositoryBindingNotFound
	}
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	reader, err := providers.RepositoryOf(adapter)
	if err != nil {
		return nil, err
	}
	ref := refOfBinding(resolved.Binding)
	var repository providers.Repository
	if err := s.cached(ctx, cacheKey(resolved), &repository, func() (any, error) {
		return reader.Repository(ctx, ref)
	}); err != nil {
		return nil, s.mapProviderError(err)
	}
	return &repository, nil
}

// Branches lists the branches of a resolved repository.
func (s *RepositoryService) Branches(ctx context.Context, resolved *ResolvedRepository, limit, page int) ([]providers.Branch, error) {
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapability(resolved, providers.CapabilityBranches, "branches"); err != nil {
		return nil, err
	}
	reader, err := providers.BranchesOf(adapter)
	if err != nil {
		return nil, err
	}
	branches, err := reader.Branches(ctx, providers.ListQuery{Ref: refOfBinding(resolved.Binding), Limit: limit, Page: page})
	if err != nil {
		return nil, s.mapProviderError(err)
	}
	return branches, nil
}

// Commits lists the commits of a resolved repository.
func (s *RepositoryService) Commits(ctx context.Context, resolved *ResolvedRepository, refName, path string, limit, page int) ([]providers.Commit, error) {
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapability(resolved, providers.CapabilityCommits, "commits"); err != nil {
		return nil, err
	}
	reader, err := providers.CommitsOf(adapter)
	if err != nil {
		return nil, err
	}
	commits, err := reader.Commits(ctx, providers.CommitQuery{
		Ref: refOfBinding(resolved.Binding), RefName: refName, Path: path, Limit: limit, Page: page,
	})
	if err != nil {
		return nil, s.mapProviderError(err)
	}
	return commits, nil
}

// Tree lists a directory of a resolved repository.
func (s *RepositoryService) Tree(ctx context.Context, resolved *ResolvedRepository, path, refName string) ([]providers.TreeEntry, error) {
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapability(resolved, providers.CapabilityFiles, "files"); err != nil {
		return nil, err
	}
	reader, err := providers.FilesOf(adapter)
	if err != nil {
		return nil, err
	}
	entries, err := reader.Tree(ctx, providers.TreeQuery{
		Ref: refOfBinding(resolved.Binding), Path: path, RefName: refName, MaxEntries: 100,
	})
	if err != nil {
		return nil, s.mapProviderError(err)
	}
	return entries, nil
}

// Blob reads a file of a resolved repository.
func (s *RepositoryService) Blob(ctx context.Context, resolved *ResolvedRepository, path, refName string) (*providers.Blob, error) {
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapability(resolved, providers.CapabilityFiles, "files"); err != nil {
		return nil, err
	}
	reader, err := providers.FilesOf(adapter)
	if err != nil {
		return nil, err
	}
	blob, err := reader.Blob(ctx, providers.BlobQuery{Ref: refOfBinding(resolved.Binding), Path: path, RefName: refName})
	if err != nil {
		return nil, s.mapProviderError(err)
	}
	return blob, nil
}

// Issues lists the issues of a resolved repository, falling back to the local
// read model when the provider is unavailable.
func (s *RepositoryService) Issues(ctx context.Context, resolved *ResolvedRepository, state string, limit, page int) ([]providers.Issue, string, error) {
	if err := s.requireCapability(resolved, providers.CapabilityIssues, "issues"); err != nil {
		return nil, "", err
	}
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, "", err
	}
	reader, err := providers.IssuesOf(adapter)
	if err != nil {
		return nil, "", err
	}
	issues, listErr := reader.Issues(ctx, providers.ListQuery{Ref: refOfBinding(resolved.Binding), State: state, Limit: limit, Page: page})
	if listErr == nil {
		return issues, "provider", nil
	}
	fallback, fallbackErr := s.issuesFromReadModel(ctx, resolved, state, limit, page)
	if fallbackErr != nil {
		return nil, "", s.mapProviderError(listErr)
	}
	s.logger.Warn("provider issues unavailable, serving read model",
		"provider", resolved.Provider, "project_id", resolved.Project.ID, "error", listErr)
	return fallback, "read_model", nil
}

// Reviews lists the change proposals of a resolved repository.
func (s *RepositoryService) Reviews(ctx context.Context, resolved *ResolvedRepository, state string, limit, page int) ([]providers.Review, string, error) {
	if err := s.requireCapability(resolved, providers.CapabilityReviews, "reviews"); err != nil {
		return nil, "", err
	}
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, "", err
	}
	reader, err := providers.ReviewsOf(adapter)
	if err != nil {
		return nil, "", err
	}
	reviews, listErr := reader.Reviews(ctx, providers.ListQuery{Ref: refOfBinding(resolved.Binding), State: state, Limit: limit, Page: page})
	if listErr == nil {
		return reviews, "provider", nil
	}
	fallback, fallbackErr := s.reviewsFromReadModel(ctx, resolved, limit, page)
	if fallbackErr != nil {
		return nil, "", s.mapProviderError(listErr)
	}
	return fallback, "read_model", nil
}

// Releases lists the releases of a resolved repository.
func (s *RepositoryService) Releases(ctx context.Context, resolved *ResolvedRepository, limit, page int) ([]providers.Release, string, error) {
	if err := s.requireCapability(resolved, providers.CapabilityReleases, "releases"); err != nil {
		return nil, "", err
	}
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, "", err
	}
	reader, err := providers.ReleasesOf(adapter)
	if err != nil {
		return nil, "", err
	}
	releases, listErr := reader.Releases(ctx, providers.ListQuery{Ref: refOfBinding(resolved.Binding), Limit: limit, Page: page})
	if listErr == nil {
		return releases, "provider", nil
	}
	fallback, fallbackErr := s.releasesFromReadModel(ctx, resolved, limit, page)
	if fallbackErr != nil {
		return nil, "", s.mapProviderError(listErr)
	}
	return fallback, "read_model", nil
}

// Contributions aggregates the recent commits of a repository by day. It is the
// data behind the profile contribution graph.
func (s *RepositoryService) Contributions(ctx context.Context, resolved *ResolvedRepository, days int) ([]ContributionDay, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	adapter, err := s.adapter(ctx, resolved)
	if err != nil {
		return nil, err
	}
	reader, err := providers.CommitsOf(adapter)
	if err != nil {
		return nil, err
	}
	commits, err := reader.Commits(ctx, providers.CommitQuery{Ref: refOfBinding(resolved.Binding), Limit: 100})
	if err != nil {
		return nil, s.mapProviderError(err)
	}
	since := time.Now().UTC().AddDate(0, 0, -days)
	counts := map[string]int{}
	for _, commit := range commits {
		if commit.CommittedAt.Before(since) {
			continue
		}
		counts[commit.CommittedAt.UTC().Format("2006-01-02")]++
	}
	out := make([]ContributionDay, 0, len(counts))
	for day, count := range counts {
		out = append(out, ContributionDay{Date: day, Count: count})
	}
	return out, nil
}

// ContributionDay is one bucket of the contribution graph.
type ContributionDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// Languages approximates the language breakdown of a repository. The provider
// only exposes a primary language, so a full linguist-style analysis is
// reported as deferred in the API documentation.
func (s *RepositoryService) Languages(ctx context.Context, resolved *ResolvedRepository) (map[string]any, error) {
	repository, err := s.Repository(ctx, resolved)
	if err != nil {
		return nil, err
	}
	languages := []map[string]any{}
	if repository.Language != "" {
		languages = append(languages, map[string]any{
			"name":       repository.Language,
			"percentage": 100.0,
		})
	}
	return map[string]any{
		"primary":     repository.Language,
		"languages":   languages,
		"topics":      repository.Topics,
		"source":      "provider",
		"approximate": true,
	}, nil
}

// ── Internals ────────────────────────────────────────────────────────────────

func (s *RepositoryService) adapter(ctx context.Context, resolved *ResolvedRepository) (providers.Adapter, error) {
	if resolved == nil || resolved.Binding == nil {
		return nil, utils.ErrRepositoryBindingNotFound
	}
	binding := resolved.Binding
	if binding.ConnectionID != nil && *binding.ConnectionID != "" {
		connection, err := s.connections.connections.GetByID(ctx, *binding.ConnectionID)
		if err != nil {
			return nil, utils.ErrProviderConnectionNotFound
		}
		return s.connections.AdapterForConnection(connection)
	}
	ownerID := ""
	if resolved.Project != nil {
		ownerID = resolved.Project.OwnerID
	}
	return s.connections.Adapter(ctx, ownerID, binding.Provider)
}

func (s *RepositoryService) requireCapability(resolved *ResolvedRepository, capability, feature string) error {
	if resolved == nil || resolved.Provider == "" {
		return utils.ErrRepositoryBindingNotFound
	}
	if s.registry.Capabilities(resolved.Provider)[capability] == providers.CapabilityUnavailable {
		return utils.ErrProviderCapabilityMissing
	}
	_ = feature
	return nil
}

// mapProviderError translates typed provider errors into the API contract.
func (s *RepositoryService) mapProviderError(err error) error {
	if err == nil {
		return nil
	}
	return providers.AsAppError(err)
}

// findBindingByExternal resolves a repository that is bound to a project but
// not (yet) addressed through its Code reference.
func (s *RepositoryService) findBindingByExternal(ctx context.Context, owner, repo string) (*models.RepositoryBinding, error) {
	return s.bindings.GetByExternal(ctx, owner, repo)
}

func (s *RepositoryService) issuesFromReadModel(ctx context.Context, resolved *ResolvedRepository, state string, limit, page int) ([]providers.Issue, error) {
	items, _, err := s.resources.ListByProject(ctx, resolved.Project.ID, models.ExternalResourceIssue, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	out := make([]providers.Issue, 0, len(items))
	for _, item := range items {
		if state != "" && !strings.EqualFold(item.State, state) {
			continue
		}
		out = append(out, providers.Issue{
			ID:          item.ExternalID,
			Number:      item.Number,
			Title:       item.Title,
			State:       item.State,
			AuthorLogin: item.AuthorLogin,
			URL:         item.URL,
			UpdatedAt:   derefTime(item.ExternalUpdatedAt),
		})
	}
	return out, nil
}

func (s *RepositoryService) reviewsFromReadModel(ctx context.Context, resolved *ResolvedRepository, limit, page int) ([]providers.Review, error) {
	items, _, err := s.resources.ListByProject(ctx, resolved.Project.ID, models.ExternalResourceReview, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	out := make([]providers.Review, 0, len(items))
	for _, item := range items {
		out = append(out, providers.Review{
			ID:          item.ExternalID,
			Number:      item.Number,
			Title:       item.Title,
			State:       item.State,
			AuthorLogin: item.AuthorLogin,
			URL:         item.URL,
			UpdatedAt:   derefTime(item.ExternalUpdatedAt),
		})
	}
	return out, nil
}

func (s *RepositoryService) releasesFromReadModel(ctx context.Context, resolved *ResolvedRepository, limit, page int) ([]providers.Release, error) {
	items, _, err := s.resources.ListByProject(ctx, resolved.Project.ID, models.ExternalResourceRelease, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	out := make([]providers.Release, 0, len(items))
	for _, item := range items {
		out = append(out, providers.Release{
			ID:          item.ExternalID,
			TagName:     item.Title,
			Name:        item.Title,
			URL:         item.URL,
			PublishedAt: derefTime(item.ExternalUpdatedAt),
		})
	}
	return out, nil
}

// cached returns a cached provider read, or computes and stores it.
func (s *RepositoryService) cached(ctx context.Context, key string, dest any, fetch func() (any, error)) error {
	if s.cache != nil && key != "" {
		if err := s.cache.Get(ctx, key, dest); err == nil {
			return nil
		}
	}
	value, err := fetch()
	if err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if s.cache != nil && key != "" {
		if storeErr := s.cache.Set(ctx, key, value, s.cacheTTL); storeErr != nil {
			s.logger.Warn("provider cache write failed", "error", storeErr)
		}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, dest)
}

func cacheKey(resolved *ResolvedRepository) string {
	if resolved == nil || resolved.Binding == nil {
		return ""
	}
	binding := resolved.Binding
	return fmt.Sprintf("provider:%s:repo:%s/%s", binding.Provider, binding.ExternalOwner, binding.ExternalRepo)
}

func refOfBinding(binding *models.RepositoryBinding) providers.RepositoryRef {
	if binding == nil {
		return providers.RepositoryRef{}
	}
	return providers.RepositoryRef{Owner: binding.ExternalOwner, Repo: binding.ExternalRepo}
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
