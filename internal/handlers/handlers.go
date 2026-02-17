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

// SendMessage - обработчик для отправки сообщений в очередь
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.MessageResponse{
			Status: "error",
			Error:  "Invalid request format: " + err.Error(),
		})
		return
	}

	// Создаем уникальный ID сообщения
	messageID := uuid.New().String()

	queueMsg := models.QueueMessage{
		ID:        messageID,
		UserID:    req.UserID,
		Action:    req.Action,
		Payload:   req.Payload,
		Timestamp: time.Now().Unix(),
		Metadata:  req.Metadata,
	}

	// Определяем очередь на основе action
	queueName := h.getQueueForAction(req.Action)

	// Публикуем сообщение в RabbitMQ
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

// GetMessageStatus - получение статуса сообщения
func (h *MessageHandler) GetMessageStatus(c *gin.Context) {
	messageID := c.Param("id")

	// Здесь можно реализовать проверку статуса из БД или очереди
	c.JSON(http.StatusOK, models.MessageResponse{
		Status:    "processing",
		MessageID: messageID,
	})
}

// HealthCheck - проверка здоровья сервиса
func (h *MessageHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// GetQueueInfo - информация об очередях
func (h *MessageHandler) GetQueueInfo(c *gin.Context) {
	// Здесь можно получить информацию о состоянии очередей
	// через Management API RabbitMQ
	c.JSON(http.StatusOK, gin.H{
		"queues": []string{"user_actions", "notifications", "data_requests"},
	})
}

func (h *MessageHandler) getQueueForAction(action string) string {
	// Маршрутизация на основе action
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
