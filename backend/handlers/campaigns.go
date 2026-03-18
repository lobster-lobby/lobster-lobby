package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type CampaignHandler struct {
	campaigns     *repository.CampaignRepository
	policies      *repository.PolicyRepository
	users         *repository.UserRepository
	events        *repository.CampaignEventRepository
	jwtSvc        *services.JWTService
	reputationSvc *services.ReputationService
	logger        *zap.Logger
}

func NewCampaignHandler(
	campaigns *repository.CampaignRepository,
	policies *repository.PolicyRepository,
	users *repository.UserRepository,
	events *repository.CampaignEventRepository,
	jwtSvc *services.JWTService,
	reputationSvc *services.ReputationService,
	logger *zap.Logger,
) *CampaignHandler {
	return &CampaignHandler{
		campaigns:     campaigns,
		policies:      policies,
		users:         users,
		events:        events,
		jwtSvc:        jwtSvc,
		reputationSvc: reputationSvc,
		logger:        logger,
	}
}

func (h *CampaignHandler) Create(c *gin.Context) {
	var req struct {
		Title       string              `json:"title"`
		PolicyID    string              `json:"policyId"`
		Objective   string              `json:"objective"`
		Target      string              `json:"target"`
		Description string              `json:"description"`
		Milestones  []models.Milestone  `json:"milestones"`
		Status      string              `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// Check reputation >= 50
	user, err := h.users.FindByID(c, userID)
	if err != nil {
		h.logger.Error("failed to find user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if user.Reputation.Score < 50 {
		c.JSON(http.StatusForbidden, gin.H{"error": "reputation must be at least 50 to create campaigns"})
		return
	}

	// Validate policyID exists and is active
	policyID, err := bson.ObjectIDFromHex(req.PolicyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	policy, err := h.policies.FindByID(c, policyID)
	if err != nil {
		h.logger.Error("failed to find policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify policy"})
		return
	}
	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}
	if policy.Status != models.PolicyStatusReadyForCampaign {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy must be ready_for_campaign to create a campaign"})
		return
	}

	campaign := &models.Campaign{
		Title:       req.Title,
		PolicyID:    policyID,
		CreatedBy:   userID,
		Objective:   req.Objective,
		Target:      req.Target,
		Description: req.Description,
		Status:      models.CampaignStatus(req.Status),
		Milestones:  req.Milestones,
	}

	if err := campaign.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.campaigns.Create(c, campaign); err != nil {
		h.logger.Error("failed to create campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create campaign"})
		return
	}

	// Create "created" event
	if h.events != nil {
		go func() {
			event := &models.CampaignEvent{
				CampaignID:  campaign.ID,
				Type:        models.CampaignEventCreated,
				Title:       "Campaign Created",
				Description: "Campaign '" + campaign.Title + "' was launched",
				Metadata: map[string]any{
					"campaignTitle": campaign.Title,
				},
			}
			if err := h.events.Create(context.Background(), event); err != nil {
				h.logger.Warn("failed to create campaign_created event", zap.Error(err))
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{"campaign": campaign})
}

func (h *CampaignHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	opts := repository.CampaignListOpts{
		Page:     page,
		PerPage:  perPage,
		Sort:     c.DefaultQuery("sort", "trending"),
		Status:   c.Query("status"),
		PolicyID: c.Query("policyId"),
		Search:   c.Query("q"),
	}

	campaigns, total, err := h.campaigns.List(c, opts)
	if err != nil {
		h.logger.Error("failed to list campaigns", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"total":     total,
		"page":      page,
		"perPage":   perPage,
	})
}

func (h *CampaignHandler) Get(c *gin.Context) {
	idOrSlug := c.Param("id")

	var campaign *models.Campaign
	var err error

	// Try ObjectID first
	if oid, parseErr := bson.ObjectIDFromHex(idOrSlug); parseErr == nil {
		campaign, err = h.campaigns.GetByID(c, oid.Hex())
	} else {
		campaign, err = h.campaigns.FindBySlug(c, idOrSlug)
	}

	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}

	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"campaign": campaign})
}

func (h *CampaignHandler) Update(c *gin.Context) {
	idStr := c.Param("id")

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	campaign, err := h.campaigns.GetByID(c, idStr)
	if err != nil {
		h.logger.Error("failed to find campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	// Check authorization: must be creator or moderator
	isCreator := campaign.CreatedBy == userID

	isModerator := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isModerator = user.Role == "moderator" || user.Role == "admin"
	}

	if !isCreator && !isModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own campaigns"})
		return
	}

	var req struct {
		Title             *string             `json:"title"`
		Objective         *string             `json:"objective"`
		Target            *string             `json:"target"`
		Description       *string             `json:"description"`
		Milestones        []models.Milestone  `json:"milestones"`
		CompletionSummary *string             `json:"completionSummary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := bson.M{}

	if req.Title != nil {
		if len(*req.Title) < 5 || len(*req.Title) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 5 and 200 characters"})
			return
		}
		updates["title"] = *req.Title
	}

	if req.Objective != nil {
		if len(*req.Objective) < 10 || len(*req.Objective) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "objective must be between 10 and 500 characters"})
			return
		}
		updates["objective"] = *req.Objective
	}

	if req.Target != nil {
		if len(*req.Target) < 5 || len(*req.Target) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target must be between 5 and 500 characters"})
			return
		}
		updates["target"] = *req.Target
	}

	if req.Description != nil {
		if len(*req.Description) < 20 || len(*req.Description) > 5000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "description must be between 20 and 5000 characters"})
			return
		}
		updates["description"] = *req.Description
	}

	if req.Milestones != nil {
		updates["milestones"] = req.Milestones
	}

	if req.CompletionSummary != nil {
		updates["completionSummary"] = *req.CompletionSummary
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	updatedCampaign, err := h.campaigns.Update(c, idStr, updates)
	if err != nil {
		h.logger.Error("failed to update campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"campaign": updatedCampaign})
}

func (h *CampaignHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	campaign, err := h.campaigns.GetByID(c, idStr)
	if err != nil {
		h.logger.Error("failed to find campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	// Check authorization: must be creator or moderator
	isCreator := campaign.CreatedBy == userID

	isModerator := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isModerator = user.Role == "moderator" || user.Role == "admin"
	}

	if !isCreator && !isModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own campaigns"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := models.CampaignStatus(req.Status)
	if status != models.CampaignStatusActive && status != models.CampaignStatusPaused &&
		status != models.CampaignStatusCompleted && status != models.CampaignStatusArchived {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of: active, paused, completed, archived"})
		return
	}

	oldStatus := campaign.Status

	if err := h.campaigns.UpdateStatus(c, idStr, status); err != nil {
		h.logger.Error("failed to update campaign status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update campaign status"})
		return
	}

	// Create "status_change" event
	if h.events != nil && oldStatus != status {
		go func() {
			event := &models.CampaignEvent{
				CampaignID:  campaign.ID,
				Type:        models.CampaignEventStatusChange,
				Title:       "Status Changed",
				Description: "Campaign status changed from " + string(oldStatus) + " to " + string(status),
				Metadata: map[string]any{
					"oldStatus": string(oldStatus),
					"newStatus": string(status),
				},
			}
			if err := h.events.Create(context.Background(), event); err != nil {
				h.logger.Warn("failed to create status_change event", zap.Error(err))
			}
		}()
	}

	updatedCampaign, err := h.campaigns.GetByID(c, idStr)
	if err != nil {
		h.logger.Error("failed to fetch updated campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated campaign"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaign": updatedCampaign})
}

func (h *CampaignHandler) ListByPolicy(c *gin.Context) {
	policyID := c.Param("id")

	const maxCampaignsPerPolicy = 50

	campaigns, err := h.campaigns.FindByPolicy(c, policyID)
	if err != nil {
		h.logger.Error("failed to list campaigns by policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list campaigns"})
		return
	}

	if len(campaigns) > maxCampaignsPerPolicy {
		campaigns = campaigns[:maxCampaignsPerPolicy]
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"total":     len(campaigns),
	})
}
