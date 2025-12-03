package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// MessageForm represents a form for sending a message
type MessageForm struct {
	Content string `form:"content" validate:"required,min=1"`
	form.Submission
}
