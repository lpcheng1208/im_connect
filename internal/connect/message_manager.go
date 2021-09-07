package connect

import (
	"github.com/golang/protobuf/proto"
)

// StoreMessageAck 存储
func StoreMessageAck(c *Conn, requestId int64, message proto.Message) {
	c.MessageAck.Store(requestId, message)
}

// GetMessageAck 获取
func GetMessageAck(c *Conn, requestId int64) proto.Message {
	value, ok := c.MessageAck.Load(requestId)
	if ok {
		return value.(proto.Message)
	}
	return nil
}

// DeleteMessageAck 删除
func DeleteMessageAck(c *Conn, requestId int64) {
	c.MessageAck.Delete(requestId)
}
