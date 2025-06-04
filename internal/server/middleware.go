package server

import (
	"context"
	"cursach/internal/pkg/auth"
	"cursach/internal/repository"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

func JWTAuthMiddleware(secret string, tokenRepo repository.TokenRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Processing request: %s %s", r.Method, r.URL.Path)
			// Для WebSocket проверяем token в query-параметре
			if strings.HasPrefix(r.URL.Path, "/api/ws/") {
				token := r.URL.Query().Get("token")
				if token == "" {
					http.Error(w, "Token required", http.StatusUnauthorized)
					return
				}

				// Проверка отозван ли токен
				revoked, err := tokenRepo.IsTokenRevoked(r.Context(), token)
				if err != nil || revoked {
					http.Error(w, "Token revoked", http.StatusUnauthorized)
					return
				}

				// Валидация токена
				claims, err := auth.ValidateToken(token, secret)
				if err != nil {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}

				// Добавляем claims в контекст
				ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Стандартная проверка для других запросов
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := splitToken[1]

			// Проверка отозван ли токен
			revoked, err := tokenRepo.IsTokenRevoked(r.Context(), tokenString)
			if err != nil || revoked {
				http.Error(w, "Token revoked", http.StatusUnauthorized)
				return
			}

			// Валидация токена
			claims, err := auth.ValidateToken(tokenString, secret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			log.Printf("Authenticated user: %s", claims.UserID)
			// Добавляем claims в контекст
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
