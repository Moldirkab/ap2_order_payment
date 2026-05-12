package provider

import (
	"net/smtp"
	"os"
)

type SMTPSender struct{}

func NewSMTPSender() EmailSender {
	return &SMTPSender{}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	msg := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	auth := smtp.PlainAuth("", user, pass, host)

	return smtp.SendMail(
		host+":"+port,
		auth,
		user,
		[]string{to},
		msg,
	)
}
