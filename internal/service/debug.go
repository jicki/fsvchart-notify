package service

import "fmt"

var debugLog bool

// debugf 输出调试日志
func debugf(format string, args ...interface{}) {
	if debugLog {
		fmt.Printf(format+"\n", args...)
	}
}

// SetDebugLog 设置是否启用调试日志
func SetDebugLog(enable bool) {
	debugLog = enable
} 