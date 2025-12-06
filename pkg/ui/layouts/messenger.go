package layouts

import (
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "github.com/mikestefanello/pagoda/pkg/ui/components"
	. "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Messenger creates a three-panel layout for the messenger interface
func Messenger(r *ui.Request, content Node) Node {
	// Get sidebar data from context
	var sidebarData SidebarData
	if data := r.Context.Get(context.MessengerSidebarKey); data != nil {
		if sd, ok := data.(SidebarData); ok {
			sidebarData = sd
		}
	}

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
				// Alpine.js store for thread panel management (replaces global window.* functions)
				ThreadPanelStore(),
				Div(
					Class("flex h-full"),
					// Left sidebar - Workspaces, Channels, DMs
					Sidebar(r, sidebarData),
					// Center panel - Messages
					Div(
						Class("flex-1 flex flex-col"),
						content,
					),
					// Right panel - Thread panel (opens when thread is selected)
					Div(
						ID("right-panel"),
						// Base classes: hidden by default, fixed width, border, background
						// Mobile: overlay/modal (fixed, full height, z-index)
						// Desktop: sidebar (static, border-left)
						Class("hidden w-80 border-l border-base-300 bg-base-200"),
						Class("lg:relative lg:block"),                          // Desktop: relative positioning, visible when not hidden
						Class("fixed inset-y-0 right-0 z-50 lg:static"),        // Mobile: fixed overlay, Desktop: static
						Class("transition-transform duration-300 ease-in-out"), // Smooth slide animation
						Class("transform translate-x-full lg:translate-x-0"),   // Mobile: slide from right, Desktop: no transform
						Attr("data-thread-id", ""),                             // Will be set when thread panel is opened
						// Right panel content will be added via HTMX when thread is opened
					),
					// Mobile overlay backdrop (only visible on mobile when panel is open)
					Div(
						ID("right-panel-backdrop"),
						Class("hidden fixed inset-0 bg-black/50 z-40 lg:hidden"),
						Class("transition-opacity duration-300 ease-in-out"),
						// Use Alpine.js store method instead of onclick
						Attr("@click", "$store.threadPanel.closeThreadPanel()"),
					),
				),
			),
			HtmxListeners(r),
		),
	)
}
