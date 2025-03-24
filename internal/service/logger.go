package service

import (
	"container/ring"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type LogManager struct {
	logs  *ring.Ring
	mutex sync.RWMutex
}

var (
	logManager *LogManager
	once       sync.Once
)

// GetLogManager 返回单例的日志管理器
func GetLogManager() *LogManager {
	once.Do(func() {
		logManager = &LogManager{
			logs: ring.New(1000), // 保存最近1000条日志
		}
	})
	return logManager
}

// AddLog 添加一条日志
func (lm *LogManager) AddLog(msg string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
	}
	lm.logs.Value = entry
	lm.logs = lm.logs.Next()
}

// GetLogs 获取所有日志
func (lm *LogManager) GetLogs() []LogEntry {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var entries []LogEntry
	lm.logs.Do(func(v interface{}) {
		if v != nil {
			entries = append(entries, v.(LogEntry))
		}
	})
	return entries
} 