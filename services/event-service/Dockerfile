# Сборка
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/event-service ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/migrator ./cmd/migrator

# Финальный образ
FROM alpine:3.18
WORKDIR /app
# Устанавливаем netcat (обязательно для entrypoint.sh)
RUN apk add --no-cache netcat-openbsd

COPY --from=builder /app/bin/ ./bin/
COPY --from=builder /app/migrations/ ./migrations/
COPY entrypoint.sh .

# Исправляем символы конца строки (на случай Windows) и даем права
RUN sed -i 's/\r$//' entrypoint.sh
RUN chmod +x entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
# Передаем путь к бинарнику в качестве аргумента скрипту
CMD ["/app/bin/event-service"]