package simulator

// SensorDef describes a sensor to register in the backend.
type SensorDef struct {
	Name     string
	Type     string
	Location string
}

// DefaultSensors defines the standard set of simulation sensors across 3 rooms.
var DefaultSensors = []SensorDef{
	{Name: "temp_living_room", Type: "temperature", Location: "living_room"},
	{Name: "humidity_living_room", Type: "humidity", Location: "living_room"},
	{Name: "light_living_room", Type: "light", Location: "living_room"},
	{Name: "motion_living_room", Type: "motion", Location: "living_room"},
	{Name: "light_kitchen", Type: "light", Location: "kitchen"},
	{Name: "motion_kitchen", Type: "motion", Location: "kitchen"},
	{Name: "light_bedroom", Type: "light", Location: "bedroom"},
	{Name: "motion_bedroom", Type: "motion", Location: "bedroom"},
}
