package email

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
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
	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		return s.sendResend(apiKey, to, subject, body)
	}
	if s.Password != "" {
		return s.sendSMTP(to, subject, body)
	}
	return fmt.Errorf("no email provider configured (set RESEND_API_KEY or SMTP_PASSWORD)")
}

func (s *Sender) sendResend(apiKey, to, subject, body string) error {
	payload := map[string]interface{}{
		"from":    s.From,
		"to":      []string{to},
		"subject": subject,
		"text":    body,
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("resend api status %d: %s", resp.StatusCode, errResp.Message)
	}

	return nil
}

func (s *Sender) sendSMTP(to, subject, body string) error {
	if s.Password == "" {
		return fmt.Errorf("SMTP password not configured")
	}

	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n", s.From, to, subject)
	msg := []byte(header + body)

	tlsConfig := &tls.Config{ServerName: s.Host}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", s.Host+":"+s.Port, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
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

	if err := s.Send(to, subject, body); err != nil {
		log.Printf("email send failed to %s (code %s logged as fallback): %v", to, code, err)
	}
	log.Printf("--- VERIFICATION CODE for %s: %s ---", to, code)
	return nil
}

func (s *Sender) SendSuggestionApproved(to, title string) error {
	subject := fmt.Sprintf("Your career suggestion \"%s\" has been approved!", title)
	body := fmt.Sprintf(`Hi there,

Your suggested career "%s" has been approved by our team!

It will be added to Carevo soon. Thank you for helping us grow our career database.

- Carevo Team`, title)

	if err := s.Send(to, subject, body); err != nil {
		log.Printf("email send failed to %s: %v", to, err)
	}
	return nil
}

func (s *Sender) SendSuggestionRejected(to, title string) error {
	subject := fmt.Sprintf("Update on your career suggestion \"%s\"", title)
	body := fmt.Sprintf(`Hi there,

Regarding your suggested career "%s" — after review, we've decided not to add it to Carevo at this time.

Thank you for your contribution!

- Carevo Team`, title)

	if err := s.Send(to, subject, body); err != nil {
		log.Printf("email send failed to %s: %v", to, err)
	}
	return nil
}
