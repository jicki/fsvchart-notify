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
	"time"
)

// HybridElement 表示一个混合元素（图表或文本）
type HybridElement struct {
	DisplayOrder int
	DisplayMode  string // "chart" 或 "text"
	PromQLName   string

	// 图表相关
	ChartData       *models.QueryDataPoints
	ChartType       string
	ShowDataLabel   bool

	// 文本相关
	TextMetrics []LatestMetric
	Unit        string
	MetricLabel string
}

// SendFeishuHybridCard 发送混合卡片消息到飞书
// 此函数用于在同一个卡片内混合展示图表和文本内容
//
// 参数说明:
//   - webhookURL: 飞书 webhook URL
//   - hybridElements: 混合元素列表（包含图表和文本），已按 display_order 排序
//   - cardTitle: 卡片标题
//   - cardTemplate: 卡片颜色主题（"blue", "red", "green" 等）
//   - buttonText: 按钮文本（可选）
//   - buttonURL: 按钮链接（可选）
func SendFeishuHybridCard(webhookURL string, hybridElements []HybridElement, cardTitle, cardTemplate, unit, buttonText, buttonURL string, showDataLabel bool) error {
	log.Printf("[SendFeishuHybridCard] ====== START ======")
	log.Printf("[SendFeishuHybridCard] Webhook: %s, CardTitle: %s", webhookURL, cardTitle)
	log.Printf("[SendFeishuHybridCard] 混合元素数量: %d", len(hybridElements))

	// 按 display_order 排序
	sort.Slice(hybridElements, func(i, j int) bool {
		if hybridElements[i].DisplayOrder != hybridElements[j].DisplayOrder {
			return hybridElements[i].DisplayOrder < hybridElements[j].DisplayOrder
		}
		// display_order 相同时，按名称排序
		return hybridElements[i].PromQLName < hybridElements[j].PromQLName
	})

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
			"elements": []interface{}{},
		},
	}

	// 获取所有元素的引用
	elements := cardData["card"].(map[string]interface{})["elements"].([]interface{})

	// 是否为多天数据（用于图表时间格式）
	isMultiDayData := false
	if len(hybridElements) > 0 {
		for _, elem := range hybridElements {
			if elem.DisplayMode == "chart" && elem.ChartData != nil && len(elem.ChartData.DataPoints) > 0 {
				// 检查是否有多个不同日期
				dates := make(map[string]bool)
				for _, dp := range elem.ChartData.DataPoints {
					date := time.Unix(dp.UnixTime, 0).In(ChinaTimezone).Format("01-02")
					dates[date] = true
					if len(dates) > 1 {
						isMultiDayData = true
						break
					}
				}
				if isMultiDayData {
					break
				}
			}
		}
	}

	log.Printf("[SendFeishuHybridCard] 是否多天数据: %v", isMultiDayData)

	// 按顺序添加元素
	for idx, elem := range hybridElements {
		log.Printf("[SendFeishuHybridCard] 处理元素 %d: %s (mode=%s, order=%d)", idx+1, elem.PromQLName, elem.DisplayMode, elem.DisplayOrder)

		if elem.DisplayMode == "text" {
			// 添加文本元素
			elements = appendTextElements(elements, elem)
		} else if elem.DisplayMode == "chart" {
			// 添加图表元素
			elements = appendChartElements(elements, elem, isMultiDayData)
		}

		// 在元素之间添加分隔线（除了最后一个）
		if idx < len(hybridElements)-1 {
			elements = append(elements, map[string]interface{}{
				"tag": "hr",
			})
		}
	}

	// 添加底部分隔线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 添加按钮（如果提供）
	if buttonText != "" && buttonURL != "" {
		elements = append(elements, map[string]interface{}{
			"tag": "action",
			"actions": []map[string]interface{}{
				{
					"tag": "button",
					"text": map[string]interface{}{
						"content": buttonText,
						"tag":     "plain_text",
					},
					"type": "default",
					"url":  buttonURL,
				},
			},
		})
	}

	// 添加时间戳
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

	// 发送消息
	err := SendFeishuCardMessageFromMap(webhookURL, cardData)
	if err != nil {
		log.Printf("[SendFeishuHybridCard] Failed to send message: %v", err)
		
		// 记录失败的发送记录
		AddSendRecord(models.SendRecord{
			Timestamp:  time.Now(),
			Status:     "error",
			Message:    fmt.Sprintf("发送失败: %v", err),
			Webhook:    webhookURL,
			TaskName:   cardTitle,
			ButtonText: buttonText,
			ButtonURL:  buttonURL,
		})
		
		return err
	}

	// 记录成功的发送记录
	AddSendRecord(models.SendRecord{
		Timestamp:  time.Now(),
		Status:     "success",
		Message:    fmt.Sprintf("成功发送混合卡片消息: %s", cardTitle),
		Webhook:    webhookURL,
		TaskName:   cardTitle,
		ButtonText: buttonText,
		ButtonURL:  buttonURL,
	})

	log.Printf("[SendFeishuHybridCard] ====== END ======")
	return nil
}

// appendTextElements 添加文本元素到卡片
func appendTextElements(elements []interface{}, elem HybridElement) []interface{} {
	// 添加 PromQL 名称和单位作为标题
	titleText := fmt.Sprintf("**%s**", elem.PromQLName)
	if elem.Unit != "" {
		titleText = fmt.Sprintf("**%s** (%s)", elem.PromQLName, elem.Unit)
	}

	elements = append(elements, map[string]interface{}{
		"tag":     "markdown",
		"content": titleText,
	})

	// 如果没有指标数据，显示无数据
	if len(elem.TextMetrics) == 0 {
		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": "└─ 暂无数据",
		})
		return elements
	}

	// 排序指标（按标签名称）
	sort.Slice(elem.TextMetrics, func(i, j int) bool {
		return elem.TextMetrics[i].Label < elem.TextMetrics[j].Label
	})

	// 显示每个指标的最新值
	for i, metric := range elem.TextMetrics {
		var prefix string
		if i < len(elem.TextMetrics)-1 {
			prefix = "├─"
		} else {
			prefix = "└─"
		}

		// 格式化值
		valueStr := formatValue(metric.Value, elem.Unit)

		// 构建显示文本
		displayText := fmt.Sprintf("%s %s: %s", prefix, metric.Label, valueStr)

		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": displayText,
		})
	}

	return elements
}

// appendChartElements 添加图表元素到卡片
func appendChartElements(elements []interface{}, elem HybridElement, isMultiDayData bool) []interface{} {
	if elem.ChartData == nil || len(elem.ChartData.DataPoints) == 0 {
		// 添加无数据提示
		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": fmt.Sprintf("**%s**", elem.PromQLName),
		})
		elements = append(elements, map[string]interface{}{
			"tag":     "markdown",
			"content": "└─ 图表数据: 暂无数据",
		})
		return elements
	}

	// 添加图表标题（使用 markdown）
	titleText := fmt.Sprintf("**%s**", elem.PromQLName)
	if elem.ChartData.Unit != "" {
		titleText = fmt.Sprintf("**%s** (%s)", elem.PromQLName, elem.ChartData.Unit)
	}
	elements = append(elements, map[string]interface{}{
		"tag":     "markdown",
		"content": titleText,
	})

	// 处理数据点 - 按飞书标准图表格式
	dataPointsMap := make(map[string]map[string]float64) // time -> seriesName -> value
	allTimes := make(map[string]bool)
	allSeries := make(map[string]bool)

	for _, dp := range elem.ChartData.DataPoints {
		// 格式化时间
		timeStr := formatTimeForChart(dp.UnixTime, isMultiDayData)
		allTimes[timeStr] = true
		allSeries[dp.Type] = true

		if dataPointsMap[timeStr] == nil {
			dataPointsMap[timeStr] = make(map[string]float64)
		}
		dataPointsMap[timeStr][dp.Type] = dp.Value
	}

	// 构建时间列表（排序）
	var times []string
	for t := range allTimes {
		times = append(times, t)
	}
	sort.Strings(times)

	// 构建系列列表
	var seriesNames []string
	for s := range allSeries {
		seriesNames = append(seriesNames, s)
	}
	sort.Strings(seriesNames)

	// 构建飞书图表数据格式（每个系列一个数据集）
	var chartData []map[string]interface{}
	for _, series := range seriesNames {
		var values []map[string]interface{}
		for _, t := range times {
			if seriesMap, ok := dataPointsMap[t]; ok {
				if val, exists := seriesMap[series]; exists {
					values = append(values, map[string]interface{}{
						"x":    t,
						"y":    val,
						"name": series,
					})
				}
			}
		}
		
		if len(values) > 0 {
			chartData = append(chartData, map[string]interface{}{
				"name":   series,
				"values": values,
			})
		}
	}

	// 构建系列配置
	var chartSeries []map[string]interface{}
	for i := range seriesNames {
		seriesConfig := map[string]interface{}{
			"type":        elem.ChartType,
			"dataIndex":   i,
			"seriesField": "name",
			"xField":      "x",
			"yField":      "y",
		}
		
		// 添加数据标签配置
		if elem.ShowDataLabel {
			labelFormatter := "{y}"
			if elem.ChartData.Unit != "" {
				if elem.ChartData.Unit == "%" {
					labelFormatter = "{y}%"
				} else {
					labelFormatter = "{y}" + elem.ChartData.Unit
				}
			}
			seriesConfig["label"] = map[string]interface{}{
				"visible":  true,
				"formatter": labelFormatter,
			}
		}
		
		chartSeries = append(chartSeries, seriesConfig)
	}

	// 使用飞书官方标准图表格式
	chartElement := map[string]interface{}{
		"tag": "chart",
		"chart_spec": map[string]interface{}{
			"type":   "common",
			"data":   chartData,
			"series": chartSeries,
			"axes": []map[string]interface{}{
				{
					"orient": "bottom",
					"label": map[string]interface{}{
						"visible":    true,
						"autoRotate": isMultiDayData,
					},
				},
				{
					"orient": "left",
					"label": map[string]interface{}{
						"visible": true,
					},
				},
			},
			"legends": map[string]interface{}{
				"position": "bottom",
			},
		},
	}

	elements = append(elements, chartElement)

	return elements
}

// formatTimeForChart 格式化时间用于图表显示
func formatTimeForChart(unixTime int64, isMultiDay bool) string {
	t := time.Unix(unixTime, 0).In(ChinaTimezone)
	if isMultiDay {
		return t.Format("01-02 15:04")
	}
	return t.Format("15:04")
}

// SendFeishuCardMessageFromMap 从 map 发送飞书消息
func SendFeishuCardMessageFromMap(webhookURL string, cardData map[string]interface{}) error {
	payload, err := json.Marshal(cardData)
	if err != nil {
		log.Printf("[SendFeishuCardMessageFromMap] JSON marshal error: %v", err)
		return fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("[SendFeishuCardMessageFromMap] Sending to webhook: %s, payload size: %d bytes", webhookURL, len(payload))

	// 重试逻辑
	var lastErr error
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("[SendFeishuCardMessageFromMap] Retry attempt %d/%d", retry, maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// 构造 POST 请求
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			lastErr = fmt.Errorf("create request error: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// 执行请求
		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http post error: %w", err)
			continue
		}

		// 读取响应
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		log.Printf("[SendFeishuCardMessageFromMap] Response status: %d, body: %s", resp.StatusCode, string(bodyBytes))

		// 解析飞书API的响应
		var feishuResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(bodyBytes, &feishuResp); err == nil {
			if feishuResp.Code == 0 {
				log.Printf("[SendFeishuCardMessageFromMap] Message sent successfully")
				return nil
			}
			lastErr = fmt.Errorf("feishu api error: code=%d, msg=%s", feishuResp.Code, feishuResp.Msg)
		} else {
			lastErr = fmt.Errorf("failed to parse feishu response: %w", err)
		}
	}

	return lastErr
}

