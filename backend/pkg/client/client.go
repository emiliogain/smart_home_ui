package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the smart home backend REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a client pointing at the given backend base URL (e.g. "http://localhost:8080").
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SensorResponse is the JSON returned by POST /api/v1/sensors.
type SensorResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayLabel string `json:"display_label"`
	Type         string `json:"type"`
	Location     string `json:"location"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// RegisterSensor creates or updates a sensor in the backend.
func (c *Client) RegisterSensor(name, sensorType, location string) (*SensorResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"name":     name,
		"type":     sensorType,
		"location": location,
	})

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/sensors",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("register sensor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("register sensor: unexpected status %d", resp.StatusCode)
	}

	var sr SensorResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode sensor response: %w", err)
	}
	return &sr, nil
}

// SubmitReading sends a single reading for the given sensor.
func (c *Client) SubmitReading(sensorID string, value float64, unit string) error {
	return c.SubmitReadingAt(sensorID, value, unit, time.Time{})
}

// SubmitReadingAt sends a reading with an optional timestamp (zero = server assigns "now").
func (c *Client) SubmitReadingAt(sensorID string, value float64, unit string, at time.Time) error {
	payload := map[string]interface{}{
		"value": value,
		"unit":  unit,
	}
	if !at.IsZero() {
		payload["timestamp"] = at.UTC().Format(time.RFC3339Nano)
	}
	body, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v1/sensors/%s/readings", c.baseURL, sensorID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("submit reading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("submit reading: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// BatchReading is one row for SubmitReadingsBatch.
type BatchReading struct {
	SensorID  string
	Value     float64
	Unit      string
	Timestamp time.Time
}

// SubmitReadingsBatch posts multiple readings and runs fusion once.
func (c *Client) SubmitReadingsBatch(readings []BatchReading) error {
	type item struct {
		SensorID  string  `json:"sensor_id"`
		Value     float64 `json:"value"`
		Unit      string  `json:"unit"`
		Timestamp *string `json:"timestamp,omitempty"`
	}
	items := make([]item, 0, len(readings))
	for _, r := range readings {
		it := item{SensorID: r.SensorID, Value: r.Value, Unit: r.Unit}
		if !r.Timestamp.IsZero() {
			s := r.Timestamp.UTC().Format(time.RFC3339Nano)
			it.Timestamp = &s
		}
		items = append(items, it)
	}
	body, _ := json.Marshal(map[string]interface{}{"readings": items})

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/sensors/readings/batch",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("submit readings batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit readings batch: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// ResetDB truncates sensor_readings and sensors via POST /api/v1/admin/reset.
func (c *Client) ResetDB() error {
	resp, err := c.httpClient.Post(c.baseURL+"/api/admin/reset", "application/json", nil)
	if err != nil {
		return fmt.Errorf("reset db: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reset db: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// ListSensors returns all sensors from GET /api/v1/sensors.
func (c *Client) ListSensors() ([]SensorResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/sensors")
	if err != nil {
		return nil, fmt.Errorf("list sensors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list sensors: unexpected status %d", resp.StatusCode)
	}

	var wrapper struct {
		Sensors []SensorResponse `json:"sensors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decode sensors: %w", err)
	}
	return wrapper.Sensors, nil
}
