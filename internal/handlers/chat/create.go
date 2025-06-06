package chat

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"cursach/internal/usecase/chat"
)

// Обработчик создания чата
type CreateHandler struct {
	useCase *chat.ChatCreator
}

// Конструктор обработчика
func NewCreateHandler(useCase *chat.ChatCreator) *CreateHandler {
	return &CreateHandler{useCase: useCase}
}

// Запрос на создание чата
type CreateRequest struct {
	UserLogin string `json:"userLogin"` // Логин пользователя для чата
}

// Ответ при успешном создании
type CreateResponse struct {
	ChatID string `json:"chat_id"`
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID текущего пользователя из контекста
	currentUserID, ok := r.Context().Value("user_id").(string)
	if !ok || currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Передаем ID текущего пользователя и логин целевого пользователя
	chatID, err := h.useCase.Execute(r.Context(), currentUserID, req.UserLogin)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, chat.ErrChatCreation):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			log.Printf("Create chat error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := CreateResponse{ChatID: chatID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode chat creation response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
