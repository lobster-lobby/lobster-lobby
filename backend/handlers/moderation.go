package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// DebateModerationStore is the subset of DebateRepository used by ModerationHandler.
type DebateModerationStore interface {
	GetFlaggedArguments(ctx context.Context) ([]models.FlaggedArgumentDetail, error)
	GetArgumentByID(ctx context.Context, id bson.ObjectID) (*models.Argument, error)
	UnflagArgument(ctx context.Context, argumentID bson.ObjectID) error
	DeleteArgument(ctx context.Context, argumentID bson.ObjectID) error
	BanUser(ctx context.Context, userID bson.ObjectID) error
}

// UserModerationStore is the subset of UserRepository used by ModerationHandler.
type UserModerationStore interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*models.User, error)
}

// Compile-time assertions: real repos satisfy the interfaces.
var _ DebateModerationStore = (*repository.DebateRepository)(nil)
var _ UserModerationStore = (*repository.UserRepository)(nil)

type ModerationHandler struct {
	debates DebateModerationStore
	users   UserModerationStore
	logger  *zap.Logger
}

func NewModerationHandler(debates *repository.DebateRepository, users *repository.UserRepository, logger *zap.Logger) *ModerationHandler {
	return &ModerationHandler{debates: debates, users: users, logger: logger}
}

func (h *ModerationHandler) GetQueue(c *gin.Context) {
	flagged, err := h.debates.GetFlaggedArguments(c)
	if err != nil {
		h.logger.Error("failed to get moderation queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get moderation queue"})
		return
	}

	if flagged == nil {
		flagged = []models.FlaggedArgumentDetail{}
	}

	c.JSON(http.StatusOK, gin.H{"queue": flagged})
}

func (h *ModerationHandler) TakeAction(c *gin.Context) {
	argIDStr := c.Param("id")
	argID, err := bson.ObjectIDFromHex(argIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument id"})
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Action != "approve" && req.Action != "remove" && req.Action != "ban" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be one of: approve, remove, ban"})
		return
	}

	argument, err := h.debates.GetArgumentByID(c, argID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "argument not found"})
		return
	}

	switch req.Action {
	case "approve":
		if err := h.debates.UnflagArgument(c, argID); err != nil {
			h.logger.Error("failed to approve argument", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve argument"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "argument approved"})

	case "remove":
		if err := h.debates.DeleteArgument(c, argID); err != nil {
			h.logger.Error("failed to remove argument", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove argument"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "argument removed"})

	case "ban":
		if err := h.debates.DeleteArgument(c, argID); err != nil {
			h.logger.Error("failed to remove argument", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove argument"})
			return
		}
		if err := h.debates.BanUser(c, argument.AuthorID); err != nil {
			h.logger.Error("failed to ban user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ban user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "argument removed and user banned"})
	}
}
