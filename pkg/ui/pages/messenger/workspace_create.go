package messenger

import (
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/form"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerForms "github.com/mikestefanello/pagoda/pkg/ui/forms/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// WorkspaceCreate renders a page for creating the first workspace
func WorkspaceCreate(ctx echo.Context) error {
	r := ui.NewRequest(ctx)
	r.Title = "Create Workspace"

	// Get form from context (will be new form or form with validation errors)
	form := form.Get[messengerForms.WorkspaceForm](ctx)

	// Get create path
	createPath := r.Path("messenger.workspace.create")
	if createPath == "" || createPath == "/" {
		createPath = "/workspace"
	}

	content := Div(
		Class("flex flex-col items-center justify-center h-full p-8"),
		Div(
			Class("text-center max-w-md w-full"),
			H1(
				Class("text-3xl font-bold mb-4"),
				Text("Welcome to Slack Clone!"),
			),
			P(
				Class("text-base-content/70 mb-8"),
				Text("Get started by creating your first workspace. A workspace is where you and your team can collaborate."),
			),
			Form(
				Method("POST"),
				Action(createPath),
				Attr("hx-post", createPath),
				Attr("hx-target", "body"),
				Attr("hx-swap", "outerHTML"),
				Class("space-y-4"),
				// CSRF token
				If(r.CSRF != "", Input(
					Type("hidden"),
					Name("csrf"),
					Value(r.CSRF),
				)),
				// Name field
				Div(
					Label(
						Class("label"),
						Span(
							Class("label-text"),
							Text("Workspace Name"),
						),
					),
					Input(
						Type("text"),
						Name("name"),
						Placeholder("My Workspace"),
						Class("input input-bordered w-full"),
						If(form.FieldHasErrors("Name"), Class("input-error")),
						If(form.Name != "", Value(form.Name)),
						Required(),
						Attr("autofocus"),
					),
					// Show validation errors
					If(form.FieldHasErrors("Name"), func() Node {
						errs := form.GetFieldErrors("Name")
						g := make(Group, len(errs))
						for i, err := range errs {
							g[i] = Div(
								Class("label"),
								Span(
									Class("label-text-alt text-error"),
									Text(err),
								),
							)
						}
						return Group(g)
					}()),
				),
				// Description field (optional)
				Div(
					Label(
						Class("label"),
						Span(
							Class("label-text"),
							Text("Description (optional)"),
						),
					),
					Textarea(
						Name("description"),
						Placeholder("A brief description of your workspace"),
						Class("textarea textarea-bordered w-full"),
						If(form.FieldHasErrors("Description"), Class("textarea-error")),
						If(form.Description != "", Text(form.Description)),
						Rows("3"),
					),
					// Show validation errors
					If(form.FieldHasErrors("Description"), func() Node {
						errs := form.GetFieldErrors("Description")
						g := make(Group, len(errs))
						for i, err := range errs {
							g[i] = Div(
								Class("label"),
								Span(
									Class("label-text-alt text-error"),
									Text(err),
								),
							)
						}
						return Group(g)
					}()),
				),
				// Submit button
				Button(
					Type("submit"),
					Class("btn btn-primary w-full"),
					Text("Create Workspace"),
				),
			),
		),
	)

	return r.Render(messengerLayouts.Messenger, content)
}
