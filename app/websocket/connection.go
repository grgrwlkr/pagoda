package websocket

// ============================================================================
// CUSTOM CODE START - Slack Messenger WebSocket Connection
// ============================================================================
// This file handles individual WebSocket connections.
//
// File location: app/websocket/
// This is YOUR code, not part of Pagoda core.
// ============================================================================

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Connection is a middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection
	WS *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// User ID associated with this connection
	UserID int64

	// Hub reference
	Hub *Hub
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.WS.Close()
	}()

	c.WS.SetReadDeadline(time.Now().Add(pongWait))
	c.WS.SetReadLimit(maxMessageSize)
	c.WS.SetPongHandler(func(string) error {
		c.WS.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.WS.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error if needed
			}
			break
		}

		// Parse incoming message
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			// Send error response
			c.SendError("invalid_message_format", "Failed to parse message")
			continue
		}

		// Handle the event
		c.handleEvent(&event)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.WS.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.WS.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.WS.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.WS.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.WS.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WS.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleEvent processes incoming events from the client.
func (c *Connection) handleEvent(event *Event) {
	switch event.Type {
	case EventTypeJoinChannel:
		c.handleJoinChannel(event)
	case EventTypeLeaveChannel:
		c.handleLeaveChannel(event)
	case EventTypeTypingStart:
		c.handleTypingStart(event)
	case EventTypeTypingStop:
		c.handleTypingStop(event)
	case EventTypeMessageSend:
		c.handleMessageSend(event)
	case EventTypeMarkRead:
		c.handleMarkRead(event)
	default:
		c.SendError("unknown_event_type", "Unknown event type: "+string(event.Type))
	}
}

// handleJoinChannel handles when a user joins a channel.
func (c *Connection) handleJoinChannel(event *Event) {
	// TODO: Implement channel join logic
	// This will validate user membership and notify others
}

// handleLeaveChannel handles when a user leaves a channel.
func (c *Connection) handleLeaveChannel(event *Event) {
	// TODO: Implement channel leave logic
}

// handleTypingStart handles typing start events.
func (c *Connection) handleTypingStart(event *Event) {
	// Broadcast typing indicator to channel members
	channelID, ok := event.Data["channel_id"].(int64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}
	response := UserTypingEvent(c.UserID, channelID)
	c.Hub.SendToChannel(channelID, response.ToJSON())
}

// handleTypingStop handles typing stop events.
func (c *Connection) handleTypingStop(event *Event) {
	// Broadcast typing stop to channel members
	// Similar to handleTypingStart
}

// handleMessageSend handles message sending via WebSocket.
func (c *Connection) handleMessageSend(event *Event) {
	// TODO: Save message to database and broadcast
	// This will be implemented when we have message handlers
}

// handleMarkRead handles read receipt events.
func (c *Connection) handleMarkRead(event *Event) {
	// TODO: Update last_read_at in database
}

// SendError sends an error message to the client.
func (c *Connection) SendError(code, message string) {
	response := &Event{
		Type: EventTypeError,
		Data: map[string]interface{}{
			"error_code":    code,
			"error_message": message,
		},
	}
	c.Send <- response.ToJSON()
}

// Close closes the connection.
func (c *Connection) Close() {
	c.WS.Close()
	close(c.Send)
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================

