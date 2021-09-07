package db

import (
	"go.uber.org/zap"
	"im_connect/pkg/logger"

	"github.com/go-redis/redis"
)

var (
	RedisCli *redis.Client
)

// InitRedis 初始化Redis
func InitRedis(addr, password string) {
	logger.Logger.Info("init redis", zap.String("addr", addr))
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: password,
	})

	_, err := RedisCli.Ping().Result()
	if err != nil {
		panic(err)
	}

	logger.Logger.Info("init redis ok")
}
