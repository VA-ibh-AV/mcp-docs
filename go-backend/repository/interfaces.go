package repository

import (
	"context"
	"mcpdocs/models"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*models.User, error)
}

// RefreshTokenRepositoryInterface defines the interface for refresh token repository operations
type RefreshTokenRepositoryInterface interface {
	Create(ctx context.Context, userID, token string) (*models.RefreshToken, error)
	FindValid(ctx context.Context, token string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// ProjectRepositoryInterface defines the interface for project repository operations
type ProjectRepositoryInterface interface {
	CreateProject(ctx context.Context, project *models.Project) error
	GetProjectsByUserID(ctx context.Context, userID string) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, projectID uint) (*models.Project, error)
	UpdateProject(ctx context.Context, project *models.Project) error
	DeleteProject(ctx context.Context, projectID uint) error
}
