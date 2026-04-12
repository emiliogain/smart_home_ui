# Smart Home System - Full Stack Application

A comprehensive smart home system for collecting sensor data, controlling devices, and providing a dynamic user interface. Built as a diploma thesis project.

## 🏗️ Project Architecture

```
smart-home-ui/
├── 🔧 backend/                   # Go Backend API
│   ├── cmd/                      # Application entry points
│   ├── internal/                 # Private application code
│   ├── pkg/                      # Public libraries
│   ├── api/                      # API specifications
│   ├── config/                   # Configuration files (config.yaml)
│   └── docs/                     # Backend documentation
│
├── 🎨 frontend/                  # Frontend Application  
│   ├── src/                      # Source code
│   ├── public/                   # Static assets
│   └── styles/                   # Styling
│
├── 🐳 docker-compose.yml         # Full stack orchestration
├── 📝 README.md                  # This file
└── ⚙️ Makefile                   # Build automation
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.22+** (for backend)
- **Node.js 18+** (for frontend) 
- **Docker & Docker Compose** (recommended)
- **PostgreSQL 15+** & **Redis 7+** (if running locally)

### Option 1: Docker (Recommended)
```bash
# Start the entire stack
make start
# or
docker-compose up --build

# Access the application
# Backend API: http://localhost:8080
# Frontend: http://localhost:3000
```

### Option 2: Local Development
```bash
# Setup development environment
make setup

# Start databases
make db-up

# In one terminal - start backend
make run-backend

# In another terminal - start frontend  
make run-frontend
```

## 🎯 System Components

### Backend (Go)
- **API Server**: RESTful API with real-time WebSocket support
- **Sensor Management**: IoT sensor data collection and processing
- **Device Control**: Smart device command execution and monitoring
- **Data Fusion**: Multi-sensor analytics and pattern recognition
- **Database**: PostgreSQL for persistence, Redis for caching

See `backend/README.md` for detailed backend documentation.

### Frontend (TypeScript/JavaScript)
- **Dashboard**: Real-time sensor data visualization
- **Device Control**: Interactive device management interface
- **Analytics**: Historical data analysis and insights
- **Responsive Design**: Mobile-friendly user interface

See `frontend/README.md` for frontend setup and development.

## 🛠️ Development Workflow

### Available Commands
```bash
make help              # Show all available commands

# Development
make setup             # Initial environment setup
make run-backend       # Start backend development server
make run-frontend      # Start frontend development server

# Building & Testing
make build             # Build both applications
make test              # Run all tests
make lint              # Run linters

# Docker Operations  
make start             # Start full stack
make stop              # Stop all services
make restart           # Restart services
make docker-logs       # View logs

# Database
make db-up             # Start only databases
make db-down           # Stop databases
```

### Technology Stack

**Backend:**
- **Language**: Go 1.22+
- **Framework**: Gin (HTTP) + Gorilla WebSocket
- **Database**: PostgreSQL + Redis
- **Architecture**: Clean Architecture with DI
- **Testing**: Go standard testing + testify
- **Deployment**: Docker containers

**Frontend:** (To be selected)
- **React + TypeScript** (Recommended for complex UIs)
- **Vue.js + TypeScript** (Alternative option)
- **Svelte** (Lightweight option)
- **Build Tools**: Vite/Webpack, ESLint, Prettier

**Infrastructure:**
- **Containerization**: Docker & Docker Compose
- **Databases**: PostgreSQL 15, Redis 7
- **Monitoring**: Structured logging, health checks
- **API**: OpenAPI 3.0 specification

## 📋 Development Roadmap

### ✅ Phase 1: Project Structure
- [x] Monorepo setup with clear separation
- [x] Backend Go project structure
- [x] Frontend directory preparation
- [x] Docker orchestration
- [x] Build automation
### 🔄 Phase 2: Backend Core (In Progress)
- [ ] Configuration management
- [ ] Database models and migrations
- [ ] REST API implementation
- [ ] WebSocket real-time communication
- [ ] Authentication & authorization

### 🎯 Phase 3: Sensor & Device Management
- [ ] Sensor registration and data collection
- [ ] Device discovery and control
- [ ] Data validation and quality scoring
- [ ] Time-series data storage

### 📊 Phase 4: Data Processing & Analytics
- [ ] Data fusion algorithms
- [ ] Pattern recognition
- [ ] Predictive analytics
- [ ] Alert and notification system

### 🎨 Phase 5: Frontend Development
- [ ] Framework selection and setup
- [ ] Component library
- [ ] Real-time dashboard
- [ ] Device control interface
- [ ] Data visualization

### 🚀 Phase 6: Production Ready
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Comprehensive testing
- [ ] Deployment automation
- [ ] Documentation completion

## 🔐 Security Considerations

- JWT-based authentication
- CORS configuration for frontend
- Input validation and sanitization
- Rate limiting and request throttling
- Secure database connections
- Environment-based configuration

## 📊 Monitoring & Observability

- Structured logging with Zap
- Health check endpoints
- Metrics collection (Prometheus compatible)
- Real-time performance monitoring
- Error tracking and alerting

## 🤝 Contributing & Development Guidelines

1. **Code Quality**: Follow language-specific best practices
2. **Testing**: Maintain high test coverage
3. **Documentation**: Keep docs updated with changes
4. **Git Workflow**: Use conventional commit messages
5. **Architecture**: Maintain separation of concerns

## 📄 License

This project is developed as part of a diploma thesis and is for educational purposes.

---

**Next Steps:**
1. Choose your frontend framework in `frontend/`
2. Implement backend API endpoints in `backend/`
3. Set up your development databases
4. Start building your smart home features!

For detailed setup instructions, see the README files in `backend/` and `frontend/` directories.
