package connect

import (
	"encoding/json"
	"im_connect/config"
	"im_connect/pkg/db"
	"im_connect/pkg/logger"
	"time"

	"go.uber.org/zap"

	"im_connect/pkg/gn"
)

var encoder = gn.NewHeaderLenEncoder(2, 1024)

var server *gn.Server

func StartTCPServer() {
	var err error
	server, err = gn.NewServer(config.Connect.TCPListenAddr, &handler{},
		gn.NewHeaderLenDecoder(2),
		gn.WithReadBufferLen(256),
		gn.WithTimeout(30*time.Second, 60*time.Second),
		gn.WithAcceptGNum(10),
		gn.WithIOGNum(100),
		gn.WithMsgResendTime(3*time.Second, 60*time.Second))
	if err != nil {
		logger.Sugar.Error(err)
		panic(err)
	}
	server.Run()
}

type handler struct{}

var Handler = new(handler)

func (*handler) OnConnect(c *gn.Conn) {
	// 初始化连接数据
	conn := &Conn{
		CoonType: CoonTypeTCP,
		TCP:      c,
	}
	c.SetData(conn)
	logger.Logger.Debug("connect:", zap.Int32("fd", c.GetFd()), zap.String("addr", c.GetAddr()))
}

func (h *handler) OnMessage(c *gn.Conn, bytes []byte) {
	conn := c.GetData().(*Conn)
	conn.HandleMessage(bytes)
}

func (*handler) OnClose(c *gn.Conn, err error) {
	conn := c.GetData().(*Conn)

	gFd := conn.TCP.GetFd()

	nowConn := GetConn(conn.UserId)
	if nowConn == nil {
		return
	}

	logger.Logger.Debug("close", zap.String("addr", c.GetAddr()), zap.String("user_id", conn.UserId), zap.String("device_id", conn.UserId), zap.Int32("FD", conn.TCP.GetFd()), zap.Error(err))

	nowFd := nowConn.TCP.GetFd()
	if nowFd != gFd {
		logger.Logger.Debug("当前链接的FD 与 gn 的不匹配, 不进行本地链接删除", zap.Int32("gnFd", gFd), zap.Int32("nowFd", nowFd))
		return
	}

	connClose := &OnEventData{
		MsgId: time.Now().UnixNano(),
		MsgBody: struct {
			UserId string `json:"user_id"`
			Type   int    `json:"type"`
		}{UserId: conn.UserId, Type: 2},
	}

	bytes, _ := json.Marshal(connClose)

	db.RedisCli.Publish("message:on_message_for_im_message", bytes).Result()
	DeleteConn(conn.UserId)
}

func (h *handler) OnCheckMessage(c *gn.Conn) {
	conn := c.GetData().(*Conn)
	conn.CheckMessageResend()
}
