package chat

import (
	"context"
	"cursach/internal/models"
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
	messageRepo repository.MessageRepository
	messageUC   *message.Sender
	connections map[string]map[*websocket.Conn]bool // chatID -> connections
	mu          sync.Mutex
}

func NewWSHandler(
	jwtSecret string,
	tokenRepo repository.TokenRepository,
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	messageUC *message.Sender,
) *WSHandler {
	return &WSHandler{
		jwtSecret:   jwtSecret,
		tokenRepo:   tokenRepo,
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		messageUC:   messageUC,
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *WSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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

	// Загрузка истории сообщений
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

func (h *WSHandler) sendHistory(conn *websocket.Conn, chatID string) error {
	messages, err := h.messageRepo.GetByChat(context.Background(), chatID, 50)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if err := conn.WriteJSON(msg); err != nil {
			return err
		}
	}
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

		var input struct{ Text string }
		if err := json.Unmarshal(msgBytes, &input); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		msg, err := h.messageUC.Execute(context.Background(), chatID, userID, input.Text)
		if err != nil {
			log.Printf("Message processing failed: %v", err)
			continue
		}

		h.broadcastMessage(chatID, msg)
	}
}

func (h *WSHandler) broadcastMessage(chatID string, msg *models.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.connections[chatID]
	if !ok {
		return
	}

	for conn := range conns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Broadcast failed: %v", err)
		}
	}
}
