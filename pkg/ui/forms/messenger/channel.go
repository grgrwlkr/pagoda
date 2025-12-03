package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// ChannelForm represents a form for creating/editing a channel
type ChannelForm struct {
	Name        string `form:"name" validate:"required,min=1,max=100"`
	Slug        string `form:"slug" validate:"omitempty,min=1,max=50"` // Optional - will be generated from name if empty
	Description string `form:"description" validate:"max=500"`
	IsPrivate   bool   `form:"is_private"`
	form.Submission
}

// ChannelUpdateForm represents a form for updating a channel (all fields optional)
type ChannelUpdateForm struct {
	Name        string `form:"name" validate:"omitempty,min=1,max=100"`
	Slug        string `form:"slug" validate:"omitempty,min=1,max=50"`
	Description string `form:"description" validate:"omitempty,max=500"`
	IsPrivate   string `form:"is_private"` // String to handle empty values
	form.Submission
}
