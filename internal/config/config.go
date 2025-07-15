package config

import (
	"fmt"
	"log"
	"os"
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
)

// Get возвращает синглтон-экземпляр конфигурации.
// При первом вызове загружает конфигурацию и кэширует ее.
func Get() *Config {
	once.Do(func() {
		var err error
		// Для основного приложения всегда ищем .env в текущей директории.
		config, err = Load(".")
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	})
	return config
}

// Load загружает конфигурацию из указанного пути.
// Эта функция является публичной и может использоваться в тестах для загрузки кастомных конфигов.
func Load(path string) (*Config, error) {
	vp := viper.New()

	// Настраиваем чтение из переменных окружения (высший приоритет)
	vp.AutomaticEnv()
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Настраиваем чтение из файла (низший приоритет)
	vp.AddConfigPath(path)
	vp.SetConfigFile(".env")

	// Пытаемся прочитать .env файл.
	if err := vp.ReadInConfig(); err != nil {
		// Мы игнорируем только ошибки 'файл не найден'. Все остальные ошибки (например, проблемы с правами доступа) являются критическими.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Если файл просто не найден, это нормально. Переменные окружения будут использованы.
	}

	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return &cfg, nil
}
