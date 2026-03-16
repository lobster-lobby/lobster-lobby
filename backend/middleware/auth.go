package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

const (
	ContextUserID   = "userID"
	ContextUserType = "userType"
)

const lastUsedDebounce = 1 * time.Minute

func RequireAuth(jwtSvc *services.JWTService, apiKeyRepo *repository.APIKeyRepository, apiKeySvc *services.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tryAPIKeyAuth(c, apiKeyRepo, apiKeySvc) {
			c.Next()
			return
		}

		claims, err := extractClaims(c, jwtSvc)
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(ContextUserID, claims.Subject)
		c.Set(ContextUserType, claims.Type)
		c.Set(ContextAuthMethod, "jwt")
		c.Next()
	}
}

func OptionalAuth(jwtSvc *services.JWTService, apiKeyRepo *repository.APIKeyRepository, apiKeySvc *services.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tryAPIKeyAuth(c, apiKeyRepo, apiKeySvc) {
			c.Next()
			return
		}

		claims, _ := extractClaims(c, jwtSvc)
		if claims != nil {
			c.Set(ContextUserID, claims.Subject)
			c.Set(ContextUserType, claims.Type)
			c.Set(ContextAuthMethod, "jwt")
		}
		c.Next()
	}
}

func tryAPIKeyAuth(c *gin.Context, apiKeyRepo *repository.APIKeyRepository, apiKeySvc *services.APIKeyService) bool {
	apiKeyHeader := c.GetHeader("X-API-Key")
	if apiKeyHeader == "" || !strings.HasPrefix(apiKeyHeader, "ll_") {
		return false
	}

	prefix := apiKeySvc.ExtractPrefix(apiKeyHeader)
	if prefix == "" {
		return false
	}

	key, err := apiKeyRepo.FindByPrefix(c, prefix)
	if err != nil || key == nil {
		return false
	}

	if key.Revoked {
		return false
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return false
	}

	if !apiKeySvc.VerifyKey(apiKeyHeader, key.KeyHash) {
		return false
	}

	c.Set(ContextUserID, key.UserID.Hex())
	c.Set(ContextUserType, "agent")
	c.Set(ContextAuthMethod, "apikey")
	c.Set("apiKeyID", key.ID.Hex())
	if key.RateLimit > 0 {
		c.Set("apiKeyRateLimit", key.RateLimit)
	}

	// Debounced lastUsedAt update
	if key.LastUsedAt == nil || time.Since(*key.LastUsedAt) > lastUsedDebounce {
		go apiKeyRepo.UpdateLastUsed(c.Copy(), key.ID)
	}

	return true
}

func extractClaims(c *gin.Context, jwtSvc *services.JWTService) (*services.Claims, error) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, nil
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	return jwtSvc.ValidateAccessToken(token)
}
