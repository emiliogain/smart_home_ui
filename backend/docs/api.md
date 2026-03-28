# API Documentation

TODO: Document the REST API endpoints

## Authentication
TODO: Describe authentication mechanism

## Endpoints

### Sensors
- GET /api/v1/sensors - List all sensors
- POST /api/v1/sensors - Create new sensor
- GET /api/v1/sensors/{id} - Get sensor by ID
- PUT /api/v1/sensors/{id} - Update sensor
- DELETE /api/v1/sensors/{id} - Delete sensor
- POST /api/v1/sensors/{id}/data - Submit sensor data

### Devices  
- GET /api/v1/devices - List all devices
- POST /api/v1/devices - Create new device
- GET /api/v1/devices/{id} - Get device by ID
- PUT /api/v1/devices/{id} - Update device
- DELETE /api/v1/devices/{id} - Delete device
- POST /api/v1/devices/{id}/command - Send command to device

### Real-time Data
- GET /api/v1/ws - WebSocket connection for real-time updates