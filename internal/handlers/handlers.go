package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"api-gateway/internal/broker"
	"api-gateway/internal/models"
)

type MessageHandler struct {
	rabbitClient *broker.RabbitMQClient
	responseMap  map[string]chan *models.MessageResponse
}

func NewMessageHandler(rabbitClient *broker.RabbitMQClient) *MessageHandler {
	return &MessageHandler{
		rabbitClient: rabbitClient,
		responseMap:  make(map[string]chan *models.MessageResponse),
	}
}

// SendMessage - handler for sending messages to queue
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.MessageResponse{
			Status: "error",
			Error:  "Invalid request format: " + err.Error(),
		})
		return
	}

	// Create unique message ID
	messageID := uuid.New().String()

	queueMsg := models.QueueMessage{
		ID:        messageID,
		UserID:    req.UserID,
		Action:    req.Action,
		Payload:   req.Payload,
		Timestamp: time.Now().Unix(),
		Metadata:  req.Metadata,
	}

	// Determine queue based on action
	queueName := h.getQueueForAction(req.Action)

	// Publish message to RabbitMQ
	err := h.rabbitClient.PublishMessage(queueName, queueMsg)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		c.JSON(http.StatusInternalServerError, models.MessageResponse{
			Status: "error",
			Error:  "Failed to send message",
		})
		return
	}

	log.Printf("Message sent: %s to queue: %s", messageID, queueName)

	c.JSON(http.StatusAccepted, models.MessageResponse{
		Status:    "accepted",
		MessageID: messageID,
	})
}

// GetMessageStatus - get message status
func (h *MessageHandler) GetMessageStatus(c *gin.Context) {
	messageID := c.Param("id")

	// Here you can implement status check from DB or queue
	c.JSON(http.StatusOK, models.MessageResponse{
		Status:    "processing",
		MessageID: messageID,
	})
}

// HealthCheck - service health check
func (h *MessageHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// GetQueueInfo - information about queues
func (h *MessageHandler) GetQueueInfo(c *gin.Context) {
	// Here you can get queue status information
	// via RabbitMQ Management API
	c.JSON(http.StatusOK, gin.H{
		"queues": []string{"user_actions", "notifications", "data_requests"},
	})
}

func (h *MessageHandler) getQueueForAction(action string) string {
	// Routing based on action
	switch action {
	case "login", "register", "logout":
		return "user_actions"
	case "send_notification":
		return "notifications"
	case "get_data", "update_data":
		return "data_requests"
	default:
		return "default_queue"
	}
}
