package services

// SMSClient is an interface for SMS service providers
type SMSClient interface {
	SendOTP(phone string) (string, error)
	ValidateOTP(token string, otpCode string) error
}
