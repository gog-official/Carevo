package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Sender struct {
	From   string
	APIKey string
}

func NewSender(from, apiKey string) *Sender {
	return &Sender{
		From:   from,
		APIKey: apiKey,
	}
}

func (s *Sender) Send(to, subject, body string) error {
	if s.APIKey == "" {
		return fmt.Errorf("SendGrid API key not set (SENDGRID_API_KEY)")
	}

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": to}}},
		},
		"from":    map[string]string{"email": s.From},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": body},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		msg := ""
		if len(errResp.Errors) > 0 {
			msg = errResp.Errors[0].Message
		}
		return fmt.Errorf("sendgrid api status %d: %s", resp.StatusCode, msg)
	}

	log.Printf("SendGrid success for %s", to)
	return nil
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
