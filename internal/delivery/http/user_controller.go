package http

import (
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	Log     *logrus.Logger
	UseCase *usecase.UserUseCase
}

func NewUserController(useCase *usecase.UserUseCase, logger *logrus.Logger) *UserController {
	return &UserController{Log: logger, UseCase: useCase}
}

// POST /auth/register
func (c *UserController) Register(ctx *fiber.Ctx) error {
	request := new(model.RegisterRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.Register(ctx.UserContext(), request); err != nil {
		c.Log.Warnf("Register failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[any]{Data: fiber.Map{"message": "OTP sent to email"}})
}

// POST /auth/verify-email
func (c *UserController) VerifyEmail(ctx *fiber.Ctx) error {
	request := new(model.VerifyOtpRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	deviceInfo := ctx.Get("User-Agent")
	ip := ctx.IP()
	tokens, err := c.UseCase.VerifyEmailOtp(ctx.UserContext(), request, deviceInfo, ip)
	if err != nil {
		c.Log.Warnf("VerifyEmail failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TokenResponse]{Data: tokens})
}

// POST /auth/login
func (c *UserController) Login(ctx *fiber.Ctx) error {
	request := new(model.LoginRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	deviceInfo := ctx.Get("User-Agent")
	ip := ctx.IP()
	tokens, err := c.UseCase.Login(ctx.UserContext(), request, deviceInfo, ip)
	if err != nil {
		c.Log.Warnf("Login failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TokenResponse]{Data: tokens})
}

// POST /auth/refresh
func (c *UserController) Refresh(ctx *fiber.Ctx) error {
	request := new(model.RefreshRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	deviceInfo := ctx.Get("User-Agent")
	ip := ctx.IP()
	tokens, err := c.UseCase.Refresh(ctx.UserContext(), request, deviceInfo, ip)
	if err != nil {
		c.Log.Warnf("Refresh failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TokenResponse]{Data: tokens})
}

// POST /auth/logout
func (c *UserController) Logout(ctx *fiber.Ctx) error {
	request := new(model.LogoutRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.Logout(ctx.UserContext(), request); err != nil {
		c.Log.Warnf("Logout failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[any]{Data: fiber.Map{"message": "logged out"}})
}

// GET /auth/me  (requires auth)
func (c *UserController) Me(ctx *fiber.Ctx) error {
	auth := middleware.GetUser(ctx)
	response, err := c.UseCase.Me(ctx.UserContext(), auth.ID)
	if err != nil {
		c.Log.Warnf("Me failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[*model.UserResponse]{Data: response})
}

// POST /auth/forgot-password
func (c *UserController) ForgotPassword(ctx *fiber.Ctx) error {
	request := new(model.ForgotPasswordRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.ForgotPassword(ctx.UserContext(), request); err != nil {
		c.Log.Warnf("ForgotPassword failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[any]{Data: fiber.Map{"message": "OTP sent if email is registered"}})
}

// POST /auth/reset-password
func (c *UserController) ResetPassword(ctx *fiber.Ctx) error {
	request := new(model.ResetPasswordRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.ResetPassword(ctx.UserContext(), request); err != nil {
		c.Log.Warnf("ResetPassword failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[any]{Data: fiber.Map{"message": "password reset successful"}})
}

// POST /auth/otp/resend
func (c *UserController) ResendOtp(ctx *fiber.Ctx) error {
	request := new(model.ResendOtpRequest)
	if err := ctx.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.ResendOtp(ctx.UserContext(), request); err != nil {
		c.Log.Warnf("ResendOtp failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[any]{Data: fiber.Map{"message": "OTP sent"}})
}

// GET /auth/google
func (c *UserController) GoogleOAuth(ctx *fiber.Ctx) error {
	redirectURL, err := c.UseCase.GoogleOAuthURL(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

// GET /auth/google/callback
func (c *UserController) GoogleOAuthCallback(ctx *fiber.Ctx) error {
	request := new(model.OAuthCallbackRequest)
	if err := ctx.QueryParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	response, err := c.UseCase.GoogleOAuthCallback(ctx.UserContext(), request, ctx.Get("User-Agent"), ctx.IP())
	if err != nil {
		c.Log.Warnf("GoogleOAuthCallback failed: %+v", err)
		return err
	}
	return ctx.JSON(model.WebResponse[*model.TokenResponse]{Data: response})
}

// GET /auth/apple
func (c *UserController) AppleOAuth(ctx *fiber.Ctx) error {
	redirectURL, err := c.UseCase.AppleOAuthURL(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}
