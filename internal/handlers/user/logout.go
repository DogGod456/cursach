package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cursach/internal/usecase/user"
)

type LogoutHandler struct {
	logoutUC *user.Logouter
}

func NewLogoutHandler(logoutUC *user.Logouter) *LogoutHandler {
	return &LogoutHandler{logoutUC: logoutUC}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем токен из заголовка Authorization
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

	// Инвалидируем токен
	err := h.logoutUC.Logout(r.Context(), token)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged out"}); err != nil {
		log.Printf("Failed to encode logout response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
