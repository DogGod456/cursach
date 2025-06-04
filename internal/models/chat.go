package models

import (
	"database/sql"
	"time"
)

// Chat представляет модель чата в системе
type Chat struct {
	ID        string       `json:"id"`         // Уникальный идентификатор чата
	CreatedAt time.Time    `json:"created_at"` // Время создания чата
	UpdatedAt sql.NullTime `json:"updated_at"` // Время последнего обновления (опционально)
}
