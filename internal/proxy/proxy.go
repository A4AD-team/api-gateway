package proxy

import (
	"api-gateway/internal/config"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProxyHandler struct {
	config *config.Config
	logger *logrus.Logger
	client *http.Client
}

func NewProxyHandler(cfg *config.Config, logger *logrus.Logger) *ProxyHandler {
	return &ProxyHandler{
		config: cfg,
		logger: logger,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// ProxyRequest проксирует запросы к микросервисам
func (h *ProxyHandler) ProxyRequest(c *gin.Context) {
	serviceName := c.Param("service")
	path := c.Param("path")

	// Определяем URL целевого сервиса
	targetURL, err := h.getServiceURL(serviceName, path)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get service URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service"})
		return
	}

	// Читаем тело запроса
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Создаем прокси запрос
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		h.logger.WithError(err).Error("Failed to create proxy request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	// Копируем заголовки
	h.copyHeaders(c.Request.Header, proxyReq.Header)

	// Добавляем метаданные
	proxyReq.Header.Set("X-Forwarded-For", c.ClientIP())
	proxyReq.Header.Set("X-Original-Host", c.Request.Host)

	// Выполняем запрос
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute proxy request")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read response body")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Копируем заголовки ответа
	h.copyHeaders(resp.Header, c.Writer.Header())

	// Отправляем ответ
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	// Логируем запрос
	h.logger.WithFields(logrus.Fields{
		"service":   serviceName,
		"method":    c.Request.Method,
		"path":      path,
		"status":    resp.StatusCode,
		"client_ip": c.ClientIP(),
	}).Info("Proxy request completed")
}

func (h *ProxyHandler) getServiceURL(service, path string) (string, error) {
	var baseURL string

	switch service {
	case "users":
		baseURL = h.config.Services.UserService.URL
	case "orders":
		baseURL = h.config.Services.OrderService.URL
	case "products":
		baseURL = h.config.Services.ProductService.URL
	default:
		return "", ErrInvalidService
	}

	// Очищаем путь от лишних слешей
	cleanPath := strings.TrimPrefix(path, "/")
	return strings.TrimSuffix(baseURL, "/") + "/" + cleanPath, nil
}

func (h *ProxyHandler) copyHeaders(source, destination http.Header) {
	for key, values := range source {
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}
