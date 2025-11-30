package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DirectMessageContent holds the schema definition for the DirectMessageContent entity.
type DirectMessageContent struct {
	ent.Schema
}

// Fields of the DirectMessageContent.
func (DirectMessageContent) Fields() []ent.Field {
	return []ent.Field{
		field.Text("content").
			NotEmpty(),
		field.Int("dm_id"),
		field.Int("user_id"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("edited_at").
			Optional().
			Nillable(),
	}
}

// Edges of the DirectMessageContent.
func (DirectMessageContent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dm", DirectMessage.Type).
			Field("dm_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
		edge.To("attachments", Attachment.Type),
	}
}

