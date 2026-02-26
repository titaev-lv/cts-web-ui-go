# Multi-stage build для минимального размера
FROM golang:1.24-alpine AS builder

# Разрешить автоматическую загрузку нужной версии Go
ENV GOTOOLCHAIN=auto

# Установить зависимости для сборки
RUN apk add --no-cache git make ca-certificates

WORKDIR /build

# Скопировать go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Скопировать исходный код
COPY . .

# Собрать бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o web-ui \
    cmd/web/main.go

# Final stage - минимальный образ
FROM alpine:latest

# Установить ca-certificates для HTTPS и wget для healthcheck
RUN apk --no-cache add ca-certificates tzdata wget

# Создать non-root пользователя
RUN addgroup -S webui && adduser -S webui -G webui

WORKDIR /app

# Скопировать бинарник из builder stage
COPY --from=builder /build/web-ui .

# Скопировать статические файлы и шаблоны
COPY --chown=webui:webui web/ ./web/

# Скопировать дефолтный конфиг (реальный в docker-compose обычно монтируется через volume)
COPY --chown=webui:webui config/config.proxy.yaml ./config/config.yaml

# Создать необходимые директории
RUN mkdir -p logs pki/ca pki/mysql config && \
    chown -R webui:webui logs pki config

# Переключиться на non-root пользователя
USER webui

# Expose порт для Web UI
EXPOSE 80

# Запуск приложения
ENTRYPOINT ["./web-ui"]
