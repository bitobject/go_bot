package config

import (
	"errors"
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
	Host     string `mapstructure:"HOST"`
	Port     string `mapstructure:"PORT"     validate:"required"`
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// Telegram Bot
	TelegramToken string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`

	// Database
	DBHost     string `mapstructure:"DB_HOST"     validate:"required"`
	DBPort     string `mapstructure:"DB_PORT"     validate:"required"`
	DBUser     string `mapstructure:"DB_USER"     validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName     string `mapstructure:"DB_NAME"     validate:"required"`

	// API Security
	RateLimitRequests      int `mapstructure:"RATE_LIMIT_REQUESTS"       validate:"gte=0"`
	RateLimitWindowMinutes int `mapstructure:"RATE_LIMIT_WINDOW_MINUTES" validate:"gte=1"`
	JWTSecretKey      string `mapstructure:"JWT_SECRET_KEY"       validate:"required,min=32"`
	JWTExpiresIn      int    `mapstructure:"JWT_EXPIRES_IN_HOURS"   validate:"gte=1"`
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
	// Указываем viper на файл .env в текущей директории.
	// Это хорошо для локальной разработки.
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Устанавливаем значения по умолчанию. Они будут использованы, если
	// переменная не найдена ни в файле, ни в окружении.
	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("RATE_LIMIT_REQUESTS", 200)
	viper.SetDefault("RATE_LIMIT_WINDOW_MINUTES", 1)
	viper.SetDefault("JWT_EXPIRES_IN_HOURS", 24)

	// Включаем автоматическое чтение переменных окружения.
	// Это приоритетный способ для работы в production (например, в Docker).
	viper.AutomaticEnv()

	// Пытаемся прочитать файл конфигурации.
	if err := viper.ReadInConfig(); err != nil {
		// Ошибку "файл не найден" можно проигнорировать, 
		// так как мы можем полностью полагаться на переменные окружения.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Если ошибка другого типа (например, синтаксическая в файле), 
			// это критично, и мы должны вернуть ошибку.
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	// Десериализуем конфигурацию в структуру.
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Валидируем структуру.
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		// Ошибка валидации очень информативна, она укажет, какое поле и почему не прошло проверку.
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
} 