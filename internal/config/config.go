package config

import (
	"fmt"
	"strings"

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

	// defaults
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("logging.level", "info")

	// env
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	_ = v.ReadInConfig() // если файла нет — не падаем

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.dsn is required")
	}
	return &cfg, nil
}
