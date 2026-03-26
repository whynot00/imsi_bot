package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/repo"
)

type IPHandler struct {
	ips *repo.IPRepo
}

func NewIPHandler(ips *repo.IPRepo) *IPHandler {
	return &IPHandler{ips: ips}
}

type createIPRequest struct {
	IP string `json:"ip" binding:"required"`
}

// POST /internal/ips
func (h *IPHandler) Create(c *gin.Context) {
	var req createIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip is required"})
		return
	}

	created, err := h.ips.Create(c.Request.Context(), req.IP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !created {
		c.JSON(http.StatusConflict, gin.H{"error": "ip already exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// DELETE /internal/ips/:id
func (h *IPHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	deleted, err := h.ips.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "ip not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /internal/ips
func (h *IPHandler) List(c *gin.Context) {
	ips, err := h.ips.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ips)
}
