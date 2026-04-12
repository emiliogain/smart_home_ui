package config

import (
	"errors"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

// Config holds all application configuration.
// Fields are populated from backend/config/config.yaml (search paths below) with env var overrides (prefix: SMARTHOME_).
type Config struct {
	Environment       string   `mapstructure:"environment"`
	ServerPort        int      `mapstructure:"server_port"`
	LogLevel          string   `mapstructure:"log_level"`
	DatabaseURL       string   `mapstructure:"database_url"`
	RedisURL          string   `mapstructure:"redis_url"`
	CORSOrigins       []string `mapstructure:"cors_origins"`
	SimulatorEnabled  bool     `mapstructure:"simulator_enabled"`
	SimulatorInterval string   `mapstructure:"simulator_interval"`
	SimulatorScenario string   `mapstructure:"simulator_scenario"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetEnvPrefix("SMARTHOME")
	viper.AutomaticEnv()

	viper.SetDefault("environment", "development")
	viper.SetDefault("server_port", 8080)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("database_url", "")
	viper.SetDefault("redis_url", "")
	viper.SetDefault("cors_origins", []string{"http://localhost:3000"})
	viper.SetDefault("simulator_enabled", true)
	viper.SetDefault("simulator_interval", "5s")
	viper.SetDefault("simulator_scenario", "comfortable")

	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, xerrors.Errorf("failed to read in config: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
