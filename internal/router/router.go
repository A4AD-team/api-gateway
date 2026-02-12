package router

import (
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewRouter(cfg *config.Config, logger *logrus.Logger) *gin.Engine {
	router := gin.New()

	// Глобальные middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())

	// Инициализация обработчиков
	proxyHandler := proxy.NewProxyHandler(cfg, logger)
	fileHandler := proxy.NewFileHandler(cfg, logger)

	// Публичные маршруты
	public := router.Group("/api")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
		})
	}

	// Защищенные маршруты (требуют аутентификации)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Прокси для REST запросов
		protected.Any("/:service/*path", proxyHandler.ProxyRequest)

		// Загрузка файлов
		protected.POST("/:service/files/upload", fileHandler.UploadFile)
		protected.GET("/:service/files/:fileId", fileHandler.DownloadFile)
	}

	return router
}
