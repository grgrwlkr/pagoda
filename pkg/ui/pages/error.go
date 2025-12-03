package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Error(ctx echo.Context, code int) error {
	r := ui.NewRequest(ctx)
	r.Title = http.StatusText(code)
	var body Node

	switch code {
	case http.StatusInternalServerError:
		body = Text("Please try again.")
	case http.StatusForbidden, http.StatusUnauthorized:
		body = Text("You are not authorized to view the requested page.")
	case http.StatusNotFound:
		body = Group{
			Text("Click "),
			A(
				Href(r.Path(routenames.MessengerRoot)),
				Text("here"),
			),
			Text(" to return to workspace."),
		}
	default:
		body = Text("Something went wrong.")
	}

	content := Div(
		Class("flex flex-col items-center justify-center h-full p-8"),
		Div(
			Class("text-center max-w-md"),
			H1(
				Class("text-3xl font-bold mb-4"),
				Text(r.Title),
			),
			P(
				Class("text-base-content/70"),
				body,
			),
		),
	)

	// For error pages, we don't have sidebar data, so use empty SidebarData
	// The Sidebar component will handle empty data gracefully
	return r.Render(messengerLayouts.Messenger, content)
}
