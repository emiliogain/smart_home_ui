package sensor

import (
	"errors"
	"time"
)

// Service contains the core business logic for sensors
// It has NO external dependencies - pure domain logic
type Service struct {
	// No infrastructure dependencies here!
}

// NewService creates a new sensor domain service
func NewService() *Service {
	return &Service{}
}

// ValidateConfiguration validates sensor configuration
func (s *Service) ValidateConfiguration(config Config) error {
	if config.ReportingInterval <= 0 {
		return errors.New("reporting interval must be positive")
	}

	if config.ReportingInterval < time.Second {
		return errors.New("reporting interval too short, minimum 1 second")
	}

	if config.ReportingInterval > 24*time.Hour {
		return errors.New("reporting interval too long, maximum 24 hours")
	}

	return nil
}

// ValidateDataValues validates sensor data values based on sensor type
func (s *Service) ValidateDataValues(sensorType SensorType, values map[string]interface{}) error {
	switch sensorType {
	case TypeTemperature:
		return s.validateTemperatureData(values)
	case TypeHumidity:
		return s.validateHumidityData(values)
	case TypePressure:
		return s.validatePressureData(values)
	case TypeMotion:
		return s.validateMotionData(values)
	case TypeLight:
		return s.validateLightData(values)
	case TypeDoor, TypeWindow:
		return s.validateContactData(values)
	case TypeSmoke, TypeGas:
		return s.validateGasData(values)
	case TypePower:
		return s.validatePowerData(values)
	default:
		return ErrInvalidType
	}
}

// CalculateDataQuality calculates the quality score for sensor data
func (s *Service) CalculateDataQuality(data *Data, sensor *Sensor) float64 {
	quality := 1.0

	// Reduce quality based on data age
	age := time.Since(data.Timestamp())
	if age > sensor.Config().ReportingInterval*2 {
		quality *= 0.8 // Old data penalty
	}

	// Reduce quality for incomplete data
	expectedFields := s.getExpectedFields(sensor.Type())
	actualFields := len(data.Values())
	if actualFields < expectedFields {
		quality *= float64(actualFields) / float64(expectedFields)
	}

	// Sensor-specific quality checks
	quality *= s.calculateTypeSpecificQuality(sensor.Type(), data.Values())

	return quality
}

// ShouldTriggerAlert determines if sensor data should trigger an alert
func (s *Service) ShouldTriggerAlert(sensor *Sensor, data *Data) (bool, string) {
	thresholds := sensor.Config().Thresholds
	if thresholds == nil {
		return false, ""
	}

	for key, value := range data.Values() {
		if threshold, exists := thresholds[key+"_max"]; exists {
			if numValue, ok := value.(float64); ok {
				if thresholdValue, ok := threshold.(float64); ok {
					if numValue > thresholdValue {
						return true, "Value exceeded maximum threshold"
					}
				}
			}
		}

		if threshold, exists := thresholds[key+"_min"]; exists {
			if numValue, ok := value.(float64); ok {
				if thresholdValue, ok := threshold.(float64); ok {
					if numValue < thresholdValue {
						return true, "Value below minimum threshold"
					}
				}
			}
		}
	}

	return false, ""
}

// AggregateData performs statistical aggregation of sensor data
func (s *Service) AggregateData(dataPoints []*Data, interval time.Duration) map[string]interface{} {
	if len(dataPoints) == 0 {
		return nil
	}

	// Group data by time intervals
	groups := s.groupDataByInterval(dataPoints, interval)

	result := make(map[string]interface{})

	// Calculate aggregates for each group
	for timestamp, group := range groups {
		aggregates := s.calculateAggregates(group)
		result[timestamp.Format(time.RFC3339)] = aggregates
	}

	return result
}

// Private helper methods

func (s *Service) validateTemperatureData(values map[string]interface{}) error {
	temp, exists := values["temperature"]
	if !exists {
		return errors.New("temperature value required")
	}

	tempFloat, ok := temp.(float64)
	if !ok {
		return errors.New("temperature must be a number")
	}

	if tempFloat < -273.15 || tempFloat > 200 {
		return errors.New("temperature out of valid range")
	}

	return nil
}

func (s *Service) validateHumidityData(values map[string]interface{}) error {
	humidity, exists := values["humidity"]
	if !exists {
		return errors.New("humidity value required")
	}

	humidityFloat, ok := humidity.(float64)
	if !ok {
		return errors.New("humidity must be a number")
	}

	if humidityFloat < 0 || humidityFloat > 100 {
		return errors.New("humidity must be between 0 and 100")
	}

	return nil
}

func (s *Service) validatePressureData(values map[string]interface{}) error {
	pressure, exists := values["pressure"]
	if !exists {
		return errors.New("pressure value required")
	}

	pressureFloat, ok := pressure.(float64)
	if !ok {
		return errors.New("pressure must be a number")
	}

	if pressureFloat < 0 {
		return errors.New("pressure cannot be negative")
	}

	return nil
}

func (s *Service) validateMotionData(values map[string]interface{}) error {
	motion, exists := values["motion"]
	if !exists {
		return errors.New("motion value required")
	}

	_, ok := motion.(bool)
	if !ok {
		return errors.New("motion must be a boolean")
	}

	return nil
}

func (s *Service) validateLightData(values map[string]interface{}) error {
	light, exists := values["light_level"]
	if !exists {
		return errors.New("light_level value required")
	}

	lightFloat, ok := light.(float64)
	if !ok {
		return errors.New("light_level must be a number")
	}

	if lightFloat < 0 {
		return errors.New("light_level cannot be negative")
	}

	return nil
}

func (s *Service) validateContactData(values map[string]interface{}) error {
	contact, exists := values["open"]
	if !exists {
		return errors.New("open status required")
	}

	_, ok := contact.(bool)
	if !ok {
		return errors.New("open status must be a boolean")
	}

	return nil
}

func (s *Service) validateGasData(values map[string]interface{}) error {
	detected, exists := values["detected"]
	if !exists {
		return errors.New("detected status required")
	}

	_, ok := detected.(bool)
	if !ok {
		return errors.New("detected status must be a boolean")
	}

	return nil
}

func (s *Service) validatePowerData(values map[string]interface{}) error {
	power, exists := values["power"]
	if !exists {
		return errors.New("power value required")
	}

	powerFloat, ok := power.(float64)
	if !ok {
		return errors.New("power must be a number")
	}

	if powerFloat < 0 {
		return errors.New("power cannot be negative")
	}

	return nil
}

func (s *Service) getExpectedFields(sensorType SensorType) int {
	switch sensorType {
	case TypeTemperature:
		return 1 // temperature
	case TypeHumidity:
		return 1 // humidity
	case TypePressure:
		return 1 // pressure
	case TypeMotion:
		return 1 // motion
	case TypeLight:
		return 1 // light_level
	case TypeDoor, TypeWindow:
		return 1 // open
	case TypeSmoke, TypeGas:
		return 1 // detected
	case TypePower:
		return 1 // power
	default:
		return 1
	}
}

func (s *Service) calculateTypeSpecificQuality(sensorType SensorType, values map[string]interface{}) float64 {
	// Add type-specific quality calculations here
	// For now, return 1.0 (perfect quality)
	return 1.0
}

func (s *Service) groupDataByInterval(dataPoints []*Data, interval time.Duration) map[time.Time][]*Data {
	groups := make(map[time.Time][]*Data)

	for _, data := range dataPoints {
		// Round timestamp to interval boundary
		rounded := data.Timestamp().Truncate(interval)
		groups[rounded] = append(groups[rounded], data)
	}

	return groups
}

func (s *Service) calculateAggregates(dataPoints []*Data) map[string]interface{} {
	if len(dataPoints) == 0 {
		return nil
	}

	aggregates := make(map[string]interface{})
	aggregates["count"] = len(dataPoints)

	// Calculate averages for numeric values
	sums := make(map[string]float64)
	counts := make(map[string]int)

	for _, data := range dataPoints {
		for key, value := range data.Values() {
			if numValue, ok := value.(float64); ok {
				sums[key] += numValue
				counts[key]++
			}
		}
	}

	for key, sum := range sums {
		if count := counts[key]; count > 0 {
			aggregates[key+"_avg"] = sum / float64(count)
		}
	}

	return aggregates
}
