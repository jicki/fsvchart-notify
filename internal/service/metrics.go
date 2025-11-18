package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"fsvchart-notify/internal/models"
)

// 使用中国时区 (GMT+8)
var ChinaTimezone = time.FixedZone("GMT+8", 8*60*60)

type VMQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// DataPointSorter 用于排序数据点
type DataPointSorter struct {
	points []models.DataPoint
}

func (s DataPointSorter) Len() int { return len(s.points) }

func (s DataPointSorter) Swap(i, j int) {
	s.points[i], s.points[j] = s.points[j], s.points[i]
}

func (s DataPointSorter) Less(i, j int) bool {
	if s.points[i].UnixTime != s.points[j].UnixTime {
		return s.points[i].UnixTime < s.points[j].UnixTime
	}
	return s.points[i].Type < s.points[j].Type
}

// DebugDataPoints 输出每个数据点的详细信息用于调试
func DebugDataPoints(points []models.DataPoint, seriesType string) {
	log.Printf("[DebugDataPoints] 处理系列 '%s' 的数据点:", seriesType)
	for j, p := range points {
		dateFromUnix := time.Unix(p.UnixTime, 0).Format("2006-01-02")
		log.Printf("[DebugDataPoints]   - 点 %d: Time='%s', UnixTime=%d, 原始日期=%s, Value=%f",
			j+1, p.Time, p.UnixTime, dateFromUnix, p.Value)
	}
}

// isMultiDayData 对于多天数据，将日期添加到时间前缀
func isMultiDayData(points []models.DataPoint) bool {
	dateSet := make(map[string]bool)
	for _, p := range points {
		date := time.Unix(p.UnixTime, 0).Format("2006-01-02")
		dateSet[date] = true
	}
	return len(dateSet) > 1
}

// addDatePrefix 如果原始格式是时间，添加日期前缀
func addDatePrefix(dp models.DataPoint) models.DataPoint {
	dpCopy := dp
	originalDate := time.Unix(dp.UnixTime, 0).Format("2006-01-02")
	date := time.Unix(dp.UnixTime, 0).Format("01/02")

	log.Printf("[SendFeishuStandardChart] 数据点处理: UnixTime=%d, 原始日期=%s, 格式化日期=%s",
		dp.UnixTime, originalDate, date)

	if len(dp.Time) == 5 && dp.Time[2] == ':' {
		dpCopy.Time = date + " " + dp.Time
	}
	return dpCopy
}

// getTimeFormat 获取合适的时间格式
func getTimeFormat(step time.Duration) string {
	switch {
	case step >= 24*time.Hour:
		return "01/02"
	case step >= time.Hour:
		return "15:04"
	default:
		return "15:04"
	}
}

// GetDurationStep calculates an appropriate step size for the given time range
func GetDurationStep(d time.Duration) time.Duration {
	// Convert duration to hours for easier calculation
	durationHours := d.Hours()
	log.Printf("[GetDurationStep] Duration in hours: %.2f", durationHours)

	// 计算天数（向上取整）
	durationDays := int(math.Ceil(durationHours / 24))
	log.Printf("[GetDurationStep] Duration in days (rounded up): %d", durationDays)

	// 根据时间范围长度动态调整步长
	var stepHours float64
	switch {
	case durationHours <= 24: // 1天以内
		// 1小时一个点，但如果时间范围小于6小时则每30分钟一个点
		if durationHours <= 6 {
			stepHours = 0.5 // 30分钟
			log.Printf("[GetDurationStep] Duration ≤ 6h: using 30min step")
		} else {
			stepHours = 1 // 1小时
			log.Printf("[GetDurationStep] Duration ≤ 24h: using 1h step")
		}

	case durationHours <= 72: // 3天以内
		// 确保每天至少4个点（6小时一个点）
		stepHours = 6
		log.Printf("[GetDurationStep] Duration ≤ 72h (3d): using 6h step")

	case durationHours <= 168: // 7天以内
		// 确保每天至少3个点（8小时一个点）
		stepHours = 8
		log.Printf("[GetDurationStep] Duration ≤ 168h (7d): using 8h step")

	case durationHours <= 360: // 15天以内
		// 确保每天至少2个点（12小时一个点）
		stepHours = 12
		log.Printf("[GetDurationStep] Duration ≤ 360h (15d): using 12h step")

	case durationHours <= 720: // 30天以内
		// 确保每天至少1个点（24小时一个点）
		stepHours = 24
		log.Printf("[GetDurationStep] Duration ≤ 720h (30d): using 24h step")

	default: // 超过30天
		// 确保每天至少1个点
		stepHours = 24
		log.Printf("[GetDurationStep] Duration > 720h: using 24h step")
	}

	// 计算预期的数据点数量
	expectedPoints := durationHours / stepHours
	log.Printf("[GetDurationStep] Expected data points with %.1fh step: %.1f", stepHours, expectedPoints)

	// 确保每天至少有一个数据点
	minPointsNeeded := float64(durationDays) // 最少需要的点数等于天数
	if expectedPoints < minPointsNeeded {
		// 调整步长以确保至少有每天一个点
		stepHours = durationHours / minPointsNeeded
		// 向下取整到最接近的整数小时
		stepHours = math.Floor(stepHours)
		if stepHours > 24 {
			stepHours = 24 // 最大步长为24小时
		}
		log.Printf("[GetDurationStep] Adjusted step to %.1fh to ensure at least one point per day", stepHours)
	}

	// 如果数据点太多，增加步长
	maxPoints := 90.0
	if expectedPoints > maxPoints {
		proposedStepHours := durationHours / maxPoints
		// 确保新步长不会导致每天少于一个点
		if proposedStepHours <= 24 {
			stepHours = proposedStepHours
			// 向上取整到最接近的整数小时
			stepHours = math.Ceil(stepHours)
			log.Printf("[GetDurationStep] Adjusted step to %.1fh to limit maximum points while maintaining daily coverage", stepHours)
		}
	}

	// 转换为Duration
	step := time.Duration(stepHours * float64(time.Hour))

	// 最终验证：确保步长不会导致某些天没有数据点
	finalExpectedPoints := durationHours / stepHours
	if finalExpectedPoints < float64(durationDays) {
		// 如果最终点数少于天数，强制使用24小时步长
		step = 24 * time.Hour
		log.Printf("[GetDurationStep] Final adjustment: forcing 24h step to ensure daily coverage")
	}

	log.Printf("[GetDurationStep] Final step duration: %v (will generate approximately %.1f points over %d days)",
		step, durationHours/step.Hours(), durationDays)

	return step
}

// FetchMetrics 从VictoriaMetrics获取指标数据并处理成前端可用的数据点格式
//
// 参数说明:
//   - baseURL: 指标源的基础URL
//   - query: PromQL查询语句
//   - start/end: 查询的时间范围
//   - step: 数据点的时间间隔
//   - seriesType: 默认的标签名称，用于从指标中提取序列名称（当customLabel为空时使用）
//   - customLabel: 自定义标签名称，用于从指标数据中提取对应的值作为数据点的Type。
//     例如，如果customLabel="resource"，则会从指标数据中提取"resource"标签对应的值（如"cpu"、"memory"、"gpu"等）
//     作为DataPoint.Type。如果指标中不存在该标签的值，则会跳过该结果。
//     此参数还会过滤数据点，确保只有customLabel对应的值会显示在图表上，而不显示其他标签的值（如team="mlp"）。
func FetchMetrics(baseURL, query string, start, end time.Time, step time.Duration, seriesType string, customLabel string, initialUnit string, targetUnit string) ([]models.DataPoint, error) {
	logMsg := fmt.Sprintf("[FetchMetrics] ====== START ======")
	log.Print(logMsg)
	GetLogManager().AddLog(logMsg)

	log.Printf("[FetchMetrics] Parameters:")
	log.Printf("[FetchMetrics] Query: %s, TimeRange: %v to %v, Step: %v, SeriesType: %s, CustomLabel: %s, InitialUnit: %s, TargetUnit: %s",
		query, start, end, step, seriesType, customLabel, initialUnit, targetUnit)

	duration := end.Sub(start)
	log.Printf("[FetchMetrics] Duration: %v, Input step: %v", duration, step)

	// 验证传入的步长是否合理
	expectedPoints := float64(duration) / float64(step)
	log.Printf("[FetchMetrics] Expected points with input step: %.2f", expectedPoints)

	// 只有在预期数据点太少时才重新计算步长
	if expectedPoints < 5 {
		log.Printf("[FetchMetrics] Expected points too few (%.2f), recalculating step", expectedPoints)
		// 使用 GetDurationStep 计算合适的步长
		newStep := GetDurationStep(duration)
		log.Printf("[FetchMetrics] Recalculated step: %v", newStep)
		step = newStep

		// 再次检查预期数据点
		expectedPoints = float64(duration) / float64(step)
		log.Printf("[FetchMetrics] New expected points: %.2f", expectedPoints)

		// 确保至少有7个数据点
		if expectedPoints < 7 {
			adjustedStep := time.Duration(float64(duration) / 10)
			log.Printf("[FetchMetrics] Still too few points, forcing step to ensure at least 10 points: %v", adjustedStep)
			step = adjustedStep
			expectedPoints = float64(duration) / float64(step)
			log.Printf("[FetchMetrics] Final expected points: %.2f", expectedPoints)
		}
	}

	// 根据时间范围的不同单位，调整时间格式
	timeFormat := getTimeFormat(step)
	log.Printf("[FetchMetrics] Using time format: %s for step: %v", timeFormat, step)

	log.Printf("[FetchMetrics] Final step: %v", step)
	log.Printf("[FetchMetrics] Expected points: %.2f", float64(duration)/float64(step))

	// 重新计算开始和结束时间，确保对齐到合适的时间点
	var alignedStart, alignedEnd time.Time

	// 对于多天查询，确保获取完整的历史数据
	if duration.Hours() > 24 {
		// 计算天数（向上取整）
		durationDays := int(math.Ceil(duration.Hours() / 24))
		log.Printf("[FetchMetrics] Query spans %d days", durationDays)

		// 使用当前时间作为结束时间，不进行截断，确保获取最新数据
		now := time.Now()
		alignedEnd = now
		log.Printf("[FetchMetrics] Using current time as end: %s", alignedEnd.Format("2006-01-02 15:04:05"))

		// 计算开始时间：从结束时间向前推指定天数
		alignedStart = alignedEnd.AddDate(0, 0, -durationDays)
		// 确保开始时间对齐到当天00:00
		alignedStart = time.Date(
			alignedStart.Year(), alignedStart.Month(), alignedStart.Day(),
			0, 0, 0, 0, alignedStart.Location(),
		)
		log.Printf("[FetchMetrics] Aligned start time: %s", alignedStart.Format("2006-01-02 15:04:05"))

		// 固定使用8小时步长，确保每天有3个点 (00:00, 08:00, 16:00)
		step = 8 * time.Hour
		log.Printf("[FetchMetrics] Using 8h step to ensure 3 points per day")

		// 生成每天的时间点列表，但不包含未来时间点
		var timePoints []time.Time
		currentTime := alignedStart
		for currentTime.Before(alignedEnd) || currentTime.Equal(alignedEnd) {
			// 为每天生成3个固定时间点
			dayStart := time.Date(
				currentTime.Year(), currentTime.Month(), currentTime.Day(),
				0, 0, 0, 0, currentTime.Location(),
			)

			// 只添加不晚于当前时间的时间点
			if !dayStart.After(alignedEnd) {
				timePoints = append(timePoints, dayStart) // 00:00
			}
			eightAM := dayStart.Add(8 * time.Hour)
			if !eightAM.After(alignedEnd) {
				timePoints = append(timePoints, eightAM) // 08:00
			}
			fourPM := dayStart.Add(16 * time.Hour)
			if !fourPM.After(alignedEnd) {
				timePoints = append(timePoints, fourPM) // 16:00
			}

			// 移动到下一天
			currentTime = currentTime.AddDate(0, 0, 1)
		}

		// 记录生成的时间点
		log.Printf("[FetchMetrics] Generated %d time points:", len(timePoints))
		for i, tp := range timePoints {
			log.Printf("  Point %d: %s", i+1, tp.Format("2006-01-02 15:04:05"))
		}

		// 更新查询参数
		// 开始时间使用第一个时间点
		// 结束时间使用当前时间（alignedEnd），这样 Prometheus 会返回到当前时间为止的所有 8 小时整点数据
		alignedStart = timePoints[0]
		// alignedEnd 保持为当前时间，不需要修改
		log.Printf("[FetchMetrics] Query will use start=%s, end=%s (current time)", 
			alignedStart.Format("2006-01-02 15:04:05"), 
			alignedEnd.Format("2006-01-02 15:04:05"))
	} else {
		// 对于小于一天的查询，使用常规对齐
		alignedStart = start.Truncate(step)
		alignedEnd = end.Truncate(step)
		if !end.Equal(alignedEnd) {
			alignedEnd = alignedEnd.Add(step)
		}
	}

	// 构建查询 URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/v1/query_range")

	// 记录最终的查询参数
	log.Printf("[FetchMetrics] Final query parameters:")
	log.Printf("  - Start: %s", alignedStart.Format("2006-01-02 15:04:05"))
	log.Printf("  - End: %s", alignedEnd.Format("2006-01-02 15:04:05"))
	log.Printf("  - Step: %v", step)
	log.Printf("  - Expected points per day: %.1f", 24*time.Hour.Hours()/step.Hours())

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", alignedStart.Unix()))
	params.Set("end", fmt.Sprintf("%d", alignedEnd.Unix()))
	params.Set("step", fmt.Sprintf("%d", int64(step.Seconds())))
	u.RawQuery = params.Encode()

	log.Printf("[FetchMetrics] Requesting URL: %s", u.String())

	// 记录详细的请求信息
	requestDetails := fmt.Sprintf("QueryParams: query=%s, start=%s, end=%s, step=%s",
		query,
		time.Unix(alignedStart.Unix(), 0).Format("2006-01-02 15:04:05"),
		time.Unix(alignedEnd.Unix(), 0).Format("2006-01-02 15:04:05"),
		step.String())
	log.Printf("[FetchMetrics] Request details: %s", requestDetails)

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("[FetchMetrics] ERROR: HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 记录响应状态码
	log.Printf("[FetchMetrics] Response status code: %d", resp.StatusCode)

	// 如果响应过大，只记录前1000个字符
	responsePreview := string(body)
	if len(responsePreview) > 1000 {
		responsePreview = responsePreview[:1000] + "... (truncated)"
	}
	log.Printf("[FetchMetrics] Response preview: %s", responsePreview)

	// 解析 JSON
	var vmResp VMQueryResponse
	if err := json.Unmarshal(body, &vmResp); err != nil {
		return nil, err
	}

	log.Printf("[FetchMetrics] Response status: %s, Result count: %d",
		vmResp.Status, len(vmResp.Data.Result))

	// 记录响应中每个结果的时间点数量和时间范围
	for i, result := range vmResp.Data.Result {
		if i < 5 { // 只记录前5个结果，避免日志过长
			log.Printf("[FetchMetrics] Result[%d] has %d values, metric=%v",
				i, len(result.Values), result.Metric)

			// 记录每个结果的时间范围，便于诊断
			if len(result.Values) > 0 {
				firstPointTime := time.Unix(int64(result.Values[0][0].(float64)), 0).Format("2006-01-02 15:04:05")
				lastPointTime := time.Unix(int64(result.Values[len(result.Values)-1][0].(float64)), 0).Format("2006-01-02 15:04:05")
				log.Printf("[FetchMetrics] Result[%d] time range: %s to %s",
					i, firstPointTime, lastPointTime)
			}
		}
	}

	// 检查响应状态
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", vmResp.Status)
	}

	// 声明所有需要的变量
	var allPoints []models.DataPoint
	var labelValues []string
	// 记录每个标签的实际数据点
	actualPoints := make(map[string]map[int64]float64)
	// 记录每个标签的有效时间戳
	validTimeStamps := make(map[string]map[int64]bool)

	// 生成所有需要的时间点
	var timeStamps []int64
	// 对于多天查询，不生成时间戳，直接使用 Prometheus 返回的时间戳
	// 对于单天查询，从结束时间向前生成时间点
	if duration.Hours() <= 24 {
		// 从结束时间向前生成时间点，确保优先显示最近的数据
		for t := alignedEnd.Unix(); t >= alignedStart.Unix(); t -= int64(step.Seconds()) {
			timeStamps = append(timeStamps, t)
		}
	}

	// 如果时间点过多，进行采样（仅对单天查询）
	const maxPoints = 90 // 最大数据点数量
	if duration.Hours() <= 24 && len(timeStamps) > maxPoints {
		log.Printf("[FetchMetrics] Too many timestamps (%d), sampling down to %d points",
			len(timeStamps), maxPoints)

		// 计算采样间隔
		sampleInterval := len(timeStamps) / maxPoints
		if sampleInterval < 1 {
			sampleInterval = 1
		}

		// 采样时间戳，优先保留最近的点
		sampledTimeStamps := make([]int64, 0, maxPoints)
		for i := 0; i < len(timeStamps); i += sampleInterval {
			sampledTimeStamps = append(sampledTimeStamps, timeStamps[i])
		}

		// 确保至少包含开始和结束时间点
		if len(sampledTimeStamps) > 0 && sampledTimeStamps[len(sampledTimeStamps)-1] != timeStamps[len(timeStamps)-1] {
			sampledTimeStamps = append(sampledTimeStamps, timeStamps[len(timeStamps)-1])
		}

		timeStamps = sampledTimeStamps
		log.Printf("[FetchMetrics] After sampling: %d points", len(timeStamps))
	}

	// 保持时间戳按照从最近到最早的顺序
	if duration.Hours() <= 24 {
		log.Printf("[FetchMetrics] Generated %d time stamps (ordered from newest to oldest)", len(timeStamps))
		// 打印所有生成的时间戳
		for i, ts := range timeStamps {
			dateTime := time.Unix(ts, 0).Format("2006-01-02 15:04:05")
			if i < 5 || i >= len(timeStamps)-5 {
				// 只打印前5个和最后5个时间戳，避免日志过长
				log.Printf("[FetchMetrics] TimeStamp[%d]: %d (%v)", i, ts, dateTime)
			} else if i == 5 {
				log.Printf("[FetchMetrics] ... skipping middle timestamps ...")
			}
		}
	} else {
		log.Printf("[FetchMetrics] Multi-day query: will use timestamps from Prometheus response")
	}

	// 记录时间范围信息
	if len(timeStamps) > 0 {
		firstTime := time.Unix(timeStamps[0], 0).Format("2006-01-02 15:04:05")
		lastTime := time.Unix(timeStamps[len(timeStamps)-1], 0).Format("2006-01-02 15:04:05")
		log.Printf("[FetchMetrics] Time range of data points (newest to oldest): %s to %s", firstTime, lastTime)
	}

	// 检查响应中是否包含所有预期的时间戳（仅对单天查询）
	if duration.Hours() <= 24 && len(timeStamps) > 0 {
		timeStampMap := make(map[int64]bool)
		for _, ts := range timeStamps {
			timeStampMap[ts] = false // 初始为未找到
		}

		// 记录时间戳覆盖情况
		for _, result := range vmResp.Data.Result {
			// 按时间戳从新到旧排序数据点
			sort.Slice(result.Values, func(i, j int) bool {
				return result.Values[i][0].(float64) > result.Values[j][0].(float64)
			})

			for _, point := range result.Values {
				ts := int64(point[0].(float64))
				if _, exists := timeStampMap[ts]; exists {
					timeStampMap[ts] = true // 标记为已找到
				}
			}
		}

		// 计算覆盖率
		foundCount := 0
		for _, found := range timeStampMap {
			if found {
				foundCount++
			}
		}

		coverageRate := float64(foundCount) / float64(len(timeStampMap)) * 100
		log.Printf("[FetchMetrics] Time range coverage: %.2f%% (%d/%d timestamps found in response)",
			coverageRate, foundCount, len(timeStampMap))

		// 如果覆盖率低于50%，记录详细信息
		if coverageRate < 50 {
			log.Printf("[FetchMetrics] WARNING: Low time range coverage, this may cause missing data!")

			// 列出缺失的时间点（从最近到最早）
			missingCount := 0
			log.Printf("[FetchMetrics] Missing timestamps (newest to oldest):")
			for _, ts := range timeStamps {
				if !timeStampMap[ts] {
					missingCount++
					if missingCount <= 10 { // 只显示前10个
						log.Printf("  - %s", time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
					}
				}
			}
			if missingCount > 10 {
				log.Printf("  ... and %d more missing timestamps", missingCount-10)
			}
		}
	} else if duration.Hours() > 24 {
		// 对于多天查询，记录Prometheus返回的实际时间戳范围
		log.Printf("[FetchMetrics] Multi-day query: using all timestamps from Prometheus response")
	}

	// 处理每个时间序列
	for _, result := range vmResp.Data.Result {
		log.Printf("[FetchMetrics] Raw metric data: %+v", result.Metric)

		// 按时间戳从新到旧排序数据点
		sort.Slice(result.Values, func(i, j int) bool {
			return result.Values[i][0].(float64) > result.Values[j][0].(float64)
		})

		// // 打印原始结果的完整数据，帮助调试
		// rawJSON, _ := json.MarshalIndent(result, "", "  ")
		// log.Printf("[FetchMetrics] Raw result data (sorted newest to oldest):\n%s", string(rawJSON))

		// 检查该结果的时间范围并记录
		if len(result.Values) > 0 {
			firstPointTime := time.Unix(int64(result.Values[0][0].(float64)), 0)
			lastPointTime := time.Unix(int64(result.Values[len(result.Values)-1][0].(float64)), 0)
			log.Printf("[FetchMetrics] Processing result with time range: %s to %s (%d points)",
				firstPointTime.Format("2006-01-02 15:04:05"),
				lastPointTime.Format("2006-01-02 15:04:05"),
				len(result.Values))

			// 检查时间覆盖率
			resultDuration := lastPointTime.Sub(firstPointTime)
			expectedDuration := alignedEnd.Sub(alignedStart)
			coveragePercent := float64(resultDuration) / float64(expectedDuration) * 100
			log.Printf("[FetchMetrics] Time coverage: %.1f%% (result: %v, expected: %v)",
				coveragePercent, resultDuration, expectedDuration)

			// 警告时间覆盖不足
			if coveragePercent < 70 {
				log.Printf("[FetchMetrics] WARNING: Result only covers %.1f%% of the requested time range",
					coveragePercent)
			}
		}

		// 如果提供了自定义标签，则使用它
		var labelValue string
		if customLabel != "" {
			// 使用指标中的自定义标签值
			labelValue = result.Metric[customLabel]
			log.Printf("[FetchMetrics] Using custom label '%s' with value '%s'", customLabel, labelValue)

			// 如果找不到该标签，跳过这个结果数据，不处理
			if labelValue == "" {
				log.Printf("[FetchMetrics] WARNING: Custom label '%s' not found in metrics, skipping this series", customLabel)

				// 增加详细日志，记录被跳过的数据点数量和时间范围
				if len(result.Values) > 0 {
					firstPoint := result.Values[0]
					lastPoint := result.Values[len(result.Values)-1]

					firstTime := time.Unix(int64(firstPoint[0].(float64)), 0).Format("2006-01-02 15:04:05")
					lastTime := time.Unix(int64(lastPoint[0].(float64)), 0).Format("2006-01-02 15:04:05")

					log.Printf("[FetchMetrics] SKIPPED DATA: %d points from %s to %s",
						len(result.Values), firstTime, lastTime)
					log.Printf("[FetchMetrics] Skipped metric labels: %v", result.Metric)

					// 列出前几个被跳过的时间点，帮助检查是否包含特定日期的数据
					log.Printf("[FetchMetrics] First few skipped timestamps:")
					limit := 5
					if len(result.Values) < limit {
						limit = len(result.Values)
					}

					for i := 0; i < limit; i++ {
						point := result.Values[i]
						pointTime := time.Unix(int64(point[0].(float64)), 0).Format("2006-01-02 15:04:05")
						log.Printf("  - [%d] %s = %s", i, pointTime, point[1])
					}
				}

				continue
			}

			// 添加额外调试信息，确认最终使用的标签值
			log.Printf("[FetchMetrics] Final label value for custom label '%s': '%s'", customLabel, labelValue)
		}

		// 如果没有自定义标签或自定义标签为空，则使用默认标签
		if labelValue == "" {
			// 按原逻辑从指标中提取标签
			labelValue = result.Metric[seriesType]
			log.Printf("[FetchMetrics] Trying to extract label '%s' from metrics: %v", seriesType, result.Metric)

			if labelValue == "" {
				log.Printf("[FetchMetrics] WARNING: Label '%s' not found in metrics, trying fallback options", seriesType)

				// 尝试特定查询的处理方法
				if strings.Contains(query, "k8s_cluster_rate_cpu") && seriesType == "cluster" {
					log.Printf("[FetchMetrics] Detected k8s_cluster_rate_cpu query with cluster label")
					// 如果是集群CPU查询且需要cluster标签，尝试使用整个查询标识为标签
					labelValue = "cluster_cpu"
					log.Printf("[FetchMetrics] Using hardcoded label 'cluster_cpu' for this specific query")
				} else {
					// 一般回退处理
					for k, v := range result.Metric {
						if k != "__name__" {
							labelValue = v
							log.Printf("[FetchMetrics] Using fallback label '%s' with value '%s'", k, v)
							break
						}
					}
				}

				// 如果仍然没有找到标签值，使用查询作为标签
				if labelValue == "" {
					log.Printf("[FetchMetrics] Still no label found, using query-based label")
					if strings.Contains(query, "cpu") {
						labelValue = "CPU使用率"
					} else if strings.Contains(query, "memory") {
						labelValue = "内存使用"
					} else {
						// 使用查询的最后部分作为标签
						parts := strings.Split(query, "_")
						if len(parts) > 0 {
							labelValue = parts[len(parts)-1]
							log.Printf("[FetchMetrics] Using last part of query as label: '%s'", labelValue)
						} else {
							labelValue = "指标" // 最后的后备标签
						}
					}
				}
			} else {
				log.Printf("[FetchMetrics] Found label '%s' with value '%s'", seriesType, labelValue)
			}
		}

		// 如果这个标签值之前没见过，记录下来
		if _, exists := validTimeStamps[labelValue]; !exists {
			validTimeStamps[labelValue] = make(map[int64]bool)
			labelValues = append(labelValues, labelValue)
			// log.Printf("[FetchMetrics] Added new label value: %s", labelValue)
		}

		if _, exists := actualPoints[labelValue]; !exists {
			actualPoints[labelValue] = make(map[int64]float64)
		}

		// 处理每个数据点
		for _, point := range result.Values {
			ts := int64(point[0].(float64))
			// 将时间戳转换为中国时区
			// timePoint := time.Unix(ts, 0).In(ChinaTimezone)
			val := point[1].(string)
			var floatVal float64
			var err error

			// ## 打印原始值和转换后的值
			// log.Printf("[FetchMetrics] Processing value: %s (type=%T) at time %v",
			// 	val, val, timePoint.Format("2006-01-02 15:04:05"))

			// 处理各种数值格式
			val = strings.TrimSpace(val)

			// 检查是否为科学计数法
			if strings.Contains(val, "e") || strings.Contains(val, "E") {
				log.Printf("[FetchMetrics] Found scientific notation: %s", val)
				floatVal, err = strconv.ParseFloat(val, 64)
				if err != nil {
					log.Printf("[FetchMetrics] ERROR: Failed to parse scientific notation: %v", err)
					continue
				}
			} else {
				// 处理百分比值
				isPercentage := strings.HasSuffix(val, "%")
				if isPercentage {
					log.Printf("[FetchMetrics] Found percentage value: %s", val)
					val = strings.TrimSuffix(val, "%")
				}

				// 处理 milli 值
				if strings.HasSuffix(val, "m") {
					log.Printf("[FetchMetrics] Found milli value: %s", val)
					val = strings.TrimSuffix(val, "m")
					floatVal, err = strconv.ParseFloat(val, 64)
					if err != nil {
						log.Printf("[FetchMetrics] ERROR: Failed to parse milli value: %v", err)
						continue
					}
					floatVal = floatVal / 1000
				} else {
					floatVal, err = strconv.ParseFloat(val, 64)
					if err != nil {
						log.Printf("[FetchMetrics] ERROR: Failed to parse value: %v", err)
						continue
					}
				}

				if isPercentage && floatVal > 1 {
					log.Printf("[FetchMetrics] Converting large percentage: %f", floatVal)
					floatVal = floatVal / 100
				}
			}

			// 应用单位转换（如果配置了 initial_unit）
			if initialUnit != "" && targetUnit != "" {
				convertedValue, err := ConvertUnit(floatVal, initialUnit, targetUnit)
				if err != nil {
					log.Printf("[FetchMetrics] WARNING: Unit conversion failed (%s -> %s): %v, using original value", initialUnit, targetUnit, err)
				} else {
					log.Printf("[FetchMetrics] Unit conversion: %.2f %s -> %.2f %s", floatVal, initialUnit, convertedValue, targetUnit)
					floatVal = convertedValue
				}
			}

			floatVal = math.Round(floatVal*100) / 100
			// ## 打印原始值和转换后的值
			// log.Printf("[FetchMetrics] Value conversion: original='%s', parsed=%f, final=%f",
			// 	point[1].(string), floatVal, floatVal)
			actualPoints[labelValue][ts] = floatVal
			validTimeStamps[labelValue][ts] = true
		}
		// log.Printf("[FetchMetrics] Processed %d points for label %s", len(result.Values), labelValue)

		// 打印该标签的所有时间点和值
		// log.Printf("[FetchMetrics] Time points for label %s:", labelValue)
		var timeKeys []int64
		for ts := range actualPoints[labelValue] {
			timeKeys = append(timeKeys, ts)
		}
		sort.Slice(timeKeys, func(i, j int) bool { return timeKeys[i] < timeKeys[j] })
		// 打印该标签的所有时间点和值
		// for _, ts := range timeKeys {
		// 	log.Printf("[FetchMetrics] %v (%s) = %f",
		// 		ts,
		// 		time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
		// 		actualPoints[labelValue][ts])
		// }

		// 如果数据点太少，进行插值填充（仅对单天查询）
		if duration.Hours() <= 24 && len(timeKeys) < 5 && len(timeKeys) > 0 {
			log.Printf("[FetchMetrics] Too few points (%d) for label %s, performing interpolation",
				len(timeKeys), labelValue)

			// 确保时间戳是有序的
			sort.Slice(timeKeys, func(i, j int) bool { return timeKeys[i] < timeKeys[j] })

			// 如果只有一个数据点，复制到所有时间戳
			if len(timeKeys) == 1 {
				singleValue := actualPoints[labelValue][timeKeys[0]]
				log.Printf("[FetchMetrics] Only one data point, duplicating value %.2f to all timestamps", singleValue)

				for _, ts := range timeStamps {
					if !validTimeStamps[labelValue][ts] {
						actualPoints[labelValue][ts] = singleValue
						validTimeStamps[labelValue][ts] = true
						log.Printf("[FetchMetrics] Added interpolated point at %v with value %.2f",
							time.Unix(ts, 0).Format("2006-01-02 15:04:05"), singleValue)
					}
				}
			} else {
				// 有多个数据点，进行线性插值
				log.Printf("[FetchMetrics] Performing linear interpolation between %d existing points", len(timeKeys))

				// 对每个缺失的时间戳进行插值
				for i, ts := range timeStamps {
					if validTimeStamps[labelValue][ts] {
						continue // 已有数据点，跳过
					}

					// 找到最近的两个有效时间戳进行插值
					var beforeTS, afterTS int64
					var beforeVal, afterVal float64
					var foundBefore, foundAfter bool

					// 向前查找
					for j := i - 1; j >= 0; j-- {
						checkTS := timeStamps[j]
						if validTimeStamps[labelValue][checkTS] {
							beforeTS = checkTS
							beforeVal = actualPoints[labelValue][checkTS]
							foundBefore = true
							break
						}
					}

					// 向后查找
					for j := i + 1; j < len(timeStamps); j++ {
						checkTS := timeStamps[j]
						if validTimeStamps[labelValue][checkTS] {
							afterTS = checkTS
							afterVal = actualPoints[labelValue][checkTS]
							foundAfter = true
							break
						}
					}

					// 根据找到的点进行插值
					if foundBefore && foundAfter {
						// 线性插值
						ratio := float64(ts-beforeTS) / float64(afterTS-beforeTS)
						interpolatedVal := beforeVal + ratio*(afterVal-beforeVal)
						interpolatedVal = math.Round(interpolatedVal*100) / 100

						actualPoints[labelValue][ts] = interpolatedVal
						validTimeStamps[labelValue][ts] = true

						log.Printf("[FetchMetrics] Interpolated point at %v: %.2f (between %.2f and %.2f)",
							time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
							interpolatedVal, beforeVal, afterVal)
					} else if foundBefore {
						// 只有前面的点，使用前面的值
						actualPoints[labelValue][ts] = beforeVal
						validTimeStamps[labelValue][ts] = true

						log.Printf("[FetchMetrics] Extended point at %v with previous value: %.2f",
							time.Unix(ts, 0).Format("2006-01-02 15:04:05"), beforeVal)
					} else if foundAfter {
						// 只有后面的点，使用后面的值
						actualPoints[labelValue][ts] = afterVal
						validTimeStamps[labelValue][ts] = true

						log.Printf("[FetchMetrics] Extended point at %v with next value: %.2f",
							time.Unix(ts, 0).Format("2006-01-02 15:04:05"), afterVal)
					}
				}
			}

			// 重新计算有效的时间点
			timeKeys = []int64{}
			for ts := range actualPoints[labelValue] {
				if validTimeStamps[labelValue][ts] {
					timeKeys = append(timeKeys, ts)
				}
			}
			log.Printf("[FetchMetrics] After interpolation: %d points for label %s", len(timeKeys), labelValue)
		}
	}

	// 对标签值进行排序
	sort.Strings(labelValues)
	log.Printf("[FetchMetrics] Sorted label values: %v", labelValues)

	// 自定义标签过滤：确保只处理含有该标签的数据
	if customLabel != "" {
		// 记录过滤前的标签值
		log.Printf("[FetchMetrics] Before filtering, label values: %v", labelValues)

		// 过滤出有效的标签值（即那些是customLabel对应的值）
		validLabelValues := []string{}
		for _, lv := range labelValues {
			// 只保留与自定义标签值相关的标签
			isValid := false
			// 检查是否为指定标签的值
			for _, result := range vmResp.Data.Result {
				if result.Metric[customLabel] == lv {
					isValid = true
					break
				}
			}

			if isValid {
				validLabelValues = append(validLabelValues, lv)
			} else {
				log.Printf("[FetchMetrics] Filtering out unrelated label value: '%s' for customLabel '%s'",
					lv, customLabel)
			}
		}

		// 用过滤后的列表替换原列表
		labelValues = validLabelValues
		log.Printf("[FetchMetrics] After filtering for customLabel='%s', valid label values: %v",
			customLabel, labelValues)
	}

	// 对于多天查询，额外获取当前时刻的值，确保图表显示最新数据
	if duration.Hours() > 24 {
		// 获取最后一个数据点的时间戳
		var lastDataTimestamp int64
		for _, timestamps := range actualPoints {
			for ts := range timestamps {
				if ts > lastDataTimestamp {
					lastDataTimestamp = ts
				}
			}
		}
		
		currentTime := time.Now()
		lastDataTime := time.Unix(lastDataTimestamp, 0)
		
		// 如果最后一个数据点距离现在超过1小时，则获取当前时刻的值
		if currentTime.Sub(lastDataTime) > time.Hour {
			log.Printf("[FetchMetrics] Last data point is at %s, fetching current value at %s",
				lastDataTime.Format("2006-01-02 15:04:05"),
				currentTime.Format("2006-01-02 15:04:05"))
			
			// 使用 query API 获取当前值
			currentValueURL, err := url.Parse(baseURL)
			if err == nil {
				currentValueURL.Path = path.Join(currentValueURL.Path, "/api/v1/query")
				params := url.Values{}
				params.Set("query", query)
				currentValueURL.RawQuery = params.Encode()
				
				resp, err := http.Get(currentValueURL.String())
				if err == nil {
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err == nil {
						// 定义 query API 的响应结构（返回 Value 而不是 Values）
						var currentResp struct {
							Status string `json:"status"`
							Data   struct {
								ResultType string `json:"resultType"`
								Result     []struct {
									Metric map[string]string `json:"metric"`
									Value  []interface{}     `json:"value"` // [timestamp, value]
								} `json:"result"`
							} `json:"data"`
						}
						
						if err := json.Unmarshal(body, &currentResp); err == nil {
							if currentResp.Status == "success" {
								log.Printf("[FetchMetrics] Successfully fetched current values, result count: %d", len(currentResp.Data.Result))
								
								// 将当前值添加到 actualPoints
								currentTimestamp := currentTime.Unix()
								for _, result := range currentResp.Data.Result {
									var labelValue string
									if customLabel != "" {
										if val, exists := result.Metric[customLabel]; exists {
											labelValue = val
										}
									} else if seriesType != "" {
										if val, exists := result.Metric[seriesType]; exists {
											labelValue = val
										}
									}
									
									if labelValue != "" {
										if len(result.Value) >= 2 {
											val := result.Value[1].(string)
											originalVal, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
											if err == nil {
												floatVal := originalVal
												// 应用单位转换
												if initialUnit != "" && targetUnit != "" {
													convertedVal, err := ConvertUnit(originalVal, initialUnit, targetUnit)
													if err == nil {
														floatVal = convertedVal
														log.Printf("[FetchMetrics] Unit conversion for current value: %.2f %s -> %.2f %s",
															originalVal, initialUnit, convertedVal, targetUnit)
													}
												}
												
												if actualPoints[labelValue] == nil {
													actualPoints[labelValue] = make(map[int64]float64)
												}
												// 四舍五入到2位小数
												floatVal = math.Round(floatVal*100) / 100
												actualPoints[labelValue][currentTimestamp] = floatVal
												log.Printf("[FetchMetrics] Added current value for %s: %.2f at %s",
													labelValue, floatVal, currentTime.Format("2006-01-02 15:04:05"))
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 处理数据点
	for _, labelValue := range labelValues {
		// 获取该标签的所有时间戳并按从新到旧排序
		var labelTimeStamps []int64
		for t := range actualPoints[labelValue] {
			labelTimeStamps = append(labelTimeStamps, t)
		}
		sort.Slice(labelTimeStamps, func(i, j int) bool {
			return labelTimeStamps[i] > labelTimeStamps[j] // 从新到旧排序
		})

		// 为每个时间戳创建数据点
		for _, ts := range labelTimeStamps {
			value := actualPoints[labelValue][ts]
			timeStr := time.Unix(ts, 0).Format("01/02 15:04")

			dp := models.DataPoint{
				Time:     timeStr,
				UnixTime: ts,
				Value:    value,
				Type:     labelValue,
			}
			allPoints = append(allPoints, dp)
		}
	}

	// 确保数据点按时间从新到旧排序
	sort.Slice(allPoints, func(i, j int) bool {
		return allPoints[i].UnixTime > allPoints[j].UnixTime
	})

	// // 打印最终的排序结果
	// log.Printf("[FetchMetrics] Final sorted points (newest to oldest):")
	// for i, p := range allPoints {
	// 	log.Printf("[FetchMetrics] [%d] Time=%s (%d) Type=%s Value=%f",
	// 		i, p.Time, p.UnixTime, p.Type, p.Value)
	// }

	// log.Printf("[FetchMetrics] ====== END ======")
	// log.Printf("[FetchMetrics] Total points generated: %d", len(allPoints))

	return allPoints, nil
}

// LatestMetric 表示一个时间序列的最新指标值
type LatestMetric struct {
	Label string    `json:"label"` // 标签值（如pod名称、namespace等）
	Value float64   `json:"value"` // 最新值
	Time  time.Time `json:"time"`  // 最新值的时间戳
}

// FetchLatestMetrics 从VictoriaMetrics获取指标的最新值
//
// 参数说明:
//   - baseURL: 指标源的基础URL
//   - query: PromQL查询语句
//   - seriesType: 默认的标签名称，用于从指标中提取序列名称（当customLabel为空时使用）
//   - customLabel: 自定义标签名称，用于从指标数据中提取对应的值作为标签
//
// 返回值:
//   - []LatestMetric: 每个时间序列的最新指标值列表
//   - error: 错误信息
func FetchLatestMetrics(baseURL, query, seriesType, customLabel, initialUnit, targetUnit string) ([]LatestMetric, error) {
	log.Printf("[FetchLatestMetrics] ====== START ======")
	log.Printf("[FetchLatestMetrics] Query: %s, SeriesType: %s, CustomLabel: %s, InitialUnit: %s, TargetUnit: %s",
		query, seriesType, customLabel, initialUnit, targetUnit)

	// 构建查询 URL - 使用 query 接口获取即时值
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/v1/query")

	params := url.Values{}
	params.Set("query", query)
	u.RawQuery = params.Encode()

	log.Printf("[FetchLatestMetrics] Requesting URL: %s", u.String())

	// 发送 HTTP 请求
	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("[FetchLatestMetrics] ERROR: HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[FetchLatestMetrics] Response status code: %d", resp.StatusCode)

	// 解析 JSON
	var vmResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"` // [timestamp, value]
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &vmResp); err != nil {
		return nil, err
	}

	log.Printf("[FetchLatestMetrics] Response status: %s, Result count: %d",
		vmResp.Status, len(vmResp.Data.Result))

	// 检查响应状态
	if vmResp.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", vmResp.Status)
	}

	// 处理每个时间序列，提取最新值
	var latestMetrics []LatestMetric
	for _, result := range vmResp.Data.Result {
		log.Printf("[FetchLatestMetrics] Raw metric data: %+v", result.Metric)

		// 确定标签值
		var labelValue string
		if customLabel != "" {
			// 使用指标中的自定义标签值
			labelValue = result.Metric[customLabel]
			log.Printf("[FetchLatestMetrics] Using custom label '%s' with value '%s'", customLabel, labelValue)

			// 如果找不到该标签，跳过这个结果
			if labelValue == "" {
				log.Printf("[FetchLatestMetrics] WARNING: Custom label '%s' not found in metrics, skipping this series", customLabel)
				continue
			}
		} else {
			// 使用默认标签
			labelValue = result.Metric[seriesType]
			log.Printf("[FetchLatestMetrics] Using default label '%s' with value '%s'", seriesType, labelValue)

			if labelValue == "" {
				// 回退处理：使用第一个非 __name__ 的标签
				for k, v := range result.Metric {
					if k != "__name__" {
						labelValue = v
						log.Printf("[FetchLatestMetrics] Using fallback label '%s' with value '%s'", k, v)
						break
					}
				}
			}

			// 如果仍然没有找到标签值，使用默认标签
			if labelValue == "" {
				labelValue = "指标"
				log.Printf("[FetchLatestMetrics] Using default label '指标'")
			}
		}

		// 解析值
		if len(result.Value) < 2 {
			log.Printf("[FetchLatestMetrics] WARNING: Invalid value format, skipping")
			continue
		}

		timestamp := int64(result.Value[0].(float64))
		valueStr := result.Value[1].(string)

	// 转换值为 float64
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("[FetchLatestMetrics] ERROR: Failed to parse value '%s': %v", valueStr, err)
		continue
	}

	// 应用单位转换（如果配置了 initial_unit）
	if initialUnit != "" && targetUnit != "" {
		convertedValue, err := ConvertUnit(value, initialUnit, targetUnit)
		if err != nil {
			log.Printf("[FetchLatestMetrics] WARNING: Unit conversion failed (%s -> %s): %v, using original value", initialUnit, targetUnit, err)
		} else {
			log.Printf("[FetchLatestMetrics] Unit conversion: %.2f %s -> %.2f %s", value, initialUnit, convertedValue, targetUnit)
			value = convertedValue
		}
	}

	// 四舍五入到两位小数
	value = math.Round(value*100) / 100

	// 创建最新指标记录
	latestMetric := LatestMetric{
		Label: labelValue,
		Value: value,
		Time:  time.Unix(timestamp, 0),
	}

		latestMetrics = append(latestMetrics, latestMetric)
		log.Printf("[FetchLatestMetrics] Added metric: Label=%s, Value=%.2f, Time=%s",
			latestMetric.Label, latestMetric.Value, latestMetric.Time.Format("2006-01-02 15:04:05"))
	}

	log.Printf("[FetchLatestMetrics] ====== END ======")
	log.Printf("[FetchLatestMetrics] Total latest metrics: %d", len(latestMetrics))

	return latestMetrics, nil
}
