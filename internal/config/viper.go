package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// NewViper loads configuration from config.json (base) and .env (override).
// Environment variables take priority over config.json values.
func NewViper() *viper.Viper {
	// Load .env file if it exists — errors are silently ignored (file is optional)
	_ = godotenv.Load(".env")
	_ = godotenv.Load("./../.env")

	config := viper.New()

	config.SetConfigName("config")
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w \n", err))
	}

	// Map SCREAMING_SNAKE_CASE env vars to viper dot-notation keys.
	// Env vars (from .env or system) take priority over config.json.
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	// Explicit bindings for nested keys
	_ = config.BindEnv("app.name", "APP_NAME")
	_ = config.BindEnv("web.port", "WEB_PORT")
	_ = config.BindEnv("web.prefork", "WEB_PREFORK")
	_ = config.BindEnv("log.level", "LOG_LEVEL")
	_ = config.BindEnv("database.driver", "DATABASE_DRIVER")
	_ = config.BindEnv("database.host", "DB_HOST")
	_ = config.BindEnv("database.port", "DB_PORT")
	_ = config.BindEnv("database.username", "DB_USERNAME")
	_ = config.BindEnv("database.password", "DB_PASSWORD")
	_ = config.BindEnv("database.name", "DB_NAME")
	_ = config.BindEnv("database.pool.idle", "DB_POOL_IDLE")
	_ = config.BindEnv("database.pool.max", "DB_POOL_MAX")
	_ = config.BindEnv("database.pool.lifetime", "DB_POOL_LIFETIME")
	_ = config.BindEnv("kafka.enabled", "KAFKA_ENABLED")
	_ = config.BindEnv("kafka.bootstrap.servers", "KAFKA_BOOTSTRAP_SERVERS")
	_ = config.BindEnv("kafka.group.id", "KAFKA_GROUP_ID")
	_ = config.BindEnv("kafka.auto.offset.reset", "KAFKA_AUTO_OFFSET_RESET")
	_ = config.BindEnv("kafka.producer.enabled", "KAFKA_PRODUCER_ENABLED")
	_ = config.BindEnv("kafka.consumer.enabled", "KAFKA_CONSUMER_ENABLED")

	return config
}
