package messenger

import (
	"github.com/labstack/echo/v4"
	messengerComponents "github.com/mikestefanello/pagoda/app/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/app/ui/layouts"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Channel renders the channel page with messages
func Channel(ctx echo.Context, channelID int64, channelName string) error {
	r := ui.NewRequest(ctx)
	r.Title = "#" + channelName

	// TODO: Load messages from database
	messages := []messengerComponents.MessageData{
		// Example messages - will be replaced with real data
	}

	content := Div(
		Class("flex flex-col h-full"),
		// Channel header
		channelHeader(r, channelID, channelName),
		// Messages list
		messengerComponents.MessageList(r, messages),
		// Typing indicator (will be updated via WebSocket)
		Div(
			ID("typing-indicator-container"),
			// TypingIndicator will be added via WebSocket updates
		),
		// Message input
		messengerComponents.MessageInput(r, channelID),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func channelHeader(r *ui.Request, channelID int64, channelName string) Node {
	return Div(
		Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
		Div(
			Class("flex items-center gap-3"),
			H2(
				Class("text-xl font-semibold"),
				Span(Class("text-base-content/60"), Text("#")),
				Text(channelName),
			),
		),
		Div(
			Class("flex items-center gap-2"),
			// Channel info button
			Button(
				Class("btn btn-sm btn-ghost"),
				Text("ℹ️"),
				Attr("title", "Channel info"),
				// TODO: Open right panel with channel info
			),
		),
	)
}
