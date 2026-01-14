package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	supa "github.com/supabase-community/supabase-go"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/models"
)

type BookingHandler struct {
	supabase *supa.Client
	config   *config.Config
}

func NewBookingHandler(supabase *supa.Client, cfg *config.Config) *BookingHandler {
	return &BookingHandler{
		supabase: supabase,
		config:   cfg,
	}
}

func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")

	query := h.supabase.From("bookings").
		Select("*", "", false).
		Eq("customer_id", userID.(string)).
		Order("appointment_date", nil)

	if status != "" {
		query = query.Eq("status", status)
	}

	var bookings []models.Booking
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &bookings)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to fetch bookings",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    bookings,
	})
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	bookingID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	query := h.supabase.From("bookings").
		Select("*", "", false).
		Eq("id", bookingID)

	// If customer, only show their own bookings
	if role.(string) == "customer" {
		query = query.Eq("customer_id", userID.(string))
	}

	var bookings []models.Booking
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &bookings)
	}

	if err != nil || len(bookings) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Booking not found",
		})
		return
	}

	// Get appointments for this booking
	var appointments []models.Appointment
	data2, _, err2 := h.supabase.From("appointments").
		Select("*", "", false).
		Eq("booking_id", bookingID).
		Execute()
	if err2 == nil {
		json.Unmarshal(data2, &appointments)
	}


	// Get customer info
	var users []models.User
	data3, _, _ := h.supabase.From("users").
		Select("full_name, phone", "", false).
		Eq("id", bookings[0].CustomerID).
		Execute()
	json.Unmarshal(data3, &users)

	bookingWithDetails := models.BookingWithDetails{
		Booking:      bookings[0],
		Appointments: appointments,
	}

	if len(users) > 0 {
		bookingWithDetails.CustomerName = users[0].FullName
		bookingWithDetails.CustomerPhone = users[0].Phone
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    bookingWithDetails,
	})
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Verify customer_id matches authenticated user (unless nurse/admin)
	role, _ := c.Get("role")
	if role.(string) == "customer" && req.CustomerID != userID.(string) {
		c.JSON(http.StatusForbidden, models.Response{
			Success: false,
			Error:   "Cannot create booking for another user",
		})
		return
	}

	// Create booking
	bookingData := map[string]interface{}{
		"customer_id":      req.CustomerID,
		"appointment_date": req.AppointmentDate,
		"status":           "pending",
		"notes":            req.Notes,
		"created_by":       userID.(string),
	}

	if req.Status != "" {
		bookingData["status"] = req.Status
	}

	var createdBookings []models.Booking
	data, _, err := h.supabase.From("bookings").
		Insert(bookingData, false, "", "", "").
		Execute()
	if err == nil {
		err = json.Unmarshal(data, &createdBookings)
	}

	if err != nil || len(createdBookings) == 0 {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to create booking",
		})
		return
	}

	booking := createdBookings[0]

	// Create appointments
	for _, apt := range req.Appointments {
		appointmentData := map[string]interface{}{
			"booking_id":   booking.ID,
			"time_slot_id": apt.TimeSlotID,
			"doctor_id":    apt.DoctorID,
			"service_type": apt.ServiceType,
			"location":     apt.Location,
			"status":       "pending",
		}

		var createdAppointments []models.Appointment
		aptData, _, err := h.supabase.From("appointments").
			Insert(appointmentData, false, "", "", "").
			Execute()

		if err != nil {
			// Rollback: delete booking if appointment creation fails
			h.supabase.From("bookings").Delete("", "").Eq("id", booking.ID).Execute()
			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Error:   "Failed to create appointments",
			})
			return
		}
		json.Unmarshal(aptData, &createdAppointments)
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Booking created successfully",
		Data:    booking,
	})
}

func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	bookingID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req models.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Build update data
	updateData := make(map[string]interface{})
	if req.AppointmentDate != nil {
		updateData["appointment_date"] = *req.AppointmentDate
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}
	if req.Notes != nil {
		updateData["notes"] = *req.Notes
	}
	updateData["updated_by"] = userID.(string)

	query := h.supabase.From("bookings").
		Update(updateData, "", "").
		Eq("id", bookingID)

	// If customer, only update their own bookings
	if role.(string) == "customer" {
		query = query.Eq("customer_id", userID.(string))
	}

	var updatedBookings []models.Booking
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &updatedBookings)
	}

	if err != nil || len(updatedBookings) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Booking not found or update failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Booking updated successfully",
		Data:    updatedBookings[0],
	})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	updateData := map[string]interface{}{
		"status":     "cancelled",
		"updated_by": userID.(string),
	}

	query := h.supabase.From("bookings").
		Update(updateData, "", "").
		Eq("id", bookingID)

	// If customer, only cancel their own bookings
	if role.(string) == "customer" {
		query = query.Eq("customer_id", userID.(string))
	}

	var updatedBookings []models.Booking
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &updatedBookings)
	}

	if err != nil || len(updatedBookings) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Booking not found or cancellation failed",
		})
		return
	}

	// Update appointments status
	h.supabase.From("appointments").
		Update(map[string]interface{}{"status": "cancelled"}, "", "").
		Eq("booking_id", bookingID).
		Execute()

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Booking cancelled successfully",
		Data:    updatedBookings[0],
	})
}
