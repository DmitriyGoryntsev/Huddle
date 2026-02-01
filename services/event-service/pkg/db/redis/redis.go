package redis

import (
	"context"
	"event-service/internal/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type Client struct {
	inner  *redis.Client
	logger *zap.SugaredLogger
	cb     *gobreaker.CircuitBreaker
}

// Inner — возвращает оригинальный клиент Redis
func (c *Client) Inner() *redis.Client {
	return c.inner
}

func NewClient(cfg *config.RedisConfig, logger *zap.SugaredLogger) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	// Создаем circuit breaker с теми же настройками
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Infof("%s circuit breaker state changed: %s -> %s", name, from.String(), to.String())
		},
	})

	// Повторные попытки подключения
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		// Создаем клиент Redis
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		// Пробуем подключиться
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err == nil {
			logger.Infof("Connected to Redis on attempt %d", attempt)
			return &Client{
				inner:  client,
				logger: logger,
				cb:     cb,
			}, nil
		}

		logger.Warnf("Redis connection failed on attempt %d, retrying in %ds", attempt, cfg.RetryDelay)
		client.Close()

		if attempt < cfg.MaxRetries {
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", cfg.MaxRetries)
}

func (c *Client) Close() {
	if c.inner != nil {
		c.inner.Close()
	}
}

// Ping — проверка соединения с circuit breaker
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, c.inner.Ping(ctx).Err()
	})
	return err
}

// Get — обернутый метод с circuit breaker
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.Get(ctx, key).Result()
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Set — обернутый метод с circuit breaker
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, c.inner.Set(ctx, key, value, ttl).Err()
	})
	return err
}

// Exists — обернутый метод с circuit breaker
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.Exists(ctx, keys...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Del — обернутый метод с circuit breaker
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.Del(ctx, keys...).Result()
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// методы для TokenService

func (c *Client) SetJTI(ctx context.Context, jti, userID string, ttl time.Duration) error {
	return c.Set(ctx, "jti:"+jti, userID, ttl)
}

func (c *Client) GetUserIDByJTI(ctx context.Context, jti string) (string, error) {
	val, err := c.Get(ctx, "jti:"+jti)
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *Client) RevokeJTI(ctx context.Context, jti string, ttl time.Duration) error {
	return c.Set(ctx, "revoked:"+jti, "1", ttl)
}

func (c *Client) IsRevoked(ctx context.Context, jti string) (bool, error) {
	count, err := c.Exists(ctx, "revoked:"+jti)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// Execute — универсальный метод для любых операций с circuit breaker
func (c *Client) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(fn)
}

// Pipeline — поддержка пайплайнов Redis
func (c *Client) Pipeline(ctx context.Context) (redis.Pipeliner, error) {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.Pipeline(), nil
	})
	if err != nil {
		return nil, err
	}
	return c.inner.Pipeline(), nil
}

// TxPipeline — поддержка транзакционных пайплайнов
func (c *Client) TxPipeline(ctx context.Context) (redis.Pipeliner, error) {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return c.inner.TxPipeline(), nil
	})
	if err != nil {
		return nil, err
	}
	return c.inner.TxPipeline(), nil
}
