package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/middleware"
	"github.com/sittawut/backend-appointment/models"
	"github.com/sittawut/backend-appointment/services"
	"github.com/supabase-community/postgrest-go"
	supa "github.com/supabase-community/supabase-go"
	"golang.org/x/crypto/bcrypt"
)

// Ensure postgrest is imported for OrderOpts
var _ = postgrest.OrderOpts{}

// OTPHandler handles OTP operations
type OTPHandler struct {
	sms      services.SMSClient
	supabase *supa.Client
	config   *config.Config
}

// NewOTPHandler creates a new OTP handler
func NewOTPHandler(supabase *supa.Client, cfg *config.Config, smsClient services.SMSClient) *OTPHandler {
	return &OTPHandler{
		supabase: supabase,
		config:   cfg,
		sms:      smsClient,
	}
}

// RequestOTPRequest represents the request body
type RequestOTPRequest struct {
	Phone string `json:"phone" binding:"required,len=10,numeric"`
}

// VerifyOTPRequest represents the request body
type VerifyOTPRequest struct {
	Phone   string `json:"phone" binding:"required,len=10,numeric"`
	OTPCode string `json:"otp_code" binding:"required,len=6,numeric"`
}

// OTPRecord represents OTP data from database
type OTPRecord struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	OTPHash   string    `json:"otp_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `json:"is_used"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// RequestOTP generates and sends OTP
func (h *OTPHandler) RequestOTP(c *gin.Context) {
	var req RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid phone number format",
		})
		return
	}

	fmt.Printf("[OTP] RequestOTP - Phone: %s\n", req.Phone)

	// Step 1: Verify user exists and get user_id
	user, err := h.getUserByPhone(req.Phone)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Phone number not found in system",
		})
		return
	}

	userID := user.ID

	// Step 2: Check rate limit
	allowed, err := h.checkRateLimit(userID, "request_otp")
	if err != nil {
		fmt.Printf("[OTP] Rate limit check error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	if !allowed {
		c.JSON(http.StatusTooManyRequests, models.Response{
			Success: false,
			Error:   "Too many OTP requests. Please try again later.",
		})
		return
	}

	// Step 3: Invalidate previous OTPs
	h.invalidatePreviousOTPs(req.Phone)

	// Step 4: Generate OTP
	otp := h.generateOTP()
	fmt.Printf("[OTP] Generated OTP for %s: %s\n", req.Phone, otp)

	// Step 5: Hash OTP
	otpHash, err := h.hashOTP(otp)
	if err != nil {
		fmt.Printf("[OTP] Hash error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to process OTP",
		})
		return
	}

	// Step 6: Save to database
	expiresAt := time.Now().Add(1 * time.Minute)
	otpID, err := h.saveOTP(userID, req.Phone, otpHash, expiresAt)
	if err != nil {
		fmt.Printf("[OTP] Save error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to save OTP",
		})
		return
	}

	// Step 7: Send SMS via THSMS
	// For THSMS, we need to send the OTP code directly
	// Cast to THSMSClientImpl to access SendOTPWithCode
	var messageID string
	if thsmsClient, ok := h.sms.(*services.THSMSClientImpl); ok {
		var err error
		messageID, err = thsmsClient.SendOTPWithCode(req.Phone, otp)
		if err != nil {
			fmt.Printf("[OTP] SMS send error: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Error:   "Failed to send OTP. Please try again.",
			})
			return
		}
	} else {
		// Fallback for other SMS providers
		var err error
		messageID, err = h.sms.SendOTP(req.Phone)
		if err != nil {
			fmt.Printf("[OTP] SMS send error: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Error:   "Failed to send OTP. Please try again.",
			})
			return
		}
	}

	// Step 8: Update rate limit
	h.updateRateLimit(userID, "request_otp")

	// Step 9: Log success
	h.logAudit(req.Phone, "request_otp", "success", fmt.Sprintf("message_id: %s", messageID), c.ClientIP())

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "OTP sent successfully",
		Data: map[string]interface{}{
			"expires_in": 60,
			"otp_id":     otpID,
		},
	})
}

// VerifyOTP verifies the OTP and logs in the user
func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	fmt.Printf("[OTP] VerifyOTP - Phone: %s\n", req.Phone)

	// Step 1: Get user by phone
	user, err := h.getUserByPhone(req.Phone)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	userID := user.ID

	// Step 2: Check rate limit
	allowed, err := h.checkRateLimit(userID, "verify_otp")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	if !allowed {
		c.JSON(http.StatusTooManyRequests, models.Response{
			Success: false,
			Error:   "Too many verification attempts. Please request a new OTP.",
		})
		return
	}

	// Step 3: Get latest OTP from database
	otpRecord, err := h.getLatestOTP(req.Phone)
	if err != nil || otpRecord == nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "No valid OTP found. Please request a new one.",
		})
		return
	}

	// Step 4: Check if OTP is expired
	if time.Now().After(otpRecord.ExpiresAt) {
		h.markOTPAsUsed(otpRecord.ID)
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "OTP has expired. Please request a new one.",
		})
		return
	}

	// Step 5: Check if already used
	if otpRecord.IsUsed {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "OTP has already been used.",
		})
		return
	}

	// Step 6: Check attempts
	if otpRecord.Attempts >= 3 {
		h.markOTPAsUsed(otpRecord.ID)
		h.logAudit(req.Phone, "verify_otp", "failed", "max_attempts_exceeded", c.ClientIP())
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Maximum verification attempts exceeded. Please request a new OTP.",
		})
		return
	}

	// Step 7: Verify OTP hash
	if !h.verifyOTPHash(req.OTPCode, otpRecord.OTPHash) {
		// Increment attempts
		h.incrementOTPAttempts(otpRecord.ID)
		h.updateRateLimit(userID, "verify_otp")

		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Invalid OTP code",
		})
		return
	}

	// Step 8: Mark OTP as used
	h.markOTPAsUsed(otpRecord.ID)

	// Step 9: Generate JWT token
	jwtToken, err := h.generateToken(user)
	if err != nil {
		fmt.Printf("[OTP] JWT generation error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	// Step 10: Set HttpOnly cookie only for local development
	// For production, frontend uses localStorage which is more reliable
	isLocalhost := c.Request.Host == "localhost:8080" || c.Request.Host == "127.0.0.1:8080" || c.Request.Host == "localhost:3000" || c.Request.Host == "127.0.0.1:3000"
	if isLocalhost {
		c.SetCookie("token", jwtToken, 86400, "/", "", false, true)
	}

	// Step 11: Log success
	h.logAudit(req.Phone, "verify_otp", "success", "login_successful", c.ClientIP())

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Login successful",
		Data: models.LoginResponse{
			Token: jwtToken,
			User:  user,
		},
	})
}

// Helper functions

func (h *OTPHandler) generateOTP() string {
	otp := ""
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", num.Int64())
	}
	return otp
}

func (h *OTPHandler) hashOTP(otp string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *OTPHandler) verifyOTPHash(plaintext, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	return err == nil
}

func (h *OTPHandler) saveOTP(userID, phone, otpHash string, expiresAt time.Time) (string, error) {
	otpData := map[string]interface{}{
		"user_id":    userID,
		"phone":      phone,
		"otp_hash":   otpHash,
		"expires_at": expiresAt,
		"is_used":    false,
		"attempts":   0,
	}

	data, _, err := h.supabase.From("otp_codes").
		Insert(otpData, false, "", "", "").
		Execute()

	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil || len(result) == 0 {
		return "", fmt.Errorf("failed to get OTP ID")
	}

	return result[0]["id"].(string), nil
}

func (h *OTPHandler) getLatestOTP(phone string) (*OTPRecord, error) {
	data, _, err := h.supabase.From("otp_codes").
		Select("*", "", false).
		Eq("phone", phone).
		Eq("is_used", "false").
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Limit(1, "").
		Execute()

	if err != nil {
		return nil, err
	}

	var otps []OTPRecord
	if err := json.Unmarshal(data, &otps); err != nil || len(otps) == 0 {
		return nil, fmt.Errorf("OTP not found")
	}

	return &otps[0], nil
}

func (h *OTPHandler) markOTPAsUsed(otpID string) error {
	updateData := map[string]interface{}{
		"is_used": true,
		"used_at": time.Now(),
	}

	_, _, err := h.supabase.From("otp_codes").
		Update(updateData, "", "").
		Eq("id", otpID).
		Execute()

	return err
}

func (h *OTPHandler) incrementOTPAttempts(otpID string) error {
	updateData := map[string]interface{}{
		"attempts": "attempts + 1",
	}

	_, _, err := h.supabase.From("otp_codes").
		Update(updateData, "", "").
		Eq("id", otpID).
		Execute()

	return err
}

func (h *OTPHandler) invalidatePreviousOTPs(phone string) {
	_, _, _ = h.supabase.From("otp_codes").
		Update(map[string]interface{}{"is_used": true}, "", "").
		Eq("phone", phone).
		Eq("is_used", "false").
		Execute()
}

func (h *OTPHandler) userExists(phone string) (bool, error) {
	data, _, err := h.supabase.From("users").
		Select("id", "", false).
		Eq("phone", phone).
		Eq("is_active", "true").
		Execute()

	if err != nil {
		return false, err
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(data, &users); err != nil {
		return false, err
	}

	return len(users) > 0, nil
}

func (h *OTPHandler) getUserByPhone(phone string) (*models.User, error) {
	data, _, err := h.supabase.From("users").
		Select("*", "", false).
		Eq("phone", phone).
		Eq("is_active", "true").
		Execute()

	if err != nil {
		return nil, err
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil || len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

func (h *OTPHandler) checkRateLimit(userID, action string) (bool, error) {
	data, _, err := h.supabase.From("rate_limits").
		Select("*", "", false).
		Eq("user_id", userID).
		Eq("action", action).
		Execute()

	if err != nil {
		return false, err
	}

	var limits []map[string]interface{}
	if err := json.Unmarshal(data, &limits); err != nil || len(limits) == 0 {
		return true, nil
	}

	limit := limits[0]
	attemptCount := int(limit["attempt_count"].(float64))
	resetAtStr := limit["reset_at"].(string)

	resetAt, err := time.Parse(time.RFC3339, resetAtStr)
	if err != nil {
		return true, nil
	}

	if time.Now().After(resetAt) {
		return true, nil
	}

	maxAttempts := 3
	if action == "request_otp" {
		maxAttempts = 3
	} else if action == "verify_otp" {
		maxAttempts = 5
	}

	return attemptCount < maxAttempts, nil
}

func (h *OTPHandler) updateRateLimit(userID, action string) error {
	data, _, err := h.supabase.From("rate_limits").
		Select("*", "", false).
		Eq("user_id", userID).
		Eq("action", action).
		Execute()

	if err != nil {
		return err
	}

	var limits []map[string]interface{}
	if err := json.Unmarshal(data, &limits); err != nil || len(limits) == 0 {
		var resetDuration time.Duration
		if action == "request_otp" {
			resetDuration = 1 * time.Hour
		} else if action == "verify_otp" {
			resetDuration = 1 * time.Minute
		}

		limitData := map[string]interface{}{
			"user_id":       userID,
			"action":        action,
			"attempt_count": 1,
			"reset_at":      time.Now().Add(resetDuration),
		}

		_, _, err := h.supabase.From("rate_limits").
			Insert(limitData, false, "", "", "").
			Execute()

		return err
	}

	return nil
}

func (h *OTPHandler) logAudit(phone, action, status, reason, ipAddress string) {
	auditData := map[string]interface{}{
		"phone":      phone,
		"action":     action,
		"status":     status,
		"reason":     reason,
		"ip_address": ipAddress,
	}

	_, _, _ = h.supabase.From("otp_audit_log").
		Insert(auditData, false, "", "", "").
		Execute()
}

// generateToken generates JWT token for customer
func (h *OTPHandler) generateToken(user *models.User) (string, error) {
	claims := middleware.Claims{
		UserID:   user.ID,
		Phone:    user.Phone,
		Role:     user.Role,
		FullName: user.FullName,
		Provider: "otp",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}
