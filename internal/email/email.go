package email

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	From     string
	Password string
	Host     string
	Port     string
}

func NewSender(from, password string) *Sender {
	return &Sender{
		From:     from,
		Password: password,
		Host:     "smtp.gmail.com",
		Port:     "587",
	}
}

func (s *Sender) Send(to, subject, body string) error {
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", s.From, to, subject, body))

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}

func (s *Sender) SendVerificationCode(to, code string) error {
	subject := "Your Carevo verification code"
	body := fmt.Sprintf(`Welcome to Carevo!

Your verification code is: %s

This code expires in 10 minutes.

If you didn't create an account, you can ignore this email.

- Carevo Team`, code)
	return s.Send(to, subject, body)
}

func (s *Sender) SendSuggestionApproved(to, title string) error {
	subject := fmt.Sprintf("Your career suggestion \"%s\" has been approved!", title)
	body := fmt.Sprintf(`Hi there,

Your suggested career "%s" has been approved by our team!

It will be added to Carevo soon. Thank you for helping us grow our career database.

- Carevo Team`, title)
	return s.Send(to, subject, body)
}

func (s *Sender) SendSuggestionRejected(to, title string) error {
	subject := fmt.Sprintf("Update on your career suggestion \"%s\"", title)
	body := fmt.Sprintf(`Hi there,

Regarding your suggested career "%s" — after review, we've decided not to add it to Carevo at this time.

Thank you for your contribution!

- Carevo Team`, title)
	return s.Send(to, subject, body)
}
