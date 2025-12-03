package messenger

import (
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Workspace renders the main workspace page (default view when no channel is selected)
func Workspace(ctx echo.Context) error {
	r := ui.NewRequest(ctx)
	r.Title = "Workspace"

	content := Div(
		Class("flex flex-col items-center justify-center h-full p-8"),
		Div(
			Class("text-center max-w-md"),
			H1(
				Class("text-3xl font-bold mb-4"),
				Text("Welcome to your workspace"),
			),
			P(
				Class("text-base-content/70 mb-6"),
				Text("Select a channel from the sidebar to start messaging, or create a new channel to get started."),
			),
			Div(
				Class("flex gap-4 justify-center"),
				Button(
					Class("btn btn-primary"),
					Text("Create Channel"),
					Attr("hx-get", r.Path("messenger.channel.create.form")),
					Attr("hx-target", "body"),
					Attr("hx-swap", "beforeend"),
				),
			),
		),
	)

	return r.Render(messengerLayouts.Messenger, content)
}
