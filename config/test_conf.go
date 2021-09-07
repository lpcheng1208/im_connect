package config

import (
	"im_connect/pkg/logger"

	"go.uber.org/zap"
)

func initTestConf() {

	Connect = ConnectConf{
		TCPListenAddr: ":9002",
		WSListenAddr:  ":8081",
		RedisIP:       "testhost:6379",
		RedisPassword: "",
		SubscribeNum:  100,
	}

	logger.Level = zap.DebugLevel
	logger.Target = logger.File
	logger.MaxSize = 5
	logger.MaxBackups = 10
	logger.MaxAge = 7
}
