package config

import (
	"fmt"
	"log"
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
	BaseURL       string `mapstructure:"BASE_URL"        validate:"required,url"`

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

	// 3x-UI API
	XUIURL      string `mapstructure:"XUI_URL"       validate:"required,url"`
	XUIUsername string `mapstructure:"XUI_USERNAME"  validate:"required"`
	XUIPassword string `mapstructure:"XUI_PASSWORD"  validate:"required"`
}

var (
	once   sync.Once
	config *Config
	v      = viper.New()
)

// Get возвращает синглтон-экземпляр конфигурации, загружая ее из стандартных путей.
func Get() *Config {
	once.Do(func() {
		var err error
		// Передаем пустую строку, чтобы loadConfig использовал стандартные пути viper
		config, err = Load("")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	})
	return config
}

// Load загружает конфигурацию из указанного файла .env.
// Если path пуст, viper будет искать .env в текущей директории.
func Load(path string) (*Config, error) {
	var cfg Config

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.AddConfigPath(".")
		v.SetConfigFile(".env")
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Если файл конфигурации не найден, это не является ошибкой,
		// так как мы можем полагаться на переменные окружения.
		// Ошибкой это будет только в том случае, если файл есть, но прочитать его не удалось.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return &cfg, nil
}

// loadConfig is a placeholder, as its logic has been moved to the new exported Load function.
func loadConfig() (*Config, error) {
	// The logic is now in Load(). This function is kept to avoid breaking
	// the existing singleton structure if it were more complex.
	return Load("")
}

func bindEnvs() {
	v.BindEnv("APP_ENV", "APP_ENV")
	v.BindEnv("HOST", "HOST")
	v.BindEnv("PORT", "PORT")
	v.BindEnv("LOG_LEVEL", "LOG_LEVEL")
	v.BindEnv("TELEGRAM_TOKEN", "TELEGRAM_TOKEN")
	v.BindEnv("BASE_URL", "BASE_URL")
	v.BindEnv("DB_HOST", "DB_HOST")
	v.BindEnv("DB_PORT", "DB_PORT")
	v.BindEnv("DB_USER", "DB_USER")
	v.BindEnv("DB_PASSWORD", "DB_PASSWORD")
	v.BindEnv("DB_NAME", "DB_NAME")
	v.BindEnv("JWT_SECRET_KEY", "JWT_SECRET_KEY")
	v.BindEnv("JWT_EXPIRES_IN_HOURS", "JWT_EXPIRES_IN_HOURS")
	v.BindEnv("RATE_LIMIT_REQUESTS", "RATE_LIMIT_REQUESTS")
	v.BindEnv("RATE_LIMIT_WINDOW_MINUTES", "RATE_LIMIT_WINDOW_MINUTES")
	v.BindEnv("READ_TIMEOUT", "READ_TIMEOUT")
	v.BindEnv("WRITE_TIMEOUT", "WRITE_TIMEOUT")
	v.BindEnv("IDLE_TIMEOUT", "IDLE_TIMEOUT")
	v.BindEnv("GRACEFUL_SHUTDOWN_TIMEOUT", "GRACEFUL_SHUTDOWN_TIMEOUT")
	v.BindEnv("XUI_URL", "XUI_URL")
	v.BindEnv("XUI_USERNAME", "XUI_USERNAME")
	v.BindEnv("XUI_PASSWORD", "XUI_PASSWORD")
}
