package connect

import (
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"im_connect/config"
	"im_connect/pkg/db"
	"im_connect/pkg/logger"
	"im_connect/pkg/pb"
	"im_connect/pkg/topic"
	"time"
)

func StartSubscribe() {
	channel := db.RedisCli.Subscribe(topic.MessageForwardTopic).Channel()
	logger.Sugar.Debug("StartSubscribe poll with ", config.Connect.SubscribeNum)
	for i := 0; i < config.Connect.SubscribeNum; i++ {
		go handleMsg(channel)
	}
}

func handleMsg(channel <-chan *redis.Message) {
	for msg := range channel {
		if msg.Channel == topic.MessageForwardTopic {
			handleMessageDeliver([]byte(msg.Payload))
		}
	}
}

func handleMessageDeliver(bytes []byte) {
	var message pb.ImMessage

	err := proto.Unmarshal(bytes, &message)
	if err != nil {
		logger.Sugar.Error(err.Error())
		return
	}

	ReceiverId := message.ReceiverId

	otherConn := GetConn(ReceiverId)
	if otherConn != nil {
		otherConn.Send(pb.IMPackageType_IM_MESSAGE, time.Now().UnixNano(), &message, nil)
		logger.Logger.Info("收到普通消息", zap.Any("sendId", message.Sender.SenderId), zap.Any("ReceiverId", ReceiverId), zap.Any("MessageType", message.MessageType), zap.Any("SceneType", message.SceneType))
	} else {
		logger.Logger.Info("收到普通消息, 转发连接不存在")
	}

}