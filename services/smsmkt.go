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
		Status  bool   `json:"status"`
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
	fmt.Printf("[SMSMKT] ValidateOTP Request - Token: %s, OTP: %s, RefCode: %s\n", token, otpCode, refCode)

	validateURL := "https://portal-otp.smsmkt.com/api/otp-validate"
	req, err := http.NewRequest("POST", validateURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("[SMSMKT] Request creation error: %v\n", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", s.APIKey)
	req.Header.Set("secret_key", s.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[SMSMKT] HTTP request error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("[SMSMKT] Response Status: %d\n", resp.StatusCode)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("SMSMKT validation failed with status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[SMSMKT] Body read error: %v\n", err)
		return err
	}

	fmt.Printf("[SMSMKT] Response Body: %s\n", string(respBody))

	var smsmktResp SMSMKTResponse
	if err := json.Unmarshal(respBody, &smsmktResp); err != nil {
		fmt.Printf("[SMSMKT] JSON unmarshal error: %v\n", err)
		return err
	}

	fmt.Printf("[SMSMKT] Parsed Response - Code: %s, Detail: %s, Status: %v\n", smsmktResp.Code, smsmktResp.Detail, smsmktResp.Result.Status)

	if smsmktResp.Code != "000" {
		return fmt.Errorf("invalid OTP: %s", smsmktResp.Detail)
	}

	// Check result.status for actual validation result
	if !smsmktResp.Result.Status {
		return fmt.Errorf("invalid OTP code")
	}

	return nil
}
