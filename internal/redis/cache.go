package redis

import (
	"context"
	"time"
)

var (
	ctx  = context.Background()
	main RedisCacher
)

type RedisCacher interface {
	Set(key string, value string, duration time.Duration) error
	Get(key string) (string, error)
}

func ResetCache() {
	main = nil //force to reinit main
}

func Cache() RedisCacher {
	if main == nil {
		main = NewRedis()
	}
	return main
}

func (redis *redis) Set(key string, value string, duration time.Duration) error {
	return redis.GetClient().Set(ctx, key, value, duration).Err()
}

func (redis *redis) Get(key string) (string, error) {
	return redis.GetClient().Get(ctx, key).Result()
}
