package server

import (
	"context"
	"cursach/internal/pkg/auth"
	"cursach/internal/repository"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

func JWTAuthMiddleware(secret string, tokenRepo repository.TokenRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := tokenParts[1]

			// Проверяем, не отозван ли токен
			revoked, err := tokenRepo.IsTokenRevoked(r.Context(), token)
			if err != nil {
				http.Error(w, "Failed to check token status", http.StatusInternalServerError)
				return
			}
			if revoked {
				http.Error(w, "Token has been revoked", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseJWT(token, secret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Добавляем данные пользователя в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
