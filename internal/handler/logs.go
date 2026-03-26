package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/handler/middleware"
)

type LogsHandler struct {
	store *middleware.LogStore
}

func NewLogsHandler(store *middleware.LogStore) *LogsHandler {
	return &LogsHandler{store: store}
}

// GET /internal/logs — all recent logs
func (h *LogsHandler) All(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.All())
}

// GET /internal/logs/errors — only errors
func (h *LogsHandler) Errors(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.Errors())
}
