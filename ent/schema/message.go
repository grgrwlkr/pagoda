package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MessageType represents the type of message.
type MessageType string

const (
	MessageTypeText        MessageType = "text"
	MessageTypeFile        MessageType = "file"
	MessageTypeThreadReply MessageType = "thread_reply"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.Text("content").
			NotEmpty(),
		field.Enum("message_type").
			Values(string(MessageTypeText), string(MessageTypeFile), string(MessageTypeThreadReply)).
			Default(string(MessageTypeText)),
		field.Int("channel_id"),
		field.Int("user_id"),
		field.Int("thread_id").
			Optional().
			Nillable(),
		field.Int("reply_count").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("edited_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
		edge.To("thread", Message.Type).
			Field("thread_id").
			Unique(),
		edge.From("replies", Message.Type).
			Ref("thread"),
		edge.To("reactions", Reaction.Type),
		edge.To("attachments", Attachment.Type),
	}
}
