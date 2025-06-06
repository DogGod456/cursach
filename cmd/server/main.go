package main

import (
	"cursach/internal/config"
	"cursach/internal/database"
	"cursach/internal/handlers"
	wbs "cursach/internal/handlers/chat"
	"cursach/internal/repository"
	"cursach/internal/server"
	"cursach/internal/usecase/chat"
	"cursach/internal/usecase/message"
	"cursach/internal/usecase/user"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// Загружаем .env файл перед запуском приложения
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Подключение к базе данных
	adminDB, err := database.New(cfg.Database, true)
	if err != nil {
		log.Fatalf("Admin DB connection failed: %v", err)
	}
	defer adminDB.Close()

	if err := adminDB.InitSchema(); err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}

	// Подключение для обычных операций
	userDB, err := database.New(cfg.Database, false)
	if err != nil {
		log.Fatalf("User DB connection failed: %v", err)
	}
	defer userDB.Close()

	log.Println("Database connection established")

	// Инициализация схемы БД
	if err := userDB.InitSchema(); err != nil {
		log.Printf("Warning: schema initialization: %v", err)
	}

	// Инициализация репозиториев
	chatRepo := repository.NewChatRepository(userDB.DB)
	userRepo := repository.NewUserRepository(userDB.DB)
	tokenRepo := repository.NewTokenRepository(userDB.DB)
	messageRepo := repository.NewMessageRepository(adminDB.DB)

	// Конфигурация аутентификации
	salt := cfg.Auth.Salt
	jwtSecret := cfg.Auth.JWTSecret

	// Инициализация use cases
	chatCreator := chat.NewChatCreator(chatRepo, userRepo)
	chatDeleter := chat.NewChatDeleter(chatRepo)
	chatLister := chat.NewChatLister(chatRepo)
	userManager := user.NewUserManager(userRepo, salt)
	userDeleter := user.NewUserDeleter(userRepo)
	userSearcher := user.NewUserSearcher(userRepo)
	authUC := user.NewAuthenticator(userRepo, salt)
	logoutUC := user.NewLogouter(tokenRepo)
	loginUpdater := user.NewLoginUpdater(userRepo)
	messageUC := message.NewSender(chatRepo, messageRepo)

	// WebSocket Handler
	wsHandler := wbs.NewWSHandler(
		jwtSecret,
		tokenRepo,
		chatRepo,
		userRepo,
		messageRepo,
		messageUC,
	)

	// Настройка маршрутов
	router := handlers.SetupRouter(
		chatCreator,
		chatDeleter,
		userManager,
		userDeleter,
		authUC,
		jwtSecret,
		tokenRepo,
		logoutUC,
		loginUpdater,
		wsHandler,
		chatLister,
		userSearcher,
	)

	// Запуск сервера
	server.Start(router, ":8080")
}
