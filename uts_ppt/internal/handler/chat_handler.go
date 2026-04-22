package handler

import (
	"strconv"
	"sync"

	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/service"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:    func(r *http.Request) bool { return true },
}

// Hub manages WebSocket connections per consultation room
type Hub struct {
	rooms map[uuid.UUID]map[*websocket.Conn]bool
	mu    sync.RWMutex
}

var globalHub = &Hub{rooms: make(map[uuid.UUID]map[*websocket.Conn]bool)}

func (h *Hub) join(consultationID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[consultationID] == nil {
		h.rooms[consultationID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[consultationID][conn] = true
}

func (h *Hub) leave(consultationID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[consultationID], conn)
}

func (h *Hub) broadcast(consultationID uuid.UUID, msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.rooms[consultationID] {
		conn.WriteJSON(msg)
	}
}

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(cs service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: cs}
}

// GET /api/consultations/:id/messages — Polling fallback
func (h *ChatHandler) GetMessages(c *gin.Context) {
	consultationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	msgs, total, err := h.chatService.GetMessages(consultationID, userID, page, limit)
	if err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	// Mark as read
	h.chatService.MarkRead(consultationID, userID)
	response.Paginated(c, "Messages retrieved", msgs, total, page, limit)
}

// POST /api/consultations/:id/messages — Send message (REST polling)
func (h *ChatHandler) SendMessage(c *gin.Context) {
	consultationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	msg, err := h.chatService.SendMessage(consultationID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	// Broadcast via WebSocket if anyone is connected
	globalHub.broadcast(consultationID, msg)
	response.Created(c, "Message sent", msg)
}

// GET /api/consultations/:id/ws — WebSocket upgrade
func (h *ChatHandler) WebSocket(c *gin.Context) {
	consultationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	globalHub.join(consultationID, conn)
	defer globalHub.leave(consultationID, conn)

	for {
		var req service.SendMessageRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}
		msg, err := h.chatService.SendMessage(consultationID, userID, &req)
		if err != nil {
			conn.WriteJSON(gin.H{"error": err.Error()})
			continue
		}
		globalHub.broadcast(consultationID, msg)
	}
}
