package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type UserHandler struct {
	users    *repository.UserRepository
	activity *repository.ActivityRepository
}

func NewUserHandler(users *repository.UserRepository, activity *repository.ActivityRepository) *UserHandler {
	return &UserHandler{users: users, activity: activity}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	user, err := h.users.FindByUsername(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	bookmarkCount := 0
	if user.Bookmarks != nil {
		bookmarkCount = len(user.Bookmarks)
	}
	stats, _ := h.activity.GetUserStats(c, user.ID, bookmarkCount)

	// Check if the requester is viewing their own profile
	isOwnProfile := false
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		if userID, err := bson.ObjectIDFromHex(userIDStr.(string)); err == nil {
			isOwnProfile = userID == user.ID
		}
	}

	// Public profile response - hide email unless viewing own profile
	response := gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"bio":         user.Bio,
		"type":        user.Type,
		"role":        user.Role,
		"reputation":  user.Reputation,
		"createdAt":   user.CreatedAt,
		"stats":       stats,
		"isOwnProfile": isOwnProfile,
	}

	if isOwnProfile {
		response["email"] = user.Email
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		Bio         string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	currentUser, err := h.users.FindByID(c, userID)
	if err != nil || currentUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Validate username if changed
	if req.Username != "" && req.Username != currentUser.Username {
		if !usernameRe.MatchString(req.Username) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username must be 3-30 alphanumeric/underscore chars"})
			return
		}
		// Check if username is taken
		existing, _ := h.users.FindByUsername(c, req.Username)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
	}

	// Validate email if changed
	if req.Email != "" && req.Email != currentUser.Email {
		if !strings.Contains(req.Email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "valid email required"})
			return
		}
		// Check if email is taken
		existing, _ := h.users.FindByEmail(c, req.Email)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
	}

	// Use current values if not provided
	username := req.Username
	if username == "" {
		username = currentUser.Username
	}
	email := req.Email
	if email == "" {
		email = currentUser.Email
	}
	displayName := req.DisplayName
	if displayName == "" {
		displayName = currentUser.DisplayName
	}

	if err := h.users.Update(c, userID, username, email, displayName, req.Bio); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	// Fetch updated user
	updatedUser, _ := h.users.FindByID(c, userID)
	c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current and new password required"})
		return
	}

	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	// Get current user
	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify current password
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := h.users.UpdatePassword(c, userID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userIDStr, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required for account deletion"})
		return
	}

	// Get current user
	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify password
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password is incorrect"})
		return
	}

	if err := h.users.Delete(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}
