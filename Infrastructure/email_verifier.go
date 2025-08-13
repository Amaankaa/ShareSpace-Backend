package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type EmailListVerifyVerifier struct {
	APIKey string
}

type emailListVerifyResponse struct {
	Email       string `json:"email"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	Reason      string `json:"reason"`
	Disposable  bool   `json:"disposable"`
	AcceptAll   bool   `json:"accept_all"`
	MXRecord    bool   `json:"mx_record"`
	SMTPCheck   bool   `json:"smtp_check"`
	Deliverable bool   `json:"deliverable"`
}

func NewEmailListVerifyVerifier() (*EmailListVerifyVerifier, error) {
	apiKey := os.Getenv("EMAILLISTVERIFY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EMAILLISTVERIFY_API_KEY is not set")
	}
	return &EmailListVerifyVerifier{APIKey: apiKey}, nil
}

func (e *EmailListVerifyVerifier) IsRealEmail(email string) (bool, error) {
	// EmailListVerify API endpoint - try the correct endpoint
	baseURL := "https://apps.emaillistverify.com/api/verifyEmail"

	// Create URL with parameters - EmailListVerify uses 'secret' and 'email' parameters
	params := url.Values{}
	params.Add("secret", e.APIKey)
	params.Add("email", email)

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())


	resp, err := http.Get(requestURL)
	if err != nil {
		return false, fmt.Errorf("failed to make request to EmailListVerify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("EmailListVerify API returned status %d", resp.StatusCode)
	}

	// Read the response body first to debug
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Check if response starts with '{' (JSON) or is plain text
	if body[0] != '{' {
		responseText := string(body)
		// Handle common plain text responses
		if responseText == "ok" {
			return true, nil
		}
		if responseText == "invalid" || responseText == "error" {
			return false, nil
		}
		return false, fmt.Errorf("EmailListVerify returned non-JSON response: %s", responseText)
	}

	var result emailListVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to decode EmailListVerify JSON response: %w. Raw response: %s", err, string(body))
	}

	isValid := result.Status == "ok" &&
		(result.Result == "deliverable" || result.Result == "risky") &&
		result.MXRecord &&
		!result.Disposable

	return isValid, nil
}
