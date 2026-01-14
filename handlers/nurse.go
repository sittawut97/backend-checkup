package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	supa "github.com/supabase-community/supabase-go"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/models"
)

type NurseHandler struct {
	supabase *supa.Client
	config   *config.Config
}

func NewNurseHandler(supabase *supa.Client, cfg *config.Config) *NurseHandler {
	return &NurseHandler{
		supabase: supabase,
		config:   cfg,
	}
}

func (h *NurseHandler) GetAllBookings(c *gin.Context) {
	status := c.Query("status")
	date := c.Query("date")

	query := h.supabase.From("bookings").
		Select("*", "", false).
		Order("appointment_date", nil)

	if status != "" {
		query = query.Eq("status", status)
	}
	if date != "" {
		query = query.Eq("appointment_date", date)
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

func (h *NurseHandler) CreateBookingForCustomer(c *gin.Context) {
	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	userID, _ := c.Get("user_id")

	bookingData := map[string]interface{}{
		"customer_id":      req.CustomerID,
		"appointment_date": req.AppointmentDate,
		"status":           req.Status,
		"notes":            req.Notes,
		"created_by":       userID.(string),
	}

	var createdBookings []models.Booking
	data, _, err := h.supabase.From("bookings").Insert(bookingData, false, "", "", "").Execute()
	if err == nil {
		err = json.Unmarshal(data, &createdBookings)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to create booking",
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Data:    createdBookings[0],
	})
}

func (h *NurseHandler) UpdateBooking(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) DeleteBooking(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) GetDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) BlockTimeSlots(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) UnblockTimeSlots(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) CreateDoctor(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) UpdateDoctor(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) DeleteDoctor(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) GetAllUsers(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) CreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}

func (h *NurseHandler) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{Success: true})
}
