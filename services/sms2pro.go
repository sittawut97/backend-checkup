package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SMS2ProClient struct {
	APIKey string
	Client *http.Client
}

type SMS2ProOTPRequest struct {
	Recipient     string `json:"recipient"`
	SenderName    string `json:"sender_name"`
	RefCode       string `json:"ref_code,omitempty"`
	Digit         int    `json:"digit,omitempty"`
	Validity      int    `json:"validity,omitempty"`
	CustomMessage string `json:"custom_message,omitempty"`
}

type SMS2ProOTPResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Token  string `json:"token"`
	Msg    string `json:"msg"`
}

type SMS2ProVerifyRequest struct {
	Token   string `json:"token"`
	OTPCode string `json:"otp_code"`
	RefCode string `json:"ref_code,omitempty"`
}

type SMS2ProVerifyResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func NewSMS2ProClient(apiKey string) *SMS2ProClient {
	return &SMS2ProClient{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

func (s *SMS2ProClient) SendOTP(phone string) (string, error) {
	payload := SMS2ProOTPRequest{
		Recipient:  phone,
		SenderName: "CHECKUP",
		RefCode:    "",
		Digit:      6,
		Validity:   1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to marshal payload: %v\n", err)
		return "", err
	}

	fmt.Printf("[SMS2PRO] SendOTP Request - Phone: %s, Payload: %s\n", phone, string(payloadBytes))

	req, err := http.NewRequest("POST", "https://portal.sms2pro.com/sms-api/otp-sms/send", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to create request: %v\n", err)
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to send request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to read response body: %v\n", err)
		return "", err
	}

	fmt.Printf("[SMS2PRO] SendOTP Response Status: %d, Body: %s\n", resp.StatusCode, string(body))

	var response SMS2ProOTPResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("[SMS2PRO] Failed to parse response: %v\n", err)
		return "", err
	}

	if response.Code != 0 {
		fmt.Printf("[SMS2PRO] SendOTP failed - Code: %d, Status: %s, Msg: %s\n", response.Code, response.Status, response.Msg)
		return "", fmt.Errorf("SMS2PRO SendOTP failed: %s", response.Msg)
	}

	fmt.Printf("[SMS2PRO] SendOTP successful - Token: %s\n", response.Token)
	return response.Token, nil
}

func (s *SMS2ProClient) ValidateOTP(token string, otpCode string) error {
	payload := SMS2ProVerifyRequest{
		Token:   token,
		OTPCode: otpCode,
		RefCode: "",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to marshal verify payload: %v\n", err)
		return err
	}

	fmt.Printf("[SMS2PRO] ValidateOTP Request - Token: %s, OTP: %s, Payload: %s\n", token, otpCode, string(payloadBytes))

	req, err := http.NewRequest("POST", "https://portal.sms2pro.com/sms-api/otp-sms/verify", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to create verify request: %v\n", err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to send verify request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[SMS2PRO] Failed to read verify response body: %v\n", err)
		return err
	}

	fmt.Printf("[SMS2PRO] ValidateOTP Response Status: %d, Body: %s\n", resp.StatusCode, string(body))

	var response SMS2ProVerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("[SMS2PRO] Failed to parse verify response: %v\n", err)
		return err
	}

	if response.Code != 0 {
		fmt.Printf("[SMS2PRO] ValidateOTP failed - Code: %d, Status: %s, Msg: %s\n", response.Code, response.Status, response.Msg)
		return fmt.Errorf("SMS2PRO ValidateOTP failed: %s", response.Msg)
	}

	fmt.Printf("[SMS2PRO] ValidateOTP successful\n")
	return nil
}
