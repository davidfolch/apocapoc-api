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
}
