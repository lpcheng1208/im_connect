package connect

import (
	"go.uber.org/zap"
	"im_connect/pkg/db"
	"im_connect/pkg/logger"
	"time"
)

type OnEventData struct {
	MsgId   int64 `json:"msg_id"`
	MsgBody struct {
		UserId string `json:"user_id"`
		Type   int    `json:"type"`
	} `json:"msg_body"`
}


func CacheUserStatus(userId string) {
	userStatusKey := "user_status:" + userId
	timeAdd := time.Second * 35
	exists, _ := db.RedisCli.Exists(userStatusKey).Result()
	if exists != 0 {
		db.RedisCli.Expire(userStatusKey, timeAdd)
	} else {
		_, err := db.RedisCli.Set(userStatusKey, "1", timeAdd).Result()
		if err != nil {
			logger.Sugar.Info("存储失败")
		}
	}
}

func DeleteUserStatus(userId string) {
	logger.Logger.Info("用户断开连接，删除用户在线状态", zap.String("userId", userId))
}
