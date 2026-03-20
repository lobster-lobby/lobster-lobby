package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type DraftsHandler struct {
	drafts *repository.DraftRepository
	users  *repository.UserRepository
	logger *zap.Logger
}

func NewDraftsHandler(drafts *repository.DraftRepository, users *repository.UserRepository, logger *zap.Logger) *DraftsHandler {
	return &DraftsHandler{drafts: drafts, users: users, logger: logger}
}

func (h *DraftsHandler) ListByPolicy(c *gin.Context) {
	policyID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	sort := c.DefaultQuery("sort", "top")
	drafts, err := h.drafts.ListByPolicy(c, policyID, sort)
	if err != nil {
		h.logger.Error("failed to list drafts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list drafts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drafts": drafts})
}

func (h *DraftsHandler) Create(c *gin.Context) {
	policyID, err := bson.ObjectIDFromHex(c.Param("id"))
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
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if len(req.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title exceeds 200 characters"})
		return
	}
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	validCategories := map[string]bool{"amendment": true, "talking-point": true, "position-statement": true, "full-text": true}
	if req.Category != "" && !validCategories[req.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must be one of: amendment, talking-point, position-statement, full-text"})
		return
	}

	authorName := ""
	if user, err := h.users.FindByID(c, userID); err == nil && user != nil {
		authorName = user.Username
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	draft := &models.Draft{
		PolicyID:   policyID,
		AuthorID:   userID,
		AuthorName: authorName,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Status:     status,
	}

	created, err := h.drafts.Create(c, draft)
	if err != nil {
		h.logger.Error("failed to create draft", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create draft"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"draft": created})
}

func (h *DraftsHandler) Update(c *gin.Context) {
	draftID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	draft, err := h.drafts.GetByID(c, draftID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}
	if draft.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the draft author can edit this draft"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" && len(req.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title exceeds 200 characters"})
		return
	}

	updated, err := h.drafts.Update(c, draftID, req.Title, req.Content, req.Category, req.Status)
	if err != nil {
		h.logger.Error("failed to update draft", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"draft": updated})
}

func (h *DraftsHandler) Delete(c *gin.Context) {
	draftID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	draft, err := h.drafts.GetByID(c, draftID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}

	isAdmin := false
	if user, err := h.users.FindByID(c, userID); err == nil && user != nil {
		isAdmin = user.Role == "admin"
	}

	if draft.AuthorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the draft author or an admin can delete this draft"})
		return
	}

	if err := h.drafts.Archive(c, draftID); err != nil {
		h.logger.Error("failed to archive draft", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "draft deleted"})
}

func (h *DraftsHandler) Endorse(c *gin.Context) {
	draftID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if _, err := h.drafts.GetByID(c, draftID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}

	endorsed, err := h.drafts.ToggleEndorsement(c, draftID, userID)
	if err != nil {
		h.logger.Error("failed to toggle endorsement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to toggle endorsement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"endorsed": endorsed})
}

func (h *DraftsHandler) ListComments(c *gin.Context) {
	draftID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	comments, err := h.drafts.ListComments(c, draftID)
	if err != nil {
		h.logger.Error("failed to list comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (h *DraftsHandler) AddComment(c *gin.Context) {
	draftID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if _, err := h.drafts.GetByID(c, draftID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
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

	authorName := ""
	if user, err := h.users.FindByID(c, userID); err == nil && user != nil {
		authorName = user.Username
	}

	comment := &models.DraftComment{
		DraftID:    draftID,
		AuthorID:   userID,
		AuthorName: authorName,
		Content:    req.Content,
	}

	created, err := h.drafts.AddComment(c, comment)
	if err != nil {
		h.logger.Error("failed to add comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": created})
}
