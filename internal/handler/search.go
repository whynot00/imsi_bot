package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/service"
)

// SearchHandler handles device lookup and autocomplete.
type SearchHandler struct {
	svc *service.SearchService
}

func NewSearchHandler(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search godoc
// GET /search?imsi=250018707381530
// GET /search?imei=353312113249201
func (h *SearchHandler) Search(c *gin.Context) {
	ctx := c.Request.Context()

	if imsi := c.Query("imsi"); imsi != "" {
		result, err := h.svc.ByIMSI(ctx, imsi)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if result == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	if imei := c.Query("imei"); imei != "" {
		result, err := h.svc.ByIMEI(ctx, imei)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if result == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "provide imsi or imei query param"})
}

// Suggest godoc
// GET /search/suggest?q=25001&kind=imsi
// GET /search/suggest?q=35331&kind=imei
func (h *SearchHandler) Suggest(c *gin.Context) {
	q := c.Query("q")
	kind := c.Query("kind")

	if q == "" || kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q and kind are required"})
		return
	}

	results, err := h.svc.Suggest(c.Request.Context(), q, kind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
