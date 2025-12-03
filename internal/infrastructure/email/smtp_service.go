package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"apocapoc-api/internal/domain/services"

	"gopkg.in/mail.v2"
)

type SMTPConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	From         string
	SupportEmail string
}

type SMTPService struct {
	config SMTPConfig
}

func NewSMTPService(config SMTPConfig) *SMTPService {
	return &SMTPService{
		config: config,
	}
}

func (s *SMTPService) Send(message services.EmailMessage) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)

	if message.IsHTML {
		m.SetBody("text/html", message.Body)
	} else {
		m.SetBody("text/plain", message.Body)
	}

	dialer := mail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	dialer.TLSConfig = &tls.Config{
		ServerName: s.config.Host,
	}

	// Use SSL from start for port 465, STARTTLS for other ports
	if s.config.Port == 465 {
		dialer.SSL = true
	}

	if err := s.sendWithRetry(dialer, m); err != nil {
		log.Printf("[EMAIL] status=failed to=%s subject=%q error=%q", message.To, message.Subject, err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL] status=sent to=%s subject=%q", message.To, message.Subject)
	return nil
}

func (s *SMTPService) sendWithRetry(dialer *mail.Dialer, message *mail.Message) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := dialer.DialAndSend(message); err == nil {
			return nil
		} else {
			lastErr = err

			if isAuthError(err) {
				return fmt.Errorf("SMTP authentication failed. Please check your SMTP credentials (username, password, and from address)")
			}

			if isConfigError(err) {
				return fmt.Errorf("SMTP configuration error: %w", err)
			}

			if i < maxRetries-1 {
				time.Sleep(time.Second * time.Duration(i+1))
			}
		}
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, lastErr)
}

func isAuthError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "535") ||
		strings.Contains(errStr, "invalid credentials")
}

func isConfigError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network is unreachable")
}

func (s *SMTPService) GetConfig() SMTPConfig {
	return s.config
}

func (s *SMTPService) HealthCheck() error {
	dialer := mail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	dialer.TLSConfig = &tls.Config{
		ServerName: s.config.Host,
	}

	if s.config.Port == 465 {
		dialer.SSL = true
	}

	smtpCloser, err := dialer.Dial()
	if err != nil {
		if isAuthError(err) {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
		if isConfigError(err) {
			return fmt.Errorf("SMTP connection failed: %w", err)
		}
		return fmt.Errorf("SMTP error: %w", err)
	}
	defer smtpCloser.Close()

	return nil
}
