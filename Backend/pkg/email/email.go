package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Config struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type Service struct {
	config Config
}

func NewService(config Config) *Service {
	return &Service{config: config}
}

func (s *Service) SendVerificationEmail(to, otp string) error {
	subject := "Verify your email"
	body := fmt.Sprintf(`
Hello,

Your verification code is: %s

This code will expire in 15 minutes.

If you didn't request this, please ignore this email.

Best regards,
TaskFlow Team
`, otp)

	return s.send(to, subject, body)
}

func (s *Service) SendPasswordResetEmail(to, token, frontendURL string) error {
	subject := "Reset your password"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)
	body := fmt.Sprintf(`
Hello,

You requested to reset your password. Click the link below to proceed:

%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email.

Best regards,
TaskFlow Team
`, resetLink)

	return s.send(to, subject, body)
}

func (s *Service) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if s.config.User != "" && s.config.Pass != "" {
		auth = smtp.PlainAuth("", s.config.User, s.config.Pass, s.config.Host)
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg.String()))
}
