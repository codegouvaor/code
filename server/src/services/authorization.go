package services

import (
	"context"
	"strings"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

// Role is a platform role. Roles are deliberately shared by organizations and
// projects so that authorization stays a single, auditable table of ranks
// instead of per-endpoint conditionals.
type Role string

const (
	RolePublic     Role = "public"
	RolePrivate    Role = "private"
	RoleMember     Role = "member"
	RoleMaintainer Role = "maintainer"
	RoleAdmin      Role = "admin"
	RoleOwner      Role = "owner"
)

// Access levels are the numeric ranks behind the roles.
const (
	levelNone       = 0
	levelPublic     = 1
	levelMember     = 3
	levelMaintainer = 4
	levelAdmin      = 5
	levelOwner      = 6
)

func roleLevel(role Role) int {
	switch strings.ToLower(string(role)) {
	case string(RoleOwner):
		return levelOwner
	case string(RoleAdmin):
		return levelAdmin
	case string(RoleMaintainer):
		return levelMaintainer
	case string(RoleMember):
		return levelMember
	case string(RolePublic), string(RolePrivate):
		return levelPublic
	default:
		return levelNone
	}
}

func levelRole(level int) Role {
	switch {
	case level >= levelOwner:
		return RoleOwner
	case level >= levelAdmin:
		return RoleAdmin
	case level >= levelMaintainer:
		return RoleMaintainer
	case level >= levelMember:
		return RoleMember
	case level >= levelPublic:
		return RolePublic
	default:
		return RolePrivate
	}
}

// Action is what a caller tries to do. Handlers never decide: they ask the
// authorizer, which knows the role model.
type Action string

const (
	ActionProjectRead       Action = "project:read"
	ActionProjectWrite      Action = "project:write"
	ActionProjectAdmin      Action = "project:admin"
	ActionProjectDelete     Action = "project:delete"
	ActionRepositorySync    Action = "repository:sync"
	ActionRepositoryConnect Action = "repository:connect"
	ActionAssetManage       Action = "asset:manage"
	ActionMemberManage      Action = "member:manage"
	ActionOrganizationRead  Action = "organization:read"
	ActionOrganizationAdmin Action = "organization:admin"
	ActionTeamManage        Action = "team:manage"
)

// requiredLevel is the single source of truth for the permission matrix.
func requiredLevel(action Action) int {
	switch action {
	case ActionProjectRead, ActionOrganizationRead:
		return levelMember
	case ActionProjectWrite:
		return levelMember
	case ActionRepositorySync, ActionAssetManage:
		return levelMaintainer
	case ActionProjectAdmin, ActionRepositoryConnect, ActionMemberManage, ActionOrganizationAdmin, ActionTeamManage:
		return levelAdmin
	case ActionProjectDelete:
		return levelOwner
	default:
		return levelOwner
	}
}

// ProjectAccess is the effective access a caller has on a project.
type ProjectAccess struct {
	// Role is the effective role of the caller.
	Role Role
	// Level is the numeric rank behind Role.
	Level int
	// Visibility explains where the access comes from.
	Visibility string
	// IsMember reports whether a membership row (direct or team) was found.
	IsMember bool
	// IsOwner reports whether the caller owns the resource.
	IsOwner bool
}

// Can reports whether the access satisfies an action. Reading a public project
// is allowed to everyone, including anonymous callers; everything else follows
// the role matrix.
func (a ProjectAccess) Can(action Action) bool {
	if action == ActionProjectRead && a.Visibility == "public" && a.Level >= levelPublic {
		return true
	}
	return a.Level >= requiredLevel(action)
}

// Authorizer centralizes every authorization decision of the platform domain.
// It resolves the effective role of a caller from ownership, project
// membership, organization membership and public visibility.
type Authorizer struct {
	projects       interfaces.ProjectRepository
	projectMembers interfaces.ProjectMemberRepository
	organizations  interfaces.OrganizationRepository
	orgMembers     interfaces.OrganizationMemberRepository
	orgTeamMembers interfaces.OrganizationTeamMemberRepository
}

// NewAuthorizer wires the authorizer with its repositories.
func NewAuthorizer(
	projects interfaces.ProjectRepository,
	projectMembers interfaces.ProjectMemberRepository,
	organizations interfaces.OrganizationRepository,
	orgMembers interfaces.OrganizationMemberRepository,
	orgTeamMembers interfaces.OrganizationTeamMemberRepository,
) *Authorizer {
	return &Authorizer{
		projects:       projects,
		projectMembers: projectMembers,
		organizations:  organizations,
		orgMembers:     orgMembers,
		orgTeamMembers: orgTeamMembers,
	}
}

// ProjectAccess computes the effective access of a user on a project. It never
// fails on a missing membership: absence of membership is a valid answer.
func (a *Authorizer) ProjectAccess(ctx context.Context, userID string, project *models.Project) (ProjectAccess, error) {
	access := ProjectAccess{Role: RolePrivate, Level: levelNone, Visibility: project.Visibility}
	if project == nil {
		return access, nil
	}
	if access.Visibility == "" {
		access.Visibility = "private"
	}

	if userID != "" {
		if project.OwnerID != "" && project.OwnerID == userID {
			access.Role = RoleOwner
			access.Level = levelOwner
			access.IsMember = true
			access.IsOwner = true
			return access, nil
		}
		member, err := a.projectMembers.Get(ctx, project.ID, userID)
		if err == nil && member != nil && member.ID != "" {
			access.Role = normalizeRole(member.Role)
			access.Level = max(access.Level, roleLevel(access.Role))
			access.IsMember = true
		} else if err != nil && utils.AsAppError(err).Code != "MEMBERSHIP_REQUIRED" {
			return access, err
		}

		if project.OrganizationID != nil && *project.OrganizationID != "" {
			organization, orgErr := a.organizations.GetByID(ctx, *project.OrganizationID)
			if orgErr != nil {
				return access, orgErr
			}
			if organization.OwnerID == userID {
				access.Role = RoleOwner
				access.Level = levelOwner
				access.IsOwner = true
			} else if orgMember, memberErr := a.orgMembers.Get(ctx, organization.ID, userID); memberErr == nil && orgMember != nil && orgMember.ID != "" {
				orgLevel := roleLevel(normalizeRole(orgMember.Role))
				access.Level = max(access.Level, orgLevel)
				access.IsMember = true
			} else if memberErr != nil && utils.AsAppError(memberErr).Code != "MEMBERSHIP_REQUIRED" {
				return access, memberErr
			}
			// Team membership widens access of organization members that are
			// not direct organization members yet (for example an outside
			// collaborator invited into one team).
			if access.Level < levelMember && a.orgTeamMembers != nil {
				if teams, teamErr := a.orgTeamMembers.ListByUser(ctx, userID); teamErr == nil && len(teams) > 0 {
					access.IsMember = true
					access.Level = max(access.Level, levelMember)
				}
			}
		}
	}

	if access.Level == levelNone && access.Visibility == "public" {
		access.Level = levelPublic
	}
	access.Role = levelRole(access.Level)
	return access, nil
}

// AuthorizeProject enforces an action on a project, returning the API error to
// send when the caller is not allowed.
func (a *Authorizer) AuthorizeProject(ctx context.Context, principal interfaces.Principal, project *models.Project, action Action) (ProjectAccess, error) {
	access, err := a.ProjectAccess(ctx, principal.UserID, project)
	if err != nil {
		return access, err
	}
	if access.Can(action) {
		return access, nil
	}
	if access.Level == levelNone {
		return access, utils.ErrUnauthorized
	}
	return access, utils.ErrForbidden
}

// AuthorizeOrganizationRead checks that the caller may see an organization.
func (a *Authorizer) AuthorizeOrganizationRead(ctx context.Context, userID string, organization *models.Organization) error {
	if organization == nil {
		return utils.ErrOrganizationNotFound
	}
	if organization.Visibility == "public" {
		return nil
	}
	if userID == "" {
		return utils.ErrUnauthorized
	}
	if organization.OwnerID == userID {
		return nil
	}
	if _, err := a.orgMembers.Get(ctx, organization.ID, userID); err == nil {
		return nil
	}
	return utils.ErrForbidden
}

// AuthorizeOrganization enforces an administrative action on an organization.
func (a *Authorizer) AuthorizeOrganization(ctx context.Context, principal interfaces.Principal, organization *models.Organization, action Action) (ProjectAccess, error) {
	access := ProjectAccess{Role: RolePrivate, Level: levelNone, Visibility: organization.Visibility}
	if organization.OwnerID != "" && organization.OwnerID == principal.UserID {
		access.Role = RoleOwner
		access.Level = levelOwner
		access.IsOwner = true
		access.IsMember = true
	} else if principal.UserID != "" {
		if member, err := a.orgMembers.Get(ctx, organization.ID, principal.UserID); err == nil && member != nil && member.ID != "" {
			access.Role = normalizeRole(member.Role)
			access.Level = roleLevel(access.Role)
			access.IsMember = true
		}
	}
	if access.Can(action) {
		return access, nil
	}
	if access.Level == levelNone {
		return access, utils.ErrUnauthorized
	}
	return access, utils.ErrForbidden
}

func normalizeRole(role string) Role {
	normalized := strings.ToLower(strings.TrimSpace(role))
	switch normalized {
	case string(RoleOwner), string(RoleAdmin), string(RoleMaintainer), string(RoleMember):
		return Role(normalized)
	case "read", "reader", "guest":
		return RoleMember
	case "write", "writer":
		return RoleMaintainer
	default:
		return RoleMember
	}
}
