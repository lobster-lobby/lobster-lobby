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
)

type ResearchHandler struct {
	research *repository.ResearchRepository
	policies *repository.PolicyRepository
	logger   *zap.Logger
}

func NewResearchHandler(research *repository.ResearchRepository, policies *repository.PolicyRepository, logger *zap.Logger) *ResearchHandler {
	return &ResearchHandler{research: research, policies: policies, logger: logger}
}

func (h *ResearchHandler) Create(c *gin.Context) {
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
		Title   string          `json:"title"`
		Type    string          `json:"type"`
		Content string          `json:"content"`
		Sources []models.Source `json:"sources"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate title (5-300 chars)
	if len(req.Title) < 5 || len(req.Title) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 5 and 300 characters"})
		return
	}

	// Validate type
	validTypes := map[string]bool{
		"analysis":   true,
		"news":       true,
		"data":       true,
		"academic":   true,
		"government": true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be one of: analysis, news, data, academic, government"})
		return
	}

	// Validate content (min 50 chars)
	if len(req.Content) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be at least 50 characters"})
		return
	}

	// Validate sources (min 1, each needs url + title)
	if len(req.Sources) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one source is required"})
		return
	}
	for i, src := range req.Sources {
		if src.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source " + strconv.Itoa(i+1) + " is missing url"})
			return
		}
		if src.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source " + strconv.Itoa(i+1) + " is missing title"})
			return
		}
	}

	research := &models.Research{
		PolicyID:   policyID,
		AuthorID:   userID,
		AuthorType: userType,
		Title:      req.Title,
		Type:       req.Type,
		Content:    req.Content,
		Sources:    req.Sources,
	}

	created, err := h.research.Create(c, research)
	if err != nil {
		h.logger.Error("failed to create research", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create research"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"research": created})
}

func (h *ResearchHandler) List(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	sort := c.DefaultQuery("sort", "newest")
	researchType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}

	var userID *bson.ObjectID
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if uid, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	items, total, err := h.research.List(c, policyID, sort, researchType, page, limit, userID)
	if err != nil {
		h.logger.Error("failed to list research", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list research"})
		return
	}

	if items == nil {
		items = []*models.ResearchResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"research": items,
		"total":    total,
		"page":     page,
	})
}

func (h *ResearchHandler) GetByID(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	researchIDStr := c.Param("researchId")
	researchID, err := bson.ObjectIDFromHex(researchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid research id"})
		return
	}

	var userID *bson.ObjectID
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if uid, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	research, err := h.research.GetByID(c, policyID, researchID, userID)
	if err != nil {
		h.logger.Error("failed to get research", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get research"})
		return
	}

	if research == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"research": research})
}

func (h *ResearchHandler) Update(c *gin.Context) {
	researchIDStr := c.Param("researchId")
	researchID, err := bson.ObjectIDFromHex(researchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid research id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Title   string          `json:"title"`
		Content string          `json:"content"`
		Sources []models.Source `json:"sources"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate title (5-300 chars)
	if len(req.Title) < 5 || len(req.Title) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 5 and 300 characters"})
		return
	}

	// Validate content (min 50 chars)
	if len(req.Content) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be at least 50 characters"})
		return
	}

	// Validate sources (min 1, each needs url + title)
	if len(req.Sources) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one source is required"})
		return
	}
	for i, src := range req.Sources {
		if src.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source " + strconv.Itoa(i+1) + " is missing url"})
			return
		}
		if src.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source " + strconv.Itoa(i+1) + " is missing title"})
			return
		}
	}

	updated, err := h.research.Update(c, researchID, userID, req.Title, req.Content, req.Sources)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own research"})
			return
		}
		h.logger.Error("failed to update research", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update research"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"research": updated})
}

func (h *ResearchHandler) Vote(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	researchIDStr := c.Param("researchId")
	researchID, err := bson.ObjectIDFromHex(researchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid research id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// Look up research to prevent self-voting
	research, err := h.research.GetByID(c, policyID, researchID, nil)
	if err != nil || research == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
		return
	}
	if userID == research.AuthorID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot vote on your own research"})
		return
	}

	var req struct {
		Type string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var value int
	switch req.Type {
	case "up":
		value = 1
	case "down":
		value = -1
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be \"up\" or \"down\""})
		return
	}

	newValue, err := h.research.ToggleVote(c, userID, researchID, value)
	if err != nil {
		h.logger.Error("failed to vote on research", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"vote": newValue})
}
