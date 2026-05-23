package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"
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
		Port:     "465",
	}
}

func (s *Sender) Send(to, subject, body string) error {
	if s.Password == "" {
		return fmt.Errorf("SMTP password not set (SMTP_PASSWORD)")
	}
	log.Printf("trying SMTP %s:%s...", s.Host, s.Port)
	return s.sendSMTP(to, subject, body)
}

func (s *Sender) sendSMTP(to, subject, body string) error {
	if s.Password == "" {
		return fmt.Errorf("SMTP password not set (SMTP_PASSWORD)")
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", s.From, to, subject, body))

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	log.Printf("SMTP dialing %s...", addr)

	addrs, lookupErr := net.LookupHost(s.Host)
	log.Printf("resolved %s -> %v (err: %v)", s.Host, addrs, lookupErr)

	var conn net.Conn
	var dialErr error

	d := net.Dialer{Timeout: 10 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, dialErr = d.DialContext(ctx, "tcp", addr)
	if dialErr != nil {
		return fmt.Errorf("tcp dial: %w", dialErr)
	}
	defer conn.Close()

	var client *smtp.Client
	var err error

	if s.Port == "465" {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: s.Host})
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("tls handshake: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, s.Host)
		if err != nil {
			return fmt.Errorf("new tls client: %w", err)
		}
	} else {
		client, err = smtp.NewClient(conn, s.Host)
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}
		if err := client.StartTLS(&tls.Config{ServerName: s.Host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err := client.Mail(s.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return client.Quit()
}

func (s *Sender) SendVerificationCode(to, code string) error {
	subject := "Your Carevo verification code"
	body := fmt.Sprintf(`Welcome to Carevo!

Your verification code is: %s

This code expires in 10 minutes.

If you didn't create an account, you can ignore this email.

- Carevo Team`, code)

	log.Printf("Sending verification code %s to %s...", code, to)
	if err := s.Send(to, subject, body); err != nil {
		log.Printf("EMAIL FAILED to %s (code %s): %v", to, code, err)
		log.Printf("--- VERIFICATION CODE for %s (email failed): %s ---", to, code)
	} else {
		log.Printf("EMAIL SENT to %s with code %s", to, code)
	}
	return nil
}

func (s *Sender) SendSuggestionApproved(to, title string) error {
	subject := fmt.Sprintf("Your career suggestion \"%s\" has been approved!", title)
	body := fmt.Sprintf(`Hi there,

Your suggested career "%s" has been approved by our team!

It will be added to Carevo soon. Thank you for helping us grow our career database.

- Carevo Team`, title)

	log.Printf("Sending approval to %s...", to)
	if err := s.Send(to, subject, body); err != nil {
		log.Printf("EMAIL FAILED approval to %s: %v", to, err)
	}
	return nil
}

func (s *Sender) SendSuggestionRejected(to, title string) error {
	subject := fmt.Sprintf("Update on your career suggestion \"%s\"", title)
	body := fmt.Sprintf(`Hi there,

Regarding your suggested career "%s" — after review, we've decided not to add it to Carevo at this time.

Thank you for your contribution!

- Carevo Team`, title)

	log.Printf("Sending rejection to %s...", to)
	if err := s.Send(to, subject, body); err != nil {
		log.Printf("EMAIL FAILED rejection to %s: %v", to, err)
	}
	return nil
}
