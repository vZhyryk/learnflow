package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"learnflow_backend/internal/infrastructure/validator"
	"os"
	"time"

	"gopkg.in/mail.v2"
)

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New returns a Mailer configured with the given SMTP credentials.
func New(port int, host, username, password, sender string) Mailer {
	// Initialize a new mail.Dialer instance with the given SMTP server settings. We also configure this to use a 5-second timeout whenever we send an email.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 1 * time.Minute
	// Return a Mailer instance containing the dialer and sender information.
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// CCuser holds the recipient's email and display name for outgoing emails.
type CCuser struct {
	Mail     string `json:"email"`
	Username string `json:"username"`
}

// Send renders the given email template and delivers it to ccUser via SMTP (3 attempts).
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

	m.dialer.TLSConfig = &tls.Config{}
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("mailer.Send: %w", err)
}

func renderEmail(templateFile string, data any) (subject, plainBody, htmlBody string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("mailer.renderEmail get pwd: %w", err)
	}

	tmpl, err := template.New("email").ParseFiles(dir + templateFile)
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
