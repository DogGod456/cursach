package models

import "time"

// RevokedToken представляет модель отозванного JWT токена
type RevokedToken struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	RevokedAt time.Time `json:"revoked_at"`
}
