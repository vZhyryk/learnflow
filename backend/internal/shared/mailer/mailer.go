package mailer

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"learnflow_backend/internal/shared/validator"
	"time"

	"gopkg.in/mail.v2"
)

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New returns a Mailer configured with the given SMTP credentials.
func New(port int, host, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 1 * time.Minute
	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// CCuser holds the recipient's email and display name for outgoing emails.
type CCuser struct {
	Mail     string `json:"email"`
	Username string `json:"username"`
}

// Send renders the given email template and delivers it to ccUser via SMTP in a single attempt.
// Callers that need retry semantics (e.g. EmailWorker) wrap this call with their own retry logic.
func (m Mailer) Send(templateFile string, data any, ccUser CCuser, attachmentList []string) error {
	if ccUser.Mail == "" || !validator.MatchesEmail(ccUser.Mail) {
		return fmt.Errorf("mailer.Send: invalid recipient address")
	}

	subject, plainBody, htmlBody, err := renderEmail(templateFile, data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", ccUser.Mail)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", plainBody)
	msg.AddAlternative("text/html", htmlBody)

	for _, path := range attachmentList {
		msg.Attach(path)
	}

	m.dialer.TLSConfig = &tls.Config{
		ServerName: m.dialer.Host,
		MinVersion: tls.VersionTLS12,
	}

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailer.Send: %w", err)
	}
	return nil
}

//go:embed templates
var templateFS embed.FS

func renderEmail(templateFile string, data any) (subject, plainBody, htmlBody string, err error) {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return "", "", "", fmt.Errorf("mailer.renderEmail parse template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "subject", data); err != nil {
		return "", "", "", fmt.Errorf("mailer.renderEmail execute subject: %w", err)
	}
	subject = buf.String()

	buf.Reset()
	if err = tmpl.ExecuteTemplate(&buf, "plainBody", data); err != nil {
		return "", "", "", fmt.Errorf("mailer.renderEmail execute plainBody: %w", err)
	}
	plainBody = buf.String()

	buf.Reset()
	if err = tmpl.ExecuteTemplate(&buf, "htmlBody", data); err != nil {
		return "", "", "", fmt.Errorf("mailer.renderEmail execute htmlBody: %w", err)
	}
	htmlBody = buf.String()

	return subject, plainBody, htmlBody, nil
}
