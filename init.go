package cache

import (
	"net"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type CacheClient struct {
	redis      *redis.Client
	appName    string
	expiration time.Duration
}

func NewCacheClient(cfg RedisConfig, appName string, expiration time.Duration) CacheClient {

	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	cc := CacheClient{
		redis:      rdb,
		appName:    appName,
		expiration: expiration,
	}

	return cc
}

type ErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

var HTTPInternalError = &ErrorResponse{
	Status: 500,
	Error:  "Internal Server Error",
}
