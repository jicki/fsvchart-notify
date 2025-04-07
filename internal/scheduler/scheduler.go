package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	// 假设你的项目中有：
	"fsvchart-notify/internal/database"
	"fsvchart-notify/internal/models"
	"fsvchart-notify/internal/service"
)

// TaskQueue 任务队列
type TaskQueue struct {
	tasks    chan int64
	running  sync.Map
	interval time.Duration
}

var (
	// 全局任务队列
	taskQueue = &TaskQueue{
		tasks:    make(chan int64, 100),
		interval: 500 * time.Millisecond,
	}
)

// webhookMutexes 用于控制对每个webhook的访问
var webhookMutexes = struct {
	sync.RWMutex
	m map[int64]*sync.Mutex
}{
	m: make(map[int64]*sync.Mutex),
}

// taskMutexes 用于控制对每个任务的访问，防止同一任务并行执行
var taskMutexes = struct {
	sync.RWMutex
	m map[int64]*sync.Mutex
}{
	m: make(map[int64]*sync.Mutex),
}

// getWebhookMutex 获取指定webhook的互斥锁
func getWebhookMutex(webhookID int64) *sync.Mutex {
	webhookMutexes.RLock()
	if mu, exists := webhookMutexes.m[webhookID]; exists {
		webhookMutexes.RUnlock()
		return mu
	}
	webhookMutexes.RUnlock()

	webhookMutexes.Lock()
	defer webhookMutexes.Unlock()

	// 双重检查
	if mu, exists := webhookMutexes.m[webhookID]; exists {
		return mu
	}

	mu := &sync.Mutex{}
	webhookMutexes.m[webhookID] = mu
	return mu
}

// getTaskMutex 获取指定任务的互斥锁
func getTaskMutex(taskID int64) *sync.Mutex {
	taskMutexes.RLock()
	if mu, exists := taskMutexes.m[taskID]; exists {
		taskMutexes.RUnlock()
		return mu
	}
	taskMutexes.RUnlock()

	taskMutexes.Lock()
	defer taskMutexes.Unlock()

	// 双重检查
	if mu, exists := taskMutexes.m[taskID]; exists {
		return mu
	}

	mu := &sync.Mutex{}
	taskMutexes.m[taskID] = mu
	return mu
}

// StartScheduler: 启动调度器
func StartScheduler() {
	log.Println("[scheduler] Starting scheduler...")

	// 启动任务执行器
	go taskQueue.run()

	// 立即执行一次，处理服务重启期间可能错过的任务
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

// run 运行任务队列
func (q *TaskQueue) run() {
	log.Printf("[TaskQueue] 任务队列启动运行")
	taskCount := 0

	for taskID := range q.tasks {
		taskCount++
		log.Printf("[TaskQueue] ====== 开始处理第 %d 个任务 [ID=%d] ======", taskCount, taskID)

		// 检查任务是否已经在运行
		if _, running := q.running.LoadOrStore(taskID, true); running {
			log.Printf("[TaskQueue] 任务 %d 已在运行中，跳过执行", taskID)
			continue
		}

		// 记录任务开始时间
		startTime := time.Now()
		log.Printf("[TaskQueue] 任务 %d 开始执行，时间: %s", taskID, startTime.Format("2006-01-02 15:04:05"))

		// 执行任务
		runSingleTaskPush(database.GetDB(), taskID)

		// 记录任务结束时间和执行时长
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		log.Printf("[TaskQueue] 任务 %d 执行完成，结束时间: %s，耗时: %.2f 秒",
			taskID, endTime.Format("2006-01-02 15:04:05"), duration.Seconds())

		// 任务完成后从运行map中移除
		q.running.Delete(taskID)
		log.Printf("[TaskQueue] 任务 %d 已从运行列表中移除", taskID)

		// 添加间隔，避免频率限制
		log.Printf("[TaskQueue] 等待 %v 后处理下一个任务...", q.interval)
		time.Sleep(q.interval)

		log.Printf("[TaskQueue] ====== 任务 [ID=%d] 处理完成 ======\n", taskID)
	}
}

// addTask 添加任务到队列
func (q *TaskQueue) addTask(taskID int64) {
	// 如果任务已经在运行，则跳过
	if _, running := q.running.Load(taskID); running {
		log.Printf("[TaskQueue] 任务 %d 已在运行中，不再加入队列", taskID)
		return
	}

	// 获取当前队列长度
	queueLen := len(q.tasks)
	log.Printf("[TaskQueue] 当前队列中有 %d 个任务等待执行", queueLen)

	select {
	case q.tasks <- taskID:
		log.Printf("[TaskQueue] 成功将任务 %d 加入队列，当前队列长度: %d", taskID, queueLen+1)
	default:
		log.Printf("[TaskQueue] 队列已满（容量=%d），任务 %d 被丢弃", cap(q.tasks), taskID)
	}
}

// processPushTasks 处理所有已启用的任务
func processPushTasks() {
	db := database.GetDB()
	now := time.Now()
	log.Printf("[scheduler] %s processPushTasks start...", now.Format("2006/01/02 15:04:05"))

	currentWeekday := int(now.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7 // 将周日从0改为7
	}
	currentTime := now.Format("15:04")

	// 查询所有启用的任务及其发送时间
	rows, err := db.Query(`
		SELECT DISTINCT 
			p.id, p.source_id, COALESCE(p.last_run_at, '') as last_run_at,
			p.schedule_interval, pts.weekday, pts.send_time
		FROM push_task p
		LEFT JOIN push_task_send_time pts ON p.id = pts.task_id
		WHERE p.enabled = 1
	`)
	if err != nil {
		log.Printf("[scheduler] Query push_task error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task struct {
			ID            int64
			SourceID      int64
			LastRunAt     string
			SchedInterval int
			Weekday       sql.NullInt64
			SendTime      sql.NullString
		}

		err := rows.Scan(
			&task.ID, &task.SourceID, &task.LastRunAt,
			&task.SchedInterval, &task.Weekday, &task.SendTime,
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

		// 将任务添加到队列
		taskQueue.addTask(task.ID)
	}
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
	// 获取任务互斥锁，确保同一任务不会并行执行
	taskMutex := getTaskMutex(taskID)

	// 尝试获取锁
	if !taskMutex.TryLock() {
		log.Printf("[TaskQueue] 任务 ID=%d 已经在执行中，跳过本次执行", taskID)
		return
	}
	defer taskMutex.Unlock()

	log.Printf("[TaskQueue] ===== 开始执行任务 ID=%d =====", taskID)

	// 获取任务详情
	var sourceID int64
	var name, timeRange, cardTitle, cardTemplate, metricLabel, unit, buttonText, buttonURL, customMetricLabel string
	var step float64
	var enabled int
	var showDataLabel sql.NullInt64

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
		log.Printf("[TaskQueue] 获取任务详情失败: %v", err)
		return
	}

	log.Printf("[TaskQueue] 任务信息: name=%s, timeRange=%s, step=%v", name, timeRange, step)

	// 检查任务是否启用
	if enabled != 1 {
		log.Printf("[TaskQueue] 任务未启用，跳过执行")
		return
	}

	// 获取数据源URL
	var sourceURL string
	err = db.QueryRow("SELECT url FROM metrics_source WHERE id = ?", sourceID).Scan(&sourceURL)
	if err != nil {
		log.Printf("[TaskQueue] 获取数据源失败: %v", err)
		return
	}

	// 使用map进行查询去重
	seenQueries := make(map[string]bool)
	var uniqueQueries []struct {
		Query           string
		ChartTemplateID int64
		PromQLName      string
	}

	// 查询任务的所有查询
	rows, err := db.Query(`
		SELECT ptp.promql_id, p.query, ptp.chart_template_id, p.name
		FROM push_task_promql ptp
		JOIN promql p ON ptp.promql_id = p.id
		WHERE ptp.task_id = ?
	`, taskID)
	if err != nil {
		log.Printf("[TaskQueue] 查询PromQL失败: %v", err)
		return
	}
	defer rows.Close()

	// 处理查询结果
	for rows.Next() {
		var promqlID int64
		var q struct {
			Query           string
			ChartTemplateID int64
			PromQLName      string
		}
		if err := rows.Scan(&promqlID, &q.Query, &q.ChartTemplateID, &q.PromQLName); err != nil {
			log.Printf("[TaskQueue] 扫描PromQL行失败: %v", err)
			continue
		}

		// 使用查询内容作为去重键
		if !seenQueries[q.Query] {
			seenQueries[q.Query] = true
			uniqueQueries = append(uniqueQueries, q)
			log.Printf("[TaskQueue] 添加唯一查询: %s", q.Query)
		} else {
			log.Printf("[TaskQueue] 跳过重复查询: %s", q.Query)
		}
	}

	// 如果没有找到查询，尝试使用旧的单查询格式
	if len(uniqueQueries) == 0 {
		log.Printf("[TaskQueue] 未找到PromQL查询，尝试使用旧格式查询")

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
					log.Printf("[TaskQueue] 扫描查询行失败: %v", err)
					continue
				}

				// 使用查询内容作为去重键
				if !seenQueries[q.Query] {
					seenQueries[q.Query] = true
					uniqueQueries = append(uniqueQueries, q)
					log.Printf("[TaskQueue] 添加唯一旧格式查询: %s", q.Query)
				}
			}
		}

		// 如果仍然没有查询，尝试从push_task表中获取
		if len(uniqueQueries) == 0 {
			var query string
			var chartTemplateID int64
			err := db.QueryRow(`
				SELECT query, chart_template_id
				FROM push_task
				WHERE id = ?
			`, taskID).Scan(&query, &chartTemplateID)

			if err == nil && query != "" && !seenQueries[query] {
				uniqueQueries = append(uniqueQueries, struct {
					Query           string
					ChartTemplateID int64
					PromQLName      string
				}{
					Query:           query,
					ChartTemplateID: chartTemplateID,
					PromQLName:      "",
				})
				log.Printf("[TaskQueue] 添加任务表中的查询: %s", query)
			}
		}
	}

	log.Printf("[TaskQueue] 共找到 %d 个唯一查询", len(uniqueQueries))

	// 如果仍然没有查询，记录错误并返回
	if len(uniqueQueries) == 0 {
		log.Printf("[TaskQueue] 未找到任何有效查询，任务终止")
		return
	}

	// 解析time_range为持续时间
	duration := parseDurationString(timeRange)
	end := time.Now()
	start := end.Add(-duration)
	log.Printf("[TaskQueue] 查询时间范围: %s 至 %s", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))

	// 获取所有绑定的webhook
	webhookRows, err := db.Query(`
		SELECT w.id, w.url 
		FROM feishu_webhook w
		JOIN push_task_webhook ptw ON w.id = ptw.webhook_id
		WHERE ptw.task_id = ?
	`, taskID)
	if err != nil {
		log.Printf("[TaskQueue] 获取webhook失败: %v", err)
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
			log.Printf("[TaskQueue] 扫描webhook行失败: %v", err)
			continue
		}
		webhooks = append(webhooks, wh)
	}

	if len(webhooks) == 0 {
		log.Printf("[TaskQueue] 未找到webhook配置，任务终止")
		return
	}

	log.Printf("[TaskQueue] 找到 %d 个webhook配置", len(webhooks))

	// 为每个查询获取数据点
	var allDataPoints []models.QueryDataPoints
	seenSeries := make(map[string]bool) // 用于系列去重

	for i, query := range uniqueQueries {
		// 获取图表类型
		var chartType string
		err = db.QueryRow("SELECT chart_type FROM chart_template WHERE id = ?", query.ChartTemplateID).Scan(&chartType)
		if err != nil {
			log.Printf("[TaskQueue] 获取图表类型失败 (ID=%d): %v，使用默认类型 'area'", query.ChartTemplateID, err)
			chartType = "area"
		}

		chartType = service.GetSupportedChartType(chartType)
		log.Printf("[TaskQueue] 查询 %d: 使用图表类型 %s", i+1, chartType)

		// 获取数据点
		log.Printf("[TaskQueue] 开始获取查询 %d 的指标数据: %s", i+1, query.Query)
		dataPoints, err := service.FetchMetrics(sourceURL, query.Query, start, end, time.Duration(step)*time.Second, metricLabel, customMetricLabel)
		if err != nil {
			log.Printf("[TaskQueue] 获取指标数据失败: %v", err)
			continue
		}

		// 生成图表标题
		chartTitle := query.PromQLName
		if chartTitle == "" {
			chartTitle = fmt.Sprintf("查询 %d", i+1)
		}

		// 检查系列是否重复
		if !seenSeries[chartTitle] {
			seenSeries[chartTitle] = true
			allDataPoints = append(allDataPoints, models.QueryDataPoints{
				DataPoints: dataPoints,
				ChartType:  chartType,
				ChartTitle: chartTitle,
			})
			log.Printf("[TaskQueue] 添加新的数据系列: %s (包含 %d 个数据点)", chartTitle, len(dataPoints))
		} else {
			log.Printf("[TaskQueue] 跳过重复的数据系列: %s", chartTitle)
		}
	}

	if len(allDataPoints) == 0 {
		log.Printf("[TaskQueue] 未获取到任何数据点，任务终止")
		return
	}

	log.Printf("[TaskQueue] 共收集到 %d 个唯一数据系列", len(allDataPoints))

	// 发送到所有绑定的webhook
	if len(webhooks) == 0 {
		log.Printf("[TaskQueue] 任务没有配置webhook，跳过发送")
	} else {
		log.Printf("[TaskQueue] 准备发送到 %d 个webhook", len(webhooks))

		// 记录已发送过的webhook，避免重复发送
		sentWebhooks := make(map[string]bool)

		for _, webhook := range webhooks {
			// 跳过重复的webhook URL
			if sentWebhooks[webhook.URL] {
				log.Printf("[TaskQueue] 跳过重复的webhook URL: %s", webhook.URL)
				continue
			}

			log.Printf("[TaskQueue] 准备发送到webhook (ID=%d)", webhook.ID)

			// 添加webhook互斥锁，避免同时向同一webhook发送多个消息
			webhookMutex := getWebhookMutex(webhook.ID)
			webhookMutex.Lock()

			err = service.SendFeishuStandardChart(webhook.URL, allDataPoints, cardTitle, cardTemplate,
				unit, buttonText, buttonURL, showDataLabel.Int64 == 1)

			if err != nil {
				log.Printf("[TaskQueue] 发送失败: %v", err)
				insertPushStatus(db, sourceID, webhook.ID, err)

				if strings.Contains(err.Error(), "frequency limited") || strings.Contains(err.Error(), "too many request") {
					waitTime := 3 * time.Second
					log.Printf("[TaskQueue] 检测到频率限制，等待 %v", waitTime)
					time.Sleep(waitTime)
				}
			} else {
				log.Printf("[TaskQueue] 发送成功: %d 个数据系列", len(allDataPoints))
				insertPushStatus(db, sourceID, webhook.ID, nil)

				// 标记此webhook URL已发送
				sentWebhooks[webhook.URL] = true
			}

			webhookMutex.Unlock()
		}
	}

	log.Printf("[TaskQueue] ===== 任务 ID=%d 执行完成 =====\n", taskID)
}

// RunSingleTaskPush 执行单个任务的推送（公共导出版本）
// 提供给server包和其他外部包调用，确保任务执行有互斥控制
func RunSingleTaskPush(db *sql.DB, taskID int64) {
	runSingleTaskPush(db, taskID)
}
