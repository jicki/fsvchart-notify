package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/handler"
	"fsvchart-notify/internal/middleware"
	"fsvchart-notify/internal/models"
	"fsvchart-notify/internal/scheduler"
	"fsvchart-notify/internal/service"
)

// -------------- 1. metrics_source --------------

type MetricsSourceReq struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GET /api/metrics_source
func getMetricsSources(c *gin.Context) {
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		// 统一返回 JSON
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := db.Query("SELECT id, name, url FROM metrics_source")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.MetricsSource
	for rows.Next() {
		var ms models.MetricsSource
		if err := rows.Scan(&ms.ID, &ms.Name, &ms.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, ms)
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/metrics_source
func createMetricsSource(c *gin.Context) {
	var req MetricsSourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果请求体不是合法 JSON 或字段不对，返回400 + JSON
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, err := db.Prepare("INSERT INTO metrics_source(name, url) VALUES (?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res, err := stmt.Exec(req.Name, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newID, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"id": newID, "name": req.Name, "url": req.URL})
}

// PUT /api/metrics_source/:id
func updateMetricsSource(c *gin.Context) {
	idStr := c.Param("id")

	var req MetricsSourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, err := db.Prepare("UPDATE metrics_source SET name=?, url=? WHERE id=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(req.Name, req.URL, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated", "id": idStr})
}

// DELETE /api/metrics_source/:id
func deleteMetricsSource(c *gin.Context) {
	idStr := c.Param("id")
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, e := db.Prepare("DELETE FROM metrics_source WHERE id=?")
	if e != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		return
	}
	res, delErr := stmt.Exec(idStr)
	if delErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": delErr.Error()})
		return
	}
	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted", "id": idStr})
}

// -------------- 2. feishu_webhook --------------

type FeishuWebhookReq struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GET /api/feishu_webhook
func getFeishuWebhooks(c *gin.Context) {
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := db.Query("SELECT id, name, url FROM feishu_webhook")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.FeishuWebhook
	for rows.Next() {
		var wb models.FeishuWebhook
		if err := rows.Scan(&wb.ID, &wb.Name, &wb.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, wb)
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/feishu_webhook
func createFeishuWebhook(c *gin.Context) {
	var req FeishuWebhookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, e := db.Prepare("INSERT INTO feishu_webhook(name, url) VALUES (?, ?)")
	if e != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		return
	}
	res, e2 := stmt.Exec(req.Name, req.URL)
	if e2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e2.Error()})
		return
	}
	newID, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"id": newID, "name": req.Name, "url": req.URL})
}

// PUT /api/feishu_webhook/:id
func updateFeishuWebhook(c *gin.Context) {
	idStr := c.Param("id") // 必须从 :id 获取

	var req FeishuWebhookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stmt, e := db.Prepare("UPDATE feishu_webhook SET name=?, url=? WHERE id=?")
	if e != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		return
	}
	res, e2 := stmt.Exec(req.Name, req.URL, idStr)
	if e2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e2.Error()})
		return
	}
	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated", "id": idStr})
}

// DELETE /api/feishu_webhook/:id
func deleteFeishuWebhook(c *gin.Context) {
	idStr := c.Param("id")
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, delErr := db.Exec("DELETE FROM feishu_webhook WHERE id=?", idStr)
	if delErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": delErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted", "id": idStr})
}

// -------------- 3. push_task & push_task_webhook --------------

// 新增：PromQL 配置结构体
type PromQLConfig struct {
	PromQLID          int64  `json:"promql_id"`
	Unit              string `json:"unit"`
	MetricLabel       string `json:"metric_label"`
	CustomMetricLabel string `json:"custom_metric_label"`
	ChartTemplateID   int64  `json:"chart_template_id"` // 每个PromQL可以有自己的图表模板
	InitialUnit       string `json:"initial_unit"`      // 初始单位，用于自动单位转换
	DisplayOrder      int    `json:"display_order"`     // 显示顺序，数字越小越靠前
}

type PushTaskReq struct {
	Name              string                `json:"name"`
	SourceID          int64                 `json:"source_id"`
	Query             string                `json:"query"`           // 保留向后兼容
	Queries           []QueryItem           `json:"queries"`         // 新增：支持多个查询
	PromQLIDs         []int64               `json:"promql_ids"`      // 保留向后兼容
	PromQLConfigs     []PromQLConfig        `json:"promql_configs"`  // 新增：每个 PromQL 的独立配置
	TimeRange         string                `json:"time_range"`
	Step              float64               `json:"step"`
	WebhookIDs        []int64               `json:"webhook_ids"`
	SchedInterval     int                   `json:"schedule_interval"`
	ChartTemplateID   int64                 `json:"chart_template_id"` // 保留向后兼容
	CardTitle         string                `json:"card_title"`
	CardTemplate      string                `json:"card_template"`
	MetricLabel       string                `json:"metric_label"`      // 保留向后兼容
	CustomMetricLabel string                `json:"custom_metric_label"` // 保留向后兼容
	Unit              string                `json:"unit"`              // 保留向后兼容
	ButtonText        string                `json:"button_text"`
	ButtonURL         string                `json:"button_url"`
	SendTimes         []models.TaskSendTime `json:"send_times"`
	ShowDataLabel     bool                  `json:"show_data_label"`
	PushMode          string                `json:"push_mode"` // 新增：推送模式 chart/text
}

// 新增：查询项结构体
type QueryItem struct {
	Query           string `json:"query"`
	ChartTemplateID int64  `json:"chart_template_id"`
}

// GET /api/push_task
func getAllPushTasks(c *gin.Context) {
	db := database.GetDB()

	// 检查表结构与数据库版本
	log.Printf("请求来自 %s 的 getAllPushTasks", c.ClientIP())
	currentVersion, err := database.GetCurrentVersion(db)
	if err != nil {
		log.Printf("获取数据库版本失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取数据库版本失败: %v", err)})
		return
	}

	log.Printf("当前数据库版本: %d", currentVersion)

	if currentVersion < 3 {
		log.Printf("数据库版本过低，需要迁移")
		err := database.MigrateDB(db)
		if err != nil {
			log.Printf("数据库迁移失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("数据库迁移失败: %v", err)})
			return
		}
		log.Printf("数据库迁移完成")
	} else {
		log.Printf("数据库已是最新版本，无需迁移")

		// 检查表结构是否与最新版本一致
		isValid, _, missingColumns, err := database.VerifyTableStructure(db, "push_task")
		if err != nil {
			log.Printf("检查表结构失败: %v", err)
		} else if !isValid {
			log.Printf("检测到表结构不一致，缺少列: %v", missingColumns)

			// 尝试自动修复表结构
			log.Printf("尝试自动修复表结构...")
			if err := database.RepairTableStructure(db, "push_task", false); err != nil {
				log.Printf("修复表结构失败: %v", err)
			} else {
				log.Printf("表结构修复完成")
			}
		}
	}

	// 检查是否存在custom_metric_label列
	hasCustomMetricLabel, err := database.ColumnExists(db, "push_task", "custom_metric_label")
	if err != nil {
		log.Printf("检查列是否存在失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("检查列是否存在失败: %v", err)})
		return
	}

	// 构建查询SQL
	var query string
	var customMetricLabelPart string
	if hasCustomMetricLabel {
		customMetricLabelPart = ", COALESCE(pt.custom_metric_label, '') as custom_metric_label"
	} else {
		customMetricLabelPart = ", '' as custom_metric_label"
	}

	query = fmt.Sprintf(`
		SELECT pt.id, pt.name, pt.source_id, pt.time_range, pt.step, 
			   pt.schedule_interval, COALESCE(pt.last_run_at, '') as last_run_at,
			   pt.enabled, pt.card_title, pt.card_template, pt.metric_label, pt.unit,
			   COALESCE((
				   SELECT ptp.chart_template_id 
				   FROM push_task_promql ptp 
				   WHERE ptp.task_id = pt.id 
				   LIMIT 1
			   ), 0) as chart_template_id%s,
			   COALESCE(pt.button_text, '') as button_text,
			   COALESCE(pt.button_url, '') as button_url,
			   COALESCE(pt.show_data_label, 0) as show_data_label,
			   COALESCE(pt.push_mode, 'chart') as push_mode
		FROM push_task pt
	`, customMetricLabelPart)

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询任务列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("查询任务列表失败: %v", err)})
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var task struct {
			ID                int64
			Name              string
			SourceID          int64
			TimeRange         string
			Step              float64
			SchedInterval     int
			LastRunAt         string
			Enabled           int
			CardTitle         string
			CardTemplate      string
			MetricLabel       string
			Unit              string
			ChartTemplateID   sql.NullInt64 // 修改为 sql.NullInt64
			CustomMetricLabel string
			ButtonText        string
			ButtonURL         string
			ShowDataLabel     int
			PushMode          string
		}

		err := rows.Scan(
			&task.ID, &task.Name, &task.SourceID, &task.TimeRange, &task.Step,
			&task.SchedInterval, &task.LastRunAt, &task.Enabled, &task.CardTitle,
			&task.CardTemplate, &task.MetricLabel, &task.Unit, &task.ChartTemplateID,
			&task.CustomMetricLabel, &task.ButtonText, &task.ButtonURL, &task.ShowDataLabel,
			&task.PushMode,
		)
		if err != nil {
			log.Printf("扫描任务数据失败: %v", err)
			continue
		}

		// 转换为map
		taskMap := map[string]interface{}{
			"id":                  task.ID,
			"name":                task.Name,
			"source_id":           task.SourceID,
			"time_range":          task.TimeRange,
			"step":                task.Step,
			"schedule_interval":   task.SchedInterval,
			"last_run_at":         task.LastRunAt,
			"enabled":             task.Enabled == 1,
			"card_title":          task.CardTitle,
			"card_template":       task.CardTemplate,
			"metric_label":        task.MetricLabel,
			"unit":                task.Unit,
			"chart_template_id":   task.ChartTemplateID.Int64, // 使用 Int64 值
			"custom_metric_label": task.CustomMetricLabel,
			"button_text":         task.ButtonText,
			"button_url":          task.ButtonURL,
			"show_data_label":     task.ShowDataLabel == 1,
			"push_mode":           task.PushMode,
		}

		// 获取任务的发送时间
		sendTimeRows, err := db.Query(`
			SELECT id, weekday, send_time
			FROM push_task_send_time
			WHERE task_id = ?
		`, task.ID)
		if err != nil {
			log.Printf("查询任务发送时间失败: %v", err)
		} else {
			var sendTimes []map[string]interface{}
			for sendTimeRows.Next() {
				var st struct {
					ID       int64
					Weekday  int
					SendTime string
				}
				if err := sendTimeRows.Scan(&st.ID, &st.Weekday, &st.SendTime); err != nil {
					log.Printf("扫描发送时间数据失败: %v", err)
					continue
				}
				sendTimes = append(sendTimes, map[string]interface{}{
					"id":        st.ID,
					"weekday":   st.Weekday,
					"send_time": st.SendTime,
				})
			}
			sendTimeRows.Close()
			taskMap["send_times"] = sendTimes
		}

	// 获取关联的PromQL IDs和每个PromQL的独立配置
	promqlRows, err := db.Query(`
		SELECT ptp.promql_id, ptp.chart_template_id, 
		       ptp.unit, ptp.metric_label, ptp.custom_metric_label, ptp.initial_unit, ptp.display_order,
		       p.name as promql_name
		FROM push_task_promql ptp
		LEFT JOIN promql p ON ptp.promql_id = p.id
		WHERE ptp.task_id = ?
		ORDER BY ptp.display_order ASC, ptp.id ASC
	`, task.ID)
	if err != nil {
		log.Printf("查询任务关联的PromQL失败: %v", err)
	} else {
		var promqlIDs []int64
		var promqlConfigs []map[string]interface{}
	// 使用第一个找到的chart_template_id
	var foundChartTemplateID int64
	for promqlRows.Next() {
		var promqlID int64
		var chartTemplateID sql.NullInt64
		var displayOrder int
		var unit, metricLabel, customMetricLabel, initialUnit, promqlName string
		if err := promqlRows.Scan(&promqlID, &chartTemplateID, 
			&unit, &metricLabel, &customMetricLabel, &initialUnit, &displayOrder, &promqlName); err != nil {
			log.Printf("扫描PromQL数据失败: %v", err)
			continue
		}
		promqlIDs = append(promqlIDs, promqlID)
		
		// 构建 PromQL 配置对象
		promqlConfig := map[string]interface{}{
			"promql_id":           promqlID,
			"promql_name":         promqlName,
			"unit":                unit,
			"metric_label":        metricLabel,
			"custom_metric_label": customMetricLabel,
			"initial_unit":        initialUnit,
			"display_order":       displayOrder,
		}
				if chartTemplateID.Valid {
					promqlConfig["chart_template_id"] = chartTemplateID.Int64
					if foundChartTemplateID == 0 {
						foundChartTemplateID = chartTemplateID.Int64
					}
				} else {
					promqlConfig["chart_template_id"] = 0
				}
				promqlConfigs = append(promqlConfigs, promqlConfig)
			}
			promqlRows.Close()
			taskMap["promql_ids"] = promqlIDs
			taskMap["promql_configs"] = promqlConfigs // 新增：返回每个PromQL的详细配置
			// 如果在关联表中找到了chart_template_id，则使用它
			if foundChartTemplateID > 0 {
				taskMap["chart_template_id"] = foundChartTemplateID
			}
		}

		// 获取绑定的webhook
		webhookRows, err := db.Query(`
			SELECT w.id, w.name, w.url
			FROM feishu_webhook w
			INNER JOIN push_task_webhook ptw ON w.id = ptw.webhook_id
			WHERE ptw.task_id = ?
		`, task.ID)
		if err != nil {
			log.Printf("查询任务绑定的webhook失败: %v", err)
		} else {
			var boundWebhooks []map[string]interface{}
			var webhookIDs []int64
			for webhookRows.Next() {
				var webhook struct {
					ID   int64
					Name string
					URL  string
				}
				if err := webhookRows.Scan(&webhook.ID, &webhook.Name, &webhook.URL); err != nil {
					log.Printf("扫描webhook数据失败: %v", err)
					continue
				}
				boundWebhooks = append(boundWebhooks, map[string]interface{}{
					"id":   webhook.ID,
					"name": webhook.Name,
					"url":  webhook.URL,
				})
				webhookIDs = append(webhookIDs, webhook.ID)
			}
			webhookRows.Close()
			taskMap["bound_webhooks"] = boundWebhooks
			taskMap["webhook_ids"] = webhookIDs
		}

		tasks = append(tasks, taskMap)
	}

	if tasks == nil || len(tasks) == 0 {
		// 创建一个包含占位任务的数组，但格式必须与真实数据相同
		tasks = []map[string]interface{}{
			{
				"id":                  -999,
				"name":                "空占位记录（请勿使用）",
				"source_id":           0,
				"query":               "",
				"time_range":          "30m",
				"step":                300,
				"schedule_interval":   3600,
				"last_run_at":         "",
				"enabled":             false, // 布尔值
				"custom_metric_label": "",
				"card_title":          "",
				"card_template":       "blue",
				"metric_label":        "pod",
				"unit":                "",
				"chart_template_id":   0,
				"bound_webhooks":      []map[string]interface{}{},
				"queries":             []map[string]interface{}{},
				"promql_ids":          []int64{},
				"button_text":         "",
				"button_url":          "",
				"send_times":          []map[string]interface{}{},
			},
		}
	}

	// 直接返回任务数组，不包装在data字段中
	c.JSON(http.StatusOK, tasks)
}

// POST /api/push_task => 创建
func createPushTask(c *gin.Context) {
	var req PushTaskReq
	if err := c.BindJSON(&req); err != nil {
		log.Printf("[createPushTask] 请求数据绑定失败: %v, 来自: %s", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无效的请求数据: %v", err)})
		return
	}

	// 数据验证与默认值设置
	if req.Name == "" {
		req.Name = "未命名任务_" + time.Now().Format("20060102150405")
	}

	if req.TimeRange == "" {
		req.TimeRange = "30m" // 默认时间范围
	}

	if req.Step <= 0 {
		req.Step = 300 // 默认步长5分钟
	}

	if req.SchedInterval <= 0 {
		req.SchedInterval = 3600 // 默认调度间隔1小时
	}

	if req.CardTemplate == "" {
		req.CardTemplate = "blue" // 默认卡片模板
	}

	if req.MetricLabel == "" {
		req.MetricLabel = "pod" // 默认指标标签
	}

	if req.PushMode == "" {
		req.PushMode = "chart" // 默认推送模式为图表
	}

	// 验证必填字段
	if req.SourceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_id is required"})
		return
	}

	// 检查查询是否存在
	hasQueries := len(req.Queries) > 0
	if !hasQueries && req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	// 如果没有提供多个查询，但提供了单个查询，则转换为多查询格式
	if !hasQueries && req.Query != "" {
		req.Queries = []QueryItem{
			{
				Query:           req.Query,
				ChartTemplateID: req.ChartTemplateID,
			},
		}
	}

	// 获取数据库连接
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析时间范围
	duration := parseTimeRange(req.TimeRange)
	log.Printf("[createPushTask] Parsed time_range '%s' to duration: %v", req.TimeRange, duration)

	// 计算步长
	stepSeconds := hoursToSeconds(req.Step)
	log.Printf("[createPushTask] Converted step %.2f hours to %d seconds", req.Step, stepSeconds)

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// 1. 插入主任务记录
	result, err := tx.Exec(`
		INSERT INTO push_task (
			name, source_id, time_range, step, schedule_interval, 
			card_title, card_template, metric_label, unit, enabled,
			custom_metric_label, button_text, button_url, push_mode
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Name, req.SourceID, req.TimeRange, stepSeconds, req.SchedInterval,
		req.CardTitle, req.CardTemplate, req.MetricLabel, req.Unit, true,
		req.CustomMetricLabel, req.ButtonText, req.ButtonURL, req.PushMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取新插入的任务ID
	taskID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 插入发送时间
	for _, sendTime := range req.SendTimes {
		_, err = tx.Exec(`
			INSERT INTO push_task_send_time (
				task_id, weekday, send_time
			) VALUES (?, ?, ?)
		`, taskID, sendTime.Weekday, sendTime.SendTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert send time: " + err.Error()})
			return
		}
	}

	// 2. 插入查询记录
	for _, query := range req.Queries {
		_, err = tx.Exec(`
			INSERT INTO push_task_query (
				task_id, query, chart_template_id
			) VALUES (?, ?, ?)
		`, taskID, query.Query, query.ChartTemplateID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert query: " + err.Error()})
			return
		}
	}

	// 3. 绑定WebHooks
	for _, webhookID := range req.WebhookIDs {
		_, err = tx.Exec(`
			INSERT INTO push_task_webhook (task_id, webhook_id)
			VALUES (?, ?)
		`, taskID, webhookID)
		if err != nil {
			// 忽略重复绑定错误
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bind webhook: " + err.Error()})
				return
			}
		}
	}

	// 在插入任务后，关联 PromQL
	// 优先使用 PromQLConfigs（新格式），如果没有则使用 PromQLIDs（向后兼容）
	if len(req.PromQLConfigs) > 0 {
		// 新格式：使用每个 PromQL 的独立配置
		for _, config := range req.PromQLConfigs {
			// 设置默认值
			unit := config.Unit
			metricLabel := config.MetricLabel
			if metricLabel == "" {
				metricLabel = req.MetricLabel // 如果没有配置，使用任务级别的
				if metricLabel == "" {
					metricLabel = "pod"
				}
			}
			customMetricLabel := config.CustomMetricLabel
			chartTemplateID := config.ChartTemplateID
			if chartTemplateID == 0 {
				chartTemplateID = req.ChartTemplateID // 如果没有配置，使用任务级别的
			}

	_, err = tx.Exec(`
		INSERT INTO push_task_promql (
			task_id, promql_id, chart_template_id, 
			unit, metric_label, custom_metric_label, initial_unit, display_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, taskID, config.PromQLID, chartTemplateID, 
		unit, metricLabel, customMetricLabel, config.InitialUnit, config.DisplayOrder)
			if err != nil {
				log.Printf("[createPushTask] Failed to insert push_task_promql with config: %v", err)
				// 继续处理其他 PromQL，不中断
			}
		}
	} else if len(req.PromQLIDs) > 0 {
		// 旧格式：向后兼容，使用任务级别的配置
		for _, promqlID := range req.PromQLIDs {
			metricLabel := req.MetricLabel
			if metricLabel == "" {
				metricLabel = "pod"
			}
	_, err = tx.Exec(`
		INSERT INTO push_task_promql (
			task_id, promql_id, chart_template_id, 
			unit, metric_label, custom_metric_label, initial_unit, display_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, taskID, promqlID, req.ChartTemplateID,
		req.Unit, metricLabel, req.CustomMetricLabel, "", 0)
			if err != nil {
				log.Printf("[createPushTask] Failed to insert push_task_promql: %v", err)
				// 继续处理其他 PromQL，不中断
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": taskID})
}

// PUT /api/push_task/:id => 更新
func updatePushTask(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" || idStr == "undefined" || idStr == "null" {
		log.Printf("[updatePushTask] 收到无效的任务ID: %s，来自: %s", idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID，请检查前端表单是否正确设置ID值"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("[updatePushTask] 转换任务ID失败: %v，原始值: %s，来自: %s", err, idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID格式"})
		return
	}

	var req PushTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[updatePushTask] 请求数据绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取原始任务数据用于对比
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		log.Printf("[updatePushTask] 数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var oldTask struct {
		Name      string
		TimeRange string
		SendTimes []models.TaskSendTime
	}
	err = db.QueryRow("SELECT name, time_range FROM push_task WHERE id = ?", id).Scan(&oldTask.Name, &oldTask.TimeRange)
	if err != nil {
		log.Printf("[updatePushTask] 获取原始任务数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取原始发送时间
	rows, err := db.Query("SELECT weekday, send_time FROM push_task_send_time WHERE task_id = ?", id)
	if err != nil {
		log.Printf("[updatePushTask] 获取原始发送时间失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var st models.TaskSendTime
			if err := rows.Scan(&st.Weekday, &st.SendTime); err != nil {
				log.Printf("[updatePushTask] 扫描发送时间数据失败: %v", err)
				continue
			}
			oldTask.SendTimes = append(oldTask.SendTimes, st)
		}
	}

	// 记录任务更新前的状态
	log.Printf("[updatePushTask] 开始更新任务 ID=%d", id)
	log.Printf("[updatePushTask] 原任务名称: %s", oldTask.Name)
	log.Printf("[updatePushTask] 新任务名称: %s", req.Name)
	log.Printf("[updatePushTask] 原时间范围: %s", oldTask.TimeRange)
	log.Printf("[updatePushTask] 新时间范围: %s", req.TimeRange)

	// 记录发送时间的变化
	var oldSendTimes []string
	for _, st := range oldTask.SendTimes {
		oldSendTimes = append(oldSendTimes, fmt.Sprintf("周%d %s", st.Weekday, st.SendTime))
	}
	var newSendTimes []string
	for _, st := range req.SendTimes {
		newSendTimes = append(newSendTimes, fmt.Sprintf("周%d %s", st.Weekday, st.SendTime))
	}
	log.Printf("[updatePushTask] 原发送时间: %s", strings.Join(oldSendTimes, ", "))
	log.Printf("[updatePushTask] 新发送时间: %s", strings.Join(newSendTimes, ", "))

	// 验证必填字段
	if req.Name == "" {
		log.Printf("[updatePushTask] 任务名称为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.SourceID == 0 {
		log.Printf("[updatePushTask] 数据源ID为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_id is required"})
		return
	}

	// 检查查询是否存在
	hasQueries := len(req.Queries) > 0
	if !hasQueries && req.Query == "" {
		log.Printf("[updatePushTask] 查询语句为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	// 如果没有提供多个查询，但提供了单个查询，则转换为多查询格式
	if !hasQueries && req.Query != "" {
		req.Queries = []QueryItem{
			{
				Query:           req.Query,
				ChartTemplateID: req.ChartTemplateID,
			},
		}
	}

	// 解析时间范围
	duration := parseTimeRange(req.TimeRange)
	log.Printf("[updatePushTask] 解析时间范围 '%s' 为: %v", req.TimeRange, duration)

	// 计算步长
	stepSeconds := hoursToSeconds(req.Step)
	log.Printf("[updatePushTask] 计算步长: %.2f 小时 -> %d 秒", req.Step, stepSeconds)

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[updatePushTask] 开始事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// 检查任务是否存在
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM push_task WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		log.Printf("[updatePushTask] 检查任务是否存在失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		log.Printf("[updatePushTask] 任务不存在: ID=%d", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 处理默认值
	pushMode := req.PushMode
	if pushMode == "" {
		pushMode = "chart"
	}

	// 更新主任务记录
	updateSQL := `
		UPDATE push_task SET
			name = ?, source_id = ?, time_range = ?, step = ?, 
			schedule_interval = ?, card_title = ?, card_template = ?, 
			metric_label = ?, unit = ?, custom_metric_label = ?, 
			button_text = ?, button_url = ?, chart_template_id = ?,
			show_data_label = ?, push_mode = ?
		WHERE id = ?
	`
	updateArgs := []interface{}{
		req.Name, req.SourceID, req.TimeRange, stepSeconds,
		req.SchedInterval, req.CardTitle, req.CardTemplate,
		req.MetricLabel, req.Unit, req.CustomMetricLabel,
		req.ButtonText, req.ButtonURL, req.ChartTemplateID,
		map[bool]int{true: 1, false: 0}[req.ShowDataLabel], // 转换布尔值为整数
		pushMode,
		id,
	}

	// 执行更新
	result, err := tx.Exec(updateSQL, updateArgs...)
	if err != nil {
		log.Printf("[updatePushTask] 更新任务主记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("[updatePushTask] 更新任务主记录成功，影响行数: %d", rowsAffected)

	// 更新发送时间
	// 1. 删除旧的发送时间
	result, err = tx.Exec("DELETE FROM push_task_send_time WHERE task_id = ?", id)
	if err != nil {
		log.Printf("[updatePushTask] 删除旧的发送时间失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old send times: " + err.Error()})
		return
	}
	rowsAffected, _ = result.RowsAffected()
	log.Printf("[updatePushTask] 删除旧的发送时间成功，删除数量: %d", rowsAffected)

	// 2. 只有在有新的发送时间时才插入
	if len(req.SendTimes) > 0 {
		for _, sendTime := range req.SendTimes {
			// 验证发送时间的格式
			if sendTime.SendTime == "" || sendTime.Weekday < 0 || sendTime.Weekday > 7 {
				log.Printf("[updatePushTask] 无效的发送时间配置: weekday=%d, time=%s",
					sendTime.Weekday, sendTime.SendTime)
				continue
			}

			_, err = tx.Exec(`
				INSERT INTO push_task_send_time (
					task_id, weekday, send_time
				) VALUES (?, ?, ?)
			`, id, sendTime.Weekday, sendTime.SendTime)
			if err != nil {
				log.Printf("[updatePushTask] 插入新的发送时间失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert send time: " + err.Error()})
				return
			}
		}
		log.Printf("[updatePushTask] 插入新的发送时间成功，插入数量: %d", len(req.SendTimes))
	} else {
		log.Printf("[updatePushTask] 没有新的发送时间需要插入")
	}

	// 更新关联的 PromQL
	// 优先使用 PromQLConfigs（新格式），如果没有则使用 PromQLIDs（向后兼容）
	if len(req.PromQLConfigs) > 0 || len(req.PromQLIDs) > 0 {
		// 先删除旧的关联
		result, err = tx.Exec("DELETE FROM push_task_promql WHERE task_id = ?", id)
		if err != nil {
			log.Printf("[updatePushTask] 删除旧的PromQL关联失败: %v", err)
		} else {
			rowsAffected, _ = result.RowsAffected()
			log.Printf("[updatePushTask] 删除旧的PromQL关联成功，删除数量: %d", rowsAffected)
		}

		// 添加新的关联
		if len(req.PromQLConfigs) > 0 {
			// 新格式：使用每个 PromQL 的独立配置
			for _, config := range req.PromQLConfigs {
				// 设置默认值
				unit := config.Unit
				metricLabel := config.MetricLabel
				if metricLabel == "" {
					metricLabel = req.MetricLabel
					if metricLabel == "" {
						metricLabel = "pod"
					}
				}
				customMetricLabel := config.CustomMetricLabel
				chartTemplateID := config.ChartTemplateID
				if chartTemplateID == 0 {
					chartTemplateID = req.ChartTemplateID
				}

		_, err = tx.Exec(`
			INSERT INTO push_task_promql (
				task_id, promql_id, chart_template_id, 
				unit, metric_label, custom_metric_label, initial_unit, display_order
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, id, config.PromQLID, chartTemplateID,
			unit, metricLabel, customMetricLabel, config.InitialUnit, config.DisplayOrder)
				if err != nil {
					log.Printf("[updatePushTask] 插入新的PromQL关联失败: promql_id=%d, error=%v", config.PromQLID, err)
				}
			}
			log.Printf("[updatePushTask] 插入新的PromQL关联成功，插入数量: %d", len(req.PromQLConfigs))
		} else if len(req.PromQLIDs) > 0 {
			// 旧格式：向后兼容
			for _, promqlID := range req.PromQLIDs {
				metricLabel := req.MetricLabel
				if metricLabel == "" {
					metricLabel = "pod"
				}
		_, err = tx.Exec(`
			INSERT INTO push_task_promql (
				task_id, promql_id, chart_template_id, 
				unit, metric_label, custom_metric_label, initial_unit, display_order
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, id, promqlID, req.ChartTemplateID,
			req.Unit, metricLabel, req.CustomMetricLabel, "", 0)
				if err != nil {
					log.Printf("[updatePushTask] 插入新的PromQL关联失败: promql_id=%d, error=%v", promqlID, err)
				}
			}
			log.Printf("[updatePushTask] 插入新的PromQL关联成功，插入数量: %d", len(req.PromQLIDs))
		}
	}

	// 更新 WebHook 绑定
	if len(req.WebhookIDs) > 0 {
		// 先删除旧的绑定
		result, err = tx.Exec("DELETE FROM push_task_webhook WHERE task_id = ?", id)
		if err != nil {
			log.Printf("[updatePushTask] 删除旧的WebHook绑定失败: %v", err)
		} else {
			rowsAffected, _ = result.RowsAffected()
			log.Printf("[updatePushTask] 删除旧的WebHook绑定成功，删除数量: %d", rowsAffected)
		}

		// 添加新的绑定
		for _, webhookID := range req.WebhookIDs {
			_, err = tx.Exec(`
				INSERT INTO push_task_webhook (task_id, webhook_id)
				VALUES (?, ?)
			`, id, webhookID)
			if err != nil {
				log.Printf("[updatePushTask] 插入新的WebHook绑定失败: webhook_id=%d, error=%v", webhookID, err)
			}
		}
		log.Printf("[updatePushTask] 插入新的WebHook绑定成功，插入数量: %d", len(req.WebhookIDs))
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Printf("[updatePushTask] 提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[updatePushTask] 任务更新成功完成: ID=%d", id)
	c.JSON(http.StatusOK, gin.H{"message": "task updated successfully"})
}

// PUT /api/push_task/:id/toggle
func togglePushTask(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" || idStr == "undefined" || idStr == "null" {
		log.Printf("[togglePushTask] 收到无效的任务ID: %s，来自: %s", idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID，请检查前端请求"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("[togglePushTask] 转换任务ID失败: %v，原始值: %s，来自: %s", err, idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID格式"})
		return
	}

	// 使用可以同时接收布尔值和整数的结构
	var req struct {
		Enabled json.RawMessage `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析enabled值，支持布尔和整数类型
	var enabledValue bool
	var enabledIntValue int

	// 尝试解析为布尔值
	if err := json.Unmarshal(req.Enabled, &enabledValue); err == nil {
		// 成功解析为布尔值
	} else {
		// 尝试解析为整数
		if err := json.Unmarshal(req.Enabled, &enabledIntValue); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "enabled must be a boolean or integer"})
			return
		}
		enabledValue = enabledIntValue != 0
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec("UPDATE push_task SET enabled = ? WHERE id = ?",
		map[bool]int{true: 1, false: 0}[enabledValue], id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DELETE /api/push_task/:id
func deletePushTask(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" || idStr == "undefined" || idStr == "null" {
		log.Printf("[deletePushTask] 收到无效的任务ID: %s，来自: %s", idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID，请检查前端请求"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("[deletePushTask] 转换任务ID失败: %v，原始值: %s，来自: %s", err, idStr, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID格式"})
		return
	}

	db, dberr := database.SetupDB("./data/app.db")
	if dberr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": dberr.Error()})
		return
	}
	
	// 删除所有关联数据
	log.Printf("[deletePushTask] 开始删除任务 ID=%d 及其关联数据", id)
	
	// 1. 删除 push_task_webhook
	_, err = db.Exec("DELETE FROM push_task_webhook WHERE task_id=?", id)
	if err != nil {
		log.Printf("[deletePushTask] 删除 push_task_webhook 失败: %v", err)
	}
	
	// 2. 删除 push_task_promql（重要！）
	result, err := db.Exec("DELETE FROM push_task_promql WHERE task_id=?", id)
	if err != nil {
		log.Printf("[deletePushTask] 删除 push_task_promql 失败: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		log.Printf("[deletePushTask] 删除 push_task_promql 成功，删除了 %d 条记录", rows)
	}
	
	// 3. 删除 push_task_send_time
	_, err = db.Exec("DELETE FROM push_task_send_time WHERE task_id=?", id)
	if err != nil {
		log.Printf("[deletePushTask] 删除 push_task_send_time 失败: %v", err)
	}
	
	// 4. 删除 push_task_query（旧格式）
	_, err = db.Exec("DELETE FROM push_task_query WHERE task_id=?", id)
	if err != nil {
		log.Printf("[deletePushTask] 删除 push_task_query 失败: %v", err)
	}
	
	// 5. 最后删除 push_task 主记录
	res, delErr := db.Exec("DELETE FROM push_task WHERE id=?", id)
	if delErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": delErr.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "push_task not found"})
		return
	}
	
	log.Printf("[deletePushTask] 任务删除成功 ID=%d", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted", "task_id": id})
}

// -------------- push_task_webhook --------------

type PushTaskWebhookReq struct {
	TaskID    int64 `json:"task_id"`
	WebhookID int64 `json:"webhook_id"`
}

// POST /api/push_task_webhook
func createPushTaskWebhook(c *gin.Context) {
	var req PushTaskWebhookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 先检查是否已存在
	var existedID int64
	err = db.QueryRow(`
        SELECT id
          FROM push_task_webhook
         WHERE task_id=? AND webhook_id=?
    `, req.TaskID, req.WebhookID).Scan(&existedID)
	if err == nil && existedID > 0 {
		// 已有记录
		c.JSON(http.StatusConflict, gin.H{"error": "already bound"})
		return
	} else if err != sql.ErrNoRows && err != nil {
		// 其它错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 不存在 -> 插入
	stmt, e := db.Prepare(`
        INSERT INTO push_task_webhook(task_id, webhook_id)
        VALUES(?, ?)
    `)
	if e != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": e.Error()})
		return
	}
	res, e2 := stmt.Exec(req.TaskID, req.WebhookID)
	if e2 != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate binding"})
		return
	}
	newID, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"id":         newID,
		"task_id":    req.TaskID,
		"webhook_id": req.WebhookID,
	})
	log.Printf("[createPushTaskWebhook] body = %#v", req)
	log.Printf("[createPushTaskWebhook] inserted => newID=%d, task_id=%d, webhook_id=%d", newID, req.TaskID, req.WebhookID)
}

// DELETE /api/push_task_webhook/:taskId/:webhookId
func deletePushTaskWebhook(c *gin.Context) {
	taskId := c.Param("taskId")
	webhookId := c.Param("webhookId")

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec(`
		DELETE FROM push_task_webhook 
		 WHERE task_id = ? AND webhook_id = ?
	`, taskId, webhookId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "binding not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unbound successfully"})
}

// ========== 可选: runSingleTaskPush + insertPushStatus (若需立即推送) ==========

func runSingleTaskPush(db *sql.DB, taskID int64) {
	// 获取任务详情
	var sourceID int64
	var name, timeRange, cardTitle, cardTemplate, metricLabel, unit, buttonText, buttonURL, customMetricLabel string
	var step float64
	var enabled int
	var showDataLabel sql.NullInt64 // 使用 sql.NullInt64 来处理 NULL 值

	log.Printf("[scheduler] Starting task execution for task ID=%d", taskID)

	err := db.QueryRow(`
		SELECT source_id, name, time_range, step, 
		       card_title, card_template, metric_label, unit,
		       button_text, button_url, enabled, COALESCE(show_data_label, 0) as show_data_label,
		       COALESCE(custom_metric_label, '') as custom_metric_label
		FROM push_task 
		WHERE id = ?
	`, taskID).Scan(&sourceID, &name, &timeRange, &step,
		&cardTitle, &cardTemplate, &metricLabel, &unit,
		&buttonText, &buttonURL, &enabled, &showDataLabel, &customMetricLabel)
	if err != nil {
		log.Printf("[scheduler] Error querying task: %v", err)
		return
	}

	// 检查任务是否启用
	if enabled != 1 {
		log.Printf("[scheduler] Task ID=%d (%s) is disabled, skipping execution", taskID, name)
		return
	}

	// 更新最后运行时间
	_, err = db.Exec("UPDATE push_task SET last_run_at = ? WHERE id = ?",
		time.Now().Format("2006-01-02 15:04:05"), taskID)
	if err != nil {
		log.Printf("[scheduler] Error updating last_run_at: %v", err)
		// 继续执行，不因更新时间失败而中断
	}

	log.Printf("[runSingleTaskPush] Processing task ID=%d, name=%s, time_range=%s, step=%v, custom_metric_label=%s",
		taskID, name, timeRange, step, customMetricLabel)

	// 获取数据源URL
	var sourceURL string
	err = db.QueryRow("SELECT url FROM metrics_source WHERE id = ?", sourceID).Scan(&sourceURL)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error querying metrics source: %v", err)
		return
	}

	// 查询任务的所有查询
	rows, err := db.Query(`
		SELECT ptp.promql_id, p.query, ptp.chart_template_id, p.name,
		       COALESCE(ptp.unit, '') as unit,
		       COALESCE(ptp.initial_unit, '') as initial_unit
		FROM push_task_promql ptp
		JOIN promql p ON ptp.promql_id = p.id
		WHERE ptp.task_id = ?
	`, taskID)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error querying task promqls: %v", err)
		return
	}
	defer rows.Close()

	var queries []struct {
		Query           string
		ChartTemplateID int64
		PromQLName      string
		Unit            string
		InitialUnit     string
	}

	for rows.Next() {
		var promqlID int64
		var q struct {
			Query           string
			ChartTemplateID int64
			PromQLName      string
			Unit            string
			InitialUnit     string
		}
		if err := rows.Scan(&promqlID, &q.Query, &q.ChartTemplateID, &q.PromQLName, &q.Unit, &q.InitialUnit); err != nil {
			log.Printf("[runSingleTaskPush] Error scanning promql row: %v", err)
			continue
		}
		queries = append(queries, q)
	}

	// 如果没有找到查询，尝试使用旧的单查询格式
	if len(queries) == 0 {
		// 尝试从push_task_query表中获取查询
		queryRows, err := db.Query(`
			SELECT query, chart_template_id
			FROM push_task_query
			WHERE task_id = ?
		`, taskID)
		if err == nil {
			defer queryRows.Close()
			for queryRows.Next() {
				var q struct {
					Query           string
					ChartTemplateID int64
					PromQLName      string
					Unit            string
					InitialUnit     string
				}
				if err := queryRows.Scan(&q.Query, &q.ChartTemplateID); err != nil {
					log.Printf("[runSingleTaskPush] Error scanning query row: %v", err)
					continue
				}
				q.PromQLName = "" // 设置为空字符串
				q.Unit = unit // 使用任务级别单位
				q.InitialUnit = "" // 旧格式没有 initial_unit
				queries = append(queries, q)
			}
		} else {
			log.Printf("[runSingleTaskPush] Error querying push_task_query: %v", err)
		}

		// 如果仍然没有查询，尝试从push_task表中获取
		if len(queries) == 0 {
			var query string
			var chartTemplateID int64
			err := db.QueryRow(`
				SELECT query, chart_template_id
				FROM push_task
				WHERE id = ?
			`, taskID).Scan(&query, &chartTemplateID)

			if err == nil && query != "" {
				queries = append(queries, struct {
					Query           string
					ChartTemplateID int64
					PromQLName      string
					Unit            string
					InitialUnit     string
				}{
					Query:           query,
					ChartTemplateID: chartTemplateID,
					PromQLName:      "",
					Unit:            unit, // 使用任务级别单位
					InitialUnit:     "", // 旧格式没有 initial_unit
				})
			}
		}
	}

	// 如果仍然没有查询，记录错误并返回
	if len(queries) == 0 {
		log.Printf("[runSingleTaskPush] No queries found for task ID=%d", taskID)
		return
	}

	// 解析time_range为持续时间
	duration := parseTimeRange(timeRange)
	log.Printf("[runSingleTaskPush] Parsed time_range '%s' to duration: %v", timeRange, duration)

	// 计算时间范围
	end := time.Now()
	start := end.Add(-duration)
	log.Printf("[runSingleTaskPush] Time range: start=%v, end=%v", start, end)

	// 确保我们获得最近的数据，特别是对于较长的时间范围
	if duration > 72*time.Hour { // 对于超过3天的查询
		log.Printf("[runSingleTaskPush] Long time range detected (%v), ensuring recent data is prioritized", duration)
		// 记录原始时间范围，以便于调试
		log.Printf("[runSingleTaskPush] Original time range: %s to %s",
			start.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"))

		// 强制使用当前时间作为结束时间
		end = time.Now()
		start = end.Add(-duration)
		log.Printf("[runSingleTaskPush] Adjusted time range to ensure recent data: %s to %s",
			start.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"))
	}

	// 原始的step（小时）转换为时间间隔
	origStepDuration := time.Duration(step*float64(time.Hour.Seconds())) * time.Second
	log.Printf("[runSingleTaskPush] Original step from database: %v seconds (%v)",
		origStepDuration.Seconds(), origStepDuration)

	// 根据时间范围计算合适的步长
	calculatedStep := service.GetDurationStep(duration)
	log.Printf("[runSingleTaskPush] Calculated step based on duration %v: %v",
		duration, calculatedStep)

	// 确保步长合理（至少产生5个数据点）
	expectedPoints := duration.Seconds() / calculatedStep.Seconds()
	log.Printf("[runSingleTaskPush] Expected data points with calculated step: %.1f", expectedPoints)

	if expectedPoints < 5 {
		// 如果数据点太少，调整步长确保至少有10个数据点
		adjustedStep := time.Duration(duration.Seconds()/10) * time.Second
		log.Printf("[runSingleTaskPush] Adjusting step to ensure at least 10 data points: %v -> %v",
			calculatedStep, adjustedStep)
		calculatedStep = adjustedStep
		expectedPoints = float64(duration) / float64(adjustedStep)
		log.Printf("[runSingleTaskPush] Final expected points: %.2f", expectedPoints)
	} else if duration > 120*time.Hour && expectedPoints > 20 { // 对于超过5天的时间范围
		// 增加日志以便于调试
		log.Printf("[runSingleTaskPush] Long time range with many data points, might need prioritization")
	}

	// 为每个查询获取数据点
	var allDataPoints []models.QueryDataPoints
	for i, query := range queries {
		// 获取图表类型
		var chartType string
		err = db.QueryRow("SELECT chart_type FROM chart_template WHERE id = ?", query.ChartTemplateID).Scan(&chartType)
		if err != nil {
			log.Printf("[runSingleTaskPush] Error querying chart template (ID=%d): %v (using default 'area')",
				query.ChartTemplateID, err)
			chartType = "area" // 默认为area类型
		}

		// 确保使用飞书支持的图表类型
		chartType = service.GetSupportedChartType(chartType)
		log.Printf("[runSingleTaskPush] Using chart type: %s for query %d", chartType, i+1)

		// 获取数据点
		log.Printf("[runSingleTaskPush] Fetching metrics for query %d: %s", i+1, query.Query)
		log.Printf("[runSingleTaskPush] Using custom_metric_label: '%s', fallback metric_label: '%s'",
			customMetricLabel, metricLabel)

	dataPoints, err := service.FetchMetrics(sourceURL, query.Query, start, end, calculatedStep, metricLabel, customMetricLabel, query.InitialUnit, query.Unit)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error fetching metrics for query %d: %v", i+1, err)
		continue
	}

		log.Printf("[runSingleTaskPush] Fetched %d data points for query %d", len(dataPoints), i+1)

	// 添加到结果集
	chartTitle := query.PromQLName
	if chartTitle == "" {
		// 如果PromQL名称为空，使用默认的"查询 N"
		chartTitle = fmt.Sprintf("查询 %d", i+1)
	}

	allDataPoints = append(allDataPoints, models.QueryDataPoints{
		DataPoints: dataPoints,
		ChartType:  chartType,
		ChartTitle: chartTitle,
		Unit:       query.Unit, // 使用每个查询的独立单位
	})
}

	// 如果没有获取到任何数据点，记录错误并返回
	if len(allDataPoints) == 0 {
		log.Printf("[runSingleTaskPush] No data points fetched for any query, task ID=%d", taskID)
		return
	}

	// 获取所有绑定的webhook
	log.Printf("[runSingleTaskPush] Querying webhooks for task ID=%d", taskID)
	webhookRows, err := db.Query(`
        SELECT w.id, w.url 
        FROM feishu_webhook w
        JOIN push_task_webhook ptw ON w.id = ptw.webhook_id
        WHERE ptw.task_id = ?
    `, taskID)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error querying webhooks: %v", err)
		return
	}
	defer webhookRows.Close()

	var webhooks []struct {
		ID  int64
		URL string
	}

	for webhookRows.Next() {
		var wh struct {
			ID  int64
			URL string
		}
		if err := webhookRows.Scan(&wh.ID, &wh.URL); err != nil {
			log.Printf("[runSingleTaskPush] Error scanning webhook row: %v", err)
			continue
		}
		webhooks = append(webhooks, wh)
	}

	// 检查是否找到了webhook
	if len(webhooks) == 0 {
		log.Printf("[runSingleTaskPush] No webhooks found for task ID=%d", taskID)
		return
	}

	log.Printf("[runSingleTaskPush] Found %d webhooks for task ID=%d", len(webhooks), taskID)

	// 发送到每个webhook
	for _, wh := range webhooks {
		log.Printf("[runSingleTaskPush] Sending to webhook ID=%d, URL=%s", wh.ID, wh.URL)
		err := service.SendFeishuStandardChart(wh.URL, allDataPoints, cardTitle, cardTemplate, unit, buttonText, buttonURL, showDataLabel.Int64 == 1)
		if err != nil {
			log.Printf("[runSingleTaskPush] Error sending to webhook: %v", err)
			insertPushStatus(db, sourceID, wh.ID, err)
		} else {
			log.Printf("[runSingleTaskPush] Successfully sent to webhook ID=%d", wh.ID)
			insertPushStatus(db, sourceID, wh.ID, nil)
		}
	}
}

func insertPushStatus(db *sql.DB, sourceID, webhookID int64, err error) {
	status := "success"
	logTxt := "ok"
	if err != nil {
		status = "failed"
		logTxt = err.Error()
	}
	db.Exec(`
        INSERT INTO push_status (source_id, webhook_id, last_run_time, status, log)
        VALUES (?, ?, datetime('now'), ?, ?)
    `, sourceID, webhookID, status, logTxt)
}

// -------------- chart_template --------------

type ChartTemplateReq struct {
	Name      string `json:"name"`
	ChartType string `json:"chart_type"`
}

// GET /api/chart_template
func getChartTemplates(c *gin.Context) {
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT id, name, chart_type FROM chart_template")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var templates []models.ChartTemplate
	for rows.Next() {
		var tmpl models.ChartTemplate
		if err := rows.Scan(&tmpl.ID, &tmpl.Name, &tmpl.ChartType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		templates = append(templates, tmpl)
	}
	c.JSON(http.StatusOK, templates)
}

// POST /api/chart_template
func createChartTemplate(c *gin.Context) {
	var req ChartTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("INSERT INTO chart_template(name, chart_type) VALUES (?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := stmt.Exec(req.Name, req.ChartType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newID, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"id":         newID,
		"name":       req.Name,
		"chart_type": req.ChartType,
	})
}

// PUT /api/chart_template/:id
func updateChartTemplate(c *gin.Context) {
	idStr := c.Param("id")
	var req ChartTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE chart_template SET name=?, chart_type=? WHERE id=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := stmt.Exec(req.Name, req.ChartType, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated", "id": idStr})
}

// DELETE /api/chart_template/:id
func deleteChartTemplate(c *gin.Context) {
	idStr := c.Param("id")
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("DELETE FROM chart_template WHERE id=?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := stmt.Exec(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted", "id": idStr})
}

// PromQLReq 是创建或更新 PromQL 的请求结构
type PromQLReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Category    string `json:"category"`
}

// 获取所有 PromQL 查询
func getPromQLs(c *gin.Context) {
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query(`
		SELECT id, name, description, query, category, created_at, updated_at
		FROM promql
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var promqls []models.PromQL
	for rows.Next() {
		var promql models.PromQL
		if err := rows.Scan(&promql.ID, &promql.Name, &promql.Description, &promql.Query, &promql.Category, &promql.CreatedAt, &promql.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		promqls = append(promqls, promql)
	}

	c.JSON(http.StatusOK, promqls)
}

// 创建新的 PromQL 查询
func createPromQL(c *gin.Context) {
	var req PromQLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and query are required"})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec(`
		INSERT INTO promql (name, description, query, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, req.Name, req.Description, req.Query, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// 更新 PromQL 查询
func updatePromQL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req PromQLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and query are required"})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec(`
		UPDATE promql
		SET name = ?, description = ?, query = ?, category = ?, updated_at = datetime('now')
		WHERE id = ?
	`, req.Name, req.Description, req.Query, req.Category, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// 删除 PromQL 查询
func deletePromQL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否有任务使用此 PromQL
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM push_task_promql WHERE promql_id = ?", id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法删除：此 PromQL 正在被 %d 个任务使用", count)})
		return
	}

	_, err = db.Exec("DELETE FROM promql WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// runPushTaskHandler 手动执行任务的处理函数
func runPushTaskHandler(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查任务是否存在且启用
	var enabled int
	err = db.QueryRow("SELECT enabled FROM push_task WHERE id = ?", taskID).Scan(&enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务状态失败"})
		}
		return
	}

	if enabled != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务未启用"})
		return
	}

	// 使用scheduler包中的ForceRunSingleTaskPush函数执行任务（跳过间隔检查）
	go func() {
		if err := scheduler.ForceRunSingleTaskPush(db, taskID); err != nil {
			log.Printf("[runPushTaskHandler] 任务 %d 执行失败: %v", taskID, err)
		} else {
			log.Printf("[runPushTaskHandler] 任务 %d 手动执行成功", taskID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "任务立即执行已开始",
		"task_id": taskID,
	})
}

// initScheduler 初始化调度器
func initScheduler(db *sql.DB) {
	// 创建一个定时器，每分钟触发一次
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				processPushTasks(db)
			}
		}
	}()
	log.Printf("[scheduler] Scheduler initialized and running every minute")
}

// RegisterRoutes: 主路由注册
func RegisterRoutes(r *gin.Engine) {
	// 初始化数据库连接
	db := database.GetDB()

	// 启动调度器
	initScheduler(db)

	// 静态文件服务 - 仅处理 /assets 路径
	r.Static("/assets", "./web/assets")

	// 公开路由 - 不需要认证
	// 添加版本API
	r.GET("/api/version", getVersionAPI)

	// 认证相关路由
	r.POST("/api/auth/login", login)

	// 需要认证的路由组
	authGroup := r.Group("/api")
	authGroup.Use(middleware.JWTAuth())
	{
		// 用户相关
		authGroup.GET("/user/current", getCurrentUser)
		authGroup.PUT("/user/password", changePassword)
		authGroup.PUT("/user/info", updateUserInfo)

		// 1. metrics_source 数据源管理
		authGroup.GET("/metrics_source", getMetricsSources)
		authGroup.POST("/metrics_source", createMetricsSource)
		authGroup.PUT("/metrics_source/:id", updateMetricsSource)
		authGroup.DELETE("/metrics_source/:id", deleteMetricsSource)

		// feishu_webhook
		authGroup.GET("/feishu_webhook", getFeishuWebhooks)
		authGroup.POST("/feishu_webhook", createFeishuWebhook)
		authGroup.PUT("/feishu_webhook/:id", updateFeishuWebhook)
		authGroup.DELETE("/feishu_webhook/:id", deleteFeishuWebhook)

		// push_task
		authGroup.GET("/push_task", getAllPushTasks)
		authGroup.POST("/push_task", createPushTask)
		authGroup.PUT("/push_task/:id", updatePushTask)
		authGroup.PUT("/push_task/:id/toggle", togglePushTask)
		authGroup.DELETE("/push_task/:id", deletePushTask)
		authGroup.POST("/push_task/:id/run", runPushTaskHandler)

		// push_task_webhook
		authGroup.POST("/push_task_webhook", createPushTaskWebhook)
		authGroup.DELETE("/push_task_webhook/:taskId/:webhookId", deletePushTaskWebhook)

		// chart_template routes
		authGroup.GET("/chart_template", getChartTemplates)
		authGroup.POST("/chart_template", createChartTemplate)
		authGroup.PUT("/chart_template/:id", updateChartTemplate)
		authGroup.DELETE("/chart_template/:id", deleteChartTemplate)

		// PromQL 相关路由
		authGroup.GET("/promqls", getPromQLs)
		authGroup.POST("/promql", createPromQL)
		authGroup.PUT("/promql/:id", updatePromQL)
		authGroup.DELETE("/promql/:id", deletePromQL)

		// 发送记录
		authGroup.GET("/send_records", handler.HandleGetSendRecords)
	}
}

// getVersionAPI 返回应用版本信息
func getVersionAPI(c *gin.Context) {
	// 读取VERSION文件
	data, err := ioutil.ReadFile("VERSION")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"version": "unknown",
			"error":   err.Error(),
		})
		return
	}

	version := strings.TrimSpace(string(data))
	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}

// 在创建和更新任务时，将小时转换为秒
func hoursToSeconds(hours float64) int {
	seconds := int(hours * 3600)
	log.Printf("[hoursToSeconds] Converting %f hours to %d seconds", hours, seconds)

	// 确保步长不会太大，对于长时间范围（如1个月）
	// 限制最大步长为 5 天 (432000 秒)
	maxStepSeconds := 5 * 24 * 3600
	if seconds > maxStepSeconds {
		log.Printf("[hoursToSeconds] Limiting step from %d to %d seconds (5 days max)",
			seconds, maxStepSeconds)
		seconds = maxStepSeconds
	}

	return seconds
}

// parseTimeRange 将时间范围字符串解析为 time.Duration
func parseTimeRange(timeRange string) time.Duration {
	log.Printf("[parseTimeRange] Parsing time range: %s", timeRange)

	// 处理天数格式 (例如 "1d", "5d")
	if strings.HasSuffix(timeRange, "d") {
		dayStr := strings.TrimSuffix(timeRange, "d")
		days, err := strconv.Atoi(dayStr)
		if err == nil && days > 0 {
			duration := time.Duration(days) * 24 * time.Hour
			log.Printf("[parseTimeRange] Parsed %s as %v (days)", timeRange, duration)
			return duration
		}
		log.Printf("[parseTimeRange] Invalid day format: %s, fallback=24h", timeRange)
		return 24 * time.Hour
	}

	// 处理月份格式 (例如 "1M", "3M")
	if strings.HasSuffix(timeRange, "M") {
		monthStr := strings.TrimSuffix(timeRange, "M")
		months, err := strconv.Atoi(monthStr)
		if err == nil && months > 0 {
			// 假设一个月为30天
			duration := time.Duration(months) * 30 * 24 * time.Hour
			log.Printf("[parseTimeRange] Parsed %s as %v (months)", timeRange, duration)
			return duration
		}
		log.Printf("[parseTimeRange] Invalid month format: %s, fallback=24h", timeRange)
		return 24 * time.Hour
	}

	// 处理标准时间单位
	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		// 如果没有单位，默认为小时
		if val, err := strconv.Atoi(timeRange); err == nil {
			duration = time.Duration(val) * time.Hour
			log.Printf("[parseTimeRange] Parsed %s as %v (hours)", timeRange, duration)
			return duration
		}

		log.Printf("[parseTimeRange] Failed to parse time range: %s, error: %v, fallback=24h", timeRange, err)
		return 24 * time.Hour
	}

	log.Printf("[parseTimeRange] Parsed %s as %v", timeRange, duration)
	return duration
}

// processPushTasks 处理所有已启用的任务
func processPushTasks(db *sql.DB) {
	now := time.Now()
	log.Printf("[scheduler] %s processPushTasks start...", now.Format("2006/01/02 15:04:05"))

	// 获取所有已启用的任务，使用 COALESCE 处理 NULL 值
	rows, err := db.Query(`
        SELECT pt.id, pt.name, pt.enabled, 
               COALESCE(pt.last_run_at, '') as last_run_at,
               GROUP_CONCAT(COALESCE(pts.weekday || ' ' || pts.send_time, '')) as send_times
        FROM push_task pt
        LEFT JOIN push_task_send_time pts ON pt.id = pts.task_id
        WHERE pt.enabled = 1
        GROUP BY pt.id
    `)
	if err != nil {
		log.Printf("[scheduler] Error querying enabled tasks: %v", err)
		return
	}
	defer rows.Close()

	// 当前时间信息
	currentWeekday := int(now.Weekday())
	currentTime := now.Format("15:04")

	log.Printf("[scheduler] Current time: %s, weekday: %d", currentTime, currentWeekday)

	// 遍历所有启用的任务
	for rows.Next() {
		var task struct {
			ID        int64
			Name      string
			Enabled   int
			LastRunAt string
			SendTimes string
		}
		if err := rows.Scan(&task.ID, &task.Name, &task.Enabled, &task.LastRunAt, &task.SendTimes); err != nil {
			log.Printf("[scheduler] Error scanning task row: %v", err)
			continue
		}

		log.Printf("[scheduler] Processing task ID=%d (%s)", task.ID, task.Name)
		log.Printf("[scheduler]   - Last Run: %s", task.LastRunAt)
		log.Printf("[scheduler]   - Send Times: %s", task.SendTimes)

		// 检查是否应该在当前时间执行
		shouldExecute := false
		if task.SendTimes != "" {
			for _, timeStr := range strings.Split(task.SendTimes, ",") {
				if timeStr == "" {
					continue
				}
				parts := strings.Split(strings.TrimSpace(timeStr), " ")
				if len(parts) != 2 {
					continue
				}

				weekday, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}

				sendTime := parts[1]
				// 如果是当前星期几且时间匹配，则执行
				if weekday == currentWeekday && sendTime == currentTime {
					shouldExecute = true
					log.Printf("[scheduler] Task ID=%d scheduled for current time (weekday=%d, time=%s)",
						task.ID, weekday, sendTime)
					break
				}
			}
		}

		if !shouldExecute {
			continue
		}

		log.Printf("[scheduler] Executing task ID=%d (%s)", task.ID, task.Name)
		// 使用scheduler包中的函数，确保任务执行有互斥控制
		go scheduler.RunSingleTaskPush(db, task.ID)
	}

	log.Printf("[scheduler] processPushTasks completed")
}
