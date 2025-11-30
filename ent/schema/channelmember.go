package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ChannelMember holds the schema definition for the ChannelMember entity.
type ChannelMember struct {
	ent.Schema
}

// Fields of the ChannelMember.
func (ChannelMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int("channel_id"),
		field.Int("user_id"),
		field.Time("joined_at").
			Default(time.Now),
		field.Time("last_read_at").
			Optional().
			Nillable(),
	}
}

// Edges of the ChannelMember.
func (ChannelMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channel", Channel.Type).
			Field("channel_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
	}
}

