package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lobster-lobby/lobster-lobby/services"
)

type RepresentativeHandler struct {
	svc *services.RepresentativeService
}

func NewRepresentativeHandler(svc *services.RepresentativeService) *RepresentativeHandler {
	return &RepresentativeHandler{svc: svc}
}

// List handles GET /api/representatives with query parameters:
// - address: lookup via Google Civic API (stateless, no storage)
// - state: lookup from DB cache (e.g., state=MI)
// - district: lookup from DB cache (e.g., district=MI-13)
func (h *RepresentativeHandler) List(c *gin.Context) {
	address := c.Query("address")
	state := c.Query("state")
	district := c.Query("district")

	// Address lookup takes precedence - calls external API, stateless
	if address != "" {
		officials, err := h.svc.LookupByAddress(c, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup representatives"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"officials": officials})
		return
	}

	// District lookup from DB cache
	if district != "" {
		reps, err := h.svc.GetByDistrict(c, district)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representatives"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"representatives": reps})
		return
	}

	// State lookup from DB cache
	if state != "" {
		reps, err := h.svc.GetByState(c, state)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch representatives"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"representatives": reps})
		return
	}

	// No query params - return empty result
	c.JSON(http.StatusOK, gin.H{"officials": []interface{}{}})
}

// GetByID handles GET /api/representatives/:id
func (h *RepresentativeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rep, err := h.svc.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "representative not found"})
		return
	}

	c.JSON(http.StatusOK, rep)
}
