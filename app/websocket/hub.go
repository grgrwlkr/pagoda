package websocket

// ============================================================================
// CUSTOM CODE START - Slack Messenger WebSocket Hub
// ============================================================================
// This file contains the WebSocket hub for real-time messaging in Slack clone.
//
// File location: app/websocket/
// This is YOUR code, not part of Pagoda core.
// ============================================================================

import (
	"context"
	"sync"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
)

// Hub maintains the set of active connections and broadcasts messages to them.
type Hub struct {
	// Registered connections, keyed by user ID
	connections map[int64]*Connection

	// Inbound messages from the connections
	broadcast chan []byte

	// Register requests from the connections
	register chan *Connection

	// Unregister requests from connections
	unregister chan *Connection

	// Mutex for thread-safe access
	mu sync.RWMutex

	// ORM client for database operations
	ORM *ent.Client
}

// Global hub instance (set by WebSocket handler)
var globalHub *Hub

// GetHub returns the global hub instance
func GetHub() *Hub {
	return globalHub
}

// SetHub sets the global hub instance
func SetHub(hub *Hub) {
	globalHub = hub
}

// NewHub creates a new Hub instance.
func NewHub(orm *ent.Client) *Hub {
	return &Hub{
		connections: make(map[int64]*Connection),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		ORM:         orm,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			// If user already has a connection, close the old one
			if oldConn, exists := h.connections[conn.UserID]; exists {
				oldConn.Close()
			}
			h.connections[conn.UserID] = conn
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.UserID]; ok {
				delete(h.connections, conn.UserID)
				close(conn.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
					delete(h.connections, conn.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser sends a message to a specific user.
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conn, ok := h.connections[userID]; ok {
		select {
		case conn.Send <- message:
		default:
			close(conn.Send)
			delete(h.connections, userID)
		}
	}
}

// SendToChannel sends a message to all users in a channel.
func (h *Hub) SendToChannel(channelID int64, message []byte) {
	// Get channel members from database
	members, err := h.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(channelID))).
		All(context.Background())

	if err != nil {
		// If we can't get members, fall back to broadcasting to all
		// This is a safety fallback, but shouldn't happen in normal operation
		h.mu.RLock()
		for _, conn := range h.connections {
			select {
			case conn.Send <- message:
			default:
				close(conn.Send)
				delete(h.connections, conn.UserID)
			}
		}
		h.mu.RUnlock()
		return
	}

	// Build set of user IDs that are members
	memberUserIDs := make(map[int64]bool)
	for _, member := range members {
		memberUserIDs[int64(member.UserID)] = true
	}

	// Send message only to channel members who are online
	h.mu.RLock()
	for userID, conn := range h.connections {
		if memberUserIDs[userID] {
			select {
			case conn.Send <- message:
			default:
				close(conn.Send)
				delete(h.connections, userID)
			}
		}
	}
	h.mu.RUnlock()
}

// Register registers a new connection.
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// GetConnectionCount returns the number of active connections.
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// IsUserOnline checks if a user is currently online.
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.connections[userID]
	return ok
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
