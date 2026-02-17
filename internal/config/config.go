package config

import "os"

type Config struct {
	RabbitMQURL string
	Port        string
}

func Load() *Config {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		RabbitMQURL: rabbitURL,
		Port:        port,
	}
}
