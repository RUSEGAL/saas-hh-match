package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

func Init() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := Client.Ping(ctx).Result()
	return err
}

func Get(key string, dest interface{}) error {
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Client.Set(ctx, key, data, ttl).Err()
}

func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}

func DeletePattern(pattern string) error {
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return Client.Del(ctx, keys...).Err()
	}
	return nil
}
