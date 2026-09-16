package services

import (
	"context"
	"strings"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

// OrganizationService owns the Code-native organization domain.
type OrganizationService struct {
	orgs       interfaces.OrganizationRepository
	members    interfaces.OrganizationMemberRepository
	teams      interfaces.OrganizationTeamRepository
	teamMember interfaces.OrganizationTeamMemberRepository
	users      interfaces.UserRepository
	events     interfaces.EventBus
	auth       *Authorizer
	owners     *OwnerService
}

// NewOrganizationService builds the organization service.
func NewOrganizationService(
	orgs interfaces.OrganizationRepository,
	members interfaces.OrganizationMemberRepository,
	teams interfaces.OrganizationTeamRepository,
	teamMember interfaces.OrganizationTeamMemberRepository,
	users interfaces.UserRepository,
	events interfaces.EventBus,
	auth *Authorizer,
	owners *OwnerService,
) *OrganizationService {
	return &OrganizationService{
		orgs:       orgs,
		members:    members,
		teams:      teams,
		teamMember: teamMember,
		users:      users,
		events:     events,
		auth:       auth,
		owners:     owners,
	}
}

// OrganizationMember is the API representation of an organization member.
type OrganizationMember struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	Username    string     `json:"username"`
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email,omitempty"`
	AvatarURL   *string    `json:"avatarUrl,omitempty"`
	Role        string     `json:"role"`
	JoinedAt    time.Time  `json:"joinedAt"`
	LastSeenAt  *time.Time `json:"lastSeenAt,omitempty"`
}

// CreateOrganizationInput is the payload accepted by the creation endpoint.
type CreateOrganizationInput struct {
	Name        string
	Slug        string
	Description string
	Visibility  string
	WebsiteURL  string
	Location    string
}

// Create creates an organization and registers its creator as owner.
func (s *OrganizationService) Create(ctx context.Context, principal interfaces.Principal, input CreateOrganizationInput) (*models.Organization, error) {
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = slugifyDisplayName(input.Name)
	}
	if err := ValidateName(slug); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, utils.ErrValidationFailed
	}
	available, err := s.owners.IsNameAvailable(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, utils.ErrOrganizationSlugTaken
	}

	visibility := normalizeVisibility(input.Visibility, "public")
	now := time.Now().UTC()
	organization := &models.Organization{
		Common:      models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		Slug:        slug,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Visibility:  visibility,
		OwnerID:     principal.UserID,
		Location:    strings.TrimSpace(input.Location),
	}
	if website := strings.TrimSpace(input.WebsiteURL); website != "" {
		organization.WebsiteURL = &website
	}
	member := &models.OrganizationMember{
		Common:         models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		OrganizationID: organization.ID,
		UserID:         principal.UserID,
		Role:           string(RoleOwner),
		JoinedAt:       now,
	}

	if err := s.orgs.Create(ctx, organization); err != nil {
		return nil, err
	}
	if err := s.members.Create(ctx, member); err != nil {
		return nil, err
	}
	s.publish(ctx, "organization.created", principal.UserID, map[string]any{
		"organizationId": organization.ID,
		"slug":           organization.Slug,
		"visibility":     organization.Visibility,
	})
	return organization, nil
}

// Get returns an organization by slug or id, enforcing read access.
func (s *OrganizationService) Get(ctx context.Context, principal interfaces.Principal, ref string) (*models.Organization, error) {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if err := s.auth.AuthorizeOrganizationRead(ctx, principal.UserID, organization); err != nil {
		return nil, err
	}
	return organization, nil
}

// Update applies a partial update to an organization.
func (s *OrganizationService) Update(ctx context.Context, principal interfaces.Principal, ref string, input CreateOrganizationInput) (*models.Organization, error) {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionOrganizationAdmin); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) != "" {
		organization.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		organization.Description = strings.TrimSpace(input.Description)
	}
	if input.Visibility != "" {
		organization.Visibility = normalizeVisibility(input.Visibility, organization.Visibility)
	}
	if input.Location != "" {
		organization.Location = strings.TrimSpace(input.Location)
	}
	if input.WebsiteURL != "" {
		website := strings.TrimSpace(input.WebsiteURL)
		organization.WebsiteURL = &website
	}
	organization.UpdatedAt = time.Now().UTC()
	if err := s.orgs.Update(ctx, organization); err != nil {
		return nil, err
	}
	s.publish(ctx, "organization.updated", principal.UserID, map[string]any{"organizationId": organization.ID})
	return organization, nil
}

// Archive soft-deletes an organization. Only the owner may archive it.
func (s *OrganizationService) Archive(ctx context.Context, principal interfaces.Principal, ref string) error {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return err
	}
	if organization.OwnerID != principal.UserID {
		return utils.ErrForbidden
	}
	if err := s.orgs.Archive(ctx, organization.ID, time.Now().UTC()); err != nil {
		return err
	}
	s.publish(ctx, "organization.archived", principal.UserID, map[string]any{"organizationId": organization.ID})
	return nil
}

// ListByViewer lists the organizations the caller belongs to.
func (s *OrganizationService) ListByViewer(ctx context.Context, principal interfaces.Principal) ([]models.Organization, error) {
	return s.orgs.ListByUser(ctx, principal.UserID)
}

// ListMembers returns the members of an organization.
func (s *OrganizationService) ListMembers(ctx context.Context, principal interfaces.Principal, ref string) ([]OrganizationMember, error) {
	organization, err := s.Get(ctx, principal, ref)
	if err != nil {
		return nil, err
	}
	items, err := s.members.ListByOrganization(ctx, organization.ID)
	if err != nil {
		return nil, err
	}
	return s.decorate(ctx, items)
}

// AddMember adds a user to the organization. Only admins and owners may do it.
func (s *OrganizationService) AddMember(ctx context.Context, principal interfaces.Principal, ref, userRef, role string) (*OrganizationMember, error) {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	access, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionMemberManage)
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
	user, err := s.resolveUser(ctx, userRef)
	if err != nil {
		return nil, err
	}
	if existing, existingErr := s.members.Get(ctx, organization.ID, user.ID); existingErr == nil && existing != nil && existing.ID != "" {
		return nil, utils.NewError(409, "ORGANIZATION_MEMBER_EXISTS", "This user is already a member of the organization.", nil)
	}
	now := time.Now().UTC()
	invitedBy := principal.UserID
	member := &models.OrganizationMember{
		Common:         models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		OrganizationID: organization.ID,
		UserID:         user.ID,
		Role:           string(normalizedRole),
		JoinedAt:       now,
		InvitedBy:      &invitedBy,
	}
	if err := s.members.Create(ctx, member); err != nil {
		return nil, err
	}
	s.publish(ctx, "member.added", principal.UserID, map[string]any{
		"organizationId": organization.ID,
		"userId":         user.ID,
		"role":           string(normalizedRole),
		"scope":          "organization",
	})
	return &OrganizationMember{
		ID:          member.ID,
		UserID:      user.ID,
		Username:    derefString(user.Username),
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// UpdateMember changes a member role.
func (s *OrganizationService) UpdateMember(ctx context.Context, principal interfaces.Principal, ref, userRef, role string) (*OrganizationMember, error) {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	access, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionMemberManage)
	if err != nil {
		return nil, err
	}
	user, err := s.resolveUser(ctx, userRef)
	if err != nil {
		return nil, err
	}
	member, err := s.members.Get(ctx, organization.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if !isValidRole(role) {
		return nil, utils.ErrValidationFailed
	}
	normalizedRole := normalizeRole(role)
	if member.Role == string(RoleOwner) && access.Role != RoleOwner {
		return nil, utils.ErrForbidden
	}
	if normalizedRole != RoleOwner && member.Role == string(RoleOwner) {
		owners, countErr := s.members.CountByOrganizationAndRole(ctx, organization.ID, string(RoleOwner))
		if countErr != nil {
			return nil, countErr
		}
		if owners <= 1 {
			return nil, utils.ErrLastOrganizationOwner
		}
	}
	member.Role = string(normalizedRole)
	member.UpdatedAt = time.Now().UTC()
	if err := s.members.Update(ctx, member); err != nil {
		return nil, err
	}
	s.publish(ctx, "member.updated", principal.UserID, map[string]any{"organizationId": organization.ID, "userId": user.ID})
	return &OrganizationMember{
		ID:          member.ID,
		UserID:      user.ID,
		Username:    derefString(user.Username),
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// RemoveMember removes a member, guarding the last owner.
func (s *OrganizationService) RemoveMember(ctx context.Context, principal interfaces.Principal, ref, userRef string) error {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return err
	}
	access, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionMemberManage)
	if err != nil {
		return err
	}
	user, err := s.resolveUser(ctx, userRef)
	if err != nil {
		return err
	}
	member, err := s.members.Get(ctx, organization.ID, user.ID)
	if err != nil {
		return err
	}
	if member.Role == string(RoleOwner) {
		if access.Role != RoleOwner {
			return utils.ErrForbidden
		}
		owners, countErr := s.members.CountByOrganizationAndRole(ctx, organization.ID, string(RoleOwner))
		if countErr != nil {
			return countErr
		}
		if owners <= 1 {
			return utils.ErrLastOrganizationOwner
		}
	}
	if err := s.members.Delete(ctx, organization.ID, user.ID); err != nil {
		return err
	}
	s.publish(ctx, "member.removed", principal.UserID, map[string]any{"organizationId": organization.ID, "userId": user.ID})
	return nil
}

// ── Teams ────────────────────────────────────────────────────────────────────

// CreateTeam creates an organization team.
func (s *OrganizationService) CreateTeam(ctx context.Context, principal interfaces.Principal, ref, name, description string) (*models.OrganizationTeam, error) {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionTeamManage); err != nil {
		return nil, err
	}
	slug := slugifyDisplayName(name)
	if slug == "" {
		return nil, utils.ErrValidationFailed
	}
	if _, err := s.teams.GetBySlug(ctx, organization.ID, slug); err == nil {
		return nil, utils.ErrTeamSlugTaken
	}
	now := time.Now().UTC()
	team := &models.OrganizationTeam{
		Common:         models.Common{ID: utils.NewID(), CreatedAt: now, UpdatedAt: now},
		OrganizationID: organization.ID,
		Slug:           slug,
		Name:           strings.TrimSpace(name),
		Description:    strings.TrimSpace(description),
		Privacy:        "visible",
		Permission:     "read",
	}
	if err := s.teams.Create(ctx, team); err != nil {
		return nil, err
	}
	s.publish(ctx, "organization.team.created", principal.UserID, map[string]any{"teamId": team.ID, "organizationId": organization.ID})
	return team, nil
}

// ListTeams returns the teams of an organization.
func (s *OrganizationService) ListTeams(ctx context.Context, principal interfaces.Principal, ref string) ([]models.OrganizationTeam, error) {
	organization, err := s.Get(ctx, principal, ref)
	if err != nil {
		return nil, err
	}
	return s.teams.ListByOrganization(ctx, organization.ID)
}

// DeleteTeam removes a team.
func (s *OrganizationService) DeleteTeam(ctx context.Context, principal interfaces.Principal, ref, teamSlug string) error {
	organization, err := s.resolve(ctx, ref)
	if err != nil {
		return err
	}
	if _, err := s.auth.AuthorizeOrganization(ctx, principal, organization, ActionTeamManage); err != nil {
		return err
	}
	team, err := s.teams.GetBySlug(ctx, organization.ID, strings.ToLower(strings.TrimSpace(teamSlug)))
	if err != nil {
		return err
	}
	return s.teams.Delete(ctx, team.ID)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (s *OrganizationService) resolve(ctx context.Context, ref string) (*models.Organization, error) {
	value := strings.TrimSpace(ref)
	if value == "" {
		return nil, utils.ErrOrganizationNotFound
	}
	if organization, err := s.orgs.GetBySlug(ctx, value); err == nil && organization.ID != "" {
		return organization, nil
	}
	return s.orgs.GetByID(ctx, value)
}

func (s *OrganizationService) resolveUser(ctx context.Context, ref string) (*models.User, error) {
	value := strings.TrimSpace(ref)
	if value == "" {
		return nil, utils.ErrNotFound
	}
	if strings.Contains(value, "@") {
		return s.users.GetByEmail(ctx, strings.ToLower(value))
	}
	if user, err := s.users.GetByID(ctx, value); err == nil && user.ID != "" {
		return user, nil
	}
	return s.users.GetByUsername(ctx, value)
}

func (s *OrganizationService) decorate(ctx context.Context, items []models.OrganizationMember) ([]OrganizationMember, error) {
	out := make([]OrganizationMember, 0, len(items))
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
	for _, item := range items {
		user := byID[item.UserID]
		out = append(out, OrganizationMember{
			ID:          item.ID,
			UserID:      item.UserID,
			Username:    derefString(user.Username),
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        item.Role,
			JoinedAt:    item.JoinedAt,
			LastSeenAt:  item.LastSeenAt,
		})
	}
	return out, nil
}

// isValidRole accepts the roles exposed by the API for organization and
// project memberships.
func isValidRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin", "maintainer", "member":
		return true
	default:
		return false
	}
}

func normalizeVisibility(value, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "public":
		return "public"
	case "private":
		return "private"
	case "internal":
		return "internal"
	default:
		return fallback
	}
}

func (s *OrganizationService) publish(ctx context.Context, eventType, actorID string, payload map[string]any) {
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
