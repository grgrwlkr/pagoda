package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Reaction holds the schema definition for the Reaction entity.
type Reaction struct {
	ent.Schema
}

// Fields of the Reaction.
func (Reaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("emoji").
			NotEmpty().
			MaxLen(10),
		field.Int("message_id"),
		field.Int("user_id"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Reaction.
func (Reaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message", Message.Type).
			Field("message_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
	}
}
