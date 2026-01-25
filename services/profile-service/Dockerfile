# Сборка
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарники сервиса и мигратора
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/profile-service ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/migrator ./cmd/migrator

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем netcat для проверки связи с БД
RUN apk add --no-cache netcat-openbsd

# Копируем бинарники из этапа сборки
COPY --from=builder /app/bin/ ./bin/
# Копируем папку с миграциями
COPY --from=builder /app/migrations/ ./migrations/
# Копируем конфиги
COPY --from=builder /app/config/ ./config/

# Копируем и настраиваем entrypoint
COPY entrypoint.sh .
RUN sed -i 's/\r$//' entrypoint.sh
RUN chmod +x entrypoint.sh

EXPOSE 8081

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["./bin/profile-service"]