package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SMSMKTClient struct {
	APIKey    string
	SecretKey string
	URL       string
}

func (s *SMSMKTClient) SendSMS(phone string, message string) error {
	payload := map[string]interface{}{
		"project_key": s.APIKey,
		"phone":       phone,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", s.URL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", s.APIKey)
	req.Header.Set("secret_key", s.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("SMSMKT failed with status %d", resp.StatusCode)
	}

	return nil
}
