package layouts

import (
	. "github.com/mikestefanello/pagoda/app/ui/components/messenger"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "github.com/mikestefanello/pagoda/pkg/ui/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Messenger creates a three-panel layout for the messenger interface
func Messenger(r *ui.Request, content Node) Node {
	return Doctype(
		HTML(
			Lang("en"),
			Data("theme", "dark"),
			Head(
				Metatags(r),
				CSS(),
				JS(),
			),
			Body(
				Class("h-screen overflow-hidden"),
				Div(
					Class("flex h-full"),
					// Left sidebar - Workspaces, Channels, DMs
					Sidebar(r),
					// Center panel - Messages
					Div(
						Class("flex-1 flex flex-col"),
						content,
					),
					// Right panel - Channel info (optional, can be toggled)
					Div(
						ID("right-panel"),
						Class("hidden lg:block w-80 border-l border-base-300 bg-base-200"),
						// Right panel content will be added via HTMX or components
					),
				),
				HtmxListeners(r),
			),
		),
	)
}
