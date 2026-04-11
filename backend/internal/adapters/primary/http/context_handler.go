package http

import (
	"net/http"

	"github.com/emiliogain/smart-home-backend/internal/app"
	"github.com/gin-gonic/gin"
)

// ContextHandler serves the current fusion context to the frontend.
type ContextHandler struct {
	svc *app.SensorService
}

// NewContextHandler creates a context HTTP handler.
func NewContextHandler(svc *app.SensorService) *ContextHandler {
	return &ContextHandler{svc: svc}
}

// RegisterRoutes wires the context endpoints onto the given router group.
func (h *ContextHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/context/current", h.GetCurrentContext)
}

// GetCurrentContext returns the latest fusion result.
func (h *ContextHandler) GetCurrentContext(c *gin.Context) {
	ctx := h.svc.GetCurrentContext()
	if ctx == nil {
		c.JSON(http.StatusOK, gin.H{
			"currentContext": "UNKNOWN",
			"confidence":     0,
			"lastUpdated":    "",
			"sensorSnapshot": nil,
		})
		return
	}
	c.JSON(http.StatusOK, ctx)
}
