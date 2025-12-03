package messenger

import (
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// WorkspaceCreate renders a page for creating the first workspace
func WorkspaceCreate(ctx echo.Context) error {
	r := ui.NewRequest(ctx)
	r.Title = "Create Workspace"

	// Формируем путь для создания workspace
	createPath := "/workspace"
	func() {
		defer func() {
			_ = recover()
		}()
		if r != nil && r.Context != nil {
			if path := r.Path("messenger.workspace.create"); path != "" && path != "/" {
				createPath = path
			}
		}
	}()

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
						Required(),
						Attr("autofocus"),
					),
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
						Rows("3"),
					),
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
