package models

import "time"

type User struct {
	ID          string    `json:"id" db:"id"`
	Phone       string    `json:"phone" db:"phone"`
	FullName    string    `json:"full_name" db:"full_name"`
	BirthDate   *string   `json:"birth_date,omitempty" db:"birth_date"`
	Gender      *string   `json:"gender,omitempty" db:"gender"`
	Email       *string   `json:"email,omitempty" db:"email"`
	Address     *string   `json:"address,omitempty" db:"address"`
	BloodType   *string   `json:"blood_type,omitempty" db:"blood_type"`
	Age         *int      `json:"age,omitempty" db:"age"`
	CompanyID   *string   `json:"company_id,omitempty" db:"company_id"`
	CompanyName *string   `json:"company_name,omitempty" db:"company_name"`
	Role        string    `json:"role" db:"role"`
	EmployeeID  *string   `json:"employee_id,omitempty" db:"employee_id"`
	Department  *string   `json:"department,omitempty" db:"department"`
	JobTitle    *string   `json:"job_title,omitempty" db:"job_title"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type RegisterRequest struct {
	Phone       string  `json:"phone" binding:"required"`
	FullName    string  `json:"full_name" binding:"required"`
	BirthDate   *string `json:"birth_date,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	Email       *string `json:"email,omitempty"`
	Address     *string `json:"address,omitempty"`
	BloodType   *string `json:"blood_type,omitempty"`
	Age         *int    `json:"age,omitempty"`
	CompanyName *string `json:"company_name,omitempty"`
}
