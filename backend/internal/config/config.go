package config

import (
	"errors"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

// Config holds all application configuration.
// Fields are populated from ./config.yaml with env var overrides (prefix: SMARTHOME_).
type Config struct {
	Environment string   `mapstructure:"environment"`
	ServerPort  int      `mapstructure:"server_port"`
	LogLevel    string   `mapstructure:"log_level"`
	DatabaseURL string   `mapstructure:"database_url"`
	RedisURL    string   `mapstructure:"redis_url"`
	CORSOrigins []string `mapstructure:"cors_origins"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("SMARTHOME")
	viper.AutomaticEnv()

	viper.SetDefault("environment", "development")
	viper.SetDefault("server_port", 8080)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("database_url", "")
	viper.SetDefault("redis_url", "")
	viper.SetDefault("cors_origins", []string{"http://localhost:3000"})

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
