package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

// usernamePattern matches the Code handle grammar: lowercase alphanumeric
// segments separated by single hyphens (no leading, trailing or doubled
// hyphen). The length is validated separately so that the pattern stays RE2
// compatible.
var usernamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const (
	usernameMinLength = 2
	usernameMaxLength = 39
)

// reservedUsernames cannot be claimed because they collide with platform
// routes or well-known pages.
var reservedUsernames = map[string]struct{}{
	"api": {}, "admin": {}, "new": {}, "dashboard": {}, "issues": {}, "pulls": {},
	"repositories": {}, "marketplace": {}, "notifications": {}, "settings": {},
	"login": {}, "logout": {}, "register": {}, "explore": {}, "search": {},
	"organizations": {}, "projects": {}, "teams": {}, "sponsors": {}, "about": {},
	"help": {}, "support": {}, "code": {}, "giteria": {}, "docs": {}, "assets": {},
	"static": {}, "public": {}, "me": {}, "owners": {}, "integrations": {},
}

// OwnerCapabilities mirrors the capability flags the platform shell needs to
// decide which navigation entries to render for an owner.
type OwnerCapabilities struct {
	Teams      bool `json:"teams"`
	People     bool `json:"people"`
	Insights   bool `json:"insights"`
	Sponsoring bool `json:"sponsoring"`
}

// OwnerCounts is the lightweight counter block rendered on profiles.
type OwnerCounts struct {
	Repositories int64 `json:"repositories"`
	Projects     int64 `json:"projects"`
	Teams        int64 `json:"teams"`
	Members      int64 `json:"members"`
	Stars        int64 `json:"stars"`
}

// Owner is the resolved identity behind a /[owner] route: either a Code user
// or a Code organization.
type Owner struct {
	Type             string            `json:"type"`
	ID               string            `json:"id"`
	Username         string            `json:"username"`
	Name             string            `json:"name"`
	DisplayName      string            `json:"displayName"`
	Description      string            `json:"description"`
	AvatarURL        *string           `json:"avatarUrl,omitempty"`
	WebsiteURL       *string           `json:"websiteUrl,omitempty"`
	Location         string            `json:"location,omitempty"`
	Visibility       string            `json:"visibility"`
	Capabilities     OwnerCapabilities `json:"capabilities"`
	Counts           OwnerCounts       `json:"counts"`
	ViewerPermission string            `json:"viewerPermission"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

// OwnerService resolves owner namespaces. Usernames and organization slugs
// share a single namespace so that /[owner] is never ambiguous.
type OwnerService struct {
	users      interfaces.UserRepository
	orgs       interfaces.OrganizationRepository
	orgMembers interfaces.OrganizationMemberRepository
	orgTeams   interfaces.OrganizationTeamRepository
	projects   interfaces.ProjectRepository
	bindings   interfaces.RepositoryBindingRepository
	stars      interfaces.ProjectStarRepository
	auth       *Authorizer
}

// NewOwnerService builds the owner service.
func NewOwnerService(
	users interfaces.UserRepository,
	orgs interfaces.OrganizationRepository,
	orgMembers interfaces.OrganizationMemberRepository,
	orgTeams interfaces.OrganizationTeamRepository,
	projects interfaces.ProjectRepository,
	bindings interfaces.RepositoryBindingRepository,
	stars interfaces.ProjectStarRepository,
	auth *Authorizer,
) *OwnerService {
	return &OwnerService{
		users:      users,
		orgs:       orgs,
		orgMembers: orgMembers,
		orgTeams:   orgTeams,
		projects:   projects,
		bindings:   bindings,
		stars:      stars,
		auth:       auth,
	}
}

// Resolve returns the owner behind a name, or a 404 API error.
func (s *OwnerService) Resolve(ctx context.Context, viewerID, name string) (*Owner, error) {
	slug := strings.ToLower(strings.TrimSpace(name))
	if slug == "" {
		return nil, utils.ErrOwnerNotFound
	}
	if user, err := s.users.GetByUsername(ctx, slug); err == nil && user != nil && user.ID != "" {
		return s.buildUserOwner(ctx, viewerID, user)
	} else if err != nil && utils.AsAppError(err).Code != "USER_NOT_FOUND" {
		return nil, err
	}
	if organization, err := s.orgs.GetBySlug(ctx, slug); err == nil && organization != nil && organization.ID != "" {
		return s.buildOrganizationOwner(ctx, viewerID, organization)
	} else if err != nil && utils.AsAppError(err).Code != "ORGANIZATION_NOT_FOUND" {
		return nil, err
	}
	return nil, utils.ErrOwnerNotFound
}

func (s *OwnerService) buildUserOwner(ctx context.Context, viewerID string, user *models.User) (*Owner, error) {
	owner := &Owner{
		Type:        "user",
		ID:          user.ID,
		Username:    derefString(user.Username),
		Name:        user.DisplayName,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Visibility:  "public",
		Capabilities: OwnerCapabilities{
			// A personal namespace has repositories and projects, but no
			// teams, no people directory and no organization insights.
			Sponsoring: false,
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	projects, err := s.projects.ListByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	owned := make([]models.Project, 0, len(projects))
	for _, project := range projects {
		if project.OwnerID == user.ID {
			owned = append(owned, project)
		}
	}
	owner.Counts.Repositories = int64(len(owned))
	owner.Counts.Projects = int64(len(owned))
	owner.Counts.Stars = s.countStars(ctx, owned)
	owner.ViewerPermission = "read"
	if viewerID != "" && viewerID == user.ID {
		owner.ViewerPermission = "admin"
	}
	return owner, nil
}

func (s *OwnerService) buildOrganizationOwner(ctx context.Context, viewerID string, organization *models.Organization) (*Owner, error) {
	if err := s.auth.AuthorizeOrganizationRead(ctx, viewerID, organization); err != nil {
		return nil, err
	}
	owner := &Owner{
		Type:        "organization",
		ID:          organization.ID,
		Username:    organization.Slug,
		Name:        organization.Name,
		DisplayName: organization.Name,
		Description: organization.Description,
		AvatarURL:   organization.AvatarURL,
		WebsiteURL:  organization.WebsiteURL,
		Location:    organization.Location,
		Visibility:  organization.Visibility,
		Capabilities: OwnerCapabilities{
			Teams:    true,
			People:   true,
			Insights: true,
		},
		CreatedAt: organization.CreatedAt,
		UpdatedAt: organization.UpdatedAt,
	}
	projects, err := s.projects.ListByOrganization(ctx, organization.ID)
	if err != nil {
		return nil, err
	}
	members, err := s.orgMembers.ListByOrganization(ctx, organization.ID)
	if err != nil {
		return nil, err
	}
	teams, err := s.orgTeams.ListByOrganization(ctx, organization.ID)
	if err != nil {
		return nil, err
	}
	owner.Counts.Repositories = int64(len(projects))
	owner.Counts.Projects = int64(len(projects))
	owner.Counts.Members = int64(len(members))
	owner.Counts.Teams = int64(len(teams))
	owner.Counts.Stars = s.countStars(ctx, projects)
	owner.ViewerPermission = "read"
	if viewerID != "" {
		if organization.OwnerID == viewerID {
			owner.ViewerPermission = "admin"
		} else if member, memberErr := s.orgMembers.Get(ctx, organization.ID, viewerID); memberErr == nil && member != nil && member.ID != "" {
			owner.ViewerPermission = string(normalizeRole(member.Role))
		}
	}
	return owner, nil
}

func (s *OwnerService) countStars(ctx context.Context, projects []models.Project) int64 {
	var total int64
	for _, project := range projects {
		count, err := s.stars.CountByProject(ctx, project.ID)
		if err != nil {
			continue
		}
		total += count
	}
	return total
}

// IsNameAvailable reports whether a username or organization slug can be
// claimed. Both share the same namespace.
func (s *OwnerService) IsNameAvailable(ctx context.Context, name string) (bool, error) {
	slug := strings.ToLower(strings.TrimSpace(name))
	if slug == "" {
		return false, nil
	}
	if user, err := s.users.GetByUsername(ctx, slug); err == nil && user != nil && user.ID != "" {
		return false, nil
	} else if err != nil && utils.AsAppError(err).Code != "USER_NOT_FOUND" {
		return false, err
	}
	if organization, err := s.orgs.GetBySlug(ctx, slug); err == nil && organization != nil && organization.ID != "" {
		return false, nil
	} else if err != nil && utils.AsAppError(err).Code != "ORGANIZATION_NOT_FOUND" {
		return false, err
	}
	return true, nil
}

// ValidateName checks the shared owner-name grammar and reserved words.
func ValidateName(name string) error {
	slug := strings.ToLower(strings.TrimSpace(name))
	if len(slug) < usernameMinLength || len(slug) > usernameMaxLength || !usernamePattern.MatchString(slug) {
		return utils.ErrUsernameInvalid
	}
	if _, reserved := reservedUsernames[slug]; reserved {
		return utils.ErrUsernameReserved
	}
	return nil
}

// ClaimUsername assigns a Code handle to an account. It is idempotent and
// rejects names already used by an organization.
func (s *OwnerService) ClaimUsername(ctx context.Context, userID, username string) (*models.User, error) {
	slug := strings.ToLower(strings.TrimSpace(username))
	if err := ValidateName(slug); err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Username != nil && *user.Username == slug {
		return user, nil
	}
	available, err := s.IsNameAvailable(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, utils.ErrUsernameTaken
	}
	user.Username = &slug
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
