package routes

import (
	controller "fiber-boilerplate/pkg/controllers"

	"github.com/gofiber/fiber/v3"
)

func registerAuthRoutes(v1 fiber.Router, authController *controller.AuthController) {
	auth := v1.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/forgot-password", authController.ForgotPassword)
	auth.Post("/otp/verify", authController.VerifyOTP)
	auth.Post("/reset-password", authController.ResetPassword)
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/logout", authController.Logout)
	auth.Get("/me", authController.Me)
	auth.Get("/sessions", authController.Sessions)
	auth.Post("/sessions/revoke", authController.RevokeSession)
	auth.Post("/sessions/revoke-all", authController.RevokeAllSessions)
}
