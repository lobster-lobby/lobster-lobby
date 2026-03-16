package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

type PolicyHandler struct {
	policies      *repository.PolicyRepository
	users         *repository.UserRepository
	jwtSvc        *services.JWTService
	logger        *zap.Logger
	reputationSvc *services.ReputationService
	searchSvc     *services.SearchService
}

func NewPolicyHandler(policies *repository.PolicyRepository, users *repository.UserRepository, jwtSvc *services.JWTService, logger *zap.Logger, reputationSvc *services.ReputationService, searchSvc *services.SearchService) *PolicyHandler {
	return &PolicyHandler{policies: policies, users: users, jwtSvc: jwtSvc, logger: logger, reputationSvc: reputationSvc, searchSvc: searchSvc}
}

func (h *PolicyHandler) Create(c *gin.Context) {
	var req struct {
		Title          string   `json:"title"`
		Summary        string   `json:"summary"`
		Type           string   `json:"type"`
		Level          string   `json:"level"`
		State          string   `json:"state"`
		Status         string   `json:"status"`
		ExternalURL    string   `json:"externalUrl"`
		BillNumber     string   `json:"billNumber"`
		Tags           []string `json:"tags"`
		LinkedPolicies []string `json:"linkedPolicies"`
		ParentPolicy   string   `json:"parentPolicy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	slug, err := h.policies.GenerateSlug(c, req.Title)
	if err != nil {
		h.logger.Error("failed to generate slug", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate slug"})
		return
	}

	policy := &models.Policy{
		Title:       req.Title,
		Slug:        slug,
		Summary:     req.Summary,
		Type:        models.PolicyType(req.Type),
		Level:       models.PolicyLevel(req.Level),
		State:       req.State,
		Status:      models.PolicyStatus(req.Status),
		ExternalURL: req.ExternalURL,
		BillNumber:  req.BillNumber,
		Tags:        req.Tags,
		CreatedBy:   userID,
	}

	if len(req.LinkedPolicies) > 0 {
		linkedIDs := make([]bson.ObjectID, 0, len(req.LinkedPolicies))
		for _, idStr := range req.LinkedPolicies {
			if oid, err := bson.ObjectIDFromHex(idStr); err == nil {
				linkedIDs = append(linkedIDs, oid)
			}
		}
		policy.LinkedPolicies = linkedIDs
	}

	if req.ParentPolicy != "" {
		if oid, err := bson.ObjectIDFromHex(req.ParentPolicy); err == nil {
			policy.ParentPolicy = &oid
		}
	}

	if err := policy.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.policies.Create(c, policy); err != nil {
		h.logger.Error("failed to create policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create policy"})
		return
	}

	go func() {
		if err := h.reputationSvc.AwardPoints(c.Request.Context(), userID, models.ActionPolicyCreated, policy.ID.Hex(), "policy"); err != nil {
			h.logger.Error("failed to award reputation points", zap.Error(err))
		}
	}()

	go func() {
		if err := h.searchSvc.IndexPolicy(c.Request.Context(), policy); err != nil {
			h.logger.Warn("failed to index policy in search", zap.String("id", policy.ID.Hex()), zap.Error(err))
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"policy": policy})
}

func (h *PolicyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	var tags []string
	if tagsParam := c.Query("tags"); tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	opts := repository.PolicyListOpts{
		Page:      page,
		PerPage:   perPage,
		Sort:      c.DefaultQuery("sort", "hot"),
		Type:      c.Query("type"),
		Level:     c.Query("level"),
		State:     c.Query("state"),
		Tags:      tags,
		CreatedBy: c.Query("createdBy"),
	}

	policies, total, err := h.policies.List(c, opts)
	if err != nil {
		h.logger.Error("failed to list policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list policies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"total":    total,
		"page":     page,
		"perPage":  perPage,
	})
}

func (h *PolicyHandler) Get(c *gin.Context) {
	idOrSlug := c.Param("idOrSlug")

	var policy *models.Policy
	var err error

	if oid, parseErr := bson.ObjectIDFromHex(idOrSlug); parseErr == nil {
		policy, err = h.policies.FindByID(c, oid)
	} else {
		policy, err = h.policies.FindBySlug(c, idOrSlug)
	}

	if err != nil {
		h.logger.Error("failed to get policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get policy"})
		return
	}

	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	go func() {
		_ = h.policies.IncrementEngagement(c.Request.Context(), policy.ID, "engagement.viewCount")
	}()

	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

func (h *PolicyHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	policy, err := h.policies.FindByID(c, policyID)
	if err != nil {
		h.logger.Error("failed to find policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find policy"})
		return
	}
	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	if policy.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own policies"})
		return
	}

	var req struct {
		Title       *string  `json:"title"`
		Summary     *string  `json:"summary"`
		Tags        []string `json:"tags"`
		Status      *string  `json:"status"`
		ExternalURL *string  `json:"externalUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := bson.M{}

	if req.Title != nil {
		if len(*req.Title) < 5 || len(*req.Title) > 300 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 5 and 300 characters"})
			return
		}
		updates["title"] = *req.Title
		newSlug, err := h.policies.GenerateSlug(c, *req.Title)
		if err != nil {
			h.logger.Error("failed to generate slug", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate slug"})
			return
		}
		updates["slug"] = newSlug
	}

	if req.Summary != nil {
		if len(*req.Summary) < 20 || len(*req.Summary) > 5000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "summary must be between 20 and 5000 characters"})
			return
		}
		updates["summary"] = *req.Summary
	}

	if req.Tags != nil {
		if len(req.Tags) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tags cannot exceed 10 items"})
			return
		}
		updates["tags"] = req.Tags
	}

	if req.Status != nil {
		status := models.PolicyStatus(*req.Status)
		if status != models.PolicyStatusActive && status != models.PolicyStatusPassed &&
			status != models.PolicyStatusFailed && status != models.PolicyStatusWithdrawn {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of: active, passed, failed, withdrawn"})
			return
		}
		updates["status"] = status
	}

	if req.ExternalURL != nil {
		updates["externalUrl"] = *req.ExternalURL
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	if err := h.policies.Update(c, policyID, updates); err != nil {
		h.logger.Error("failed to update policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update policy"})
		return
	}

	updatedPolicy, _ := h.policies.FindByID(c, policyID)
	if updatedPolicy != nil {
		go func() {
			if err := h.searchSvc.IndexPolicy(c.Request.Context(), updatedPolicy); err != nil {
				h.logger.Warn("failed to re-index policy in search", zap.String("id", policyID.Hex()), zap.Error(err))
			}
		}()
	}
	c.JSON(http.StatusOK, gin.H{"policy": updatedPolicy})
}

func (h *PolicyHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	policy, err := h.policies.FindByID(c, policyID)
	if err != nil {
		h.logger.Error("failed to find policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find policy"})
		return
	}
	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	isOwner := policy.CreatedBy == userID

	isModerator := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isModerator = user.Role == "moderator" || user.Role == "admin"
	}

	if !isOwner && !isModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own policies"})
		return
	}

	if err := h.policies.Delete(c, policyID); err != nil {
		h.logger.Error("failed to delete policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete policy"})
		return
	}

	go func() {
		if err := h.searchSvc.RemovePolicy(c.Request.Context(), policyID.Hex()); err != nil {
			h.logger.Warn("failed to remove policy from search", zap.String("id", policyID.Hex()), zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "policy archived"})
}
