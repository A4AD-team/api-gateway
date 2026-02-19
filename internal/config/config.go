package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	AppName    string
	AppEnv     string
	AppVersion string
	LogLevel   string
	Port       string

	// RabbitMQ
	RabbitMQ *RabbitMQConfig

	// Services
	Services map[string]*ServiceConfig

	// JWT
	JWT *JWTConfig

	// Server
	Server *ServerConfig

	// Rate Limiting
	RateLimit *RateLimitConfig

	// CORS
	CORS *CORSConfig

	// Redis
	Redis *RedisConfig

	// Metrics
	Metrics *MetricsConfig

	// Feature Flags
	Features map[string]bool
}

type RabbitMQConfig struct {
	URL               string
	Host              string
	Port              string
	User              string
	Pass              string
	VHost             string
	Heartbeat         int
	ConnectionTimeout time.Duration
	ChannelMax        int
	FrameMax          int

	Queues      []QueueConfig
	Exchanges   []ExchangeConfig
	Bindings    []BindingConfig
	Policies    []PolicyConfig
	Users       []UserConfig
	Permissions []PermissionConfig
}

type QueueConfig struct {
	Name       string                 `json:"name"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Exclusive  bool                   `json:"exclusive"`
	Arguments  map[string]interface{} `json:"arguments"`
}

type ExchangeConfig struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	Arguments  map[string]interface{} `json:"arguments"`
}

type BindingConfig struct {
	Source          string `json:"source"`
	Destination     string `json:"destination"`
	DestinationType string `json:"destination_type"`
	RoutingKey      string `json:"routing_key"`
}

type PolicyConfig struct {
	Name       string                 `json:"name"`
	Pattern    string                 `json:"pattern"`
	Definition map[string]interface{} `json:"definition"`
	Priority   int                    `json:"priority"`
}

type UserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Tags     string `json:"tags"`
}

type PermissionConfig struct {
	User      string `json:"user"`
	VHost     string `json:"vhost"`
	Configure string `json:"configure"`
	Write     string `json:"write"`
	Read      string `json:"read"`
}

type ServiceConfig struct {
	Name           string
	URL            string
	Timeout        time.Duration
	RetryCount     int
	CircuitBreaker bool
	MaxConnections int
	Weight         int
}

type JWTConfig struct {
	Secret            string
	Expiration        time.Duration
	RefreshExpiration time.Duration
	Issuer            string
	Audience          string
}

type ServerConfig struct {
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
}

type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
	Strategy string
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	MaxAge           time.Duration
	AllowCredentials bool
}

type RedisConfig struct {
	Enabled      bool
	URL          string
	Host         string
	Port         string
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

type MetricsConfig struct {
	Enabled         bool
	Path            string
	Port            string
	TracingEnabled  bool
	TracingProvider string
	TracingEndpoint string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	// Загружаем .env файл если существует
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		AppName:    getEnv("APP_NAME", "api-gateway"),
		AppEnv:     getEnv("APP_ENV", "development"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		Port:       getEnv("PORT", "8080"),

		RabbitMQ:  loadRabbitMQConfig(),
		Services:  loadServicesConfig(),
		JWT:       loadJWTConfig(),
		Server:    loadServerConfig(),
		RateLimit: loadRateLimitConfig(),
		CORS:      loadCORSConfig(),
		Redis:     loadRedisConfig(),
		Metrics:   loadMetricsConfig(),
		Features:  loadFeatureFlags(),
	}

	// Валидация
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	cfg.logConfig()
	return cfg
}

func loadRabbitMQConfig() *RabbitMQConfig {
	// Базовые настройки
	cfg := &RabbitMQConfig{
		Host:              getEnv("RABBITMQ_HOST", "localhost"),
		Port:              getEnv("RABBITMQ_PORT", "5672"),
		User:              getEnv("RABBITMQ_USER", "guest"),
		Pass:              getEnv("RABBITMQ_PASS", "guest"),
		VHost:             getEnv("RABBITMQ_VHOST", "/"),
		Heartbeat:         getIntEnv("RABBITMQ_HEARTBEAT", 30),
		ConnectionTimeout: getDurationEnv("RABBITMQ_CONNECTION_TIMEOUT", 30*time.Second),
		ChannelMax:        getIntEnv("RABBITMQ_CHANNEL_MAX", 100),
		FrameMax:          getIntEnv("RABBITMQ_FRAME_MAX", 131072),
	}

	// Формируем URL если не задан явно
	url := getEnv("RABBITMQ_URL", "")
	if url == "" {
		url = fmt.Sprintf("amqp://%s:%s@%s:%s%s",
			cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.VHost)
	}
	cfg.URL = url

	// Загружаем очереди из JSON
	if queuesJSON := getEnv("RABBITMQ_QUEUES", ""); queuesJSON != "" {
		var queues []QueueConfig
		if err := json.Unmarshal([]byte(queuesJSON), &queues); err == nil {
			cfg.Queues = queues
		} else {
			log.Printf("Error parsing RABBITMQ_QUEUES: %v", err)
		}
	}

	// Загружаем обменники
	if exchangesJSON := getEnv("RABBITMQ_EXCHANGES", ""); exchangesJSON != "" {
		var exchanges []ExchangeConfig
		if err := json.Unmarshal([]byte(exchangesJSON), &exchanges); err == nil {
			cfg.Exchanges = exchanges
		}
	}

	// Загружаем биндинги
	if bindingsJSON := getEnv("RABBITMQ_BINDINGS", ""); bindingsJSON != "" {
		var bindings []BindingConfig
		if err := json.Unmarshal([]byte(bindingsJSON), &bindings); err == nil {
			cfg.Bindings = bindings
		}
	}

	// Загружаем политики
	if policiesJSON := getEnv("RABBITMQ_POLICIES", ""); policiesJSON != "" {
		var policies []PolicyConfig
		if err := json.Unmarshal([]byte(policiesJSON), &policies); err == nil {
			cfg.Policies = policies
		}
	}

	// Загружаем пользователей (с подстановкой паролей)
	if usersJSON := getEnv("RABBITMQ_USERS", ""); usersJSON != "" {
		// Заменяем ${VAR} на значения из env
		usersJSON = expandEnvVars(usersJSON)
		var users []UserConfig
		if err := json.Unmarshal([]byte(usersJSON), &users); err == nil {
			cfg.Users = users
		}
	}

	return cfg
}

func loadServicesConfig() map[string]*ServiceConfig {
	services := make(map[string]*ServiceConfig)

	// Список сервисов
	serviceNames := []string{"AUTH", "PROFILE", "POST", "COMMENT"}

	for _, name := range serviceNames {
		prefix := strings.ToUpper(name)
		serviceName := getEnv(prefix+"_SERVICE_NAME", strings.ToLower(name)+"-service")

		cfg := &ServiceConfig{
			Name:           serviceName,
			URL:            getEnv(prefix+"_SERVICE_URL", fmt.Sprintf("http://%s:8080", serviceName)),
			Timeout:        getDurationEnv(prefix+"_SERVICE_TIMEOUT", 10*time.Second),
			RetryCount:     getIntEnv(prefix+"_SERVICE_RETRY_COUNT", 3),
			CircuitBreaker: getBoolEnv(prefix+"_SERVICE_CIRCUIT_BREAKER", true),
			MaxConnections: getIntEnv(prefix+"_SERVICE_MAX_CONNECTIONS", 100),
			Weight:         getIntEnv(prefix+"_SERVICE_WEIGHT", 1),
		}

		services[strings.ToLower(name)] = cfg
	}

	return services
}

func loadJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:            mustGetEnv("JWT_SECRET"), // Обязательное поле
		Expiration:        getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		RefreshExpiration: getDurationEnv("JWT_REFRESH_EXPIRATION", 168*time.Hour),
		Issuer:            getEnv("JWT_ISSUER", "api-gateway"),
		Audience:          getEnv("JWT_AUDIENCE", "mobile-client"),
	}
}

func loadServerConfig() *ServerConfig {
	return &ServerConfig{
		ReadTimeout:    getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:   getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:    getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		MaxHeaderBytes: getIntEnv("SERVER_MAX_HEADER_BYTES", 1048576),
	}
}

func loadRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:  getBoolEnv("RATE_LIMIT_ENABLED", true),
		Requests: getIntEnv("RATE_LIMIT_REQUESTS", 1000),
		Window:   getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
		Strategy: getEnv("RATE_LIMIT_STRATEGY", "token-bucket"),
	}
}

func loadCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     getSliceEnv("CORS_ALLOW_ORIGINS", []string{"*"}),
		AllowMethods:     getSliceEnv("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowHeaders:     getSliceEnv("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
		ExposeHeaders:    getSliceEnv("CORS_EXPOSE_HEADERS", []string{"Content-Length"}),
		MaxAge:           getDurationEnv("CORS_MAX_AGE", 12*time.Hour),
		AllowCredentials: getBoolEnv("CORS_ALLOW_CREDENTIALS", true),
	}
}

func loadRedisConfig() *RedisConfig {
	cfg := &RedisConfig{
		Enabled:      getBoolEnv("REDIS_ENABLED", false),
		Host:         getEnv("REDIS_HOST", "redis"),
		Port:         getEnv("REDIS_PORT", "6379"),
		Password:     getEnv("REDIS_PASS", ""),
		DB:           getIntEnv("REDIS_DB", 0),
		MaxRetries:   getIntEnv("REDIS_MAX_RETRIES", 3),
		PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
		MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
	}

	// Формируем URL если не задан явно
	url := getEnv("REDIS_URL", "")
	if url == "" {
		if cfg.Password != "" {
			url = fmt.Sprintf("redis://:%s@%s:%s/%d",
				cfg.Password, cfg.Host, cfg.Port, cfg.DB)
		} else {
			url = fmt.Sprintf("redis://%s:%s/%d",
				cfg.Host, cfg.Port, cfg.DB)
		}
	}
	cfg.URL = url

	return cfg
}

func loadMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:         getBoolEnv("METRICS_ENABLED", true),
		Path:            getEnv("METRICS_PATH", "/metrics"),
		Port:            getEnv("METRICS_PORT", "9090"),
		TracingEnabled:  getBoolEnv("TRACING_ENABLED", false),
		TracingProvider: getEnv("TRACING_PROVIDER", "jaeger"),
		TracingEndpoint: getEnv("TRACING_ENDPOINT", "http://jaeger:14268"),
	}
}

func loadFeatureFlags() map[string]bool {
	return map[string]bool{
		"async_messaging": getBoolEnv("FEATURE_ASYNC_MESSAGING", true),
		"sync_messaging":  getBoolEnv("FEATURE_SYNC_MESSAGING", true),
		"caching":         getBoolEnv("FEATURE_CACHING", false),
		"compression":     getBoolEnv("FEATURE_COMPRESSION", true),
	}
}

// Валидация конфигурации
func (c *Config) Validate() error {
	if c.AppEnv == "production" {
		if c.JWT.Secret == "change-this-in-production" ||
			c.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
			return fmt.Errorf("JWT_SECRET must be changed in production")
		}

		if c.RabbitMQ.Pass == "guest" {
			return fmt.Errorf("RABBITMQ_PASS must be changed in production")
		}
	}

	return nil
}

// Логирование конфигурации (без секретов)
func (c *Config) logConfig() {
	log.Println("=== Configuration ===")
	log.Printf("App: %s v%s (%s)", c.AppName, c.AppVersion, c.AppEnv)
	log.Printf("Port: %s", c.Port)
	log.Printf("Log Level: %s", c.LogLevel)

	log.Printf("RabbitMQ: %s@%s:%s",
		c.RabbitMQ.User, c.RabbitMQ.Host, c.RabbitMQ.Port)
	log.Printf("Queues: %d", len(c.RabbitMQ.Queues))

	log.Printf("Services: %d", len(c.Services))
	for name, svc := range c.Services {
		log.Printf("  %s: %s (timeout: %v)", name, svc.URL, svc.Timeout)
	}

	log.Printf("Redis Enabled: %v", c.Redis.Enabled)
	log.Printf("Metrics Enabled: %v", c.Metrics.Enabled)
	log.Printf("Rate Limit: %d/%v", c.RateLimit.Requests, c.RateLimit.Window)
	log.Println("======================")
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		parts := strings.Split(value, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Required environment variable %s is not set", key)
	return ""
}

func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}
