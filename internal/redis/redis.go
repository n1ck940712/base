package redis

import (
	"strings"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	r "github.com/go-redis/redis/v8"
)

var (
	useTestRedis = false
)

type Redis interface {
	RedisCacher
	GetClient() *r.Client
}

type redis struct {
	client *r.Client
}

func UseLocalhost() {
	useTestRedis = true
}

func addr() string {
	if useTestRedis {
		return strings.Replace(settings.REDIS_HOST, "redis:", "localhost:", 1) //TODO: support for localhost
	}
	return settings.REDIS_HOST
}

func NewRedis() Redis {
	client := r.NewClient(&r.Options{
		Addr: addr(),
		DB:   0,
	})

	return &redis{client: client}
}

func (r *redis) GetClient() *r.Client {
	return r.client
}
