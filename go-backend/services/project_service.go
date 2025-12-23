package services

import (
	"context"
	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"time"
)

type ProjectServiceInterface interface {
	CreateProject(ctx context.Context, userID string, req schema.CreateProjectRequest) (*models.Project, error)
	GetProjects(ctx context.Context, userID string) ([]*models.Project, error)
	GetProject(ctx context.Context, projectID uint) (*models.Project, error)
	UpdateProject(ctx context.Context, projectID uint, req schema.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, projectID uint) error
}

type ProjectService struct {
	projectRepository repository.ProjectRepositoryInterface
}

func NewProjectService(projectRepository repository.ProjectRepositoryInterface) *ProjectService {
	return &ProjectService{projectRepository: projectRepository}
}

func (s *ProjectService) CreateProject(ctx context.Context, userID string, req schema.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Url:         req.Url,
		Status:      "pending", // Default status
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.projectRepository.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetProjects(ctx context.Context, userID string) ([]*models.Project, error) {
	return s.projectRepository.GetProjectsByUserID(ctx, userID)
}

func (s *ProjectService) GetProject(ctx context.Context, projectID uint) (*models.Project, error) {
	return s.projectRepository.GetProjectByID(ctx, projectID)
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID uint, req schema.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.projectRepository.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Url != "" {
		project.Url = req.Url
	}
	if req.Status != "" {
		project.Status = req.Status
	}
	project.UpdatedAt = time.Now()

	if err := s.projectRepository.UpdateProject(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID uint) error {
	return s.projectRepository.DeleteProject(ctx, projectID)
}
