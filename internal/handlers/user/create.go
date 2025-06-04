package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"cursach/internal/usecase/user"
)

// CreateHandler обрабатывает HTTP запросы для создания/авторизации пользователей
type CreateHandler struct {
	useCase *user.UserManager
}

// NewCreateHandler создает новый экземпляр CreateHandler
func NewCreateHandler(useCase *user.UserManager) *CreateHandler {
	return &CreateHandler{useCase: useCase}
}

// CreateRequest представляет структуру запроса для создания/авторизации пользователя
type CreateRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

// CreateResponse представляет структуру ответа после создания/авторизации пользователя
type CreateResponse struct {
	UserID string `json:"user_id"`
}

// ServeHTTP обрабатывает HTTP запрос для создания/авторизации пользователя
// Метод: POST
// Параметры: JSON с login, password и опционально role
// Возвращает: JSON с user_id или сообщение об ошибке
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

	userID, err := h.useCase.CreateOrGetUser(r.Context(), req.Login, req.Password, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmptyCredentials),
			errors.Is(err, user.ErrInvalidRole):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, user.ErrInvalidCredentials):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := CreateResponse{UserID: userID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode user creation response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
