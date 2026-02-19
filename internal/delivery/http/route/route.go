package route

import (
	"golang-clean-architecture/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App            *fiber.App
	UserController *http.UserController
	AuthMiddleware fiber.Handler
}

func (c *RouteConfig) Setup() {
	c.SetupGuestRoute()
	c.SetupAuthRoute()
}

// SetupGuestRoute mendaftarkan endpoint yang tidak memerlukan autentikasi.
func (c *RouteConfig) SetupGuestRoute() {
	auth := c.App.Group("/api/auth")

	auth.Post("/register", c.UserController.Register)
	auth.Post("/verify-email", c.UserController.VerifyEmail)
	auth.Post("/login", c.UserController.Login)
	auth.Post("/refresh", c.UserController.Refresh)
	auth.Post("/otp/resend", c.UserController.ResendOtp)
	auth.Post("/forgot-password", c.UserController.ForgotPassword)
	auth.Post("/reset-password", c.UserController.ResetPassword)

	auth.Get("/google", c.UserController.GoogleOAuth)
	auth.Get("/google/callback", c.UserController.GoogleOAuthCallback)
	auth.Get("/apple", c.UserController.AppleOAuth)
}

// SetupAuthRoute mendaftarkan endpoint yang memerlukan JWT.
func (c *RouteConfig) SetupAuthRoute() {
	auth := c.App.Group("/api/auth", c.AuthMiddleware)

	auth.Get("/me", c.UserController.Me)
	auth.Post("/logout", c.UserController.Logout)
}
