package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// InviteForm represents a form for inviting a user
type InviteForm struct {
	UserID int64  `form:"user_id" validate:"required"`
	Email  string `form:"email" validate:"required,email"`
	form.Submission
}

// AddMemberForm represents a form for adding a member to workspace/channel
type AddMemberForm struct {
	UserID int    `form:"user_id" validate:"required"`
	Role   string `form:"role" validate:"omitempty,oneof=owner admin member"` // Optional, defaults to member
	form.Submission
}

// ReactionForm represents a form for adding a reaction
type ReactionForm struct {
	Emoji string `form:"emoji" validate:"required,min=1,max=50"`
	form.Submission
}
