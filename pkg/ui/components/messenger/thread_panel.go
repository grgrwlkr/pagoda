package messenger

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ThreadPanel renders the thread panel that appears on the right side of the messenger.
// It shows the parent message, all replies, and a form to add new replies.
// Parameters:
//   - r: UI request with context
//   - parent: Parent message data
//   - replies: Array of reply messages
//   - hasMoreReplies: Whether there are more replies to load
//   - nextPage: Next page number for loading older replies (0 if no more)
//   - threadID: Thread ID for pagination URLs
//
// Returns:
//   - HTML node with the thread panel
func ThreadPanel(r *ui.Request, parent MessageData, replies []MessageData, hasMoreReplies bool, nextPage int, threadID int64) Node {
	return Div(
		Class("flex flex-col h-full bg-base-200"),
		Attr("data-thread-id", fmt.Sprintf("%d", parent.ID)), // Store thread ID for WebSocket filtering
		// Thread panel header
		Div(
			Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
			Div(
				Class("flex items-center gap-2"),
				H3(
					Class("text-lg font-semibold"),
					Text("Thread"),
				),
			),
			// Close button
			Button(
				Class("btn btn-ghost btn-sm btn-circle"),
				Attr("onclick", "closeThreadPanel();"),
				Text("✕"),
			),
		),
		// Thread content - scrollable area
		Div(
			ID("thread-panel-content"),
			Class("flex-1 overflow-y-auto p-4 space-y-4"),
			// Parent message
			Div(
				Class("bg-base-100 rounded-lg p-4"),
				MessageItem(r, parent),
			),
			// Replies section with ID for HTMX target
			Div(
				ID("thread-panel-replies"),
				Class("ml-8 border-l-2 border-base-300 pl-4 space-y-4"),
				// Load more button (for older replies) - shown at the top
				If(hasMoreReplies && nextPage > 0,
					Div(
						ID("thread-load-more-container"),
						Class("mb-4 flex justify-center"),
						Button(
							Type("button"),
							Class("btn btn-sm btn-ghost"),
							Attr("hx-get", fmt.Sprintf("%s?page=%d", r.Path("messenger.message.thread.panel", threadID), nextPage)),
							Attr("hx-target", "#thread-panel-replies"),
							Attr("hx-swap", "afterbegin"),                                                   // Insert at the beginning (older replies go on top)
							Attr("hx-select", "#thread-panel-replies > *:not(#thread-load-more-container)"), // Select only replies, not buttons
							Attr("hx-on::after-request", `
								// Remove the load more button if we've loaded all replies
								// The new content will include updated pagination info
								var loadMoreBtn = document.getElementById('thread-load-more-container');
								if (loadMoreBtn && event.detail.xhr.status === 200) {
									// Check if new content has load more button
									var newContent = event.detail.xhr.responseText;
									if (!newContent.includes('thread-load-more-container')) {
										// No more replies to load, remove button
										setTimeout(() => {
											if (loadMoreBtn) loadMoreBtn.remove();
										}, 100);
									}
								}
							`),
							Text("Load older replies"),
						),
					),
				),
				If(len(replies) > 0,
					Div(
						Class("text-sm font-semibold text-base-content/70 mb-2"),
						Text(fmt.Sprintf("%d %s", len(replies), pluralize(len(replies), "reply", "replies"))),
					),
				),
				Group(renderThreadReplies(r, replies)),
			),
		),
		// Reply form at the bottom
		Div(
			Class("border-t border-base-300 p-4 bg-base-100"),
			renderThreadPanelReplyForm(r, parent.ID),
		),
	)
}

// renderThreadReplies renders all reply messages in the thread
func renderThreadReplies(r *ui.Request, replies []MessageData) Group {
	group := make(Group, 0, len(replies))
	for _, reply := range replies {
		group = append(group, renderThreadReplyItem(r, reply))
	}
	return group
}

// renderThreadReplyItem renders a single reply item wrapped in a div
// This is used both for initial rendering and for HTMX responses
func renderThreadReplyItem(r *ui.Request, reply MessageData) Node {
	return Div(
		Class("bg-base-100 rounded-lg p-3 mb-2"),
		MessageItem(r, reply),
	)
}

// renderThreadPanelReplyForm creates a form for replying in the thread panel
func renderThreadPanelReplyForm(r *ui.Request, messageID int64) Node {
	return Form(
		Class("flex gap-2"),
		Method("POST"),
		Action(r.Path("messenger.message.reply", messageID)),
		Attr("hx-post", r.Path("messenger.message.reply", messageID)),
		Attr("hx-target", "#thread-panel-replies"),
		Attr("hx-swap", "beforeend"),
		Attr("hx-on::after-request", `
			this.querySelector('textarea').value = '';
			this.querySelector('textarea').style.height = 'auto';
			// Scroll to new reply after adding it
			if (typeof scrollToNewReply === 'function') {
				setTimeout(scrollToNewReply, 100);
			}
		`),
		// CSRF token handling: update token from main form before submit
		// Use both onsubmit and htmx:before-request to ensure token is updated
		Attr("onsubmit", `
			// Get CSRF token from main message form and update this form's token
			var mainForm = document.querySelector('#message-form');
			if (mainForm) {
				var mainCSRF = mainForm.querySelector('input[name="csrf"]');
				if (mainCSRF && mainCSRF.value) {
					// Update the hidden input in this form to match the main form's token
					// This ensures the token matches the cookie set by the server
					var panelCSRF = this.querySelector('input[name="csrf"]');
					if (panelCSRF) {
						panelCSRF.value = mainCSRF.value;
					}
				}
			}
			return true; // Allow form submission to continue
		`),
		Attr("hx-on::htmx:before-request", `
			// Also update token in HTMX request parameters
			var mainForm = document.querySelector('#message-form');
			if (mainForm) {
				var mainCSRF = mainForm.querySelector('input[name="csrf"]');
				if (mainCSRF && mainCSRF.value) {
					// Update the hidden input in this form
					var panelCSRF = event.target.querySelector('input[name="csrf"]');
					if (panelCSRF) {
						panelCSRF.value = mainCSRF.value;
					}
					// Also update in request parameters
					if (event.detail && event.detail.parameters) {
						event.detail.parameters['csrf'] = mainCSRF.value;
					}
				}
			}
		`),
		// CSRF token in form (will be updated by JavaScript to match main form)
		If(r.CSRF != "", Input(
			Type("hidden"),
			Name("csrf"),
			ID("thread-panel-csrf"), // Add ID for easier access
			Value(r.CSRF),
		)),
		Div(
			Class("flex-1"),
			Textarea(
				Name("content"),
				Class("textarea textarea-bordered w-full resize-none text-sm"),
				Placeholder("Write a reply..."),
				Rows("2"),
				Required(),
				Attr("x-data", `{
					resize() {
						this.$el.style.height = "auto";
						this.$el.style.height = this.$el.scrollHeight + "px";
					}
				}`),
				Attr("@input", "resize()"),
			),
		),
		Div(
			Class("flex flex-col gap-2"),
			Button(
				Type("submit"),
				Class("btn btn-primary btn-sm"),
				Text("Reply"),
			),
		),
	)
}
