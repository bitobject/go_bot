package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server settings
	Port string
	Host string

	// Database
	DatabaseURL string

	// Telegram
	TelegramToken string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Logging
	LogLevel string

	// Metrics
	MetricsEnabled bool
}

var AppConfig *Config

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		// .env file is optional
	}
}

func Init() {
	AppConfig = &Config{
		Port:        getEnv("APP_PORT", "8080"),
		Host:        getEnv("APP_HOST", "0.0.0.0"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),

		// JWT settings
		JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),

		// Rate limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 200),
		RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Metrics
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
} 