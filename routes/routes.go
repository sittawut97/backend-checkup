package routes

import (
	"github.com/gin-gonic/gin"
	supa "github.com/supabase-community/supabase-go"
	"github.com/sittawut/backend-appointment/config"
	"github.com/sittawut/backend-appointment/handlers"
	"github.com/sittawut/backend-appointment/middleware"
	"github.com/sittawut/backend-appointment/services"
)

func SetupRoutes(router *gin.Engine, supabaseClient *supa.Client, cfg *config.Config, smsClient services.SMSClient) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(supabaseClient, cfg, smsClient)
	bookingHandler := handlers.NewBookingHandler(supabaseClient, cfg)
	doctorHandler := handlers.NewDoctorHandler(supabaseClient, cfg)
	nurseHandler := handlers.NewNurseHandler(supabaseClient, cfg)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "Server is running",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", authHandler.RequestOTP)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/register", authHandler.Register)
		}

		// Public routes - Doctors and schedules (no auth required)
		v1.GET("/doctors", doctorHandler.GetDoctors)
		v1.GET("/doctors/:id", doctorHandler.GetDoctorByID)
		v1.GET("/schedules", doctorHandler.GetSchedules)
		v1.GET("/time-slots", doctorHandler.GetTimeSlots)
		v1.GET("/time-slots/available", doctorHandler.GetAvailableSlots)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// User profile
			protected.GET("/auth/me", authHandler.GetMe)
			protected.PUT("/auth/me", authHandler.UpdateProfile)

			// Customer bookings
			customer := protected.Group("/bookings")
			{
				customer.GET("", bookingHandler.GetMyBookings)
				customer.POST("", bookingHandler.CreateBooking)
				customer.GET("/:id", bookingHandler.GetBookingByID)
				customer.PUT("/:id", bookingHandler.UpdateBooking)
				customer.DELETE("/:id", bookingHandler.CancelBooking)
			}

			// Nurse routes
			nurse := protected.Group("/nurse")
			nurse.Use(middleware.RoleMiddleware("nurse", "admin"))
			{
				// Booking management
				nurse.GET("/bookings", nurseHandler.GetAllBookings)
				nurse.POST("/bookings", nurseHandler.CreateBookingForCustomer)
				nurse.PUT("/bookings/:id", nurseHandler.UpdateBooking)
				nurse.DELETE("/bookings/:id", nurseHandler.DeleteBooking)
				nurse.GET("/dashboard", nurseHandler.GetDashboard)

				// Slot management
				nurse.POST("/slots/block", nurseHandler.BlockTimeSlots)
				nurse.POST("/slots/unblock", nurseHandler.UnblockTimeSlots)

				// Doctor management
				nurse.POST("/doctors", nurseHandler.CreateDoctor)
				nurse.PUT("/doctors/:id", nurseHandler.UpdateDoctor)
				nurse.DELETE("/doctors/:id", nurseHandler.DeleteDoctor)

				// User management
				nurse.GET("/users", nurseHandler.GetAllUsers)
				nurse.POST("/users", nurseHandler.CreateUser)
				nurse.PUT("/users/:id", nurseHandler.UpdateUser)
			}
		}
	}
}
