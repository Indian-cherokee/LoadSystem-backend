// internal/app/redis/redis.go
package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"web/internal/app/config"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const servicePrefix = "Load_System."
const jwtPrefix = "jwt."
const sessionPrefix = "session:"

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		Username: cfg.User,
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{client: redisClient, cfg: cfg}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// МЕТОДЫ ДЛЯ РАБОТЫ С JWT BLACKLIST

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	key := getJWTKey(jwtStr)
	err := c.client.Set(ctx, key, true, jwtTTL).Err()
	if err != nil {
		logrus.Errorf("Failed to write JWT to blacklist: %v", err)
		return err
	}
	logrus.Infof("JWT token added to blacklist: %s (TTL: %v)", key, jwtTTL)
	return nil
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Get(ctx, getJWTKey(jwtStr)).Err()
}

// МЕТОДЫ ДЛЯ РАБОТЫ С СЕССИЯМИ

func getSessionKey(sessionID string) string {
	return sessionPrefix + sessionID
}

// CreateSession создает новую сессию в Redis
// sessionID - UUID сессии, userID - ID пользователя, ttl - время жизни сессии
func (c *Client) CreateSession(ctx context.Context, sessionID string, userID uint, ttl time.Duration) error {
	key := getSessionKey(sessionID)
	err := c.client.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		logrus.Errorf("Failed to create session: %v", err)
		return err
	}
	logrus.Infof("Session created: %s (userID: %d, TTL: %v)", key, userID, ttl)
	return nil
}

// DeleteSessionByUserID удаляет все сессии пользователя из Redis
// Ищет все ключи session:* со значением userID и удаляет их
func (c *Client) DeleteSessionByUserID(ctx context.Context, userID uint) error {
	// Ищем все ключи session:*
	keys, err := c.client.Keys(ctx, sessionPrefix+"*").Result()
	if err != nil {
		logrus.Errorf("Failed to get session keys: %v", err)
		return err
	}

	deleted := 0
	for _, key := range keys {
		val, err := c.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		// Проверяем, что значение соответствует userID
		valUint, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			continue
		}
		if uint(valUint) == userID {
			if err := c.client.Del(ctx, key).Err(); err != nil {
				logrus.Errorf("Failed to delete session key %s: %v", key, err)
			} else {
				deleted++
				logrus.Infof("Session deleted: %s (userID: %d)", key, userID)
			}
		}
	}

	if deleted == 0 {
		logrus.Warnf("No sessions found for userID: %d", userID)
	} else {
		logrus.Infof("Deleted %d session(s) for userID: %d", deleted, userID)
	}
	return nil
}

// GetSession получает userID по sessionID
func (c *Client) GetSession(ctx context.Context, sessionID string) (uint, error) {
	key := getSessionKey(sessionID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("session not found: %w", err)
	}
	userID, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid session userID: %w", err)
	}
	return uint(userID), nil
}
