package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"telegram-bot/internal/bot/states"
	"telegram-bot/internal/config"
)

type Cache struct {
	client *redis.Client
}

func NewCache(cfg *config.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &Cache{client: client}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) GetUserState(userID int64) (*states.UserStateData, error) {
	key := fmt.Sprintf("user_state:%d", userID)
	data, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var state states.UserStateData
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (c *Cache) SetUserState(userID int64, state *states.UserStateData, ttl time.Duration) error {
	key := fmt.Sprintf("user_state:%d", userID)
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

func (c *Cache) DeleteUserState(userID int64) error {
	key := fmt.Sprintf("user_state:%d", userID)
	return c.client.Del(context.Background(), key).Err()
}

func (c *Cache) GetCachedPayment(userID int64) ([]byte, error) {
	key := fmt.Sprintf("payment:%d", userID)
	return c.client.Get(context.Background(), key).Bytes()
}

func (c *Cache) SetCachedPayment(userID int64, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("payment:%d", userID)
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

func (c *Cache) GetCachedResumes(userID int64) ([]byte, error) {
	key := fmt.Sprintf("resumes:%d", userID)
	return c.client.Get(context.Background(), key).Bytes()
}

func (c *Cache) SetCachedResumes(userID int64, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("resumes:%d", userID)
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

func (c *Cache) InvalidateResumesCache(userID int64) error {
	key := fmt.Sprintf("resumes:%d", userID)
	return c.client.Del(context.Background(), key).Err()
}

func (c *Cache) GetCachedStats(userID int64) ([]byte, error) {
	key := fmt.Sprintf("stats:%d", userID)
	return c.client.Get(context.Background(), key).Bytes()
}

func (c *Cache) SetCachedStats(userID int64, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("stats:%d", userID)
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

func (c *Cache) SetRateLimit(userID int64, action string, limit int, window time.Duration) error {
	key := fmt.Sprintf("ratelimit:%d:%s", userID, action)
	count, err := c.client.Incr(context.Background(), key).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		c.client.Expire(context.Background(), key, window)
	}

	if count > int64(limit) {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (c *Cache) GetScheduledJobs() (map[string]bool, error) {
	key := "scheduled_jobs"
	data, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return make(map[string]bool), nil
		}
		return nil, err
	}

	var jobs map[string]bool
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *Cache) SetScheduledJobs(jobs map[string]bool) error {
	key := "scheduled_jobs"
	data, err := json.Marshal(jobs)
	if err != nil {
		return err
	}
	return c.client.Set(context.Background(), key, data, 24*time.Hour).Err()
}
