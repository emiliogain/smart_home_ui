# Pure Hexagonal Architecture - Smart Home Backend

This backend follows **pure hexagonal architecture** (Ports & Adapters pattern) for maximum flexibility and testability.

## 🏗️ Architecture Overview

```
                    ┌─────────────────────────────────────────────┐
                    │              PRIMARY ADAPTERS              │
                    │         (Driving Adapters)                 │
                    └─────────────────────────────────────────────┘
                                          │
                    ┌─────────────────────────────────────────────┐
                    │               PRIMARY PORTS                 │
                    │              (Interfaces)                   │
                    └─────────────────────────────────────────────┘
                                          │
                    ┌─────────────────────────────────────────────┐
                    │            APPLICATION LAYER                │
                    │           (Orchestration)                   │
                    └─────────────────────────────────────────────┘
                                          │
                    ┌─────────────────────────────────────────────┐
                    │              DOMAIN LAYER                   │
                    │          (Pure Business Logic)              │
                    └─────────────────────────────────────────────┘
                                          │
                    ┌─────────────────────────────────────────────┐
                    │              SECONDARY PORTS                │
                    │              (Interfaces)                   │
                    └─────────────────────────────────────────────┘
                                          │
                    ┌─────────────────────────────────────────────┐
                    │             SECONDARY ADAPTERS              │
                    │          (Driven Adapters)                  │
                    └─────────────────────────────────────────────┘
```

## 📁 Directory Structure

```
internal/
├── domain/                    # 💎 CORE - Pure Business Logic
│   ├── sensor/               
│   │   ├── entity.go         # Sensor entity with business rules
│   │   ├── service.go        # Domain services (validation, calculations)  
│   │   └── errors.go         # Domain-specific errors
│   └── device/
│       ├── entity.go         # Device entity with business rules
│       ├── service.go        # Device domain services
│       └── errors.go         # Domain-specific errors
│
├── ports/                    # 🔌 INTERFACES - Contract Definitions
│   ├── primary/              # Incoming interfaces (what adapters call)
│   │   ├── sensor_service.go # SensorService interface
│   │   └── device_service.go # DeviceService interface  
│   └── secondary/            # Outgoing interfaces (what domain needs)
│       ├── sensor_repository.go  # Persistence contracts
│       └── device_repository.go  # Persistence contracts
│
├── app/                      # 🎭 ORCHESTRATION - Application Services
│   ├── sensor_service.go     # Implements primary.SensorService
│   └── device_service.go     # Implements primary.DeviceService
│
└── adapters/                 # 🔧 INFRASTRUCTURE - External World
    ├── primary/              # Driving adapters (initiate actions)
    │   ├── http/             # REST API handlers
    │   │   ├── sensor_handler.go
    │   │   └── device_handler.go
    │   └── cli/              # Command-line interface (future)
    │
    └── secondary/            # Driven adapters (called by domain)
        ├── database/         # Database implementations
        │   ├── sensor_repository.go
        │   └── device_repository.go  
        └── external/         # External services
            ├── event_publisher.go    # Message queue
            └── device_controller.go  # IoT device communication
```

## 🎯 Key Principles

### 1. **Dependency Inversion**
- **Domain** has NO dependencies on infrastructure
- **Application** depends only on domain and port interfaces  
- **Adapters** implement port interfaces and depend on infrastructure

### 2. **Pure Domain Logic**
```go
// ✅ GOOD: Pure business logic, no infrastructure dependencies
func (s *Service) CalculateDataQuality(data *Data, sensor *Sensor) float64 {
    quality := 1.0
    
    // Business rules only
    age := time.Since(data.Timestamp())
    if age > sensor.Config().ReportingInterval*2 {
        quality *= 0.8 // Old data penalty
    }
    
    return quality
}
```

### 3. **Port Interfaces Define Contracts**
```go
// Primary Port (incoming)
type SensorService interface {
    CreateSensor(ctx context.Context, req CreateSensorRequest) (*sensor.Sensor, error)
    GetSensor(ctx context.Context, id string) (*sensor.Sensor, error)
    // ...
}

// Secondary Port (outgoing)  
type SensorRepository interface {
    Save(ctx context.Context, sensor *sensor.Sensor) error
    FindByID(ctx context.Context, id string) (*sensor.Sensor, error)
    // ...
}
```

### 4. **Adapter Implementations**
```go
// HTTP Adapter (Primary)
type SensorHandler struct {
    sensorService primary.SensorService // Uses interface, not concrete type
}

// Database Adapter (Secondary)
type sensorRepository struct {
    db *sql.DB
}

func (r *sensorRepository) Save(ctx context.Context, s *sensor.Sensor) error {
    // Database-specific implementation
}
```

## 🔄 Data Flow

### Incoming Request (Create Sensor):
1. **HTTP Handler** receives REST request
2. **Handler** calls `primary.SensorService.CreateSensor()`
3. **App Service** orchestrates using domain services
4. **Domain Service** validates business rules
5. **App Service** calls `secondary.SensorRepository.Save()`
6. **Database Adapter** persists to PostgreSQL
7. **Event Publisher** notifies other services

### Domain Rule Example:
```go
// Domain enforces business rules
func (s *Sensor) UpdateStatus(status Status) error {
    if s.status == StatusError && status == StatusActive {
        // Business rule: Must reset error before going active
        return ErrMustResetErrorFirst
    }
    s.status = status
    return nil
}
```

## ✅ Benefits of This Architecture

### **Testability**
- Domain logic tested in isolation (no mocks needed)
- Application services tested with interface mocks
- Adapters tested independently

### **Flexibility**  
- Swap databases without touching business logic
- Add new interfaces (GraphQL, gRPC) easily
- Change external services without domain changes

### **Maintainability**
- Clear separation of concerns
- Dependencies point inward
- Business logic is protected and explicit

### **Technology Independence**
- Domain is pure Go (no frameworks)
- Can run without HTTP, database, etc.
- Framework choices don't affect business logic

## 🧪 Testing Strategy

### **Unit Tests: Domain Layer**
```go
func TestSensorValidation(t *testing.T) {
    sensor, err := NewSensor("id", "name", TypeTemperature, "kitchen")
    assert.NoError(t, err)
    assert.Equal(t, StatusActive, sensor.Status())
}
```

### **Integration Tests: Application Layer**  
```go
func TestCreateSensor(t *testing.T) {
    // Use mock repositories
    mockRepo := &MockSensorRepository{}
    mockPublisher := &MockEventPublisher{}
    
    service := NewSensorService(domainService, mockRepo, mockPublisher)
    
    sensor, err := service.CreateSensor(ctx, req)
    assert.NoError(t, err)
    
    // Verify interactions
    mockRepo.AssertCalled(t, "Save", mock.Anything)
}
```

### **Contract Tests: Adapters**
- Test that adapters correctly implement port interfaces
- Verify database adapters work with real database
- Test HTTP handlers with real HTTP requests

## 🚀 Deployment Benefits

- **Microservice Ready**: Easy to extract services  
- **Cloud Native**: Adapters for different cloud providers
- **Observability**: Easy to add monitoring at adapter boundaries
- **Scalability**: Stateless application layer

---

**The hexagonal architecture ensures your business logic stays clean, testable, and independent of infrastructure concerns.**