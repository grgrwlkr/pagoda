package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ChannelData represents channel data for rendering
type ChannelData struct {
	ID       int64
	Slug     string
	Name     string
	IsActive bool
}

// ChannelList renders the list of channels
func ChannelList(r *ui.Request, channels []ChannelData, workspaceID int64) Node {
	channelItems := make(Group, 0, len(channels))
	for _, ch := range channels {
		channelItems = append(channelItems, channelItem(r, ch.ID, ch.Slug, ch.Name, ch.IsActive))
	}

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
				Attr("onclick", "channel_create_modal.showModal()"),
				Attr("hx-get", r.Path("messenger.channel.create.form")),
				Attr("hx-target", "#channel-create-modal"),
				Attr("hx-swap", "outerHTML"),
			),
		),
		// Channel items
		Ul(
			Class("space-y-1"),
			ID("channel-list"),
			channelItems,
		),
	)
}

func channelItem(r *ui.Request, id int64, slug, name string, isActive bool) Node {
	return Li(
		A(
			Href(r.Path("messenger.channel.view", id)),
			Class("flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-base-300 transition-colors"),
			If(isActive, Class("bg-base-300")),
			Attr("hx-boost", "true"),
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

// DMData represents direct message data for rendering
type DMData struct {
	ID       int64
	UserID   int64
	UserName string
	IsActive bool
}

// DirectMessagesList renders the list of direct message conversations
func DirectMessagesList(r *ui.Request, dms []DMData) Node {
	dmItems := make(Group, 0, len(dms))
	for _, dm := range dms {
		dmItems = append(dmItems, dmItem(r, dm.ID, dm.UserID, dm.UserName, dm.IsActive))
	}

	return Div(
		Class("space-y-2"),
		H3(
			Class("text-sm font-semibold uppercase text-base-content/70 mb-2"),
			Text("Direct Messages"),
		),
		Ul(
			Class("space-y-1"),
			ID("dm-list"),
			dmItems,
		),
	)
}

func dmItem(r *ui.Request, id, userID int64, name string, isActive bool) Node {
	return Li(
		A(
			Href(r.Path("messenger.dm.view", id)),
			Class("flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-base-300 transition-colors"),
			If(isActive, Class("bg-base-300")),
			Attr("hx-boost", "true"),
			UserAvatar(r, userID, name, "xs"),
			Span(
				Class("flex-1 truncate"),
				Text(name),
			),
		),
	)
}
