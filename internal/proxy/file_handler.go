package proxy

import (
	"api-gateway/internal/config"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var ErrInvalidService = errors.New("invalid service")

type FileHandler struct {
	config *config.Config
	logger *logrus.Logger
	client *http.Client
}

func NewFileHandler(cfg *config.Config, logger *logrus.Logger) *FileHandler {
	return &FileHandler{
		config: cfg,
		logger: logger,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// UploadFile обрабатывает загрузку файлов
func (h *FileHandler) UploadFile(c *gin.Context) {
	service := c.Param("service")

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Проверяем размер файла
	if header.Size > h.config.MaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File too large. Max size: %d bytes", h.config.MaxFileSize),
		})
		return
	}

	// Читаем файл
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Получаем URL сервиса для загрузки
	serviceURL, err := h.getServiceUploadURL(service)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create form file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	_, err = part.Write(fileBytes)
	if err != nil {
		h.logger.WithError(err).Error("Failed to write file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	writer.Close()

	// Создаем запрос к микросервису
	req, err := http.NewRequest(http.MethodPost, serviceURL, body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Копируем auth заголовки
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Отправляем запрос
	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upload file to service")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Отправляем ответ клиенту
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	h.logger.WithFields(logrus.Fields{
		"service":   service,
		"file_name": header.Filename,
		"file_size": header.Size,
		"client_ip": c.ClientIP(),
		"status":    resp.StatusCode,
	}).Info("File uploaded successfully")
}

// DownloadFile скачивает файл из микросервиса
func (h *FileHandler) DownloadFile(c *gin.Context) {
	service := c.Param("service")
	fileID := c.Param("fileId")

	serviceURL, err := h.getServiceDownloadURL(service, fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest(http.MethodGet, serviceURL, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to download file")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "File not found"})
		return
	}

	// Читаем файл
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Определяем Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Получаем имя файла
	fileName := filepath.Base(fileID)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", len(fileBytes)))

	c.Data(http.StatusOK, contentType, fileBytes)

	h.logger.WithFields(logrus.Fields{
		"service":   service,
		"file_id":   fileID,
		"file_size": len(fileBytes),
		"client_ip": c.ClientIP(),
	}).Info("File downloaded successfully")
}

func (h *FileHandler) getServiceUploadURL(service string) (string, error) {
	switch service {
	case "users":
		return h.config.Services.UserService.URL + "/api/files/upload", nil
	case "orders":
		return h.config.Services.OrderService.URL + "/api/files/upload", nil
	case "products":
		return h.config.Services.ProductService.URL + "/api/files/upload", nil
	default:
		return "", ErrInvalidService
	}
}

func (h *FileHandler) getServiceDownloadURL(service, fileID string) (string, error) {
	switch service {
	case "users":
		return h.config.Services.UserService.URL + "/api/files/" + fileID, nil
	case "orders":
		return h.config.Services.OrderService.URL + "/api/files/" + fileID, nil
	case "products":
		return h.config.Services.ProductService.URL + "/api/files/" + fileID, nil
	default:
		return "", ErrInvalidService
	}
}
