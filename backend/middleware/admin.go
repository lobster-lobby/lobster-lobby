package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/lobster-lobby/lobster-lobby/repository"
)

// RequireAdmin aborts with 403 if the authenticated user is not an admin.
// Must be used after RequireAuth (relies on ContextUserID being set).
func RequireAdmin(users *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get(ContextUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, err := bson.ObjectIDFromHex(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}
		user, err := users.FindByID(c, userID)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
