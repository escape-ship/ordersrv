package config

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Database Database `mapstructure:"database"`
	}

	Database struct {
		Host         string `mapstructure:"host"`          // DATABASE_HOST
		Port         int    `mapstructure:"port"`          // DATABASE_PORT
		User         string `mapstructure:"user"`          // DATABASE_USER
		Password     string `mapstructure:"password"`      // DATABASE_PASSWORD
		DataBaseName string `mapstructure:"database_name"` // DATABASE_DATABASE_NAME
		SchemaName   string `mapstructure:"schema_name"`   // DATABASE_SCHEMA_NAME
		SSLMode      string `mapstructure:"ssl_mode"`      // DATABASE_SSL_MODE
	}
)

func New(path string) (*Config, error) {
	vp := viper.New()
	vp.SetConfigFile(path)
	vp.AutomaticEnv()

	dir, err := os.Getwd()
	if err != nil {
		slog.Error("App: get current directory error", "error", err)
		os.Exit(1)
	}
	slog.Info("App: current directory", "dir", dir)

	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
