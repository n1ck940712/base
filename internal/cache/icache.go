package cache

import (
	"encoding/json"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var Main = NewCache() //static redis cache

type RedisCache struct {
	cache    *redis.Client
	gin      *gin.Context
	settings *settings.Settings
}

type ICache interface {
	Set(key string, value interface{}, timeout time.Duration)
	Get(key string) (interface{}, error)
	Touch(key string, timeout *uint)
	Del(key string)
	GetOrig(key string) (string, error)
}

func NewCache() *RedisCache {
	cache := redis.NewClient(&redis.Options{
		Addr: settings.REDIS_HOST,
		DB:   0,
	})

	return &RedisCache{
		cache,
		nil,
		nil,
	}
}

func (c *RedisCache) Get(key string) (interface{}, error) {
	res, err := c.cache.Get(c.gin, key).Result()
	var result interface{}
	json.Unmarshal([]byte(res), &result)

	return result, err
}

func (c *RedisCache) Set(key string, value interface{}, timeout time.Duration) {
	_ = c.cache.Set(c.gin, key, value, timeout).Err()
}

func (c *RedisCache) Touch(key string, timeout *uint) {
	c.cache.Touch(c.gin, key)
}

func (c *RedisCache) Del(key string) {
	c.cache.Del(c.gin, key)
}

func (c *RedisCache) GetOrig(key string) (string, error) {
	res, err := c.cache.Get(c.gin, key).Result()

	return res, err
}
