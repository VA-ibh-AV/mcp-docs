package repository

import (
	"context"
	"mcpdocs/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) GetProjectsByUserID(ctx context.Context, userID string) ([]*models.Project, error) {
	var projects []*models.Project
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, projectID uint) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).First(&project, projectID).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, projectID).Error
}
