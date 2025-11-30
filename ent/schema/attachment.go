package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.String("filename").
			NotEmpty().
			MaxLen(255),
		field.String("filepath").
			NotEmpty().
			MaxLen(500),
		field.Int64("file_size").
			NonNegative(),
		field.String("mime_type").
			NotEmpty().
			MaxLen(100),
		field.Int("message_id").
			Optional().
			Nillable(),
		field.Int("dm_content_id").
			Optional().
			Nillable(),
		field.Int("uploaded_by"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Unique(),
		edge.To("dm_content", DirectMessageContent.Type).
			Field("dm_content_id").
			Unique(),
		edge.To("uploader", User.Type).
			Field("uploaded_by").
			Required().
			Unique(),
	}
}

