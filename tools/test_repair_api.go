package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("=== API自动修复测试工具 ===")

	// 准备测试数据库
	testDBPath := "./data/test/old_db.db"

	// 连接测试数据库
	db, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 添加测试数据
	_, err = db.Exec(`
		INSERT INTO push_task (id, name, source_id, query, time_range, step, schedule_interval, enabled)
		VALUES (1, '测试任务', 1, 'up', '30m', 60, 3600, 1)
	`)
	if err != nil {
		log.Fatalf("添加测试数据失败: %v", err)
	}

	// 显示初始状态
	log.Println("【初始数据库状态】:")
	logTableStructure(db, "push_task")
	logDBVersion(db)

	// 启动一个临时的HTTP服务器，模拟API请求
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Println("收到测试请求，开始处理...")

		// 模拟API处理逻辑
		hasColumn, err := columnExists(db, "push_task", "custom_metric_label")
		if err != nil {
			log.Printf("检查列是否存在失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "检查列是否存在失败: %v", err)
			return
		}

		log.Printf("custom_metric_label列存在: %v", hasColumn)

		if !hasColumn {
			// 检查数据库版本
			var version int
			err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
			if err != nil {
				log.Printf("获取数据库版本失败: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "获取数据库版本失败: %v", err)
				return
			}

			log.Printf("数据库版本: %d", version)

			// 如果版本是最新的但结构不匹配，尝试修复
			if version == 3 {
				log.Printf("检测到数据库版本是最新的(3)但结构不匹配，尝试修复...")

				// 添加缺失的列
				_, err := db.Exec("ALTER TABLE push_task ADD COLUMN custom_metric_label TEXT DEFAULT ''")
				if err != nil {
					log.Printf("修复表结构失败: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "修复表结构失败: %v", err)
					return
				}

				log.Printf("表结构修复完成，已添加custom_metric_label列")
				hasColumn = true
			}
		}

		// 根据列是否存在选择查询
		var query string
		if hasColumn {
			query = `
				SELECT id, name, source_id, query, time_range, step, schedule_interval, 
				datetime(last_run_at) as last_run_at, enabled, custom_metric_label 
				FROM push_task
			`
		} else {
			query = `
				SELECT id, name, source_id, query, time_range, step, schedule_interval, 
				datetime(last_run_at) as last_run_at, enabled, '' as custom_metric_label 
				FROM push_task
			`
		}

		// 执行查询
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("查询失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "查询失败: %v", err)
			return
		}
		defer rows.Close()

		// 处理结果
		var tasks []map[string]interface{}
		for rows.Next() {
			var id, sourceID, step, scheduleInterval int
			var name, query, timeRange, lastRunAt, customMetricLabel string
			var enabled int

			err := rows.Scan(&id, &name, &sourceID, &query, &timeRange, &step, &scheduleInterval, &lastRunAt, &enabled, &customMetricLabel)
			if err != nil {
				log.Printf("扫描行数据失败: %v", err)
				continue
			}

			task := map[string]interface{}{
				"id":                  id,
				"name":                name,
				"source_id":           sourceID,
				"query":               query,
				"time_range":          timeRange,
				"step":                step,
				"schedule_interval":   scheduleInterval,
				"last_run_at":         lastRunAt,
				"enabled":             enabled == 1,
				"custom_metric_label": customMetricLabel,
			}

			tasks = append(tasks, task)
		}

		// 返回结果
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\n  \"success\": true,\n  \"data\": %v,\n  \"message\": \"查询成功，共%d条记录\"\n}", formatTasks(tasks), len(tasks))
	})

	// 启动服务器
	go func() {
		log.Println("启动测试服务器在 http://localhost:8888/test")
		if err := http.ListenAndServe(":8888", nil); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 发送测试请求
	log.Println("\n【发送测试请求】:")
	resp, err := http.Get("http://localhost:8888/test")
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	log.Printf("响应状态码: %d", resp.StatusCode)
	log.Printf("响应内容: %s", buf[:n])

	// 检查修复后的状态
	log.Println("\n【修复后数据库状态】:")
	logTableStructure(db, "push_task")

	log.Println("测试完成!")
	os.Exit(0)
}

func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("SELECT count(*) FROM pragma_table_info('%s') WHERE name='%s'", tableName, columnName)
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查列是否存在失败: %w", err)
	}
	return count > 0, nil
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

	fmt.Printf("数据库版本: %d\n", version)

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

func formatTasks(tasks []map[string]interface{}) string {
	if len(tasks) == 0 {
		return "[]"
	}

	var result string
	result = "[\n"

	for i, task := range tasks {
		result += "    {\n"

		for k, v := range task {
			var valueStr string
			switch val := v.(type) {
			case string:
				valueStr = fmt.Sprintf("\"%s\"", val)
			case bool:
				valueStr = fmt.Sprintf("%t", val)
			default:
				valueStr = fmt.Sprintf("%v", val)
			}

			result += fmt.Sprintf("      \"%s\": %s", k, valueStr)
			result += ",\n"
		}

		// 移除最后一个逗号
		result = result[:len(result)-2] + "\n"

		result += "    }"
		if i < len(tasks)-1 {
			result += ","
		}
		result += "\n"
	}

	result += "  ]"
	return result
}
