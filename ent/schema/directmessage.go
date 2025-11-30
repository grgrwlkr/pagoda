package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DirectMessage holds the schema definition for the DirectMessage entity.
type DirectMessage struct {
	ent.Schema
}

// Fields of the DirectMessage.
func (DirectMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user1_id"),
		field.Int("user2_id"),
		field.Time("last_message_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the DirectMessage.
func (DirectMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user1", User.Type).
			Field("user1_id").
			Required().
			Unique(),
		edge.To("user2", User.Type).
			Field("user2_id").
			Required().
			Unique(),
		edge.To("messages", DirectMessageContent.Type),
	}
}

// Indexes of the DirectMessage.
func (DirectMessage) Indexes() []ent.Index {
	return []ent.Index{
		// Unique index for user1_id and user2_id combination
		// This ensures no duplicate direct message conversations
		index.Fields("user1_id", "user2_id").Unique(),
	}
}

