package chat

import (
	"cursach/internal/usecase/chat"
	"encoding/json"
	"net/http"
)

type GetChatsHandler struct {
	chatLister *chat.ChatLister
}

func NewGetChatsHandler(chatLister *chat.ChatLister) *GetChatsHandler {
	return &GetChatsHandler{chatLister: chatLister}
}

func (h *GetChatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	chats, err := h.chatLister.Execute(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
