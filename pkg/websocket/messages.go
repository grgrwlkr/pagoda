package websocket

// ============================================================================
// Slack Messenger WebSocket Message Types
// ============================================================================
// This file defines message structures for WebSocket communication.
//
// File location: pkg/websocket/
//
// ============================================================================

import "time"

// Message represents a chat message structure.
type Message struct {
	ID        int64      `json:"id"`
	Content   string     `json:"content"`
	ChannelID int64      `json:"channel_id"`
	UserID    int64      `json:"user_id"`
	UserName  string     `json:"user_name,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
}

// TypingIndicator represents a typing indicator.
type TypingIndicator struct {
	UserID    int64  `json:"user_id"`
	ChannelID int64  `json:"channel_id"`
	UserName  string `json:"user_name,omitempty"`
}

// ReadReceipt represents a read receipt.
type ReadReceipt struct {
	UserID    int64     `json:"user_id"`
	ChannelID int64     `json:"channel_id"`
	ReadAt    time.Time `json:"read_at"`
}

// UserStatus represents a user's online status.
type UserStatus struct {
	UserID int64  `json:"user_id"`
	Status string `json:"status"` // online, away, busy, offline
}

// ReactionUpdate represents a reaction addition or removal.
type ReactionUpdate struct {
	MessageID int64  `json:"message_id"`
	UserID    int64  `json:"user_id"`
	Emoji     string `json:"emoji"`
	Action    string `json:"action"` // "add" or "remove"
}

// ChannelUpdate represents a channel update.
type ChannelUpdate struct {
	ChannelID   int64  `json:"channel_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// MemberUpdate represents a member joining or leaving.
type MemberUpdate struct {
	ChannelID int64  `json:"channel_id"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name,omitempty"`
	Action    string `json:"action"` // "join" or "leave"
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
