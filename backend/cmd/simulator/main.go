package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
	"github.com/emiliogain/smart-home-backend/internal/simulator"
	"github.com/emiliogain/smart-home-backend/pkg/client"
)

// httpReadingSaver wraps the HTTP client to satisfy simulator.ReadingSaver.
type httpReadingSaver struct {
	client *client.Client
}

func (h *httpReadingSaver) SaveReading(_ context.Context, r sensor.Reading) (*secondary.FusionResult, error) {
	err := h.client.SubmitReading(r.SensorID, r.Value, r.Unit)
	return nil, err
}

func (h *httpReadingSaver) SaveReadingsBatch(ctx context.Context, readings []sensor.Reading) (*secondary.FusionResult, error) {
	for _, r := range readings {
		if _, err := h.SaveReading(ctx, r); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func main() {
	backendURL := flag.String("backend", "http://localhost:8080", "Backend base URL")
	scenarioFlag := flag.String("scenario", "", "Run a single scenario")
	interval := flag.Duration("interval", 5*time.Second, "Interval between readings")
	flag.Parse()

	c := client.New(*backendURL)

	// Register sensors via HTTP.
	sensorIDs := make(map[string]string)
	for _, s := range simulator.DefaultSensors {
		resp, err := c.RegisterSensor(s.Name, s.Type, s.Location)
		if err != nil {
			log.Fatalf("Failed to register sensor %s: %v", s.Name, err)
		}
		sensorIDs[s.Name] = resp.ID
		log.Printf("Registered sensor %s (id=%s)", s.Name, resp.ID)
	}

	// Build engine options.
	opts := []simulator.Option{simulator.WithInterval(*interval)}
	if *scenarioFlag != "" {
		opts = append(opts, simulator.WithScenario(*scenarioFlag))
		// Validate the scenario name.
		found := false
		for _, name := range simulator.ScenarioNames() {
			if strings.EqualFold(name, *scenarioFlag) {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Unknown scenario %q. Available: %s\n",
				*scenarioFlag, strings.Join(simulator.ScenarioNames(), ", "))
			os.Exit(1)
		}
	}

	saver := &httpReadingSaver{client: c}
	engine := simulator.NewEngine(saver, sensorIDs, opts...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go engine.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	engine.Stop()
	log.Println("Simulator exited")
}
