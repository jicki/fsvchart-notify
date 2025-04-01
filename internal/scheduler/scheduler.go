package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	// 假设你的项目中有：
	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/models"
	"fsvchart-notify/internal/service"
)

// StartScheduler: 启动调度器，每5分钟检查一次任务
func StartScheduler() {
	// 立即执行一次，处理服务重启期间可能错过的任务
	log.Println("[scheduler] Starting scheduler...")
	processPushTasks()

	// 创建定时器，每1分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)

	// 启动后台goroutine持续运行
	go func() {
		defer ticker.Stop() // 确保ticker最终会被停止

		for range ticker.C {
			log.Println("[scheduler] Ticker triggered, processing tasks...")
			processPushTasks()
		}
	}()

	log.Println("[scheduler] Scheduler started successfully")
}

// parseDurationString: 把 "5m"/"30m"/"1h"/"1d"/"1M" 等字符串解析为 time.Duration
// 若解析失败，fallback 30m
func parseDurationString(s string) time.Duration {
	// 处理天数格式 (例如 "1d", "5d")
	if strings.HasSuffix(s, "d") {
		dayStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(dayStr)
		if err == nil && days > 0 {
			duration := time.Duration(days) * 24 * time.Hour
			log.Printf("[scheduler] Parsed %s as %v (days)", s, duration)
			return duration
		}
		log.Printf("[scheduler] Invalid day format: %s, fallback=30m", s)
		return 30 * time.Minute
	}

	// 处理月份格式 (例如 "1M", "3M")
	if strings.HasSuffix(s, "M") {
		monthStr := strings.TrimSuffix(s, "M")
		months, err := strconv.Atoi(monthStr)
		if err == nil && months > 0 {
			// 假设一个月为30天
			return time.Duration(months) * 30 * 24 * time.Hour
		}
		log.Printf("[scheduler] Invalid month format: %s, fallback=30m", s)
		return 30 * time.Minute
	}

	// 处理标准时间单位
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("[scheduler] parseDurationString error: %v, fallback=30m", err)
		return 30 * time.Minute
	}
	return d
}

// secondsToHours: 将秒转换为小时
func secondsToHours(seconds int) float64 {
	return float64(seconds) / 3600
}

// processPushTasks: 查询 push_task -> 获取 metrics_source -> FetchMetrics -> 发送至 WebHook
func processPushTasks() {
	db, err := database.SetupDB("./data/app.db")
	if err != nil {
		log.Printf("[scheduler] DB init error: %v", err)
		return
	}

	log.Println("[scheduler] processPushTasks start...")

	// 1. 查询所有启用的任务及其发送时间
	rows, err := db.Query(`
		SELECT DISTINCT 
			p.id, p.source_id, COALESCE(p.query, '') as query, p.time_range, p.step, 
			p.schedule_interval, COALESCE(p.last_run_at, '') as last_run_at,
			COALESCE(ct.chart_type, 'area') as chart_type,
			COALESCE(p.card_title, '') as card_title,
			COALESCE(p.card_template, 'red') as card_template,
			COALESCE(p.metric_label, 'pod') as metric_label,
			COALESCE(p.unit, '') as unit,
			COALESCE(p.button_text, '') as button_text,
			COALESCE(p.button_url, '') as button_url,
			pts.weekday,
			pts.send_time
		FROM push_task p
		LEFT JOIN chart_template ct ON p.chart_template_id = ct.id
		LEFT JOIN push_task_send_time pts ON p.id = pts.task_id
		WHERE p.enabled = 1
	`)
	if err != nil {
		log.Printf("[scheduler] Query push_task error: %v", err)
		return
	}
	defer rows.Close()

	now := time.Now()
	currentWeekday := int(now.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7 // 将周日从0改为7
	}
	currentTime := now.Format("15:04")

	for rows.Next() {
		var task struct {
			ID            int64
			SourceID      int64
			Query         string
			TimeRange     string
			Step          int
			SchedInterval int
			LastRunAt     string
			ChartType     string
			CardTitle     string
			CardTemplate  string
			MetricLabel   string
			Unit          string
			ButtonText    string
			ButtonURL     string
			Weekday       sql.NullInt64
			SendTime      sql.NullString
		}

		err := rows.Scan(
			&task.ID, &task.SourceID, &task.Query, &task.TimeRange, &task.Step,
			&task.SchedInterval, &task.LastRunAt, &task.ChartType,
			&task.CardTitle, &task.CardTemplate, &task.MetricLabel, &task.Unit,
			&task.ButtonText, &task.ButtonURL, &task.Weekday, &task.SendTime,
		)
		if err != nil {
			log.Printf("[scheduler] Scan task error: %v", err)
			continue
		}

		// 检查是否是当前星期几
		if !task.Weekday.Valid || int(task.Weekday.Int64) != currentWeekday {
			continue
		}

		// 检查是否到达发送时间
		if !task.SendTime.Valid || task.SendTime.String != currentTime {
			continue
		}

		// 检查上次运行时间
		if task.LastRunAt != "" {
			lastRun, err := time.Parse("2006-01-02 15:04:05", task.LastRunAt)
			if err != nil {
				log.Printf("[scheduler] Parse last_run_at error: %v", err)
				continue
			}

			// 如果距离上次运行时间不足调度间隔，则跳过
			if now.Sub(lastRun).Seconds() < float64(task.SchedInterval) {
				continue
			}
		}

		// 更新最后运行时间
		_, err = db.Exec("UPDATE push_task SET last_run_at = ? WHERE id = ?",
			now.Format("2006-01-02 15:04:05"), task.ID)
		if err != nil {
			log.Printf("[scheduler] Update last_run_at error: %v", err)
			continue
		}

		// 执行任务
		go runSingleTaskPush(db, task.ID)
	}
}

// insertPushStatus: 记录推送结果(成功/失败)
func insertPushStatus(db *sql.DB, sourceID, webhookID int64, err error) {
	status := "success"
	logTxt := "ok"
	if err != nil {
		status = "failed"
		logTxt = err.Error()
	}
	_, e := db.Exec(`
		INSERT INTO push_status (source_id, webhook_id, last_run_time, status, log)
		VALUES (?, ?, datetime('now'), ?, ?)
	`, sourceID, webhookID, status, logTxt)
	if e != nil {
		log.Printf("[scheduler] insertPushStatus error: %v", e)
	}
}

// runSingleTaskPush 执行单个任务的推送
func runSingleTaskPush(db *sql.DB, taskID int64) {
	// 获取任务详情
	var sourceID int64
	var name, timeRange, cardTitle, cardTemplate, metricLabel, unit, buttonText, buttonURL string
	var step float64
	var enabled int
	var showDataLabel sql.NullInt64 // 使用 sql.NullInt64 来处理 NULL 值

	err := db.QueryRow(`
		SELECT source_id, name, time_range, step, 
			   card_title, card_template, metric_label, unit,
			   button_text, button_url, enabled, COALESCE(show_data_label, 0) as show_data_label
		FROM push_task 
		WHERE id = ?
	`, taskID).Scan(&sourceID, &name, &timeRange, &step,
		&cardTitle, &cardTemplate, &metricLabel, &unit,
		&buttonText, &buttonURL, &enabled, &showDataLabel)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error querying task: %v", err)
		return
	}

	// 获取数据源URL
	var sourceURL string
	err = db.QueryRow("SELECT url FROM metrics_source WHERE id = ?", sourceID).Scan(&sourceURL)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error querying metrics source: %v", err)
		return
	}

	// 查询任务的所有查询
	rows, err := db.Query(`
		SELECT ptp.promql_id, p.query, ptp.chart_template_id, p.name
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
	}

	for rows.Next() {
		var promqlID int64
		var q struct {
			Query           string
			ChartTemplateID int64
			PromQLName      string
		}
		if err := rows.Scan(&promqlID, &q.Query, &q.ChartTemplateID, &q.PromQLName); err != nil {
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
				}
				if err := queryRows.Scan(&q.Query, &q.ChartTemplateID); err != nil {
					log.Printf("[runSingleTaskPush] Error scanning query row: %v", err)
					continue
				}
				q.PromQLName = "" // 设置为空字符串
				queries = append(queries, q)
			}
		}
	}

	// 如果仍然没有查询，记录错误并返回
	if len(queries) == 0 {
		log.Printf("[runSingleTaskPush] No queries found for task ID=%d", taskID)
		return
	}

	// 解析time_range为持续时间
	duration := parseDurationString(timeRange)
	end := time.Now()
	start := end.Add(-duration)

	// 获取所有绑定的webhook
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

	// 修改发送逻辑：先循环查询获取数据，再发送到所有webhook
	// 这样可以避免因飞书API频率限制导致的重复查询和发送
	type QueryResult struct {
		DataPoints []models.DataPoint
		ChartType  string
		ChartTitle string
		QueryName  string
	}

	// 存储每个查询的结果
	var queryResults []QueryResult

	// 用于去重的map
	seenQueries := make(map[string]bool)

	// 先获取所有查询的数据点
	for _, query := range queries {
		// 使用查询名称作为唯一标识
		queryKey := query.PromQLName
		if queryKey == "" {
			queryKey = query.Query
		}

		// 跳过重复的查询
		if seenQueries[queryKey] {
			log.Printf("[runSingleTaskPush] 跳过重复的查询: %s", queryKey)
			continue
		}
		seenQueries[queryKey] = true

		var chartType string
		err = db.QueryRow("SELECT chart_type FROM chart_template WHERE id = ?", query.ChartTemplateID).Scan(&chartType)
		if err != nil {
			log.Printf("[runSingleTaskPush] Error querying chart template (ID=%d): %v", query.ChartTemplateID, err)
			chartType = "area" // 默认为area类型
		}

		// 获取数据点
		dataPoints, err := service.FetchMetrics(sourceURL, query.Query, start, end, time.Duration(step)*time.Second, metricLabel, "")
		if err != nil {
			log.Printf("[runSingleTaskPush] Error fetching metrics for query '%s': %v", query.PromQLName, err)
			// 记录错误但继续处理其他查询
			continue
		}

		// 保存查询结果
		queryResults = append(queryResults, QueryResult{
			DataPoints: dataPoints,
			ChartType:  chartType,
			ChartTitle: cardTitle,
			QueryName:  queryKey,
		})

		log.Printf("[runSingleTaskPush] Successfully fetched metrics for query '%s': %d data points",
			queryKey, len(dataPoints))
	}

	// 检查是否有有效的查询结果
	if len(queryResults) == 0 {
		log.Printf("[runSingleTaskPush] No valid query results for task ID=%d", taskID)
		for _, wh := range webhooks {
			insertPushStatus(db, sourceID, wh.ID, fmt.Errorf("no valid data from any query"))
		}
		return
	}

	// 然后将所有查询结果发送到每个webhook
	for _, wh := range webhooks {
		// 为每个webhook只发送一次，包含所有查询结果
		var allQueryDataPoints []models.QueryDataPoints

		// 使用map进行系列去重
		seenSeries := make(map[string]bool)

		// 收集所有查询的数据点
		for _, result := range queryResults {
			// 使用查询名称作为唯一标识
			seriesKey := result.QueryName
			if seriesKey == "" {
				// 如果没有查询名称，使用ChartTitle
				seriesKey = result.ChartTitle
			}

			// 跳过重复的系列
			if seenSeries[seriesKey] {
				log.Printf("[runSingleTaskPush] 跳过重复的系列: %s", seriesKey)
				continue
			}
			seenSeries[seriesKey] = true

			// 添加到发送队列
			allQueryDataPoints = append(allQueryDataPoints, models.QueryDataPoints{
				DataPoints: result.DataPoints,
				ChartType:  result.ChartType,
				ChartTitle: seriesKey,
			})
		}

		// 记录实际发送的系列数量
		log.Printf("[runSingleTaskPush] 向webhook ID=%d 发送 %d 个唯一系列", wh.ID, len(allQueryDataPoints))

		// 发送飞书消息，一次性发送所有查询结果
		err = service.SendFeishuStandardChart(wh.URL, allQueryDataPoints, cardTitle, cardTemplate,
			unit, buttonText, buttonURL, showDataLabel.Int64 == 1)

		if err != nil {
			log.Printf("[runSingleTaskPush] Error sending to webhook ID=%d: %v", wh.ID, err)
			insertPushStatus(db, sourceID, wh.ID, err)

			// 如果是频率限制错误，增加等待时间
			if strings.Contains(err.Error(), "frequency limited") || strings.Contains(err.Error(), "too many request") {
				waitTime := 3 * time.Second
				log.Printf("[runSingleTaskPush] 检测到频率限制错误，等待 %v 后继续", waitTime)
				time.Sleep(waitTime)
				continue // 遇到频率限制错误时，跳过当前发送，继续下一个
			}
		} else {
			log.Printf("[runSingleTaskPush] Successfully sent %d unique series to webhook ID=%d",
				len(allQueryDataPoints), wh.ID)
			insertPushStatus(db, sourceID, wh.ID, nil)

			// 成功发送后也增加短暂间隔，避免频率限制
			time.Sleep(500 * time.Millisecond)
		}

		// 每个webhook只发送一次，发送完就退出循环
		break
	}

	// 更新任务的last_run_at
	_, err = db.Exec("UPDATE push_task SET last_run_at = datetime('now') WHERE id = ?", taskID)
	if err != nil {
		log.Printf("[runSingleTaskPush] Error updating last_run_at: %v", err)
	}
}
