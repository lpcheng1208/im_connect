package connect

import (
	"go.uber.org/zap"
	"im_connect/pkg/logger"
	"im_connect/pkg/pb"
	"sync"
)

var ConnsManager = sync.Map{}

// SetConn 存储
func SetConn(UserId string, conn *Conn) {
	ConnsManager.Store(UserId, conn)
}

// GetConn 获取
func GetConn(UserId string) *Conn {
	value, ok := ConnsManager.Load(UserId)
	if ok {
		return value.(*Conn)
	}
	return nil
}

// DeleteConn 删除
func DeleteConn(UserId string) {
	// 删除用户的在线状态
	DeleteUserStatus(UserId)
	ConnsManager.Delete(UserId)
}

func PushAll(message *pb.ImMessage) {
	ConnsManager.Range(func(key, value interface{}) bool {
		conn := value.(*Conn)
		conn.Send(pb.IMPackageType_IM_MESSAGE, 0, message, nil)
		return true
	})
}

func GetAllConn() {
	var allUser []interface{}
	ConnsManager.Range(func(key, value interface{}) bool {
		allUser = append(allUser, key)
		return true
	})
	logger.Logger.Debug("All Conns", zap.Any("allUsers", allUser))

}
