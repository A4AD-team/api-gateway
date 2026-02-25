package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"api-gateway/internal/broker"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize RabbitMQ connection
	rabbitClient, err := broker.NewRabbitMQClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	// Declare queues
	queues := []string{"user_actions", "notifications", "data_requests", "default_queue"}
	for _, queue := range queues {
		if err := rabbitClient.DeclareQueue(queue); err != nil {
			log.Printf("Failed to declare queue %s: %v", queue, err)
		} else {
			log.Printf("Queue %s declared successfully", queue)
		}
	}

	// Create handler
	handler := handlers.NewMessageHandler(rabbitClient)

	// Setup router
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Proxy routes
	proxy := NewReverseProxy(cfg)

	// Auth service routes
	authGroup := router.Group("/api/v1")
	{
		authGroup.Any("/auth/*path", proxy.proxyHandler("auth"))
		authGroup.Any("/users/*path", proxy.proxyHandler("auth"))
		authGroup.Any("/roles/*path", proxy.proxyHandler("auth"))
		authGroup.Any("/permissions/*path", proxy.proxyHandler("auth"))
	}

	// Post service routes
	authGroup.Any("/posts/*path", proxy.proxyHandler("post"))

	// Comment service routes
	authGroup.Any("/comments/*path", proxy.proxyHandler("comment"))

	// Start server
	log.Printf("API Gateway starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// ReverseProxy handles routing to backend services
type ReverseProxy struct {
	config *config.Config
}

func NewReverseProxy(cfg *config.Config) *ReverseProxy {
	return &ReverseProxy{config: cfg}
}

func (p *ReverseProxy) proxyHandler(serviceKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		svc, ok := p.config.Services[serviceKey]
		if !ok || svc == nil || svc.URL == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service not configured: " + serviceKey})
			return
		}

		targetURL, err := url.Parse(svc.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid service URL"})
			return
		}

		// Get the full path including the route prefix
		fullPath := c.Request.URL.Path
		// Remove /api/v1 prefix since backend services expect paths without it
		apiPath := strings.TrimPrefix(fullPath, "/api/v1")

		log.Printf("Proxying %s %s to %s", c.Request.Method, fullPath, svc.URL+apiPath)

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = apiPath
			req.URL.RawQuery = c.Request.URL.RawQuery
			req.Host = targetURL.Host
			req.Header = c.Request.Header.Clone()
			if _, exists := req.Header["User-Agent"]; !exists {
				req.Header["User-Agent"] = []string{"api-gateway"}
			}
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// CORS middleware for mobile clients
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
