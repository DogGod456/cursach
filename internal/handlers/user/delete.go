package user

import (
	"errors"
	"log"
	"net/http"

	_ "cursach/internal/server"
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
	log.Println("correct")
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем currentUserID из контекста
	currentUserID, ok := r.Context().Value("user_id").(string)
	if !ok || currentUserID == "" {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestedUserID := vars["user_id"]
	log.Println("delete", requestedUserID)
	log.Println("delete", currentUserID)

	// Определяем ID пользователя для удаления
	userIDToDelete := requestedUserID
	if requestedUserID == "me" {
		userIDToDelete = currentUserID
	}

	// Проверяем права доступа (ТОЛЬКО ЗДЕСЬ)
	if currentUserID != userIDToDelete {
		http.Error(w, "You can only delete your own account", http.StatusForbidden)
		return
	}

	// Выполняем удаление (передаем ТОЛЬКО userIDToDelete)
	err := h.useCase.Execute(r.Context(), userIDToDelete)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
