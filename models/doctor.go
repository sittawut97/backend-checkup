package models

import "time"

type Doctor struct {
	ID        string    `json:"id" db:"id"`
	FullName  string    `json:"full_name" db:"full_name"`
	Title     *string   `json:"title,omitempty" db:"title"`
	Specialty string    `json:"specialty" db:"specialty"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type DoctorSchedule struct {
	ID           string    `json:"id" db:"id"`
	DoctorID     string    `json:"doctor_id" db:"doctor_id"`
	ScheduleDate string    `json:"schedule_date" db:"schedule_date"`
	DayOfWeek    *string   `json:"day_of_week,omitempty" db:"day_of_week"`
	IsAvailable  bool      `json:"is_available" db:"is_available"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type TimeSlot struct {
	ID               string    `json:"id" db:"id"`
	DoctorScheduleID string    `json:"doctor_schedule_id" db:"doctor_schedule_id"`
	StartTime        string    `json:"start_time" db:"start_time"`
	EndTime          string    `json:"end_time" db:"end_time"`
	Status           string    `json:"status" db:"status"`
	MaxCapacity      int       `json:"max_capacity" db:"max_capacity"`
	CurrentBookings  int       `json:"current_bookings" db:"current_bookings"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type TimeSlotWithDoctor struct {
	TimeSlot
	DoctorID    string  `json:"doctor_id"`
	DoctorName  string  `json:"doctor_name"`
	DoctorTitle *string `json:"doctor_title,omitempty"`
	Specialty   string  `json:"specialty"`
}

type AvailableSlot struct {
	TimeSlotID      string  `json:"time_slot_id"`
	DoctorID        string  `json:"doctor_id"`
	DoctorName      string  `json:"doctor_name"`
	DoctorTitle     *string `json:"doctor_title,omitempty"`
	Specialty       string  `json:"specialty"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	AvailableSlots  int     `json:"available_slots"`
}
