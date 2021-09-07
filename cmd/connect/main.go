package main

import (
	"im_connect/config"
	"im_connect/internal/connect"
	"im_connect/pkg/db"
	"im_connect/pkg/logger"
)

func main() {
	logger.Init()

	db.InitRedis(config.Connect.RedisIP, config.Connect.RedisPassword)

	defer db.RedisCli.Close()

	connect.StartSubscribe()

	connect.StartTCPServer()

}
