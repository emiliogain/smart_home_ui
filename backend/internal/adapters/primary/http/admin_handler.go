package http

import (
	_ "embed"
	"net/http"
	"strconv"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/adapters/secondary/fusion"
	"github.com/emiliogain/smart-home-backend/internal/app"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/gin-gonic/gin"
)

//go:embed admin.html
var adminHTML string

// AdminHandler provides REST endpoints for controlling the embedded simulator,
// switching fusion models, and serving the admin panel HTML page.
type AdminHandler struct {
	sim   *simulator.Engine
	svc   *app.SensorService
	multi *fusion.MultiPredictor
}

// NewAdminHandler creates an admin handler. sim and multi may be nil when disabled.
func NewAdminHandler(sim *simulator.Engine, svc *app.SensorService, multi *fusion.MultiPredictor) *AdminHandler {
	return &AdminHandler{sim: sim, svc: svc, multi: multi}
}

// RegisterAPIRoutes mounts admin API endpoints under the given group.
func (h *AdminHandler) RegisterAPIRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	{
		// Simulator control
		sim := admin.Group("/simulator")
		sim.GET("/status", h.GetStatus)
		sim.POST("/scenario", h.SetScenario)
		sim.POST("/control", h.Control)
		sim.POST("/interval", h.SetInterval)
		sim.POST("/inject", h.Inject)

		// Fusion model control
		f := admin.Group("/fusion")
		f.GET("/models", h.GetModels)
		f.POST("/model", h.SetModel)
		f.GET("/comparison", h.GetComparison)
	}
}

// ServePage returns the embedded admin HTML page.
func (h *AdminHandler) ServePage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(adminHTML))
}

// ── Simulator endpoints ───────────────────────────────────────────────────────

// GetStatus returns the simulator state plus the current fusion context.
func (h *AdminHandler) GetStatus(c *gin.Context) {
	resp := gin.H{"simulatorEnabled": h.sim != nil}

	if h.sim != nil {
		s := h.sim.GetStatus()
		resp["running"] = s.Running
		resp["paused"] = s.Paused
		resp["currentScenario"] = s.CurrentScenario
		resp["intervalMs"] = s.IntervalMs
		resp["availableScenarios"] = s.AvailableScenarios
		resp["lastTick"] = s.LastTick
		resp["tickCount"] = s.TickCount
	}

	if ctx := h.svc.GetCurrentContext(); ctx != nil {
		resp["currentContext"] = ctx.CurrentContext
		resp["confidence"] = ctx.Confidence
		resp["lastUpdated"] = ctx.LastUpdated
		resp["sensorSnapshot"] = ctx.SensorSnapshot
	}

	c.JSON(http.StatusOK, resp)
}

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
		if h.sim.GetStatus().Paused {
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
	c.JSON(http.StatusOK, gin.H{"sensor": req.SensorName, "value": req.Value, "unit": req.Unit})
}

// ── Fusion model endpoints ────────────────────────────────────────────────────

// GetModels returns the available models and the currently active one.
func (h *AdminHandler) GetModels(c *gin.Context) {
	if h.multi == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fusion multi-predictor not available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"models": h.multi.GetModelNames(),
		"active": h.multi.GetActiveModel(),
	})
}

// SetModel switches the active fusion model.
func (h *AdminHandler) SetModel(c *gin.Context) {
	if h.multi == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fusion multi-predictor not available"})
		return
	}
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.multi.SetActiveModel(req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"active": req.Model})
}

// GetComparison returns recent side-by-side predictions and aggregate stats.
// Query param: n (default 20, max 100).
func (h *AdminHandler) GetComparison(c *gin.Context) {
	if h.multi == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fusion multi-predictor not available"})
		return
	}
	n := 20
	if raw := c.Query("n"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			n = parsed
		}
	}
	c.JSON(http.StatusOK, h.multi.GetComparison(n))
}
