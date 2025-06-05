package user

import (
	"encoding/json"
	"log"
	"net/http"

	"cursach/internal/usecase/user"
)

type GetUserHandler struct {
	userManager *user.UserManager
}

func NewGetUserHandler(userManager *user.UserManager) *GetUserHandler {
	return &GetUserHandler{userManager: userManager}
}

func (h *GetUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userData, err := h.userManager.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Формируем ответ с информацией о пользователе и его чатах
	type ChatResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"` // Имя собеседника
	}

	type UserResponse struct {
		ID    string         `json:"id"`
		Login string         `json:"login"`
		Role  string         `json:"role"`
		Chats []ChatResponse `json:"chats"` // Список чатов
	}

	response := UserResponse{
		ID:    userData.ID,
		Login: userData.Login,
		Role:  userData.Role,
		Chats: make([]ChatResponse, 0, len(userData.Chats)),
	}

	// Преобразуем чаты в нужный формат
	for _, chat := range userData.Chats {
		response.Chats = append(response.Chats, ChatResponse{
			ID:   chat.Chat.ID,
			Name: chat.User.Login, // Имя собеседника
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode user response: %v", err)
	}
}
