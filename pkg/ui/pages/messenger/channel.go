package messenger

import (
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Channel renders the channel page with messages
func Channel(ctx echo.Context, channelID int64, channelName string, messages []messengerComponents.MessageData) error {
	r := ui.NewRequest(ctx)
	r.Title = "#" + channelName

	content := Div(
		Class("flex flex-col h-full"),
		Attr("hx-ext", "ws"), // Enable HTMX WebSocket extension for this container
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
		messengerComponents.MessageInput(r, channelID, false), // false = not a DM
		// WebSocket connection via HTMX (replaces JavaScript WebSocket)
		messengerComponents.WebSocketConnection(r, "/ws", channelID, 0),
		// Hidden HTMX buttons for reloading thread panels (triggered by WebSocket events via Alpine.js)
		// These replace htmx.ajax() and fetch() calls
		renderWebSocketReloadButtons(r, messages),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

// renderWebSocketReloadButtons creates hidden HTMX button for reloading thread panel.
// This button is triggered by WebSocket events via Alpine.js, replacing htmx.ajax() and fetch() calls.
// The button URL is updated dynamically via Alpine.js before triggering.
func renderWebSocketReloadButtons(r *ui.Request, messages []messengerComponents.MessageData) Node {
	// Create a single universal button that can reload any thread panel
	// Alpine.js will update the hx-get attribute and trigger click
	return Button(
		ID("ws-reload-thread-button"),
		Type("button"),
		Style("display: none;"),
		Attr("hx-get", ""), // Will be set dynamically by Alpine.js
		Attr("hx-target", "#right-panel"),
		Attr("hx-swap", "innerHTML"),
		Attr("hx-on::after-request", `
			if (typeof scrollToNewReply === 'function') {
				setTimeout(scrollToNewReply, 100);
			}
		`),
	)
}

// websocketScript removed - replaced by messengerComponents.WebSocketConnection() which uses HTMX WebSocket extension

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
