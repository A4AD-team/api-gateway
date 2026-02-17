# Dockerfile
# Этап 1: Сборка
FROM golang:1.21-alpine AS builder

# Установка необходимых системных пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/api-gateway ./main.go

# Этап 2: Финальный образ
FROM alpine:latest

# Копирование сертификатов и временной зоны из builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Создание непривилегированного пользователя
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Копирование бинарного файла
COPY --from=builder --chown=appuser:appgroup /app/api-gateway /app/api-gateway

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Запуск приложения
ENTRYPOINT ["/app/api-gateway"]