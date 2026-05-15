package push

import (
	"fmt"
	"strings"
	"time"

	"note_feed/internal/config"
	"note_feed/internal/search"
)

func BuildDigestHTML(articles []search.Article, dateStr string, categories []string) string {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Group articles by category
	byCategory := make(map[string][]search.Article)
	categoryOrder := categories
	for _, a := range articles {
		cat := a.Category
		if cat == "" {
			cat = "其他"
		}
		byCategory[cat] = append(byCategory[cat], a)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif; max-width: 640px; margin: 0 auto; padding: 20px;">
	<div style="border-bottom: 2px solid #4361ee; padding-bottom: 12px; margin-bottom: 24px;">
		<h1 style="font-size: 20px; color: #1a1a2e; margin: 0;">📚 今日学习推送</h1>
		<p style="color: #6c757d; margin: 4px 0 0 0; font-size: 14px;">%s</p>
	</div>`, dateStr))

	for _, cat := range categoryOrder {
		items, ok := byCategory[cat]
		if !ok || len(items) == 0 {
			continue
		}

		body.WriteString(fmt.Sprintf(`
	<div style="margin-bottom: 24px;">
		<h2 style="font-size: 15px; color: #4361ee; margin: 0 0 12px 0; padding-bottom: 4px; border-bottom: 1px solid #e9ecef;">
			━━ %s ━━
		</h2>`, cat))

		for _, a := range items {
			body.WriteString(fmt.Sprintf(`
		<div style="margin-bottom: 12px; padding: 12px; background: #f8f9fa; border-radius: 6px;">
			<a href="%s" style="font-size: 14px; font-weight: 600; color: #1a1a2e; text-decoration: none; display: block; margin-bottom: 4px;">%s</a>
			<p style="font-size: 13px; color: #6c757d; margin: 0;">%s</p>
		</div>`, a.URL, a.Title, a.Description))
		}

		body.WriteString(`	</div>`)
	}

	body.WriteString(`
	<div style="border-top: 1px solid #e9ecef; padding-top: 12px; margin-top: 24px;">
		<p style="font-size: 12px; color: #adb5bd; margin: 0;">由笔记推送系统自动生成 · 祝你学习愉快</p>
	</div>
</body>
</html>`)

	return body.String()
}

func SendDigest(cfg config.Config, articles []search.Article, categories []string, dateStr string) error {
	subject := fmt.Sprintf("📚 今日学习推送 (%s)", dateStr)
	htmlBody := BuildDigestHTML(articles, dateStr, categories)
	return SendEmail(cfg.Push.Email, subject, htmlBody)
}
