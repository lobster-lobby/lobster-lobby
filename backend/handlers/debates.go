package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type DebatesHandler struct {
	debates       *repository.DebateRepository
	logger        *zap.Logger
	reputationSvc *services.ReputationService
}

func NewDebatesHandler(debates *repository.DebateRepository, logger *zap.Logger, reputationSvc *services.ReputationService) *DebatesHandler {
	return &DebatesHandler{debates: debates, logger: logger, reputationSvc: reputationSvc}
}

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRegexp.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

func (h *DebatesHandler) CreateDebate(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "title exceeds maximum length of 200 characters"})
		return
	}
	if len(req.Description) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description exceeds maximum length of 5000 characters"})
		return
	}

	slug := slugify(req.Title)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must contain alphanumeric characters"})
		return
	}

	debate := &models.Debate{
		Slug:        slug,
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   userID,
	}

	created, err := h.debates.CreateDebate(c, debate)
	if err != nil {
		h.logger.Error("failed to create debate", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create debate"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"debate": created})
}

func (h *DebatesHandler) ListDebates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	debates, total, err := h.debates.ListDebates(c, page, perPage)
	if err != nil {
		h.logger.Error("failed to list debates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list debates"})
		return
	}

	if debates == nil {
		debates = []models.DebateResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"debates": debates,
		"total":   total,
		"page":    page,
	})
}

func (h *DebatesHandler) GetDebate(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	debate, err := h.debates.GetDebateBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debate not found"})
		return
	}

	sort := c.DefaultQuery("sort", "top")
	if sort != "newest" && sort != "top" && sort != "controversial" {
		sort = "top"
	}

	var arguments []models.ArgumentResponse
	if sort == "controversial" {
		arguments, err = h.debates.ListArgumentsControversial(c, debate.ID)
	} else {
		arguments, err = h.debates.ListArguments(c, debate.ID, sort)
	}
	if err != nil {
		h.logger.Error("failed to list arguments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list arguments"})
		return
	}

	// Enrich with user votes if authenticated
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if userID, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			for i := range arguments {
				arguments[i].UserVote = h.debates.GetUserVote(c, userID, arguments[i].ID)
			}
		}
	}

	if arguments == nil {
		arguments = []models.ArgumentResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"debate":    debate,
		"arguments": arguments,
	})
}

func (h *DebatesHandler) CreateArgument(c *gin.Context) {
	slug := c.Param("slug")
	debate, err := h.debates.GetDebateBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debate not found"})
		return
	}

	if debate.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "debate is closed"})
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
		Content string `json:"content"`
		Side    string `json:"side"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Side != "pro" && req.Side != "con" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "side must be one of: pro, con"})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	if len(req.Content) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content exceeds maximum length of 10000 characters"})
		return
	}

	arg := &models.Argument{
		DebateID:   debate.ID,
		AuthorID:   userID,
		AuthorType: userType,
		Side:       req.Side,
		Content:    req.Content,
	}

	created, err := h.debates.CreateArgument(c, arg)
	if err != nil {
		h.logger.Error("failed to create argument", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create argument"})
		return
	}

	// Award reputation points
	go func() {
		if err := h.reputationSvc.AwardPoints(context.Background(), userID, models.ActionCommentPosted, created.ID.Hex(), "argument"); err != nil {
			h.logger.Warn("failed to award reputation points", zap.Error(err))
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"argument": created})
}

func (h *DebatesHandler) VoteOnArgument(c *gin.Context) {
	slug := c.Param("slug")
	debate, err := h.debates.GetDebateBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debate not found"})
		return
	}

	argIDStr := c.Param("id")
	argID, err := bson.ObjectIDFromHex(argIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument id"})
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

	if req.Value != 1 && req.Value != -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 1 or -1"})
		return
	}

	newValue, err := h.debates.ToggleVote(c, userID, argID, debate.ID, req.Value)
	if err != nil {
		h.logger.Error("failed to vote on argument", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record vote"})
		return
	}

	// Award reputation to the argument author
	go func() {
		args, listErr := h.debates.ListArguments(context.Background(), debate.ID, "top")
		if listErr != nil {
			return
		}
		for _, a := range args {
			if a.ID == argID {
				action := models.ActionUpvoteReceived
				if req.Value == -1 {
					action = models.ActionDownvoteReceived
				}
				if err := h.reputationSvc.AwardPoints(context.Background(), a.AuthorID, action, argID.Hex(), "argument"); err != nil {
					h.logger.Warn("failed to award vote reputation", zap.Error(err))
				}
				break
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"vote": newValue})
}
