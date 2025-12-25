package domain

import "strings"

// ExtractJSON 从可能包含 Markdown 代码块的字符串中提取 JSON
func ExtractJSON(content string) string {
	// 尝试提取 ```json 代码块
	if jsonStart := findCodeBlock(content, "```json"); jsonStart != "" {
		return strings.TrimSpace(jsonStart)
	}
	// 尝试提取普通 ``` 代码块
	if codeStart := findCodeBlock(content, "```"); codeStart != "" {
		return strings.TrimSpace(codeStart)
	}
	return strings.TrimSpace(content)
}

// findCodeBlock 查找并提取代码块内容
func findCodeBlock(content, marker string) string {
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	start += len(marker)

	// 跳过可能的换行符
	if start < len(content) && content[start] == '\n' {
		start++
	}

	end := strings.Index(content[start:], "```")
	if end == -1 {
		return ""
	}
	return content[start : start+end]
}
