package simulator

import (
	"strings"
	"time"
)

// Option configures the simulation engine.
type Option func(*Engine)

// WithInterval sets the initial tick interval.
func WithInterval(d time.Duration) Option {
	return func(e *Engine) {
		if d >= time.Second && d <= 30*time.Second {
			e.interval = d
		}
	}
}

// WithScenario sets the initial scenario by name.
func WithScenario(name string) Option {
	return func(e *Engine) {
		for _, sc := range e.allScenarios {
			if strings.EqualFold(sc.Name, name) {
				e.scenario = sc
				return
			}
		}
	}
}
