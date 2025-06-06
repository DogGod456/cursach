package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWithEnv = errors.New("environment variable error")
)

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	User          string `yaml:"user"`           // Для обычных операций
	Password      string `yaml:"password"`       // Для обычных операций
	AdminUser     string `yaml:"admin_user"`     // Для админских операций
	AdminPassword string `yaml:"admin_password"` // Для админских операций
	DBName        string `yaml:"dbname"`
	SSLMode       string `yaml:"sslmode"`
}

type AuthConfig struct {
	Salt      string
	JWTSecret string
	JWTExpiry time.Duration
}

func LoadConfig() (*Config, error) {
	// Получаем все переменные окружения с проверкой ошибок
	host, err := getEnv("PGHOST")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGHOST")
	}

	portStr, err := getEnv("PGPORT")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGPORT")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PGPORT: %w", err)
	}

	user, err := getEnv("PGUSER")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGUSER")
	}

	password, err := getEnv("PGPASSWORD")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGPASSWORD")
	}

	adminUser, err := getEnv("PGADMIN_USER")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGADMIN_USER")
	}
	adminPassword, err := getEnv("PGADMIN_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGADMIN_PASSWORD")
	}

	dbName, err := getEnv("PGDATABASE")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGDATABASE")
	}

	sslModeRaw, err := getEnv("PGSSLMODE")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "PGSSLMODE")
	}
	sslMode := strings.ToLower(sslModeRaw)

	if err := validateSSLMode(sslMode); err != nil {
		return nil, err
	}

	if sslMode == "disable" {
		fmt.Println("WARNING: SSL is disabled - not recommended for production!")
	}

	salt := os.Getenv("AUTH_SALT")
	if salt == "" {
		return nil, fmt.Errorf("AUTH_SALT is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:          host,
			Port:          port,
			User:          user,
			Password:      password,
			AdminUser:     adminUser,
			AdminPassword: adminPassword,
			DBName:        dbName,
			SSLMode:       sslMode,
		},
		Auth: AuthConfig{
			Salt:      salt,
			JWTSecret: jwtSecret,
			JWTExpiry: 24 * time.Hour,
		},
	}, nil
}

func validateSSLMode(mode string) error {
	validModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid SSL mode: %s (allowed: disable, allow, prefer, require, verify-ca, verify-full)", mode)
	}
	return nil
}

func getEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%w: %s is not set", ErrWithEnv, key)
	}
	return value, nil
}
