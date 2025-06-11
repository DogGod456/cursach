package handlers

import (
	chathandler "cursach/internal/handlers/chat"
	userhandler "cursach/internal/handlers/user"
	"cursach/internal/repository"
	"cursach/internal/server"
	chatusecase "cursach/internal/usecase/chat"
	userusecase "cursach/internal/usecase/user"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupRouter создает и настраивает маршрутизатор HTTP Вынести, надо посмотреть как сделат не больше 5 методов, посмотреть как через конфиг это организовать yaml
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
	chatLister *chatusecase.ChatLister,
	userSearcher *userusecase.UserSearcher,
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

	r.HandleFunc("/ws/{chat_id}", wsHandler.Handle)
	protected.Handle("/chats", chathandler.NewCreateHandler(chatCreator)).Methods("POST")
	protected.Handle("/chats", chathandler.NewGetChatsHandler(chatLister)).Methods("GET")
	protected.Handle("/chats/{chat_id}", chathandler.NewDeleteHandler(chatDeleter)).Methods("DELETE")
	protected.Handle("/user", userhandler.NewGetUserHandler(userManager)).Methods("GET")
	// protected.Handle("/users/me", userhandler.NewDeleteHandler(userDeleter)).Methods("DELETE")
	protected.Handle("/users/{user_id}", userhandler.NewDeleteHandler(userDeleter)).Methods("DELETE")
	protected.Handle("/logout", logoutHandler).Methods("POST")
	protected.Handle("/users/search", userhandler.NewSearchUsersHandler(userSearcher)).Methods("GET")
	protected.Handle("/users/login", userhandler.NewUpdateLoginHandler(loginUpdater)).Methods("PUT")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	return r
}
