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
