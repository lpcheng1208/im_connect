package connect

import (
	"container/list"
	"encoding/json"
	"errors"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"im_connect/config"
	"im_connect/pkg/db"
	"im_connect/pkg/gn"
	"im_connect/pkg/logger"
	"im_connect/pkg/pb"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	CoonTypeTCP int8 = 1 // tcp连接
	ConnTypeWS  int8 = 2 // websocket连接
	KeyUserConn string = "Cache#UserConnAddr:"
)

type Conn struct {
	CoonType   int8            // 连接类型
	TCP        *gn.Conn        // tcp连接
	WSMutex    sync.Mutex      // WS写锁
	WS         *websocket.Conn // websocket连接
	UserId     string          // 用户id
	DeviceId   string          // 设备id
	RoomId     int64           // 订阅的房间ID
	Element    *list.Element   // 链表节点
	MessageAck sync.Map        // 待确认待消息
	ReSendData sync.Map        // 重发的次数
}

// Write 写入数据
func (c *Conn) Write(bytes []byte) error {
	if c.CoonType == CoonTypeTCP {
		return encoder.EncodeToWriter(c.TCP, bytes)
	} else if c.CoonType == ConnTypeWS {
		return c.WriteToWS(bytes)
	}
	logger.Logger.Error("unknown conn type", zap.Any("conn", c))
	return nil
}

func (c *Conn) WriteToWS(bytes []byte) error {
	c.WSMutex.Lock()
	defer c.WSMutex.Unlock()

	return c.WS.WriteMessage(websocket.BinaryMessage, bytes)
}

// Close 关闭
func (c *Conn) Close() error {
	// 取消设备和连接的对应关系
	if c.UserId != "" {
		DeleteConn(c.UserId)
	}

	if c.CoonType == CoonTypeTCP {
		return c.TCP.Close()
	} else if c.CoonType == ConnTypeWS {
		return c.WS.Close()
	}
	return nil
}

func (c *Conn) GetAddr() string {
	if c.CoonType == CoonTypeTCP {
		return c.TCP.GetAddr()
	} else if c.CoonType == ConnTypeWS {
		return c.WS.RemoteAddr().String()
	}
	return ""
}

func (c *Conn) HandleMessage(bytesData []byte) {
	var input pb.ImInput
	err := proto.Unmarshal(bytesData, &input)
	if err != nil {
		logger.Logger.Error("unmarshal error", zap.Error(err))
		return
	}

	logger.Logger.Debug("收到消息，进行分配处理", zap.Any("inputType", input.Type), zap.Int64("requestId", input.RequestId))
	// 对未登录的用户进行拦截
	if input.Type != pb.IMPackageType_IM_SIGN_IN && c.UserId == "" {
		// 应该告诉用户没有登录
		return
	}
	switch input.Type {
	case pb.IMPackageType_IM_SIGN_IN:
		c.SignIn(input)
	case pb.IMPackageType_IM_HEARTBEAT:
		c.Heartbeat(input)
	case pb.IMPackageType_IM_MESSAGE:
		c.MessageDeliver(input)
	case pb.IMPackageType_IM_MESSAGE_ACK:
		c.ImMessageACK(input)
	}
}

// Send 下发消息
func (c *Conn) Send(pt pb.IMPackageType, requestId int64, message proto.Message, err error) {
	var output = pb.ImOutput{
		Type:      pt,
		RequestId: requestId,
	}

	if err != nil {
		status, _ := status.FromError(err)
		output.Code = int32(status.Code())
		output.Message = status.Message()
	}

	if message != nil {
		msgBytes, err := proto.Marshal(message)
		if err != nil {
			logger.Sugar.Error(err)
			return
		}
		output.Data = msgBytes
	}

	outputBytes, err := proto.Marshal(&output)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}

	err = c.Write(outputBytes)
	if err != nil {
		logger.Sugar.Error(err)
		c.Close()
		return
	}
	if pt != pb.IMPackageType_IM_MESSAGE {
		return
	}
	// 存储待确认的消息
	StoreMessageAck(c, requestId, message)
}

// SignIn 登录R
func (c *Conn) SignIn(input pb.ImInput) {
	var signIn pb.ImSignInInput
	err := proto.Unmarshal(input.Data, &signIn)

	//if GetConn(signIn.UserId) != nil {
	//	DeleteConn(c.UserId)
	//	logger.Sugar.Error("用户已经连接成功")
	//}

	inputUserId := signIn.UserId
	message := &pb.ImSignInOutput{
		UserId:  inputUserId,
		Success: 0,
	}

	if err != nil {
		logger.Sugar.Error(err)
		message.Success = -1
		c.Send(pb.IMPackageType_IM_SIGN_IN, input.RequestId, message, err)
		return
	}
	userToken, _ := db.RedisCli.Get("user_token_" + inputUserId).Result()

	if signIn.Token != userToken {
		logger.Logger.Error("Im SignIn Error", zap.String("userId", inputUserId), zap.String("token", signIn.Token), zap.String("redisToken", userToken))
		message.Success = -1
		err = errors.New("token error")
		c.Send(pb.IMPackageType_IM_SIGN_IN, input.RequestId, message, err)
		return
	}

	logger.Logger.Debug("Im SignIn", zap.String("userId", signIn.UserId), zap.String("token", signIn.Token))
	c.Send(pb.IMPackageType_IM_SIGN_IN, input.RequestId, message, err)

	c.UserId = inputUserId
	c.DeviceId = signIn.DeviceId

	connOk := &OnEventData{
		MsgId: time.Now().UnixNano(),
		MsgBody: struct {
			UserId string `json:"user_id"`
			Type   int    `json:"type"`
		}{UserId: signIn.UserId, Type: 1},
	}

	bytes, _ := json.Marshal(connOk)

	db.RedisCli.Publish("message:on_message_for_im_message", bytes).Result()

	SetConn(signIn.UserId, c)
}

// Heartbeat 心跳
func (c *Conn) Heartbeat(input pb.ImInput) {
	//CacheUserStatus(c.UserId)
	c.Send(pb.IMPackageType_IM_HEARTBEAT, input.RequestId, nil, nil)
	key := "Cache#UserConnAddr:" + c.UserId
	db.RedisCli.Set(key, config.Connect.TCPListenAddr, time.Second * 35)
	logger.Sugar.Infow("心跳消息，维持长链接状态", "device_id", c.DeviceId, "user_id", c.UserId)
}

// ImMessageACK 消息收到回执
func (c *Conn) ImMessageACK(input pb.ImInput) {
	RequestId := input.RequestId
	DeleteMessageAck(c, RequestId)
	logger.Logger.Info("收到 客户端回包消息", zap.Int64("RequestId", RequestId))

}

func (c *Conn) MessageDeliver(input pb.ImInput) {

	var ImMessage pb.ImMessage

	err := proto.Unmarshal(input.Data, &ImMessage)
	if err != nil {
		logger.Sugar.Error(err.Error())
		return
	}

	// 发送消息回包
	c.Send(pb.IMPackageType_IM_MESSAGE_ACK, input.RequestId, nil, nil)

	MessageType := ImMessage.MessageType
	ReceiverId := ImMessage.ReceiverId

	dbAddr, _ := db.RedisCli.Get(KeyUserConn + ReceiverId).Result()
	localAdd := config.Connect.TCPListenAddr

	if dbAddr != localAdd {
		// todo 发送到redis 广播 进行消息投递
	} else {
		// todo 发送到本地连接
	}

	receiveConn := GetConn(ReceiverId)
	if receiveConn == nil {
		logger.Sugar.Debug("获取不到对方连接, 不进行转发")
		return
	}
	receiveConn.Send(pb.IMPackageType_IM_MESSAGE, time.Now().UnixNano(), &ImMessage, nil)
	logger.Logger.Info("收到普通消息", zap.Any("sendId", c.UserId), zap.Any("ReceiverId", ReceiverId), zap.Any("MessageType", MessageType), zap.Any("SceneType", ImMessage.SceneType))

}

func (c *Conn) CheckMessageResend() {
	var messageAck []int64
	c.MessageAck.Range(func(key, value interface{}) bool {
		message := value.(proto.Message)
		requestId := key.(int64)
		v, ok := c.ReSendData.Load(requestId)

		if time.Now().UnixNano() - requestId < 3 * 1000 * 1000 * 1000{
			return true
		}

		var num int32 = 1
		if ok {
			num = v.(int32)
			if num > 3 {
				logger.Logger.Info("消息重发已经超过三次， 丢弃该条消息", zap.Int64("requestId", requestId))
				c.ReSendData.Delete(requestId)
				DeleteMessageAck(c, requestId)
				return true
			}
		}
		num += 1
		c.ReSendData.Store(requestId, num)

		messageAck = append(messageAck, key.(int64))
		logger.Logger.Info("重发消息", zap.String("userId", c.UserId), zap.Int64("requestId", requestId))
		c.Send(pb.IMPackageType_IM_MESSAGE, key.(int64), message, nil)
		return true
	})
	if len(messageAck) != 0 {
		logger.Logger.Debug("RangeMessageAck", zap.String("userId", c.UserId), zap.Any("messageAck", messageAck))
	}
}

func sendToDevice(c *Conn, ImMessage pb.ImMessage)  {

	MessageType := ImMessage.MessageType
	ReceiverId := ImMessage.ReceiverId
	MessageContent := ImMessage.MessageContent

	switch MessageType {
	case pb.ImMessageType_IM_MT_COMMAND:
		var command pb.ImCommand
		err := proto.Unmarshal(MessageContent, &command)
		if err != nil {
			logger.Sugar.Error(err.Error())
			return
		}
		commandType := command.CommandCode
		commandData := command.CommandData
		switch commandType {
		case pb.ImCommandType_IM_COMMAND_CHAT_REQUEST:


		case pb.ImCommandType_IM_COMMAND_CHAT_OPERATE:
			var operate pb.CommandChatOperate
			err = proto.Unmarshal(commandData, &operate)
			if err != nil {
				logger.Sugar.Error(err.Error())
				return
			}
			action := operate.Action
			if action == 1{

			}
		}
	default:
		c.Send(pb.IMPackageType_IM_MESSAGE, time.Now().UnixNano(), &ImMessage, nil)
		logger.Logger.Info("收到普通消息", zap.Any("sendId", c.UserId), zap.Any("ReceiverId", ReceiverId), zap.Any("MessageType", MessageType), zap.Any("SceneType", ImMessage.SceneType))
	}
}
