package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config хранит всю конфигурацию приложения.
// Теги mapstructure используются для чтения данных из viper,
// а теги validate - для валидации полей.
type Config struct {
	// Server settings
	Host     string `mapstructure:"HOST" validate:"required"`
	Port     string `mapstructure:"PORT"     validate:"required"`
	LogLevel string `mapstructure:"LOG_LEVEL" validate:"required"`

	// Telegram Bot
	TelegramToken string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`
	WebhookURL    string `mapstructure:"WEBHOOK_URL"      validate:"required,url"`

	// Database
	DBHost     string `mapstructure:"DB_HOST"     validate:"required"`
	DBPort     string `mapstructure:"DB_PORT"     validate:"required"`
	DBUser     string `mapstructure:"DB_USER"     validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName     string `mapstructure:"DB_NAME"     validate:"required"`

	// API Security
	RateLimitRequests      int    `mapstructure:"RATE_LIMIT_REQUESTS"       validate:"required,gte=0"`
	RateLimitWindowMinutes int    `mapstructure:"RATE_LIMIT_WINDOW_MINUTES" validate:"required,gte=1"`
	JWTSecretKey           string `mapstructure:"JWT_SECRET_KEY"       validate:"required,min=32"`
	JWTExpiresIn           int    `mapstructure:"JWT_EXPIRES_IN_HOURS"   validate:"gte=1"`
}

var (
	once   sync.Once
	config *Config
)

// Get возвращает синглтон-экземпляр конфигурации.
// При первом вызове он загружает и валидирует конфигурацию.
// Если загрузка или валидация не удаются, функция паникует,
// так как приложение не может работать без корректной конфигурации.
func Get() *Config {
	once.Do(func() {
		var err error
		config, err = loadConfig()
		if err != nil {
			log.Fatalf("failed to load configuration: %v", err)
		}
	})
	return config
}

// loadConfig выполняет фактическую загрузку конфигурации.
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	bindEnvs()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func bindEnvs() {
	viper.BindEnv("APP_ENV", "APP_ENV")
	viper.BindEnv("HOST", "HOST")
	viper.BindEnv("PORT", "PORT")
	viper.BindEnv("LOG_LEVEL", "LOG_LEVEL")
	viper.BindEnv("TELEGRAM_TOKEN", "TELEGRAM_TOKEN")
	viper.BindEnv("WEBHOOK_URL", "WEBHOOK_URL")
	viper.BindEnv("DB_HOST", "DB_HOST")
	viper.BindEnv("DB_PORT", "DB_PORT")
	viper.BindEnv("DB_USER", "DB_USER")
	viper.BindEnv("DB_PASSWORD", "DB_PASSWORD")
	viper.BindEnv("DB_NAME", "DB_NAME")
	viper.BindEnv("JWT_SECRET_KEY", "JWT_SECRET_KEY")
	viper.BindEnv("JWT_EXPIRES_IN_HOURS", "JWT_EXPIRES_IN_HOURS")
	viper.BindEnv("RATE_LIMIT_REQUESTS", "RATE_LIMIT_REQUESTS")
	viper.BindEnv("RATE_LIMIT_WINDOW_MINUTES", "RATE_LIMIT_WINDOW_MINUTES")
	viper.BindEnv("READ_TIMEOUT", "READ_TIMEOUT")
	viper.BindEnv("WRITE_TIMEOUT", "WRITE_TIMEOUT")
	viper.BindEnv("IDLE_TIMEOUT", "IDLE_TIMEOUT")
	viper.BindEnv("GRACEFUL_SHUTDOWN_TIMEOUT", "GRACEFUL_SHUTDOWN_TIMEOUT")
}
