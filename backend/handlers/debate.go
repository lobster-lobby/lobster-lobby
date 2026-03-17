package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type DebateHandler struct {
	comments      *repository.CommentRepository
	policies      *repository.PolicyRepository
	logger        *zap.Logger
	reputationSvc *services.ReputationService
}

func NewDebateHandler(comments *repository.CommentRepository, policies *repository.PolicyRepository, logger *zap.Logger, reputationSvc *services.ReputationService) *DebateHandler {
	return &DebateHandler{comments: comments, policies: policies, logger: logger, reputationSvc: reputationSvc}
}

func (h *DebateHandler) CreateComment(c *gin.Context) {
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

	userType := "human"
	if ut, exists := c.Get(middleware.ContextUserType); exists {
		userType = ut.(string)
	}

	var req struct {
		Content  string `json:"content"`
		Position string `json:"position"`
		ParentID string `json:"parentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Position != "support" && req.Position != "oppose" && req.Position != "neutral" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "position must be one of: support, oppose, neutral"})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	comment := &models.Comment{
		PolicyID:   policyID,
		AuthorID:   userID,
		AuthorType: userType,
		Position:   req.Position,
		Content:    req.Content,
	}

	if req.ParentID != "" {
		parentID, err := bson.ObjectIDFromHex(req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent id"})
			return
		}
		comment.ParentID = &parentID
	}

	created, err := h.comments.Create(c, comment)
	if err != nil {
		h.logger.Error("failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	// Auto-set user's stance
	go func() {
		if err := h.comments.SetStance(c.Request.Context(), userID, policyID, req.Position); err != nil {
			h.logger.Warn("failed to set stance", zap.Error(err))
		}
	}()

	// Award reputation points
	go func() {
		if err := h.reputationSvc.AwardPoints(c.Request.Context(), userID, models.ActionCommentPosted, created.ID.Hex(), "comment"); err != nil {
			h.logger.Error("failed to award reputation points", zap.Error(err))
		}
	}()

	// Increment debate engagement count
	go func() {
		if err := h.policies.IncrementEngagement(c.Request.Context(), policyID, "engagement.debateCount"); err != nil {
			h.logger.Warn("failed to increment debate count", zap.Error(err))
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"comment": created})
}

func (h *DebateHandler) ListComments(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	sort := c.DefaultQuery("sort", "best")
	position := c.DefaultQuery("position", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	comments, total, positions, err := h.comments.FindByPolicy(c, policyID, sort, position, page, perPage)
	if err != nil {
		h.logger.Error("failed to list comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	// Enrich with user reactions if authenticated
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if userID, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			for i := range comments {
				comments[i].UserReaction = h.comments.GetUserReaction(c, userID, comments[i].ID)
			}
		}
	}

	if comments == nil {
		comments = []models.CommentResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  comments,
		"total":     total,
		"page":      page,
		"positions": positions,
	})
}

func (h *DebateHandler) GetReplies(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := bson.ObjectIDFromHex(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	replies, err := h.comments.FindReplies(c, commentID)
	if err != nil {
		h.logger.Error("failed to get replies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get replies"})
		return
	}

	// Enrich with user reactions if authenticated
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if userID, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			for i := range replies {
				replies[i].UserReaction = h.comments.GetUserReaction(c, userID, replies[i].ID)
			}
		}
	}

	if replies == nil {
		replies = []models.CommentResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"replies": replies})
}

func (h *DebateHandler) EditComment(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := bson.ObjectIDFromHex(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	if err := h.comments.Update(c, commentID, userID, req.Content); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own comments"})
			return
		}
		h.logger.Error("failed to edit comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to edit comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment updated"})
}

func (h *DebateHandler) ReactToComment(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := bson.ObjectIDFromHex(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Value int `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Value != 1 && req.Value != -1 && req.Value != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 1, -1, or 0"})
		return
	}

	if err := h.comments.React(c, userID, commentID, req.Value); err != nil {
		h.logger.Error("failed to react to comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to react to comment"})
		return
	}

	// Award reputation to the comment author
	if req.Value != 0 {
		go func() {
			comment, err := h.comments.FindByID(c.Request.Context(), commentID)
			if err != nil || comment == nil {
				return
			}
			action := models.ActionUpvoteReceived
			if req.Value == -1 {
				action = models.ActionDownvoteReceived
			}
			if err := h.reputationSvc.AwardPoints(c.Request.Context(), comment.AuthorID, action, commentID.Hex(), "comment"); err != nil {
				h.logger.Warn("failed to award reaction reputation", zap.Error(err))
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "reaction recorded"})
}

func (h *DebateHandler) SetStance(c *gin.Context) {
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
		Position string `json:"position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Position != "support" && req.Position != "oppose" && req.Position != "neutral" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "position must be one of: support, oppose, neutral"})
		return
	}

	if err := h.comments.SetStance(c, userID, policyID, req.Position); err != nil {
		h.logger.Error("failed to set stance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set stance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stance updated", "position": req.Position})
}

func (h *DebateHandler) GetStance(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	userIDStr, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusOK, gin.H{"stance": nil})
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"stance": nil})
		return
	}

	stance, err := h.comments.GetStance(c, userID, policyID)
	if err != nil {
		h.logger.Error("failed to get stance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stance": stance})
}
