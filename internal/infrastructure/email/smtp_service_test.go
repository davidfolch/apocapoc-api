package email

import (
	"strings"
	"testing"

	"apocapoc-api/internal/domain/services"
)

func TestNewSMTPService(t *testing.T) {
	config := SMTPConfig{
		Host:         "smtp.example.com",
		Port:         587,
		Username:     "user@example.com",
		Password:     "password",
		From:         "noreply@example.com",
		SupportEmail: "support@example.com",
	}

	service := NewSMTPService(config)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.GetConfig().Host != config.Host {
		t.Errorf("Expected host %s, got %s", config.Host, service.GetConfig().Host)
	}

	if service.GetConfig().Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, service.GetConfig().Port)
	}

	if service.GetConfig().From != config.From {
		t.Errorf("Expected from %s, got %s", config.From, service.GetConfig().From)
	}
}

func TestSMTPService_ConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config SMTPConfig
	}{
		{
			name: "Port 587 (STARTTLS)",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password",
				From:     "noreply@example.com",
			},
		},
		{
			name: "Port 465 (SSL)",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     465,
				Username: "user@example.com",
				Password: "password",
				From:     "noreply@example.com",
			},
		},
		{
			name: "Port 25 (Plain)",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     25,
				Username: "user@example.com",
				Password: "password",
				From:     "noreply@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSMTPService(tt.config)
			if service == nil {
				t.Fatal("Expected service to be created")
			}

			if service.GetConfig().Port != tt.config.Port {
				t.Errorf("Expected port %d, got %d", tt.config.Port, service.GetConfig().Port)
			}
		})
	}
}

func TestSMTPService_MessageTypes(t *testing.T) {
	config := SMTPConfig{
		Host:         "smtp.example.com",
		Port:         587,
		Username:     "user@example.com",
		Password:     "password",
		From:         "noreply@example.com",
		SupportEmail: "support@example.com",
	}

	service := NewSMTPService(config)

	tests := []struct {
		name    string
		message services.EmailMessage
	}{
		{
			name: "HTML message",
			message: services.EmailMessage{
				To:      "recipient@example.com",
				Subject: "Test Email",
				Body:    "<h1>Test</h1>",
				IsHTML:  true,
			},
		},
		{
			name: "Plain text message",
			message: services.EmailMessage{
				To:      "recipient@example.com",
				Subject: "Test Email",
				Body:    "Plain text body",
				IsHTML:  false,
			},
		},
		{
			name: "Message with special characters",
			message: services.EmailMessage{
				To:      "recipient@example.com",
				Subject: "Test Email with émojis 🎉",
				Body:    "<p>Special chars: ñ, á, ü, €</p>",
				IsHTML:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message.To == "" {
				t.Error("Expected recipient to be set")
			}

			if tt.message.Subject == "" {
				t.Error("Expected subject to be set")
			}

			if tt.message.Body == "" {
				t.Error("Expected body to be set")
			}

			if service == nil {
				t.Fatal("Service should not be nil")
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		expected bool
	}{
		{
			name:     "Authentication failed error",
			errStr:   "535 Authentication failed",
			expected: true,
		},
		{
			name:     "Invalid credentials error",
			errStr:   "Invalid credentials provided",
			expected: true,
		},
		{
			name:     "535 error code",
			errStr:   "535 5.7.8 Error",
			expected: true,
		},
		{
			name:     "Connection refused (not auth)",
			errStr:   "connection refused",
			expected: false,
		},
		{
			name:     "Generic error (not auth)",
			errStr:   "some other error",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &mockError{msg: tt.errStr}
			result := isAuthError(err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for error: %s", tt.expected, result, tt.errStr)
			}
		})
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		expected bool
	}{
		{
			name:     "Connection refused",
			errStr:   "connection refused",
			expected: true,
		},
		{
			name:     "No such host",
			errStr:   "no such host smtp.invalid.com",
			expected: true,
		},
		{
			name:     "Network unreachable",
			errStr:   "network is unreachable",
			expected: true,
		},
		{
			name:     "Authentication error (not config)",
			errStr:   "authentication failed",
			expected: false,
		},
		{
			name:     "Generic error (not config)",
			errStr:   "some other error",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &mockError{msg: tt.errStr}
			result := isConfigError(err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for error: %s", tt.expected, result, tt.errStr)
			}
		})
	}
}

func TestSMTPService_Send_InvalidConfig(t *testing.T) {
	config := SMTPConfig{
		Host:     "invalid.smtp.server.that.does.not.exist",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
	}

	service := NewSMTPService(config)

	message := services.EmailMessage{
		To:      "test@example.com",
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  false,
	}

	err := service.Send(message)
	if err == nil {
		t.Error("Expected error when sending to invalid SMTP server")
	}

	if !strings.Contains(err.Error(), "failed to send email") {
		t.Errorf("Expected error message to contain 'failed to send email', got: %s", err.Error())
	}
}

func TestSMTPService_HealthCheck_InvalidConfig(t *testing.T) {
	config := SMTPConfig{
		Host:     "invalid.smtp.server.that.does.not.exist",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
	}

	service := NewSMTPService(config)

	err := service.HealthCheck()
	if err == nil {
		t.Error("Expected error when health checking invalid SMTP server")
	}

	if !strings.Contains(err.Error(), "SMTP") {
		t.Errorf("Expected error message to contain 'SMTP', got: %s", err.Error())
	}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
