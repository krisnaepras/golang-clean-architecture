package config

import (
	"golang-clean-architecture/internal/gateway/email"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewMailer creates a Mailer from SMTP settings in viper config.
// Returns nil if smtp.host is empty (SMTP not configured), which causes the
// usecase to fall back to logging the OTP code instead of emailing it.
func NewMailer(cfg *viper.Viper, log *logrus.Logger) *email.Mailer {
	host := cfg.GetString("smtp.host")
	if host == "" {
		log.Warn("SMTP host not configured — OTP emails will only be logged (set SMTP_HOST to enable)")
		return nil
	}

	port := cfg.GetInt("smtp.port")
	if port == 0 {
		port = 587
	}

	appName := cfg.GetString("app.name")
	if appName == "" {
		appName = "App"
	}

	mailer := email.NewMailer(
		host,
		port,
		cfg.GetString("smtp.username"),
		cfg.GetString("smtp.password"),
		cfg.GetString("smtp.from"),
		cfg.GetString("smtp.from_name"),
		appName,
	)

	log.Infof("Mailer configured: %s:%d (from: %s)", host, port, cfg.GetString("smtp.from"))
	return mailer
}
