package handlers

import (
	chathandler "cursach/internal/handlers/chat"
	userhandler "cursach/internal/handlers/user"
	"cursach/internal/repository"
	"cursach/internal/server"
	chatusecase "cursach/internal/usecase/chat"
	userusecase "cursach/internal/usecase/user"
	"github.com/gorilla/mux"
)

// SetupRouter создает и настраивает маршрутизатор HTTP
// Параметры:
//   - chatCreator: use case для создания чатов
//   - chatDeleter: use case для удаления чатов
//   - userManager: use case для управления пользователями
//   - userDeleter: use case для удаления пользователей
//
// Возвращает: настроенный маршрутизатор Gorilla Mux
func SetupRouter(
	chatCreator *chatusecase.ChatCreator,
	chatDeleter *chatusecase.ChatDeleter,
	userManager *userusecase.UserManager,
	userDeleter *userusecase.UserDeleter,
	authUC *userusecase.Authenticator,
	jwtSecret string,
	tokenRepo repository.TokenRepository,
	logoutUC *userusecase.Logouter,
	loginUpdater *userusecase.LoginUpdater,
	wsHandler *chathandler.WSHandler,
) *mux.Router {
	r := mux.NewRouter()

	// Public routes
	authHandler := userhandler.NewAuthHandler(authUC, jwtSecret)
	r.Handle("/api/auth", authHandler).Methods("POST")                                // Вход
	r.Handle("/api/users", userhandler.NewCreateHandler(userManager)).Methods("POST") // Регистрация

	logoutHandler := userhandler.NewLogoutHandler(logoutUC)

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(server.JWTAuthMiddleware(jwtSecret, tokenRepo))

	protected.Handle("/chats", chathandler.NewCreateHandler(chatCreator)).Methods("POST")
	protected.Handle("/chats/{chat_id}", chathandler.NewDeleteHandler(chatDeleter)).Methods("DELETE")
	protected.Handle("/users/{user_id}", userhandler.NewDeleteHandler(userDeleter)).Methods("DELETE")
	protected.Handle("/logout", logoutHandler).Methods("POST")
	protected.Handle("/users/login", userhandler.NewUpdateLoginHandler(loginUpdater)).Methods("PUT")
	protected.HandleFunc("/ws/{chat_id}", wsHandler.Handle)

	return r
}
