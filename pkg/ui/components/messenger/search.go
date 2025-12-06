package messenger

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SearchBox renders a search input box for the sidebar
func SearchBox(r *ui.Request, workspaceID int64) Node {
	return Div(
		Class("p-4 border-b border-base-300"),
		// Alpine.js для управления видимостью результатов поиска
		Attr("x-data", `{
			open: false,
			showResults(status) {
				this.open = (status === 200);
			}
		}`),
		Div(
			Class("relative"),
			Input(
				Type("search"),
				ID("search-input"),
				Name("q"),
				Class("input input-bordered w-full pl-10"),
				Placeholder("Search messages..."),
				Attr("hx-get", r.Path(routenames.MessengerSearchMessages)),
				Attr("hx-trigger", "keyup changed delay:500ms, search"),
				Attr("hx-target", "#search-results"),
				Attr("hx-swap", "innerHTML"),
				Attr("hx-include", "[name='workspace_id']"),
				Attr("@focus", "open = true"),
				Attr("@click", "open = true"),
				// Show results when HTMX updates content
				Attr("hx-on::after-request", "showResults(event.detail.xhr.status)"),
			),
			// Search icon
			Span(
				Class("absolute left-3 top-1/2 transform -translate-y-1/2 text-base-content/50"),
				Text("🔍"),
			),
			// Hidden input for workspace_id
			Input(
				Type("hidden"),
				Name("workspace_id"),
				Value(fmt.Sprintf("%d", workspaceID)),
			),
		),
		// Search results container - visibility controlled by Alpine.js
		Div(
			ID("search-results"),
			Class("absolute z-50 w-64 mt-2 bg-base-100 border border-base-300 rounded-lg shadow-lg max-h-96 overflow-y-auto"),
			// Use Alpine.js x-show instead of classList manipulation
			Attr("x-show", "open"),
			Style("display: none;"), // Hidden by default
		),
	)
}

// SearchResults renders search results
func SearchResults(r *ui.Request, messages []MessageSearchResult, users []UserSearchResult, query string) Node {
	group := make(Group, 0)

	// Messages section
	if len(messages) > 0 {
		group = append(group,
			Div(
				Class("p-2 border-b border-base-300"),
				H4(
					Class("text-sm font-semibold text-base-content/70"),
					Text("Messages"),
				),
			),
		)

		for _, msg := range messages {
			group = append(group, SearchMessageItem(r, msg, query))
		}
	}

	// Users section
	if len(users) > 0 {
		group = append(group,
			Div(
				Class("p-2 border-b border-base-300"),
				H4(
					Class("text-sm font-semibold text-base-content/70"),
					Text("Users"),
				),
			),
		)

		for _, user := range users {
			group = append(group, SearchUserItem(r, user))
		}
	}

	// No results
	if len(messages) == 0 && len(users) == 0 {
		group = append(group,
			Div(
				Class("p-4 text-center text-base-content/50"),
				Text("No results found"),
			),
		)
	}

	return Div(
		Class("p-2"),
		group,
	)
}

// MessageSearchResult represents a message search result (exported for use in handlers)
type MessageSearchResult struct {
	MessageID   int64
	Content     string
	ChannelID   int64
	ChannelName string
	UserID      int
	UserName    string
	CreatedAt   time.Time
}

// UserSearchResult represents a user search result (exported for use in handlers)
type UserSearchResult struct {
	UserID    int
	Name      string
	Email     string
	AvatarURL string
}

// SearchMessageItem renders a single message search result
func SearchMessageItem(r *ui.Request, msg MessageSearchResult, query string) Node {
	return A(
		Href(r.Path(routenames.MessengerChannelView, msg.ChannelID)),
		Class("block p-3 hover:bg-base-200 border-b border-base-300"),
		Div(
			Class("flex items-start gap-2"),
			Div(
				Class("flex-1 min-w-0"),
				Div(
					Class("flex items-center gap-2 mb-1"),
					Span(
						Class("font-semibold text-sm"),
						Text(msg.UserName),
					),
					Span(
						Class("text-xs text-base-content/60"),
						Text("#"+msg.ChannelName),
					),
					Span(
						Class("text-xs text-base-content/40"),
						Text(msg.CreatedAt.Format("Jan 2, 15:04")),
					),
				),
				Div(
					Class("text-sm text-base-content/90 line-clamp-2"),
					// Highlight search query in content
					Raw(highlightText(msg.Content, query)),
				),
			),
		),
	)
}

// highlightText highlights the search query in the text
func highlightText(text, query string) string {
	if query == "" {
		return html.EscapeString(text)
	}

	// Escape HTML in text first
	escapedText := html.EscapeString(text)

	// Simple case-insensitive highlighting
	lowerText := strings.ToLower(escapedText)
	lowerQuery := strings.ToLower(query)

	if !strings.Contains(lowerText, lowerQuery) {
		return escapedText
	}

	// Find and replace (case-insensitive)
	result := ""
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerQuery)
		if idx == -1 {
			result += escapedText[start:]
			break
		}
		idx += start
		// Extract the original case from escaped text
		matchedText := escapedText[idx : idx+len(query)]
		result += escapedText[start:idx] + "<mark class='bg-warning text-warning-content'>" + matchedText + "</mark>"
		start = idx + len(query)
	}

	return result
}

// SearchUserItem renders a single user search result
func SearchUserItem(r *ui.Request, user UserSearchResult) Node {
	return A(
		Href("#"), // TODO: Link to user profile or DM
		Class("block p-3 hover:bg-base-200 border-b border-base-300"),
		Div(
			Class("flex items-center gap-3"),
			UserAvatar(r, int64(user.UserID), user.Name, "sm", nil),
			Div(
				Class("flex-1 min-w-0"),
				Div(
					Class("font-medium truncate"),
					Text(user.Name),
				),
				Div(
					Class("text-xs text-base-content/60 truncate"),
					Text(user.Email),
				),
			),
		),
	)
}
