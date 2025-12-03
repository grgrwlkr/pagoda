package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SidebarData contains data for rendering the sidebar
type SidebarData struct {
	WorkspaceID     int64
	WorkspaceName   string
	Channels        []ChannelData
	DirectMessages  []DMData
	ActiveChannelID *int64
	ActiveDMID      *int64
}

// Sidebar renders the left sidebar with workspaces, channels, and direct messages
func Sidebar(r *ui.Request, data SidebarData) Node {
	// Mark active channels/DMs
	channels := make([]ChannelData, len(data.Channels))
	for i, ch := range data.Channels {
		channels[i] = ch
		if data.ActiveChannelID != nil && ch.ID == *data.ActiveChannelID {
			channels[i].IsActive = true
		}
	}

	dms := make([]DMData, len(data.DirectMessages))
	for i, dm := range data.DirectMessages {
		dms[i] = dm
		if data.ActiveDMID != nil && dm.ID == *data.ActiveDMID {
			dms[i].IsActive = true
		}
	}

	return Div(
		Class("w-64 bg-base-200 border-r border-base-300 flex flex-col"),
		// Header with user profile
		Div(
			Class("p-4 border-b border-base-300"),
			userProfileSection(r),
		),
		// Workspace selector
		Div(
			Class("p-4 border-b border-base-300"),
			H3(
				Class("text-sm font-semibold uppercase text-base-content/70 mb-2"),
				Text("Workspace"),
			),
			Div(
				Class("text-base font-medium"),
				Text(data.WorkspaceName),
			),
		),
		// Channels section
		Div(
			Class("flex-1 overflow-y-auto p-4"),
			ChannelList(r, channels, data.WorkspaceID),
		),
		// Direct Messages section
		Div(
			Class("p-4 border-t border-base-300"),
			DirectMessagesList(r, dms),
		),
	)
}

func userProfileSection(r *ui.Request) Node {
	if !r.IsAuth {
		return Div(Text("Not authenticated"))
	}

	return Div(
		Class("flex items-center gap-3"),
		UserAvatar(r, int64(r.AuthUser.ID), r.AuthUser.Name, "sm"),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(r.AuthUser.Name),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(r.AuthUser.Email),
			),
		),
	)
}
