// Package http contains HTTP adapters (primary adapters)
package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/primary"
	"github.com/gin-gonic/gin"
)

// SensorHandler handles HTTP requests for sensor operations
type SensorHandler struct {
	sensorService primary.SensorService
}

// NewSensorHandler creates a new sensor HTTP handler
func NewSensorHandler(sensorService primary.SensorService) *SensorHandler {
	return &SensorHandler{
		sensorService: sensorService,
	}
}

// CreateSensor handles POST /api/v1/sensors
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var req primary.CreateSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensorEntity, err := h.sensorService.CreateSensor(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sensorToResponse(sensorEntity))
}

// GetSensor handles GET /api/v1/sensors/:id
func (h *SensorHandler) GetSensor(c *gin.Context) {
	id := c.Param("id")

	sensorEntity, err := h.sensorService.GetSensor(c.Request.Context(), id)
	if err != nil {
		if err == sensor.ErrSensorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sensorToResponse(sensorEntity))
}

// ListSensors handles GET /api/v1/sensors
func (h *SensorHandler) ListSensors(c *gin.Context) {
	// Parse query parameters
	sensorType := c.Query("type")
	location := c.Query("location")

	var sensors []*sensor.Sensor
	var err error

	switch {
	case sensorType != "":
		sensors, err = h.sensorService.GetSensorsByType(c.Request.Context(), sensor.SensorType(sensorType))
	case location != "":
		sensors, err = h.sensorService.GetSensorsByLocation(c.Request.Context(), location)
	default:
		sensors, err = h.sensorService.GetAllSensors(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]SensorResponse, len(sensors))
	for i, s := range sensors {
		response[i] = sensorToResponse(s)
	}

	c.JSON(http.StatusOK, gin.H{"sensors": response})
}

// UpdateSensor handles PUT /api/v1/sensors/:id
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id := c.Param("id")

	var req primary.UpdateSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensorEntity, err := h.sensorService.UpdateSensor(c.Request.Context(), id, req)
	if err != nil {
		if err == sensor.ErrSensorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sensorToResponse(sensorEntity))
}

// DeleteSensor handles DELETE /api/v1/sensors/:id
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id := c.Param("id")

	err := h.sensorService.DeleteSensor(c.Request.Context(), id)
	if err != nil {
		if err == sensor.ErrSensorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SubmitSensorData handles POST /api/v1/sensors/:id/data
func (h *SensorHandler) SubmitSensorData(c *gin.Context) {
	sensorID := c.Param("id")

	var req primary.SubmitSensorDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.sensorService.SubmitSensorData(c.Request.Context(), sensorID, req)
	if err != nil {
		if err == sensor.ErrSensorNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dataToResponse(data))
}

// GetSensorData handles GET /api/v1/sensors/:id/data
func (h *SensorHandler) GetSensorData(c *gin.Context) {
	sensorID := c.Param("id")

	// Parse time range parameters
	from, err := parseTime(c.Query("from"), time.Now().Add(-24*time.Hour))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from time"})
		return
	}

	to, err := parseTime(c.Query("to"), time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to time"})
		return
	}

	// Check for aggregation request
	if interval := c.Query("interval"); interval != "" {
		duration, err := time.ParseDuration(interval)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interval"})
			return
		}

		aggregated, err := h.sensorService.GetAggregatedData(c.Request.Context(), sensorID, from, to, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": aggregated})
		return
	}

	// Get raw data
	data, err := h.sensorService.GetSensorData(c.Request.Context(), sensorID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]SensorDataResponse, len(data))
	for i, d := range data {
		response[i] = dataToResponse(d)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetSensorReadings handles GET /api/v1/sensors/:id/readings
func (h *SensorHandler) GetSensorReadings(c *gin.Context) {
	sensorID := c.Param("id")

	limit := 100 // default
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	readings, err := h.sensorService.GetSensorReadings(c.Request.Context(), sensorID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]SensorReadingResponse, len(readings))
	for i, r := range readings {
		response[i] = readingToResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{"readings": response})
}

// Response DTOs

type SensorResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        sensor.SensorType `json:"type"`
	Location    string            `json:"location"`
	Description string            `json:"description,omitempty"`
	Status      sensor.Status     `json:"status"`
	LastSeen    *time.Time        `json:"last_seen,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Config      sensor.Config     `json:"config"`
}

type SensorDataResponse struct {
	ID        string                 `json:"id"`
	SensorID  string                 `json:"sensor_id"`
	Timestamp time.Time              `json:"timestamp"`
	Values    map[string]interface{} `json:"values"`
	Quality   float64                `json:"quality"`
	CreatedAt time.Time              `json:"created_at"`
}

type SensorReadingResponse struct {
	SensorID  string            `json:"sensor_id"`
	Type      sensor.SensorType `json:"type"`
	Location  string            `json:"location"`
	Timestamp time.Time         `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Unit      string            `json:"unit,omitempty"`
	Quality   float64           `json:"quality"`
}

// Conversion helpers

func sensorToResponse(s *sensor.Sensor) SensorResponse {
	return SensorResponse{
		ID:          s.ID(),
		Name:        s.Name(),
		Type:        s.Type(),
		Location:    s.Location(),
		Description: s.Description(),
		Status:      s.Status(),
		LastSeen:    s.LastSeen(),
		CreatedAt:   s.CreatedAt(),
		UpdatedAt:   s.UpdatedAt(),
		Config:      s.Config(),
	}
}

func dataToResponse(d *sensor.Data) SensorDataResponse {
	return SensorDataResponse{
		ID:        d.ID(),
		SensorID:  d.SensorID(),
		Timestamp: d.Timestamp(),
		Values:    d.Values(),
		Quality:   d.Quality(),
		CreatedAt: d.CreatedAt(),
	}
}

func readingToResponse(r *sensor.Reading) SensorReadingResponse {
	return SensorReadingResponse{
		SensorID:  r.SensorID,
		Type:      r.Type,
		Location:  r.Location,
		Timestamp: r.Timestamp,
		Value:     r.Value,
		Unit:      r.Unit,
		Quality:   r.Quality,
	}
}

// Helper functions

func parseTime(timeStr string, defaultTime time.Time) (time.Time, error) {
	if timeStr == "" {
		return defaultTime, nil
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}
