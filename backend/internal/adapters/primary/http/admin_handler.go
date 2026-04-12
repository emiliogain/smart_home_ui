package http

import (
	_ "embed"
	"net/http"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/app"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/gin-gonic/gin"
)

//go:embed admin.html
var adminHTML string

// AdminHandler provides REST endpoints for controlling the embedded simulator
// and serves the admin panel HTML page.
type AdminHandler struct {
	sim *simulator.Engine
	svc *app.SensorService
}

// NewAdminHandler creates an admin handler. sim may be nil if the simulator is disabled.
func NewAdminHandler(sim *simulator.Engine, svc *app.SensorService) *AdminHandler {
	return &AdminHandler{sim: sim, svc: svc}
}

// RegisterAPIRoutes mounts admin API endpoints under the given group.
func (h *AdminHandler) RegisterAPIRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin/simulator")
	{
		admin.GET("/status", h.GetStatus)
		admin.POST("/scenario", h.SetScenario)
		admin.POST("/control", h.Control)
		admin.POST("/interval", h.SetInterval)
		admin.POST("/inject", h.Inject)
	}
}

// ServePage returns the embedded admin HTML page.
func (h *AdminHandler) ServePage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(adminHTML))
}

// GetStatus returns the simulator state plus the current fusion context.
func (h *AdminHandler) GetStatus(c *gin.Context) {
	resp := gin.H{
		"simulatorEnabled": h.sim != nil,
	}

	if h.sim != nil {
		status := h.sim.GetStatus()
		resp["running"] = status.Running
		resp["paused"] = status.Paused
		resp["currentScenario"] = status.CurrentScenario
		resp["intervalMs"] = status.IntervalMs
		resp["availableScenarios"] = status.AvailableScenarios
		resp["lastTick"] = status.LastTick
		resp["tickCount"] = status.TickCount
	}

	if ctx := h.svc.GetCurrentContext(); ctx != nil {
		resp["currentContext"] = ctx.CurrentContext
		resp["confidence"] = ctx.Confidence
		resp["lastUpdated"] = ctx.LastUpdated
		resp["sensorSnapshot"] = ctx.SensorSnapshot
	}

	c.JSON(http.StatusOK, resp)
}

// SetScenario switches the active simulation scenario.
func (h *AdminHandler) SetScenario(c *gin.Context) {
	if h.sim == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "simulator is disabled"})
		return
	}

	var req struct {
		Scenario string `json:"scenario" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sim.SetScenario(req.Scenario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scenario": req.Scenario})
}

// Control handles pause/resume/toggle actions.
func (h *AdminHandler) Control(c *gin.Context) {
	if h.sim == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "simulator is disabled"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Action {
	case "pause":
		h.sim.Pause()
	case "resume":
		h.sim.Resume()
	case "toggle":
		status := h.sim.GetStatus()
		if status.Paused {
			h.sim.Resume()
		} else {
			h.sim.Pause()
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be pause, resume, or toggle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": req.Action})
}

// SetInterval changes the simulation tick rate.
func (h *AdminHandler) SetInterval(c *gin.Context) {
	if h.sim == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "simulator is disabled"})
		return
	}

	var req struct {
		Interval string `json:"interval" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := time.ParseDuration(req.Interval)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid duration: " + err.Error()})
		return
	}

	if err := h.sim.SetInterval(d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"intervalMs": d.Milliseconds()})
}

// Inject sends a one-off manual sensor reading.
func (h *AdminHandler) Inject(c *gin.Context) {
	if h.sim == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "simulator is disabled"})
		return
	}

	var req struct {
		SensorName string  `json:"sensorName" binding:"required"`
		Value      float64 `json:"value"`
		Unit       string  `json:"unit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sim.Inject(req.SensorName, req.Value, req.Unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sensor": req.SensorName,
		"value":  req.Value,
		"unit":   req.Unit,
	})
}
