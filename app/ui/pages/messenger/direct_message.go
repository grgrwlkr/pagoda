package messenger

import (
	"github.com/labstack/echo/v4"
	messengerComponents "github.com/mikestefanello/pagoda/app/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/app/ui/layouts"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// DirectMessage renders the direct message page
func DirectMessage(ctx echo.Context, dmID int64, otherUserName string) error {
	r := ui.NewRequest(ctx)
	r.Title = otherUserName

	// TODO: Load messages from database
	messages := []messengerComponents.MessageData{
		// Example messages - will be replaced with real data
	}

	content := Div(
		Class("flex flex-col h-full"),
		// DM header
		dmHeader(r, dmID, otherUserName),
		// Messages list
		messengerComponents.MessageList(r, messages),
		// Typing indicator (will be updated via WebSocket)
		Div(
			ID("typing-indicator-container"),
			// TypingIndicator will be added via WebSocket updates
		),
		// Message input
		messengerComponents.MessageInput(r, dmID), // Using dmID as channelID for now
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func dmHeader(r *ui.Request, dmID int64, userName string) Node {
	return Div(
		Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
		Div(
			Class("flex items-center gap-3"),
			messengerComponents.UserAvatar(r, dmID, userName, "sm"),
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
