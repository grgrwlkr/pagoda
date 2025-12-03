package messenger

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// UserAvatar renders a user avatar
// size can be: "xs", "sm", "md", "lg"
func UserAvatar(r *ui.Request, userID int64, userName string, size string) Node {
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

	// TODO: Add avatar URL support when UserProfile is implemented
	// For now, use placeholder with initials
	return Div(
		Class(fmt.Sprintf("avatar placeholder %s", sizeClass)),
		Div(
			Class("bg-neutral text-neutral-content rounded-full flex items-center justify-center font-semibold"),
			Text(initials),
		),
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
