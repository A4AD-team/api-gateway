package main

import (
	"log"

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

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/messages", handler.SendMessage)
		api.GET("/messages/:id", handler.GetMessageStatus)
		api.GET("/health", handler.HealthCheck)
		api.GET("/queues", handler.GetQueueInfo)
	}

	// Start server
	log.Printf("API Gateway starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
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
