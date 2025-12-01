package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageInput renders the message input form at the bottom of the chat
func MessageInput(r *ui.Request, channelID int64) Node {
	return Div(
		Class("border-t border-base-300 p-4 bg-base-100"),
		Form(
			Class("flex gap-2"),
			ID("message-form"),
			// TODO: Add WebSocket connection and message sending
			// TODO: Add HTMX for form submission
			Div(
				Class("flex-1"),
				Textarea(
					ID("message-input"),
					Name("content"),
					Class("textarea textarea-bordered w-full resize-none"),
					Placeholder("Type a message..."),
					Rows("1"),
					Attr("x-data", `{
						resize() {
							this.$el.style.height = "auto";
							this.$el.style.height = this.$el.scrollHeight + "px";
						}
					}`),
					Attr("@input", "resize()"),
					Attr("@keydown.enter", "if(!event.shiftKey) { event.preventDefault(); document.getElementById('message-form').requestSubmit(); }"),
				),
			),
			Div(
				Class("flex flex-col gap-2"),
				// File upload button
				Button(
					Type("button"),
					Class("btn btn-circle btn-ghost"),
					Title("Upload file"),
					Text("📎"),
					// TODO: Add file upload handler
				),
				// Send button
				Button(
					Type("submit"),
					Class("btn btn-primary btn-circle"),
					Title("Send message"),
					Text("➤"),
				),
			),
		),
	)
}
