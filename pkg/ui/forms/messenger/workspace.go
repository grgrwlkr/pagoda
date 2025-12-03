package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// WorkspaceForm represents a form for creating/editing a workspace
type WorkspaceForm struct {
	Name        string `form:"name" validate:"required,min=1,max=100"`
	Slug        string `form:"slug" validate:"required,min=1,max=50"`
	Description string `form:"description" validate:"max=500"`
	form.Submission
}
