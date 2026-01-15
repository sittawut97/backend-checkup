package models

import "time"

type Booking struct {
	ID              string    `json:"id" db:"id"`
	BookingNumber   string    `json:"booking_number" db:"booking_number"`
	CustomerID      string    `json:"customer_id" db:"customer_id"`
	AppointmentDate string    `json:"appointment_date" db:"appointment_date"`
	Status          string    `json:"status" db:"status"`
	Notes           *string   `json:"notes,omitempty" db:"notes"`
	CreatedBy       *string   `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy       *string   `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type BookingWithDetails struct {
	Booking
	CustomerName  string                   `json:"customer_name"`
	CustomerPhone string                   `json:"customer_phone"`
	Appointments  []AppointmentWithDetails `json:"appointments"`
}

type CreateBookingRequest struct {
	CustomerID      string                     `json:"customer_id" binding:"required"`
	AppointmentDate string                     `json:"appointment_date" binding:"required"`
	Status          string                     `json:"status"`
	Notes           *string                    `json:"notes,omitempty"`
	Appointments    []CreateAppointmentRequest `json:"appointments" binding:"required,min=1"`
}

type UpdateBookingRequest struct {
	AppointmentDate *string `json:"appointment_date,omitempty"`
	Status          *string `json:"status,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}
