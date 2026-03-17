package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/repository"
)

type CampaignEventHandler struct {
	events    *repository.CampaignEventRepository
	campaigns *repository.CampaignRepository
	logger    *zap.Logger
}

func NewCampaignEventHandler(
	events *repository.CampaignEventRepository,
	campaigns *repository.CampaignRepository,
	logger *zap.Logger,
) *CampaignEventHandler {
	return &CampaignEventHandler{
		events:    events,
		campaigns: campaigns,
		logger:    logger,
	}
}

// List handles GET /api/campaigns/:slug/events
func (h *CampaignEventHandler) List(c *gin.Context) {
	idOrSlug := c.Param("id")

	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	events, err := h.events.ListByCampaign(c, campaign.ID.Hex())
	if err != nil {
		h.logger.Error("failed to list events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// GetActivity handles GET /api/campaigns/:slug/metrics/activity
func (h *CampaignEventHandler) GetActivity(c *gin.Context) {
	idOrSlug := c.Param("id")

	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	// Get daily activity for last 30 days
	dailyActivity, err := h.events.GetActivityByDay(c, campaign.ID.Hex(), 30)
	if err != nil {
		h.logger.Error("failed to get activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activity"})
		return
	}

	// Get recent events for feed
	recentEvents, err := h.events.GetRecentEvents(c, campaign.ID.Hex(), 10)
	if err != nil {
		h.logger.Error("failed to get recent events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recent events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dailyActivity": dailyActivity,
		"recentEvents":  recentEvents,
		"metrics":       campaign.Metrics,
	})
}
