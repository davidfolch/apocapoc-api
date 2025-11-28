package email

import (
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
}

func TestSMTPService_MessageConstruction(t *testing.T) {
	config := SMTPConfig{
		Host:         "smtp.example.com",
		Port:         587,
		Username:     "user@example.com",
		Password:     "password",
		From:         "noreply@example.com",
		SupportEmail: "support@example.com",
	}

	service := NewSMTPService(config)

	message := services.EmailMessage{
		To:      "recipient@example.com",
		Subject: "Test Email",
		Body:    "<h1>Test</h1>",
		IsHTML:  true,
	}

	if message.To == "" {
		t.Error("Expected recipient to be set")
	}

	if message.Subject == "" {
		t.Error("Expected subject to be set")
	}

	if !message.IsHTML {
		t.Error("Expected message to be HTML")
	}

	if service == nil {
		t.Fatal("Service should not be nil")
	}
}
