package messenger

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// UserAvatar renders a user avatar with optional online status indicator
// size can be: "xs", "sm", "md", "lg"
// isOnline: optional online status (nil = unknown, true = online, false = offline)
func UserAvatar(r *ui.Request, userID int64, userName string, size string, isOnline *bool) Node {
	// Generate initials from name
	initials := getInitials(userName)

	// Size classes
	sizeClasses := map[string]string{
		"xs": "w-6 h-6 text-xs",
		"sm": "w-8 h-8 text-sm",
		"md": "w-12 h-12 text-base",
		"lg": "w-16 h-16 text-lg",
	}

	sizeClass := sizeClasses[size]
	if sizeClass == "" {
		sizeClass = sizeClasses["md"]
	}

	// Online status indicator size
	statusSize := "w-2 h-2"
	if size == "xs" {
		statusSize = "w-1.5 h-1.5"
	} else if size == "lg" {
		statusSize = "w-3 h-3"
	}

	// TODO: Add avatar URL support when UserProfile is implemented
	// For now, use placeholder with initials

	// Build online status indicator if isOnline is not nil
	var statusIndicator Node
	if isOnline != nil {
		var statusClass string
		var statusValue string
		if *isOnline {
			statusClass = "bg-success"
			statusValue = "true"
		} else {
			statusClass = "bg-base-300"
			statusValue = "false"
		}
		statusIndicator = Div(
			Class(fmt.Sprintf("absolute bottom-0 right-0 %s rounded-full border-2 border-base-100 %s", statusSize, statusClass)),
			Attr("data-online", statusValue),
		)
	}

	return Div(
		Class(fmt.Sprintf("avatar placeholder %s relative", sizeClass)),
		ID(fmt.Sprintf("user-avatar-%d", userID)),
		Div(
			Class("bg-neutral text-neutral-content rounded-full flex items-center justify-center font-semibold"),
			Text(initials),
		),
		// Online status indicator
		If(statusIndicator != nil, statusIndicator),
	)
}

func getInitials(name string) string {
	if len(name) == 0 {
		return "?"
	}

	words := []rune(name)
	if len(words) == 1 {
		return string(words[0:1])
	}

	// Get first letter of first word and first letter of last word
	first := string(words[0])
	last := ""
	for i := len(words) - 1; i >= 0; i-- {
		if words[i] != ' ' {
			last = string(words[i])
			break
		}
	}

	if last == "" {
		return first
	}

	return first + last
}
