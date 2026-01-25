# Сборка
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем оба бинарника: сервис и мигратор
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/auth-service ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/migrator ./cmd/migrator

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем netcat для проверки доступности БД в entrypoint
RUN apk add --no-cache netcat-openbsd

# Копируем бинарники
COPY --from=builder /app/bin/ ./bin/
# Копируем папку с миграциями (важно, чтобы мигратор их нашел)
COPY --from=builder /app/migrations/ ./migrations/
# Копируем конфиги (если нужны дефолтные значения, хотя мы все берем из ENV)
COPY --from=builder /app/config/ ./config/

# Копируем entrypoint и даем права на выполнение
COPY entrypoint.sh .

RUN sed -i 's/\r$//' entrypoint.sh
RUN chmod +x entrypoint.sh

EXPOSE 8080

# Указываем, что сначала запускается скрипт
ENTRYPOINT ["/app/entrypoint.sh"]

# А это аргумент для entrypoint (само приложение)
CMD ["./bin/auth-service"]