package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/middleware"
	"github.com/sittawut/backend-appointment/models"
	"github.com/sittawut/backend-appointment/services"
	"github.com/supabase-community/postgrest-go"
	supa "github.com/supabase-community/supabase-go"
)

type AuthHandler struct {
	supabase *supa.Client
	config   *config.Config
	sms      services.SMSClient
}

func NewAuthHandler(supabase *supa.Client, cfg *config.Config, smsClient services.SMSClient) *AuthHandler {
	return &AuthHandler{
		supabase: supabase,
		config:   cfg,
		sms:      smsClient,
	}
}

// RequestOTP generates and sends OTP to user's phone
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Failed to read request body",
		})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req models.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Check if user exists
	var users []models.User
	fmt.Printf("[RequestOTP] Phone: %s\n", req.Phone)
	data, _, err := h.supabase.From("users").
		Select("id,phone,full_name,is_active", "", false).
		Eq("phone", req.Phone).
		Eq("is_active", "true").
		Execute()

	if err != nil {
		fmt.Printf("[RequestOTP] Database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Database query failed",
		})
		return
	}

	fmt.Printf("[RequestOTP] Query data: %s\n", string(data))
	if err := json.Unmarshal(data, &users); err != nil {
		fmt.Printf("[RequestOTP] Unmarshal error: %v\n", err)
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Phone number not found",
		})
		return
	}

	if len(users) == 0 {
		fmt.Printf("[RequestOTP] No users found for phone: %s\n", req.Phone)
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Phone number not found",
		})
		return
	}

	// Invalidate all previous unused OTPs for this phone
	if _, _, err := h.supabase.From("otp_codes").
		Update(map[string]interface{}{"is_used": true}, "", "").
		Eq("phone", req.Phone).
		Eq("is_used", "false").
		Execute(); err != nil {
		fmt.Printf("[RequestOTP] Warning: Failed to invalidate previous OTPs: %v\n", err)
	}

	// Send OTP via SMSMKT
	fmt.Printf("[RequestOTP] Sending OTP to phone: %s\n", req.Phone)
	token, err := h.sms.SendOTP(req.Phone)
	if err != nil {
		fmt.Printf("[RequestOTP] SMSMKT error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to send OTP: %v", err),
		})
		return
	}

	fmt.Printf("[RequestOTP] OTP token received: %s\n", token)
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

	fmt.Printf("[RequestOTP] Saving OTP data: %+v\n", otpData)
	_, _, err = h.supabase.From("otp_codes").
		Insert(otpData, false, "", "", "").
		Execute()

	if err != nil {
		fmt.Printf("[RequestOTP] Database insert error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save OTP token: %v", err),
		})
		return
	}

	fmt.Printf("[RequestOTP] OTP sent successfully\n")

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
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Get latest unused OTP for this phone
	var otps []models.OTP
	data, _, err := h.supabase.From("otp_codes").
		Select("*", "", false).
		Eq("phone", req.Phone).
		Eq("is_used", "false").
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Database query failed",
		})
		return
	}

	if err := json.Unmarshal(data, &otps); err != nil || len(otps) == 0 {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "No valid OTP found",
		})
		return
	}

	// We ordered by created_at DESC, so the latest record is the first element.
	otp := otps[0]
	token := otp.ID

	// Check if OTP is expired
	if time.Now().After(otp.ExpiresAt) {
		// Mark OTP as used so it cannot be reused.
		if _, _, err := h.supabase.From("otp_codes").
			Update(map[string]interface{}{"is_used": true}, "", "").
			Eq("id", otp.ID).
			Execute(); err != nil {
			fmt.Printf("[VerifyOTP] Warning: Failed to mark expired OTP as used: %v\n", err)
		}

		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "OTP has expired",
		})
		return
	}

	// Check attempts
	if otp.Attempts >= 3 {
		// Mark OTP as used so it cannot be reused.
		if _, _, err := h.supabase.From("otp_codes").
			Update(map[string]interface{}{"is_used": true}, "", "").
			Eq("id", otp.ID).
			Execute(); err != nil {
			fmt.Printf("[VerifyOTP] Warning: Failed to mark OTP as used after max attempts: %v\n", err)
		}

		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Too many attempts. Please request a new OTP",
		})
		return
	}

	// Validate OTP with SMS2PRO
	if err := h.sms.ValidateOTP(token, req.OTPCode); err != nil {
		// Increment attempts
		updateData := map[string]interface{}{
			"attempts": otp.Attempts + 1,
		}
		if _, _, updateErr := h.supabase.From("otp_codes").
			Update(updateData, "", "").
			Eq("id", otp.ID).
			Execute(); updateErr != nil {
			fmt.Printf("[VerifyOTP] Warning: Failed to increment attempts: %v\n", updateErr)
		}

		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Error:   "Invalid OTP code",
		})
		return
	}

	// Mark OTP as used
	updateData := map[string]interface{}{
		"is_used": true,
	}
	if _, _, err := h.supabase.From("otp_codes").
		Update(updateData, "", "").
		Eq("id", otp.ID).
		Execute(); err != nil {
		fmt.Printf("[VerifyOTP] Warning: Failed to mark OTP as used: %v\n", err)
	}

	// Get user details
	var users []models.User
	userData, _, err := h.supabase.From("users").
		Select("*", "", false).
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

	if err := json.Unmarshal(userData, &users); err != nil || len(users) == 0 {
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

	// Set HttpOnly cookie for middleware to read
	// Use parent domain for cross-domain cookie sharing
	// Secure=false for localhost development, true for production
	secure := c.Request.Host != "localhost:8080" && c.Request.Host != "127.0.0.1:8080"

	// Determine domain based on request host
	domain := ""
	if secure {
		// Production: use parent domain for cross-domain cookies
		if strings.Contains(c.Request.Host, "railway.app") {
			domain = ".up.railway.app"
		} else if strings.Contains(c.Request.Host, "vercel.app") {
			domain = ".vercel.app"
		}
	}

	c.SetCookie("token", jwtToken, 86400, "/", domain, secure, true)

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Login successful",
		Data: models.LoginResponse{
			Token: jwtToken,
			User:  &user,
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

	// Build new user object, including only provided optional fields
	newUser := map[string]interface{}{
		"id":        uuid.New().String(),
		"phone":     req.Phone,
		"full_name": req.FullName,
		"role":      "customer",
		"is_active": true,
	}

	// helper to add string pointer if not nil/empty
	addString := func(key string, val *string) {
		if val != nil && *val != "" {
			newUser[key] = *val
		}
	}

	addString("birth_date", req.BirthDate)
	{
		genderVal := "other"
		if req.Gender != nil {
			g := strings.ToLower(strings.TrimSpace(*req.Gender))
			if g == "male" || g == "female" || g == "other" {
				genderVal = g
			}
		}
		newUser["gender"] = genderVal
	}
	addString("email", req.Email)
	addString("address", req.Address)
	addString("blood_type", req.BloodType)
	addString("company_name", req.CompanyName)

	if req.Age != nil {
		newUser["age"] = *req.Age
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
			User:  &user,
		},
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	provider, _ := c.Get("provider")
	getString := func(key string) string {
		if value, exists := c.Get(key); exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
		return ""
	}

	if provider == "azure" {
		user := models.User{
			ID:       getString("user_id"),
			Phone:    getString("phone"),
			FullName: getString("full_name"),
			Role:     getString("role"),
			IsActive: true,
		}
		if email := getString("email"); email != "" {
			user.Email = &email
		}
		if employeeID := getString("employee_id"); employeeID != "" {
			user.EmployeeID = &employeeID
		}
		if department := getString("department"); department != "" {
			user.Department = &department
		}
		if jobTitle := getString("job_title"); jobTitle != "" {
			user.JobTitle = &jobTitle
		}

		c.JSON(http.StatusOK, models.Response{
			Success: true,
			Data:    user,
		})
		return
	}

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
