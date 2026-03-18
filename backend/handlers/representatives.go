package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// RepresentativeStore abstracts representative data access for testability.
type RepresentativeStore interface {
	Create(ctx context.Context, rep *models.Representative) error
	FindByID(ctx context.Context, id bson.ObjectID) (*models.Representative, error)
	FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]models.Representative, error)
	Update(ctx context.Context, id bson.ObjectID, updates bson.M) error
	Delete(ctx context.Context, id bson.ObjectID) error
	List(ctx context.Context, opts repository.RepListOpts) ([]models.Representative, int64, error)
}

// VotingRecordStore abstracts voting record data access for testability.
type VotingRecordStore interface {
	Create(ctx context.Context, vr *models.VotingRecord) error
	FindByRepresentative(ctx context.Context, repID bson.ObjectID, opts repository.VoteListOpts) ([]models.VotingRecord, int64, error)
	FindByPolicy(ctx context.Context, policyID bson.ObjectID, opts repository.VoteListOpts) ([]models.VotingRecord, int64, error)
	GetSummary(ctx context.Context, repID bson.ObjectID) (*models.VotingSummary, error)
	GetPolicySummary(ctx context.Context, policyID bson.ObjectID) (*models.VotingSummary, error)
}

// CivicLookupService abstracts the Google Civic API lookup.
type CivicLookupService interface {
	LookupByAddress(ctx context.Context, address string) ([]models.CivicOfficial, error)
}

type RepresentativeHandler struct {
	reps   RepresentativeStore
	votes  VotingRecordStore
	civic  CivicLookupService
	logger *zap.Logger
}

func NewRepresentativeHandler(reps RepresentativeStore, votes VotingRecordStore, civic CivicLookupService, logger *zap.Logger) *RepresentativeHandler {
	return &RepresentativeHandler{reps: reps, votes: votes, civic: civic, logger: logger}
}

// List handles GET /api/representatives
// Supports: ?address= (civic lookup), or ?search=&party=&state=&chamber=&page=&perPage= (DB listing)
func (h *RepresentativeHandler) List(c *gin.Context) {
	// Address lookup takes precedence - calls external API
	if address := c.Query("address"); address != "" {
		if h.civic == nil {
			c.JSON(http.StatusOK, gin.H{"officials": []interface{}{}})
			return
		}
		officials, err := h.civic.LookupByAddress(c, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup representatives"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"officials": officials})
		return
	}

	// DB listing with filters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	opts := repository.RepListOpts{
		Page:    page,
		PerPage: perPage,
		Search:  c.Query("search"),
		Party:   c.Query("party"),
		State:   c.Query("state"),
		Chamber: c.Query("chamber"),
	}

	reps, total, err := h.reps.List(c, opts)
	if err != nil {
		h.logger.Error("failed to list representatives", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list representatives"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"representatives": reps,
		"total":           total,
		"page":            opts.Page,
		"perPage":         opts.PerPage,
	})
}

// GetByID handles GET /api/representatives/:id — returns profile with voting summary
func (h *RepresentativeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid representative id"})
		return
	}

	rep, err := h.reps.FindByID(c, id)
	if err != nil {
		h.logger.Error("failed to find representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representative"})
		return
	}
	if rep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	summary, err := h.votes.GetSummary(c, id)
	if err != nil {
		h.logger.Error("failed to get voting summary", zap.Error(err))
		// Still return rep, just with empty summary
		summary = &models.VotingSummary{}
	}

	c.JSON(http.StatusOK, gin.H{
		"representative": rep,
		"votingSummary":  summary,
	})
}

// Create handles POST /api/representatives (admin-only)
func (h *RepresentativeHandler) Create(c *gin.Context) {
	var req struct {
		Name        string             `json:"name"`
		Title       string             `json:"title"`
		Party       string             `json:"party"`
		State       string             `json:"state"`
		District    string             `json:"district"`
		Chamber     string             `json:"chamber"`
		Level       string             `json:"level"`
		Bio         string             `json:"bio"`
		PhotoURL    string             `json:"photoUrl"`
		ContactInfo models.ContactInfo `json:"contactInfo"`
		SocialMedia map[string]string  `json:"socialMedia"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rep := &models.Representative{
		Name:        req.Name,
		Title:       req.Title,
		Party:       req.Party,
		State:       req.State,
		District:    req.District,
		Chamber:     req.Chamber,
		Level:       req.Level,
		Bio:         req.Bio,
		PhotoURL:    req.PhotoURL,
		ContactInfo: req.ContactInfo,
		SocialMedia: req.SocialMedia,
	}

	if err := rep.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reps.Create(c, rep); err != nil {
		h.logger.Error("failed to create representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create representative"})
		return
	}

	c.JSON(http.StatusCreated, rep)
}

// Update handles PUT /api/representatives/:id (admin-only)
func (h *RepresentativeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid representative id"})
		return
	}

	existing, err := h.reps.FindByID(c, id)
	if err != nil {
		h.logger.Error("failed to find representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representative"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only allow updating specific fields
	allowed := map[string]string{
		"name": "name", "title": "title", "party": "party",
		"state": "state", "district": "district", "chamber": "chamber",
		"level": "level", "bio": "bio", "photoUrl": "photoUrl",
		"contactInfo": "contactInfo", "socialMedia": "socialMedia",
	}

	updates := bson.M{}
	for key, bsonKey := range allowed {
		if val, ok := req[key]; ok {
			updates[bsonKey] = val
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields to update"})
		return
	}

	if err := h.reps.Update(c, id, updates); err != nil {
		h.logger.Error("failed to update representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update representative"})
		return
	}

	// Re-fetch updated representative
	updated, err := h.reps.FindByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "updated but failed to fetch"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// Delete handles DELETE /api/representatives/:id (admin-only)
func (h *RepresentativeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid representative id"})
		return
	}

	existing, err := h.reps.FindByID(c, id)
	if err != nil {
		h.logger.Error("failed to find representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representative"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	if err := h.reps.Delete(c, id); err != nil {
		h.logger.Error("failed to delete representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete representative"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "representative deleted"})
}

// ListVotes handles GET /api/representatives/:id/votes — paginated voting records
func (h *RepresentativeHandler) ListVotes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid representative id"})
		return
	}

	// Verify representative exists
	rep, err := h.reps.FindByID(c, id)
	if err != nil {
		h.logger.Error("failed to find representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representative"})
		return
	}
	if rep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	records, total, err := h.votes.FindByRepresentative(c, id, repository.VoteListOpts{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		h.logger.Error("failed to list voting records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list voting records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"votes":   records,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}

// RecordVote handles POST /api/representatives/:id/votes (admin-only)
func (h *RepresentativeHandler) RecordVote(c *gin.Context) {
	idStr := c.Param("id")
	repID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid representative id"})
		return
	}

	// Verify representative exists
	rep, err := h.reps.FindByID(c, repID)
	if err != nil {
		h.logger.Error("failed to find representative", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representative"})
		return
	}
	if rep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	var req struct {
		PolicyID string          `json:"policyId"`
		Vote     models.VoteType `json:"vote"`
		Date     string          `json:"date"`
		Session  string          `json:"session"`
		Notes    string          `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policyID, err := bson.ObjectIDFromHex(req.PolicyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	voteDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be in YYYY-MM-DD format"})
		return
	}

	vr := &models.VotingRecord{
		RepresentativeID: repID,
		PolicyID:         policyID,
		Vote:             req.Vote,
		Date:             voteDate,
		Session:          req.Session,
		Notes:            req.Notes,
	}

	if err := vr.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.votes.Create(c, vr); err != nil {
		h.logger.Error("failed to record vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record vote"})
		return
	}

	c.JSON(http.StatusCreated, vr)
}

// ListVotesByPolicy handles GET /api/policies/:id/votes — voting records for a policy with representative details
func (h *RepresentativeHandler) ListVotesByPolicy(c *gin.Context) {
	idStr := c.Param("id")
	policyID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "50"))

	records, total, err := h.votes.FindByPolicy(c, policyID, repository.VoteListOpts{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		h.logger.Error("failed to list policy voting records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list voting records"})
		return
	}

	summary, err := h.votes.GetPolicySummary(c, policyID)
	if err != nil {
		h.logger.Error("failed to get policy voting summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get voting summary"})
		return
	}

	// Resolve representative details via batch query
	type VoteWithRep struct {
		models.VotingRecord
		Representative *models.Representative `json:"representative,omitempty"`
	}

	// Collect unique rep IDs
	seen := make(map[bson.ObjectID]struct{})
	var repIDs []bson.ObjectID
	for _, rec := range records {
		if _, ok := seen[rec.RepresentativeID]; !ok {
			seen[rec.RepresentativeID] = struct{}{}
			repIDs = append(repIDs, rec.RepresentativeID)
		}
	}

	// Batch fetch all representatives
	repMap := make(map[bson.ObjectID]*models.Representative)
	if len(repIDs) > 0 {
		reps, err := h.reps.FindByIDs(c, repIDs)
		if err == nil {
			for i := range reps {
				repMap[reps[i].ID] = &reps[i]
			}
		}
	}

	enriched := make([]VoteWithRep, len(records))
	for i, rec := range records {
		enriched[i] = VoteWithRep{VotingRecord: rec}
		if rep, ok := repMap[rec.RepresentativeID]; ok {
			enriched[i].Representative = rep
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"votes":   enriched,
		"summary": summary,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}
