package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type AuthHandler struct {
	users  *repository.UserRepository
	tokens *repository.RefreshTokenRepository
	jwtSvc *services.JWTService
}

func NewAuthHandler(users *repository.UserRepository, tokens *repository.RefreshTokenRepository, jwtSvc *services.JWTService) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens, jwtSvc: jwtSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Type        string `json:"type"`
		DisplayName string `json:"displayName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !usernameRe.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be 3-30 alphanumeric/underscore chars"})
		return
	}

	if req.Type == "" {
		req.Type = "human"
	}
	if req.Type != "human" && req.Type != "agent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be human or agent"})
		return
	}

	if req.Type == "human" && (req.Email == "" || !strings.Contains(req.Email, "@")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid email required for human accounts"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	if existing, _ := h.users.FindByUsername(c, req.Username); existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}
	if req.Email != "" {
		if existing, _ := h.users.FindByEmail(c, req.Email); existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	user := &models.User{
		Username:          req.Username,
		Email:             req.Email,
		PasswordHash:      string(hash),
		Type:              req.Type,
		Verified:          false,
		VerificationLevel: "none",
		DisplayName:       displayName,
		Reputation:        models.ReputationScore{Score: 0, Contributions: 0, Tier: "new"},
		Bookmarks:         []bson.ObjectID{},
	}

	if err := h.users.Create(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	access, refresh, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user, "accessToken": access, "refreshToken": refresh})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user *models.User
	if req.Email != "" {
		user, _ = h.users.FindByEmail(c, req.Email)
	} else if req.Username != "" {
		user, _ = h.users.FindByUsername(c, req.Username)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or username required"})
		return
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	_ = h.users.UpdateLastLogin(c, user.ID)

	access, refresh, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "accessToken": access, "refreshToken": refresh})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refreshToken required"})
		return
	}

	rt, err := h.tokens.FindByToken(c, req.RefreshToken)
	if err != nil || rt == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	_ = h.tokens.DeleteByToken(c, req.RefreshToken)

	user, err := h.users.FindByID(c, rt.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	access, refresh, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": access, "refreshToken": refresh})
}

func (h *AuthHandler) Me(c *gin.Context) {
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

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) issueTokenPair(c *gin.Context, user *models.User) (string, string, error) {
	access, err := h.jwtSvc.GenerateAccessToken(user.ID.Hex(), user.Type)
	if err != nil {
		return "", "", err
	}

	refresh, expiresAt, err := h.jwtSvc.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	rt := &repository.RefreshToken{
		Token:     refresh,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	if err := h.tokens.Create(c, rt); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}
