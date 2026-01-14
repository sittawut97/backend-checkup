package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	supa "github.com/supabase-community/supabase-go"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/models"
)

type DoctorHandler struct {
	supabase *supa.Client
	config   *config.Config
}

func NewDoctorHandler(supabase *supa.Client, cfg *config.Config) *DoctorHandler {
	return &DoctorHandler{
		supabase: supabase,
		config:   cfg,
	}
}

func (h *DoctorHandler) GetDoctors(c *gin.Context) {
	specialty := c.Query("specialty")

	query := h.supabase.From("doctors").
		Select("*", "", false).
		Eq("is_active", "true").
		Order("full_name", nil)

	if specialty != "" {
		query = query.Eq("specialty", specialty)
	}

	var doctors []models.Doctor
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &doctors)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to fetch doctors",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    doctors,
	})
}

func (h *DoctorHandler) GetDoctorByID(c *gin.Context) {
	doctorID := c.Param("id")

	var doctors []models.Doctor
	data, _, err := h.supabase.From("doctors").
		Select("*", "", false).
		Eq("id", doctorID).
		Execute()
	if err == nil {
		err = json.Unmarshal(data, &doctors)
	}

	if err != nil || len(doctors) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Error:   "Doctor not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    doctors[0],
	})
}

func (h *DoctorHandler) GetSchedules(c *gin.Context) {
	doctorID := c.Query("doctor_id")
	date := c.Query("date")

	query := h.supabase.From("doctor_schedules").
		Select("*", "", false).
		Eq("is_available", "true")

	if doctorID != "" {
		query = query.Eq("doctor_id", doctorID)
	}
	if date != "" {
		query = query.Eq("schedule_date", date)
	}

	var schedules []models.DoctorSchedule
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &schedules)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to fetch schedules",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    schedules,
	})
}

func (h *DoctorHandler) GetTimeSlots(c *gin.Context) {
	scheduleID := c.Query("schedule_id")
	status := c.Query("status")

	if scheduleID == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "schedule_id is required",
		})
		return
	}

	query := h.supabase.From("time_slots").
		Select("*", "", false).
		Eq("doctor_schedule_id", scheduleID).
		Order("start_time", nil)

	if status != "" {
		query = query.Eq("status", status)
	}

	var timeSlots []models.TimeSlot
	data, _, err := query.Execute()
	if err == nil {
		err = json.Unmarshal(data, &timeSlots)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Error:   "Failed to fetch time slots",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    timeSlots,
	})
}

func (h *DoctorHandler) GetAvailableSlots(c *gin.Context) {
	date := c.Query("date")
	_ = c.Query("specialty") // TODO: implement specialty filter

	if date == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Error:   "date is required",
		})
		return
	}

	// This would require a custom SQL query or multiple queries
	// For simplicity, returning a basic response
	// In production, you'd want to use a stored procedure or complex query

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Available slots endpoint - implement custom query",
		Data:    []models.AvailableSlot{},
	})
}
