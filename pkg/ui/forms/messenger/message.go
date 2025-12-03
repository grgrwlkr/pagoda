package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// MessageForm represents a form for sending a message
// Content is optional if files are provided
type MessageForm struct {
	Content string `form:"content" validate:"omitempty"`
	form.Submission
}
