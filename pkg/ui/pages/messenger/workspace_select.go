package messenger

import (
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// WorkspaceSelect renders a page for selecting a workspace when user has multiple workspaces
func WorkspaceSelect(ctx echo.Context, workspaces []*ent.Workspace) error {
	r := ui.NewRequest(ctx)
	r.Title = "Select Workspace"

	// Формируем путь для создания нового workspace
	// r.Path() уже имеет встроенную защиту от паники и fallback
	createFormPath := r.Path("messenger.workspace.create.form")

	// Создаем список workspace карточек
	workspaceCards := make(Group, 0, len(workspaces))
	for _, ws := range workspaces {
		// Формируем путь для перехода в workspace
		// r.Path() уже имеет встроенную защиту от паники и fallback
		workspacePath := r.Path("messenger.workspace.view", ws.ID)

		workspaceCards = append(workspaceCards,
			A(
				Href(workspacePath),
				Class("card bg-base-200 hover:bg-base-300 transition-colors cursor-pointer"),
				Div(
					Class("card-body"),
					H2(
						Class("card-title"),
						Text(ws.Name),
					),
					If(ws.Description != "", P(
						Class("text-sm text-base-content/70"),
						Text(ws.Description),
					)),
					Div(
						Class("card-actions justify-end mt-4"),
						Button(
							Class("btn btn-primary btn-sm"),
							Text("Open"),
						),
					),
				),
			),
		)
	}

	content := Div(
		Class("flex flex-col h-full p-8"),
		Div(
			Class("max-w-4xl mx-auto w-full"),
			// Header
			Div(
				Class("mb-8 text-center"),
				H1(
					Class("text-3xl font-bold mb-2"),
					Text("Select Workspace"),
				),
				P(
					Class("text-base-content/70"),
					Text("Choose a workspace to continue, or create a new one."),
				),
			),
			// Workspace grid
			Div(
				Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8"),
				workspaceCards,
			),
			// Create new workspace button
			Div(
				Class("text-center"),
				Button(
					Class("btn btn-outline btn-primary"),
					Text("+ Create New Workspace"),
					Attr("onclick", "workspace_create_modal.showModal()"),
					Attr("hx-get", createFormPath),
					Attr("hx-target", "#workspace-create-modal"),
					Attr("hx-swap", "outerHTML"),
				),
			),
		),
		// Modal для создания нового workspace
		Div(
			ID("workspace-create-modal"),
			Class("modal"),
		),
	)

	return r.Render(messengerLayouts.Messenger, content)
}
