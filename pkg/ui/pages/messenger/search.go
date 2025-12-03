package messenger

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SearchPage renders the search results page
func SearchPage(ctx echo.Context, query string, workspaceID int64, messages []messengerComponents.MessageSearchResult, users []messengerComponents.UserSearchResult) error {
	r := ui.NewRequest(ctx)
	r.Title = "Search Results"

	content := Div(
		Class("flex flex-col h-full"),
		// Search header
		Div(
			Class("border-b border-base-300 p-4 bg-base-100"),
			H1(
				Class("text-2xl font-bold mb-2"),
				Text("Search Results"),
			),
			Div(
				Class("text-sm text-base-content/60"),
				Text(fmt.Sprintf("Searching for: \"%s\"", query)),
			),
		),
		// Search results
		Div(
			Class("flex-1 overflow-y-auto p-4"),
			messengerComponents.SearchResults(r, messages, users, query),
		),
	)

	return r.Render(messengerLayouts.Messenger, content)
}
