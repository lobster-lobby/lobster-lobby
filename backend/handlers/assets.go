package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

const (
	maxImageSize = 10 * 1024 * 1024  // 10MB
	maxPDFSize   = 25 * 1024 * 1024  // 25MB
	assetDataDir = "data/campaign-assets"
)

var allowedMimeTypes = map[string]bool{
	"image/png":       true,
	"image/jpeg":      true,
	"image/svg+xml":   true,
	"application/pdf": true,
}

type AssetHandler struct {
	assets    *repository.AssetRepository
	campaigns *repository.CampaignRepository
	users     *repository.UserRepository
	logger    *zap.Logger
}

func NewAssetHandler(
	assets *repository.AssetRepository,
	campaigns *repository.CampaignRepository,
	users *repository.UserRepository,
	logger *zap.Logger,
) *AssetHandler {
	return &AssetHandler{
		assets:    assets,
		campaigns: campaigns,
		users:     users,
		logger:    logger,
	}
}

// CreateTextAsset handles POST /api/campaigns/:id/assets
func (h *AssetHandler) CreateTextAsset(c *gin.Context) {
	campaignID := c.Param("id")

	campaign, err := h.campaigns.GetByID(c, campaignID)
	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Title               string `json:"title"`
		Type                string `json:"type"`
		Content             string `json:"content"`
		Description         string `json:"description"`
		SubjectLine         string `json:"subjectLine"`
		SuggestedRecipients string `json:"suggestedRecipients"`
		AIGenerated         bool   `json:"aiGenerated"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset := &models.CampaignAsset{
		CampaignID:          campaign.ID,
		CreatedBy:           userID,
		CreatedByUsername:   user.Username,
		Title:               req.Title,
		Type:                models.AssetType(req.Type),
		Content:             req.Content,
		Description:         req.Description,
		SubjectLine:         req.SubjectLine,
		SuggestedRecipients: req.SuggestedRecipients,
		AIGenerated:         req.AIGenerated,
	}

	if err := asset.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !asset.IsTextBased() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this endpoint is for text-based assets only, use /upload for file assets"})
		return
	}

	if err := h.assets.Create(c, asset); err != nil {
		h.logger.Error("failed to create asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset"})
		return
	}

	// Update campaign asset count
	h.updateCampaignAssetCount(c, campaignID)

	c.JSON(http.StatusCreated, gin.H{"asset": asset})
}

// UploadAsset handles POST /api/campaigns/:id/assets/upload
func (h *AssetHandler) UploadAsset(c *gin.Context) {
	campaignID := c.Param("id")

	campaign, err := h.campaigns.GetByID(c, campaignID)
	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.FindByID(c, userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxPDFSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Detect MIME type from file bytes (first 512), ignoring client-supplied Content-Type.
	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	mimeType := http.DetectContentType(sniff[:n])
	// Trim any parameters (e.g. "; charset=utf-8")
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	// http.DetectContentType can't distinguish SVG — check original header as fallback
	// only when the sniffed type is generic text/xml or text/plain.
	if mimeType == "text/xml; charset=utf-8" || mimeType == "text/plain; charset=utf-8" || mimeType == "text/xml" || mimeType == "text/plain" {
		ct := header.Header.Get("Content-Type")
		if ct == "image/svg+xml" {
			mimeType = "image/svg+xml"
		}
	}
	// Seek back to beginning so the full file can be read during save.
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}
	if !allowedMimeTypes[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed, must be PNG, JPEG, SVG, or PDF"})
		return
	}

	// Validate size
	maxSize := int64(maxImageSize)
	if mimeType == "application/pdf" {
		maxSize = maxPDFSize
	}
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file too large, max %dMB", maxSize/(1024*1024))})
		return
	}

	// Get form values
	title := c.PostForm("title")
	assetType := c.PostForm("type")
	description := c.PostForm("description")
	aiGenerated := c.PostForm("aiGenerated") == "true"

	// Create asset first to get ID
	asset := &models.CampaignAsset{
		ID:                bson.NewObjectID(),
		CampaignID:        campaign.ID,
		CreatedBy:         userID,
		CreatedByUsername: user.Username,
		Title:             title,
		Type:              models.AssetType(assetType),
		Description:       description,
		FileName:          header.Filename,
		FileSize:          header.Size,
		MimeType:          mimeType,
		AIGenerated:       aiGenerated,
	}

	// Basic validation
	if len(asset.Title) < 3 || len(asset.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 3 and 200 characters"})
		return
	}

	if !models.ValidAssetTypes[asset.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset type"})
		return
	}

	// Create directory structure
	dirPath := filepath.Join(assetDataDir, campaignID, asset.ID.Hex())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		h.logger.Error("failed to create asset directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Sanitize filename
	safeFilename := sanitizeFilename(header.Filename)
	filePath := filepath.Join(dirPath, safeFilename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("failed to create file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error("failed to write file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Set file URL
	asset.FileURL = fmt.Sprintf("/api/campaigns/%s/assets/%s/file", campaignID, asset.ID.Hex())
	asset.FileName = safeFilename

	if err := h.assets.Create(c, asset); err != nil {
		// Clean up file on error
		os.RemoveAll(dirPath)
		h.logger.Error("failed to create asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset"})
		return
	}

	// Update campaign asset count
	h.updateCampaignAssetCount(c, campaignID)

	c.JSON(http.StatusCreated, gin.H{"asset": asset})
}

// List handles GET /api/campaigns/:id/assets
func (h *AssetHandler) List(c *gin.Context) {
	campaignID := c.Param("id")

	campaign, err := h.campaigns.GetByID(c, campaignID)
	if err != nil {
		h.logger.Error("failed to get campaign", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get campaign"})
		return
	}
	if campaign == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	opts := repository.AssetListOpts{
		CampaignID: campaignID,
		Page:       page,
		PerPage:    perPage,
		Sort:       c.DefaultQuery("sort", "top"),
		Type:       c.Query("type"),
	}

	assets, total, err := h.assets.List(c, opts)
	if err != nil {
		h.logger.Error("failed to list assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets":  assets,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}

// Get handles GET /api/campaigns/:id/assets/:assetId
func (h *AssetHandler) Get(c *gin.Context) {
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil {
		h.logger.Error("failed to get asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get asset"})
		return
	}
	if asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Get user vote if authenticated
	userVote := 0
	if userIDStr, exists := c.Get(middleware.ContextUserID); exists {
		vote, _ := h.assets.GetVote(c, assetID, userIDStr.(string))
		if vote != nil {
			userVote = vote.Value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"asset":    asset,
		"userVote": userVote,
	})
}

// ServeFile handles GET /api/campaigns/:id/assets/:assetId/file
func (h *AssetHandler) ServeFile(c *gin.Context) {
	campaignID := c.Param("id")
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	if asset.FileName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no file associated with this asset"})
		return
	}

	filePath := filepath.Join(assetDataDir, campaignID, assetID, asset.FileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", asset.FileName))
	c.File(filePath)
}

// Download handles POST /api/campaigns/:id/assets/:assetId/download
func (h *AssetHandler) Download(c *gin.Context) {
	campaignID := c.Param("id")
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Increment download count
	if err := h.assets.IncrementDownload(c, assetID); err != nil {
		h.logger.Warn("failed to increment download count", zap.Error(err))
	}

	// Update campaign metrics
	h.incrementCampaignDownloads(c, campaignID)

	if asset.FileName != "" {
		// File asset - serve file
		filePath := filepath.Join(assetDataDir, campaignID, assetID, asset.FileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", asset.FileName))
		c.File(filePath)
	} else {
		// Text asset - return JSON
		c.JSON(http.StatusOK, gin.H{
			"asset":   asset,
			"message": "download tracked",
		})
	}
}

// BatchVotes handles GET /api/campaigns/:slug/assets/votes?ids=id1,id2,...
func (h *AssetHandler) BatchVotes(c *gin.Context) {
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusOK, gin.H{"votes": map[string]int{}})
		return
	}

	ids := strings.Split(idsParam, ",")
	// Trim whitespace
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	userIDStr, exists := c.Get(middleware.ContextUserID)
	if !exists {
		c.JSON(http.StatusOK, gin.H{"votes": map[string]int{}})
		return
	}

	votes, err := h.assets.GetBatchVotes(c, ids, userIDStr.(string))
	if err != nil {
		h.logger.Error("failed to get batch votes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get votes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"votes": votes})
}

// Vote handles POST /api/campaigns/:id/assets/:assetId/vote
func (h *AssetHandler) Vote(c *gin.Context) {
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)

	var req struct {
		Value int `json:"value"` // 1, -1, or 0 to clear
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Value != 1 && req.Value != -1 && req.Value != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 1, -1, or 0"})
		return
	}

	if err := h.assets.SetVote(c, assetID, userIDStr.(string), req.Value); err != nil {
		h.logger.Error("failed to set vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to vote"})
		return
	}

	// Get updated asset
	updatedAsset, _ := h.assets.GetByID(c, assetID)

	c.JSON(http.StatusOK, gin.H{
		"asset":    updatedAsset,
		"userVote": req.Value,
	})
}

// Share handles POST /api/campaigns/:id/assets/:assetId/share
func (h *AssetHandler) Share(c *gin.Context) {
	campaignID := c.Param("id")
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	var req struct {
		Platform string `json:"platform"` // twitter, facebook, email, print, other
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validPlatforms := map[string]bool{
		"twitter":  true,
		"facebook": true,
		"email":    true,
		"print":    true,
		"other":    true,
	}

	if !validPlatforms[req.Platform] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform must be twitter, facebook, email, print, or other"})
		return
	}

	if err := h.assets.IncrementShare(c, assetID, req.Platform); err != nil {
		h.logger.Error("failed to increment share", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track share"})
		return
	}

	// Update campaign metrics
	h.incrementCampaignShares(c, campaignID, req.Platform)

	// Get updated asset
	updatedAsset, _ := h.assets.GetByID(c, assetID)

	c.JSON(http.StatusOK, gin.H{
		"asset":   updatedAsset,
		"message": "share tracked",
	})
}

// Update handles PATCH /api/campaigns/:id/assets/:assetId
func (h *AssetHandler) Update(c *gin.Context) {
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	// Check ownership
	if asset.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own assets"})
		return
	}

	var req struct {
		Title               *string `json:"title"`
		Content             *string `json:"content"`
		Description         *string `json:"description"`
		SubjectLine         *string `json:"subjectLine"`
		SuggestedRecipients *string `json:"suggestedRecipients"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := bson.M{}

	if req.Title != nil {
		if len(*req.Title) < 3 || len(*req.Title) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 3 and 200 characters"})
			return
		}
		updates["title"] = *req.Title
	}

	if req.Content != nil {
		if len(*req.Content) < 10 || len(*req.Content) > 50000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content must be between 10 and 50000 characters"})
			return
		}
		updates["content"] = *req.Content
	}

	if req.Description != nil {
		if len(*req.Description) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "description must be at most 1000 characters"})
			return
		}
		updates["description"] = *req.Description
	}

	if req.SubjectLine != nil {
		updates["subjectLine"] = *req.SubjectLine
	}

	if req.SuggestedRecipients != nil {
		updates["suggestedRecipients"] = *req.SuggestedRecipients
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	updatedAsset, err := h.assets.Update(c, assetID, updates)
	if err != nil {
		h.logger.Error("failed to update asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"asset": updatedAsset})
}

// Delete handles DELETE /api/campaigns/:id/assets/:assetId
func (h *AssetHandler) Delete(c *gin.Context) {
	campaignID := c.Param("id")
	assetID := c.Param("assetId")

	asset, err := h.assets.GetByID(c, assetID)
	if err != nil || asset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	// Check authorization: owner or moderator
	isOwner := asset.CreatedBy == userID

	isModerator := false
	user, err := h.users.FindByID(c, userID)
	if err == nil && user != nil {
		isModerator = user.Role == "moderator" || user.Role == "admin"
	}

	if !isOwner && !isModerator {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own assets"})
		return
	}

	// Delete file if exists
	if asset.FileName != "" {
		dirPath := filepath.Join(assetDataDir, campaignID, assetID)
		os.RemoveAll(dirPath)
	}

	if err := h.assets.Delete(c, assetID); err != nil {
		h.logger.Error("failed to delete asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete asset"})
		return
	}

	// Update campaign asset count
	h.updateCampaignAssetCount(c, campaignID)

	c.JSON(http.StatusOK, gin.H{"message": "asset deleted"})
}

// Helper functions

func (h *AssetHandler) updateCampaignAssetCount(c *gin.Context, campaignID string) {
	count, err := h.assets.CountByCampaign(c, campaignID)
	if err != nil {
		h.logger.Warn("failed to count assets", zap.Error(err))
		return
	}
	// Asset count is always a recount (set, not delta), so $set it atomically.
	if _, err := h.campaigns.Update(c, campaignID, bson.M{
		"metrics.assetCount": int(count),
	}); err != nil {
		h.logger.Warn("failed to update campaign asset count", zap.Error(err))
	}
}

func (h *AssetHandler) incrementCampaignDownloads(c *gin.Context, campaignID string) {
	if err := h.campaigns.IncrMetric(c, campaignID, "totalDownloads", 1); err != nil {
		h.logger.Warn("failed to update campaign downloads", zap.Error(err))
	}
	h.campaigns.RecalcTrending(c, campaignID)
}

func (h *AssetHandler) incrementCampaignShares(c *gin.Context, campaignID, platform string) {
	if err := h.campaigns.IncrMetric(c, campaignID, "totalShares", 1); err != nil {
		h.logger.Warn("failed to update campaign shares", zap.Error(err))
	}
	if err := h.campaigns.IncrMetric(c, campaignID, "sharesByPlatform."+platform, 1); err != nil {
		h.logger.Warn("failed to update campaign shares by platform", zap.Error(err))
	}
	h.campaigns.RecalcTrending(c, campaignID)
}

func sanitizeFilename(filename string) string {
	// Get base name and extension
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filepath.Base(filename), ext)

	// Remove any path traversal attempts
	base = strings.ReplaceAll(base, "..", "")
	base = strings.ReplaceAll(base, "/", "")
	base = strings.ReplaceAll(base, "\\", "")

	// Replace spaces and special chars
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, base)

	// Truncate if too long
	if len(base) > 100 {
		base = base[:100]
	}

	if base == "" {
		base = "file"
	}

	return base + ext
}
