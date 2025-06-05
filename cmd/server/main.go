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
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func(db *database.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	log.Println("Database connection established")

	// Инициализация схемы БД
	if err := db.InitSchema(); err != nil {
		log.Printf("Warning: schema initialization: %v", err)
	}

	// Инициализация репозиториев
	chatRepo := repository.NewChatRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	tokenRepo := repository.NewTokenRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)

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
