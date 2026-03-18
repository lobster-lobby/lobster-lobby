package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/repository"
)

type CampaignActivityHandler struct {
	activities *repository.CampaignActivityRepository
	campaigns  *repository.CampaignRepository
	logger     *zap.Logger
}

func NewCampaignActivityHandler(
	activities *repository.CampaignActivityRepository,
	campaigns *repository.CampaignRepository,
	logger *zap.Logger,
) *CampaignActivityHandler {
	return &CampaignActivityHandler{
		activities: activities,
		campaigns:  campaigns,
		logger:     logger,
	}
}

// ListActivity handles GET /api/campaigns/:id/activity
func (h *CampaignActivityHandler) ListActivity(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	opts := repository.CampaignActivityListOpts{
		CampaignID: campaign.ID.Hex(),
		Page:       page,
		Limit:      limit,
	}

	activities, total, err := h.activities.ListByCampaign(c, opts)
	if err != nil {
		h.logger.Error("failed to list activities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// GetReachMetrics handles GET /api/campaigns/:id/reach
func (h *CampaignActivityHandler) GetReachMetrics(c *gin.Context) {
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

	// Count unique participants from activity records
	uniqueParticipants, err := h.activities.CountUniqueUsers(c, campaign.ID.Hex())
	if err != nil {
		h.logger.Warn("failed to count unique users", zap.Error(err))
	}

	// Use the larger of the two participant counts
	participants := campaign.Metrics.UniqueParticipants
	if uniqueParticipants > participants {
		participants = uniqueParticipants
	}

	// totalSupporters = uniqueParticipants (users who have interacted)
	c.JSON(http.StatusOK, gin.H{
		"totalSupporters":    participants,
		"totalShares":        campaign.Metrics.TotalShares,
		"totalDownloads":     campaign.Metrics.TotalDownloads,
		"uniqueParticipants": participants,
		"trendingScore":      campaign.TrendingScore,
	})
}
