package handlers

import (
	"mcpdocs/models"
	"mcpdocs/repository"
	"mcpdocs/schema"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionRepo *repository.SubscriptionRepository
	planRepo         *repository.PlanRepository
}

func NewSubscriptionHandler(subscriptionRepo *repository.SubscriptionRepository, planRepo *repository.PlanRepository) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var subscription schema.CreateSubscriptionInput

	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newSubscription := &models.Subcription{
		UserID:      subscription.UserID,
		PlanID:      subscription.PlanID,
		Status:      "active",
		PeriodStart: time.Now(),
		PeriodEnd:   time.Now().AddDate(0, 1, 0),
	}

	if err := h.subscriptionRepo.CreateSubscription(newSubscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Subscription created successfully", "subscription": newSubscription})
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	userIDParam := c.GetString("userID")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID parameter is required"})
		return
	}

	subscription, err := h.subscriptionRepo.GetActiveSubscriptionByUserID(userIDParam, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	subscriptionIDParam := c.Param("subscriptionID")
	if subscriptionIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscriptionID parameter is required"})
		return
	}

	err := h.subscriptionRepo.CancelSubscriptionByIDString(subscriptionIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription canceled successfully"})
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userIDParam := c.GetString("userID")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID parameter is required"})
		return
	}

	subscriptions, err := h.subscriptionRepo.ListSubscriptionsByUser(userIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	var subscription schema.UpdateSubscriptionInput

	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingSubscription, err := h.subscriptionRepo.GetActiveSubscriptionByUserID(subscription.UserID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if existingSubscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
		return
	}

	existingSubscription.PlanID = subscription.PlanID
	existingSubscription.Status = subscription.Status

	if err := h.subscriptionRepo.UpdateSubscription(existingSubscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully", "subscription": existingSubscription})
}

func (h *SubscriptionHandler) HandleRenewalWebhook(c *gin.Context) {
	var webhookData schema.SubscriptionRenewalWebhook

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionRepo.GetActiveSubscriptionByUserID(webhookData.UserID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
		return
	}

	subscription.PeriodStart = webhookData.NewPeriodStart
	subscription.PeriodEnd = webhookData.NewPeriodEnd

	if err := h.subscriptionRepo.UpdateSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription renewed successfully"})
}

func (h *SubscriptionHandler) HandleCancellationWebhook(c *gin.Context) {
	var webhookData schema.SubscriptionCancellationWebhook

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionRepo.GetSubscriptionByStripeID(webhookData.StripeSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	subscription.Status = "canceled"

	if err := h.subscriptionRepo.UpdateSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription canceled successfully"})
}

func (h *SubscriptionHandler) HandleFailedPaymentWebhook(c *gin.Context) {
	var webhookData schema.SubscriptionCancellationWebhook

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionRepo.GetSubscriptionByStripeID(webhookData.StripeSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	subscription.Status = "past_due"

	if err := h.subscriptionRepo.UpdateSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription marked as past due successfully"})
}

func (h *SubscriptionHandler) HandleSuccessfulPaymentWebhook(c *gin.Context) {
	var webhookData schema.SubscriptionRenewalWebhook

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionRepo.GetSubscriptionByStripeID(webhookData.StripeSubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	subscription.Status = "active"
	subscription.PeriodStart = webhookData.NewPeriodStart
	subscription.PeriodEnd = webhookData.NewPeriodEnd

	if err := h.subscriptionRepo.UpdateSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription marked as active successfully"})
}
