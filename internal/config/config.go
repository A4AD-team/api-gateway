package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Port        string        `json:"port"`
	Environment string        `json:"environment"`
	Services    Services      `json:"services"`
	Timeout     time.Duration `json:"timeout"`
	MaxFileSize int64         `json:"max_file_size"`
}

type Services struct {
	UserService    ServiceConfig `json:"user_service"`
	OrderService   ServiceConfig `json:"order_service"`
	ProductService ServiceConfig `json:"product_service"`
}

type ServiceConfig struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Default values
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 << 20 // 10 MB
	}

	return &config, nil
}
