package infrastructure

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type BrevoEmailSender struct{}

func NewBrevoEmailSender() *BrevoEmailSender {
	return &BrevoEmailSender{}
}

func (b *BrevoEmailSender) SendEmail(to, subject, content string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	fromEmail := os.Getenv("FROM_EMAIL")
	fromName := os.Getenv("FROM_NAME")

	if apiKey == "" || fromEmail == "" || fromName == "" {
		return errors.New("missing Brevo config in environment")
	}

	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  fromName,
			"email": fromEmail,
		},
		"to": []map[string]string{
			{"email": to},
		},
		"subject":     subject,
		"htmlContent": "<p>" + content + "</p>",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to send email: " + resp.Status)
	}

	return nil
}
