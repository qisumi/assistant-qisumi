package common

import (
	"strings"
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

// ExtractJSON 从可能包含 Markdown 代码块的字符串中提取 JSON
func ExtractJSON(content string) string {
	// 尝试提取 ```json 代码块
	if jsonStart := findCodeBlock(content, "```json"); jsonStart != "" {
		return jsonStart
	}
	// 尝试提取普通 ``` 代码块
	if codeStart := findCodeBlock(content, "```"); codeStart != "" {
		return codeStart
	}
	return content
}

// findCodeBlock 查找并提取代码块内容
func findCodeBlock(content, marker string) string {
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	start += len(marker)
	
	end := strings.Index(content[start:], "```")
	if end == -1 {
		return ""
	}
	return content[start : start+end]
}
