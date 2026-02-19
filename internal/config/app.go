package config

import (
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/delivery/http/route"
	"golang-clean-architecture/internal/gateway/email"
	"golang-clean-architecture/internal/repository"
	"golang-clean-architecture/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Mailer   *email.Mailer // optional — nil = log OTP codes instead of emailing
}

func Bootstrap(config *BootstrapConfig) {
	// JWT secret dari config (wajib diset di config.json / env)
	jwtSecret := config.Config.GetString("jwt.secret")

	// setup repositories
	userRepository := repository.NewUserRepository(config.Log)
	authProviderRepository := repository.NewAuthProviderRepository(config.Log)
	refreshTokenRepository := repository.NewRefreshTokenRepository(config.Log)
	otpRepository := repository.NewOtpRepository(config.Log)
	oauthStateRepository := repository.NewOauthStateRepository(config.Log)
	otpDeliveryRepository := repository.NewOtpDeliveryRepository(config.Log)

	// setup use cases
	userUseCase := usecase.NewUserUseCase(
		config.DB,
		config.Log,
		config.Validate,
		jwtSecret,
		userRepository,
		authProviderRepository,
		refreshTokenRepository,
		otpRepository,
		oauthStateRepository,
		config.Mailer,
		otpDeliveryRepository,
		config.Config.GetString("google.oauth.client_id"),
		config.Config.GetString("google.oauth.client_secret"),
		config.Config.GetString("google.oauth.redirect_uri"),
	)

	// setup controllers
	userController := http.NewUserController(userUseCase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		App:            config.App,
		UserController: userController,
		AuthMiddleware: authMiddleware,
	}
	routeConfig.Setup()
}
