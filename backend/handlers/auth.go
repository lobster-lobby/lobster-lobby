package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	users    *repository.UserRepository
	tokens   *repository.RefreshTokenRepository
	jwtSvc   *services.JWTService
	secureCookie bool
}

func NewAuthHandler(users *repository.UserRepository, tokens *repository.RefreshTokenRepository, jwtSvc *services.JWTService, env string) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens, jwtSvc: jwtSvc, secureCookie: env == "production"}
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

	access, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user, "token": access})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalise: prefer the combined "identifier" field, fall back to explicit fields.
	// We treat the identifier as an email when it contains "@" — a lightweight heuristic
	// that avoids a round-trip DB look-up while covering all realistic login inputs
	// (usernames are restricted to alphanumeric + underscore and cannot contain "@").
	email := req.Email
	username := req.Username
	if req.Identifier != "" {
		if strings.Contains(req.Identifier, "@") {
			email = req.Identifier
		} else {
			username = req.Identifier
		}
	}

	var user *models.User
	if email != "" {
		user, _ = h.users.FindByEmail(c, email)
	} else if username != "" {
		user, _ = h.users.FindByUsername(c, username)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or username required"})
		return
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	_ = h.users.UpdateLastLogin(c, user.ID)

	access, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": access})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	tokenVal, err := c.Cookie(refreshCookieName)
	if err != nil || tokenVal == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no refresh token"})
		return
	}

	rt, err := h.tokens.FindByToken(c, tokenVal)
	if err != nil || rt == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	_ = h.tokens.DeleteByToken(c, tokenVal)

	user, err := h.users.FindByID(c, rt.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	access, err := h.issueTokenPair(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": access})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	tokenVal, err := c.Cookie(refreshCookieName)
	if err == nil && tokenVal != "" {
		_ = h.tokens.DeleteByToken(c, tokenVal)
	}
	c.SetCookie(refreshCookieName, "", -1, "/api/auth", "", h.secureCookie, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
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

func (h *AuthHandler) issueTokenPair(c *gin.Context, user *models.User) (string, error) {
	access, err := h.jwtSvc.GenerateAccessToken(user.ID.Hex(), user.Type, user.Role, user.Username, user.Email)
	if err != nil {
		return "", err
	}

	refresh, expiresAt, err := h.jwtSvc.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	rt := &repository.RefreshToken{
		Token:     refresh,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	if err := h.tokens.Create(c, rt); err != nil {
		return "", err
	}

	maxAge := int(time.Until(expiresAt).Seconds())
	// Scope the refresh-token cookie to /api/auth so it is never sent with
	// regular API calls — this limits the attack surface if XSS occurs.
	// The token-refresh and logout endpoints both live under this path, so
	// restricting the scope does not break any legitimate flow.
	c.SetCookie(refreshCookieName, refresh, maxAge, "/api/auth", "", h.secureCookie, true)

	return access, nil
}
