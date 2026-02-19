# Forum API Gateway

Единая точка входа для фронтенда и мобильных клиентов форума.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-2+-00ADD8?logo=go&logoColor=white)](https://gofiber.io/)

## Назначение

- Маршрутизация запросов к backend-сервисам
- Централизованная проверка JWT (от auth-service)
- Rate limiting, CORS, request logging, tracing
- Health-check и readiness/liveness

## Технологии

- Go 1.23+
- Fiber (или Gin/Chi)
- golang-jwt
- OpenTelemetry + Prometheus
- Redis (rate-limit / session cache)

## Структура
Планируемая
```
api-gateway/
├── cmd/
│   └── gateway/
├── internal/
│   ├── config/
│   ├── middleware/
│   ├── routes/
│   └── proxy/
├── Dockerfile
└── docker-compose.yml
```
Нынешняя

```
api-gateway/
├── cmd/
│   └── main.go
├── internal/
│   ├── broker/
│   ├── config/
│   ├── handlers/
│   ├── models/
│   └── router/
├── Dockerfile
└── docker-compose.yml
```


## Быстрый старт

```bash
# Зависимости
docker compose up -d redis

# Запуск
go run ./cmd/gateway -config=config/local.yaml
# или
make run
```

API доступен: <http://localhost:8080>

## Основные маршруты

/auth/*→ auth-service (без JWT)
/api/v1/profile/* → profile-service
/api/v1/posts/*→ post-service
/api/v1/comments/* → comment-service
/health           → health check

## Docker Compose (фрагмент)

```yaml
services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8081
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
```
