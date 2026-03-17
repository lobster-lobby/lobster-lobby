package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type SummaryHandler struct {
	summary       *repository.SummaryPointRepository
	comments      *repository.CommentRepository
	policies      *repository.PolicyRepository
	users         *repository.UserRepository
	logger        *zap.Logger
	reputationSvc *services.ReputationService
}

func NewSummaryHandler(
	summary *repository.SummaryPointRepository,
	comments *repository.CommentRepository,
	policies *repository.PolicyRepository,
	users *repository.UserRepository,
	logger *zap.Logger,
	reputationSvc *services.ReputationService,
) *SummaryHandler {
	return &SummaryHandler{
		summary:       summary,
		comments:      comments,
		policies:      policies,
		users:         users,
		logger:        logger,
		reputationSvc: reputationSvc,
	}
}

func (h *SummaryHandler) ListSummary(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	includeHidden := c.DefaultQuery("all", "") == "true"

	points, err := h.summary.ListByPolicy(c, policyID, includeHidden)
	if err != nil {
		h.logger.Error("failed to list summary points", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list summary points"})
		return
	}

	// Get current user ID if authenticated
	var userID *bson.ObjectID
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if uid, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	// Group by position
	grouped := map[string][]models.SummaryPointResponse{
		"support":   {},
		"oppose":    {},
		"consensus": {},
	}

	for _, p := range points {
		resp := h.summary.EnrichResponse(c, p, userID)
		grouped[p.Position] = append(grouped[p.Position], resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"support":   grouped["support"],
		"oppose":    grouped["oppose"],
		"consensus": grouped["consensus"],
	})
}

func (h *SummaryHandler) CreatePoint(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Content         string `json:"content"`
		Position        string `json:"position"`
		SourceCommentID string `json:"sourceCommentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	point := &models.SummaryPoint{
		PolicyID: policyID,
		AuthorID: userID,
		Content:  req.Content,
		Position: req.Position,
	}

	if req.SourceCommentID != "" {
		srcID, err := bson.ObjectIDFromHex(req.SourceCommentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source comment id"})
			return
		}
		// Verify the source comment exists and belongs to this policy
		comment, err := h.comments.FindByID(c, srcID)
		if err != nil || comment == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source comment not found"})
			return
		}
		if comment.PolicyID != policyID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source comment belongs to a different policy"})
			return
		}
		point.SourceCommentID = &srcID
	}

	if err := point.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.summary.Create(c, point)
	if err != nil {
		h.logger.Error("failed to create summary point", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create summary point"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"point": created})
}

func (h *SummaryHandler) EndorsePoint(c *gin.Context) {
	pointIDStr := c.Param("pointId")
	pointID, err := bson.ObjectIDFromHex(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid point id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Position string `json:"position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Position != "support" && req.Position != "oppose" && req.Position != "consensus" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "position must be one of: support, oppose, consensus"})
		return
	}

	// Verify point exists
	point, err := h.summary.FindByID(c, pointID)
	if err != nil || point == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "summary point not found"})
		return
	}

	// Cannot endorse own point
	if point.AuthorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot endorse your own summary point"})
		return
	}

	// Fetch user for verified status and rep tier
	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	endorsement := models.Endorsement{
		UserID:   userID,
		Position: req.Position,
		Verified: user.Verified,
		RepTier:  user.Reputation.Tier,
	}

	if err := h.summary.AddEndorsement(c, pointID, endorsement); err != nil {
		h.logger.Error("failed to endorse point", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to endorse point"})
		return
	}

	// Award reputation to point author
	go func() {
		if err := h.reputationSvc.AwardPoints(context.Background(), point.AuthorID, models.ActionEndorsementReceived, pointID.Hex(), "summaryPoint"); err != nil {
			h.logger.Warn("failed to award endorsement reputation", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "endorsement recorded"})
}

func (h *SummaryHandler) RemoveEndorsement(c *gin.Context) {
	pointIDStr := c.Param("pointId")
	pointID, err := bson.ObjectIDFromHex(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid point id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.summary.RemoveEndorsement(c, pointID, userID); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "endorsement not found"})
			return
		}
		h.logger.Error("failed to remove endorsement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove endorsement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "endorsement removed"})
}
