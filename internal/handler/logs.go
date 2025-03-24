package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"fsvchart-notify/internal/service"
)

func HandleGetRunLogs(w http.ResponseWriter, r *http.Request) {
	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理 OPTIONS 请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只允许 GET 请求
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs := service.GetLogManager().GetLogs()
	
	// 格式化日志消息
	var formattedLogs []string
	for _, entry := range logs {
		formattedLogs = append(formattedLogs, 
			fmt.Sprintf("%s %s", 
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.Message))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(formattedLogs); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
} 