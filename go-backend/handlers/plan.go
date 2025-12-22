package handlers

import (
	"mcpdocs/models"
	"mcpdocs/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	planRepo *repository.PlanRepository
}

func NewPlanHandler(planRepo *repository.PlanRepository) *PlanHandler {
	return &PlanHandler{planRepo: planRepo}
}

func (h *PlanHandler) ListPlans(c *gin.Context) {
	plans, err := h.planRepo.ListPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *PlanHandler) GetPlan(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	plan, err := h.planRepo.GetPlanByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.planRepo.CreatePlan(&plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Plan created successfully", "plan": plan})
}
