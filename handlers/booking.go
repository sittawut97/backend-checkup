package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/models"
	supa "github.com/supabase-community/supabase-go"
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

	var bookingDetails []models.BookingWithDetails
	for _, b := range bookings {
		var appointments []models.Appointment
		aptData, _, aptErr := h.supabase.From("appointments").
			Select("*", "", false).
			Eq("booking_id", b.ID).
			Execute()
		if aptErr == nil {
			json.Unmarshal(aptData, &appointments)
		}

		var appointmentDetails []models.AppointmentWithDetails
		for _, apt := range appointments {
			var doctors []models.Doctor
			doctorData, _, _ := h.supabase.From("doctors").
				Select("id, full_name, title", "", false).
				Eq("id", apt.DoctorID).
				Execute()
			json.Unmarshal(doctorData, &doctors)

			var slots []models.TimeSlot
			slotData, _, _ := h.supabase.From("time_slots").
				Select("id, start_time, end_time", "", false).
				Eq("id", apt.TimeSlotID).
				Execute()
			json.Unmarshal(slotData, &slots)

			aptWithDetails := models.AppointmentWithDetails{Appointment: apt}
			if len(doctors) > 0 {
				aptWithDetails.DoctorName = doctors[0].FullName
				if doctors[0].Title != nil {
					aptWithDetails.DoctorTitle = *doctors[0].Title
				}
			}
			if len(slots) > 0 {
				aptWithDetails.StartTime = slots[0].StartTime
				aptWithDetails.EndTime = slots[0].EndTime
			}
			appointmentDetails = append(appointmentDetails, aptWithDetails)
		}

		bookingDetails = append(bookingDetails, models.BookingWithDetails{
			Booking:      b,
			Appointments: appointmentDetails,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    bookingDetails,
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
	roleStr, ok := role.(string)
	userIDStr, ok2 := userID.(string)
	if ok && ok2 && roleStr == "customer" {
		query = query.Eq("customer_id", userIDStr)
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

	var appointmentDetails []models.AppointmentWithDetails
	for _, apt := range appointments {
		var doctors []models.Doctor
		doctorData, _, _ := h.supabase.From("doctors").
			Select("id, full_name, title", "", false).
			Eq("id", apt.DoctorID).
			Execute()
		json.Unmarshal(doctorData, &doctors)

		var slots []models.TimeSlot
		slotData, _, _ := h.supabase.From("time_slots").
			Select("id, start_time, end_time", "", false).
			Eq("id", apt.TimeSlotID).
			Execute()
		json.Unmarshal(slotData, &slots)

		aptWithDetails := models.AppointmentWithDetails{Appointment: apt}
		if len(doctors) > 0 {
			aptWithDetails.DoctorName = doctors[0].FullName
			if doctors[0].Title != nil {
				aptWithDetails.DoctorTitle = *doctors[0].Title
			}
		}
		if len(slots) > 0 {
			aptWithDetails.StartTime = slots[0].StartTime
			aptWithDetails.EndTime = slots[0].EndTime
		}
		appointmentDetails = append(appointmentDetails, aptWithDetails)
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
		Appointments: appointmentDetails,
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
		fmt.Printf("[CreateBooking] Bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	fmt.Printf("[CreateBooking] Request received - CustomerID: %s, AppointmentDate: %s, Appointments: %d\n",
		req.CustomerID, req.AppointmentDate, len(req.Appointments))

	// Default customer_id to authenticated user if not provided
	if req.CustomerID == "" {
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Error:   "Invalid user context",
			})
			return
		}
		req.CustomerID = userIDStr
	}

	// Verify customer_id matches authenticated user (unless nurse/admin)
	role, _ := c.Get("role")
	roleStr, ok := role.(string)
	userIDStr, ok2 := userID.(string)
	if ok && ok2 && roleStr == "customer" && req.CustomerID != userIDStr {
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
		"created_by":       userIDStr,
		"updated_by":       userIDStr,
	}

	if req.Status != "" {
		bookingData["status"] = req.Status
	}

	fmt.Printf("[CreateBooking] Booking data to insert: %+v\n", bookingData)

	var createdBookings []models.Booking
	data, _, err := h.supabase.From("bookings").
		Insert(bookingData, false, "", "", "").
		Execute()

	fmt.Printf("[CreateBooking] Response data: %s, Error: %v\n", string(data), err)

	if err != nil {
		fmt.Printf("[CreateBooking] Insert error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to create booking: %v", err),
		})
		return
	}

	if err := json.Unmarshal(data, &createdBookings); err != nil {
		fmt.Printf("[CreateBooking] Unmarshal error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse booking response: %v", err),
		})
		return
	}

	if len(createdBookings) == 0 {
		fmt.Printf("[CreateBooking] No bookings returned from insert\n")
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "No booking created",
		})
		return
	}

	booking := createdBookings[0]

	// Create appointments
	for i, apt := range req.Appointments {
		fmt.Printf("[CreateBooking] Creating appointment %d: TimeSlotID=%s, DoctorID=%s, ServiceType=%s\n",
			i, apt.TimeSlotID, apt.DoctorID, apt.ServiceType)

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

		fmt.Printf("[CreateBooking] Appointment %d response: %s, Error: %v\n", i, string(aptData), err)

		if err != nil {
			fmt.Printf("[CreateBooking] Appointment creation failed, rolling back booking\n")
			// Rollback: delete booking if appointment creation fails
			h.supabase.From("bookings").Delete("", "").Eq("id", booking.ID).Execute()
			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Error:   fmt.Sprintf("Failed to create appointments: %v", err),
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
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Invalid user context",
		})
		return
	}
	updateData["updated_by"] = userIDStr

	query := h.supabase.From("bookings").
		Update(updateData, "", "").
		Eq("id", bookingID)

	// If customer, only update their own bookings
	roleStr, okRole := role.(string)
	userIDStr2, okUserID := userID.(string)
	if okRole && okUserID && roleStr == "customer" {
		query = query.Eq("customer_id", userIDStr2)
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

	// If customer, ensure booking belongs to them
	if role.(string) == "customer" {
		// Check ownership
		var rows []map[string]interface{}
		data, _, err := h.supabase.From("bookings").
			Select("id", "", false).
			Eq("id", bookingID).
			Eq("customer_id", userID.(string)).
			Execute()
		if err != nil || json.Unmarshal(data, &rows) != nil || len(rows) == 0 {
			c.JSON(http.StatusForbidden, models.Response{Success: false, Error: "Not allowed"})
			return
		}
	}

	// Delete child appointments first
	_, _, _ = h.supabase.From("appointments").Delete("", "").Eq("booking_id", bookingID).Execute()

	// Delete booking
	delResp, _, err := h.supabase.From("bookings").Delete("", "").Eq("id", bookingID).Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{Success: false, Error: "Failed to delete booking"})
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Message: "Booking cancelled successfully", Data: string(delResp)})
}
