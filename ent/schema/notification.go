package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeMention      NotificationType = "mention"
	NotificationTypeChannelMention NotificationType = "channel_mention"
	NotificationTypeInvite        NotificationType = "invite"
	NotificationTypeDirectMessage NotificationType = "direct_message"
	NotificationTypeReaction      NotificationType = "reaction"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
	ent.Schema
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Enum("type").
			Values(
				string(NotificationTypeMention),
				string(NotificationTypeChannelMention),
				string(NotificationTypeInvite),
				string(NotificationTypeDirectMessage),
				string(NotificationTypeReaction),
			).
			Default(string(NotificationTypeMention)),
		field.String("title").
			NotEmpty().
			MaxLen(200),
		field.Text("content").
			NotEmpty(),
		field.String("link").
			Optional().
			MaxLen(500),
		field.Bool("read").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
	}
}

