package util

import (
	"fmt"
	"runtime"
	"im_connect/pkg/logger"

	"go.uber.org/zap"
)

// RecoverPanic 恢复panic
func RecoverPanic() {
	err := recover()
	if err != nil {
		logger.Logger.DPanic("panic", zap.Any("panic", err), zap.String("stack", GetStackInfo()))
	}
}

// GetStackInfo 获取Panic堆栈信息
func GetStackInfo() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return fmt.Sprintf("%s", buf[:n])
}
