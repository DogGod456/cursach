package database

import (
	"context"
	"database/sql"
	"fmt"
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
