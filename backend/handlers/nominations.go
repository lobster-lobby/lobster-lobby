package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type NominationHandler struct {
	nominations *repository.NominationRepository
	policies    *repository.PolicyRepository
	users       *repository.UserRepository
	logger      *zap.Logger
}

func NewNominationHandler(
	nominations *repository.NominationRepository,
	policies *repository.PolicyRepository,
	users *repository.UserRepository,
	logger *zap.Logger,
) *NominationHandler {
	return &NominationHandler{
		nominations: nominations,
		policies:    policies,
		users:       users,
		logger:      logger,
	}
}

// Nominate creates a campaign nomination for a policy.
func (h *NominationHandler) Nominate(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
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

	// Fetch policy
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

	// Check no existing nomination
	existing, err := h.nominations.FindByPolicyID(c, policyID)
	if err != nil {
		h.logger.Error("failed to check existing nomination", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing nomination"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "a nomination already exists for this policy"})
		return
	}

	// Check eligibility: >= 10 debate comments
	debateCount, err := h.nominations.CountDebateComments(c, policyID)
	if err != nil {
		h.logger.Error("failed to count debate comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check eligibility"})
		return
	}
	if debateCount < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy must have at least 10 debate comments"})
		return
	}

	// Check eligibility: >= 3 research submissions
	researchCount, err := h.nominations.CountResearchSubmissions(c, policyID)
	if err != nil {
		h.logger.Error("failed to count research submissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check eligibility"})
		return
	}
	if researchCount < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy must have at least 3 research submissions"})
		return
	}

	// Check nominator: must be policy creator OR have reputation >= 200
	user, err := h.users.FindByID(c, userID)
	if err != nil {
		h.logger.Error("failed to find user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	isCreator := policy.CreatedBy == userID
	hasReputation := user.Reputation.Score >= 200
	if !isCreator && !hasReputation {
		c.JSON(http.StatusForbidden, gin.H{"error": "you must be the policy creator or have reputation >= 200 to nominate"})
		return
	}

	nomination := &models.CampaignNomination{
		PolicyID:    policyID,
		NominatedBy: userID,
		Status:      models.NominationStatusPending,
	}

	if err := nomination.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.nominations.Create(c, nomination); err != nil {
		h.logger.Error("failed to create nomination", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create nomination"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"nomination": nomination})
}

// Endorse adds an endorsement to an existing nomination.
func (h *NominationHandler) Endorse(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
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

	// Check user reputation >= 50
	user, err := h.users.FindByID(c, userID)
	if err != nil {
		h.logger.Error("failed to find user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if user.Reputation.Score < models.EndorserMinReputation {
		c.JSON(http.StatusForbidden, gin.H{"error": "reputation must be at least 50 to endorse nominations"})
		return
	}

	// Check nomination exists
	existing, err := h.nominations.FindByPolicyID(c, policyID)
	if err != nil {
		h.logger.Error("failed to find nomination", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find nomination"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no nomination found for this policy"})
		return
	}
	if existing.Status != models.NominationStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nomination is no longer pending"})
		return
	}

	// Check not already endorsed by this user
	for _, e := range existing.Endorsers {
		if e.UserID == userID {
			c.JSON(http.StatusConflict, gin.H{"error": "you have already endorsed this nomination"})
			return
		}
	}

	nomination, err := h.nominations.AddEndorsement(c, policyID, userID)
	if err != nil {
		h.logger.Error("failed to add endorsement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add endorsement"})
		return
	}
	if nomination == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "endorsement could not be added"})
		return
	}

	// Auto-transition: when 5th endorsement hits, set policy status to ready_for_campaign
	if len(nomination.Endorsers) >= models.EndorsementsRequired && nomination.Status == models.NominationStatusPending {
		go h.transitionToReadyForCampaign(context.Background(), policyID)
	}

	c.JSON(http.StatusOK, gin.H{"nomination": nomination})
}

func (h *NominationHandler) transitionToReadyForCampaign(ctx context.Context, policyID bson.ObjectID) {
	if err := h.nominations.UpdateStatus(ctx, policyID, models.NominationStatusApproved); err != nil {
		h.logger.Error("failed to update nomination status", zap.Error(err))
		return
	}

	if err := h.policies.Update(ctx, policyID, bson.M{"status": string(models.PolicyStatusReadyForCampaign)}); err != nil {
		h.logger.Error("failed to update policy status to ready_for_campaign", zap.Error(err))
	}
}

// CampaignReadiness checks the eligibility of a policy for campaign nomination.
func (h *NominationHandler) CampaignReadiness(c *gin.Context) {
	policyIDStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(policyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
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

	debateCount, _ := h.nominations.CountDebateComments(c, policyID)
	researchCount, _ := h.nominations.CountResearchSubmissions(c, policyID)

	nomination, _ := h.nominations.FindByPolicyID(c, policyID)

	endorsementCount := 0
	nominationStatus := ""
	if nomination != nil {
		endorsementCount = len(nomination.Endorsers)
		nominationStatus = string(nomination.Status)
	}

	missing := []string{}
	if debateCount < 10 {
		missing = append(missing, "needs at least 10 debate comments")
	}
	if researchCount < 3 {
		missing = append(missing, "needs at least 3 research submissions")
	}
	if nomination == nil {
		missing = append(missing, "nomination not yet created")
	} else if endorsementCount < models.EndorsementsRequired {
		missing = append(missing, "needs more endorsements")
	}

	eligible := debateCount >= 10 && researchCount >= 3
	isReady := policy.Status == models.PolicyStatusReadyForCampaign

	c.JSON(http.StatusOK, gin.H{
		"eligible":              eligible,
		"isReady":               isReady,
		"debateCount":           debateCount,
		"debateRequired":        10,
		"researchCount":         researchCount,
		"researchRequired":      3,
		"endorsementCount":      endorsementCount,
		"endorsementsRequired":  models.EndorsementsRequired,
		"nominationStatus":      nominationStatus,
		"missing":               missing,
	})
}
