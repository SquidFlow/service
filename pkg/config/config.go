package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// KubernetesMode represents the mode of Kubernetes connection
type KubernetesMode string

const (
	KubeconfigMode KubernetesMode = "kubeconfig"
	InClusterMode  KubernetesMode = "incluster"
)

// file: deploy/service/templates/config.toml
type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Server   struct {
		Address string `mapstructure:"address"`
		Port    int    `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
	} `mapstructure:"database"`
	Kubernetes struct {
		Mode       KubernetesMode `mapstructure:"mode"`
		KubeConfig struct {
			Path string `mapstructure:"path"`
		} `mapstructure:"kubeconfig"`
	} `mapstructure:"kubernetes"`
	ApplicationRepo struct {
		Provider    []string `mapstructure:"provider"`
		LocalPath   string   `mapstructure:"local_path"`
		RemoteURL   string   `mapstructure:"remote_url"`
		AccessToken string   `mapstructure:"access_token"`
	} `mapstructure:"application_repo"`
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

	//  check if the config is valid
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	if config.Kubernetes.Mode == "" {
		return fmt.Errorf("kubernetes mode is required")
	}
	if config.Kubernetes.Mode != KubeconfigMode && config.Kubernetes.Mode != InClusterMode {
		return fmt.Errorf("invalid kubernetes mode: must be either 'kubeconfig' or 'incluster'")
	}
	if config.Kubernetes.Mode == KubeconfigMode && config.Kubernetes.KubeConfig.Path == "" {
		return fmt.Errorf("kubernetes kubeconfig path is required when mode is kubeconfig")
	}

	if len(config.ApplicationRepo.Provider) == 0 {
		return fmt.Errorf("at least one application repo provider is required")
	}

	if config.ApplicationRepo.RemoteURL == "" {
		return fmt.Errorf("application repo remote URL is required")
	}
	if config.ApplicationRepo.Provider[0] == "github" && config.ApplicationRepo.AccessToken == "" {
		_ = os.Setenv("GIT_TOKEN", config.ApplicationRepo.AccessToken)
	}

	return nil
}
