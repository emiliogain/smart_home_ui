package client

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
	body, _ := json.Marshal(map[string]interface{}{
		"value": value,
		"unit":  unit,
	})

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
