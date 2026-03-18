package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type CampaignCommentHandler struct {
	comments  *repository.CampaignCommentRepository
	campaigns *repository.CampaignRepository
	users     *repository.UserRepository
	events    *repository.CampaignEventRepository
	logger    *zap.Logger
}

func NewCampaignCommentHandler(
	comments *repository.CampaignCommentRepository,
	campaigns *repository.CampaignRepository,
	users *repository.UserRepository,
	events *repository.CampaignEventRepository,
	logger *zap.Logger,
) *CampaignCommentHandler {
	return &CampaignCommentHandler{
		comments:  comments,
		campaigns: campaigns,
		users:     users,
		events:    events,
		logger:    logger,
	}
}

// List handles GET /api/campaigns/:slug/comments
func (h *CampaignCommentHandler) List(c *gin.Context) {
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

	opts := repository.CampaignCommentListOpts{
		CampaignID: campaign.ID.Hex(),
		Sort:       c.DefaultQuery("sort", "newest"),
	}

	comments, err := h.comments.ListByCampaign(c, opts)
	if err != nil {
		h.logger.Error("failed to list comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	// If authenticated, get user votes
	userVotes := map[string]int{}
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		commentIDs := make([]string, len(comments))
		for i, comment := range comments {
			commentIDs[i] = comment.ID.Hex()
		}
		if len(commentIDs) > 0 {
			votes, err := h.comments.GetBatchUserVotes(c, commentIDs, userIDStr.(string))
			if err == nil {
				userVotes = votes
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  comments,
		"userVotes": userVotes,
	})
}

// Create handles POST /api/campaigns/:slug/comments
func (h *CampaignCommentHandler) Create(c *gin.Context) {
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

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Body     string  `json:"body"`
		ParentID *string `json:"parentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := &models.CampaignComment{
		CampaignID: campaign.ID,
		AuthorID:   userID,
		AuthorName: user.Username,
		Body:       req.Body,
	}

	// Handle parent ID for replies
	if req.ParentID != nil && *req.ParentID != "" {
		parentOID, err := bson.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent id"})
			return
		}
		// Verify parent exists and belongs to same campaign
		parent, err := h.comments.GetByID(c, *req.ParentID)
		if err != nil || parent == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "parent comment not found"})
			return
		}
		if parent.CampaignID != campaign.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent comment belongs to different campaign"})
			return
		}
		// Enforce single level of nesting - cannot reply to a reply
		if parent.ParentID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reply to a reply"})
			return
		}
		comment.ParentID = &parentOID
	}

	if err := comment.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.comments.Create(c, comment); err != nil {
		h.logger.Error("failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	// Update campaign comment count asynchronously
	campaignID := campaign.ID.Hex()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		count, err := h.comments.CountByCampaign(ctx, campaignID)
		if err == nil {
			h.campaigns.Update(ctx, campaignID, bson.M{
				"metrics.commentCount": int(count),
			})
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

// Update handles PUT /api/campaigns/:slug/comments/:id
func (h *CampaignCommentHandler) Update(c *gin.Context) {
	idOrSlug := c.Param("id")
	commentID := c.Param("commentId")

	// Resolve campaign to verify comment belongs to it
	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil || campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	comment, err := h.comments.GetByID(c, commentID)
	if err != nil {
		h.logger.Error("failed to get comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comment"})
		return
	}
	if comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	// Verify comment belongs to this campaign
	if comment.CampaignID != campaign.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	// Only owner can edit
	if comment.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own comments"})
		return
	}

	var req struct {
		Body *string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Body == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	if len(*req.Body) < 1 || len(*req.Body) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must be between 1 and 2000 characters"})
		return
	}

	updates := bson.M{"body": *req.Body}

	updatedComment, err := h.comments.Update(c, commentID, updates)
	if err != nil {
		h.logger.Error("failed to update comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": updatedComment})
}

// Delete handles DELETE /api/campaigns/:slug/comments/:id
func (h *CampaignCommentHandler) Delete(c *gin.Context) {
	idOrSlug := c.Param("id")
	commentID := c.Param("commentId")

	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil || campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	comment, err := h.comments.GetByID(c, commentID)
	if err != nil {
		h.logger.Error("failed to get comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comment"})
		return
	}
	if comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	// Verify comment belongs to this campaign
	if comment.CampaignID != campaign.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	// Check authorization: owner or moderator
	isOwner := comment.AuthorID == userID

	isModerator := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isModerator = user.Role == "moderator" || user.Role == "admin"
	}

	if !isOwner && !isModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own comments"})
		return
	}

	if err := h.comments.Delete(c, commentID); err != nil {
		h.logger.Error("failed to delete comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}

	// Update campaign comment count asynchronously
	// Use background context since request context may be cancelled after response
	campaignID := campaign.ID.Hex()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		count, err := h.comments.CountByCampaign(ctx, campaignID)
		if err == nil {
			h.campaigns.Update(ctx, campaignID, bson.M{
				"metrics.commentCount": int(count),
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

// TogglePin handles PUT /api/campaigns/:id/comments/:commentId/pin
func (h *CampaignCommentHandler) TogglePin(c *gin.Context) {
	idOrSlug := c.Param("id")
	commentID := c.Param("commentId")

	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil || campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	comment, err := h.comments.GetByID(c, commentID)
	if err != nil {
		h.logger.Error("failed to get comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comment"})
		return
	}
	if comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if comment.CampaignID != campaign.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	// Only campaign creator or admin can pin
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	isCreator := campaign.CreatedBy == userID
	isAdmin := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isAdmin = user.Role == "admin"
	}

	if !isCreator && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the campaign creator or an admin can pin comments"})
		return
	}

	// If pinning (not already pinned), enforce max 3
	if !comment.Pinned {
		pinnedCount, err := h.comments.CountPinnedByCampaign(c, campaign.ID.Hex())
		if err != nil {
			h.logger.Error("failed to count pinned comments", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle pin"})
			return
		}
		if pinnedCount >= 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum of 3 pinned comments per campaign"})
			return
		}
	}

	updatedComment, err := h.comments.Update(c, commentID, bson.M{"pinned": !comment.Pinned})
	if err != nil {
		h.logger.Error("failed to toggle pin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle pin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": updatedComment})
}

// Vote handles POST /api/campaigns/:slug/comments/:id/vote
func (h *CampaignCommentHandler) Vote(c *gin.Context) {
	idOrSlug := c.Param("id")
	commentID := c.Param("commentId")

	// Resolve campaign to verify comment belongs to it
	campaign, err := ResolveCampaign(c, h.campaigns, idOrSlug)
	if err != nil || campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	comment, err := h.comments.GetByID(c, commentID)
	if err != nil {
		h.logger.Error("failed to get comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comment"})
		return
	}
	if comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	// Verify comment belongs to this campaign
	if comment.CampaignID != campaign.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)

	var req struct {
		Value int `json:"value"` // 1, -1
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Value != 1 && req.Value != -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 1 or -1"})
		return
	}

	newVote, err := h.comments.ToggleVote(c, commentID, userIDStr.(string), req.Value)
	if err != nil {
		h.logger.Error("failed to toggle vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to vote"})
		return
	}

	// Get updated comment
	updatedComment, _ := h.comments.GetByID(c, commentID)

	c.JSON(http.StatusOK, gin.H{
		"comment":  updatedComment,
		"userVote": newVote,
	})
}
