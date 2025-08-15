package infrastructure

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"nhooyr.io/websocket"
)

type Hub struct {
	mu sync.RWMutex
	// userID -> set of conns
	conns map[primitive.ObjectID]map[*websocket.Conn]struct{}
	// conversationID -> participant userIDs (lazy cache)
	rooms   map[primitive.ObjectID][]primitive.ObjectID
	usecase msgpkg.IMessagingUsecase
}

func NewHub(uc msgpkg.IMessagingUsecase) *Hub {
	return &Hub{
		conns:   map[primitive.ObjectID]map[*websocket.Conn]struct{}{},
		rooms:   map[primitive.ObjectID][]primitive.ObjectID{},
		usecase: uc,
	}
}

func (h *Hub) addConn(userID primitive.ObjectID, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*websocket.Conn]struct{})
	}
	h.conns[userID][c] = struct{}{}
}

func (h *Hub) removeConn(userID primitive.ObjectID, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.conns[userID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.conns, userID)
		}
	}
}

func (h *Hub) broadcastToConversation(ctx context.Context, convID primitive.ObjectID, payload any) {
	h.mu.RLock()
	participants := h.rooms[convID]
	h.mu.RUnlock()
	if len(participants) == 0 {
		return
	}
	b, _ := json.Marshal(payload)
	for _, uid := range participants {
		h.mu.RLock()
		set := h.conns[uid]
		h.mu.RUnlock()
		for c := range set {
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			_ = c.Write(wctx, websocket.MessageText, b)
			cancel()
		}
	}
}

// WSHandler upgrades the connection and listens for frames
func (h *Hub) WSHandler(c *gin.Context) {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		userIDStr = c.GetString("user_id")
	}
	if userIDStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Println("ws accept error:", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	h.addConn(uid, conn)
	defer h.removeConn(uid, conn)

	// derive a cancelable context for this connection
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// heartbeat ping loop
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pctx, pcancel := context.WithTimeout(ctx, 5*time.Second)
				_ = conn.Ping(pctx)
				pcancel()
			}
		}
	}()

	// simple read loop with per-iteration read timeout
	for {
		var frame msgpkg.MessageFrame
		rctx, rcancel := context.WithTimeout(ctx, 60*time.Second)
		_, data, err := conn.Read(rctx)
		rcancel()
		if err != nil {
			// classify close
			if cs := websocket.CloseStatus(err); cs != -1 {
				log.Printf("ws closed: user=%s code=%d", uid.Hex(), cs)
			} else {
				log.Printf("ws read error: user=%s err=%v", uid.Hex(), err)
			}
			return
		}
		if err := json.Unmarshal(data, &frame); err != nil {
			continue
		}

		switch frame.Type {
		case "message":
			convID, err := primitive.ObjectIDFromHex(frame.ConversationID)
			if err != nil {
				continue
			}
			msg, err := h.usecase.SendMessage(ctx, uid, convID, frame.Content)
			if err != nil {
				continue
			}
			// ensure room participants are tracked lazily using usecase data
			h.mu.Lock()
			if _, ok := h.rooms[convID]; !ok {
				if conv, err := h.usecase.GetConversation(ctx, convID); err == nil {
					h.rooms[convID] = conv.ParticipantIDs
				} else {
					h.rooms[convID] = []primitive.ObjectID{uid}
				}
			}
			h.mu.Unlock()
			h.broadcastToConversation(ctx, convID, gin.H{"type": "message", "message": msg})
		case "typing":
			convID, err := primitive.ObjectIDFromHex(frame.ConversationID)
			if err != nil {
				continue
			}
			// populate room if first time
			h.mu.Lock()
			if _, ok := h.rooms[convID]; !ok {
				if conv, err := h.usecase.GetConversation(ctx, convID); err == nil {
					h.rooms[convID] = conv.ParticipantIDs
				} else {
					h.rooms[convID] = []primitive.ObjectID{uid}
				}
			}
			h.mu.Unlock()
			h.broadcastToConversation(ctx, convID, gin.H{"type": "typing", "userId": uid.Hex(), "ts": time.Now().UTC()})
		}
	}
}
