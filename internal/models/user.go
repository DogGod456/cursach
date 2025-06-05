package models

import (
	"database/sql"
	"time"
)

// User представляет модель пользователя в системе
type User struct {
	ID        string         `json:"id"`         // Уникальный идентификатор пользователя
	Login     string         `json:"login"`      // Логин пользователя (уникальный)
	Password  string         `json:"-"`          // Хэш пароля (не экспортируется в JSON)
	Role      string         `json:"role"`       // Роль пользователя (user/admin)
	CreatedAt time.Time      `json:"created_at"` // Время создания пользователя
	UpdatedAt sql.NullTime   `json:"updated_at"` // Время последнего обновления (опционально)
	Chats     []ChatWithUser `json:"chats,omitempty"`
}

// ChatWithUser представляет чат с информацией о собеседнике
type ChatWithUser struct {
	Chat Chat `json:"chat"`
	User User `json:"user"`
}
