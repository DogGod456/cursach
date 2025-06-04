package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// HashPassword создает SHA-256 хеш пароля
func HashPassword(password, salt string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyPassword проверяет соответствие пароля хешу
func VerifyPassword(password, salt, hash string) bool {
	hashed, err := HashPassword(password, salt)
	if err != nil {
		return false
	}
	return hashed == hash
}
