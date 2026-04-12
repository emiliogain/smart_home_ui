package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	httphandler "github.com/emiliogain/smart-home-backend/internal/adapters/primary/http"
	ws "github.com/emiliogain/smart-home-backend/internal/adapters/primary/websocket"
	"github.com/emiliogain/smart-home-backend/internal/adapters/secondary/database"
	"github.com/emiliogain/smart-home-backend/internal/adapters/secondary/fusion"
	"github.com/emiliogain/smart-home-backend/internal/app"
	"github.com/emiliogain/smart-home-backend/internal/config"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 2. Initialize structured logger
	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer logger.Sync()

	// 3. Apply embedded SQL migrations (local `make run-backend` and Docker entrypoint)
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database_url is required")
	}
	logger.Info("running database migrations")
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	logger.Info("database migrations applied")

	// 4. Connect to PostgreSQL
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL")

	// 5. Create WebSocket hub
	hub, err := ws.NewHub()
	if err != nil {
		return fmt.Errorf("create websocket hub: %w", err)
	}
	defer hub.Close()

	// 6. Create adapters & service
	repo := database.NewSensorRepository(pool)
	predictor := fusion.NewRuleBasedPredictor(fusion.DefaultThresholds())
	svc := app.NewSensorService(repo, predictor, hub)

	// 7. Start embedded simulator (if enabled)
	var sim *simulator.Engine
	if cfg.SimulatorEnabled {
		sensorIDs, regErr := registerSimulationSensors(ctx, svc)
		if regErr != nil {
			return fmt.Errorf("register simulation sensors: %w", regErr)
		}

		interval, _ := time.ParseDuration(cfg.SimulatorInterval)
		if interval == 0 {
			interval = 5 * time.Second
		}

		sim = simulator.NewEngine(svc, sensorIDs,
			simulator.WithInterval(interval),
			simulator.WithScenario(cfg.SimulatorScenario),
		)
		go sim.Start(ctx)
		defer sim.Stop()
		logger.Info("embedded simulator started",
			zap.String("scenario", cfg.SimulatorScenario),
			zap.Duration("interval", interval),
		)
	}

	// 8. Setup Gin router with CORS
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/api/admin/simulator/status", "/socket.io/"},
	}))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000", "http://localhost:5173", "http://localhost:8080",
			"http://127.0.0.1:3000", "http://127.0.0.1:5173", "http://127.0.0.1:8080",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// 9. Mount Socket.IO handler
	router.GET("/socket.io/*any", gin.WrapH(hub.Handler()))
	router.POST("/socket.io/*any", gin.WrapH(hub.Handler()))

	// 10. Register REST API routes
	v1 := router.Group("/api/v1")
	sensorHandler := httphandler.NewSensorHandler(svc)
	sensorHandler.RegisterRoutes(v1)

	api := router.Group("/api")
	sensorHandler.RegisterRoutes(api)
	contextHandler := httphandler.NewContextHandler(svc)
	contextHandler.RegisterRoutes(api)

	// 11. Register admin panel routes
	adminHandler := httphandler.NewAdminHandler(sim, svc)
	adminHandler.RegisterAPIRoutes(api)
	router.GET("/admin", adminHandler.ServePage)

	// 12. Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", zap.Int("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
		}
	}()

	// 13. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}

// registerSimulationSensors creates the default sensors and returns a name→ID map.
func registerSimulationSensors(ctx context.Context, svc *app.SensorService) (map[string]string, error) {
	ids := make(map[string]string)
	for _, def := range simulator.DefaultSensors {
		id := simulator.DeterministicSensorID(def.Name)
		s := sensor.Sensor{
			ID:       id,
			Name:     def.Name,
			Type:     sensor.SensorType(def.Type),
			Location: def.Location,
		}
		if err := svc.CreateSensor(ctx, s); err != nil {
			return nil, fmt.Errorf("create sensor %s: %w", def.Name, err)
		}
		ids[def.Name] = id
	}
	return ids, nil
}

func newLogger(level string) (*zap.Logger, error) {
	var zapLevel zap.AtomicLevel
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zap.NewAtomicLevel()
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zapLevel
	return cfg.Build()
}
