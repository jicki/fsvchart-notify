package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 当前数据库结构版本
const CurrentSchemaVersion = 13 // 版本13: 添加 display_mode 字段支持 PromQL 级别的展示模式

// 表结构定义，用于验证和修复
type TableStructure struct {
	RequiredColumns map[string]string // 列名 -> 列类型
}

// 最新版本的表结构定义
var currentStructures = map[string]TableStructure{
	"push_task": {
		RequiredColumns: map[string]string{
			"id":                  "INTEGER",
			"name":                "TEXT",
			"source_id":           "INTEGER",
			"query":               "TEXT",
			"time_range":          "TEXT",
			"step":                "INTEGER",
			"schedule_interval":   "INTEGER",
			"last_run_at":         "DATETIME",
			"enabled":             "INTEGER",
			"custom_metric_label": "TEXT",
			"button_text":         "TEXT",
			"button_url":          "TEXT",
			"show_data_label":     "INTEGER",
			"push_mode":           "TEXT",
		},
	},
	"push_task_promql": {
		RequiredColumns: map[string]string{
			"id":                  "INTEGER",
			"task_id":             "INTEGER",
			"promql_id":           "INTEGER",
			"chart_template_id":   "INTEGER",
			"unit":                "TEXT",
			"metric_label":        "TEXT",
			"custom_metric_label": "TEXT",
			"initial_unit":        "TEXT",
			"display_order":       "INTEGER",
			"display_mode":        "TEXT",
		},
	},
	// 其他表可以按需添加...
}

// 迁移信息
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// 定义所有迁移脚本，从旧到新
var migrations = []Migration{
	{
		Version:     1,
		Description: "初始结构",
		SQL: `
		CREATE TABLE IF NOT EXISTS metrics_source (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS feishu_webhook (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS push_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			source_id INTEGER NOT NULL,
			query TEXT,
			time_range TEXT NOT NULL DEFAULT '30m',
			step INTEGER NOT NULL DEFAULT 300,
			schedule_interval INTEGER DEFAULT 0,
			last_run_at DATETIME,
			chart_template_id INTEGER,
			card_title TEXT,
			card_template TEXT DEFAULT 'red',
			metric_label TEXT,
			unit TEXT,
			enabled INTEGER DEFAULT 1,
			initial_send_time TEXT
		);

		CREATE TABLE IF NOT EXISTS push_task_query (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			query TEXT NOT NULL,
			chart_template_id INTEGER,
			FOREIGN KEY(task_id) REFERENCES push_task(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS push_task_webhook (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			webhook_id INTEGER NOT NULL,
			UNIQUE(task_id, webhook_id),
			FOREIGN KEY(task_id) REFERENCES push_task(id) ON DELETE CASCADE,
			FOREIGN KEY(webhook_id) REFERENCES feishu_webhook(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS push_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL,
			webhook_id INTEGER NOT NULL,
			last_run_time DATETIME NOT NULL,
			status TEXT NOT NULL,
			log TEXT,
			FOREIGN KEY(source_id) REFERENCES metrics_source(id) ON DELETE CASCADE,
			FOREIGN KEY(webhook_id) REFERENCES feishu_webhook(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS chart_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chart_type TEXT NOT NULL DEFAULT 'area'
		);

		CREATE TABLE IF NOT EXISTS promql (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			query TEXT NOT NULL,
			category TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS push_task_promql (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			promql_id INTEGER NOT NULL,
			chart_template_id INTEGER,
			UNIQUE(task_id, promql_id),
			FOREIGN KEY(task_id) REFERENCES push_task(id) ON DELETE CASCADE,
			FOREIGN KEY(promql_id) REFERENCES promql(id) ON DELETE CASCADE,
			FOREIGN KEY(chart_template_id) REFERENCES chart_template(id) ON DELETE SET NULL
		);
		
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL,
			description TEXT
		);
		`,
	},
	{
		Version:     2,
		Description: "添加custom_chart_label和custom_metric_label列到push_task表",
		SQL: `
		ALTER TABLE push_task ADD COLUMN custom_chart_label TEXT DEFAULT '';
		ALTER TABLE push_task ADD COLUMN custom_metric_label TEXT DEFAULT '';
		`,
	},
	{
		Version:     3,
		Description: "创建send_record表和索引",
		SQL: `
		CREATE TABLE IF NOT EXISTS send_record (
			id INTEGER PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			status TEXT NOT NULL,
			message TEXT,
			webhook TEXT,
			task_name TEXT
		);
		
		-- 添加查询性能索引
		CREATE INDEX IF NOT EXISTS idx_send_record_timestamp ON send_record(timestamp);
		CREATE INDEX IF NOT EXISTS idx_send_record_status ON send_record(status);
		`,
	},
	{
		Version:     4,
		Description: "创建users表和默认admin用户",
		SQL: `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			display_name TEXT,
			email TEXT,
			role TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		-- 添加默认管理员用户 (密码: 123456)
		INSERT OR IGNORE INTO users (username, password, display_name, role) 
		VALUES ('admin', '$2a$10$8KVj4kFyEQT9U.J9Y4Ix3OdPZR9kZ4XiKw/t/G/XJgWJsjxRwueim', 'Administrator', 'admin');
		`,
	},
	{
		Version:     5,
		Description: "修复admin用户密码哈希",
		SQL: `
		-- 更新admin用户密码为123456
		UPDATE users 
		SET password = '$2a$10$iBJJ1cMetVj5uZoOQa4n/OgI8JvRnpGisyVkZoRPGrMRnK40m6Hi2' 
		WHERE username = 'admin';
		`,
	},
	{
		Version:     6,
		Description: "添加button_text和button_url列到push_task表",
		SQL: `
		ALTER TABLE push_task ADD COLUMN button_text TEXT DEFAULT '';
		ALTER TABLE push_task ADD COLUMN button_url TEXT DEFAULT '';
		`,
	},
	{
		Version:     7,
		Description: "添加show_data_label列到push_task表",
		SQL: `
		ALTER TABLE push_task ADD COLUMN show_data_label INTEGER DEFAULT 0;
		`,
	},
	{
		Version:     8,
		Description: "为push_task_promql表添加unit、metric_label、custom_metric_label列",
		SQL: `
		ALTER TABLE push_task_promql ADD COLUMN unit TEXT DEFAULT '';
		ALTER TABLE push_task_promql ADD COLUMN metric_label TEXT DEFAULT 'pod';
		ALTER TABLE push_task_promql ADD COLUMN custom_metric_label TEXT DEFAULT '';
		`,
	},
	{
		Version:     9,
		Description: "添加push_mode列到push_task表，并迁移配置数据",
		SQL: `
		-- 添加push_mode列
		ALTER TABLE push_task ADD COLUMN push_mode TEXT DEFAULT 'chart';
		
		-- 迁移现有任务的配置到push_task_promql表
		-- 为所有现有的push_task_promql记录设置unit和metric_label
		UPDATE push_task_promql
		SET 
			unit = (SELECT unit FROM push_task WHERE push_task.id = push_task_promql.task_id),
			metric_label = (SELECT COALESCE(metric_label, 'pod') FROM push_task WHERE push_task.id = push_task_promql.task_id),
			custom_metric_label = (SELECT COALESCE(custom_metric_label, '') FROM push_task WHERE push_task.id = push_task_promql.task_id)
		WHERE EXISTS (SELECT 1 FROM push_task WHERE push_task.id = push_task_promql.task_id);
		`,
	},
	{
		Version:     10,
		Description: "清理孤立的关联记录（已删除任务遗留的记录）",
		SQL: `
		-- 清理孤立的 push_task_promql 记录
		-- 这些记录引用的 task_id 在 push_task 表中已不存在
		DELETE FROM push_task_promql 
		WHERE task_id NOT IN (SELECT id FROM push_task);
		
		-- 清理孤立的 push_task_webhook 记录
		DELETE FROM push_task_webhook 
		WHERE task_id NOT IN (SELECT id FROM push_task);
		
		-- 清理孤立的 push_task_send_time 记录
		DELETE FROM push_task_send_time 
		WHERE task_id NOT IN (SELECT id FROM push_task);
		
		-- 清理孤立的 push_task_query 记录（旧格式）
		DELETE FROM push_task_query 
		WHERE task_id NOT IN (SELECT id FROM push_task);
		`,
	},
	{
		Version:     11,
		Description: "添加 initial_unit 字段支持单位自动转换",
		SQL: `
		-- 为 push_task_promql 表添加 initial_unit 字段
		-- 用于存储原始单位，系统会自动转换为目标单位
		-- 支持: bytes (B/KB/MB/GB/TB/PB), time (ns/μs/ms/s/m/h), ms 等
		ALTER TABLE push_task_promql ADD COLUMN initial_unit TEXT DEFAULT '';
		`,
	},
	{
		Version:     12,
		Description: "添加 display_order 字段支持自定义显示顺序",
		SQL: `
		-- 为 push_task_promql 表添加 display_order 字段
		-- 用于控制 PromQL 在卡片中的显示顺序
		-- 默认值为 0，数字越小越靠前
		ALTER TABLE push_task_promql ADD COLUMN display_order INTEGER DEFAULT 0;
		
		-- 为现有记录设置默认顺序（按 id 升序）
		UPDATE push_task_promql SET display_order = id WHERE display_order = 0;
		`,
	},
	{
		Version:     13,
		Description: "添加 display_mode 字段支持 PromQL 级别的展示模式",
		SQL: `
		-- 为 push_task_promql 表添加 display_mode 字段
		-- 用于控制每个 PromQL 的展示模式：chart(图表), text(文本), both(混合)
		-- 默认值为 'chart'，保持向后兼容
		ALTER TABLE push_task_promql ADD COLUMN display_mode TEXT DEFAULT 'chart';
		`,
	},
}

var (
	dbInstance *sql.DB // 全局数据库连接
	dbPath     string  // 数据库文件路径
)

// GetDB 返回数据库连接实例
func GetDB() *sql.DB {
	if dbInstance == nil {
		var err error
		dbInstance, err = SetupDB("./data/app.db")
		if err != nil {
			log.Printf("[数据库] 获取数据库连接失败: %v", err)
			return nil
		}
	}
	return dbInstance
}

// 创建版本记录表
func createVersionTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME NOT NULL,
		description TEXT
	)`)
	return err
}

// 获取当前数据库版本
func getCurrentVersion(db *sql.DB) (int, error) {
	// 检查版本表是否存在
	var count int
	err := db.QueryRow(`SELECT count(name) FROM sqlite_master WHERE type='table' AND name='schema_version'`).Scan(&count)
	if err != nil {
		return 0, err
	}

	// 如果表不存在，返回版本0
	if count == 0 {
		return 0, nil
	}

	// 获取最高版本
	var version int
	err = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// 记录版本迁移
func recordMigration(db *sql.DB, version int, description string) error {
	_, err := db.Exec(
		`INSERT INTO schema_version (version, applied_at, description) VALUES (?, ?, ?)`,
		version, time.Now(), description,
	)
	return err
}

// VerifyTableStructure 验证表的结构是否符合预期
// 返回：是否有效，现有列的映射，缺失列的列表，错误
func VerifyTableStructure(db *sql.DB, tableName string) (bool, map[string]bool, []string, error) {
	// 获取表的所有列信息
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, nil, nil, fmt.Errorf("验证表结构失败: %w", err)
	}
	defer rows.Close()

	// 获取当前表结构，如果未定义则返回错误
	expectedStructure, exists := currentStructures[tableName]
	if !exists {
		return false, nil, nil, fmt.Errorf("表 %s 的预期结构未定义", tableName)
	}

	// 检查现有列
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, type_ string
		var notnull, pk int
		var dflt_value sql.NullString
		if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
			return false, nil, nil, fmt.Errorf("读取列信息失败: %w", err)
		}
		existingColumns[name] = true
	}

	// 检查是否有缺失的列
	var missingColumns []string
	for colName := range expectedStructure.RequiredColumns {
		if !existingColumns[colName] {
			missingColumns = append(missingColumns, colName)
		}
	}

	// 如果有缺失列，则结构不有效
	return len(missingColumns) == 0, existingColumns, missingColumns, nil
}

// RepairTableStructure 修复表结构
// forceRecreate=true 时会重新创建表，否则只添加缺失的列
func RepairTableStructure(db *sql.DB, tableName string, forceRecreate bool) error {
	isValid, existingColumns, missingColumns, err := VerifyTableStructure(db, tableName)
	if err != nil {
		return err
	}

	// 如果结构已经有效，不需要修复
	if isValid {
		log.Printf("表 %s 结构已经是最新，无需修复", tableName)
		return nil
	}

	expectedStructure, exists := currentStructures[tableName]
	if !exists {
		return fmt.Errorf("表 %s 的预期结构未定义", tableName)
	}

	// 开始事务以确保操作的原子性
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if forceRecreate {
		// 方法1：完全重建表（适用于SQLite）
		log.Printf("正在重建表 %s...", tableName)

		// 1. 创建临时表
		tempTableName := tableName + "_temp"
		createTableSQL := fmt.Sprintf("CREATE TABLE %s (", tempTableName)
		var columnDefs []string
		for colName, colType := range expectedStructure.RequiredColumns {
			// 处理主键
			if colName == "id" && colType == "INTEGER" {
				columnDefs = append(columnDefs, "id INTEGER PRIMARY KEY AUTOINCREMENT")
			} else {
				columnDefs = append(columnDefs, fmt.Sprintf("%s %s", colName, colType))
			}
		}
		createTableSQL += strings.Join(columnDefs, ", ")
		createTableSQL += ")"

		_, err = tx.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("创建临时表失败: %w", err)
		}

		// 2. 复制数据（只复制存在于两个表中的列）
		var commonColumns []string
		for colName := range expectedStructure.RequiredColumns {
			if existingColumns[colName] {
				commonColumns = append(commonColumns, colName)
			}
		}

		if len(commonColumns) > 0 {
			copyDataSQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s",
				tempTableName,
				strings.Join(commonColumns, ", "),
				strings.Join(commonColumns, ", "),
				tableName)

			_, err = tx.Exec(copyDataSQL)
			if err != nil {
				return fmt.Errorf("复制数据失败: %w", err)
			}
		}

		// 3. 删除原表
		_, err = tx.Exec(fmt.Sprintf("DROP TABLE %s", tableName))
		if err != nil {
			return fmt.Errorf("删除原表失败: %w", err)
		}

		// 4. 重命名临时表
		_, err = tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tempTableName, tableName))
		if err != nil {
			return fmt.Errorf("重命名临时表失败: %w", err)
		}

		log.Printf("表 %s 已重建完成，现在具有最新结构", tableName)
	} else {
		// 方法2：只添加缺失的列（如果SQLite版本支持ALTER TABLE ADD COLUMN）
		log.Printf("正在添加表 %s 的缺失列...", tableName)
		for _, colName := range missingColumns {
			colType := expectedStructure.RequiredColumns[colName]

			// 为TEXT类型的列添加默认空字符串，避免NULL值问题
			var alterSQL string
			if strings.Contains(colType, "TEXT") {
				alterSQL = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s DEFAULT ''", tableName, colName, colType)
			} else {
				alterSQL = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, colName, colType)
			}

			_, err = tx.Exec(alterSQL)
			if err != nil {
				return fmt.Errorf("添加列 %s 失败: %w", colName, err)
			}
			log.Printf("已添加列: %s %s", colName, colType)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// migrate 执行迁移
func migrate(db *sql.DB) error {
	// 创建版本表（如果不存在）
	if err := createVersionTable(db); err != nil {
		return fmt.Errorf("创建版本表失败: %w", err)
	}

	// 获取当前版本
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}

	log.Printf("[数据库] 当前数据库版本: %d, 最新版本: %d", currentVersion, CurrentSchemaVersion)

	// 如果已经是最新版本，不需要迁移
	if currentVersion >= CurrentSchemaVersion {
		log.Printf("[数据库] 数据库已是最新版本，无需迁移")
		return nil
	}

	// 特殊情况处理: 没有版本记录但实际结构可能已经是新版本
	// 检查是否存在 custom_metric_label 列来判断是否已经是新版本
	if currentVersion == 0 {
		var count int
		err := db.QueryRow(`
			SELECT count(*) FROM pragma_table_info('push_task') 
			WHERE name='custom_metric_label'
		`).Scan(&count)

		// 如果查询成功且列已存在，直接设置为新版本
		if err == nil && count > 0 {
			log.Printf("[数据库] 检测到数据库结构已经是新版本，正在记录版本信息")

			// 记录版本1和2已完成
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("开启事务失败: %w", err)
			}

			// 记录版本1
			_, err = tx.Exec(
				`INSERT INTO schema_version (version, applied_at, description) VALUES (?, ?, ?)`,
				1, time.Now(), "初始结构（自动检测）",
			)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("记录版本1失败: %w", err)
			}

			// 记录版本2
			_, err = tx.Exec(
				`INSERT INTO schema_version (version, applied_at, description) VALUES (?, ?, ?)`,
				2, time.Now(), "添加custom_chart_label和custom_metric_label列（自动检测）",
			)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("记录版本2失败: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}

			// 更新当前版本
			currentVersion = 2
			log.Printf("[数据库] 成功记录版本信息，当前版本: %d", currentVersion)

			// 继续执行更高版本的迁移，不返回
		}
	}

	// 开始迁移
	log.Printf("[数据库] 开始迁移数据库，从版本 %d 升级到 %d", currentVersion, CurrentSchemaVersion)

	// 分别执行每个迁移版本，每个版本单独使用事务
	for _, migration := range migrations {
		// 只执行比当前版本更高的迁移
		if migration.Version <= currentVersion {
			continue
		}

		log.Printf("[数据库] 执行迁移: 版本 %d - %s", migration.Version, migration.Description)

		// 为每个版本迁移单独创建事务
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("为版本 %d 创建事务失败: %w", migration.Version, err)
		}

		// 执行SQL脚本
		_, err = tx.Exec(migration.SQL)
		if err != nil {
			tx.Rollback()

			// 版本2特殊处理：如果是列已存在错误，可以认为结构已经是最新的，只需记录版本即可
			if migration.Version == 2 && (err.Error() == "duplicate column name: custom_chart_label" ||
				err.Error() == "duplicate column name: custom_metric_label") {
				log.Printf("[数据库] 检测到列已存在，迁移脚本2不需要执行，只记录版本信息")

				// 记录版本2已完成
				err = recordMigration(db, migration.Version, migration.Description+" (列已存在，跳过执行)")
				if err != nil {
					return fmt.Errorf("记录版本信息失败: %w", err)
				}

				// 更新当前版本，继续下一个迁移
				currentVersion = migration.Version
				continue
			}

			// 其他错误返回
			return fmt.Errorf("执行迁移脚本失败 (版本 %d): %w", migration.Version, err)
		}

		// 记录已执行的迁移
		_, err = tx.Exec(
			`INSERT INTO schema_version (version, applied_at, description) VALUES (?, ?, ?)`,
			migration.Version, time.Now(), migration.Description,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("记录迁移版本失败 (版本 %d): %w", migration.Version, err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败 (版本 %d): %w", migration.Version, err)
		}

		log.Printf("[数据库] 迁移版本 %d 执行成功", migration.Version)

		// 版本10特殊处理：记录清理统计
		if migration.Version == 10 {
			// 查询每个表剩余的记录数，以确认清理效果
			var promqlCount, webhookCount, sendtimeCount, queryCount int
			
			db.QueryRow("SELECT COUNT(*) FROM push_task_promql").Scan(&promqlCount)
			db.QueryRow("SELECT COUNT(*) FROM push_task_webhook").Scan(&webhookCount)
			db.QueryRow("SELECT COUNT(*) FROM push_task_send_time").Scan(&sendtimeCount)
			db.QueryRow("SELECT COUNT(*) FROM push_task_query").Scan(&queryCount)
			
			log.Printf("[数据库清理] 孤立记录清理完成")
			log.Printf("[数据库清理] 当前记录数 - push_task_promql: %d, push_task_webhook: %d, push_task_send_time: %d, push_task_query: %d",
				promqlCount, webhookCount, sendtimeCount, queryCount)
			log.Printf("[数据库清理] 现在可以正常删除不再使用的 PromQL 了")
		}

		// 更新当前版本
		currentVersion = migration.Version
	}

	// 执行完迁移后，验证表结构并修复（如果需要）
	if currentVersion == CurrentSchemaVersion {
		log.Printf("[数据库] 迁移完成，验证表结构")

		// 验证并修复push_task表
		isValid, _, missingColumns, err := VerifyTableStructure(db, "push_task")
		if err != nil {
			log.Printf("[数据库] 警告: 验证表结构失败: %v", err)
		} else if !isValid {
			log.Printf("[数据库] 警告: 表结构不一致! 缺少列: %v", missingColumns)

			// 自动修复表结构
			log.Printf("[数据库] 尝试自动修复表结构...")
			repairErr := RepairTableStructure(db, "push_task", false) // 不删除多余的列，只添加缺少的列
			if repairErr != nil {
				log.Printf("[数据库] 警告: 表结构修复失败: %v", repairErr)
			} else {
				log.Printf("[数据库] 表结构修复成功")
			}
		}
	}

	log.Printf("[数据库] 迁移完成，当前版本: %d", CurrentSchemaVersion)
	return nil
}

// SetupDB 初始化数据库连接并进行迁移
func SetupDB(dbPath string) (*sql.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败: %w", err)
		}
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 执行数据库迁移
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}

// ColumnExists 检查某个表中是否存在指定的列
func ColumnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("SELECT count(*) FROM pragma_table_info('%s') WHERE name='%s'", tableName, columnName)
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查列是否存在失败: %w", err)
	}
	return count > 0, nil
}

// GetCurrentVersion 获取数据库当前版本的导出函数
func GetCurrentVersion(db *sql.DB) (int, error) {
	return getCurrentVersion(db)
}

// MigrateDB 执行数据库迁移
func MigrateDB(db *sql.DB) error {
	return migrate(db)
}
