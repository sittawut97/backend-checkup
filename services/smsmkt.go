package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SMSMKTClient struct {
	APIKey     string
	SecretKey  string
	ProjectKey string
	URL        string
}

type SMSMKTResponse struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
	Result struct {
		Token   string `json:"token"`
		RefCode string `json:"ref_code"`
	} `json:"result"`
}

func (s *SMSMKTClient) SendOTP(phone string) (string, error) {
	payload := map[string]interface{}{
		"project_key": s.ProjectKey,
		"phone":       phone,
		"ref_code":    "CHECKUP",
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", s.URL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", s.APIKey)
	req.Header.Set("secret_key", s.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("SMSMKT failed with status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var smsmktResp SMSMKTResponse
	if err := json.Unmarshal(respBody, &smsmktResp); err != nil {
		return "", err
	}

	if smsmktResp.Code != "000" {
		return "", fmt.Errorf("SMSMKT error: %s", smsmktResp.Detail)
	}

	return smsmktResp.Result.Token, nil
}

func (s *SMSMKTClient) ValidateOTP(token, otpCode, refCode string) error {
	payload := map[string]interface{}{
		"token":    token,
		"otp_code": otpCode,
		"ref_code": refCode,
	}

	body, _ := json.Marshal(payload)

	validateURL := "https://portal-otp.smsmkt.com/api/otp-validate"
	req, err := http.NewRequest("POST", validateURL, bytes.NewBuffer(body))
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
		return fmt.Errorf("SMSMKT validation failed with status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var smsmktResp SMSMKTResponse
	if err := json.Unmarshal(respBody, &smsmktResp); err != nil {
		return err
	}

	if smsmktResp.Code != "000" {
		return fmt.Errorf("invalid OTP: %s", smsmktResp.Detail)
	}

	return nil
}
