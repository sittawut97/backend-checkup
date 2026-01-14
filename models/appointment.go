package models

import "time"

type Appointment struct {
	ID          string    `json:"id" db:"id"`
	BookingID   string    `json:"booking_id" db:"booking_id"`
	TimeSlotID  string    `json:"time_slot_id" db:"time_slot_id"`
	DoctorID    string    `json:"doctor_id" db:"doctor_id"`
	ServiceType string    `json:"service_type" db:"service_type"`
	Location    *string   `json:"location,omitempty" db:"location"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AppointmentWithDetails struct {
	Appointment
	DoctorName  string `json:"doctor_name"`
	DoctorTitle string `json:"doctor_title"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

type CreateAppointmentRequest struct {
	TimeSlotID  string  `json:"time_slot_id" binding:"required"`
	DoctorID    string  `json:"doctor_id" binding:"required"`
	ServiceType string  `json:"service_type" binding:"required"`
	Location    *string `json:"location,omitempty"`
}
