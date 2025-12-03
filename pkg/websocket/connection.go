package websocket

// ============================================================================
// Slack Messenger WebSocket Connection
// ============================================================================
// This file handles individual WebSocket connections.
//
// File location: pkg/websocket/
//
// ============================================================================

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/message"
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

	// Context for database operations
	Ctx context.Context

	// ORM client for database operations
	ORM *ent.Client

	// Logger for logging events
	Logger *slog.Logger
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Connection) ReadPump() {
	if c.Logger != nil {
		c.Logger.Info("=== WEBSOCKET READ PUMP START ===", "user_id", c.UserID)
	}

	defer func() {
		if c.Logger != nil {
			c.Logger.Info("WebSocket read pump closing", "user_id", c.UserID)
		}

		// Send offline status before disconnecting
		offlineEvent := UserOfflineEvent(c.UserID)
		c.Hub.Broadcast(offlineEvent.ToJSON())

		c.Hub.unregister <- c
		c.WS.Close()

		if c.Logger != nil {
			c.Logger.Info("=== WEBSOCKET READ PUMP END ===", "user_id", c.UserID)
		}
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
				if c.Logger != nil {
					c.Logger.Warn("WebSocket unexpected close", "user_id", c.UserID, "error", err)
				}
			} else if c.Logger != nil {
				c.Logger.Debug("WebSocket read error (normal close)", "user_id", c.UserID, "error", err)
			}
			break
		}

		// Parse incoming message
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			if c.Logger != nil {
				c.Logger.Warn("Failed to parse WebSocket message", "user_id", c.UserID, "error", err, "message_length", len(message))
			}
			c.SendError("invalid_message_format", "Failed to parse message")
			continue
		}

		// Log incoming event
		if c.Logger != nil {
			c.Logger.Info("WebSocket event received", "user_id", c.UserID, "event_type", event.Type)
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
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Verify user is a member of the channel
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Notify other channel members
	response := MemberJoinedEvent(chID, c.UserID)
	c.Hub.SendToChannel(chID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("User joined channel", "user_id", c.UserID, "channel_id", chID)
	}
}

// handleLeaveChannel handles when a user leaves a channel.
func (c *Connection) handleLeaveChannel(event *Event) {
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Verify user is a member of the channel
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Notify other channel members
	response := &Event{
		Type: EventTypeMemberLeft,
		Data: map[string]interface{}{
			"channel_id": chID,
			"user_id":    c.UserID,
		},
	}
	c.Hub.SendToChannel(chID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("User left channel", "user_id", c.UserID, "channel_id", chID)
	}
}

// handleTypingStart handles typing start events.
func (c *Connection) handleTypingStart(event *Event) {
	// Broadcast typing indicator to channel members
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Verify user is a member of the channel
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil || !exists {
		// Silently ignore typing events for non-members
		return
	}

	response := UserTypingEvent(c.UserID, chID)
	c.Hub.SendToChannel(chID, response.ToJSON())
}

// handleTypingStop handles typing stop events.
func (c *Connection) handleTypingStop(event *Event) {
	// Typing stop is typically handled client-side
	// We can implement this if needed for server-side cleanup
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		return
	}

	chID := int64(channelID)

	// Verify user is a member of the channel
	// Typing stop is typically handled client-side, but we validate membership
	_, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		// Silently ignore errors for typing stop
		return
	}

	// Could send a typing_stop event if needed
	// For now, we'll just validate and return
}

// handleMessageSend handles message sending via WebSocket.
func (c *Connection) handleMessageSend(event *Event) {
	// Extract message data
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	content, ok := event.Data["content"].(string)
	if !ok || content == "" {
		c.SendError("invalid_content", "Message content is required")
		return
	}

	chID := int64(channelID)

	// Verify user is a member of the channel
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Create message in database
	msg, err := c.ORM.Message.
		Create().
		SetContent(content).
		SetMessageType(message.MessageTypeText).
		SetChannelID(int(chID)).
		SetUserID(int(c.UserID)).
		Save(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to save message", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to save message")
		return
	}

	// Broadcast message to channel members
	response := MessageNewEvent(int64(msg.ID), chID, c.UserID, content)
	c.Hub.SendToChannel(chID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("Message sent", "user_id", c.UserID, "channel_id", chID, "message_id", msg.ID)
	}
}

// handleMarkRead handles read receipt events.
func (c *Connection) handleMarkRead(event *Event) {
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Update last_read_at for the user in this channel
	_, err := c.ORM.ChannelMember.
		Update().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		SetLastReadAt(time.Now()).
		Save(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to update last_read_at", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to update read status")
		return
	}

	if c.Logger != nil {
		c.Logger.Debug("Read receipt updated", "user_id", c.UserID, "channel_id", chID)
	}
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
