package user

import (
	"errors"
	"net/http"

	"cursach/internal/server"
	"cursach/internal/usecase/user"
	"github.com/gorilla/mux"
)

// DeleteHandler обрабатывает HTTP запросы для удаления пользователей
type DeleteHandler struct {
	useCase *user.UserDeleter
}

// NewDeleteHandler создает новый экземпляр DeleteHandler
func NewDeleteHandler(useCase *user.UserDeleter) *DeleteHandler {
	return &DeleteHandler{useCase: useCase}
}

// ServeHTTP обрабатывает HTTP запрос для удаления пользователя
// Метод: DELETE
// Параметры: user_id в URL, текущий user_id из контекста JWT
// Возвращает: HTTP статус 204 при успехе или сообщение об ошибке
func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	userID := vars["user_id"]

	// Получаем currentUserID из контекста (установлено в JWT middleware)
	currentUserID, ok := r.Context().Value(server.UserIDKey).(string)
	if !ok || currentUserID == "" {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	err := h.useCase.Execute(r.Context(), userID, currentUserID)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, user.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
