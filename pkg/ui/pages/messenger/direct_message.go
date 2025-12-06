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

// DirectMessage renders the direct message page
func DirectMessage(ctx echo.Context, dmID int64, otherUserName string, messages []messengerComponents.MessageData) error {
	r := ui.NewRequest(ctx)
	r.Title = otherUserName

	content := Div(
		Class("flex flex-col h-full"),
		// DM header
		dmHeader(r, dmID, otherUserName),
		// Messages list
		messengerComponents.MessageList(r, messages),
		// Typing indicator (managed via Alpine.js, updated via WebSocket)
		messengerComponents.TypingIndicator(),
		// Message input
		messengerComponents.MessageInput(r, dmID, true), // true = is a DM
		// WebSocket connection via HTMX (replaces JavaScript WebSocket)
		messengerComponents.WebSocketConnection(r, "/ws", 0, dmID),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func dmHeader(r *ui.Request, dmID int64, userName string) Node {
	return Div(
		Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
		Div(
			Class("flex items-center gap-3"),
			Div(
				ID(fmt.Sprintf("dm-avatar-%d", dmID)),
				messengerComponents.UserAvatar(r, dmID, userName, "sm", nil), // nil = unknown online status, will be updated via WebSocket
			),
			H2(
				Class("text-xl font-semibold"),
				Text(userName),
			),
		),
		Div(
			Class("flex items-center gap-2"),
			// User info button
			Button(
				Class("btn btn-sm btn-ghost"),
				Text("ℹ️"),
				Attr("title", "User info"),
				// TODO: Open right panel with user info
			),
		),
	)
}

// dmWebsocketScript removed - replaced by messengerComponents.WebSocketConnection() which uses HTMX WebSocket extension
