package services

import (
	"context"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/utils"
)

// SearchResultKind enumerates the resources returned by the global search.
type SearchResultKind string

const (
	SearchKindProject       SearchResultKind = "project"
	SearchKindOrganization  SearchResultKind = "organization"
	SearchKindUser          SearchResultKind = "user"
	SearchKindRepository    SearchResultKind = "repository"
	SearchKindDocumentation SearchResultKind = "documentation"
	SearchKindAPI           SearchResultKind = "api"
	SearchKindSDK           SearchResultKind = "sdk"
	SearchKindService       SearchResultKind = "service"
	SearchKindStandard      SearchResultKind = "standard"
)

// SearchHit is one result of the global search.
type SearchHit struct {
	Kind        SearchResultKind `json:"kind"`
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Reference   string           `json:"reference,omitempty"`
	URL         string           `json:"url"`
	Visibility  string           `json:"visibility,omitempty"`
	Provider    string           `json:"provider,omitempty"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	Score       float64          `json:"score,omitempty"`
}

// SearchResponse is the grouped result of a search request.
type SearchResponse struct {
	Query   string      `json:"query"`
	Kinds   []string    `json:"kinds"`
	Total   int64       `json:"total"`
	HasMore bool        `json:"hasMore"`
	Hits    []SearchHit `json:"hits"`
}

// SearchEngine is the abstraction behind the global search. The platform ships
// a PostgreSQL implementation; a dedicated engine can replace it later without
// touching handlers or the domain.
type SearchEngine interface {
	Search(ctx context.Context, viewerID string, filter interfaces.SearchFilter) ([]SearchHit, int64, error)
	Name() string
}

// SearchService orchestrates the configured engine.
type SearchService struct {
	engines []SearchEngine
}

// NewSearchService builds the search service.
func NewSearchService(engines ...SearchEngine) *SearchService {
	return &SearchService{engines: engines}
}

// Search runs a global search across the requested kinds.
func (s *SearchService) Search(ctx context.Context, viewerID string, filter interfaces.SearchFilter) (*SearchResponse, error) {
	query := strings.TrimSpace(filter.Query)
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	filter.Query = query
	response := &SearchResponse{Query: query, Kinds: filter.Kinds, Hits: []SearchHit{}}
	for _, engine := range s.engines {
		hits, total, err := engine.Search(ctx, viewerID, filter)
		if err != nil {
			return nil, err
		}
		response.Hits = append(response.Hits, hits...)
		response.Total += total
	}
	if len(response.Hits) > filter.Limit {
		response.HasMore = true
		response.Hits = response.Hits[:filter.Limit]
	}
	return response, nil
}

// Engines exposes the configured engine names, for the discovery endpoint.
func (s *SearchService) Engines() []string {
	out := make([]string, 0, len(s.engines))
	for _, engine := range s.engines {
		out = append(out, engine.Name())
	}
	return out
}

// NormalizeSearchKinds validates and normalises the requested kinds.
func NormalizeSearchKinds(kinds []string) ([]string, error) {
	if len(kinds) == 0 {
		return nil, nil
	}
	allowed := map[string]struct{}{
		string(SearchKindProject): {}, string(SearchKindOrganization): {}, string(SearchKindUser): {},
		string(SearchKindRepository): {}, string(SearchKindDocumentation): {}, string(SearchKindAPI): {},
		string(SearchKindSDK): {}, string(SearchKindService): {}, string(SearchKindStandard): {},
	}
	out := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		value := strings.ToLower(strings.TrimSpace(kind))
		if value == "" {
			continue
		}
		if _, ok := allowed[value]; !ok {
			return nil, utils.ErrValidationFailed
		}
		out = append(out, value)
	}
	return out, nil
}

// PostgresSearchEngine implements full-text-ish search directly on PostgreSQL.
// It is intentionally simple (ILIKE on indexed columns) and becomes the
// read-side of a future dedicated index without any domain change.
type PostgresSearchEngine struct {
	projects      interfaces.ProjectRepository
	organizations interfaces.OrganizationRepository
	assets        interfaces.ProjectAssetRepository
	users         interfaces.UserRepository
}

// NewPostgresSearchEngine builds the default engine.
func NewPostgresSearchEngine(
	projects interfaces.ProjectRepository,
	organizations interfaces.OrganizationRepository,
	assets interfaces.ProjectAssetRepository,
	users interfaces.UserRepository,
) *PostgresSearchEngine {
	return &PostgresSearchEngine{projects: projects, organizations: organizations, assets: assets, users: users}
}

// Name identifies the engine.
func (e *PostgresSearchEngine) Name() string { return "postgres" }

// Search queries the platform tables.
func (e *PostgresSearchEngine) Search(ctx context.Context, _ string, filter interfaces.SearchFilter) ([]SearchHit, int64, error) {
	kinds := map[string]struct{}{}
	for _, kind := range filter.Kinds {
		kinds[kind] = struct{}{}
	}
	all := len(kinds) == 0
	hits := make([]SearchHit, 0, filter.Limit)
	var total int64

	if all || hasKind(kinds, SearchKindProject) || hasKind(kinds, SearchKindRepository) {
		projects, count, err := e.projects.Search(ctx, filter.Query, filter.Offset, filter.Limit)
		if err != nil {
			return nil, 0, err
		}
		total += count
		for _, project := range projects {
			kind := SearchKindProject
			if hasKind(kinds, SearchKindRepository) && !hasKind(kinds, SearchKindProject) {
				kind = SearchKindRepository
			}
			hits = append(hits, SearchHit{
				Kind:        kind,
				ID:          project.ID,
				Title:       project.Name,
				Description: project.Description,
				Reference:   project.Reference,
				URL:         "/" + project.Reference,
				Visibility:  project.Visibility,
				UpdatedAt:   project.UpdatedAt,
			})
		}
	}

	if all || hasKind(kinds, SearchKindOrganization) {
		organizations, count, err := e.organizations.Search(ctx, filter.Query, filter.Offset, filter.Limit)
		if err != nil {
			return nil, 0, err
		}
		total += count
		for _, organization := range organizations {
			hits = append(hits, SearchHit{
				Kind:        SearchKindOrganization,
				ID:          organization.ID,
				Title:       organization.Name,
				Description: organization.Description,
				Reference:   organization.Slug,
				URL:         "/" + organization.Slug,
				Visibility:  organization.Visibility,
				UpdatedAt:   organization.UpdatedAt,
			})
		}
	}

	if all || hasKind(kinds, SearchKindDocumentation) || hasKind(kinds, SearchKindAPI) ||
		hasKind(kinds, SearchKindSDK) || hasKind(kinds, SearchKindService) || hasKind(kinds, SearchKindStandard) {
		assets, count, err := e.assets.Search(ctx, filter.Query, filter.Offset, filter.Limit)
		if err != nil {
			return nil, 0, err
		}
		total += count
		for _, asset := range assets {
			if !all && !hasKind(kinds, SearchResultKind(asset.Kind)) {
				continue
			}
			hits = append(hits, SearchHit{
				Kind:        SearchResultKind(asset.Kind),
				ID:          asset.ID,
				Title:       asset.Name,
				Description: asset.Description,
				URL:         derefString(asset.URL),
				Visibility:  asset.Visibility,
				UpdatedAt:   asset.UpdatedAt,
			})
		}
	}

	if all || hasKind(kinds, SearchKindUser) {
		users, count, err := e.users.Search(ctx, filter.Query, filter.Offset, filter.Limit)
		if err != nil {
			return nil, 0, err
		}
		total += count
		for _, user := range users {
			username := derefString(user.Username)
			if username == "" {
				continue
			}
			hits = append(hits, SearchHit{
				Kind:      SearchKindUser,
				ID:        user.ID,
				Title:     user.DisplayName,
				Reference: username,
				URL:       "/" + username,
				UpdatedAt: user.UpdatedAt,
			})
		}
	}

	return hits, total, nil
}

func hasKind(kinds map[string]struct{}, kind SearchResultKind) bool {
	_, ok := kinds[string(kind)]
	return ok
}

// ensure the engine satisfies the interface.
var _ SearchEngine = (*PostgresSearchEngine)(nil)
