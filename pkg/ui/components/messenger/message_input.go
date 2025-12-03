package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageInput renders the message input form at the bottom of the chat
// channelIDOrDMID: channel ID for channels, DM ID for direct messages
// isDM: true if this is for a direct message, false for a channel
func MessageInput(r *ui.Request, channelIDOrDMID int64, isDM bool) Node {
	var actionRoute string
	if isDM {
		actionRoute = r.Path("messenger.dm.message.create", channelIDOrDMID)
	} else {
		actionRoute = r.Path("messenger.message.create", channelIDOrDMID)
	}

	return Div(
		Class("border-t border-base-300 p-4 bg-base-100"),
		Form(
			Class("flex gap-2"),
			ID("message-form"),
			Method("POST"),
			Action(actionRoute),
			Attr("hx-post", actionRoute),
			Attr("hx-target", "#message-list"),
			Attr("hx-swap", "beforeend"),
			Attr("hx-on::after-request", "this.querySelector('textarea').value = ''; this.querySelector('textarea').style.height = 'auto';"),
			// CSRF token
			If(r.CSRF != "", Input(
				Type("hidden"),
				Name("csrf"),
				Value(r.CSRF),
			)),
			Div(
				Class("flex-1"),
				Textarea(
					ID("message-input"),
					Name("content"),
					Class("textarea textarea-bordered w-full resize-none"),
					Placeholder("Type a message..."),
					Rows("1"),
					Required(),
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
					ID("message-send"),
					Class("btn btn-primary btn-circle"),
					Title("Send message"),
					Text("➤"),
				),
			),
		),
	)
}
