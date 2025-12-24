package services

import (
	"context"
	"encoding/json"
	"mcpdocs/kafka"
	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"net/url"
	"time"
)

type IndexingServiceInterface interface {
	CreateIndexingRequest(ctx context.Context, userID string, req schema.CreateIndexingRequestRequest) (*models.IndexingRequest, error)
	GetIndexingRequest(ctx context.Context, requestID uint) (*models.IndexingRequest, error)
	GetIndexingRequestsByProject(ctx context.Context, projectID uint) ([]*models.IndexingRequest, error)
	UpdateIndexingRequestStatus(ctx context.Context, requestID uint, req schema.UpdateIndexingRequestStatusRequest) (*models.IndexingRequest, error)
	UpdateTotalJobs(ctx context.Context, requestID uint, totalJobs int) error

	CreateIndexingJob(ctx context.Context, req schema.CreateIndexingJobRequest) (*models.IndexingJob, error)
	GetIndexingJob(ctx context.Context, jobID uint) (*models.IndexingJob, error)
	GetIndexingJobsByRequest(ctx context.Context, requestID uint) ([]*models.IndexingJob, error)
	UpdateIndexingJobStatus(ctx context.Context, jobID uint, req schema.UpdateIndexingJobStatusRequest) (*models.IndexingJob, error)
}

type IndexingService struct {
	requestRepo repository.IndexingRequestRepositoryInterface
	jobRepo     repository.IndexingJobRepositoryInterface
	producer    *kafka.Producer
}

func NewIndexingService(requestRepo repository.IndexingRequestRepositoryInterface, jobRepo repository.IndexingJobRepositoryInterface, producer *kafka.Producer) *IndexingService {
	return &IndexingService{
		requestRepo: requestRepo,
		jobRepo:     jobRepo,
		producer:    producer,
	}
}

func (s *IndexingService) CreateIndexingRequest(ctx context.Context, userID string, req schema.CreateIndexingRequestRequest) (*models.IndexingRequest, error) {
	request := &models.IndexingRequest{
		UserID:    userID,
		ProjectID: req.ProjectID,
		Endpoint:  req.Endpoint,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.requestRepo.CreateIndexingRequest(ctx, request); err != nil {
		return nil, err
	}

	// Push to Kafka
	if s.producer != nil {
		u, err := url.Parse(req.Endpoint)
		if err == nil {
			domain := u.Hostname()
			payload, _ := json.Marshal(request)
			s.producer.SendMessage("indexing_requests", domain, string(payload))
		}
	}

	return request, nil
}

func (s *IndexingService) GetIndexingRequest(ctx context.Context, requestID uint) (*models.IndexingRequest, error) {
	return s.requestRepo.GetIndexingRequestByID(ctx, requestID)
}

func (s *IndexingService) GetIndexingRequestsByProject(ctx context.Context, projectID uint) ([]*models.IndexingRequest, error) {
	return s.requestRepo.GetIndexingRequestsByProjectID(ctx, projectID)
}

func (s *IndexingService) UpdateIndexingRequestStatus(ctx context.Context, requestID uint, req schema.UpdateIndexingRequestStatusRequest) (*models.IndexingRequest, error) {
	request, err := s.requestRepo.GetIndexingRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "" {
		request.Status = req.Status
	}
	if req.ErrorMsg != "" {
		request.ErrorMsg = req.ErrorMsg
	}
	if req.IncrementCompletedJobs {
		request.CompletedJobs++
	}
	request.UpdatedAt = time.Now()
	if err := s.requestRepo.UpdateIndexingRequest(ctx, request); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *IndexingService) UpdateTotalJobs(ctx context.Context, requestID uint, totalJobs int) error {
	return s.requestRepo.UpdateTotalJobs(ctx, requestID, totalJobs)
}

func (s *IndexingService) CreateIndexingJob(ctx context.Context, req schema.CreateIndexingJobRequest) (*models.IndexingJob, error) {
	job := &models.IndexingJob{
		ProjectID: req.ProjectID,
		RequestID: req.RequestID,
		Url:       req.Url,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.jobRepo.CreateIndexingJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *IndexingService) GetIndexingJob(ctx context.Context, jobID uint) (*models.IndexingJob, error) {
	return s.jobRepo.GetIndexingJobByID(ctx, jobID)
}

func (s *IndexingService) GetIndexingJobsByRequest(ctx context.Context, requestID uint) ([]*models.IndexingJob, error) {
	return s.jobRepo.GetIndexingJobsByRequestID(ctx, requestID)
}

func (s *IndexingService) UpdateIndexingJobStatus(ctx context.Context, jobID uint, req schema.UpdateIndexingJobStatusRequest) (*models.IndexingJob, error) {
	job, err := s.jobRepo.GetIndexingJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if req.Status != "" {
		job.Status = req.Status
	}
	if req.ErrorMsg != "" {
		job.ErrorMsg = req.ErrorMsg
	}
	job.UpdatedAt = time.Now()
	if err := s.jobRepo.UpdateIndexingJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}
