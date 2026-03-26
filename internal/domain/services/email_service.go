package services

type EmailMessage struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

type EmailService interface {
	Send(message EmailMessage) error
	HealthCheck() error
	IsEnabled() bool
}

type NoOpEmailService struct{}

func (n *NoOpEmailService) Send(_ EmailMessage) error { return nil }
func (n *NoOpEmailService) HealthCheck() error        { return nil }
func (n *NoOpEmailService) IsEnabled() bool           { return false }
