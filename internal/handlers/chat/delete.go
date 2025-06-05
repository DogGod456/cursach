package chat

import (
	"errors"
	"log"
	"net/http"

	"cursach/internal/usecase/chat"
	"github.com/gorilla/mux"
)

// DeleteHandler обрабатывает HTTP запросы для удаления чатов
type DeleteHandler struct {
	useCase *chat.ChatDeleter
}

// NewDeleteHandler создает новый экземпляр DeleteHandler
func NewDeleteHandler(useCase *chat.ChatDeleter) *DeleteHandler {
	return &DeleteHandler{useCase: useCase}
}

// ServeHTTP обрабатывает HTTP запрос для удаления чата
func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	log.Println(vars)
	chatID := vars["chat_id"]

	// Получаем userID из контекста (установлено в JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if chatID == "" {
		http.Error(w, "Missing chat_id parameter", http.StatusBadRequest)
		return
	}
	log.Println(chatID)
	err := h.useCase.Execute(r.Context(), chatID, userID)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrUserNotInChat),
			errors.Is(err, chat.ErrChatNotFound):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, chat.ErrChatDeletion):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Отправка успешного ответа без содержимого
	w.WriteHeader(http.StatusNoContent)
}
