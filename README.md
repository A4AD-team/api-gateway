# API Gateway

Единая точка входа для всех внешних и внутренних клиентов нашего движка бизнес-процессов.  
Обеспечивает маршрутизацию, аутентификацию, rate limiting, observability и базовую валидацию запросов.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-2+-00ADD8?logo=go&logoColor=white)](https://gofiber.io/)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-1.0+-yellowgreen)](https://opentelemetry.io/)

## Назначение и ключевые обязанности

- Единый вход для всех клиентов (веб, мобильное приложение, интеграции, партнеры)
- Централизованная **аутентификация** и **авторизация** (JWT от Auth Service + RBAC)
- **Маршрутизация** запросов к backend-сервисам (Request, Workflow, Comment и т.д.)
- **Rate limiting** и **throttling** по пользователям / ролям / IP
- **CORS**, **request/response transformation** (если нужно)
- **Logging**, **tracing**, **metrics** (Prometheus + OpenTelemetry)
- **Health-check** и **readiness/liveness** для Kubernetes
- Защита от базовых атак (размер тела, заголовки, таймауты)

## Технологический стек

- Go 1.23+
- Fiber (или Gin / Chi / Echo — Fiber выбран за скорость и удобство middlewares)
- JWT-go / golang-jwt (валидация токенов от Auth Service)
- go-rate-limiter или redis-based limiter
- OpenTelemetry (tracing + metrics)
- Prometheus client_golang
- Zerolog / slog (structured logging)
- Viper / envconfig (конфигурация)
- Docker + Compose + Kubernetes-ready

## Структура проекта

```
api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── config/
│   ├── middleware/
│   │   ├── auth/
│   │   ├── logging/
│   │   ├── ratelimit/
│   │   ├── tracing/
│   │   └── cors/
│   ├── proxy/
│   │   └── reverse_proxy.go
│   ├── routes/
│   └── health/
├── pkg/
│   └── limiter/
├── config/
│   └── config.yaml
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Быстрый старт (локально)

```bash
# 1. Запускаем зависимости (Redis для rate-limit + Auth Service для тестов)
docker compose up -d redis auth-service

# 2. Запуск API Gateway
cp .env.example .env
# отредактируйте .env (AUTH_SERVICE_URL, JWT_SECRET и т.д.)

go run ./cmd/gateway
# или
make run
```

API будет доступен по умолчанию на <http://localhost:8080>

## Docker Compose пример

```yaml
services:
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8081
      - JWT_SECRET=super-secret-key-change-me
      - REDIS_ADDR=redis:6379
      - LOG_LEVEL=debug
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
    depends_on:
      - redis
      - auth-service

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

## Основные маршруты (примеры)

TODO: Сделать это нормальной таблицей
Метод,Путь,Описание,Middleware
POST,/api/v1/auth/login,Прокси на Auth Service (login),без auth
POST,/api/v1/auth/register,Прокси на Auth Service (register),без auth
_,/api/v1/requests/_,Заявки,auth + rbac + ratelimit
_,/api/v1/workflows/_,Маршруты согласования,auth + rbac
\_,/api/v1/comments/\*,Комментарии,auth
GET,/health,Health-check,—
GET,/metrics,Prometheus metrics,—

## Конфигурация (через .env + config.yaml)

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 30s

auth:
  service_url: http://auth-service:8081
  jwks_url: http://auth-service:8081/.well-known/jwks.json # если используем JWKS

routes:
  - prefix: /api/v1/requests
    backend: http://request-service:8083
    auth_required: true
    rate_limit:
      rps: 50
      burst: 100

tracing:
  enabled: true
  exporter: http://otel-collector:4317
```

## Observability

- Tracing — OpenTelemetry → Jaeger / Tempo / Zipkin
- Metrics — Prometheus → Grafana (request latency, error rate, upstream status)
- Logs — structured JSON → Loki / ELK / CloudWatch
