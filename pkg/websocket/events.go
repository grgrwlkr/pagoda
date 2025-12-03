package websocket

// ============================================================================
// Slack Messenger WebSocket Events
// ============================================================================
// This file defines event types and structures for WebSocket communication.
//
// File location: pkg/websocket/
//
// ============================================================================

import "encoding/json"

// EventType represents the type of WebSocket event.
type EventType string

const (
	// Client events (from client to server)
	EventTypeJoinChannel  EventType = "join_channel"
	EventTypeLeaveChannel EventType = "leave_channel"
	EventTypeTypingStart  EventType = "typing_start"
	EventTypeTypingStop   EventType = "typing_stop"
	EventTypeMessageSend  EventType = "message_send"
	EventTypeMarkRead     EventType = "mark_read"

	// Server events (from server to client)
	EventTypeMessageNew      EventType = "message_new"
	EventTypeMessageEdited   EventType = "message_edited"
	EventTypeMessageDeleted  EventType = "message_deleted"
	EventTypeUserTyping      EventType = "user_typing"
	EventTypeUserOnline      EventType = "user_online"
	EventTypeUserOffline     EventType = "user_offline"
	EventTypeReactionAdded   EventType = "reaction_added"
	EventTypeReactionRemoved EventType = "reaction_removed"
	EventTypeChannelUpdated  EventType = "channel_updated"
	EventTypeMemberJoined    EventType = "member_joined"
	EventTypeMemberLeft      EventType = "member_left"
	EventTypeError           EventType = "error"
)

// Event represents a WebSocket event.
type Event struct {
	Type EventType              `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// ToJSON converts the event to JSON bytes.
func (e *Event) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// FromJSON parses an event from JSON bytes.
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// MessageNewEvent creates a message_new event.
func MessageNewEvent(messageID, channelID, userID int64, content string) *Event {
	return &Event{
		Type: EventTypeMessageNew,
		Data: map[string]interface{}{
			"message_id": messageID,
			"channel_id": channelID,
			"user_id":    userID,
			"content":    content,
		},
	}
}

// UserTypingEvent creates a user_typing event.
func UserTypingEvent(userID, channelID int64) *Event {
	return &Event{
		Type: EventTypeUserTyping,
		Data: map[string]interface{}{
			"user_id":    userID,
			"channel_id": channelID,
		},
	}
}

// UserOnlineEvent creates a user_online event.
func UserOnlineEvent(userID int64) *Event {
	return &Event{
		Type: EventTypeUserOnline,
		Data: map[string]interface{}{
			"user_id": userID,
		},
	}
}

// UserOfflineEvent creates a user_offline event.
func UserOfflineEvent(userID int64) *Event {
	return &Event{
		Type: EventTypeUserOffline,
		Data: map[string]interface{}{
			"user_id": userID,
		},
	}
}

// ReactionAddedEvent creates a reaction_added event.
func ReactionAddedEvent(messageID, userID int64, emoji string) *Event {
	return &Event{
		Type: EventTypeReactionAdded,
		Data: map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
			"emoji":      emoji,
		},
	}
}

// ChannelUpdatedEvent creates a channel_updated event.
func ChannelUpdatedEvent(channelID int64) *Event {
	return &Event{
		Type: EventTypeChannelUpdated,
		Data: map[string]interface{}{
			"channel_id": channelID,
		},
	}
}

// MemberJoinedEvent creates a member_joined event.
func MemberJoinedEvent(channelID, userID int64) *Event {
	return &Event{
		Type: EventTypeMemberJoined,
		Data: map[string]interface{}{
			"channel_id": channelID,
			"user_id":    userID,
		},
	}
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
