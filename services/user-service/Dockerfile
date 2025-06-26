# Dockerfile для User Service
FROM golang:1.23.2-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates

# Рабочая директория
WORKDIR /app

# Копирование go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service ./cmd/server

# Финальный образ
FROM alpine:latest

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копирование бинарного файла
COPY --from=builder /app/user-service .

# Порт для сервиса
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск сервиса
CMD ["./user-service"]