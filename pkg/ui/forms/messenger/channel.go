package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// ChannelForm represents a form for creating/editing a channel
type ChannelForm struct {
	Name        string `form:"name" validate:"required,min=1,max=100"`
	Description string `form:"description" validate:"max=500"`
	IsPrivate   bool   `form:"is_private"`
	form.Submission
}
