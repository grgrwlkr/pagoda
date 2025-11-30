package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserStatus represents the online status of a user.
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusAway    UserStatus = "away"
	UserStatusBusy    UserStatus = "busy"
	UserStatusOffline UserStatus = "offline"
)

// UserProfile holds the schema definition for the UserProfile entity.
// This extends the User entity with Slack-specific fields.
type UserProfile struct {
	ent.Schema
}

// Fields of the UserProfile.
func (UserProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").
			Unique(),
		field.String("avatar_url").
			Optional().
			MaxLen(500).
			StorageKey("avatar_url"),
		field.Enum("status").
			Values(string(UserStatusOnline), string(UserStatusAway), string(UserStatusBusy), string(UserStatusOffline)).
			Default(string(UserStatusOffline)),
		field.String("status_message").
			Optional().
			MaxLen(100),
		field.String("timezone").
			Optional().
			MaxLen(50),
	}
}

// Edges of the UserProfile.
func (UserProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
	}
}
