package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
		Port:     "587",
	}
}

func (s *Sender) Send(to, subject, body string) error {
	if s.Password != "" {
		log.Printf("trying SMTP %s:%s...", s.Host, s.Port)
		err := s.sendSMTP(to, subject, body)
		if err == nil {
			return nil
		}
		log.Printf("SMTP failed: %v", err)
	}
	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		log.Printf("trying Resend API...")
		return s.sendResend(apiKey, to, subject, body)
	}
	return fmt.Errorf("no email provider could send (set SMTP_PASSWORD or RESEND_API_KEY)")
}

func (s *Sender) sendResend(apiKey, to, subject, body string) error {
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "Carevo <onboarding@resend.dev>"
	}
	payload := map[string]interface{}{
		"from":    from,
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

	log.Printf("Resend API success for %s", to)
	return nil
}

func (s *Sender) sendSMTP(to, subject, body string) error {
	if s.Password == "" {
		return fmt.Errorf("SMTP password not set (SMTP_PASSWORD)")
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", s.From, to, subject, body))

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	log.Printf("SMTP dialing %s...", addr)
	err := sendMailTimeout(addr, auth, s.From, []string{to}, msg, 15*time.Second)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	log.Printf("SMTP success for %s", to)
	return nil
}

func sendMailTimeout(addr string, a smtp.Auth, from string, to []string, msg []byte, timeout time.Duration) error {
	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		err := smtp.SendMail(addr, a, from, to, msg)
		ch <- result{err}
	}()
	select {
	case r := <-ch:
		return r.err
	case <-time.After(timeout):
		return fmt.Errorf("connection timed out after %v", timeout)
	}
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
