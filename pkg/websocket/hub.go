package websocket

// ============================================================================
// Slack Messenger WebSocket Hub
// ============================================================================
// This file contains the WebSocket hub for real-time messaging in Slack clone.
// ============================================================================

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
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
	slog.Info("WebSocket hub started")
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			// If user already has a connection, close the old one
			if oldConn, exists := h.connections[conn.UserID]; exists {
				slog.Info("Closing existing connection for user", "user_id", conn.UserID)
				oldConn.Close()
			}
			h.connections[conn.UserID] = conn
			connectionCount := len(h.connections)
			h.mu.Unlock()
			slog.Info("Connection registered", "user_id", conn.UserID, "total_connections", connectionCount)

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.UserID]; ok {
				delete(h.connections, conn.UserID)
				close(conn.Send)
				connectionCount := len(h.connections)
				slog.Info("Connection unregistered", "user_id", conn.UserID, "total_connections", connectionCount)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			connectionCount := len(h.connections)
			sentCount := 0
			for _, conn := range h.connections {
				select {
				case conn.Send <- message:
					sentCount++
				default:
					close(conn.Send)
					delete(h.connections, conn.UserID)
				}
			}
			h.mu.RUnlock()
			if connectionCount > 0 {
				slog.Debug("Message broadcast", "total_connections", connectionCount, "sent_to", sentCount)
			}
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
			slog.Debug("Message sent to user", "user_id", userID)
		default:
			slog.Warn("Failed to send message to user, closing connection", "user_id", userID)
			close(conn.Send)
			delete(h.connections, userID)
		}
	} else {
		slog.Debug("User not connected, message not sent", "user_id", userID)
	}
}

// SendToWorkspace sends a message to all users in a workspace.
func (h *Hub) SendToWorkspace(ctx context.Context, workspaceID int64, message []byte) {
	slog.Info("Sending message to workspace", "workspace_id", workspaceID)

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get workspace members from the database
	slog.Debug("Loading workspace members", "workspace_id", workspaceID)
	members, err := h.ORM.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(int(workspaceID))).
		WithUser().
		All(ctx)

	if err != nil {
		slog.Error("Failed to get workspace members for SendToWorkspace", "workspace_id", workspaceID, "error", err)
		// Fallback: broadcast to all if we can't get specific members
		slog.Warn("Falling back to broadcast to all connections", "workspace_id", workspaceID)
		sentCount := 0
		for _, conn := range h.connections {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				close(conn.Send)
				delete(h.connections, conn.UserID)
			}
		}
		slog.Info("Message broadcasted to all (fallback)", "workspace_id", workspaceID, "sent_to", sentCount)
		return
	}
	slog.Info("Workspace members loaded", "workspace_id", workspaceID, "members_count", len(members))

	// Build set of user IDs that are members
	memberUserIDs := make(map[int64]bool)
	for _, member := range members {
		memberUserIDs[int64(member.UserID)] = true
	}

	// Send message only to workspace members who are online
	sentCount := 0
	for userID, conn := range h.connections {
		if memberUserIDs[userID] {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				slog.Warn("Failed to send message to workspace member, closing connection", "workspace_id", workspaceID, "user_id", userID)
				close(conn.Send)
				delete(h.connections, userID)
			}
		}
	}
	slog.Info("Message sent to workspace members", "workspace_id", workspaceID, "sent_to", sentCount, "total_members", len(members))
}

// SendToChannel sends a message to all users in a channel.
func (h *Hub) SendToChannel(channelID int64, message []byte) {
	slog.Info("Sending message to channel", "channel_id", channelID)

	// Get channel members from database
	slog.Debug("Loading channel members", "channel_id", channelID)
	members, err := h.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(channelID))).
		All(context.Background())

	if err != nil {
		// If we can't get members, fall back to broadcasting to all
		// This is a safety fallback, but shouldn't happen in normal operation
		slog.Error("Failed to get channel members, falling back to broadcast", "channel_id", channelID, "error", err)
		h.mu.RLock()
		sentCount := 0
		for _, conn := range h.connections {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				close(conn.Send)
				delete(h.connections, conn.UserID)
			}
		}
		h.mu.RUnlock()
		slog.Info("Message broadcasted to all (fallback)", "channel_id", channelID, "sent_to", sentCount)
		return
	}
	slog.Info("Channel members loaded", "channel_id", channelID, "members_count", len(members))

	// Build set of user IDs that are members
	memberUserIDs := make(map[int64]bool)
	for _, member := range members {
		memberUserIDs[int64(member.UserID)] = true
	}

	// Send message only to channel members who are online
	h.mu.RLock()
	sentCount := 0
	for userID, conn := range h.connections {
		if memberUserIDs[userID] {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				slog.Warn("Failed to send message to channel member, closing connection", "channel_id", channelID, "user_id", userID)
				close(conn.Send)
				delete(h.connections, userID)
			}
		}
	}
	h.mu.RUnlock()
	slog.Info("Message sent to channel members", "channel_id", channelID, "sent_to", sentCount, "total_members", len(members))
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
