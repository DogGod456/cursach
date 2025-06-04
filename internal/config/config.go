package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrWithEnv = errors.New("environment variable error")
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
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

	return &Config{
		Database: DatabaseConfig{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			DBName:   dbName,
			SSLMode:  sslMode,
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
