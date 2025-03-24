package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run tools/fix_null_values.go <数据库路径>")
		fmt.Println("示例: go run tools/fix_null_values.go ./data/app.db")
		os.Exit(1)
	}

	dbPath := os.Args[1]
	log.Printf("=== 数据库NULL值修复工具 ===")
	log.Printf("数据库路径: %s", dbPath)

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 显示表结构
	printTableStructure(db, "push_task")

	// 检查并修复NULL值
	fixNullValues(db)

	log.Println("NULL值修复完成!")
}

func printTableStructure(db *sql.DB, tableName string) {
	log.Printf("表 %s 结构:", tableName)

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		log.Fatalf("获取表结构失败: %v", err)
	}
	defer rows.Close()

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

		log.Printf("  列 %d: %s (%s), NOT NULL=%d, DEFAULT=%s, PK=%d",
			cid, name, type_, notnull, defaultValue, pk)
	}
}

func fixNullValues(db *sql.DB) {
	// 先查看NULL值记录
	checkNullColumns(db)

	// 定义需要修复的文本类型列
	textColumns := []string{
		"name", "query", "time_range", "card_title",
		"card_template", "metric_label", "unit",
		"initial_send_time", "custom_metric_label", "custom_chart_label",
	}

	// 修复文本列的NULL值
	for _, column := range textColumns {
		fixNullTextColumn(db, column)
	}

	// 修复整数列的NULL值（使用默认值0）
	intColumns := []string{
		"source_id", "step", "schedule_interval",
		"chart_template_id", "enabled",
	}

	for _, column := range intColumns {
		fixNullIntColumn(db, column)
	}

	// 最后再次检查NULL值情况
	log.Println("\n修复后的NULL值情况:")
	checkNullColumns(db)
}

func checkNullColumns(db *sql.DB) {
	log.Println("\n检查push_task表中的NULL值情况:")

	// 获取表的所有列
	columnRows, err := db.Query("PRAGMA table_info(push_task)")
	if err != nil {
		log.Fatalf("获取列信息失败: %v", err)
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
			log.Fatalf("读取列信息失败: %v", err)
			continue
		}
		columns = append(columns, name)
	}

	// 检查每一列的NULL值数量
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
		}
	}
}

func fixNullTextColumn(db *sql.DB, column string) {
	// 首先检查该列是否存在NULL值
	var nullCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM push_task WHERE %s IS NULL", column)
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
	default:
		defaultValue = ""
	}

	// 更新NULL值为默认值
	updateQuery := fmt.Sprintf("UPDATE push_task SET %s = ? WHERE %s IS NULL", column, column)
	result, err := db.Exec(updateQuery, defaultValue)
	if err != nil {
		log.Printf("修复 %s 列的NULL值失败: %v", column, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %s 列的 %d 个NULL值为 '%s'", column, rowsAffected, defaultValue)
}

func fixNullIntColumn(db *sql.DB, column string) {
	// 首先检查该列是否存在NULL值
	var nullCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM push_task WHERE %s IS NULL", column)
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
	updateQuery := fmt.Sprintf("UPDATE push_task SET %s = ? WHERE %s IS NULL", column, column)
	result, err := db.Exec(updateQuery, defaultValue)
	if err != nil {
		log.Printf("修复 %s 列的NULL值失败: %v", column, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("已修复 %s 列的 %d 个NULL值为 %d", column, rowsAffected, defaultValue)
}
