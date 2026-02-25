package mail

import (
	"fmt"
	"net"
	"net/smtp"
)

// Mailer sends emails via SMTP.
type Mailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Send sends an email to the given recipient with the specified subject and body.
func (m *Mailer) Send(to, subject, body string) error {
	addr := net.JoinHostPort(m.Host, m.Port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", m.From, to, subject, body)

	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}

	return nil
}
