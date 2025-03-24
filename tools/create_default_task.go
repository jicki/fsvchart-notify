package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./data/app.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	log.Printf("=== 创建默认任务工具 ===")
	log.Printf("数据库路径: %s", dbPath)

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查是否已有任务
	var taskCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task").Scan(&taskCount)
	if err != nil {
		log.Fatalf("查询任务数量失败: %v", err)
	}

	if taskCount > 0 {
		log.Printf("系统中已有 %d 个任务，无需创建默认任务", taskCount)
		return
	}

	// 查找默认数据源
	var sourceID int64
	err = db.QueryRow("SELECT id FROM metrics_source LIMIT 1").Scan(&sourceID)
	if err != nil {
		log.Printf("未找到默认数据源，将创建一个: %v", err)

		// 创建默认数据源
		result, err := db.Exec(
			"INSERT INTO metrics_source (name, url) VALUES (?, ?)",
			"默认数据源", "http://localhost:9090",
		)
		if err != nil {
			log.Fatalf("创建默认数据源失败: %v", err)
		}

		sourceID, _ = result.LastInsertId()
		log.Printf("已创建默认数据源，ID: %d", sourceID)
	}

	// 查找默认Webhook
	var webhookID int64
	err = db.QueryRow("SELECT id FROM feishu_webhook LIMIT 1").Scan(&webhookID)
	if err != nil {
		log.Printf("未找到默认Webhook，将创建一个: %v", err)

		// 创建默认Webhook
		result, err := db.Exec(
			"INSERT INTO feishu_webhook (name, url) VALUES (?, ?)",
			"默认Webhook", "https://open.feishu.cn/webhook/测试占位链接",
		)
		if err != nil {
			log.Fatalf("创建默认Webhook失败: %v", err)
		}

		webhookID, _ = result.LastInsertId()
		log.Printf("已创建默认Webhook，ID: %d", webhookID)
	}

	// 创建默认任务
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.Exec(`
		INSERT INTO push_task (
			name, source_id, query, time_range, step, 
			schedule_interval, last_run_at, chart_template_id, 
			card_title, card_template, metric_label, unit, 
			enabled, initial_send_time, custom_chart_label, custom_metric_label
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		"示例任务", sourceID, "up", "30m", 300,
		3600, now, 1,
		"系统监控", "blue", "instance", "%",
		1, "08:00", "", "",
	)

	if err != nil {
		log.Fatalf("创建默认任务失败: %v", err)
	}

	taskID, _ := result.LastInsertId()
	log.Printf("已创建默认任务，ID: %d", taskID)

	// 关联任务和Webhook
	_, err = db.Exec(
		"INSERT INTO push_task_webhook (task_id, webhook_id) VALUES (?, ?)",
		taskID, webhookID,
	)

	if err != nil {
		log.Printf("关联任务和Webhook失败: %v", err)
	} else {
		log.Printf("已关联任务和Webhook")
	}

	log.Printf("默认任务创建完成！")
}
