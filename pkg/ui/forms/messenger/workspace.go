package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/form"
)

// WorkspaceForm represents a form for creating/editing a workspace
type WorkspaceForm struct {
	Name        string `form:"name" validate:"required,min=1,max=100"`
	Slug        string `form:"slug" validate:"omitempty,min=1,max=50"` // Optional - will be generated from name if empty
	Description string `form:"description" validate:"max=500"`
	form.Submission
}

// WorkspaceUpdateForm represents a form for updating a workspace (all fields optional)
type WorkspaceUpdateForm struct {
	Name        string `form:"name" validate:"omitempty,min=1,max=100"`
	Slug        string `form:"slug" validate:"omitempty,min=1,max=50"`
	Description string `form:"description" validate:"omitempty,max=500"`
	form.Submission
}
