package user

import (
	"cursach/internal/usecase/user"
	"encoding/json"
	_ "github.com/gorilla/mux"
	"net/http"
)

type SearchUsersHandler struct {
	userSearcher *user.UserSearcher
}

func NewSearchUsersHandler(userSearcher *user.UserSearcher) *SearchUsersHandler {
	return &SearchUsersHandler{userSearcher: userSearcher}
}

func (h *SearchUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	if login == "" {
		http.Error(w, "login parameter is required", http.StatusBadRequest)
		return
	}

	users, err := h.userSearcher.Execute(r.Context(), login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
