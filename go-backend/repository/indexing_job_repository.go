package repository

import (
	"context"
	"mcpdocs/models"

	"gorm.io/gorm"
)

type IndexingJobRepository struct {
	db *gorm.DB
}

func NewIndexingJobRepository(db *gorm.DB) *IndexingJobRepository {
	return &IndexingJobRepository{db: db}
}

func (r *IndexingJobRepository) CreateIndexingJob(ctx context.Context, job *models.IndexingJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *IndexingJobRepository) GetIndexingJobByID(ctx context.Context, jobID uint) (*models.IndexingJob, error) {
	var job models.IndexingJob
	if err := r.db.WithContext(ctx).First(&job, jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *IndexingJobRepository) GetIndexingJobsByRequestID(ctx context.Context, requestID uint) ([]*models.IndexingJob, error) {
	var jobs []*models.IndexingJob
	if err := r.db.WithContext(ctx).Where("request_id = ?", requestID).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *IndexingJobRepository) UpdateIndexingJob(ctx context.Context, job *models.IndexingJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}
