package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/whynot00/imsi-bot/internal/parser"
	"github.com/whynot00/imsi-bot/internal/service"
)

// InternalUploadHandler handles large file uploads via API key auth.
type InternalUploadHandler struct {
	svc *service.ImportService
}

func NewInternalUploadHandler(svc *service.ImportService) *InternalUploadHandler {
	return &InternalUploadHandler{svc: svc}
}

// UploadParametr godoc
// POST /internal/upload/parametr  (multipart/form-data, field "file")
func (h *InternalUploadHandler) UploadParametr(c *gin.Context) {
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer f.Close()

	result, err := parser.Parse(f)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	imported, err := h.svc.ImportParametr(c.Request.Context(), result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices":   imported.Devices,
		"locations": imported.Locations,
		"sightings": imported.Sightings,
		"skipped":   imported.Skipped,
	})
}

// UploadRK godoc
// POST /internal/upload/rk  (multipart/form-data, field "file")
func (h *InternalUploadHandler) UploadRK(c *gin.Context) {
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer f.Close()

	devices, err := parser.ParseRaw(f)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	imported, err := h.svc.ImportRK(c.Request.Context(), devices)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices":   imported.Devices,
		"sightings": imported.Sightings,
		"skipped":   imported.Skipped,
	})
}
