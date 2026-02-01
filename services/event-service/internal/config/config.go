package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServerConfig struct {
	Port       string `env:"HTTP_SERVER_PORT" env-default:"8080" validate:"required,numeric"`
	MaxRetries int    `env:"HTTP_SERVER_MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `env:"HTTP_SERVER_RETRY_DELAY" env-default:"5" validate:"gte=1"`
}

type PostgresConfig struct {
	Host       string `env:"POSTGRES_HOST" validate:"required"`
	Port       string `env:"POSTGRES_PORT" env-default:"5432" validate:"numeric"`
	User       string `env:"POSTGRES_USER" validate:"required"`
	Password   string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DBNAME" validate:"required"`
	SSLMode    string `env:"POSTGRES_SSLMODE" env-default:"disable"`
	MaxConns   int32  `env:"POSTGRES_MAX_CONNS" env-default:"50"`
	MinConns   int32  `env:"POSTGRES_MIN_CONNS" env-default:"10"`
	Timeout    int    `env:"POSTGRES_TIMEOUT" env-default:"5"`
	MaxRetries int    `env:"POSTGRES_MAX_RETRIES" env-default:"5"`
	RetryDelay int    `env:"POSTGRES_RETRY_DELAY" env-default:"2"`
}

type LoggerConfig struct {
	Level            string   `env:"LOG_LEVEL" env-default:"info" validate:"oneof=debug info warn error dpanic panic fatal"`
	Format           string   `env:"LOG_FORMAT" env-default:"console" validate:"oneof=console json"`
	OutputPaths      []string `env:"LOG_OUTPUT_PATHS" env-default:"stdout" env-separator:","`
	ErrorOutputPaths []string `env:"LOG_ERROR_OUTPUT_PATHS" env-default:"stderr" env-separator:","`
	Development      bool     `env:"LOG_DEVELOPMENT" env-default:"false"`
	EnableStacktrace bool     `env:"LOG_ENABLE_STACKTRACE" env-default:"false"`
	TimeFormat       string   `env:"LOG_TIME_FORMAT" env-default:"iso8601" validate:"oneof=iso8601 rfc3339 epoch millis"`
}

type RedisConfig struct {
	Host       string `env:"REDIS_HOST" env-default:"localhost" validate:"required"`
	Port       string `env:"REDIS_PORT" env-default:"6379" validate:"required,numeric"`
	Password   string `env:"REDIS_PASSWORD" env-default:""`
	DB         int    `env:"REDIS_DB" env-default:"0" validate:"gte=0,lte=15"`
	MaxRetries int    `env:"REDIS_MAX_RETRIES" env-default:"5" validate:"gte=1"`
	RetryDelay int    `env:"REDIS_RETRY_DELAY" env-default:"2" validate:"gte=1"`
	Timeout    int    `env:"REDIS_TIMEOUT" env-default:"5" validate:"gte=1"`
}

type Config struct {
	Env        string `env:"ENV" env-default:"development" validate:"oneof=development production"`
	HTTPServer HTTPServerConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	Logger     LoggerConfig
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
