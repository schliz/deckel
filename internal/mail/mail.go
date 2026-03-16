package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const dialTimeout = 10 * time.Second

// Mailer sends emails via SMTP.
type Mailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
}

// formatFrom returns the From header value, using RFC 5322 format when FromName is set.
func (m *Mailer) formatFrom() string {
	if m.FromName == "" {
		return m.From
	}
	return (&mail.Address{Name: m.FromName, Address: m.From}).String()
}

// Send sends an email to the given recipient with the specified subject and body.
func (m *Mailer) Send(to, subject, body string) error {
	addr := net.JoinHostPort(m.Host, m.Port)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.formatFrom(), to, subject, body,
	)

	tlsConfig := &tls.Config{ServerName: m.Host}
	dialer := &net.Dialer{Timeout: dialTimeout}

	var conn net.Conn
	var err error

	// Port 465 uses implicit TLS (SMTPS); all others use STARTTLS.
	if m.Port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	// STARTTLS for non-465 ports that advertise it (port 25 may not support TLS).
	if m.Port != "465" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	// Authenticate if credentials are provided.
	if m.Username != "" {
		auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	// Send the message.
	if err := client.Mail(m.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to %s: %w", to, err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := strings.NewReader(msg).WriteTo(w); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}
