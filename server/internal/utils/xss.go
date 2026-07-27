package utils

import (
	"regexp"
	"strings"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// SanitizeString 清理字符串，移除 HTML 标签并转义特殊字符，防止 XSS
func SanitizeString(input string) string {
	// 移除潜在的 HTML 标签
	cleaned := htmlTagPattern.ReplaceAllString(input, "")

	// 转义 HTML 特殊字符
	cleaned = strings.ReplaceAll(cleaned, "&", "&amp;")
	cleaned = strings.ReplaceAll(cleaned, "<", "&lt;")
	cleaned = strings.ReplaceAll(cleaned, ">", "&gt;")
	cleaned = strings.ReplaceAll(cleaned, "\"", "&quot;")
	cleaned = strings.ReplaceAll(cleaned, "'", "&#39;")

	return strings.TrimSpace(cleaned)
}
