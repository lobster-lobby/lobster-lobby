package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type APIKeyHandler struct {
	keys      *repository.APIKeyRepository
	apiKeySvc *services.APIKeyService
}

func NewAPIKeyHandler(keys *repository.APIKeyRepository, apiKeySvc *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keys: keys, apiKeySvc: apiKeySvc}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		ExpiresIn   *int     `json:"expiresIn"` // days until expiration, optional
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	fullKey, prefix, hash, err := h.apiKeySvc.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}

	if req.Permissions == nil {
		req.Permissions = []string{"read"}
	}

	apiKey := &models.APIKey{
		UserID:      userID,
		Name:        req.Name,
		KeyHash:     hash,
		KeyPrefix:   prefix,
		Permissions: req.Permissions,
	}

	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		expiresAt := time.Now().UTC().AddDate(0, 0, *req.ExpiresIn)
		apiKey.ExpiresAt = &expiresAt
	}

	if err := h.keys.Create(c, apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":     fullKey,
		"id":      apiKey.ID,
		"name":    apiKey.Name,
		"prefix":  apiKey.KeyPrefix,
		"message": "Store this key securely — it will not be shown again",
	})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	keys, err := h.keys.FindByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}

	if keys == nil {
		keys = []models.APIKey{}
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	keyID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	if err := h.keys.Revoke(c, keyID, userID); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
}
