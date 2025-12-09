package common

import (
	"time"
)

// 时间格式常量
const (
	ISO8601Format = "2006-01-02T15:04:05"
)

// ParseISO8601Time 解析ISO 8601格式的时间字符串
func ParseISO8601Time(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	return time.Parse(ISO8601Format, timeStr)
}

// FormatISO8601Time 将时间格式化为ISO 8601字符串
func FormatISO8601Time(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(ISO8601Format)
}

// DetectLanguage 检测文本语言（简单实现，主要用于返回对应语言的回复）
func DetectLanguage(text string) string {
	// 简单实现：检测是否包含中文字符
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return "zh"
		}
	}
	return "en"
}

// IsValidID 验证ID是否有效
func IsValidID(id uint64) bool {
	return id > 0
}
