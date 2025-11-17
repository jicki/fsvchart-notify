package service

import (
	"fmt"
	"math"
	"strings"
)

// ConvertUnit 单位转换函数
// fromUnit: 原始单位 (如 "bytes", "B", "KB", "ms", "s" 等)
// toUnit: 目标单位 (如 "TB", "GB", "s", "m" 等)
// value: 原始值
// 返回: 转换后的值
func ConvertUnit(value float64, fromUnit, toUnit string) (float64, error) {
	// 如果没有指定初始单位或单位相同，直接返回原值
	if fromUnit == "" || toUnit == "" || strings.EqualFold(fromUnit, toUnit) {
		return value, nil
	}

	// 标准化单位名称（转为小写）
	fromUnit = strings.ToLower(strings.TrimSpace(fromUnit))
	toUnit = strings.ToLower(strings.TrimSpace(toUnit))

	// 根据单位类型选择转换方式
	if isBytesUnit(fromUnit) && isBytesUnit(toUnit) {
		return convertBytes(value, fromUnit, toUnit)
	} else if isTimeUnit(fromUnit) && isTimeUnit(toUnit) {
		return convertTime(value, fromUnit, toUnit)
	}

	// 单位类型不匹配，返回原值
	return value, fmt.Errorf("无法在不同类型的单位之间转换: %s -> %s", fromUnit, toUnit)
}

// isBytesUnit 判断是否为字节单位
func isBytesUnit(unit string) bool {
	bytesUnits := []string{"b", "bytes", "byte", "kb", "mb", "gb", "tb", "pb", "eb",
		"kib", "mib", "gib", "tib", "pib", "eib"}
	unit = strings.ToLower(unit)
	for _, u := range bytesUnits {
		if u == unit {
			return true
		}
	}
	return false
}

// isTimeUnit 判断是否为时间单位
func isTimeUnit(unit string) bool {
	timeUnits := []string{"ns", "nanosecond", "nanoseconds",
		"us", "μs", "microsecond", "microseconds",
		"ms", "millisecond", "milliseconds",
		"s", "second", "seconds",
		"m", "min", "minute", "minutes",
		"h", "hour", "hours",
		"d", "day", "days"}
	unit = strings.ToLower(unit)
	for _, u := range timeUnits {
		if u == unit {
			return true
		}
	}
	return false
}

// convertBytes 字节单位转换
func convertBytes(value float64, fromUnit, toUnit string) (float64, error) {
	// 定义单位到字节的转换比率
	unitToBytes := map[string]float64{
		"b":     1,
		"byte":  1,
		"bytes": 1,
		// 十进制单位 (1000)
		"kb": 1000,
		"mb": 1000 * 1000,
		"gb": 1000 * 1000 * 1000,
		"tb": 1000 * 1000 * 1000 * 1000,
		"pb": 1000 * 1000 * 1000 * 1000 * 1000,
		"eb": 1000 * 1000 * 1000 * 1000 * 1000 * 1000,
		// 二进制单位 (1024)
		"kib": 1024,
		"mib": 1024 * 1024,
		"gib": 1024 * 1024 * 1024,
		"tib": 1024 * 1024 * 1024 * 1024,
		"pib": 1024 * 1024 * 1024 * 1024 * 1024,
		"eib": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}

	fromRatio, ok1 := unitToBytes[strings.ToLower(fromUnit)]
	toRatio, ok2 := unitToBytes[strings.ToLower(toUnit)]

	if !ok1 {
		return value, fmt.Errorf("未知的字节单位: %s", fromUnit)
	}
	if !ok2 {
		return value, fmt.Errorf("未知的字节单位: %s", toUnit)
	}

	// 先转换为字节，再转换为目标单位
	bytes := value * fromRatio
	result := bytes / toRatio

	return result, nil
}

// convertTime 时间单位转换
func convertTime(value float64, fromUnit, toUnit string) (float64, error) {
	// 定义单位到纳秒的转换比率
	unitToNanoseconds := map[string]float64{
		"ns":          1,
		"nanosecond":  1,
		"nanoseconds": 1,
		"us":          1000,
		"μs":          1000,
		"microsecond": 1000,
		"microseconds": 1000,
		"ms":           1000 * 1000,
		"millisecond":  1000 * 1000,
		"milliseconds": 1000 * 1000,
		"s":            1000 * 1000 * 1000,
		"second":       1000 * 1000 * 1000,
		"seconds":      1000 * 1000 * 1000,
		"m":            60 * 1000 * 1000 * 1000,
		"min":          60 * 1000 * 1000 * 1000,
		"minute":       60 * 1000 * 1000 * 1000,
		"minutes":      60 * 1000 * 1000 * 1000,
		"h":            60 * 60 * 1000 * 1000 * 1000,
		"hour":         60 * 60 * 1000 * 1000 * 1000,
		"hours":        60 * 60 * 1000 * 1000 * 1000,
		"d":            24 * 60 * 60 * 1000 * 1000 * 1000,
		"day":          24 * 60 * 60 * 1000 * 1000 * 1000,
		"days":         24 * 60 * 60 * 1000 * 1000 * 1000,
	}

	fromRatio, ok1 := unitToNanoseconds[strings.ToLower(fromUnit)]
	toRatio, ok2 := unitToNanoseconds[strings.ToLower(toUnit)]

	if !ok1 {
		return value, fmt.Errorf("未知的时间单位: %s", fromUnit)
	}
	if !ok2 {
		return value, fmt.Errorf("未知的时间单位: %s", toUnit)
	}

	// 先转换为纳秒，再转换为目标单位
	nanoseconds := value * fromRatio
	result := nanoseconds / toRatio

	return result, nil
}

// FormatUnitValue 格式化单位值（保留合适的小数位数）
func FormatUnitValue(value float64, precision int) float64 {
	if precision < 0 {
		precision = 2 // 默认保留2位小数
	}
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}

