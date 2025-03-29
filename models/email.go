package models

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wneessen/go-mail"
)

const (
	DefaultSender = "support@imago.com"
)

type Email struct {
	From      string
	To        string
	Subject   string
	Plaintext string
	HTML      string
}

type EmailService struct {
	// DefaultSender is the default sender email address when none is provided in the request.
	// It is used also when an email is predetermined, like in the forgotten password
	// flow, where the email is not provided by the user.
	DefaultSender string

	client *mail.Client
}

func NewEmailService() (*EmailService, error) {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("email: %w", err)
	}

	options := []mail.Option{
		mail.WithPort(port),
		mail.WithUsername(os.Getenv("SMTP_USERNAME")),
		mail.WithPassword(os.Getenv("SMTP_PASSWORD")),
	}

	switch os.Getenv("ENV") {
	case "development":
		options = append(options, mail.WithTLSPortPolicy(mail.NoTLS))
		options = append(options, mail.WithSMTPAuth(mail.SMTPAuthPlain))
	default:
		options = append(options, mail.WithTLSPortPolicy(mail.TLSMandatory))
		options = append(options, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover))
	}

	client, err := mail.NewClient(os.Getenv("SMTP_SERVER"), options...)
	if err != nil {
		return nil, fmt.Errorf("email: %w", err)
	}

	return &EmailService{
		DefaultSender: DefaultSender,
		client:        client,
	}, nil
}

func (e *EmailService) Send(email Email) error {
	message := mail.NewMsg()

	if err := message.From(e.getFrom(email)); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	if err := message.To(email.To); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	message.Subject(email.Subject)

	switch {
	case email.Plaintext != "" && email.HTML != "":
		message.SetBodyString(mail.TypeTextHTML, email.HTML)
		message.AddAlternativeString(mail.TypeTextPlain, email.Plaintext)
	case email.Plaintext != "":
		message.SetBodyString(mail.TypeTextPlain, email.Plaintext)
	case email.HTML != "":
		message.SetBodyString(mail.TypeTextHTML, email.HTML)
	default:
		return fmt.Errorf("send: no body provided")
	}

	if err := e.client.DialAndSend(message); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (e *EmailService) SendForgottenPassword(to string, resetURL string) error {
	email := Email{
		To:        to,
		Subject:   "Imago - Reset password",
		Plaintext: fmt.Sprintf("Click the link to reset your password: %s", resetURL),
		HTML:      fmt.Sprintf("<p>Click the link to reset your password: </p><a href=\"%s\">%s</a>", resetURL, resetURL),
	}

	return e.Send(email)
}

func (e *EmailService) getFrom(email Email) string {
	if email.From != "" {
		return email.From
	}

	if e.DefaultSender != "" {
		return e.DefaultSender
	}

	return DefaultSender
}
