package models

import "time"

// User представляет модель пользователя в системе
type User struct {
	ID        string    `json:"id"`         // Уникальный идентификатор пользователя
	Login     string    `json:"login"`      // Логин пользователя (уникальный)
	Password  string    `json:"-"`          // Хэш пароля (не экспортируется в JSON)
	Role      string    `json:"role"`       // Роль пользователя (user/admin)
	CreatedAt time.Time `json:"created_at"` // Время создания пользователя
	UpdatedAt time.Time `json:"updated_at"` // Время последнего обновления (опционально)
}
