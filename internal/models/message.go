package models

import (
	"database/sql"
	"time"
)

// Message представляет модель сообщения в чате
type Message struct {
	ID          string       `json:"id"`           // Уникальный идентификатор сообщения
	ChatID      string       `json:"chat_id"`      // ID чата, к которому относится сообщение
	UserID      string       `json:"user_id"`      // ID отправителя сообщения
	Login       string       `json:"login"`        // Логин отправителя
	Text        string       `json:"text"`         // Текст сообщения
	ReplyTo     string       `json:"reply_to"`     // ID сообщения, на которое дан ответ (опционально) НЕ БУДЕТ
	IsDraft     bool         `json:"is_draft"`     // Флаг черновика НЕ БУДЕТ
	SendingTime time.Time    `json:"sending_time"` // Время отправки сообщения
	UpdatedAt   sql.NullTime `json:"updated_at"`   // Время последнего обновления (опционально)
}
