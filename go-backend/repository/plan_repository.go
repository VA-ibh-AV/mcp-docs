package repository

import (
	"mcpdocs/models"

	"gorm.io/gorm"
)

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) GetPlanByID(planID uint) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.First(&plan, planID).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) CreatePlan(plan *models.Plan) error {
	return r.db.Create(plan).Error
}

func (r *PlanRepository) UpdatePlan(plan *models.Plan) error {
	return r.db.Save(plan).Error
}

func (r *PlanRepository) DeletePlan(planID uint) error {
	return r.db.Delete(&models.Plan{}, planID).Error
}

func (r *PlanRepository) ListPlans() ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.Find(&plans).Error
	if err != nil {
		return nil, err
	}
	return plans, nil
}
