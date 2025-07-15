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

type Config struct {
	Host     string `mapstructure:"HOST" validate:"required"`
	Port     string `mapstructure:"PORT"     validate:"required"`
	LogLevel string `mapstructure:"LOG_LEVEL" validate:"required"`

	TelegramToken string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`
	BaseURL       string `mapstructure:"BASE_URL"        validate:"required,url"`

	DBHost     string `mapstructure:"DB_HOST"     validate:"required"`
	DBPort     string `mapstructure:"DB_PORT"     validate:"required"`
	DBUser     string `mapstructure:"DB_USER"     validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName     string `mapstructure:"DB_NAME"     validate:"required"`

	RateLimitRequests      int    `mapstructure:"RATE_LIMIT_REQUESTS"       validate:"required,gte=0"`
	RateLimitWindowMinutes int    `mapstructure:"RATE_LIMIT_WINDOW_MINUTES" validate:"required,gte=1"`
	JWTSecretKey           string `mapstructure:"JWT_SECRET_KEY"       validate:"required,min=32"`
	JWTExpiresIn           int    `mapstructure:"JWT_EXPIRES_IN_HOURS"   validate:"gte=1"`

	XUIURL      string `mapstructure:"XUI_URL"       validate:"required,url"`
	XUIUsername string `mapstructure:"XUI_USERNAME"  validate:"required"`
	XUIPassword string `mapstructure:"XUI_PASSWORD"  validate:"required"`
}

var (
	cfg  *Config
	once sync.Once
)

func Get() *Config {
	once.Do(func() {
		var err error
		cfg, err = Load()
		if err != nil {
			log.Fatalf("failed to load configuration: %v", err)
		}
	})
	return cfg
}

func bindEnvs() {
	viper.BindEnv("HOST")
	viper.BindEnv("PORT")
	viper.BindEnv("LOG_LEVEL")
	viper.BindEnv("TELEGRAM_TOKEN")
	viper.BindEnv("BASE_URL")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("RATE_LIMIT_REQUESTS")
	viper.BindEnv("RATE_LIMIT_WINDOW_MINUTES")
	viper.BindEnv("JWT_SECRET_KEY")
	viper.BindEnv("JWT_EXPIRES_IN_HOURS")
	viper.BindEnv("XUI_URL")
	viper.BindEnv("XUI_USERNAME")
	viper.BindEnv("XUI_PASSWORD")
	viper.BindEnv("XUI_SERVICE")
}

func Load() (*Config, error) {
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	bindEnvs()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

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
