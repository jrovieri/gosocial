package mailer

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	FromName            = "GopherSocial"
	FromEmail           = "contact@gophersocial.com"
	MaxRetry            = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandBox bool) (int, error)
}

type MailSender struct {
	from   string
	apiKey string
	client *sendgrid.Client
}

func NewMailSender(apiKey, fromEmail string) (*MailSender, error) {

	if apiKey == "" {
		return &MailSender{}, errors.New("mail apiKey is required")
	}

	client := sendgrid.NewSendClient(apiKey)
	return &MailSender{
		from:   FromEmail,
		apiKey: apiKey,
		client: client,
	}, nil
}

func (m *MailSender) Send(templateFile, username, email string, data any, inSandbox bool) (int, error) {

	from := mail.NewEmail(FromName, m.from)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &inSandbox,
		},
	})

	for i := 0; i < MaxRetry; i++ {
		response, err := m.client.Send(message)

		if err != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return response.StatusCode, nil
	}
	return -1, fmt.Errorf("failed to send email after %d attempt, error: %v", MaxRetry, err)
}
