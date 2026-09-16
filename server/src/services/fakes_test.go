package services

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
)

// The fakes below implement the platform repository interfaces in memory. They
// deliberately reproduce the error contract of the GORM implementations
// (not-found codes, membership errors) so that service tests exercise the same
// branches as production.

// ── Users ────────────────────────────────────────────────────────────────────

type fakeUserRepo struct {
	mu    sync.Mutex
	items map[string]models.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{items: map[string]models.User{}}
}

func (r *fakeUserRepo) add(user models.User) models.User {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user.ID == "" {
		user.ID = utils.NewID()
	}
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
	r.items[user.ID] = user
	return user
}

func (r *fakeUserRepo) Create(_ context.Context, user *models.User) error {
	stored := r.add(*user)
	*user = stored
	return nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.items[id]
	if !ok {
		return &models.User{}, utils.NewError(404, "USER_NOT_FOUND", "not found", nil)
	}
	return &user, nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, user := range r.items {
		if strings.EqualFold(user.Email, email) {
			return &user, nil
		}
	}
	return &models.User{}, utils.NewError(404, "USER_NOT_FOUND", "not found", nil)
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, user := range r.items {
		if user.Username != nil && strings.EqualFold(*user.Username, username) {
			return &user, nil
		}
	}
	return &models.User{}, utils.NewError(404, "USER_NOT_FOUND", "not found", nil)
}

func (r *fakeUserRepo) Search(_ context.Context, query string, offset, limit int) ([]models.User, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := []models.User{}
	for _, user := range r.items {
		if user.Username == nil {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(*user.Username), strings.ToLower(query)) {
			matched = append(matched, user)
		}
	}
	total := int64(len(matched))
	return paginate(matched, offset, limit), total, nil
}

func (r *fakeUserRepo) ListByIDs(_ context.Context, ids []string) ([]models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.User{}
	for _, id := range ids {
		if user, ok := r.items[id]; ok {
			out = append(out, user)
		}
	}
	return out, nil
}

func (r *fakeUserRepo) ListStale(context.Context, time.Time, int) ([]models.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user.UpdatedAt = time.Now().UTC()
	r.items[user.ID] = *user
	return nil
}

// ── Organizations ────────────────────────────────────────────────────────────

type fakeOrgRepo struct {
	mu    sync.Mutex
	items map[string]models.Organization
}

func newFakeOrgRepo() *fakeOrgRepo { return &fakeOrgRepo{items: map[string]models.Organization{}} }

func (r *fakeOrgRepo) add(organization models.Organization) models.Organization {
	r.mu.Lock()
	defer r.mu.Unlock()
	if organization.ID == "" {
		organization.ID = utils.NewID()
	}
	now := time.Now().UTC()
	if organization.CreatedAt.IsZero() {
		organization.CreatedAt = now
	}
	organization.UpdatedAt = now
	r.items[organization.ID] = organization
	return organization
}

func (r *fakeOrgRepo) Create(_ context.Context, organization *models.Organization) error {
	*organization = r.add(*organization)
	return nil
}

func (r *fakeOrgRepo) GetByID(_ context.Context, id string) (*models.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	organization, ok := r.items[id]
	if !ok {
		return &models.Organization{}, utils.ErrOrganizationNotFound
	}
	return &organization, nil
}

func (r *fakeOrgRepo) GetBySlug(_ context.Context, slug string) (*models.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, organization := range r.items {
		if strings.EqualFold(organization.Slug, slug) {
			return &organization, nil
		}
	}
	return &models.Organization{}, utils.ErrOrganizationNotFound
}

func (r *fakeOrgRepo) ListByUser(_ context.Context, userID string) ([]models.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.Organization{}
	for _, organization := range r.items {
		if organization.OwnerID == userID {
			out = append(out, organization)
		}
	}
	return out, nil
}

func (r *fakeOrgRepo) Search(_ context.Context, query string, offset, limit int) ([]models.Organization, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := []models.Organization{}
	for _, organization := range r.items {
		if organization.Visibility != "public" {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(organization.Name), strings.ToLower(query)) {
			matched = append(matched, organization)
		}
	}
	return paginate(matched, offset, limit), int64(len(matched)), nil
}

func (r *fakeOrgRepo) Update(_ context.Context, organization *models.Organization) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	organization.UpdatedAt = time.Now().UTC()
	r.items[organization.ID] = *organization
	return nil
}

func (r *fakeOrgRepo) Archive(_ context.Context, id string, archivedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	organization, ok := r.items[id]
	if !ok {
		return utils.ErrOrganizationNotFound
	}
	organization.ArchivedAt = &archivedAt
	r.items[id] = organization
	return nil
}

type fakeOrgMemberRepo struct {
	mu    sync.Mutex
	items map[string]models.OrganizationMember
}

func newFakeOrgMemberRepo() *fakeOrgMemberRepo {
	return &fakeOrgMemberRepo{items: map[string]models.OrganizationMember{}}
}

func (r *fakeOrgMemberRepo) add(member models.OrganizationMember) models.OrganizationMember {
	r.mu.Lock()
	defer r.mu.Unlock()
	if member.ID == "" {
		member.ID = utils.NewID()
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now().UTC()
	}
	r.items[member.OrganizationID+"|"+member.UserID] = member
	return member
}

func (r *fakeOrgMemberRepo) Create(_ context.Context, member *models.OrganizationMember) error {
	*member = r.add(*member)
	return nil
}

func (r *fakeOrgMemberRepo) Get(_ context.Context, organizationID, userID string) (*models.OrganizationMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	member, ok := r.items[organizationID+"|"+userID]
	if !ok {
		return &models.OrganizationMember{}, utils.ErrMembershipRequired
	}
	return &member, nil
}

func (r *fakeOrgMemberRepo) ListByOrganization(_ context.Context, organizationID string) ([]models.OrganizationMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.OrganizationMember{}
	for _, member := range r.items {
		if member.OrganizationID == organizationID {
			out = append(out, member)
		}
	}
	return out, nil
}

func (r *fakeOrgMemberRepo) Update(_ context.Context, member *models.OrganizationMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[member.OrganizationID+"|"+member.UserID] = *member
	return nil
}

func (r *fakeOrgMemberRepo) Delete(_ context.Context, organizationID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, organizationID+"|"+userID)
	return nil
}

func (r *fakeOrgMemberRepo) CountByOrganizationAndRole(_ context.Context, organizationID, role string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, member := range r.items {
		if member.OrganizationID == organizationID && member.Role == role {
			count++
		}
	}
	return count, nil
}

type fakeOrgTeamRepo struct {
	mu    sync.Mutex
	items map[string]models.OrganizationTeam
}

func newFakeOrgTeamRepo() *fakeOrgTeamRepo {
	return &fakeOrgTeamRepo{items: map[string]models.OrganizationTeam{}}
}

func (r *fakeOrgTeamRepo) Create(_ context.Context, team *models.OrganizationTeam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if team.ID == "" {
		team.ID = utils.NewID()
	}
	r.items[team.ID] = *team
	return nil
}

func (r *fakeOrgTeamRepo) GetByID(_ context.Context, id string) (*models.OrganizationTeam, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	team, ok := r.items[id]
	if !ok {
		return &models.OrganizationTeam{}, utils.ErrTeamNotFound
	}
	return &team, nil
}

func (r *fakeOrgTeamRepo) GetBySlug(_ context.Context, organizationID, slug string) (*models.OrganizationTeam, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, team := range r.items {
		if team.OrganizationID == organizationID && team.Slug == slug {
			return &team, nil
		}
	}
	return &models.OrganizationTeam{}, utils.ErrTeamNotFound
}

func (r *fakeOrgTeamRepo) ListByOrganization(_ context.Context, organizationID string) ([]models.OrganizationTeam, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.OrganizationTeam{}
	for _, team := range r.items {
		if team.OrganizationID == organizationID {
			out = append(out, team)
		}
	}
	return out, nil
}

func (r *fakeOrgTeamRepo) Update(_ context.Context, team *models.OrganizationTeam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[team.ID] = *team
	return nil
}

func (r *fakeOrgTeamRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

type fakeOrgTeamMemberRepo struct {
	mu    sync.Mutex
	items map[string]models.OrganizationTeamMember
}

func newFakeOrgTeamMemberRepo() *fakeOrgTeamMemberRepo {
	return &fakeOrgTeamMemberRepo{items: map[string]models.OrganizationTeamMember{}}
}

func (r *fakeOrgTeamMemberRepo) Create(_ context.Context, member *models.OrganizationTeamMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if member.ID == "" {
		member.ID = utils.NewID()
	}
	r.items[member.TeamID+"|"+member.UserID] = *member
	return nil
}

func (r *fakeOrgTeamMemberRepo) Get(_ context.Context, teamID, userID string) (*models.OrganizationTeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	member, ok := r.items[teamID+"|"+userID]
	if !ok {
		return &models.OrganizationTeamMember{}, utils.ErrMembershipRequired
	}
	return &member, nil
}

func (r *fakeOrgTeamMemberRepo) ListByTeam(_ context.Context, teamID string) ([]models.OrganizationTeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.OrganizationTeamMember{}
	for _, member := range r.items {
		if member.TeamID == teamID {
			out = append(out, member)
		}
	}
	return out, nil
}

func (r *fakeOrgTeamMemberRepo) ListByUser(_ context.Context, userID string) ([]models.OrganizationTeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.OrganizationTeamMember{}
	for _, member := range r.items {
		if member.UserID == userID {
			out = append(out, member)
		}
	}
	return out, nil
}

func (r *fakeOrgTeamMemberRepo) Delete(_ context.Context, teamID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, teamID+"|"+userID)
	return nil
}

// ── Projects ─────────────────────────────────────────────────────────────────

type fakeProjectRepo struct {
	mu    sync.Mutex
	items map[string]models.Project
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{items: map[string]models.Project{}}
}

func (r *fakeProjectRepo) add(project models.Project) models.Project {
	r.mu.Lock()
	defer r.mu.Unlock()
	if project.ID == "" {
		project.ID = utils.NewID()
	}
	now := time.Now().UTC()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = now
	}
	project.UpdatedAt = now
	r.items[project.ID] = project
	return project
}

func (r *fakeProjectRepo) Create(_ context.Context, project *models.Project) error {
	*project = r.add(*project)
	return nil
}

func (r *fakeProjectRepo) GetByID(_ context.Context, id string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.items[id]
	if !ok {
		return &models.Project{}, utils.ErrProjectNotFound
	}
	return &project, nil
}

func (r *fakeProjectRepo) GetByReference(_ context.Context, reference string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, project := range r.items {
		if project.Reference == reference || project.Slug == reference {
			return &project, nil
		}
	}
	return &models.Project{}, utils.ErrProjectNotFound
}

func (r *fakeProjectRepo) GetByNamespaceAndSlug(_ context.Context, namespace, slug string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, project := range r.items {
		if strings.EqualFold(project.Namespace, namespace) && strings.EqualFold(project.Slug, slug) {
			return &project, nil
		}
	}
	return &models.Project{}, utils.ErrProjectNotFound
}

func (r *fakeProjectRepo) ListByUser(_ context.Context, userID string) ([]models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.Project{}
	for _, project := range r.items {
		if project.OwnerID == userID {
			out = append(out, project)
		}
	}
	return out, nil
}

func (r *fakeProjectRepo) ListByOrganization(_ context.Context, organizationID string) ([]models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.Project{}
	for _, project := range r.items {
		if project.OrganizationID != nil && *project.OrganizationID == organizationID {
			out = append(out, project)
		}
	}
	return out, nil
}

func (r *fakeProjectRepo) List(_ context.Context, filter interfaces.ProjectFilter) ([]models.Project, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := []models.Project{}
	for _, project := range r.items {
		if filter.Visibility != "" && project.Visibility != filter.Visibility {
			continue
		}
		if filter.OwnerID != "" && project.OwnerID != filter.OwnerID {
			continue
		}
		if filter.OrganizationID != "" && (project.OrganizationID == nil || *project.OrganizationID != filter.OrganizationID) {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(project.Reference), strings.ToLower(filter.Query)) {
			continue
		}
		matched = append(matched, project)
	}
	return paginate(matched, filter.Offset, filter.Limit), int64(len(matched)), nil
}

func (r *fakeProjectRepo) Search(_ context.Context, query string, offset, limit int) ([]models.Project, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matched := []models.Project{}
	for _, project := range r.items {
		if project.Visibility != "public" {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(project.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(project.Reference), strings.ToLower(query)) {
			matched = append(matched, project)
		}
	}
	return paginate(matched, offset, limit), int64(len(matched)), nil
}

func (r *fakeProjectRepo) Update(_ context.Context, project *models.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	project.UpdatedAt = time.Now().UTC()
	r.items[project.ID] = *project
	return nil
}

func (r *fakeProjectRepo) Archive(_ context.Context, id string, archivedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.items[id]
	if !ok {
		return utils.ErrProjectNotFound
	}
	project.ArchivedAt = &archivedAt
	r.items[id] = project
	return nil
}

type fakeProjectMemberRepo struct {
	mu    sync.Mutex
	items map[string]models.ProjectMember
}

func newFakeProjectMemberRepo() *fakeProjectMemberRepo {
	return &fakeProjectMemberRepo{items: map[string]models.ProjectMember{}}
}

func (r *fakeProjectMemberRepo) add(member models.ProjectMember) models.ProjectMember {
	r.mu.Lock()
	defer r.mu.Unlock()
	if member.ID == "" {
		member.ID = utils.NewID()
	}
	r.items[member.ProjectID+"|"+member.UserID] = member
	return member
}

func (r *fakeProjectMemberRepo) Create(_ context.Context, member *models.ProjectMember) error {
	*member = r.add(*member)
	return nil
}

func (r *fakeProjectMemberRepo) Get(_ context.Context, projectID, userID string) (*models.ProjectMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	member, ok := r.items[projectID+"|"+userID]
	if !ok {
		return &models.ProjectMember{}, utils.ErrMembershipRequired
	}
	return &member, nil
}

func (r *fakeProjectMemberRepo) ListByProject(_ context.Context, projectID string) ([]models.ProjectMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ProjectMember{}
	for _, member := range r.items {
		if member.ProjectID == projectID {
			out = append(out, member)
		}
	}
	return out, nil
}

func (r *fakeProjectMemberRepo) Update(_ context.Context, member *models.ProjectMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[member.ProjectID+"|"+member.UserID] = *member
	return nil
}

func (r *fakeProjectMemberRepo) Delete(_ context.Context, projectID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, projectID+"|"+userID)
	return nil
}

type fakeProjectAssetRepo struct {
	mu    sync.Mutex
	items map[string]models.ProjectAsset
}

func newFakeProjectAssetRepo() *fakeProjectAssetRepo {
	return &fakeProjectAssetRepo{items: map[string]models.ProjectAsset{}}
}

func (r *fakeProjectAssetRepo) Create(_ context.Context, asset *models.ProjectAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if asset.ID == "" {
		asset.ID = utils.NewID()
	}
	r.items[asset.ID] = *asset
	return nil
}

func (r *fakeProjectAssetRepo) GetByID(_ context.Context, id string) (*models.ProjectAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	asset, ok := r.items[id]
	if !ok {
		return &models.ProjectAsset{}, utils.ErrProjectAssetNotFound
	}
	return &asset, nil
}

func (r *fakeProjectAssetRepo) GetByProjectAndSlug(_ context.Context, projectID, kind, slug string) (*models.ProjectAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, asset := range r.items {
		if asset.ProjectID == projectID && asset.Kind == kind && asset.Slug == slug {
			return &asset, nil
		}
	}
	return &models.ProjectAsset{}, utils.ErrProjectAssetNotFound
}

func (r *fakeProjectAssetRepo) ListByProject(_ context.Context, projectID, kind string) ([]models.ProjectAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ProjectAsset{}
	for _, asset := range r.items {
		if asset.ProjectID != projectID {
			continue
		}
		if kind != "" && asset.Kind != kind {
			continue
		}
		out = append(out, asset)
	}
	return out, nil
}

func (r *fakeProjectAssetRepo) Search(_ context.Context, query string, offset, limit int) ([]models.ProjectAsset, int64, error) {
	return nil, 0, nil
}

func (r *fakeProjectAssetRepo) Update(_ context.Context, asset *models.ProjectAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[asset.ID] = *asset
	return nil
}

func (r *fakeProjectAssetRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

type fakeProjectStarRepo struct {
	mu    sync.Mutex
	items map[string]models.ProjectStar
}

func newFakeProjectStarRepo() *fakeProjectStarRepo {
	return &fakeProjectStarRepo{items: map[string]models.ProjectStar{}}
}

func (r *fakeProjectStarRepo) Create(_ context.Context, star *models.ProjectStar) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if star.ID == "" {
		star.ID = utils.NewID()
	}
	r.items[star.ProjectID+"|"+star.UserID] = *star
	return nil
}

func (r *fakeProjectStarRepo) Delete(_ context.Context, projectID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, projectID+"|"+userID)
	return nil
}

func (r *fakeProjectStarRepo) Get(_ context.Context, projectID, userID string) (*models.ProjectStar, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	star, ok := r.items[projectID+"|"+userID]
	if !ok {
		return &models.ProjectStar{}, utils.ErrNotFound
	}
	return &star, nil
}

func (r *fakeProjectStarRepo) ListByUser(_ context.Context, userID string, offset, limit int) ([]models.ProjectStar, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ProjectStar{}
	for _, star := range r.items {
		if star.UserID == userID {
			out = append(out, star)
		}
	}
	return paginate(out, offset, limit), int64(len(out)), nil
}

func (r *fakeProjectStarRepo) CountByProject(_ context.Context, projectID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, star := range r.items {
		if star.ProjectID == projectID {
			count++
		}
	}
	return count, nil
}

type fakeProjectWatchRepo struct {
	mu    sync.Mutex
	items map[string]models.ProjectWatch
}

func newFakeProjectWatchRepo() *fakeProjectWatchRepo {
	return &fakeProjectWatchRepo{items: map[string]models.ProjectWatch{}}
}

func (r *fakeProjectWatchRepo) Upsert(_ context.Context, watch *models.ProjectWatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if watch.ID == "" {
		watch.ID = utils.NewID()
	}
	r.items[watch.ProjectID+"|"+watch.UserID] = *watch
	return nil
}

func (r *fakeProjectWatchRepo) Delete(_ context.Context, projectID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, projectID+"|"+userID)
	return nil
}

func (r *fakeProjectWatchRepo) Get(_ context.Context, projectID, userID string) (*models.ProjectWatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	watch, ok := r.items[projectID+"|"+userID]
	if !ok {
		return &models.ProjectWatch{}, utils.ErrNotFound
	}
	return &watch, nil
}

func (r *fakeProjectWatchRepo) ListByUser(_ context.Context, userID string) ([]models.ProjectWatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ProjectWatch{}
	for _, watch := range r.items {
		if watch.UserID == userID {
			out = append(out, watch)
		}
	}
	return out, nil
}

// ── Bindings, connections and synchronisation ────────────────────────────────

type fakeBindingRepo struct {
	mu    sync.Mutex
	items map[string]models.RepositoryBinding
}

func newFakeBindingRepo() *fakeBindingRepo {
	return &fakeBindingRepo{items: map[string]models.RepositoryBinding{}}
}

func (r *fakeBindingRepo) add(binding models.RepositoryBinding) models.RepositoryBinding {
	r.mu.Lock()
	defer r.mu.Unlock()
	if binding.ID == "" {
		binding.ID = utils.NewID()
	}
	now := time.Now().UTC()
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = now
	}
	binding.UpdatedAt = now
	r.items[binding.ID] = binding
	return binding
}

func (r *fakeBindingRepo) Create(_ context.Context, binding *models.RepositoryBinding) error {
	*binding = r.add(*binding)
	return nil
}

func (r *fakeBindingRepo) GetByID(_ context.Context, id string) (*models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	binding, ok := r.items[id]
	if !ok {
		return &models.RepositoryBinding{}, utils.ErrRepositoryBindingNotFound
	}
	return &binding, nil
}

func (r *fakeBindingRepo) GetByProjectAndProvider(_ context.Context, projectID, provider string) (*models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, binding := range r.items {
		if binding.ProjectID == projectID && binding.Provider == provider {
			return &binding, nil
		}
	}
	return &models.RepositoryBinding{}, utils.ErrRepositoryBindingNotFound
}

func (r *fakeBindingRepo) GetByProjectProviderAndExternalID(_ context.Context, projectID, provider, externalID string) (*models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, binding := range r.items {
		if binding.ProjectID == projectID && binding.Provider == provider && binding.ExternalID == externalID {
			return &binding, nil
		}
	}
	return &models.RepositoryBinding{}, utils.ErrRepositoryBindingNotFound
}

func (r *fakeBindingRepo) GetByExternal(_ context.Context, owner, repo string) (*models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, binding := range r.items {
		if strings.EqualFold(binding.ExternalOwner, owner) && strings.EqualFold(binding.ExternalRepo, repo) {
			return &binding, nil
		}
	}
	return &models.RepositoryBinding{}, utils.ErrRepositoryBindingNotFound
}

func (r *fakeBindingRepo) ListByProject(_ context.Context, projectID string) ([]models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.RepositoryBinding{}
	for _, binding := range r.items {
		if binding.ProjectID == projectID {
			out = append(out, binding)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *fakeBindingRepo) ListSyncable(_ context.Context, staleBefore time.Time, limit int) ([]models.RepositoryBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.RepositoryBinding{}
	for _, binding := range r.items {
		if !binding.SyncEnabled || binding.SyncStatus == models.BindingSyncDetached {
			continue
		}
		if binding.LastSyncedAt != nil && binding.LastSyncedAt.After(staleBefore) {
			continue
		}
		out = append(out, binding)
	}
	return paginate(out, 0, limit), nil
}

func (r *fakeBindingRepo) Update(_ context.Context, binding *models.RepositoryBinding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	binding.UpdatedAt = time.Now().UTC()
	r.items[binding.ID] = *binding
	return nil
}

func (r *fakeBindingRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

type fakeConnectionRepo struct {
	mu    sync.Mutex
	items map[string]models.ProviderConnection
}

func newFakeConnectionRepo() *fakeConnectionRepo {
	return &fakeConnectionRepo{items: map[string]models.ProviderConnection{}}
}

func (r *fakeConnectionRepo) Create(_ context.Context, connection *models.ProviderConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if connection.ID == "" {
		connection.ID = utils.NewID()
	}
	r.items[connection.ID] = *connection
	return nil
}

func (r *fakeConnectionRepo) GetByID(_ context.Context, id string) (*models.ProviderConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	connection, ok := r.items[id]
	if !ok {
		return &models.ProviderConnection{}, utils.ErrProviderConnectionNotFound
	}
	return &connection, nil
}

func (r *fakeConnectionRepo) GetByProviderAccount(_ context.Context, provider, providerAccountID string) (*models.ProviderConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, connection := range r.items {
		if connection.Provider == provider && connection.ProviderAccountID == providerAccountID {
			return &connection, nil
		}
	}
	return &models.ProviderConnection{}, utils.ErrProviderConnectionNotFound
}

func (r *fakeConnectionRepo) GetByUserAndProvider(_ context.Context, userID, provider string) (*models.ProviderConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, connection := range r.items {
		if connection.UserID == userID && connection.Provider == provider && connection.Status == models.ConnectionActive {
			return &connection, nil
		}
	}
	return &models.ProviderConnection{}, utils.ErrProviderConnectionNotFound
}

func (r *fakeConnectionRepo) ListByUser(_ context.Context, userID string) ([]models.ProviderConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ProviderConnection{}
	for _, connection := range r.items {
		if connection.UserID == userID {
			out = append(out, connection)
		}
	}
	return out, nil
}

func (r *fakeConnectionRepo) Update(_ context.Context, connection *models.ProviderConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[connection.ID] = *connection
	return nil
}

func (r *fakeConnectionRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

type fakeCursorRepo struct {
	mu    sync.Mutex
	items map[string]models.SyncCursor
}

func newFakeCursorRepo() *fakeCursorRepo {
	return &fakeCursorRepo{items: map[string]models.SyncCursor{}}
}

func (r *fakeCursorRepo) Get(_ context.Context, bindingID, resource string) (*models.SyncCursor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cursor, ok := r.items[bindingID+"|"+resource]
	if !ok {
		return &models.SyncCursor{}, utils.ErrSyncCursorNotFound
	}
	return &cursor, nil
}

func (r *fakeCursorRepo) ListByBinding(_ context.Context, bindingID string) ([]models.SyncCursor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.SyncCursor{}
	for _, cursor := range r.items {
		if cursor.BindingID == bindingID {
			out = append(out, cursor)
		}
	}
	return out, nil
}

func (r *fakeCursorRepo) Upsert(_ context.Context, cursor *models.SyncCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cursor.ID == "" {
		cursor.ID = utils.NewID()
	}
	r.items[cursor.BindingID+"|"+cursor.Resource] = *cursor
	return nil
}

type fakeSubscriptionRepo struct {
	mu    sync.Mutex
	items map[string]models.WebhookSubscription
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{items: map[string]models.WebhookSubscription{}}
}

func (r *fakeSubscriptionRepo) Create(_ context.Context, subscription *models.WebhookSubscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if subscription.ID == "" {
		subscription.ID = utils.NewID()
	}
	r.items[subscription.ID] = *subscription
	return nil
}

func (r *fakeSubscriptionRepo) GetByID(_ context.Context, id string) (*models.WebhookSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	subscription, ok := r.items[id]
	if !ok {
		return &models.WebhookSubscription{}, utils.ErrWebhookSubscriptionNotFound
	}
	return &subscription, nil
}

func (r *fakeSubscriptionRepo) GetByBindingAndExternal(_ context.Context, bindingID, externalID string) (*models.WebhookSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, subscription := range r.items {
		if subscription.BindingID == bindingID && subscription.ExternalID == externalID {
			return &subscription, nil
		}
	}
	return &models.WebhookSubscription{}, utils.ErrWebhookSubscriptionNotFound
}

func (r *fakeSubscriptionRepo) ListByBinding(_ context.Context, bindingID string) ([]models.WebhookSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.WebhookSubscription{}
	for _, subscription := range r.items {
		if subscription.BindingID == bindingID {
			out = append(out, subscription)
		}
	}
	return out, nil
}

func (r *fakeSubscriptionRepo) ListActive(context.Context, int) ([]models.WebhookSubscription, error) {
	return nil, nil
}

func (r *fakeSubscriptionRepo) Update(_ context.Context, subscription *models.WebhookSubscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[subscription.ID] = *subscription
	return nil
}

func (r *fakeSubscriptionRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

type fakeJobRepo struct {
	mu    sync.Mutex
	items map[string]models.SyncJob
}

func newFakeJobRepo() *fakeJobRepo { return &fakeJobRepo{items: map[string]models.SyncJob{}} }

func (r *fakeJobRepo) Create(_ context.Context, job *models.SyncJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.items {
		if existing.IdempotencyKey == job.IdempotencyKey {
			return utils.ErrValidationFailed
		}
	}
	if job.ID == "" {
		job.ID = utils.NewID()
	}
	r.items[job.ID] = *job
	return nil
}

func (r *fakeJobRepo) GetByID(_ context.Context, id string) (*models.SyncJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.items[id]
	if !ok {
		return &models.SyncJob{}, utils.ErrSyncJobNotFound
	}
	return &job, nil
}

func (r *fakeJobRepo) GetByIdempotencyKey(_ context.Context, key string) (*models.SyncJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, job := range r.items {
		if job.IdempotencyKey == key {
			return &job, nil
		}
	}
	return &models.SyncJob{}, utils.ErrSyncJobNotFound
}

func (r *fakeJobRepo) ClaimNext(_ context.Context, workerID string, now time.Time, lockFor time.Duration) (*models.SyncJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, job := range r.items {
		if job.Status != models.SyncJobQueued && job.Status != models.SyncJobFailed {
			continue
		}
		if job.RunAfter.After(now) {
			continue
		}
		if job.LockedUntil != nil && job.LockedUntil.After(now) && job.LockedBy != workerID {
			continue
		}
		lockedUntil := now.Add(lockFor)
		job.Status = models.SyncJobRunning
		job.Attempts++
		job.LockedBy = workerID
		job.LockedUntil = &lockedUntil
		job.StartedAt = &now
		r.items[job.ID] = job
		return &job, nil
	}
	return &models.SyncJob{}, utils.ErrSyncJobNotFound
}

func (r *fakeJobRepo) Update(_ context.Context, job *models.SyncJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[job.ID] = *job
	return nil
}

func (r *fakeJobRepo) ListByStatus(_ context.Context, status string, limit int) ([]models.SyncJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.SyncJob{}
	for _, job := range r.items {
		if job.Status == status {
			out = append(out, job)
		}
	}
	return paginate(out, 0, limit), nil
}

func (r *fakeJobRepo) CountByStatus(_ context.Context, status string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, job := range r.items {
		if job.Status == status {
			count++
		}
	}
	return count, nil
}

type fakeResourceRepo struct {
	mu    sync.Mutex
	items map[string]models.ExternalResource
}

func newFakeResourceRepo() *fakeResourceRepo {
	return &fakeResourceRepo{items: map[string]models.ExternalResource{}}
}

func (r *fakeResourceRepo) Upsert(_ context.Context, resource *models.ExternalResource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if resource.ID == "" {
		resource.ID = utils.NewID()
	}
	r.items[resource.BindingID+"|"+resource.Kind+"|"+resource.ExternalID] = *resource
	return nil
}

func (r *fakeResourceRepo) GetByBindingAndExternal(_ context.Context, bindingID, kind, externalID string) (*models.ExternalResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resource, ok := r.items[bindingID+"|"+kind+"|"+externalID]
	if !ok {
		return &models.ExternalResource{}, utils.ErrExternalResourceNotFound
	}
	return &resource, nil
}

func (r *fakeResourceRepo) ListByProject(_ context.Context, projectID, kind string, offset, limit int) ([]models.ExternalResource, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []models.ExternalResource{}
	for _, resource := range r.items {
		if resource.ProjectID != projectID {
			continue
		}
		if kind != "" && resource.Kind != kind {
			continue
		}
		out = append(out, resource)
	}
	return paginate(out, offset, limit), int64(len(out)), nil
}

func (r *fakeResourceRepo) DeleteByBinding(_ context.Context, bindingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, resource := range r.items {
		if resource.BindingID == bindingID {
			delete(r.items, key)
		}
	}
	return nil
}

// paginate applies offset/limit to an in-memory slice.
func paginate[T any](items []T, offset, limit int) []T {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []T{}
	}
	end := len(items)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return items[offset:end]
}
