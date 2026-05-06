package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration for email notifications.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// EmailNotifier sends alert notifications via email.
type EmailNotifier struct {
	cfg EmailConfig
}

// NewEmailNotifier creates a new EmailNotifier with the given config.
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{cfg: cfg}
}

// Send delivers the alert message to all configured recipients via SMTP.
func (e *EmailNotifier) Send(msg string) error {
	if len(e.cfg.To) == 0 {
		return fmt.Errorf("email notifier: no recipients configured")
	}

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}

	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: cronwatch alert\r\n\r\n%s",
		e.cfg.From,
		strings.Join(e.cfg.To, ", "),
		msg,
	)

	err := smtp.SendMail(addr, auth, e.cfg.From, e.cfg.To, []byte(body))
	if err != nil {
		return fmt.Errorf("email notifier: send failed: %w", err)
	}
	return nil
}
