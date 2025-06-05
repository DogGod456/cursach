package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"

	"cursach/internal/config"

	_ "github.com/lib/pq"
)

// DB обертка над sql.DB для добавления методов
type DB struct {
	*sql.DB
}

// New создает новое подключение к PostgreSQL
// Принимает конфигурацию и возвращает *DB или ошибку
func New(cfg config.DatabaseConfig) (*DB, error) {
	// Сначала проверяем существование базы данных
	exists, err := databaseExists(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	// Если базы не существует - создаем
	if !exists {
		log.Printf("Database %s does not exist, creating...", cfg.DBName)
		if err := createDatabase(cfg); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Формируем строку подключения
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем соединение с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL database")
	return &DB{db}, nil
}

// InitSchema инициализирует схему базы данных
func (db *DB) InitSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Выполняем SQL-скрипт создания схемы
	_, err := db.ExecContext(ctx, Schema)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}

	log.Println("Database schema initialized")
	return nil
}

// Close закрывает соединение с базой данных
func (db *DB) Close() error {
	if err := db.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	log.Println("Database connection closed")
	return nil
}

// databaseExists проверяет существование базы данных
func databaseExists(cfg config.DatabaseConfig) (bool, error) {
	// Подключаемся к системной базе данных
	sysConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		return false, fmt.Errorf("failed to open system database: %w", err)
	}
	defer sysDB.Close()

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sysDB.PingContext(ctx); err != nil {
		return false, fmt.Errorf("failed to ping system database: %w", err)
	}

	// Проверяем существование базы
	query := `SELECT 1 FROM pg_database WHERE datname = $1`
	var exists int
	err = sysDB.QueryRowContext(ctx, query, cfg.DBName).Scan(&exists)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("database check failed: %w", err)
	default:
		return true, nil
	}
}

// createDatabase создает новую базу данных
func createDatabase(cfg config.DatabaseConfig) error {
	// Подключаемся к системной базе данных
	sysConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode,
	)

	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		return fmt.Errorf("failed to open system database: %w", err)
	}
	defer sysDB.Close()

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sysDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping system database: %w", err)
	}

	// Создаем базу данных с экранированием имени
	query := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(cfg.DBName))
	_, err = sysDB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}
