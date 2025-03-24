package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run tools/data_sync.go <数据库路径>")
		fmt.Println("示例: go run tools/data_sync.go ./data/app.db")
		os.Exit(1)
	}

	dbPath := os.Args[1]

	log.Printf("=== 数据同步工具 ===")
	log.Printf("数据库路径: %s", dbPath)

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查表结构
	log.Println("检查表结构...")
	checkTableStructure(db, "push_task")

	// 检查数据一致性
	log.Println("\n检查数据一致性...")
	syncPushTaskData(db)

	log.Println("数据同步完成!")
}

func checkTableStructure(db *sql.DB, tableName string) {
	// 获取表的结构信息
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		log.Fatalf("获取表结构失败: %v", err)
	}
	defer rows.Close()

	fmt.Printf("表 %s 结构:\n", tableName)
	hasCustomMetricLabel := false

	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value sql.NullString
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			log.Fatalf("读取列信息失败: %v", err)
		}

		defaultValue := "NULL"
		if dflt_value.Valid {
			defaultValue = dflt_value.String
		}

		fmt.Printf("  列 %d: %s (%s), NOT NULL=%d, DEFAULT=%s, PK=%d\n",
			cid, name, type_, notnull, defaultValue, pk)

		if name == "custom_metric_label" {
			hasCustomMetricLabel = true
		}
	}

	// 检查是否缺少必要的列
	if !hasCustomMetricLabel {
		log.Println("\n需要添加custom_metric_label列")
		_, err := db.Exec("ALTER TABLE push_task ADD COLUMN custom_metric_label TEXT DEFAULT ''")
		if err != nil {
			log.Fatalf("添加custom_metric_label列失败: %v", err)
		}
		log.Println("已添加custom_metric_label列")
	}
}

func syncPushTaskData(db *sql.DB) {
	// 获取任务总数
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM push_task").Scan(&count)
	if err != nil {
		log.Fatalf("获取任务总数失败: %v", err)
	}
	log.Printf("共有 %d 个任务", count)

	// 检查并修复空值
	fixNullValues(db)

	// 检查与其他表的关联数据
	checkTaskRelations(db)
}

func fixNullValues(db *sql.DB) {
	// 检查并修复NULL值
	log.Println("检查NULL值...")

	// 更新custom_metric_label为空的行
	result, err := db.Exec(`
		UPDATE push_task 
		SET custom_metric_label = '' 
		WHERE custom_metric_label IS NULL
	`)
	if err != nil {
		log.Fatalf("修复custom_metric_label NULL值失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %d 行中的NULL值", rowsAffected)

	// 其他可能需要修复的列
	columnsToFix := []string{"name", "query", "time_range", "card_title", "card_template", "metric_label", "unit", "initial_send_time"}

	for _, column := range columnsToFix {
		result, err := db.Exec(fmt.Sprintf(`
			UPDATE push_task 
			SET %s = COALESCE(%s, '') 
			WHERE %s IS NULL
		`, column, column, column))

		if err != nil {
			log.Printf("修复 %s NULL值失败: %v", column, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("已修复 %d 行的 %s 列中的NULL值", rowsAffected, column)
		}
	}
}

func checkTaskRelations(db *sql.DB) {
	// 检查任务与webhook的关联
	log.Println("\n检查任务与webhook的关联...")
	rows, err := db.Query(`
		SELECT p.id, p.name, COUNT(ptw.webhook_id) as webhook_count
		FROM push_task p
		LEFT JOIN push_task_webhook ptw ON p.id = ptw.task_id
		GROUP BY p.id
	`)
	if err != nil {
		log.Fatalf("检查任务与webhook关联失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		var webhookCount int
		if err := rows.Scan(&id, &name, &webhookCount); err != nil {
			log.Fatalf("读取任务webhook关联数据失败: %v", err)
		}

		if webhookCount == 0 {
			log.Printf("警告: 任务 [%d] %s 没有关联任何webhook", id, name)
		} else {
			log.Printf("任务 [%d] %s 关联了 %d 个webhook", id, name, webhookCount)
		}
	}

	// 检查孤立的push_task_webhook记录
	var orphanedCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM push_task_webhook ptw
		LEFT JOIN push_task pt ON ptw.task_id = pt.id
		WHERE pt.id IS NULL
	`).Scan(&orphanedCount)

	if err != nil {
		log.Printf("检查孤立的push_task_webhook记录失败: %v", err)
	} else if orphanedCount > 0 {
		log.Printf("发现 %d 条孤立的push_task_webhook记录", orphanedCount)

		// 清理孤立记录
		result, err := db.Exec(`
			DELETE FROM push_task_webhook
			WHERE task_id NOT IN (SELECT id FROM push_task)
		`)
		if err != nil {
			log.Printf("清理孤立的push_task_webhook记录失败: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("已清理 %d 条孤立的push_task_webhook记录", rowsAffected)
		}
	}

	// 补充最后运行时间
	var nullLastRunCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task WHERE last_run_at IS NULL").Scan(&nullLastRunCount)

	if err != nil {
		log.Printf("检查last_run_at为NULL的记录失败: %v", err)
	} else if nullLastRunCount > 0 {
		log.Printf("发现 %d 条记录的last_run_at为NULL", nullLastRunCount)

		// 更新为当前时间
		now := time.Now().Format("2006-01-02 15:04:05")
		result, err := db.Exec(`
			UPDATE push_task 
			SET last_run_at = ? 
			WHERE last_run_at IS NULL
		`, now)

		if err != nil {
			log.Printf("更新last_run_at失败: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("已更新 %d 条记录的last_run_at为 %s", rowsAffected, now)
		}
	}
}
