package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkspaceRole represents the role of a user in a workspace.
type WorkspaceRole string

const (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
)

// WorkspaceMember holds the schema definition for the WorkspaceMember entity.
type WorkspaceMember struct {
	ent.Schema
}

// Fields of the WorkspaceMember.
func (WorkspaceMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int("workspace_id"),
		field.Int("user_id"),
		field.Enum("role").
			Values(string(WorkspaceRoleOwner), string(WorkspaceRoleAdmin), string(WorkspaceRoleMember)).
			Default(string(WorkspaceRoleMember)),
		field.Time("joined_at").
			Default(time.Now),
	}
}

// Edges of the WorkspaceMember.
func (WorkspaceMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workspace", Workspace.Type).
			Field("workspace_id").
			Required().
			Unique(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique(),
	}
}

