package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lobster-lobby/lobster-lobby/services"
)

const (
	ContextUserID   = "userID"
	ContextUserType = "userType"
)

func RequireAuth(jwtSvc *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := extractClaims(c, jwtSvc)
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(ContextUserID, claims.Subject)
		c.Set(ContextUserType, claims.Type)
		c.Next()
	}
}

func OptionalAuth(jwtSvc *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := extractClaims(c, jwtSvc)
		if claims != nil {
			c.Set(ContextUserID, claims.Subject)
			c.Set(ContextUserType, claims.Type)
		}
		c.Next()
	}
}

func extractClaims(c *gin.Context, jwtSvc *services.JWTService) (*services.Claims, error) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, nil
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	return jwtSvc.ValidateAccessToken(token)
}
