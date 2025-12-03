package messenger

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TypingIndicator renders a typing indicator showing who is typing
func TypingIndicator(r *ui.Request, userNames []string) Node {
	if len(userNames) == 0 {
		return Div()
	}

	var text string
	if len(userNames) == 1 {
		text = userNames[0] + " is typing..."
	} else if len(userNames) == 2 {
		text = userNames[0] + " and " + userNames[1] + " are typing..."
	} else {
		text = userNames[0] + " and " + fmt.Sprintf("%d others", len(userNames)-1) + " are typing..."
	}

	return Div(
		ID("typing-indicator"),
		Class("px-4 py-2 text-sm text-base-content/60 italic"),
		Text(text),
		Span(
			Class("inline-block ml-1"),
			Attr("x-data", `{
				dots: '.',
				init() {
					setInterval(() => {
						this.dots = this.dots.length >= 3 ? '.' : this.dots + '.';
					}, 500);
				}
			}`),
			Attr("x-text", "dots"),
		),
	)
}
