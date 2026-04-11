package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/app"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SensorHandler handles HTTP requests for sensor operations.
type SensorHandler struct {
	svc *app.SensorService
}

// NewSensorHandler creates a new sensor HTTP handler.
func NewSensorHandler(svc *app.SensorService) *SensorHandler {
	return &SensorHandler{svc: svc}
}

// RegisterRoutes wires all sensor endpoints onto the given router group.
func (h *SensorHandler) RegisterRoutes(rg *gin.RouterGroup) {
	sensors := rg.Group("/sensors")
	{
		sensors.POST("", h.CreateSensor)
		sensors.GET("", h.ListSensors)
		sensors.GET("/:id", h.GetSensor)
		sensors.POST("/:id/readings", h.SubmitReading)
		sensors.GET("/:id/readings", h.GetReadings)
	}
}

// --- request / response DTOs (private to this file) ---

type createSensorRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Location string `json:"location" binding:"required"`
}

type submitReadingRequest struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type sensorResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Location  string    `json:"location"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type readingResponse struct {
	ID        string    `json:"id"`
	SensorID  string    `json:"sensor_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

// --- handlers ---

func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var req createSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s := sensor.Sensor{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Type:     sensor.SensorType(req.Type),
		Location: req.Location,
	}

	if err := h.svc.CreateSensor(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toSensorResponse(s))
}

func (h *SensorHandler) GetSensor(c *gin.Context) {
	s, err := h.svc.GetSensor(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toSensorResponse(*s))
}

func (h *SensorHandler) ListSensors(c *gin.Context) {
	sensors, err := h.svc.ListSensors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]sensorResponse, len(sensors))
	for i, s := range sensors {
		out[i] = toSensorResponse(s)
	}
	c.JSON(http.StatusOK, gin.H{"sensors": out})
}

func (h *SensorHandler) SubmitReading(c *gin.Context) {
	var req submitReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r := sensor.Reading{
		ID:       uuid.NewString(),
		SensorID: c.Param("id"),
		Value:    req.Value,
		Unit:     req.Unit,
	}

	result, err := h.svc.SaveReading(c.Request.Context(), r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"reading": toReadingResponse(r),
		"fusion":  result,
	})
}

func (h *SensorHandler) GetReadings(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	readings, err := h.svc.GetLatestReadings(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]readingResponse, len(readings))
	for i, r := range readings {
		out[i] = toReadingResponse(r)
	}
	c.JSON(http.StatusOK, gin.H{"readings": out})
}

// --- converters ---

func toSensorResponse(s sensor.Sensor) sensorResponse {
	return sensorResponse{
		ID: s.ID, Name: s.Name, Type: string(s.Type),
		Location: s.Location, Status: s.Status,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

func toReadingResponse(r sensor.Reading) readingResponse {
	return readingResponse{
		ID: r.ID, SensorID: r.SensorID,
		Value: r.Value, Unit: r.Unit, Timestamp: r.Timestamp,
	}
}
