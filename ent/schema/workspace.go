package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Workspace holds the schema definition for the Workspace entity.
type Workspace struct {
	ent.Schema
}

// Fields of the Workspace.
func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.String("slug").
			NotEmpty().
			Unique().
			MaxLen(50),
		field.String("description").
			Optional().
			MaxLen(500),
		field.Int("owner_id"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Workspace.
func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("owner", User.Type).
			Field("owner_id").
			Required().
			Unique(),
		edge.To("members", WorkspaceMember.Type),
		edge.To("channels", Channel.Type),
	}
}

