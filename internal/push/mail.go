package push

import (
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"note_feed/internal/config"
)

func SendEmail(cfg config.EmailConfig, subject, htmlBody string) error {
	if cfg.From == "" {
		cfg.From = cfg.Username
	}

	headers := make(map[string]string)
	headers["From"] = cfg.From
	headers["To"] = cfg.To
	headers["Subject"] = mime.BEncoding.Encode("utf-8", subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	if err := smtp.SendMail(addr, auth, cfg.From, []string{cfg.To}, []byte(msg.String())); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

func SendTestEmail(cfg config.EmailConfig) error {
	subject := "笔记推送系统 — 测试邮件"
	body := `<h2>✅ 配置验证成功</h2><p>如果你收到这封邮件，说明 SMTP 配置正确，邮件推送功能可用。</p>`
	return SendEmail(cfg, subject, body)
}
