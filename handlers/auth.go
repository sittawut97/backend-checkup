package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"
	"github.com/supabase-community/postgrest-go"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/middleware"
	"github.com/sittawut/backend-appointment/models"
	"github.com/sittawut/backend-appointment/services"
)

type AuthHandler struct {
	supabase *supa.Client
	config   *config.Config
	sms      *services.SMSMKTClient
}

func NewAuthHandler(supabase *supa.Client, cfg *config.Config, smsClient *services.SMSMKTClient) *AuthHandler {
	return &AuthHandler{
		supabase: supabase,
		config:   cfg,
		sms:      smsClient,
	}
}

// RequestOTP generates and sends OTP to user's phone
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	// Log raw body for debugging
	bodyBytes, _ := c.GetRawData()
	fmt.Printf("Raw request body: %s\n", string(bodyBytes))
	
	// Reset body for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	var req models.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Request OTP binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	
	fmt.Printf("Request OTP for phone: %s\n", req.Phone)

	// Check if user exists
	var users []models.User
	data, _, err := h.supabase.From("users").
		Select("id,phone,full_name,is_active", "", false).
		Eq("phone", req.Phone).
		Eq("is_active", "true").
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Database query failed",
		})
		return
	}

	if err := json.Unmarshal(data, &users); err != nil || len(users) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Phone number not found",
		})
		return
	}

	// Invalidate all previous unused OTPs for this phone
	h.supabase.From("otp_codes").
		Update(map[string]interface{}{"is_used": true}, "", "").
		Eq("phone", req.Phone).
		Eq("is_used", "false").
		Execute()

	// Send OTP via SMSMKT
	token, err := h.sms.SendOTP(req.Phone)
	if err != nil {
		fmt.Printf("Failed to send SMS: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to send OTP",
		})
		return
	}

	fmt.Printf("OTP sent to %s, token: %s\n", req.Phone, token)

	expiresAt := time.Now().Add(5 * time.Minute)

	// Save token to database (use id field to store token)
	otpData := map[string]interface{}{
		"id":         token,
		"phone":      req.Phone,
		"otp_code":   "000000",
		"expires_at": expiresAt,
		"is_used":    false,
		"attempts":   0,
	}

	_, _, err = h.supabase.From("otp_codes").
		Insert(otpData, false, "", "", "").
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to save OTP token",
		})
		return
	}
	

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "OTP sent successfully",
		Data: models.OTPResponse{
			Message:   "OTP has been sent to your phone",
			ExpiresAt: expiresAt,
		},
	})
}

// VerifyOTP verifies the OTP and logs in the user
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Verify OTP binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	
	fmt.Printf("Verify OTP - Phone: %s, OTP: %s\n", req.Phone, req.OTPCode)

	// Get latest unused OTP for this phone
	var otps []models.OTP
	data, _, err := h.supabase.From("otp_codes").
		Select("*", "", false).
		Eq("phone", req.Phone).
		Eq("is_used", "false").
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	if err != nil {
		fmt.Printf("Database query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Database query failed",
		})
		return
	}

	if err := json.Unmarshal(data, &otps); err != nil || len(otps) == 0 {
		fmt.Printf("Unmarshal error or no OTPs found. Error: %v, OTPs count: %d\n", err, len(otps))
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "No valid OTP found",
		})
		return
	}

	otp := otps[len(otps)-1] // Get the latest OTP (token in id field)
	token := otp.ID
	fmt.Printf("Found token: %s, Expires: %v, IsUsed: %v, Attempts: %d\n", token, otp.ExpiresAt, otp.IsUsed, otp.Attempts)

	// Check if OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "OTP has expired",
		})
		return
	}

	// Check attempts
	if otp.Attempts >= 3 {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Too many attempts. Please request a new OTP",
		})
		return
	}

	// Validate OTP with SMSMKT
	if err := h.sms.ValidateOTP(token, req.OTPCode, "CHECKUP"); err != nil {
		fmt.Printf("SMSMKT validation error: %v\n", err)
		
		// Increment attempts
		updateData := map[string]interface{}{
			"attempts": otp.Attempts + 1,
		}
		h.supabase.From("otp_codes").
			Update(updateData, "", "").
			Eq("id", otp.ID).
			Execute()

		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Invalid OTP code",
		})
		return
	}
	
	fmt.Printf("OTP validated successfully for phone: %s\n", req.Phone)

	// Mark OTP as used
	updateData := map[string]interface{}{
		"is_used": true,
	}
	h.supabase.From("otp_codes").
		Update(updateData, "", "").
		Eq("id", otp.ID).
		Execute()

	// Get user details
	var users []models.User
	userData, _, err := h.supabase.From("users").
		Select("*", "", false).
		Eq("phone", req.Phone).
		Eq("is_active", "true").
		Execute()

	if err != nil || json.Unmarshal(userData, &users) != nil || len(users) == 0 {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to get user details",
		})
		return
	}

	user := users[0]

	// ✅ แก้ไข - ใช้ชื่อตัวแปรใหม่
	jwtToken, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Login successful",
		Data: models.LoginResponse{
			Token: jwtToken,  // ← เปลี่ยนเป็น jwtToken
			User:  user,
		},
	})

}

// generateOTP creates a random 6-digit OTP
func (h *AuthHandler) generateOTP() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Check if phone already exists
	var existingUsers []models.User
	data, _, err := h.supabase.From("users").
		Select("id", "", false).
		Eq("phone", req.Phone).
		Execute()

	if err == nil {
		json.Unmarshal(data, &existingUsers)
		if len(existingUsers) > 0 {
			c.JSON(http.StatusConflict, models.Response{
				Success: false,
				Error:   "Phone number already registered",
			})
			return
		}
	}

	// Create new user
	newUser := map[string]interface{}{
		"id":         uuid.New().String(),
		"phone":      req.Phone,
		"full_name":  req.FullName,
		"birth_date":    req.BirthDate,
		"gender":        req.Gender,
		"email":         req.Email,
		"address":       req.Address,
		"blood_type":    req.BloodType,
		"age":           req.Age,
		"company_name":  req.CompanyName,
		"role":          "customer",
		"is_active":     true,
	}

	var createdUsers []models.User
	data, _, err = h.supabase.From("users").
		Insert(newUser, false, "", "", "").
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	if err := json.Unmarshal(data, &createdUsers); err != nil || len(createdUsers) == 0 {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "User created but not returned",
		})
		return
	}

	user := createdUsers[0]

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "User created but failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Registration successful",
		Data: models.LoginResponse{
			Token: token,
			User:  user,
		},
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var users []models.User
	data, _, err := h.supabase.From("users").
		Select("*", "", false).
		Eq("id", userID.(string)).
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Database query failed",
		})
		return
	}

	if err := json.Unmarshal(data, &users); err != nil || len(users) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    users[0],
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Remove fields that shouldn't be updated
	delete(req, "id")
	delete(req, "phone")
	delete(req, "role")
	delete(req, "created_at")

	var updatedUsers []models.User
	data, _, err := h.supabase.From("users").
		Update(req, "", "").
		Eq("id", userID.(string)).
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to update profile",
		})
		return
	}

	if err := json.Unmarshal(data, &updatedUsers); err != nil || len(updatedUsers) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data:    updatedUsers[0],
	})
}

func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := middleware.Claims{
		UserID: user.ID,
		Phone:  user.Phone,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}
