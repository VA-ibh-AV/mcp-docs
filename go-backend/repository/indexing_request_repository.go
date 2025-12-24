package repository

import (
	"context"
	"mcpdocs/models"

	"gorm.io/gorm"
)

type IndexingRequestRepository struct {
	db *gorm.DB
}

func NewIndexingRequestRepository(db *gorm.DB) *IndexingRequestRepository {
	return &IndexingRequestRepository{db: db}
}

func (r *IndexingRequestRepository) CreateIndexingRequest(ctx context.Context, request *models.IndexingRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *IndexingRequestRepository) GetIndexingRequestByID(ctx context.Context, requestID uint) (*models.IndexingRequest, error) {
	var request models.IndexingRequest
	if err := r.db.WithContext(ctx).First(&request, requestID).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *IndexingRequestRepository) GetIndexingRequestsByProjectID(ctx context.Context, projectID uint) ([]*models.IndexingRequest, error) {
	var requests []*models.IndexingRequest
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *IndexingRequestRepository) UpdateIndexingRequest(ctx context.Context, request *models.IndexingRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

func (r *IndexingRequestRepository) UpdateTotalJobs(ctx context.Context, requestID uint, totalJobs int) error {
	return r.db.WithContext(ctx).Model(&models.IndexingRequest{}).Where("id = ?", requestID).Update("total_jobs", totalJobs).Error
}

