package config

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// NewViper loads configuration from .env (primary) and config.json (fallback for non-DB settings).
// Database configuration is read exclusively from .env / environment variables.
func NewViper() *viper.Viper {
	// Load .env file if it exists — errors are silently ignored (file is optional)
	_ = godotenv.Load(".env")
	_ = godotenv.Load("./../.env")

	config := viper.New()

	// config.json is optional — used only for non-DB settings like app, web, log, kafka
	config.SetConfigName("config")
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	_ = config.ReadInConfig() // ignore error — .env is the primary source

	// Defaults for database (overridden by .env / environment variables)
	config.SetDefault("database.driver", "postgres")
	config.SetDefault("database.host", "localhost")
	config.SetDefault("database.port", 5432)
	config.SetDefault("database.username", "postgres")
	config.SetDefault("database.password", "")
	config.SetDefault("database.name", "dreampod")
	config.SetDefault("database.pool.idle", 10)
	config.SetDefault("database.pool.max", 100)
	config.SetDefault("database.pool.lifetime", 300)

	// Map SCREAMING_SNAKE_CASE env vars to viper dot-notation keys.
	// Env vars (from .env or system) take priority over defaults and config.json.
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
	_ = config.BindEnv("jwt.secret", "JWT_SECRET")

	return config
}
