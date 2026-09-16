package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/providers"
	"github.com/codegouvaor/code/server/src/utils"
	"gorm.io/datatypes"
)

// ProjectService owns the Code-native project domain. A project is the stable
// identity of a piece of software on the platform; external forges are only
// attached to it through repository bindings.
type ProjectService struct {
	projects   interfaces.ProjectRepository
	members    interfaces.ProjectMemberRepository
	assets     interfaces.ProjectAssetRepository
	stars      interfaces.ProjectStarRepository
	watches    interfaces.ProjectWatchRepository
	bindings   interfaces.RepositoryBindingRepository
	orgs       interfaces.OrganizationRepository
	orgMembers interfaces.OrganizationMemberRepository
	users      interfaces.UserRepository
	registry   *providers.Registry
	events     interfaces.EventBus
	auth       *Authorizer
}

// NewProjectService builds the project service.
func NewProjectService(
	projects interfaces.ProjectRepository,
	members interfaces.ProjectMemberRepository,
	assets interfaces.ProjectAssetRepository,
	stars interfaces.ProjectStarRepository,
	watches interfaces.ProjectWatchRepository,
	bindings interfaces.RepositoryBindingRepository,
	orgs interfaces.OrganizationRepository,
	orgMembers interfaces.OrganizationMemberRepository,
	users interfaces.UserRepository,
	registry *providers.Registry,
	events interfaces.EventBus,
	auth *Authorizer,
) *ProjectService {
	return &ProjectService{
		projects:   projects,
		members:    members,
		assets:     assets,
		stars:      stars,
		watches:    watches,
		bindings:   bindings,
		orgs:       orgs,
		orgMembers: orgMembers,
		users:      users,
		registry:   registry,
		events:     events,
		auth:       auth,
	}
}

// ProjectView is the API representation of a project: the Code-native record
// plus the computed fields the platform shell renders.
type ProjectView struct {
	ID               string             `json:"id"`
	Reference        string             `json:"reference"`
	Namespace        string             `json:"namespace"`
	Slug             string             `json:"slug"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Visibility       string             `json:"visibility"`
	OwnerID          string             `json:"ownerId"`
	OwnerLogin       string             `json:"ownerLogin"`
	OrganizationID   *string            `json:"organizationId,omitempty"`
	OrganizationSlug string             `json:"organizationSlug,omitempty"`
	Topics           []string           `json:"topics"`
	Metadata         map[string]any     `json:"metadata,omitempty"`
	HomepageURL      *string            `json:"homepageUrl,omitempty"`
	DefaultProvider  string             `json:"defaultProvider,omitempty"`
	Repository       *RepositorySummary `json:"repository,omitempty"`
	Capabilities     map[string]string  `json:"capabilities"`
	ViewerPermission string             `json:"viewerPermission"`
	Starred          bool               `json:"starred"`
	Watching         bool               `json:"watching"`
	Stars            int64              `json:"stars"`
	Watchers         int64              `json:"watchers"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
	ArchivedAt       *time.Time         `json:"archivedAt,omitempty"`
}

// RepositorySummary is the platform-level view of an attached repository,
// independent from the provider that hosts it.
type RepositorySummary struct {
	Provider      string            `json:"provider"`
	ProviderName  string            `json:"providerName"`
	ExternalID    string            `json:"externalId"`
	ExternalURL   string            `json:"externalUrl"`
	Owner         string            `json:"owner"`
	Name          string            `json:"name"`
	FullName      string            `json:"fullName"`
	DefaultBranch string            `json:"defaultBranch"`
	SyncStatus    string            `json:"syncStatus"`
	LastSyncedAt  *time.Time        `json:"lastSyncedAt,omitempty"`
	LastSyncError string            `json:"lastSyncError,omitempty"`
	IsPrimary     bool              `json:"isPrimary"`
	SyncEnabled   bool              `json:"syncEnabled"`
	Archived      bool              `json:"archived"`
	BindingID     string            `json:"bindingId"`
	Capabilities  map[string]string `json:"capabilities"`
}

// CreateProjectInput is the payload accepted by the creation endpoint.
type CreateProjectInput struct {
	Owner         string
	Organization  string
	Name          string
	Slug          string
	Description   string
	Visibility    string
	Topics        []string
	HomepageURL   string
	DefaultBranch string
}

// Create creates a Code project, its reference and its initial membership.
func (s *ProjectService) Create(ctx context.Context, principal interfaces.Principal, input CreateProjectInput) (*models.Project, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, utils.ErrValidationFailed
	}
	owner, err := s.resolveOwner(ctx, principal, input)
	if err != nil {
		return nil, err
	}

	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = slugifyDisplayName(name)
	}
	if !utils.ValidWorkspaceSlug(slug) {
		return nil, utils.NewError(400, "VALIDATION_ERROR", "The project slug is invalid.", map[string]any{"field": "slug"})
	}

	topics := normalizeTopics(input.Topics)
	visibility := normalizeVisibility(input.Visibility, "public")
	now := time.Now().UTC()
	project := &models.Project{
		Common:          models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		OwnerID:         owner.userID,
		OrganizationID:  owner.organizationID,
		Namespace:       owner.namespace,
		Slug:            slug,
		Name:            name,
		Description:     strings.TrimSpace(input.Description),
		Visibility:      visibility,
		Topics:          encodeStringSlice(topics),
		DefaultProvider: firstRegisteredProvider(s.registry),
	}
	project.Reference = owner.namespace + "/" + slug
	if homepage := strings.TrimSpace(input.HomepageURL); homepage != "" {
		project.HomepageURL = &homepage
	}
	// The reference must stay unique across the platform: appending a suffix is
	// preferable to failing a legitimate creation.
	for attempt := 2; attempt <= 20; attempt++ {
		if _, err := s.projects.GetByReference(ctx, project.Reference); err != nil {
			if utils.AsAppError(err).Code == "PROJECT_NOT_FOUND" {
				break
			}
			return nil, err
		}
		project.Reference = fmt.Sprintf("%s/%s-%d", owner.namespace, slug, attempt)
	}
	if _, err := s.projects.GetByReference(ctx, project.Reference); err == nil {
		return nil, utils.ErrProjectSlugTaken
	}
	if err := s.projects.Create(ctx, project); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, utils.ErrProjectSlugTaken
		}
		return nil, err
	}
	member := &models.ProjectMember{
		Common:    models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID: project.ID,
		UserID:    principal.UserID,
		Role:      string(RoleOwner),
		JoinedAt:  now,
	}
	if err := s.members.Create(ctx, member); err != nil {
		return nil, err
	}
	s.publish(ctx, "project.created", principal.UserID, map[string]any{
		"projectId":  project.ID,
		"reference":  project.Reference,
		"visibility": project.Visibility,
	})
	return project, nil
}

// Get resolves a project by id or reference and enforces read access.
func (s *ProjectService) Get(ctx context.Context, principal interfaces.Principal, ref string) (*models.Project, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectRead); err != nil {
		return nil, err
	}
	return project, nil
}

// View builds the enriched representation of a project.
func (s *ProjectService) View(ctx context.Context, principal interfaces.Principal, ref string) (*ProjectView, error) {
	project, err := s.Get(ctx, principal, ref)
	if err != nil {
		return nil, err
	}
	return s.buildView(ctx, principal, project)
}

func (s *ProjectService) buildView(ctx context.Context, principal interfaces.Principal, project *models.Project) (*ProjectView, error) {
	access, err := s.auth.ProjectAccess(ctx, principal.UserID, project)
	if err != nil {
		return nil, err
	}
	view := &ProjectView{
		ID:               project.ID,
		Reference:        project.Reference,
		Namespace:        project.Namespace,
		Slug:             project.Slug,
		Name:             project.Name,
		Description:      project.Description,
		Visibility:       project.Visibility,
		OwnerID:          project.OwnerID,
		OrganizationID:   project.OrganizationID,
		Topics:           decodeStringSlice(project.Topics),
		Metadata:         decodeMap(project.Metadata),
		HomepageURL:      project.HomepageURL,
		DefaultProvider:  project.DefaultProvider,
		Capabilities:     map[string]string{},
		ViewerPermission: string(access.Role),
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
		ArchivedAt:       project.ArchivedAt,
	}
	if owner, ownerErr := s.users.GetByID(ctx, project.OwnerID); ownerErr == nil {
		view.OwnerLogin = derefString(owner.Username)
	}
	if project.OrganizationID != nil && *project.OrganizationID != "" {
		if organization, orgErr := s.orgs.GetByID(ctx, *project.OrganizationID); orgErr == nil {
			view.OrganizationSlug = organization.Slug
		}
	}
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
		view.Repository = summaryOfBinding(primary, s.registry)
		view.Capabilities = stringifyCapabilities(s.registry.Capabilities(primary.Provider))
	} else {
		view.Capabilities = codeOnlyCapabilities()
	}
	stars, err := s.stars.CountByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	view.Stars = stars
	if principal.UserID != "" {
		if _, starErr := s.stars.Get(ctx, project.ID, principal.UserID); starErr == nil {
			view.Starred = true
		}
		if _, watchErr := s.watches.Get(ctx, project.ID, principal.UserID); watchErr == nil {
			view.Watching = true
		}
	}
	return view, nil
}

// List returns the visible projects of an owner namespace.
func (s *ProjectService) List(ctx context.Context, principal interfaces.Principal, filter interfaces.ProjectFilter) ([]ProjectView, int64, error) {
	if filter.Visibility == "" {
		filter.Visibility = "public"
	}
	projects, total, err := s.projects.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	views := make([]ProjectView, 0, len(projects))
	for index := range projects {
		project := projects[index]
		access, accessErr := s.auth.ProjectAccess(ctx, principal.UserID, &project)
		if accessErr != nil {
			return nil, 0, accessErr
		}
		if !access.Can(ActionProjectRead) {
			total--
			continue
		}
		view, viewErr := s.buildView(ctx, principal, &project)
		if viewErr != nil {
			return nil, 0, viewErr
		}
		views = append(views, *view)
	}
	return views, total, nil
}

// ListByOwner lists the projects of a user or organization namespace.
func (s *ProjectService) ListByOwner(ctx context.Context, principal interfaces.Principal, owner string) ([]ProjectView, error) {
	resolved, err := s.resolveOwnerRef(ctx, owner)
	if err != nil {
		return nil, err
	}
	var projects []models.Project
	if resolved.organizationID != nil {
		projects, err = s.projects.ListByOrganization(ctx, *resolved.organizationID)
	} else {
		all, listErr := s.projects.ListByUser(ctx, resolved.userID)
		if listErr != nil {
			return nil, listErr
		}
		projects = make([]models.Project, 0, len(all))
		for _, project := range all {
			if project.OwnerID == resolved.userID {
				projects = append(projects, project)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	views := make([]ProjectView, 0, len(projects))
	for index := range projects {
		project := projects[index]
		access, accessErr := s.auth.ProjectAccess(ctx, principal.UserID, &project)
		if accessErr != nil {
			return nil, accessErr
		}
		if !access.Can(ActionProjectRead) {
			continue
		}
		view, viewErr := s.buildView(ctx, principal, &project)
		if viewErr != nil {
			return nil, viewErr
		}
		views = append(views, *view)
	}
	return views, nil
}

// Update applies a partial update to a project.
func (s *ProjectService) Update(ctx context.Context, principal interfaces.Principal, ref string, input CreateProjectInput) (*ProjectView, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectAdmin); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) != "" {
		project.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		project.Description = strings.TrimSpace(input.Description)
	}
	if input.Visibility != "" {
		project.Visibility = normalizeVisibility(input.Visibility, project.Visibility)
	}
	if input.Topics != nil {
		project.Topics = encodeStringSlice(normalizeTopics(input.Topics))
	}
	if input.HomepageURL != "" {
		homepage := strings.TrimSpace(input.HomepageURL)
		project.HomepageURL = &homepage
	}
	project.UpdatedAt = time.Now().UTC()
	project.LastActivityAt = &project.UpdatedAt
	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}
	s.publish(ctx, "project.updated", principal.UserID, map[string]any{"projectId": project.ID})
	return s.buildView(ctx, principal, project)
}

// Archive soft-deletes a project.
func (s *ProjectService) Archive(ctx context.Context, principal interfaces.Principal, ref string) error {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectDelete); err != nil {
		return err
	}
	if err := s.projects.Archive(ctx, project.ID, time.Now().UTC()); err != nil {
		return err
	}
	s.publish(ctx, "project.archived", principal.UserID, map[string]any{"projectId": project.ID})
	return nil
}

// ── Members ──────────────────────────────────────────────────────────────────

// Member is the API representation of a project member.
type Member struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl,omitempty"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

// ListMembers returns the direct members of a project.
func (s *ProjectService) ListMembers(ctx context.Context, principal interfaces.Principal, ref string) ([]Member, error) {
	project, err := s.Get(ctx, principal, ref)
	if err != nil {
		return nil, err
	}
	items, err := s.members.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.UserID)
	}
	users, err := s.users.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]models.User, len(users))
	for _, user := range users {
		byID[user.ID] = user
	}
	out := make([]Member, 0, len(items))
	for _, item := range items {
		user := byID[item.UserID]
		out = append(out, Member{
			ID:          item.ID,
			UserID:      item.UserID,
			Username:    derefString(user.Username),
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        item.Role,
			JoinedAt:    item.JoinedAt,
		})
	}
	return out, nil
}

// AddMember adds a member to a project.
func (s *ProjectService) AddMember(ctx context.Context, principal interfaces.Principal, ref, userRef, role string) (*Member, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	access, err := s.auth.AuthorizeProject(ctx, principal, project, ActionMemberManage)
	if err != nil {
		return nil, err
	}
	if !isValidRole(role) {
		return nil, utils.ErrValidationFailed
	}
	normalizedRole := normalizeRole(role)
	if normalizedRole == RoleOwner && access.Role != RoleOwner {
		return nil, utils.ErrForbidden
	}
	if user, lookupErr := s.users.GetByUsername(ctx, userRef); lookupErr == nil && user.ID != "" {
		return s.attachMember(ctx, principal, project, normalizedRole, user)
	}
	if user, lookupErr := s.users.GetByID(ctx, userRef); lookupErr == nil && user.ID != "" {
		return s.attachMember(ctx, principal, project, normalizedRole, user)
	}
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(userRef)))
	if err != nil {
		return nil, err
	}
	return s.attachMember(ctx, principal, project, normalizedRole, user)
}

func (s *ProjectService) attachMember(ctx context.Context, principal interfaces.Principal, project *models.Project, role Role, user *models.User) (*Member, error) {
	if existing, err := s.members.Get(ctx, project.ID, user.ID); err == nil && existing != nil && existing.ID != "" {
		return nil, utils.NewError(409, "PROJECT_MEMBER_EXISTS", "This user is already a member of the project.", nil)
	}
	now := time.Now().UTC()
	member := &models.ProjectMember{
		Common:    models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      string(role),
		JoinedAt:  now,
	}
	if err := s.members.Create(ctx, member); err != nil {
		return nil, err
	}
	s.publish(ctx, "member.added", principal.UserID, map[string]any{
		"projectId": project.ID,
		"userId":    user.ID,
		"role":      string(role),
		"scope":     "project",
	})
	return &Member{
		ID:          member.ID,
		UserID:      user.ID,
		Username:    derefString(user.Username),
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// UpdateMember changes a project member role.
func (s *ProjectService) UpdateMember(ctx context.Context, principal interfaces.Principal, ref, userRef, role string) (*Member, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionMemberManage); err != nil {
		return nil, err
	}
	if !isValidRole(role) {
		return nil, utils.ErrValidationFailed
	}
	user, err := s.resolveMemberUser(ctx, userRef)
	if err != nil {
		return nil, err
	}
	member, err := s.members.Get(ctx, project.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if member.UserID == project.OwnerID && normalizeRole(role) != RoleOwner {
		return nil, utils.ErrForbidden
	}
	member.Role = string(normalizeRole(role))
	member.UpdatedAt = time.Now().UTC()
	if err := s.members.Update(ctx, member); err != nil {
		return nil, err
	}
	return &Member{
		ID:          member.ID,
		UserID:      user.ID,
		Username:    derefString(user.Username),
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// RemoveMember removes a member from a project.
func (s *ProjectService) RemoveMember(ctx context.Context, principal interfaces.Principal, ref, userRef string) error {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionMemberManage); err != nil {
		return err
	}
	user, err := s.resolveMemberUser(ctx, userRef)
	if err != nil {
		return err
	}
	if user.ID == project.OwnerID {
		return utils.NewError(409, "PROJECT_OWNER_REQUIRED", "The project owner cannot be removed.", nil)
	}
	return s.members.Delete(ctx, project.ID, user.ID)
}

func (s *ProjectService) resolveMemberUser(ctx context.Context, userRef string) (*models.User, error) {
	if user, err := s.users.GetByID(ctx, userRef); err == nil && user.ID != "" {
		return user, nil
	}
	if user, err := s.users.GetByUsername(ctx, userRef); err == nil && user.ID != "" {
		return user, nil
	}
	return s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(userRef)))
}

// ── Code-native assets ───────────────────────────────────────────────────────

// CreateAssetInput is the payload accepted by the asset endpoints.
type CreateAssetInput struct {
	Kind        string
	Slug        string
	Name        string
	Description string
	URL         string
	Visibility  string
	Position    int
	Metadata    map[string]any
}

// ListAssets returns the Code-native resources attached to a project.
func (s *ProjectService) ListAssets(ctx context.Context, principal interfaces.Principal, ref, kind string) ([]models.ProjectAsset, error) {
	project, err := s.Get(ctx, principal, ref)
	if err != nil {
		return nil, err
	}
	if kind != "" && !ValidAssetKind(kind) {
		return nil, utils.ErrValidationFailed
	}
	return s.assets.ListByProject(ctx, project.ID, kind)
}

// CreateAsset attaches a Code-native resource to a project.
func (s *ProjectService) CreateAsset(ctx context.Context, principal interfaces.Principal, ref string, input CreateAssetInput) (*models.ProjectAsset, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionAssetManage); err != nil {
		return nil, err
	}
	if !ValidAssetKind(input.Kind) {
		return nil, utils.ErrValidationFailed
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, utils.ErrValidationFailed
	}
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = slugifyDisplayName(name)
	}
	if _, err := s.assets.GetByProjectAndSlug(ctx, project.ID, input.Kind, slug); err == nil {
		return nil, utils.ErrSlugAlreadyExists
	}
	now := time.Now().UTC()
	asset := &models.ProjectAsset{
		Common:      models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID:   project.ID,
		Kind:        input.Kind,
		Slug:        slug,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Visibility:  normalizeVisibility(input.Visibility, "public"),
		Position:    input.Position,
		Metadata:    encodeMap(input.Metadata),
	}
	if url := strings.TrimSpace(input.URL); url != "" {
		asset.URL = &url
	}
	if err := s.assets.Create(ctx, asset); err != nil {
		return nil, err
	}
	s.publish(ctx, "project.asset.created", principal.UserID, map[string]any{"projectId": project.ID, "assetId": asset.ID, "kind": asset.Kind})
	return asset, nil
}

// UpdateAsset updates a Code-native resource.
func (s *ProjectService) UpdateAsset(ctx context.Context, principal interfaces.Principal, projectRef, assetID string, input CreateAssetInput) (*models.ProjectAsset, error) {
	project, err := s.resolve(ctx, projectRef)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionAssetManage); err != nil {
		return nil, err
	}
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}
	if asset.ProjectID != project.ID {
		return nil, utils.ErrProjectAssetNotFound
	}
	if strings.TrimSpace(input.Name) != "" {
		asset.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		asset.Description = strings.TrimSpace(input.Description)
	}
	if input.Visibility != "" {
		asset.Visibility = normalizeVisibility(input.Visibility, asset.Visibility)
	}
	if input.URL != "" {
		url := strings.TrimSpace(input.URL)
		asset.URL = &url
	}
	if input.Metadata != nil {
		asset.Metadata = encodeMap(input.Metadata)
	}
	asset.UpdatedAt = time.Now().UTC()
	if err := s.assets.Update(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

// DeleteAsset removes a Code-native resource.
func (s *ProjectService) DeleteAsset(ctx context.Context, principal interfaces.Principal, projectRef, assetID string) error {
	project, err := s.resolve(ctx, projectRef)
	if err != nil {
		return err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionAssetManage); err != nil {
		return err
	}
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset.ProjectID != project.ID {
		return utils.ErrProjectAssetNotFound
	}
	return s.assets.Delete(ctx, asset.ID)
}

// ValidAssetKind validates the resource discriminator.
func ValidAssetKind(kind string) bool {
	switch models.ProjectAssetKind(strings.ToLower(strings.TrimSpace(kind))) {
	case models.ProjectAssetDocumentation, models.ProjectAssetAPI, models.ProjectAssetSDK,
		models.ProjectAssetService, models.ProjectAssetStandard, models.ProjectAssetIntegration:
		return true
	default:
		return false
	}
}

// ── Stars and watches ────────────────────────────────────────────────────────

// Star marks a project as starred by the caller.
func (s *ProjectService) Star(ctx context.Context, principal interfaces.Principal, ref string) (*ProjectView, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectRead); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	star := &models.ProjectStar{
		Common:    models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID: project.ID,
		UserID:    principal.UserID,
		StarredAt: now,
	}
	if err := s.stars.Create(ctx, star); err != nil {
		return nil, err
	}
	s.publish(ctx, "project.starred", principal.UserID, map[string]any{"projectId": project.ID})
	return s.buildView(ctx, principal, project)
}

// Unstar removes a star.
func (s *ProjectService) Unstar(ctx context.Context, principal interfaces.Principal, ref string) (*ProjectView, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if err := s.stars.Delete(ctx, project.ID, principal.UserID); err != nil {
		return nil, err
	}
	return s.buildView(ctx, principal, project)
}

// Watch subscribes the caller to a project's activity.
func (s *ProjectService) Watch(ctx context.Context, principal interfaces.Principal, ref, level string) (*ProjectView, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeProject(ctx, principal, project, ActionProjectRead); err != nil {
		return nil, err
	}
	if level == "" {
		level = "participating"
	}
	now := time.Now().UTC()
	watch := &models.ProjectWatch{
		Common:    models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		ProjectID: project.ID,
		UserID:    principal.UserID,
		Level:     level,
		WatchedAt: now,
	}
	if err := s.watches.Upsert(ctx, watch); err != nil {
		return nil, err
	}
	return s.buildView(ctx, principal, project)
}

// Unwatch removes a subscription.
func (s *ProjectService) Unwatch(ctx context.Context, principal interfaces.Principal, ref string) (*ProjectView, error) {
	project, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if err := s.watches.Delete(ctx, project.ID, principal.UserID); err != nil {
		return nil, err
	}
	return s.buildView(ctx, principal, project)
}

// StarredProjects lists the projects starred by the caller.
func (s *ProjectService) StarredProjects(ctx context.Context, principal interfaces.Principal, offset, limit int) ([]ProjectView, int64, error) {
	stars, total, err := s.stars.ListByUser(ctx, principal.UserID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	views := make([]ProjectView, 0, len(stars))
	for _, star := range stars {
		project, projectErr := s.projects.GetByID(ctx, star.ProjectID)
		if projectErr != nil {
			continue
		}
		view, viewErr := s.buildView(ctx, principal, project)
		if viewErr != nil {
			continue
		}
		views = append(views, *view)
	}
	return views, total, nil
}

// ── Internals ────────────────────────────────────────────────────────────────

type ownerRef struct {
	userID         string
	organizationID *string
	namespace      string
}

// resolveOwner validates the requested namespace for a creation.
func (s *ProjectService) resolveOwner(ctx context.Context, principal interfaces.Principal, input CreateProjectInput) (ownerRef, error) {
	requested := strings.TrimSpace(input.Organization)
	if requested == "" {
		requested = strings.TrimSpace(input.Owner)
	}
	if requested == "" {
		user, err := s.users.GetByID(ctx, principal.UserID)
		if err != nil {
			return ownerRef{}, err
		}
		if user.Username == nil || *user.Username == "" {
			return ownerRef{}, utils.ErrUsernameInvalid
		}
		return ownerRef{userID: user.ID, namespace: *user.Username}, nil
	}

	if organization, err := s.orgs.GetBySlug(ctx, requested); err == nil && organization.ID != "" {
		if _, authErr := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionProjectWrite); authErr != nil {
			return ownerRef{}, authErr
		}
		organizationID := organization.ID
		return ownerRef{userID: principal.UserID, organizationID: &organizationID, namespace: organization.Slug}, nil
	}

	if user, err := s.users.GetByUsername(ctx, requested); err == nil && user.ID != "" {
		if user.ID != principal.UserID {
			return ownerRef{}, utils.ErrForbidden
		}
		return ownerRef{userID: user.ID, namespace: derefString(user.Username)}, nil
	}
	return ownerRef{}, utils.ErrValidationFailed
}

func (s *ProjectService) resolveOwnerRef(ctx context.Context, owner string) (ownerRef, error) {
	value := strings.ToLower(strings.TrimSpace(owner))
	if value == "" {
		return ownerRef{}, utils.ErrOwnerNotFound
	}
	if user, err := s.users.GetByUsername(ctx, value); err == nil && user.ID != "" {
		return ownerRef{userID: user.ID, namespace: derefString(user.Username)}, nil
	}
	if organization, err := s.orgs.GetBySlug(ctx, value); err == nil && organization.ID != "" {
		organizationID := organization.ID
		return ownerRef{organizationID: &organizationID, namespace: organization.Slug}, nil
	}
	return ownerRef{}, utils.ErrOwnerNotFound
}

// resolve accepts a project id, a full reference ("owner/slug") or a slug.
func (s *ProjectService) resolve(ctx context.Context, ref string) (*models.Project, error) {
	value := strings.TrimSpace(ref)
	if value == "" {
		return nil, utils.ErrProjectNotFound
	}
	if project, err := s.projects.GetByID(ctx, value); err == nil && project.ID != "" {
		return project, nil
	}
	if project, err := s.projects.GetByReference(ctx, value); err == nil && project.ID != "" {
		return project, nil
	}
	if strings.Contains(value, "/") {
		parts := strings.SplitN(value, "/", 2)
		return s.projects.GetByNamespaceAndSlug(ctx, parts[0], parts[1])
	}
	return s.projects.GetByReference(ctx, value)
}

func (s *ProjectService) publish(ctx context.Context, eventType, actorID string, payload map[string]any) {
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

func summaryOfBinding(binding models.RepositoryBinding, registry *providers.Registry) *RepositorySummary {
	summary := &RepositorySummary{
		Provider:      binding.Provider,
		ExternalID:    binding.ExternalID,
		ExternalURL:   binding.ExternalURL,
		Owner:         binding.ExternalOwner,
		Name:          binding.ExternalRepo,
		FullName:      binding.ExternalOwner + "/" + binding.ExternalRepo,
		DefaultBranch: binding.DefaultBranch,
		SyncStatus:    binding.SyncStatus,
		LastSyncedAt:  binding.LastSyncedAt,
		LastSyncError: binding.LastSyncError,
		IsPrimary:     binding.IsPrimary,
		SyncEnabled:   binding.SyncEnabled,
		Archived:      binding.ExternalArchived,
		BindingID:     binding.ID,
		Capabilities:  map[string]string{},
	}
	if registry != nil {
		if descriptor, ok := registry.Descriptor(binding.Provider); ok {
			summary.ProviderName = descriptor.DisplayName
			summary.Capabilities = stringifyCapabilities(descriptor.Capabilities)
		}
	}
	if summary.ProviderName == "" {
		summary.ProviderName = binding.Provider
	}
	return summary
}

func stringifyCapabilities(capabilities map[string]providers.Capability) map[string]string {
	out := make(map[string]string, len(capabilities))
	for key, value := range capabilities {
		out[key] = string(value)
	}
	return out
}

// codeOnlyCapabilities is what a project can do without any bound repository:
// every Code-native resource stays available.
func codeOnlyCapabilities() map[string]string {
	native := []string{
		providers.CapabilityDocumentation, providers.CapabilityAPIs, providers.CapabilitySDKs,
		providers.CapabilityServices, providers.CapabilityStandards, providers.CapabilityRepositories,
	}
	out := make(map[string]string, len(native)+8)
	for _, key := range native {
		out[key] = string(providers.CapabilityNative)
	}
	unavailable := []string{
		providers.CapabilityBranches, providers.CapabilityCommits, providers.CapabilityFiles,
		providers.CapabilityIssues, providers.CapabilityReviews, providers.CapabilityReleases,
		providers.CapabilityOrganizations, providers.CapabilityWebhooks, providers.CapabilityCICD,
		providers.CapabilityPackages,
	}
	for _, key := range unavailable {
		out[key] = string(providers.CapabilityUnavailable)
	}
	return out
}

func normalizeTopics(topics []string) []string {
	out := make([]string, 0, len(topics))
	seen := map[string]struct{}{}
	for _, topic := range topics {
		value := strings.ToLower(strings.TrimSpace(topic))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func firstRegisteredProvider(registry *providers.Registry) string {
	if registry == nil {
		return ""
	}
	names := registry.Names()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func encodeStringSlice(values []string) datatypes.JSON {
	if values == nil {
		values = []string{}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(data)
}

func decodeStringSlice(data datatypes.JSON) []string {
	if len(data) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return []string{}
	}
	return values
}

func encodeMap(values map[string]any) datatypes.JSON {
	if values == nil {
		return datatypes.JSON("{}")
	}
	data, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON("{}")
	}
	return datatypes.JSON(data)
}

func decodeMap(data datatypes.JSON) map[string]any {
	if len(data) == 0 {
		return nil
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil
	}
	return values
}
