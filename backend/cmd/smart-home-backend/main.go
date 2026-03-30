package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// TODO: uncomment as you implement each package
	// "github.com/emiliogain/smart-home-backend/internal/config"
	// "github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	// "github.com/emiliogain/smart-home-backend/internal/domain/device"
	// "github.com/emiliogain/smart-home-backend/internal/app"
	// "github.com/emiliogain/smart-home-backend/internal/adapters/primary/http"
	// "github.com/emiliogain/smart-home-backend/internal/adapters/secondary/database"
	// "github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/emiliogain/smart-home-backend/internal/adapters/secondary/database"
	"github.com/emiliogain/smart-home-backend/internal/config"
	"golang.org/x/xerrors"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run is the real entrypoint -- returns an error instead of calling os.Exit directly.
// This pattern makes the function testable and ensures deferred calls execute.
func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		return xerrors.Errorf("loading config: %w", err)
	}

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		return xerrors.Errorf("initializing logger: %w", err)
	}
	defer logger.Sync()

	// ---------------------------------------------------------------
	// 3. Connect to PostgreSQL
	// ---------------------------------------------------------------
	if cfg.DatabaseURL == "" {
		return xerrors.Errorf("database URL is empty")
	}

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		return xerrors.Errorf("database migrations: %w", err)
	}
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return xerrors.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL")

	// ---------------------------------------------------------------
	// 4. Create secondary adapters (repositories)
	// ---------------------------------------------------------------
	_ = database.NewSensorRepository(pool)
	_ = database.NewDeviceRepository(pool)

	// ---------------------------------------------------------------
	// 5. Create domain services
	// ---------------------------------------------------------------
	// TODO: sensorDomainSvc := sensor.NewService()
	// TODO: deviceDomainSvc := device.NewService()

	// ---------------------------------------------------------------
	// 6. Create application services
	// ---------------------------------------------------------------
	// TODO: sensorAppSvc := app.NewSensorService(sensorDomainSvc, sensorRepo, eventPublisher)
	// TODO: deviceAppSvc := app.NewDeviceService(deviceDomainSvc, deviceRepo, deviceController, eventPublisher)

	// ---------------------------------------------------------------
	// 7. Create primary adapters (HTTP handlers)
	// ---------------------------------------------------------------
	// TODO: sensorHandler := httpAdapter.NewSensorHandler(sensorAppSvc)
	// TODO: deviceHandler := httpAdapter.NewDeviceHandler(deviceAppSvc)
	// TODO: wsHandler     := httpAdapter.NewWebSocketHandler(...)

	// ---------------------------------------------------------------
	// 8. Setup Gin router with middleware
	// ---------------------------------------------------------------
	// TODO: router := httpAdapter.NewRouter(sensorHandler, deviceHandler, wsHandler)

	// ---------------------------------------------------------------
	// 9. Start HTTP server
	// ---------------------------------------------------------------
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.ServerPort),
		// Handler: router, // TODO: uncomment once router is built
	}

	go func() {
		fmt.Printf("starting server on :%d\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	// ---------------------------------------------------------------
	// 10. Graceful shutdown on SIGINT / SIGTERM
	// ---------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	fmt.Println("server stopped")
	return nil
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
