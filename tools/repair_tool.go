package main

import (
	"database/sql"
	"fmt"
	"fsvchart-notify/internal/database"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("=== 表结构自动修复工具 ===")

	testDBPath := "./data/old_test.db"

	// 连接旧数据库
	db, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查当前结构
	log.Println("【初始状态】:")
	logTableStructure(db, "push_task")
	logDBVersion(db)

	// 调用验证和修复函数
	log.Println("\n【执行修复】:")
	isValid, cols, missing, err := database.VerifyTableStructure(db, "push_task")
	if err != nil {
		log.Fatalf("验证表结构失败: %v", err)
	}

	fmt.Printf("结构是否有效: %v\n", isValid)
	fmt.Printf("现有列: %v\n", getColumnNames(cols))
	fmt.Printf("缺失列: %v\n", missing)

	if !isValid {
		err = database.RepairTableStructure(db, "push_task", true)
		if err != nil {
			log.Fatalf("修复表结构失败: %v", err)
		}
		log.Println("修复完成")
	}

	// 检查修复后的结构
	log.Println("\n【修复后状态】:")
	logTableStructure(db, "push_task")

	log.Println("测试完成!")
	os.Exit(0)
}

func logTableStructure(db *sql.DB, tableName string) {
	fmt.Printf("表 %s 结构:\n", tableName)

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		fmt.Printf("  读取表结构失败: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value sql.NullString
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			fmt.Printf("  读取列信息失败: %v\n", err)
			continue
		}

		defaultValue := "NULL"
		if dflt_value.Valid {
			defaultValue = dflt_value.String
		}

		fmt.Printf("  列 %d: %s (%s), NOT NULL=%d, DEFAULT=%s, PK=%d\n",
			cid, name, type_, notnull, defaultValue, pk)
	}
}

func logDBVersion(db *sql.DB) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		fmt.Printf("读取数据库版本失败: %v\n", err)
		return
	}

	fmt.Printf("数据库版本: %d (当前最新版本: %d)\n", version, database.CurrentSchemaVersion)

	rows, err := db.Query("SELECT version, applied_at, description FROM schema_version ORDER BY version")
	if err != nil {
		fmt.Printf("读取版本记录失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("版本记录:")
	for rows.Next() {
		var ver int
		var appliedAt, desc string
		if err := rows.Scan(&ver, &appliedAt, &desc); err != nil {
			fmt.Printf("  读取版本记录失败: %v\n", err)
			continue
		}

		fmt.Printf("  版本 %d: 应用于 %s - %s\n", ver, appliedAt, desc)
	}
}

func getColumnNames(columns map[string]bool) []string {
	var names []string
	for name := range columns {
		names = append(names, name)
	}
	return names
}
