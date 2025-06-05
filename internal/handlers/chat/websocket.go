package chat

import (
	"context"
	"cursach/internal/pkg/auth"
	"cursach/internal/repository"
	"cursach/internal/usecase/message"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 // 1KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Для разработки
	},
}

type WSHandler struct {
	jwtSecret   string
	tokenRepo   repository.TokenRepository
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository // Добавлен UserRepository
	messageRepo repository.MessageRepository
	messageUC   *message.Sender
	connections map[string]map[*websocket.Conn]bool // chatID -> connections
	mu          sync.Mutex
}

func NewWSHandler(
	jwtSecret string,
	tokenRepo repository.TokenRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository, // Добавлен UserRepository
	messageRepo repository.MessageRepository,
	messageUC *message.Sender,
) *WSHandler {
	return &WSHandler{
		jwtSecret:   jwtSecret,
		tokenRepo:   tokenRepo,
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		messageUC:   messageUC,
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *WSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection requested")
	conn, err := upgrader.Upgrade(w, r, nil)
	log.Println("conn", conn)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Аутентификация
	claims, err := h.authenticate(r)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "Auth failed"))
		return
	}

	// Проверка доступа к чату
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4002, "Chat ID not provided"))
		return
	}

	if !h.validateChatAccess(claims.UserID, chatID) {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4003, "Access denied"))
		return
	}

	// Регистрация соединения
	h.registerConnection(chatID, conn)
	defer h.unregisterConnection(chatID, conn)

	// Запускаем горутину для отправки пингов
	go h.keepAlive(conn)

	// Отправляем информацию о чате
	h.sendChatInfo(conn, chatID, claims.UserID)

	// Загружаем и отправляем историю сообщений
	if err := h.sendHistory(conn, chatID); err != nil {
		log.Printf("Failed to send history: %v", err)
	}

	// Обработка входящих сообщений
	h.handleMessages(conn, chatID, claims.UserID)
}

func (h *WSHandler) authenticate(r *http.Request) (*auth.Claims, error) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return nil, errors.New("missing token")
	}

	// Проверяем, не отозван ли токен
	revoked, err := h.tokenRepo.IsTokenRevoked(context.Background(), tokenString)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, errors.New("token revoked")
	}

	return auth.ValidateToken(tokenString, h.jwtSecret)
}

func (h *WSHandler) validateChatAccess(userID, chatID string) bool {
	isMember, err := h.chatRepo.IsUserInChat(context.Background(), chatID, userID)
	return err == nil && isMember
}

func (h *WSHandler) registerConnection(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.connections[chatID]; !ok {
		h.connections[chatID] = make(map[*websocket.Conn]bool)
	}
	h.connections[chatID][conn] = true
}

func (h *WSHandler) unregisterConnection(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.connections[chatID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.connections, chatID)
		}
	}
}

func (h *WSHandler) keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) sendChatInfo(conn *websocket.Conn, chatID, userID string) {
	// Получаем пользователей чата
	users, err := h.chatRepo.GetChatUsers(context.Background(), chatID)
	if err != nil {
		log.Printf("Failed to get chat users: %v", err)
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "Failed to get chat info",
		})
		return
	}

	// Находим собеседника (исключая текущего пользователя)
	var interlocutorName string
	for _, user := range users {
		if user.ID != userID {
			interlocutorName = user.Login
			break
		}
	}

	if interlocutorName == "" {
		interlocutorName = "Unknown"
	}

	// Отправляем информацию о чате
	conn.WriteJSON(map[string]interface{}{
		"type": "chat_info",
		"name": interlocutorName,
	})
}

func (h *WSHandler) sendHistory(conn *websocket.Conn, chatID string) error {
	// Получаем историю сообщений
	messages, err := h.messageRepo.GetByChat(context.Background(), chatID, 500)
	if err != nil {
		return err
	}

	// Отправляем историю
	conn.WriteJSON(map[string]interface{}{
		"type":     "history",
		"messages": messages,
	})

	return nil
}

func (h *WSHandler) handleMessages(conn *websocket.Conn, chatID, userID string) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var input struct {
			Type string `json:"type"`
			Text string `json:"text,omitempty"`
		}

		if err := json.Unmarshal(msgBytes, &input); err != nil {
			log.Printf("Invalid message format: %v", err)
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"message": "Invalid message format",
			})
			continue
		}

		switch input.Type {
		case "message":
			// Обработка нового сообщения
			if input.Text == "" {
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": "Message text is empty",
				})
				continue
			}

			msg, err := h.messageUC.Execute(context.Background(), chatID, userID, input.Text)
			if err != nil {
				log.Printf("Message processing failed: %v", err)
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": "Failed to send message",
				})
				continue
			}

			// Добавляем логин отправителя
			user, err := h.userRepo.GetUserByID(context.Background(), userID)
			if err == nil && user != nil {
				msg.Login = user.Login
			}

			// Рассылаем сообщение всем участникам чата
			h.broadcastMessage(chatID, map[string]interface{}{
				"type":    "message",
				"message": msg,
			})

		default:
			log.Printf("Unknown message type: %s", input.Type)
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"message": "Unknown message type",
			})
		}
	}
}

func (h *WSHandler) broadcastMessage(chatID string, msg interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.connections[chatID]
	if !ok {
		return
	}

	for conn := range conns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Broadcast failed: %v", err)
			conn.Close()
			delete(conns, conn)
		}
	}
}
