package user

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cursach/internal/pkg/auth"
	"cursach/internal/usecase/user"
)

type AuthHandler struct {
	authUC    *user.Authenticator
	jwtSecret string
}

func NewAuthHandler(authUC *user.Authenticator, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authUC:    authUC,
		jwtSecret: jwtSecret,
	}
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Исправлено: конфликт имен (user -> authUser)
	authUser, err := h.authUC.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateJWT(authUser, h.jwtSecret, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")

	// Добавлена обработка ошибки кодирования
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode auth response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
