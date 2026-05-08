package service

import (
	"fmt"
	"net/smtp"

	"github.com/chrpnnkv/SmartBattle/internal/config"
)

type EmailService interface {
	SendPasswordResetEmail(toEmail, token string) error
}

type SMTPEmailService struct {
	cfg *config.Config
}

func NewSMTPEmailService(cfg *config.Config) *SMTPEmailService {
	return &SMTPEmailService{cfg: cfg}
}

func (s *SMTPEmailService) SendPasswordResetEmail(toEmail, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, token)

	headerFrom := fmt.Sprintf("From: %s\r\n", s.cfg.SMTPFrom)
	headerTo := fmt.Sprintf("To: %s\r\n", toEmail)
	subject := "Subject: Сброс пароля в SmartBattle\r\n"
	mime := "MIME-version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"

	body := fmt.Sprintf("Здравствуйте!\n\nВы запросили сброс пароля. Перейдите по ссылке, чтобы установить новый пароль:\n%s\n\nЕсли вы не запрашивали сброс, просто проигнорируйте это письмо.\n\nКоманда SmartBattle", resetLink)

	msg := []byte(headerFrom + headerTo + subject + mime + body)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)

	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{toEmail}, msg)
}
