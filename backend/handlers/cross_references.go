package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// CrossRefStore is the subset of CrossReferenceRepository used by CrossReferenceHandler.
type CrossRefStore interface {
	Create(ctx context.Context, ref *models.CrossReference) (*models.CrossReferenceResponse, error)
	GetForEntity(ctx context.Context, entityType string, entityID bson.ObjectID) ([]*models.CrossReferenceResponse, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	GetByID(ctx context.Context, id bson.ObjectID) (*models.CrossReference, error)
}

var _ CrossRefStore = (*repository.CrossReferenceRepository)(nil)

type CrossReferenceHandler struct {
	refs   CrossRefStore
	logger *zap.Logger
}

func NewCrossReferenceHandler(refs *repository.CrossReferenceRepository, logger *zap.Logger) *CrossReferenceHandler {
	return &CrossReferenceHandler{refs: refs, logger: logger}
}

var validRefTypes = map[string]bool{"research": true, "debate": true, "policy": true}

// Create handles POST /api/cross-references
func (h *CrossReferenceHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		SourceType string `json:"sourceType"`
		SourceID   string `json:"sourceId"`
		TargetType string `json:"targetType"`
		TargetID   string `json:"targetId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validRefTypes[req.SourceType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sourceType must be one of: research, debate, policy"})
		return
	}
	if !validRefTypes[req.TargetType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "targetType must be one of: research, debate, policy"})
		return
	}

	sourceID, err := bson.ObjectIDFromHex(req.SourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sourceId"})
		return
	}
	targetID, err := bson.ObjectIDFromHex(req.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid targetId"})
		return
	}

	if sourceID == targetID && req.SourceType == req.TargetType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reference an entity to itself"})
		return
	}

	ref := &models.CrossReference{
		SourceType: req.SourceType,
		SourceID:   sourceID,
		TargetType: req.TargetType,
		TargetID:   targetID,
		CreatedBy:  userID,
	}

	created, err := h.refs.Create(c, ref)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateCrossReference) {
			c.JSON(http.StatusConflict, gin.H{"error": "this cross-reference already exists"})
			return
		}
		h.logger.Error("failed to create cross-reference", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cross-reference"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"reference": created})
}

// List handles GET /api/cross-references?type=research&id=XXX
func (h *CrossReferenceHandler) List(c *gin.Context) {
	entityType := c.Query("type")
	entityIDStr := c.Query("id")

	if !validRefTypes[entityType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be one of: research, debate, policy"})
		return
	}

	entityID, err := bson.ObjectIDFromHex(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	refs, err := h.refs.GetForEntity(c, entityType, entityID)
	if err != nil {
		h.logger.Error("failed to list cross-references", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cross-references"})
		return
	}

	if refs == nil {
		refs = []*models.CrossReferenceResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"references": refs})
}

// Delete handles DELETE /api/cross-references/:id
func (h *CrossReferenceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.refs.Delete(c, id)
	if err != nil {
		h.logger.Error("failed to delete cross-reference", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "cross-reference not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cross-reference deleted"})
}
