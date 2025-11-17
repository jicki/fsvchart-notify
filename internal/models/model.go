package models

import (
	"time"
)

type MetricsSource struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type FeishuWebhook struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PushStatus struct {
	ID          int64  `json:"id"`
	SourceID    int64  `json:"source_id"`
	WebhookID   int64  `json:"webhook_id"`
	LastRunTime string `json:"last_run_time"`
	Status      string `json:"status"`
	Log         string `json:"log"`
}

type DataPoint struct {
	Time     string  `json:"x"`
	UnixTime int64   `json:"-"` // 存储Unix时间戳
	Value    float64 `json:"y"`
	Type     string  `json:"type"`
}

// 新增：查询数据点结构体，用于支持多个查询
type QueryDataPoints struct {
	DataPoints []DataPoint `json:"data_points"`
	ChartType  string      `json:"chart_type"`
	ChartTitle string      `json:"chart_title"`
	Unit       string      `json:"unit"` // 每个查询的独立单位
}

// 新增：发送记录结构体
type SendRecord struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Webhook    string    `json:"webhook"`
	TaskName   string    `json:"task_name"`
	ButtonText string    `json:"button_text"`
	ButtonURL  string    `json:"button_url"`
}

// 新增：图表模板结构体
type ChartTemplate struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ChartType string `json:"chart_type"`
}

// PromQL 查询结构体
type PromQL struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`        // 查询名称
	Description string `json:"description"` // 查询描述
	Query       string `json:"query"`       // 查询语句
	Category    string `json:"category"`    // 查询分类
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}

// TaskSendTime 任务发送时间结构体
type TaskSendTime struct {
	ID       int64  `json:"id"`
	TaskID   int64  `json:"task_id"`
	Weekday  int    `json:"weekday"`   // 1-7 代表周一到周日
	SendTime string `json:"send_time"` // 格式如 "09:00"
}

// User 用户结构体
type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"` // 不输出到JSON
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

// ChangePasswordRequest 修改密码请求结构体
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateUserRequest 更新用户信息请求结构体
type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}
