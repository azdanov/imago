package models

import (
	"errors"
	"fmt"

	"github.com/azdanov/imago/config"
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

func NewEmailService(cnf *config.Config) (*EmailService, error) {
	options := []mail.Option{
		mail.WithPort(cnf.SMTP.Port),
		mail.WithUsername(cnf.SMTP.Username),
		mail.WithPassword(cnf.SMTP.Password),
	}

	if cnf.SMTP.SSLMode {
		options = append(options, mail.WithTLSPortPolicy(mail.TLSMandatory))
		options = append(options, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover))
	} else {
		options = append(options, mail.WithTLSPortPolicy(mail.NoTLS))
		options = append(options, mail.WithSMTPAuth(mail.SMTPAuthPlain))
	}

	client, err := mail.NewClient(cnf.SMTP.Host, options...)
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
		return errors.New("send: no body provided")
	}

	if err := e.client.DialAndSend(message); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (e *EmailService) SendResetPassword(to string, resetURL string) error {
	email := Email{
		To:        to,
		Subject:   "Imago - Reset password",
		Plaintext: fmt.Sprintf("Click the link to reset your password: %s", resetURL),
		HTML: fmt.Sprintf("<p>Click the link to reset your password: </p><a href=\"%s\">%s</a>",
			resetURL, resetURL),
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
