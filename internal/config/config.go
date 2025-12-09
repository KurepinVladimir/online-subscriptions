package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

func Load() (*Config, error) {
	v := viper.New()

	// yaml config
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("logging.level", "info")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.dsn is required")
	}

	return &cfg, nil
}
