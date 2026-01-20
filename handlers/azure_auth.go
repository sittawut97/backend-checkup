package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/middleware"
	"github.com/sittawut/backend-appointment/models"
	supa "github.com/supabase-community/supabase-go"
)

type AzureAuthHandler struct {
	supabase *supa.Client
	config   *config.Config
}

func NewAzureAuthHandler(supabase *supa.Client, cfg *config.Config) *AzureAuthHandler {
	return &AzureAuthHandler{
		supabase: supabase,
		config:   cfg,
	}
}

// AzureTokenResponse represents Azure AD token response
type AzureTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

// AzureUserInfo represents user info from Azure AD
// type AzureUserInfo struct {
// 	ID          string `json:"id"`
// 	Email       string `json:"userPrincipalName"`
// 	Name        string `json:"displayName"`
// 	MobilePhone string `json:"mobilePhone"`
// }

type AzureUserInfo struct {
	ID          string `json:"id"`
	Mail        string `json:"mail"`
	UPN         string `json:"userPrincipalName"`
	Name        string `json:"displayName"`
	GivenName   string `json:"givenName"`
	Surname     string `json:"surname"`
	MobilePhone string `json:"mobilePhone"`
	JobTitle    string `json:"jobTitle"`
	EmployeeID  string `json:"employeeId"`
	Department  string `json:"department"`
}

// AzureCallback handles Azure AD OAuth callback
func (h *AzureAuthHandler) AzureCallback(c *gin.Context) {
	code := c.Query("code")
	// state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Missing authorization code",
		})
		return
	}

	// Exchange code for token
	token, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to authenticate with Azure",
		})
		return
	}

	// Get user info from Azure AD
	userInfo, err := h.getUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to get user information",
		})
		return
	}

	azureUser := buildAzureUser(userInfo)
	jwtToken, err := h.generateToken(azureUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	// Set HttpOnly cookie
	secure := c.Request.Host != "localhost:3000" && c.Request.Host != "127.0.0.1:3000"
	domain := ""
	if secure {
		if strings.Contains(c.Request.Host, "railway.app") {
			domain = ".up.railway.app"
		} else if strings.Contains(c.Request.Host, "vercel.app") {
			domain = ".vercel.app"
		}
	}

	c.SetCookie("token", jwtToken, 86400, "/", domain, secure, true)

	// Return success response
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Azure authentication successful",
		Data: models.LoginResponse{
			Token: jwtToken,
			User:  azureUser,
		},
	})
}

// exchangeCodeForToken exchanges authorization code for access token
func (h *AzureAuthHandler) exchangeCodeForToken(code string) (*AzureTokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", h.config.AzureTenantID)

	data := url.Values{
		"client_id":     {h.config.AzureClientID},
		"client_secret": {h.config.AzureClientSecret},
		"code":          {code},
		"redirect_uri":  {h.config.AzureRedirectURI},
		"grant_type":    {"authorization_code"},
		"scope":         {"openid profile email offline_access User.Read"},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp AzureTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	return &tokenResp, nil
}

// getUserInfo fetches user information from Azure AD
func (h *AzureAuthHandler) getUserInfo(accessToken string) (*AzureUserInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me?$select=id,displayName,jobTitle,mail,userPrincipalName,mobilePhone,employeeId,department,givenName,surname", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo AzureUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	if userInfo.Mail == "" {
		userInfo.Mail = userInfo.UPN
	}
	if userInfo.ID == "" {
		userInfo.ID = userInfo.UPN
	}

	// Log the raw response for debugging
	fmt.Printf("[Azure] Raw response: %s\n", string(body))
	fmt.Printf("[Azure] Parsed userInfo: ID=%s, Email=%s, GivenName=%s, Surname=%s, Phone=%s, JobTitle=%s, EmployeeID=%s, Department=%s\n", userInfo.ID, userInfo.Mail, userInfo.GivenName, userInfo.Surname, userInfo.MobilePhone, userInfo.JobTitle, userInfo.EmployeeID, userInfo.Department)

	return &userInfo, nil
}

func buildAzureUser(userInfo *AzureUserInfo) *models.User {
	email := userInfo.Mail
	phone := userInfo.MobilePhone
	if phone == "" {
		phone = "0000000000"
	}

	fullName := strings.TrimSpace(userInfo.GivenName + " " + userInfo.Surname)
	if fullName == "" {
		fullName = userInfo.Name
	}

	user := &models.User{
		ID:       userInfo.ID,
		Phone:    phone,
		FullName: fullName,
		Role:     "nurse",
		IsActive: true,
	}

	if email != "" {
		user.Email = &email
	}
	if userInfo.EmployeeID != "" {
		employeeID := userInfo.EmployeeID
		user.EmployeeID = &employeeID
	}
	if userInfo.Department != "" {
		department := userInfo.Department
		user.Department = &department
	}
	if userInfo.JobTitle != "" {
		jobTitle := userInfo.JobTitle
		user.JobTitle = &jobTitle
	}

	return user
}

// generateToken generates JWT token for nurse
func (h *AzureAuthHandler) generateToken(user *models.User) (string, error) {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	employeeID := ""
	if user.EmployeeID != nil {
		employeeID = *user.EmployeeID
	}
	department := ""
	if user.Department != nil {
		department = *user.Department
	}
	jobTitle := ""
	if user.JobTitle != nil {
		jobTitle = *user.JobTitle
	}

	claims := middleware.Claims{
		UserID:     user.ID,
		Phone:      user.Phone,
		Role:       user.Role,
		Email:      email,
		FullName:   user.FullName,
		EmployeeID: employeeID,
		Department: department,
		JobTitle:   jobTitle,
		Provider:   "azure",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}

// Logout handles logout request
func (h *AzureAuthHandler) Logout(c *gin.Context) {
	secure := !strings.Contains(c.Request.Host, "localhost")

	c.SetCookie(
		"token",
		"",
		-1, // expire ทันที
		"/",
		"",
		secure,
		true, // HttpOnly
	)

	c.SetCookie(
		"user",
		"",
		-1, // expire ทันที
		"/",
		"",
		secure,
		false,
	)

	c.JSON(200, gin.H{
		"success": true,
	})
}
