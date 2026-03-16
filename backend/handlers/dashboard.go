package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type DashboardHandler struct {
	users         *repository.UserRepository
	policies      *repository.PolicyRepository
	activity      *repository.ActivityRepository
	reputationSvc *services.ReputationService
	logger        *zap.Logger
}

func NewDashboardHandler(
	users *repository.UserRepository,
	policies *repository.PolicyRepository,
	activity *repository.ActivityRepository,
	reputationSvc *services.ReputationService,
	logger *zap.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		users:         users,
		policies:      policies,
		activity:      activity,
		reputationSvc: reputationSvc,
		logger:        logger,
	}
}

func (h *DashboardHandler) BookmarkToggle(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	policy, err := h.policies.FindByID(c, policyID)
	if err != nil {
		h.logger.Error("failed to find policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find policy"})
		return
	}
	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	bookmarked, err := h.users.ToggleBookmark(c, userID, policyID)
	if err != nil {
		h.logger.Error("failed to toggle bookmark", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle bookmark"})
		return
	}

	// Update policy bookmark count atomically
	go func() {
		if bookmarked {
			_ = h.policies.IncrementEngagement(c.Request.Context(), policyID, "engagement.bookmarkCount")
		} else {
			_ = h.policies.DecrementEngagement(c.Request.Context(), policyID, "engagement.bookmarkCount")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"bookmarked": bookmarked})
}

func (h *DashboardHandler) BookmarkList(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	policies, total, err := h.users.FindBookmarkedPolicies(c, userID, page, limit)
	if err != nil {
		h.logger.Error("failed to list bookmarks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookmarks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *DashboardHandler) Dashboard(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		h.logger.Error("failed to find user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	// Reputation + recent events
	recentEvents, _, err := h.reputationSvc.GetHistory(c, userID, 1)
	if err != nil {
		h.logger.Error("failed to get reputation history", zap.Error(err))
		recentEvents = nil
	}
	if len(recentEvents) > 5 {
		recentEvents = recentEvents[:5]
	}

	reputation := gin.H{
		"score":        user.Reputation.Score,
		"tier":         user.Reputation.Tier,
		"recentEvents": recentEvents,
	}

	// Stats
	stats, err := h.activity.GetUserStats(c, userID, len(user.Bookmarks))
	if err != nil {
		h.logger.Error("failed to get user stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	// Recent activity
	recentActivity, err := h.activity.GetRecentActivity(c, userID, 20)
	if err != nil {
		h.logger.Error("failed to get recent activity", zap.Error(err))
		recentActivity = []repository.ActivityItem{}
	}

	// Top 5 bookmarked policies
	bookmarkedPolicies, _, err := h.users.FindBookmarkedPolicies(c, userID, 1, 5)
	if err != nil {
		h.logger.Error("failed to get bookmarked policies", zap.Error(err))
		bookmarkedPolicies = nil
	}

	// Build user response without passwordHash (json:"-" already handles this)
	c.JSON(http.StatusOK, gin.H{
		"user":               user,
		"reputation":         reputation,
		"stats":              stats,
		"recentActivity":     recentActivity,
		"bookmarkedPolicies": bookmarkedPolicies,
	})
}
