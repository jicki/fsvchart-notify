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
	if len(os.Args) != 3 {
		fmt.Println("使用方法: go run tools/db_maintenance.go <命令> <数据库路径>")
		fmt.Println("命令:")
		fmt.Println("  fix-null     - 修复数据库中的NULL值")
		fmt.Println("  check        - 检查数据库结构和数据")
		fmt.Println("  cleanup      - 清理无效数据和孤立记录")
		fmt.Println("  repair-table - 修复表结构")
		fmt.Println("  all          - 执行所有维护操作")
		fmt.Println("示例: go run tools/db_maintenance.go all ./data/app.db")
		os.Exit(1)
	}

	command := os.Args[1]
	dbPath := os.Args[2]

	log.Printf("=== 数据库维护工具 ===")
	log.Printf("数据库路径: %s", dbPath)
	log.Printf("执行命令: %s", command)

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 备份数据库
	backupDB(dbPath)

	switch command {
	case "fix-null":
		fixNullValues(db)
	case "check":
		checkDatabase(db)
	case "cleanup":
		cleanupDatabase(db)
	case "repair-table":
		repairTables(db)
	case "all":
		checkDatabase(db)
		repairTables(db)
		fixNullValues(db)
		cleanupDatabase(db)
		validateData(db)
	default:
		log.Fatalf("未知命令: %s", command)
	}

	log.Println("数据库维护操作完成!")
}

// 备份数据库
func backupDB(dbPath string) {
	backupPath := fmt.Sprintf("%s.backup.%s", dbPath, time.Now().Format("20060102150405"))
	log.Printf("备份数据库到 %s", backupPath)

	// 读取原始数据库文件
	data, err := os.ReadFile(dbPath)
	if err != nil {
		log.Printf("读取数据库文件失败: %v", err)
		return
	}

	// 写入备份文件
	err = os.WriteFile(backupPath, data, 0644)
	if err != nil {
		log.Printf("写入备份文件失败: %v", err)
		return
	}

	log.Printf("数据库备份成功")
}

// 检查数据库
func checkDatabase(db *sql.DB) {
	log.Println("\n=== 数据库检查 ===")

	// 检查表结构
	checkTableStructure(db, "push_task")
	checkTableStructure(db, "feishu_webhook")
	checkTableStructure(db, "push_task_webhook")
	checkTableStructure(db, "send_record")

	// 检查数据库版本
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		log.Printf("获取数据库版本失败: %v", err)
	} else {
		log.Printf("数据库当前版本: %d", version)
	}

	// 检查NULL值
	checkNullColumns(db)

	// 检查任务数量
	var taskCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task").Scan(&taskCount)
	if err != nil {
		log.Printf("获取任务数量失败: %v", err)
	} else {
		log.Printf("任务数量: %d", taskCount)
	}

	// 检查Webhook数量
	var webhookCount int
	err = db.QueryRow("SELECT COUNT(*) FROM feishu_webhook").Scan(&webhookCount)
	if err != nil {
		log.Printf("获取Webhook数量失败: %v", err)
	} else {
		log.Printf("Webhook数量: %d", webhookCount)
	}

	// 检查关联记录
	var taskWebhookCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task_webhook").Scan(&taskWebhookCount)
	if err != nil {
		log.Printf("获取任务Webhook关联数量失败: %v", err)
	} else {
		log.Printf("任务Webhook关联数量: %d", taskWebhookCount)
	}

	// 检查孤立记录
	var orphanedTaskWebhookCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM push_task_webhook ptw
		LEFT JOIN push_task pt ON ptw.task_id = pt.id
		LEFT JOIN feishu_webhook fw ON ptw.webhook_id = fw.id
		WHERE pt.id IS NULL OR fw.id IS NULL
	`).Scan(&orphanedTaskWebhookCount)
	if err != nil {
		log.Printf("检查孤立关联记录失败: %v", err)
	} else if orphanedTaskWebhookCount > 0 {
		log.Printf("警告: 发现 %d 条孤立的任务Webhook关联记录", orphanedTaskWebhookCount)
	}

	// 检查无效的任务ID
	var invalidIDCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task WHERE id <= 0").Scan(&invalidIDCount)
	if err != nil {
		log.Printf("检查无效ID失败: %v", err)
	} else if invalidIDCount > 0 {
		log.Printf("警告: 发现 %d 条ID无效的任务记录", invalidIDCount)
	}
}

// 检查表结构
func checkTableStructure(db *sql.DB, tableName string) {
	log.Printf("检查表 %s 结构:", tableName)

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		log.Printf("  获取表结构失败: %v", err)
		return
	}
	defer rows.Close()

	var columnCount int
	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value sql.NullString
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			log.Printf("  读取列信息失败: %v", err)
			continue
		}

		defaultValue := "NULL"
		if dflt_value.Valid {
			defaultValue = dflt_value.String
		}

		log.Printf("  列 %d: %s (%s), NOT NULL=%d, DEFAULT=%s, PK=%d",
			cid, name, type_, notnull, defaultValue, pk)
		columnCount++
	}

	log.Printf("  表 %s 共有 %d 列", tableName, columnCount)
}

// 修复NULL值
func fixNullValues(db *sql.DB) {
	log.Println("\n=== 修复NULL值 ===")

	// 先检查NULL值数量
	checkNullColumns(db)

	// 定义需要修复的文本类型列
	textColumns := []string{
		"name", "query", "time_range", "card_title",
		"card_template", "metric_label", "unit",
		"initial_send_time", "custom_metric_label", "custom_chart_label",
	}

	// 修复文本列的NULL值
	for _, column := range textColumns {
		fixNullTextColumn(db, "push_task", column)
	}

	// 修复整数列的NULL值（使用默认值0）
	intColumns := []string{
		"source_id", "step", "schedule_interval",
		"chart_template_id", "enabled",
	}

	for _, column := range intColumns {
		fixNullIntColumn(db, "push_task", column)
	}

	// 修复日期时间列的NULL值
	fixDateTimeColumn(db, "push_task", "last_run_at")

	// 最后再次检查NULL值情况
	log.Println("\n修复后的NULL值情况:")
	checkNullColumns(db)
}

// 检查NULL值
func checkNullColumns(db *sql.DB) {
	log.Println("\n检查表中的NULL值情况:")

	// 获取表的所有列
	columnRows, err := db.Query("PRAGMA table_info(push_task)")
	if err != nil {
		log.Printf("获取列信息失败: %v", err)
		return
	}
	defer columnRows.Close()

	var columns []string
	for columnRows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value sql.NullString
		if err := columnRows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			log.Printf("读取列信息失败: %v", err)
			continue
		}
		columns = append(columns, name)
	}

	// 检查每一列的NULL值数量
	var totalNullCount int
	for _, column := range columns {
		var nullCount int
		query := fmt.Sprintf("SELECT COUNT(*) FROM push_task WHERE %s IS NULL", column)
		err := db.QueryRow(query).Scan(&nullCount)
		if err != nil {
			log.Printf("检查 %s 列的NULL值失败: %v", column, err)
			continue
		}

		if nullCount > 0 {
			log.Printf("  列 %s 有 %d 个NULL值", column, nullCount)
			totalNullCount += nullCount
		}
	}

	if totalNullCount == 0 {
		log.Printf("  数据库中没有NULL值")
	} else {
		log.Printf("  数据库中共有 %d 个NULL值需要修复", totalNullCount)
	}
}

// 修复文本类型列的NULL值
func fixNullTextColumn(db *sql.DB, tableName, column string) {
	// 首先检查该列是否存在NULL值
	var nullCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", tableName, column)
	err := db.QueryRow(countQuery).Scan(&nullCount)
	if err != nil {
		log.Printf("检查 %s 列的NULL值失败: %v", column, err)
		return
	}

	if nullCount == 0 {
		return // 没有NULL值，无需修复
	}

	// 根据列名确定默认值
	var defaultValue string
	switch column {
	case "time_range":
		defaultValue = "30m"
	case "card_template":
		defaultValue = "blue"
	case "metric_label":
		defaultValue = "pod"
	case "name":
		defaultValue = "未命名任务_" + time.Now().Format("20060102150405")
	default:
		defaultValue = ""
	}

	// 更新NULL值为默认值
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s IS NULL", tableName, column, column)
	result, err := db.Exec(updateQuery, defaultValue)
	if err != nil {
		log.Printf("修复 %s 列的NULL值失败: %v", column, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %s 列的 %d 个NULL值为 '%s'", column, rowsAffected, defaultValue)
}

// 修复整数类型列的NULL值
func fixNullIntColumn(db *sql.DB, tableName, column string) {
	// 首先检查该列是否存在NULL值
	var nullCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", tableName, column)
	err := db.QueryRow(countQuery).Scan(&nullCount)
	if err != nil {
		log.Printf("检查 %s 列的NULL值失败: %v", column, err)
		return
	}

	if nullCount == 0 {
		return // 没有NULL值，无需修复
	}

	// 根据列名确定默认值
	var defaultValue int
	switch column {
	case "step":
		defaultValue = 300
	case "schedule_interval":
		defaultValue = 3600
	case "enabled":
		defaultValue = 1
	default:
		defaultValue = 0
	}

	// 更新NULL值为默认值
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s IS NULL", tableName, column, column)
	result, err := db.Exec(updateQuery, defaultValue)
	if err != nil {
		log.Printf("修复 %s 列的NULL值失败: %v", column, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %s 列的 %d 个NULL值为 %d", column, rowsAffected, defaultValue)
}

// 修复日期时间类型列的NULL值
func fixDateTimeColumn(db *sql.DB, tableName, column string) {
	// 首先检查该列是否存在NULL值
	var nullCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", tableName, column)
	err := db.QueryRow(countQuery).Scan(&nullCount)
	if err != nil {
		log.Printf("检查 %s 列的NULL值失败: %v", column, err)
		return
	}

	if nullCount == 0 {
		return // 没有NULL值，无需修复
	}

	// 使用当前时间作为默认值
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = datetime('now') WHERE %s IS NULL", tableName, column, column)
	result, err := db.Exec(updateQuery)
	if err != nil {
		log.Printf("修复 %s 列的NULL值失败: %v", column, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %s 列的 %d 个NULL值为当前时间", column, rowsAffected)
}

// 清理数据库
func cleanupDatabase(db *sql.DB) {
	log.Println("\n=== 清理数据库 ===")

	// 清理孤立的webhook关联
	result, err := db.Exec(`
		DELETE FROM push_task_webhook 
		WHERE task_id NOT IN (SELECT id FROM push_task)
		OR webhook_id NOT IN (SELECT id FROM feishu_webhook)
	`)

	if err != nil {
		log.Printf("清理孤立的webhook关联失败: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("已清理 %d 条孤立的webhook关联记录", rowsAffected)
		} else {
			log.Printf("没有孤立的webhook关联需要清理")
		}
	}

	// 清理无效ID的任务
	result, err = db.Exec("DELETE FROM push_task WHERE id <= 0")
	if err != nil {
		log.Printf("清理无效ID的任务失败: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("已清理 %d 条无效ID的任务", rowsAffected)
		} else {
			log.Printf("没有无效ID的任务需要清理")
		}
	}

	// 清理旧的发送记录（可选，保留最近100条）
	result, err = db.Exec(`
		DELETE FROM send_record 
		WHERE id NOT IN (
			SELECT id FROM send_record 
			ORDER BY sent_at DESC 
			LIMIT 100
		)
	`)

	if err != nil {
		log.Printf("清理旧的发送记录失败: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("已清理 %d 条旧的发送记录", rowsAffected)
		} else {
			log.Printf("没有旧的发送记录需要清理")
		}
	}
}

// 修复表结构
func repairTables(db *sql.DB) {
	log.Println("\n=== 修复表结构 ===")

	// 修复push_task表
	log.Printf("添加缺失的列...")

	// 检查并添加custom_metric_label列
	var hasColumn bool
	err := db.QueryRow("SELECT COUNT(*) > 0 FROM pragma_table_info('push_task') WHERE name='custom_metric_label'").Scan(&hasColumn)
	if err != nil {
		log.Printf("检查custom_metric_label列失败: %v", err)
	} else if !hasColumn {
		_, err = db.Exec("ALTER TABLE push_task ADD COLUMN custom_metric_label TEXT DEFAULT ''")
		if err != nil {
			log.Printf("添加custom_metric_label列失败: %v", err)
		} else {
			log.Printf("已添加custom_metric_label列")
		}
	}

	// 检查并添加custom_chart_label列
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM pragma_table_info('push_task') WHERE name='custom_chart_label'").Scan(&hasColumn)
	if err != nil {
		log.Printf("检查custom_chart_label列失败: %v", err)
	} else if !hasColumn {
		_, err = db.Exec("ALTER TABLE push_task ADD COLUMN custom_chart_label TEXT DEFAULT ''")
		if err != nil {
			log.Printf("添加custom_chart_label列失败: %v", err)
		} else {
			log.Printf("已添加custom_chart_label列")
		}
	}

	// 执行VACUUM操作优化数据库
	log.Printf("执行VACUUM操作优化数据库...")
	_, err = db.Exec("VACUUM")
	if err != nil {
		log.Printf("执行VACUUM失败: %v", err)
	} else {
		log.Printf("VACUUM操作完成")
	}
}

// 验证数据有效性
func validateData(db *sql.DB) {
	log.Println("\n=== 验证数据有效性 ===")

	// 检查任务ID是否合法
	var invalidIDCount int
	err := db.QueryRow("SELECT COUNT(*) FROM push_task WHERE id <= 0").Scan(&invalidIDCount)
	if err != nil {
		log.Printf("检查无效ID失败: %v", err)
	} else if invalidIDCount > 0 {
		log.Printf("警告: 仍有 %d 条ID无效的任务记录", invalidIDCount)
	} else {
		log.Printf("所有任务ID均有效")
	}

	// 检查必要字段是否为空
	requiredColumns := []string{"name", "source_id", "time_range", "step"}
	for _, column := range requiredColumns {
		var emptyCount int
		query := fmt.Sprintf("SELECT COUNT(*) FROM push_task WHERE %s IS NULL OR %s = ''", column, column)
		err := db.QueryRow(query).Scan(&emptyCount)
		if err != nil {
			log.Printf("检查 %s 列是否为空失败: %v", column, err)
			continue
		}

		if emptyCount > 0 {
			log.Printf("警告: 有 %d 条记录的 %s 列为空", emptyCount, column)
		} else {
			log.Printf("所有记录的 %s 列均有有效值", column)
		}
	}

	// 检查时间范围格式是否有效
	var invalidTimeRangeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task WHERE time_range NOT LIKE '%h' AND time_range NOT LIKE '%m' AND time_range NOT LIKE '%s'").Scan(&invalidTimeRangeCount)
	if err != nil {
		log.Printf("检查无效时间范围失败: %v", err)
	} else if invalidTimeRangeCount > 0 {
		log.Printf("警告: 有 %d 条记录的时间范围格式无效", invalidTimeRangeCount)
	} else {
		log.Printf("所有时间范围格式均有效")
	}

	// 检查是否存在孤立记录
	var orphanedCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM push_task_webhook ptw
		LEFT JOIN push_task pt ON ptw.task_id = pt.id
		LEFT JOIN feishu_webhook fw ON ptw.webhook_id = fw.id
		WHERE pt.id IS NULL OR fw.id IS NULL
	`).Scan(&orphanedCount)
	if err != nil {
		log.Printf("检查孤立关联记录失败: %v", err)
	} else if orphanedCount > 0 {
		log.Printf("警告: 仍有 %d 条孤立的任务Webhook关联记录", orphanedCount)
	} else {
		log.Printf("没有孤立的任务Webhook关联记录")
	}

	log.Printf("数据验证完成")
}
