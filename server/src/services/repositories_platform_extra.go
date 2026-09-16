package services

import (
	"context"

	"github.com/codegouvaor/code/server/src/interfaces"
	"github.com/codegouvaor/code/server/src/models"
	"github.com/codegouvaor/code/server/src/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repositories) OrganizationTeams() interfaces.OrganizationTeamRepository {
	return &organizationTeamRepository{db: r.db}
}

func (r *Repositories) OrganizationTeamMembers() interfaces.OrganizationTeamMemberRepository {
	return &organizationTeamMemberRepository{db: r.db}
}

func (r *Repositories) ProjectStars() interfaces.ProjectStarRepository {
	return &projectStarRepository{db: r.db}
}

func (r *Repositories) ProjectWatches() interfaces.ProjectWatchRepository {
	return &projectWatchRepository{db: r.db}
}

type organizationTeamRepository struct{ db *gorm.DB }

func (r *organizationTeamRepository) Create(ctx context.Context, team *models.OrganizationTeam) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *organizationTeamRepository) GetByID(ctx context.Context, id string) (*models.OrganizationTeam, error) {
	var item models.OrganizationTeam
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	return &item, normalizeNotFound(err, utils.ErrTeamNotFound)
}

func (r *organizationTeamRepository) GetBySlug(ctx context.Context, organizationID, slug string) (*models.OrganizationTeam, error) {
	var item models.OrganizationTeam
	err := r.db.WithContext(ctx).First(&item, "organization_id = ? AND slug = ?", organizationID, slug).Error
	return &item, normalizeNotFound(err, utils.ErrTeamNotFound)
}

func (r *organizationTeamRepository) ListByOrganization(ctx context.Context, organizationID string) ([]models.OrganizationTeam, error) {
	var items []models.OrganizationTeam
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("name asc").Find(&items).Error
	return items, err
}

func (r *organizationTeamRepository) Update(ctx context.Context, team *models.OrganizationTeam) error {
	return r.db.WithContext(ctx).Save(team).Error
}

func (r *organizationTeamRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.OrganizationTeam{}, "id = ?", id).Error
}

type organizationTeamMemberRepository struct{ db *gorm.DB }

func (r *organizationTeamMemberRepository) Create(ctx context.Context, member *models.OrganizationTeamMember) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "user_id"}},
		UpdateAll: true,
	}).Create(member).Error
}

func (r *organizationTeamMemberRepository) Get(ctx context.Context, teamID, userID string) (*models.OrganizationTeamMember, error) {
	var item models.OrganizationTeamMember
	err := r.db.WithContext(ctx).First(&item, "team_id = ? AND user_id = ?", teamID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrMembershipRequired)
}

func (r *organizationTeamMemberRepository) ListByTeam(ctx context.Context, teamID string) ([]models.OrganizationTeamMember, error) {
	var items []models.OrganizationTeamMember
	err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Order("joined_at asc").Find(&items).Error
	return items, err
}

func (r *organizationTeamMemberRepository) ListByUser(ctx context.Context, userID string) ([]models.OrganizationTeamMember, error) {
	var items []models.OrganizationTeamMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *organizationTeamMemberRepository) Delete(ctx context.Context, teamID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.OrganizationTeamMember{}, "team_id = ? AND user_id = ?", teamID, userID).Error
}

type projectStarRepository struct{ db *gorm.DB }

func (r *projectStarRepository) Create(ctx context.Context, star *models.ProjectStar) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(star).Error
}

func (r *projectStarRepository) Delete(ctx context.Context, projectID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectStar{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}

func (r *projectStarRepository) Get(ctx context.Context, projectID, userID string) (*models.ProjectStar, error) {
	var item models.ProjectStar
	err := r.db.WithContext(ctx).First(&item, "project_id = ? AND user_id = ?", projectID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrNotFound)
}

func (r *projectStarRepository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]models.ProjectStar, int64, error) {
	var items []models.ProjectStar
	var count int64
	builder := r.db.WithContext(ctx).Model(&models.ProjectStar{}).Where("user_id = ?", userID)
	if err := builder.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := builder.Order("starred_at desc").Offset(offset).Limit(limit).Find(&items).Error
	return items, count, err
}

func (r *projectStarRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ProjectStar{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

type projectWatchRepository struct{ db *gorm.DB }

func (r *projectWatchRepository) Upsert(ctx context.Context, watch *models.ProjectWatch) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "user_id"}},
		UpdateAll: true,
	}).Create(watch).Error
}

func (r *projectWatchRepository) Delete(ctx context.Context, projectID, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectWatch{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}

func (r *projectWatchRepository) Get(ctx context.Context, projectID, userID string) (*models.ProjectWatch, error) {
	var item models.ProjectWatch
	err := r.db.WithContext(ctx).First(&item, "project_id = ? AND user_id = ?", projectID, userID).Error
	return &item, normalizeNotFound(err, utils.ErrNotFound)
}

func (r *projectWatchRepository) ListByUser(ctx context.Context, userID string) ([]models.ProjectWatch, error) {
	var items []models.ProjectWatch
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}
