package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config matches config.toml structure
type Config struct {
	Log struct {
		Level string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	} `mapstructure:"log"`

	Server struct {
		Address string `mapstructure:"address" validate:"required,ip"`
		Port    int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	} `mapstructure:"server"`

	ArgoCD struct {
		ServerAddress string `mapstructure:"server_address" validate:"required,hostname_port"`
		Username      string `mapstructure:"username" validate:"required"`
		Password      string `mapstructure:"password" validate:"required"`
	} `mapstructure:"argocd"`

	ApplicationRepo struct {
		Provider    []string `mapstructure:"provider" validate:"required,min=1,dive,oneof=github"`
		RemoteURL   string   `mapstructure:"remote_url" validate:"required,url"`
		AccessToken string   `mapstructure:"access_token" validate:"required"`
	} `mapstructure:"application_repo"`
}

func init() {
	// support environment variable
	viper.AutomaticEnv()
	viper.SetEnvPrefix("H4")

	// set default value
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log.level", "info")
}

func ParseConfig(configFilePath string) (*Config, error) {
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config using validator
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Set environment variable for git token if provided
	if config.ApplicationRepo.AccessToken != "" {
		_ = os.Setenv("GIT_TOKEN", config.ApplicationRepo.AccessToken)
	}

	return &config, nil
}
