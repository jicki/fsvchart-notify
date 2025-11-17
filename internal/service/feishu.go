package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsvchart-notify/internal/models"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// =====================
// 1. 基础结构定义
// =====================

// DataPoint 代表一个时序数据点，用于示例图表数据
type DataPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// FeishuCard 代表飞书的 Interactive 卡片消息顶层结构
type FeishuCard struct {
	MsgType string         `json:"msg_type"` // 固定 "interactive"
	Card    FeishuCardBody `json:"card"`
}

// FeishuCardBody 包含卡片的主体内容：配置、头部、元素
type FeishuCardBody struct {
	Config   *FeishuCardConfig   `json:"config,omitempty"`
	Header   *FeishuCardHeader   `json:"header,omitempty"`
	Elements []FeishuCardElement `json:"elements,omitempty"`
}

// FeishuCardConfig 配置卡片是否宽屏、是否允许转发等
type FeishuCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

// FeishuCardHeader 卡片头部：标题与主题颜色
type FeishuCardHeader struct {
	Title    *FeishuCardHeaderTitle `json:"title,omitempty"`
	Template string                 `json:"template,omitempty"` // "blue", "red", "green", etc
}

// FeishuCardHeaderTitle 头部标题
type FeishuCardHeaderTitle struct {
	Content string `json:"content,omitempty"`
	Tag     string `json:"tag,omitempty"` // "plain_text", "lark_md"
}

// FeishuCardElement 代表 elements 数组中的单个组件
type FeishuCardElement struct {
	// 基础
	Tag string `json:"tag"` // "markdown", "hr", "chart", "action", "note" 等

	// markdown 专用
	Content string `json:"content,omitempty"`

	// chart 专用
	ChartSpec *FeishuChartSpec `json:"chart_spec,omitempty"`

	// action 专用
	Actions []FeishuAction `json:"actions,omitempty"`

	// note 专用
	Elements []FeishuNoteElement `json:"elements,omitempty"`
}

// ChartTemplate 代表图表模板
type ChartTemplate struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ChartType string `json:"chart_type"`
}

// FeishuChartSeries 新增 Series 配置结构
type FeishuChartSeries struct {
	Type          string                 `json:"type"`
	Stack         bool                   `json:"stack,omitempty"`
	DataIndex     int                    `json:"dataIndex"`
	Label         map[string]interface{} `json:"label,omitempty"`
	SeriesField   string                 `json:"seriesField"`
	XField        string                 `json:"xField"`
	YField        string                 `json:"yField"`
	ShowDataLabel bool                   `json:"show_data_label"`
}

// FeishuChartAxis 新增坐标轴配置结构
type FeishuChartAxis struct {
	Orient    string                 `json:"orient"`
	TickCount int                    `json:"tickCount,omitempty"`
	Label     map[string]interface{} `json:"label,omitempty"`
}

// FeishuChartTooltip 新增图表工具提示结构
type FeishuChartTooltip struct {
	Mark      map[string]interface{} `json:"mark,omitempty"`
	Dimension map[string]interface{} `json:"dimension,omitempty"`
}

// FeishuChartSpec 代表 "chart" 类型元素下的 "chart_spec"
type FeishuChartSpec struct {
	Type    string                   `json:"type"` // 使用 "common"
	Title   map[string]interface{}   `json:"title,omitempty"`
	Data    []map[string]interface{} `json:"data"`    // 改为数组支持多系列
	Series  []FeishuChartSeries      `json:"series"`  // 新增 series 配置
	Axes    []FeishuChartAxis        `json:"axes"`    // 新增坐标轴配置
	Legends map[string]interface{}   `json:"legends"` // 新增图例配置
	Tooltip *FeishuChartTooltip      `json:"tooltip,omitempty"`
	Layout  map[string]interface{}   `json:"layout,omitempty"`
}

// FeishuAction 代表 "action" 元素中的按钮或选择等交互组件
type FeishuAction struct {
	Tag   string            `json:"tag"`             // 一般 "button"
	Text  *FeishuActionText `json:"text,omitempty"`
	Type  string            `json:"type,omitempty"`  // "primary", "default", ...
	URL   string            `json:"url,omitempty"`   // 跳转链接
	Value map[string]string `json:"value,omitempty"` // 自定义键值对（已废弃）
}

// FeishuActionText 代表按钮上的文字
type FeishuActionText struct {
	Content string `json:"content,omitempty"`
	Tag     string `json:"tag,omitempty"` // "plain_text", "lark_md"
}

// FeishuNoteElement 代表 "note" 元素中的子元素
type FeishuNoteElement struct {
	Tag     string `json:"tag"` // "lark_md", "plain_text"
	Content string `json:"content,omitempty"`
}

// =====================
// 2. 发送卡片消息的函数
// =====================

// ========== 1. 定义带超时的 http.Client ==========

var httpClient = &http.Client{
	// 增加超时时间到30秒，避免网络波动导致的超时
	Timeout: 30 * time.Second,
}

// 最大重试次数
const maxRetries = 3

// SendFeishuCardMessage 用于发送任意自定义的 FeishuCard
func SendFeishuCardMessage(webhookURL string, card *FeishuCard) error {
	payload, err := json.Marshal(card)
	if err != nil {
		log.Printf("[SendFeishuCardMessage] JSON marshal error: %v", err)
		return fmt.Errorf("json marshal error: %w", err)
	}

	// 记录发送的URL和数据大小
	log.Printf("[SendFeishuCardMessage] Sending to webhook URL: %s, payload size: %d bytes", webhookURL, len(payload))

	// 添加重试逻辑
	var lastErr error
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("[SendFeishuCardMessage] Retry attempt %d/%d after error: %v", retry, maxRetries, lastErr)
			// 重试前等待一段时间，避免立即重试
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// 构造 POST 请求
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("[SendFeishuCardMessage] Create request error: %v", err)
			lastErr = fmt.Errorf("create request error: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// 用我们带 Timeout 的 httpClient 执行
		log.Printf("[SendFeishuCardMessage] Executing HTTP request (attempt %d/%d)...", retry+1, maxRetries)
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("[SendFeishuCardMessage] HTTP post error: %v", err)
			lastErr = fmt.Errorf("http post error: %w", err)
			continue
		}

		// 确保响应体被关闭
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		log.Printf("[SendFeishuCardMessage] Response status: %d, body: %s", resp.StatusCode, string(bodyBytes))

		// 解析飞书API的响应
		var feishuResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(bodyBytes, &feishuResp); err == nil {
			if feishuResp.Code != 0 {
				log.Printf("[SendFeishuCardMessage] Feishu API returned error code: %d, message: %s", feishuResp.Code, feishuResp.Msg)
				lastErr = fmt.Errorf("feishu API error: code=%d, msg=%s", feishuResp.Code, feishuResp.Msg)
				// 如果是飞书API错误，继续重试
				continue
			}
			log.Printf("[SendFeishuCardMessage] Feishu API success: code=%d", feishuResp.Code)
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("feishu webhook returned status=%d, body=%s", resp.StatusCode, string(bodyBytes))
			continue
		}

		// 成功发送，返回nil
		log.Printf("[SendFeishuCardMessage] Successfully sent message to webhook")
		return nil
	}

	// 所有重试都失败了，返回最后一个错误
	log.Printf("[SendFeishuCardMessage] All %d retry attempts failed, last error: %v", maxRetries, lastErr)
	return lastErr
}

// =====================
// 3. 构建并发送示例卡片
// =====================

var sendRecords []models.SendRecord
var recordMutex sync.Mutex

// AddSendRecord 添加发送记录
func AddSendRecord(record models.SendRecord) {
	recordMutex.Lock()
	defer recordMutex.Unlock()

	// 保持最近1000条记录
	if len(sendRecords) >= 1000 {
		sendRecords = sendRecords[1:]
	}
	sendRecords = append(sendRecords, record)
}

// GetSendRecords 获取发送记录
func GetSendRecords() []models.SendRecord {
	recordMutex.Lock()
	defer recordMutex.Unlock()

	return sendRecords
}

// GetSupportedChartType 获取飞书支持的图表类型
func GetSupportedChartType(chartType string) string {
	// 飞书支持的图表类型
	supportedTypes := map[string]bool{
		"line":    true,
		"bar":     true,
		"pie":     true,
		"area":    true,
		"scatter": true,
		"bubble":  true,
	}

	// 如果是支持的类型，直接返回
	if supportedTypes[chartType] {
		return chartType
	}

	// 不支持的类型映射到支持的类型
	typeMapping := map[string]string{
		"bar3d":  "bar",
		"line3d": "line",
		"radar":  "line",
		"funnel": "bar",
		"gauge":  "pie",
	}

	if mappedType, ok := typeMapping[chartType]; ok {
		log.Printf("[GetSupportedChartType] Unsupported chart type '%s' mapped to '%s'", chartType, mappedType)
		return mappedType
	}

	// 默认返回折线图
	log.Printf("[GetSupportedChartType] Unknown chart type '%s', using default 'line'", chartType)
	return "line"
}

// SendFeishuStandardChart 严格按照飞书官方文档构建图表消息
func SendFeishuStandardChart(webhookURL string, queryDataPoints []models.QueryDataPoints, cardTitle, cardTemplate, unit, buttonText, buttonURL string, showDataLabel bool) error {
	// 添加发送前的日志
	log.Printf("[SendFeishuStandardChart] 准备发送消息到 webhook: %s", webhookURL)
	log.Printf("[SendFeishuStandardChart] 标题: %s, 系列数量: %d", cardTitle, len(queryDataPoints))

	// 检查参数
	if len(queryDataPoints) == 0 {
		return fmt.Errorf("no data points provided")
	}

	// 对每个查询的数据点进行预处理
	for i := range queryDataPoints {
		if len(queryDataPoints[i].DataPoints) == 0 {
			continue
		}

		// 按类型分组数据点
		typeGroups := make(map[string][]models.DataPoint)
		for _, dp := range queryDataPoints[i].DataPoints {
			typeGroups[dp.Type] = append(typeGroups[dp.Type], dp)
		}

		// 找出整体的时间范围
		var minTime, maxTime int64
		first := true
		for _, points := range typeGroups {
			for _, p := range points {
				if first || p.UnixTime < minTime {
					minTime = p.UnixTime
				}
				if first || p.UnixTime > maxTime {
					maxTime = p.UnixTime
				}
				first = false
			}
		}

		// 确定时间间隔
		interval := int64(8 * 3600) // 默认8小时
		for _, points := range typeGroups {
			if len(points) > 1 {
				// 对每个类型的点按时间排序
				sort.Slice(points, func(i, j int) bool {
					return points[i].UnixTime < points[j].UnixTime
				})
				// 计算最小时间间隔
				for i := 1; i < len(points); i++ {
					diff := points[i].UnixTime - points[i-1].UnixTime
					if diff > 0 && diff < interval {
						interval = diff
					}
				}
			}
		}

		// 处理每个类型的数据点
		var processedPoints []models.DataPoint
		for seriesType, points := range typeGroups {
			// 创建时间点映射
			timeMap := make(map[int64]models.DataPoint)
			for _, p := range points {
				timeMap[p.UnixTime] = p
			}

			// 补充缺失的时间点
			for t := minTime; t <= maxTime; t += interval {
				if dp, exists := timeMap[t]; exists {
					processedPoints = append(processedPoints, dp)
				} else {
					// 补充缺失的点，值设为0
					processedPoints = append(processedPoints, models.DataPoint{
						Time:     time.Unix(t, 0).Format("15:04"),
						UnixTime: t,
						Value:    0,
						Type:     seriesType,
					})
				}
			}
		}

		// 最终按时间和类型排序
		sort.Slice(processedPoints, func(i, j int) bool {
			if processedPoints[i].UnixTime == processedPoints[j].UnixTime {
				return processedPoints[i].Type < processedPoints[j].Type
			}
			return processedPoints[i].UnixTime < processedPoints[j].UnixTime
		})

		// 更新处理后的数据点
		queryDataPoints[i].DataPoints = processedPoints
	}

	// 对数据进行去重
	uniqueDataPoints := make([]models.QueryDataPoints, 0)
	seenSeries := make(map[string]bool)

	for _, qdp := range queryDataPoints {
		// 使用 ChartTitle 作为唯一标识
		if !seenSeries[qdp.ChartTitle] {
			seenSeries[qdp.ChartTitle] = true
			uniqueDataPoints = append(uniqueDataPoints, qdp)
		} else {
			log.Printf("[SendFeishuStandardChart] 跳过重复的系列: %s", qdp.ChartTitle)
		}
	}

	log.Printf("[SendFeishuStandardChart] 去重后的系列数量: %d", len(uniqueDataPoints))

	// 使用去重后的数据点继续处理
	queryDataPoints = uniqueDataPoints

	if cardTitle == "" {
		cardTitle = "数据推送"
	}
	if cardTemplate == "" {
		cardTemplate = "blue"
	}
	// 设置默认按钮文本和URL
	if buttonText == "" {
		buttonText = "节点池资源总览"
	}
	if buttonURL == "" {
		buttonURL = "https://grafana.deeproute.cn/d/aede74qtud2iod/e88a82-e782b9-e6b1a0-e8b584-e6ba90-e7bb9f-e8aea1?orgId=1"
	}

	// 改进多天数据检测逻辑
	isMultiDayData := true // 默认添加日期前缀

	// 通过分析时间点来检测多天数据
	if len(queryDataPoints) > 0 && len(queryDataPoints[0].DataPoints) > 0 {
		// 方法1: 检查是否有相同的时间点（例如多个"00:00"）
		timeCount := make(map[string]int)
		for _, dp := range queryDataPoints[0].DataPoints {
			timeCount[dp.Time]++
		}

		// 如果有任何时间点出现多次，说明是跨天数据
		hasDuplicateTimes := false
		for time, count := range timeCount {
			if count > 1 {
				hasDuplicateTimes = true
				log.Printf("[SendFeishuStandardChart] 检测到时间点 '%s' 重复出现 %d 次，确认为多天数据",
					time, count)
				break
			}
		}

		// 方法2: 检查数据点数量与时间格式的关系
		// 如果时间格式为"00:00", "08:00", "16:00"等，且数据点超过3个，
		// 说明可能是多天数据（每天3个点）
		timeFormat := ""
		if len(queryDataPoints[0].DataPoints) > 0 {
			timeStr := queryDataPoints[0].DataPoints[0].Time
			if len(timeStr) == 5 && timeStr[2] == ':' {
				timeFormat = "HH:MM"
			}
		}

		// 检查是否可能是8小时步长的多天数据
		possibleMultiDay := false
		if timeFormat == "HH:MM" {
			uniqueTimes := len(timeCount)
			if uniqueTimes <= 3 && len(queryDataPoints[0].DataPoints) > uniqueTimes {
				possibleMultiDay = true
				log.Printf("[SendFeishuStandardChart] 检测到固定时间格式且点数(%d)>唯一时间数(%d)，判断为多天数据",
					len(queryDataPoints[0].DataPoints), uniqueTimes)
			}
		}

		// 如果检测到了重复时间点或符合多天特征，则强制设置为多天数据
		if hasDuplicateTimes || possibleMultiDay {
			isMultiDayData = true
			log.Printf("[SendFeishuStandardChart] 基于时间点重复或数据特征，强制设置为多天数据")
		} else {
			// 只有在没有其他明显多天特征的情况下，才参考时间跨度
			// 计算时间范围长度作为额外参考
			minTime := int64(^uint64(0) >> 1)
			maxTime := int64(0)

			for _, dp := range queryDataPoints[0].DataPoints {
				if dp.UnixTime > maxTime {
					maxTime = dp.UnixTime
				}
				if dp.UnixTime < minTime {
					minTime = dp.UnixTime
				}
			}

			hoursDiff := (maxTime - minTime) / 3600
			log.Printf("[SendFeishuStandardChart] 数据时间跨度: %d小时", hoursDiff)

			// 只有在时间跨度小于20小时（而不是之前的24小时）且无重复时间点时，才不添加日期
			if hoursDiff < 20 && !hasDuplicateTimes && !possibleMultiDay {
				isMultiDayData = false
				log.Printf("[SendFeishuStandardChart] 时间跨度小于20小时且无多天特征，不添加日期前缀")
			} else {
				log.Printf("[SendFeishuStandardChart] 时间跨度>=20小时，添加日期前缀")
			}
		}
	}

	// 创建一个卡片
	cardData := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
				"enable_forward":   true,
			},
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": cardTitle,
				},
				"template": cardTemplate,
			},
			"elements": []interface{}{
				// 添加标题
				map[string]interface{}{
					"tag":     "markdown",
					"content": "**图表数据**\n",
				},
			},
		},
	}

	// 获取所有元素的引用
	elements := cardData["card"].(map[string]interface{})["elements"].([]interface{})

	// 检查是否所有查询都没有数据
	allEmpty := true
	for _, queryData := range queryDataPoints {
		if len(queryData.DataPoints) > 0 {
			allEmpty = false
			break
		}
	}

	// 如果所有查询都没有数据，添加一个全局无数据提示
	if allEmpty && len(queryDataPoints) > 0 {
		log.Printf("[SendFeishuStandardChart] 所有查询均无数据，添加全局无数据提示")

		// 记录更详细的诊断信息
		log.Printf("[SendFeishuStandardChart] 诊断信息:")
		for i, queryData := range queryDataPoints {
			log.Printf("[SendFeishuStandardChart]   - 查询 %d: 标题='%s', 类型='%s', 数据点数=0",
				i+1, queryData.ChartTitle, queryData.ChartType)
		}

		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": "## 📈 数据查询结果为空\n\n**当前所有查询均未返回数据点**\n\n可能的原因：\n- 所选时间范围内没有数据收集\n- 数据源暂时不可用或连接中断\n- 查询参数配置需要调整\n\n您可以尝试以下操作：\n- 调整查询的时间范围\n- 稍后重试查询\n- 检查数据源状态和查询参数",
		})

		// 添加提示的时间和查询信息
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": fmt.Sprintf("*查询时间: %s*", timestamp),
		})

		// 添加一个分隔线
		elements = append(elements, map[string]interface{}{
			"tag": "hr",
		})
	}

	// 为每个查询添加图表
	for i, queryData := range queryDataPoints {
		if len(queryData.DataPoints) == 0 {
			// 记录该查询无数据的详细信息
			log.Printf("[SendFeishuStandardChart] 查询 '%s' 无数据，添加无数据提示", queryData.ChartTitle)

			// 添加查询标题和无数据提示
			elements = append(elements, map[string]interface{}{
				"tag":     "markdown",
				"content": fmt.Sprintf("**%s**\n", queryData.ChartTitle),
			})

			// 添加无数据提示信息
			elements = append(elements, map[string]interface{}{
				"tag":     "markdown",
				"content": "📊 *暂无数据* - 当前查询时间范围内未获取到数据点\n\n可能的原因：\n- 所选时间范围内没有数据\n- 数据源暂时不可用\n- 查询参数配置问题\n\n请稍后重试或调整查询参数。",
			})

			// 显示查询时间
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			elements = append(elements, map[string]interface{}{
				"tag":     "markdown",
				"content": fmt.Sprintf("*查询时间: %s*", timestamp),
			})

			// 如果不是最后一个查询，添加分隔线
			if i < len(queryDataPoints)-1 {
				elements = append(elements, map[string]interface{}{
					"tag": "hr",
				})
			}

			// 继续处理下一个查询
			continue
		}

		// 添加查询标题 - 使用飞书支持的格式
		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": fmt.Sprintf("**%s**\n", queryData.ChartTitle),
		})

		// 组织数据点
		seriesData := make(map[string][]models.DataPoint)

		// 统计未处理前的原始数据点数量
		rawDataPointCount := 0
		for range queryData.DataPoints {
			rawDataPointCount++
		}
		log.Printf("[SendFeishuStandardChart] 系列共有 %d 个原始数据点", rawDataPointCount)

		// 收集所有数据点的日期，检查是否真的有多天数据
		datesFound := make(map[string]bool)
		timesByDate := make(map[string]map[string]bool)

		// 记录所有Unix时间戳的日期映射，用于验证
		unixTimeDateMap := make(map[int64]string)

		for _, dp := range queryData.DataPoints {
			dateStr := time.Unix(dp.UnixTime, 0).Format("2006-01-02")
			datesFound[dateStr] = true
			unixTimeDateMap[dp.UnixTime] = dateStr

			// 按日期记录时间点
			if _, exists := timesByDate[dateStr]; !exists {
				timesByDate[dateStr] = make(map[string]bool)
			}
			timesByDate[dateStr][dp.Time] = true
		}

		uniqueDatesCount := len(datesFound)
		log.Printf("[SendFeishuStandardChart] 检测到 %d 个不同的日期: %v",
			uniqueDatesCount, datesFound)

		// 如果日期数量异常少（例如6天查询只有2天数据），检查是否应该有更多日期
		if durationDays := len(queryData.DataPoints) / 10; durationDays > uniqueDatesCount {
			log.Printf("[SendFeishuStandardChart] 警告: 数据点数量(%d)表明应有~%d天数据，但只找到%d天",
				len(queryData.DataPoints), durationDays, uniqueDatesCount)

			// 输出所有数据点的日期分布
			dateCounts := make(map[string]int)
			for _, dp := range queryData.DataPoints {
				date := time.Unix(dp.UnixTime, 0).Format("2006-01-02")
				dateCounts[date]++
			}

			log.Printf("[SendFeishuStandardChart] 数据点日期分布详情:")
			for date, count := range dateCounts {
				log.Printf(" - 日期 %s: %d 个数据点", date, count)
			}

			// 检查是否只有较早的日期
			now := time.Now()
			today := now.Format("2006-01-02")
			yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

			if datesFound[today] || datesFound[yesterday] {
				log.Printf("[SendFeishuStandardChart] 数据包含最近日期(今天或昨天)")
			} else {
				log.Printf("[SendFeishuStandardChart] 警告: 数据不包含最近日期(今天:%s 昨天:%s)",
					today, yesterday)

				// 记录最近的日期，以便诊断
				var latestDate string
				latestTime := time.Time{}
				for dateStr := range datesFound {
					d, _ := time.Parse("2006-01-02", dateStr)
					if d.After(latestTime) {
						latestTime = d
						latestDate = dateStr
					}
				}

				if latestDate != "" {
					daysSinceLatest := int(now.Sub(latestTime).Hours() / 24)
					log.Printf("[SendFeishuStandardChart] 最近的日期是 %s (距今 %d 天)",
						latestDate, daysSinceLatest)
				}
			}
		}

		// 如果日期数量异常少（5天查询应该至少有3-5天数据），输出详细日志
		if len(queryData.DataPoints) > uniqueDatesCount*3 {
			log.Printf("[SendFeishuStandardChart] 警告: 数据点数量(%d)远大于日期数量(%d)的3倍，可能存在日期检测问题",
				len(queryData.DataPoints), uniqueDatesCount)

			// 输出所有数据点的日期分布
			dateCounts := make(map[string]int)
			for _, dp := range queryData.DataPoints {
				date := time.Unix(dp.UnixTime, 0).Format("2006-01-02")
				dateCounts[date]++
			}

			log.Printf("[SendFeishuStandardChart] 数据点日期分布详情:")
			for date, count := range dateCounts {
				log.Printf(" - 日期 %s: %d 个数据点", date, count)
			}
		}

		// 判断是否保持多天设置
		// 注意：以下逻辑已调整为优先考虑特征检测而不仅是日期数量
		if uniqueDatesCount <= 1 && !isMultiDayData {
			log.Printf("[SendFeishuStandardChart] 实际只有一天的日期且无多天特征，不添加日期前缀")
			isMultiDayData = false
		} else if isMultiDayData {
			log.Printf("[SendFeishuStandardChart] 检测到多天数据特征，强制启用日期前缀")
		} else {
			log.Printf("[SendFeishuStandardChart] 检测到 %d 天的数据，添加日期前缀", uniqueDatesCount)
			isMultiDayData = true
		}

		// 记录处理前的总数据点数量，用于后续验证
		totalOriginalPoints := len(queryData.DataPoints)
		log.Printf("[SendFeishuStandardChart] 处理前总数据点数量: %d", totalOriginalPoints)

		// 记录每种类型的时间点，用于检查可能的重复
		timePointsByType := make(map[string]map[string]bool)

		for _, dp := range queryData.DataPoints {
			if _, exists := seriesData[dp.Type]; !exists {
				seriesData[dp.Type] = make([]models.DataPoint, 0)
				timePointsByType[dp.Type] = make(map[string]bool)
			}

			// 对于多天数据，将日期添加到时间前缀
			if isMultiDayData {
				// 创建一个带日期的副本
				dpCopy := dp

				// 提取实际日期（从时间戳中获取）
				realDate := time.Unix(dp.UnixTime, 0).Format("01/02")

				// 检查是否所有时间点都有相同的时间戳日期（这是一种异常情况）
				// 主检测条件：如果只有一个唯一日期但检测到了多天数据特征
				hasTimeRepetition := false
				// 检查这个类型的时间点是否有重复
				timeCounts := make(map[string]int)
				for _, p := range queryData.DataPoints {
					if p.Type == dp.Type {
						timeCounts[p.Time]++
					}
				}
				for _, count := range timeCounts {
					if count > 1 {
						hasTimeRepetition = true
						break
					}
				}

				// 检查是否有时间点重复但时间戳日期相同的情况
				sameTimestampDates := uniqueDatesCount <= 1 && (hasTimeRepetition || len(queryData.DataPoints) > len(timeCounts))

				// log.Printf("[SendFeishuStandardChart] 处理数据点: Type=%s, Time=%s, UnixTime=%d, 实际日期=%s",
				// 	dp.Type, dp.Time, dp.UnixTime, realDate)

				// 生成不同的虚拟日期以区分数据点
				// 这种情况是数据有同一天的时间戳但逻辑上是多天的数据
				if sameTimestampDates {
					log.Printf("[SendFeishuStandardChart] 检测到时间戳日期相同但时间点重复的情况，使用改进的虚拟日期生成")

					// 收集所有有效的时间戳，按时间顺序排序
					var allTimestamps []int64
					seenTimestamps := make(map[int64]bool)

					for _, p := range queryData.DataPoints {
						if p.Type == dp.Type && !seenTimestamps[p.UnixTime] {
							allTimestamps = append(allTimestamps, p.UnixTime)
							seenTimestamps[p.UnixTime] = true
						}
					}

					// 确保时间戳按顺序排列
					sort.Slice(allTimestamps, func(i, j int) bool {
						return allTimestamps[i] < allTimestamps[j]
					})

					// 记录排序后的时间戳顺序
					if len(allTimestamps) > 0 {
						log.Printf("[SendFeishuStandardChart] 时间戳排序: 共 %d 个时间戳", len(allTimestamps))
						showCount := 5
						if len(allTimestamps) < showCount {
							showCount = len(allTimestamps)
						}
						for i := 0; i < showCount; i++ {
							ts := allTimestamps[i]
							log.Printf(" - 第%d个时间戳: %d (%s)",
								i+1, ts, time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
						}
					}

					// 使用当前序列作为基准，计算偏移日期
					// 收集这个数据点的系列中所有相同时间格式的点
					seriesPoints := make([]models.DataPoint, 0)
					for _, p := range queryData.DataPoints {
						if p.Type == dp.Type && p.Time == dp.Time {
							seriesPoints = append(seriesPoints, p)
						}
					}

					// 对所有具有相同时间格式的点进行排序
					sort.Slice(seriesPoints, func(i, j int) bool {
						return seriesPoints[i].UnixTime < seriesPoints[j].UnixTime
					})

					// 找到当前点在序列中的位置
					pointIndex := -1
					for i, p := range seriesPoints {
						if p.UnixTime == dp.UnixTime {
							pointIndex = i
							break
						}
					}

					// 根据点的位置，生成虚拟日期
					if pointIndex >= 0 {
						// 改进虚拟日期生成逻辑，确保生成真实的日期差异

						// 1. 如果allTimestamps有足够多的时间戳，直接使用不同天的时间戳
						if len(allTimestamps) > pointIndex {
							realTS := allTimestamps[pointIndex]
							// 使用实际时间戳对应的日期，而不是加偏移
							realDate := time.Unix(realTS, 0).Format("01/02")

							log.Printf("[SendFeishuStandardChart] 使用实际时间戳日期: 位置=%d, 时间戳=%d, 实际日期=%s",
								pointIndex, realTS, realDate)

							// 如果原始格式是时间，添加日期前缀
							if len(dp.Time) == 5 && dp.Time[2] == ':' {
								dpCopy.Time = realDate + " " + dp.Time
								log.Printf("[SendFeishuStandardChart] 添加真实日期前缀: %s -> %s", dp.Time, dpCopy.Time)
							} else if !strings.Contains(dp.Time, realDate) {
								// 对于其他格式，也添加日期前缀
								dpCopy.Time = realDate + " " + dp.Time
							}
						} else {
							// 2. 如果没有足够多的时间戳，才使用原来的偏移方法（不应该走到这个逻辑）
							log.Printf("[SendFeishuStandardChart] 警告: 使用回退的虚拟日期生成方法")

							baseTime := time.Unix(dp.UnixTime, 0)
							virtualDate := baseTime.AddDate(0, 0, pointIndex)
							date := virtualDate.Format("01/02")

							log.Printf("[SendFeishuStandardChart] 为重复时间点生成虚拟日期: 位置=%d, 原始日期=%s, 虚拟日期=%s",
								pointIndex, realDate, date)

							// 如果原始格式是时间，添加日期前缀
							if len(dp.Time) == 5 && dp.Time[2] == ':' {
								dpCopy.Time = date + " " + dp.Time
								log.Printf("[SendFeishuStandardChart] 添加虚拟日期前缀: %s -> %s", dp.Time, dpCopy.Time)
							} else if !strings.Contains(dp.Time, date) {
								// 对于其他格式，也添加日期前缀
								dpCopy.Time = date + " " + dp.Time
							}
						}
					} else {
						// 无法确定点的位置，使用默认日期前缀
						if len(dp.Time) == 5 && dp.Time[2] == ':' {
							dpCopy.Time = realDate + " " + dp.Time
							log.Printf("[SendFeishuStandardChart] 添加默认日期前缀: %s -> %s", dp.Time, dpCopy.Time)
						}
					}
				} else {
					// 正常多天数据处理
					// 如果原始格式是时间，添加日期前缀
					if len(dp.Time) == 5 && dp.Time[2] == ':' {
						dpCopy.Time = realDate + " " + dp.Time
						log.Printf("[SendFeishuStandardChart] 添加日期前缀: %s -> %s", dp.Time, dpCopy.Time)
					} else {
						// 如果时间已经有格式，检查是否已包含日期
						if !strings.Contains(dp.Time, realDate) {
							// 没有包含正确的日期，尝试添加
							log.Printf("[SendFeishuStandardChart] 时间格式已有, 但不包含正确日期: %s, 添加前缀: %s", dp.Time, realDate)
							dpCopy.Time = realDate + " " + dp.Time
						}
					}
				}

				// 检查该类型是否已经有相同的时间点
				if timePointsByType[dp.Type][dpCopy.Time] {
					log.Printf("[SendFeishuStandardChart] 警告: 系列 '%s' 中发现重复的时间点 '%s' (Unix: %d)，生成唯一标识",
						dp.Type, dpCopy.Time, dp.UnixTime)
					// 为避免重复，添加Unix时间戳后缀
					dpCopy.Time = fmt.Sprintf("%s.%d", dpCopy.Time, dp.UnixTime)
				}

				// 记录该时间点
				timePointsByType[dp.Type][dpCopy.Time] = true
				seriesData[dp.Type] = append(seriesData[dp.Type], dpCopy)
			} else {
				// 对于单天数据，确保时间点格式一致
				dpCopy := dp
				// 标准化时间格式，确保所有时间格式一致
				if len(dp.Time) == 5 && dp.Time[2] == ':' {
					// 确保小时格式为两位数
					timeParts := strings.Split(dp.Time, ":")
					if len(timeParts) == 2 {
						hour, _ := strconv.Atoi(timeParts[0])
						minute, _ := strconv.Atoi(timeParts[1])
						dpCopy.Time = fmt.Sprintf("%02d:%02d", hour, minute)
					}
				}
				seriesData[dp.Type] = append(seriesData[dp.Type], dpCopy)
			}
		}

		// 在处理结束后再次检查数据点数量
		finalPointCount := 0
		for _, points := range seriesData {
			finalPointCount += len(points)
		}

		// 对比原始和最终数据点数量，详细记录差异
		log.Printf("[SendFeishuStandardChart] 系列 '%s' 原始数据点: %d, 处理后数据点: %d, 差异: %d",
			queryData.ChartTitle, rawDataPointCount, finalPointCount, rawDataPointCount-finalPointCount)

		if finalPointCount < rawDataPointCount {
			log.Printf("[SendFeishuStandardChart] 警告: 处理后数据点少于原始数据点，可能有合并或丢失")

			// 检查是否有某些日期的数据被全部丢失
			processedDates := make(map[string]bool)
			for _, points := range seriesData {
				for _, p := range points {
					date := time.Unix(p.UnixTime, 0).Format("2006-01-02")
					processedDates[date] = true
				}
			}

			log.Printf("[SendFeishuStandardChart] 处理后保留了 %d/%d 个日期",
				len(processedDates), uniqueDatesCount)

			// 找出丢失的日期
			missingDates := []string{}
			for date := range datesFound {
				if !processedDates[date] {
					missingDates = append(missingDates, date)
				}
			}

			if len(missingDates) > 0 {
				log.Printf("[SendFeishuStandardChart] 以下日期的数据点被完全丢失: %v", missingDates)
			}
		}

		// 获取有序的系列
		var seriesTypes []string
		for t := range seriesData {
			seriesTypes = append(seriesTypes, t)
		}
		sort.Strings(seriesTypes)

		// 构建图表数据
		var chartData []map[string]interface{}
		var chartSeries []map[string]interface{}

		// 预先检查每个系列的数据点数量
		seriesPointCounts := make(map[string]int)
		for _, seriesType := range seriesTypes {
			seriesPointCounts[seriesType] = len(seriesData[seriesType])
		}

		log.Printf("[SendFeishuStandardChart] 各系列数据点统计:")
		for seriesType, count := range seriesPointCounts {
			log.Printf(" - 系列 '%s': %d 个数据点", seriesType, count)
		}

		// 检查是否存在某些系列数据点数量明显少于其他系列的情况
		var avgPointCount float64
		totalPoints := 0
		for _, count := range seriesPointCounts {
			totalPoints += count
		}

		if len(seriesPointCounts) > 0 {
			avgPointCount = float64(totalPoints) / float64(len(seriesPointCounts))

			log.Printf("[SendFeishuStandardChart] 平均每个系列有 %.1f 个数据点", avgPointCount)

			// 检查不平衡的系列
			for seriesType, count := range seriesPointCounts {
				if float64(count) < avgPointCount*0.7 {
					log.Printf("[SendFeishuStandardChart] 警告: 系列 '%s' 的数据点数量(%d)明显少于平均值(%.1f)",
						seriesType, count, avgPointCount)
				}
			}
		}

		// 检查所有系列的时间范围是否一致
		seriesTimeRanges := make(map[string][2]int64) // seriesType -> [minTime, maxTime]
		for seriesType, points := range seriesData {
			if len(points) == 0 {
				continue
			}

			// 初始化为第一个点的时间
			minTime := points[0].UnixTime
			maxTime := points[0].UnixTime

			// 查找最小和最大时间
			for _, p := range points {
				if p.UnixTime < minTime {
					minTime = p.UnixTime
				}
				if p.UnixTime > maxTime {
					maxTime = p.UnixTime
				}
			}

			seriesTimeRanges[seriesType] = [2]int64{minTime, maxTime}
		}

		// 输出各系列的时间范围
		if len(seriesTimeRanges) > 0 {
			log.Printf("[SendFeishuStandardChart] 各系列的时间范围:")
			for seriesType, timeRange := range seriesTimeRanges {
				minTimeStr := time.Unix(timeRange[0], 0).Format("2006-01-02 15:04:05")
				maxTimeStr := time.Unix(timeRange[1], 0).Format("2006-01-02 15:04:05")
				log.Printf(" - 系列 '%s': %s ~ %s", seriesType, minTimeStr, maxTimeStr)
			}
		}

		for i, seriesType := range seriesTypes {
			points := seriesData[seriesType]
			// 重要：按UnixTime排序，确保时间点顺序正确
			sort.Slice(points, func(i, j int) bool {
				return points[i].UnixTime < points[j].UnixTime
			})

			// 记录排序后的日期顺序，以便验证数据完整性
			if len(points) > 0 {
				// log.Printf("[SendFeishuStandardChart] 系列 '%s' 的时间点日期顺序:", seriesType)
				dateOrder := []string{}
				for _, p := range points {
					dateStr := time.Unix(p.UnixTime, 0).Format("2006-01-02")
					dateOrder = append(dateOrder, dateStr)
				}
				// 只显示不重复的日期顺序
				uniqueDates := []string{}
				seenDates := make(map[string]bool)
				for _, date := range dateOrder {
					if !seenDates[date] {
						uniqueDates = append(uniqueDates, date)
						seenDates[date] = true
					}
				}
				log.Printf(" - 日期顺序: %v", uniqueDates)
			}

			// 创建数据点
			var chartPoints []map[string]interface{}

			// 输出每个数据点的详细信息用于调试
			// log.Printf("[SendFeishuStandardChart] 处理系列 '%s' 的数据点:", seriesType)
			for j, p := range points {
				// log.Printf("[SendFeishuStandardChart]   - 点 %d: Time='%s', UnixTime=%d, Value=%f",
				// 	j+1, p.Time, p.UnixTime, p.Value)

				// 确保每个数据点被正确添加到chartPoints
				chartPoints = append(chartPoints, map[string]interface{}{
					"x":    p.Time,
					"y":    p.Value,
					"name": p.Type,
					// 添加额外的时间戳信息，确保点的唯一性和正确排序
					"unix": p.UnixTime,
					// 添加额外的序号属性，确保不同日期的相同时间点能够区分
					"seq": j,
				})
			}

			// log.Printf("[SendFeishuStandardChart] 系列 '%s' 最终生成了 %d 个图表数据点",
			// 	seriesType, len(chartPoints))

			chartData = append(chartData, map[string]interface{}{
				"values": chartPoints,
			})

		// 确定使用的单位：优先使用该查询的独立单位，否则使用任务级别的单位
		currentUnit := queryData.Unit
		if currentUnit == "" {
			currentUnit = unit
		}
		log.Printf("[SendFeishuStandardChart] 系列 '%s' 使用单位: '%s' (queryData.Unit='%s', task.unit='%s')",
			queryData.ChartTitle, currentUnit, queryData.Unit, unit)

		// 添加系列配置 - 只使用飞书支持的类型
		chartSeries = append(chartSeries, map[string]interface{}{
			"type":      GetSupportedChartType(queryData.ChartType), // 使用queryData的图表类型，确保受飞书支持
			"stack":     false,                                      // 设置stack为false，禁用堆叠效果
			"dataIndex": i,
			// 添加数据标签配置，显示单位
			"label": map[string]interface{}{
				"visible": showDataLabel, // 使用传入的 showDataLabel 参数
				"formatter": func() string {
					if currentUnit == "%" {
						return "{y}%"
					} else if currentUnit != "" {
						return "{y}" + currentUnit
					}
					return "{y}"
				}(),
			},
			"seriesField": "name",
				"xField": func() interface{} {
					if GetSupportedChartType(queryData.ChartType) == "bar" {
						return []string{"x", "name"}
					}
					return "x"
				}(),
				"yField": "y",
			})
		}

		// 构建完全符合飞书API的图表元素
		chartElement := map[string]interface{}{
			"tag": "chart",
			"chart_spec": map[string]interface{}{
				"type":   "common",
				"data":   chartData,
				"series": chartSeries,
				// 改进坐标轴配置，优化多天标签显示
				"axes": []map[string]interface{}{
					{
						"orient": "bottom",
						// 如果是多天数据，调整X轴标签显示
						"label": map[string]interface{}{
							"visible":      true,
							"autoRotate":   isMultiDayData, // 多天数据时自动旋转
							"autoHide":     false,          // 不自动隐藏标签
							"autoEllipsis": false,          // 不自动省略标签
							// 对于多天数据，调整标签字体大小和旋转角度
							"style": map[string]interface{}{
								"fontSize": func() int {
									if isMultiDayData {
										return 10 // 多天数据使用较小字体
									}
									return 12 // 单天数据使用正常字体
								}(),
								"angle": func() int {
									if isMultiDayData {
										return 45 // 多天数据旋转标签
									}
									return 0 // 单天数据不旋转
								}(),
							},
							// 确保标签文本不被截断
							"autoLimit": false,
							"maxWidth":  150, // 增加最大宽度
							"minWidth":  40,  // 设置最小宽度
						},
						// 对于多天数据，优化刻度点数量
						"tickCount": func() int {
							// 获取所有数据点的数量，确保刻度数量合理
							totalPoints := 0
							uniqueTimePoints := make(map[string]bool)

							for _, data := range chartData {
								if values, ok := data["values"].([]map[string]interface{}); ok {
									totalPoints += len(values)
									// 收集唯一的时间点
									for _, v := range values {
										if x, ok := v["x"].(string); ok {
											uniqueTimePoints[x] = true
										}
									}
								}
							}

							uniquePointCount := len(uniqueTimePoints)
							log.Printf("[SendFeishuStandardChart] 图表共有 %d 个唯一时间点, 总计 %d 个数据点",
								uniquePointCount, totalPoints)

							if isMultiDayData {
								// 根据总点数调整刻度数量，确保不会太密集也不会太稀疏
								if uniquePointCount > 20 {
									return 15 // 大量数据点时，较多刻度
								} else if uniquePointCount > 10 {
									return 10 // 中等数据点，适中刻度
								} else if uniquePointCount > 5 {
									return uniquePointCount // 少量点时显示全部
								}
								return 5 // 默认刻度数
							}
							return 5 // 单天数据使用5个刻度点
						}(),
						// 优化刻度线配置
						"grid": map[string]interface{}{
							"visible":   true,
							"alignTick": true,
						},
					},
				{
					"orient": "left",
					// 添加y轴标签配置，显示该查询的独立单位
					"label": map[string]interface{}{
						"visible": true,
						"formatter": func() string {
							currentYUnit := queryData.Unit
							if currentYUnit == "" {
								currentYUnit = unit // 向后兼容：如果没有设置独立单位，使用任务级别的单位
							}
							if currentYUnit == "%" {
								return "{label}%"
							} else if currentYUnit != "" {
								return "{label}" + currentYUnit
							}
							return "{label}"
						}(),
					},
				},
				},
				// 图例配置
				"legends": map[string]interface{}{
					"position": "bottom",
				},
			// 添加与飞书要求匹配的tooltip配置，使用该查询的独立单位
			"tooltip": map[string]interface{}{
				"mark": map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"valueFormatter": func() string {
								tooltipUnit := queryData.Unit
								if tooltipUnit == "" {
									tooltipUnit = unit // 向后兼容
								}
								if tooltipUnit == "%" {
									return "{name}: {y}%"
								} else if tooltipUnit != "" {
									return "{name}: {y}" + tooltipUnit
								}
								return "{name}: {y}"
							}(),
						},
					},
				},
				"dimension": map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"valueFormatter": func() string {
								tooltipUnit := queryData.Unit
								if tooltipUnit == "" {
									tooltipUnit = unit // 向后兼容
								}
								if tooltipUnit == "%" {
									return "{name}: {y}%"
								} else if tooltipUnit != "" {
									return "{name}: {y}" + tooltipUnit
								}
								return "{name}: {y}"
							}(),
						},
					},
				},
			},
			},
		}

		elements = append(elements, chartElement)

		// 如果不是最后一个查询，添加分隔线
		if i < len(queryDataPoints)-1 {
			elements = append(elements, map[string]interface{}{
				"tag": "hr",
			})
		}
	}

	// 添加底部元素
	// 先添加分割线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 添加按钮
	elements = append(elements, map[string]interface{}{
		"tag": "action",
		"actions": []map[string]interface{}{
			{
				"tag": "button",
				"text": map[string]interface{}{
					"content": buttonText,
					"tag":     "plain_text",
				},
				"type": "primary",
				"url":  buttonURL,
			},
		},
	})

	// 添加原来的底部元素
	elements = append(elements, map[string]interface{}{
		"tag": "note",
		"elements": []map[string]interface{}{
			{
				"tag":     "lark_md",
				"content": "DeepRoute.ai " + time.Now().Format("2006-01-02 15:04:05"),
			},
		},
	})

	// 更新卡片中的元素
	cardData["card"].(map[string]interface{})["elements"] = elements

	// // 打印最终的卡片数据用于调试
	// debugData, _ := json.MarshalIndent(cardData, "", "  ")
	// log.Printf("[SendFeishuStandardChart] 完整卡片数据:\n%s", string(debugData))

	// 添加图表数据统计和诊断信息，增强问题排查能力
	if chartConfig, ok := cardData["card"].(map[string]interface{})["elements"].([]map[string]interface{}); ok {
		var chartDataPoints []map[string]interface{}
		var chartSeriesItems []map[string]interface{}

		// 查找图表组件和其中的数据点
		for _, element := range chartConfig {
			if element["tag"] == "chart" && element["chart_id"] == "standard_chart" {
				if chartData, ok := element["data"].(map[string]interface{}); ok {
					if dataPoints, ok := chartData["data"].([]map[string]interface{}); ok {
						chartDataPoints = dataPoints
					}
					if series, ok := chartData["series"].([]map[string]interface{}); ok {
						chartSeriesItems = series
					}
				}
				break
			}
		}

		// 输出图表数据统计
		if len(chartDataPoints) > 0 || len(chartSeriesItems) > 0 {
			log.Printf("[SendFeishuStandardChart] 图表数据统计: %d个系列, %d个数据点",
				len(chartSeriesItems), len(chartDataPoints))

			// 检查数据点和日期格式
			timeFormats := make(map[string]int)
			for _, data := range chartDataPoints {
				if timeStr, ok := data["time"].(string); ok && len(timeStr) >= 2 {
					timeFormats[timeStr[:2]] = timeFormats[timeStr[:2]] + 1
				}
			}

			if len(timeFormats) > 0 {
				log.Printf("[SendFeishuStandardChart] 图表中的日期前缀统计:")
				for prefix, count := range timeFormats {
					log.Printf(" - 日期前缀 '%s': %d个数据点", prefix, count)
				}
			}

			// 输出最终发送到飞书的前几个数据点作为示例
			maxSamplePoints := 5 // 最多显示5个样本
			if len(chartDataPoints) > 0 {
				sampleCount := len(chartDataPoints)
				if sampleCount > maxSamplePoints {
					sampleCount = maxSamplePoints
				}

				log.Printf("[SendFeishuStandardChart] 前%d个数据点示例:", sampleCount)
				for i := 0; i < sampleCount; i++ {
					dataJSON, _ := json.Marshal(chartDataPoints[i])
					log.Printf(" - 数据点%d: %s", i+1, string(dataJSON))
				}
			}
		}
	}

	// 直接使用 HTTP 请求发送到飞书
	jsonData, err := json.Marshal(cardData)
	if err != nil {
		return fmt.Errorf("JSON编码错误: %w", err)
	}

	// 在发送前对系列数据进行去重
	if elements, ok := cardData["card"].(map[string]interface{})["elements"].([]interface{}); ok {
		for _, element := range elements {
			if elem, ok := element.(map[string]interface{}); ok {
				if elem["tag"] == "chart" {
					if chartSpec, ok := elem["chart_spec"].(map[string]interface{}); ok {
						if data, ok := chartSpec["data"].([]map[string]interface{}); ok {
							// 使用map进行去重
							seenSeries := make(map[string]bool)
							var uniqueData []map[string]interface{}

							for _, series := range data {
								if values, ok := series["values"].([]map[string]interface{}); ok && len(values) > 0 {
									// 使用第一个数据点的name作为系列标识
									if name, ok := values[0]["name"].(string); ok {
										if !seenSeries[name] {
											seenSeries[name] = true
											uniqueData = append(uniqueData, series)
										} else {
											log.Printf("[SendFeishuStandardChart] 跳过重复的系列: %s", name)
										}
									}
								}
							}
							chartSpec["data"] = uniqueData
						}
					}
				}
			}
		}
	}

	log.Printf("[SendFeishuStandardChart] 发送图表卡片 (长度: %d 字节) 到 webhook: %s", len(jsonData), webhookURL)

	// 重试逻辑
	var lastErr error
	retryCount := 0
	maxRetries := 3
	baseWaitTime := 5 * time.Second

	for retryCount < maxRetries {
		if retryCount > 0 {
			// 使用指数退避策略，每次重试等待时间翻倍
			waitTime := baseWaitTime * time.Duration(1<<uint(retryCount-1))
			log.Printf("[SendFeishuStandardChart] 第%d次重试，等待 %v 后继续...", retryCount, waitTime)
			time.Sleep(waitTime)
		}

		// 创建请求
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("创建请求错误: %w", err)
			retryCount++
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP请求失败: %w", err)
			retryCount++
			continue
		}

		// 读取响应
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("[SendFeishuStandardChart] 响应状态: %d, 内容: %s", resp.StatusCode, string(body))

		// 解析响应JSON
		var result struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("解析响应JSON失败: %w, 响应内容: %s", err, string(body))
			retryCount++
			continue
		}

		// 检查飞书API返回的状态码
		if result.Code != 0 {
			// 特殊处理频率限制错误
			if result.Code == 9499 || result.Code == 11232 || strings.Contains(result.Msg, "frequency limited") || strings.Contains(result.Msg, "too many request") {
				lastErr = fmt.Errorf("飞书API频率限制: code=%d, msg=%s", result.Code, result.Msg)
				retryCount++
				continue
			}

			lastErr = fmt.Errorf("飞书API错误: code=%d, msg=%s", result.Code, result.Msg)
			retryCount++
			continue
		}

		// 成功发送
		log.Printf("[SendFeishuStandardChart] 发送成功！")

		// 记录发送记录，添加按钮文本和按钮链接信息
		AddSendRecord(models.SendRecord{
			Timestamp:  time.Now(),
			Status:     "success",
			Message:    fmt.Sprintf("成功发送图表消息: %s (按钮: %s)", cardTitle, buttonText),
			Webhook:    webhookURL,
			TaskName:   cardTitle,
			ButtonText: buttonText,
			ButtonURL:  buttonURL,
		})

		return nil
	}

	log.Printf("[SendFeishuStandardChart] 所有重试都失败，最后错误: %v", lastErr)

	// 记录失败的发送记录
	AddSendRecord(models.SendRecord{
		Timestamp:  time.Now(),
		Status:     "error",
		Message:    fmt.Sprintf("发送失败: %v", lastErr),
		Webhook:    webhookURL,
		TaskName:   cardTitle,
		ButtonText: buttonText,
		ButtonURL:  buttonURL,
	})

	return lastErr
}

// SendFeishuTextCard 发送文本卡片消息到飞书
// 此函数用于以纯文本格式展示多个 PromQL 查询的最新值
//
// 参数说明:
//   - webhookURL: 飞书 webhook URL
//   - promqlMetrics: 每个 PromQL 的最新指标值列表，格式为 map[promqlName][]LatestMetric
//   - promqlConfigs: 每个 PromQL 的配置（单位、标签等）
//   - cardTitle: 卡片标题
//   - cardTemplate: 卡片颜色主题（"blue", "red", "green" 等）
//   - buttonText: 按钮文本（可选）
//   - buttonURL: 按钮链接（可选）
func SendFeishuTextCard(webhookURL string, promqlMetrics map[string][]LatestMetric, promqlConfigs map[string]struct {
	Name              string
	Unit              string
	MetricLabel       string
	CustomMetricLabel string
	InitialUnit       string
}, promqlOrder []string, cardTitle, cardTemplate, buttonText, buttonURL string) error {
	log.Printf("[SendFeishuTextCard] ====== START ======")
	log.Printf("[SendFeishuTextCard] Webhook: %s, CardTitle: %s", webhookURL, cardTitle)
	log.Printf("[SendFeishuTextCard] PromQL 显示顺序: %v", promqlOrder)

	// 构建卡片
	card := &FeishuCard{
		MsgType: "interactive",
		Card: FeishuCardBody{
			Config: &FeishuCardConfig{
				WideScreenMode: true,
				EnableForward:  true,
			},
			Header: &FeishuCardHeader{
				Title: &FeishuCardHeaderTitle{
					Content: cardTitle,
					Tag:     "plain_text",
				},
				Template: cardTemplate,
			},
			Elements: []FeishuCardElement{},
		},
	}

	// 按照指定的顺序为每个 PromQL 添加一个部分
	for _, promqlName := range promqlOrder {
		metrics, exists := promqlMetrics[promqlName]
		if !exists {
			log.Printf("[SendFeishuTextCard] Warning: PromQL '%s' not found in metrics", promqlName)
			continue
		}
		config, hasConfig := promqlConfigs[promqlName]
		if !hasConfig {
			log.Printf("[SendFeishuTextCard] Warning: No config found for PromQL '%s'", promqlName)
			continue
		}

		// 添加 PromQL 名称和单位作为标题
		titleText := fmt.Sprintf("**%s**", promqlName)
		if config.Unit != "" {
			titleText = fmt.Sprintf("**%s** (%s)", promqlName, config.Unit)
		}

		card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
			Tag:     "markdown",
			Content: titleText,
		})

		// 如果没有指标数据，显示无数据
		if len(metrics) == 0 {
			card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
				Tag:     "markdown",
				Content: "└─ 暂无数据",
			})
		} else {
			// 排序指标（按标签名称）
			sort.Slice(metrics, func(i, j int) bool {
				return metrics[i].Label < metrics[j].Label
			})

			// 显示每个指标的最新值
			for i, metric := range metrics {
				var prefix string
				if i < len(metrics)-1 {
					prefix = "├─"
				} else {
					prefix = "└─"
				}

				// 格式化值
				valueStr := formatValue(metric.Value, config.Unit)

				// 构建显示文本
				displayText := fmt.Sprintf("%s %s: %s", prefix, metric.Label, valueStr)

				card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
					Tag:     "markdown",
					Content: displayText,
				})
			}
		}
	}

	// 添加分割线
	card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
		Tag: "hr",
	})

	// 添加按钮（如果提供）
	if buttonText != "" && buttonURL != "" {
		card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
			Tag: "action",
			Actions: []FeishuAction{
				{
					Tag: "button",
					Text: &FeishuActionText{
						Content: buttonText,
						Tag:     "plain_text",
					},
					Type: "default",
					URL:  buttonURL,
				},
			},
		})
	}

	// 添加数据采集时间
	now := time.Now().In(ChinaTimezone)
	timeText := fmt.Sprintf("⏰ 数据时间: %s", now.Format("2006-01-02 15:04"))
	card.Card.Elements = append(card.Card.Elements, FeishuCardElement{
		Tag:     "markdown",
		Content: timeText,
	})

	// 发送消息
	err := SendFeishuCardMessage(webhookURL, card)
	if err != nil {
		log.Printf("[SendFeishuTextCard] Failed to send message: %v", err)
		return err
	}

	log.Printf("[SendFeishuTextCard] ====== END ======")
	return nil
}

// formatValue 格式化值，添加单位
func formatValue(value float64, unit string) string {
	// 格式化为两位小数
	valueStr := fmt.Sprintf("%.2f", value)

	// 移除尾部多余的零
	valueStr = strings.TrimRight(valueStr, "0")
	valueStr = strings.TrimRight(valueStr, ".")

	// 添加单位
	if unit != "" {
		return fmt.Sprintf("%s %s", valueStr, unit)
	}

	return valueStr
}
