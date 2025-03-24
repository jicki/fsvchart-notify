// This file has been deprecated and its functionality has been merged into database.go
// Please use database.SetupDB() instead of database.InitDB()

package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		dsn := dbPath + "?_busy_timeout=5000"
		db, err = sql.Open("sqlite3", dsn)
		if err != nil {
			log.Printf("[InitDB] Open db error: %v", err)
			return
		}
		// 允许并发：5个连接
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)

		// 启用 WAL 模式
		if _, e := db.Exec("PRAGMA journal_mode=WAL;"); e != nil {
			log.Printf("[InitDB] set WAL mode error: %v", e)
		}
		// busy_timeout=5000
		if _, e := db.Exec("PRAGMA busy_timeout=5000;"); e != nil {
			log.Printf("[InitDB] set busy_timeout error: %v", e)
		}

		// 初始化表结构
		if initErr := initSchema(db); initErr != nil {
			log.Printf("[InitDB] initSchema error: %v", initErr)
		}
	})
	return db, err
}

func initSchema(d *sql.DB) error {
	// 首先启用外键约束
	if _, err := d.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}

	// 基础表结构创建
	baseStatements := []string{
		`CREATE TABLE IF NOT EXISTS metrics_source (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS feishu_webhook (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS push_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER,
			query TEXT,
			time_range TEXT,
			step INTEGER,
			name TEXT,
			schedule_interval INTEGER DEFAULT 0,
			last_run_at DATETIME,
			FOREIGN KEY(source_id) REFERENCES metrics_source(id)
		)`,
		`CREATE TABLE IF NOT EXISTS push_task_webhook (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER,
			webhook_id INTEGER,
			FOREIGN KEY(task_id) REFERENCES push_task(id),
			FOREIGN KEY(webhook_id) REFERENCES feishu_webhook(id)
		)`,
		`CREATE TABLE IF NOT EXISTS push_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER,
			webhook_id INTEGER,
			last_run_time DATETIME,
			status TEXT,
			log TEXT,
			FOREIGN KEY(source_id) REFERENCES metrics_source(id),
			FOREIGN KEY(webhook_id) REFERENCES feishu_webhook(id)
		)`,
		`CREATE TABLE IF NOT EXISTS chart_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chart_type TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS push_task_send_time (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			weekday INTEGER NOT NULL,  -- 1-7 代表周一到周日
			send_time TEXT NOT NULL,   -- 格式如 "09:00"
			FOREIGN KEY(task_id) REFERENCES push_task(id) ON DELETE CASCADE
		)`,
	}

	// 执行基础表结构创建
	for _, stmt := range baseStatements {
		if _, err := d.Exec(stmt); err != nil {
			return err
		}
	}

	// 检查并添加新列
	columns := []struct {
		table, column, dataType string
	}{
		{"push_task", "last_run_at", "DATETIME"},
		{"push_task", "chart_template_id", "INTEGER REFERENCES chart_template(id)"},
		{"push_task", "card_title", "TEXT"},
		{"push_task", "card_template", "TEXT DEFAULT 'red'"},
		{"push_task", "metric_label", "TEXT DEFAULT 'pod'"},
		{"push_task", "enabled", "INTEGER DEFAULT 1"},
		{"push_task", "unit", "TEXT DEFAULT ''"},
		{"push_task", "initial_send_time", "TEXT DEFAULT ''"},
		{"push_task", "weekday", "INTEGER DEFAULT 1"}, // 1-7 代表周一到周日
	}

	for _, col := range columns {
		var count int
		err := d.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?",
			col.table, col.column,
		).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			if _, err := d.Exec(
				fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
					col.table, col.column, col.dataType),
			); err != nil {
				return err
			}
		}
	}

	return nil
}
