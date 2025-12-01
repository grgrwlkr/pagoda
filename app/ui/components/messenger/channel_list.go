package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ChannelList renders the list of channels
func ChannelList(r *ui.Request) Node {
	return Div(
		Class("space-y-2"),
		// Header with create button
		Div(
			Class("flex items-center justify-between mb-4"),
			H3(
				Class("text-sm font-semibold uppercase text-base-content/70"),
				Text("Channels"),
			),
			Button(
				Class("btn btn-sm btn-circle btn-ghost"),
				Text("+"),
				Attr("title", "Create channel"),
				// TODO: Add click handler to open create channel modal
			),
		),
		// Channel items (will be populated from data)
		Ul(
			Class("space-y-1"),
			ID("channel-list"),
			// TODO: Load channels from database and render
			channelItem(r, 1, "general", "General", true),
			channelItem(r, 2, "random", "Random", false),
		),
	)
}

func channelItem(r *ui.Request, id int64, slug, name string, isActive bool) Node {
	return Li(
		A(
			Href("#"), // TODO: Use route name
			Class("flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-base-300 transition-colors"),
			If(isActive, Class("bg-base-300")),
			Span(
				Class("text-lg"),
				Text("#"),
			),
			Span(
				Class("flex-1 truncate"),
				Text(name),
			),
			// Unread badge (if any)
			// Span(Class("badge badge-sm badge-primary"), Text("3")),
		),
	)
}

// DirectMessagesList renders the list of direct message conversations
func DirectMessagesList(r *ui.Request) Node {
	return Div(
		Class("space-y-2"),
		H3(
			Class("text-sm font-semibold uppercase text-base-content/70 mb-2"),
			Text("Direct Messages"),
		),
		Ul(
			Class("space-y-1"),
			ID("dm-list"),
			// TODO: Load DMs from database and render
			// dmItem(r, 1, "John Doe", true),
		),
	)
}

func dmItem(r *ui.Request, id int64, name string, isActive bool) Node {
	return Li(
		A(
			Href("#"), // TODO: Use route name
			Class("flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-base-300 transition-colors"),
			If(isActive, Class("bg-base-300")),
			UserAvatar(r, id, name, "xs"),
			Span(
				Class("flex-1 truncate"),
				Text(name),
			),
		),
	)
}
