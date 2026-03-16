package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/services"
)

// SearchHandler handles search requests.
type SearchHandler struct {
	search *services.SearchService
	logger *zap.Logger
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(search *services.SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{search: search, logger: logger}
}

// Search handles GET /api/search
func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filters := services.SearchFilters{
		Type:   c.Query("type"),
		Level:  c.Query("level"),
		State:  c.Query("state"),
		Status: c.Query("status"),
	}

	res, err := h.search.SearchPolicies(c.Request.Context(), q, filters, page, perPage)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, res)
}
