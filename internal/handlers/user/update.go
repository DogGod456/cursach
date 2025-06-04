package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"cursach/internal/server"
	"cursach/internal/usecase/user"
)

type UpdateLoginHandler struct {
	updateUC *user.LoginUpdater
}

func NewUpdateLoginHandler(updateUC *user.LoginUpdater) *UpdateLoginHandler {
	return &UpdateLoginHandler{updateUC: updateUC}
}

type UpdateLoginRequest struct {
	NewLogin string `json:"new_login"`
}

func (h *UpdateLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем текущего пользователя из контекста
	currentUserID, ok := r.Context().Value(server.UserIDKey).(string)
	if !ok || currentUserID == "" {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req UpdateLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewLogin == "" {
		http.Error(w, "New login is required", http.StatusBadRequest)
		return
	}

	// Обновляем логин
	err := h.updateUC.UpdateLogin(r.Context(), currentUserID, req.NewLogin)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrLoginAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
