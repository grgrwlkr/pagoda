package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Channel holds the schema definition for the Channel entity.
type Channel struct {
	ent.Schema
}

// Fields of the Channel.
func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.String("slug").
			NotEmpty().
			MaxLen(50),
		field.String("description").
			Optional().
			MaxLen(500),
		field.Bool("is_private").
			Default(false),
		field.Int("workspace_id"),
		field.Int("created_by"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Channel.
func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Required().
			Unique(),
		edge.To("creator", User.Type).
			Field("created_by").
			Required().
			Unique(),
		edge.To("members", ChannelMember.Type),
		edge.To("messages", Message.Type),
	}
}

