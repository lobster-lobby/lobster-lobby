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

type ModerationHandler struct {
	debates *repository.DebateRepository
	users   *repository.UserRepository
	logger  *zap.Logger
}

func NewModerationHandler(debates *repository.DebateRepository, users *repository.UserRepository, logger *zap.Logger) *ModerationHandler {
	return &ModerationHandler{debates: debates, users: users, logger: logger}
}

func (h *ModerationHandler) requireAdmin(c *gin.Context) (bson.ObjectID, bool) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return bson.ObjectID{}, false
	}

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return bson.ObjectID{}, false
	}

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return bson.ObjectID{}, false
	}

	return userID, true
}

func (h *ModerationHandler) GetQueue(c *gin.Context) {
	if _, ok := h.requireAdmin(c); !ok {
		return
	}

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
	if _, ok := h.requireAdmin(c); !ok {
		return
	}

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
