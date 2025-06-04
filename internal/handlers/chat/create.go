package chat

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"cursach/internal/usecase/chat"
)

// CreateHandler обрабатывает HTTP запросы для создания новых чатов
type CreateHandler struct {
	useCase *chat.ChatCreator
}

// NewCreateHandler создает новый экземпляр CreateHandler
func NewCreateHandler(useCase *chat.ChatCreator) *CreateHandler {
	return &CreateHandler{useCase: useCase}
}

// CreateRequest представляет структуру запроса на создание чата
type CreateRequest struct {
	UserIDs []string `json:"user_ids"`
}

// CreateResponse представляет структуру ответа после создания чата
type CreateResponse struct {
	ChatID string `json:"chat_id"`
}

// ServeHTTP обрабатывает HTTP запрос для создания нового чата
// Метод: POST
// Параметры: JSON с массивом user_ids
// Возвращает: JSON с chat_id или сообщение об ошибке
func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	chatID, err := h.useCase.Execute(r.Context(), req.UserIDs)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrEmptyUsers),
			errors.Is(err, chat.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, chat.ErrUserCheckFailed),
			errors.Is(err, chat.ErrChatCreation):
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
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
