package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type PollsHandler struct {
	polls  *repository.PollRepository
	users  *repository.UserRepository
	logger *zap.Logger
}

func NewPollsHandler(polls *repository.PollRepository, users *repository.UserRepository, logger *zap.Logger) *PollsHandler {
	return &PollsHandler{polls: polls, users: users, logger: logger}
}

func (h *PollsHandler) ListByPolicy(c *gin.Context) {
	policyID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	polls, err := h.polls.ListByPolicy(c, policyID)
	if err != nil {
		h.logger.Error("failed to list polls", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list polls"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (h *PollsHandler) Create(c *gin.Context) {
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
		Question    string     `json:"question"`
		Options     []string   `json:"options"`
		MultiSelect bool       `json:"multiSelect"`
		EndsAt      *time.Time `json:"endsAt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	if len(req.Question) > 280 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question exceeds 280 characters"})
		return
	}
	if len(req.Options) < 2 || len(req.Options) > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "polls require 2-6 options"})
		return
	}

	options := make([]models.PollOption, len(req.Options))
	for i, text := range req.Options {
		if text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "option text cannot be empty"})
			return
		}
		options[i] = models.PollOption{Text: text}
	}

	authorName := ""
	if user, err := h.users.FindByID(c, userID); err == nil && user != nil {
		authorName = user.Username
	}

	poll := &models.Poll{
		PolicyID:    policyID,
		AuthorID:    userID,
		AuthorName:  authorName,
		Question:    req.Question,
		Options:     options,
		MultiSelect: req.MultiSelect,
		EndsAt:      req.EndsAt,
	}

	created, err := h.polls.Create(c, poll)
	if err != nil {
		h.logger.Error("failed to create poll", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create poll"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"poll": created})
}

func (h *PollsHandler) Vote(c *gin.Context) {
	pollID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		OptionIDs []string `json:"optionIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.OptionIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one option is required"})
		return
	}

	poll, err := h.polls.GetByID(c, pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}
	if poll.Status == "closed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "poll is closed"})
		return
	}
	if !poll.MultiSelect && len(req.OptionIDs) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this poll only allows a single selection"})
		return
	}

	optionIDs := make([]bson.ObjectID, 0, len(req.OptionIDs))
	for _, idStr := range req.OptionIDs {
		oid, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option id: " + idStr})
			return
		}
		found := false
		for _, opt := range poll.Options {
			if opt.ID == oid {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "option not found in poll"})
			return
		}
		optionIDs = append(optionIDs, oid)
	}

	updated, err := h.polls.Vote(c, pollID, userID, optionIDs)
	if err != nil {
		h.logger.Error("failed to vote on poll", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"poll": updated})
}

func (h *PollsHandler) Delete(c *gin.Context) {
	pollID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	poll, err := h.polls.GetByID(c, pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	// Check if user is admin
	isAdmin := false
	if user, err := h.users.FindByID(c, userID); err == nil && user != nil {
		isAdmin = user.Role == "admin"
	}

	if poll.AuthorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the poll author or an admin can delete this poll"})
		return
	}

	if err := h.polls.Delete(c, pollID); err != nil {
		h.logger.Error("failed to delete poll", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete poll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "poll deleted"})
}
