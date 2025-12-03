package messenger

import (
	"fmt"
	"time"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageList renders the list of messages in a channel or DM
func MessageList(r *ui.Request, messages []MessageData) Node {
	return Div(
		ID("message-list"),
		Class("flex-1 overflow-y-auto p-4 space-y-4"),
		Group(renderMessages(r, messages)),
	)
}

type MessageData struct {
	ID          int64
	Content     string
	UserID      int64
	UserName    string
	CreatedAt   time.Time
	EditedAt    *time.Time
	Reactions   []ReactionData
	Attachments []FileAttachmentData
	ReplyCount  int // Number of replies in thread
}

type ReactionData struct {
	Emoji   string
	Count   int
	UserIDs []int64
}

func renderMessages(r *ui.Request, messages []MessageData) Group {
	group := make(Group, 0, len(messages))

	for _, msg := range messages {
		group = append(group, messageItem(r, msg))
	}

	return group
}

// MessageItem renders a single message item (exported for HTMX)
func MessageItem(r *ui.Request, msg MessageData) Node {
	return messageItem(r, msg)
}

func messageItem(r *ui.Request, msg MessageData) Node {
	return Div(
		Class("flex gap-3 hover:bg-base-200/50 p-2 rounded-lg transition-colors"),
		ID(fmt.Sprintf("message-%d", msg.ID)),
		// Avatar
		UserAvatar(r, msg.UserID, msg.UserName, "sm", nil), // nil = unknown online status
		// Message content
		Div(
			Class("flex-1 min-w-0"),
			// Header with name and timestamp
			Div(
				Class("flex items-center gap-2 mb-1"),
				Span(
					Class("font-semibold"),
					Text(msg.UserName),
				),
				Span(
					Class("text-xs text-base-content/60"),
					Text(msg.CreatedAt.Format("15:04")),
				),
				If(msg.EditedAt != nil,
					Span(
						Class("text-xs text-base-content/40 italic"),
						Text("(edited)"),
					),
				),
			),
			// Message content
			If(msg.Content != "",
				Div(
					Class("text-base-content/90 whitespace-pre-wrap break-words"),
					Text(msg.Content),
				),
			),
			// Attachments (if any)
			If(len(msg.Attachments) > 0,
				Div(
					Class("mt-2 space-y-2"),
					Group(renderAttachments(r, msg.Attachments)),
				),
			),
			// Reactions (if any)
			If(len(msg.Reactions) > 0,
				Div(
					Class("flex flex-wrap gap-1 mt-2"),
					Group(renderReactions(r, msg.Reactions, msg.ID)),
				),
			),
			// Thread actions
			Div(
				Class("flex items-center gap-2 mt-2 text-sm"),
				// Reply button
				Button(
					Class("btn btn-ghost btn-sm gap-1"),
					Attr("hx-get", r.Path("messenger.message.replies", msg.ID)),
					Attr("hx-target", fmt.Sprintf("#thread-%d", msg.ID)),
					Attr("hx-swap", "innerHTML"),
					Attr("onclick", fmt.Sprintf("document.getElementById('thread-reply-form-%d').classList.toggle('hidden'); return false;", msg.ID)),
					Text("💬 Reply"),
				),
				// Reply count (if any)
				If(msg.ReplyCount > 0,
					A(
						Href(r.Path("messenger.message.thread", msg.ID)),
						Class("text-base-content/60 hover:text-base-content"),
						Text(fmt.Sprintf("%d %s", msg.ReplyCount, pluralize(msg.ReplyCount, "reply", "replies"))),
					),
				),
			),
			// Thread replies container (initially hidden)
			Div(
				ID(fmt.Sprintf("thread-%d", msg.ID)),
				Class("mt-2 ml-8 border-l-2 border-base-300 pl-4"),
			),
			// Thread reply form (initially hidden)
			Div(
				ID(fmt.Sprintf("thread-reply-form-%d", msg.ID)),
				Class("hidden mt-2 ml-8"),
				renderThreadReplyForm(r, msg.ID),
			),
		),
	)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func renderThreadReplyForm(r *ui.Request, messageID int64) Node {
	return Form(
		Class("flex gap-2"),
		Method("POST"),
		Action(r.Path("messenger.message.reply", messageID)),
		Attr("hx-post", r.Path("messenger.message.reply", messageID)),
		Attr("hx-target", fmt.Sprintf("#thread-%d", messageID)),
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
			Button(
				Type("button"),
				Class("btn btn-ghost btn-sm"),
				Attr("onclick", fmt.Sprintf("document.getElementById('thread-reply-form-%d').classList.add('hidden');", messageID)),
				Text("Cancel"),
			),
		),
	)
}

func renderReactions(r *ui.Request, reactions []ReactionData, messageID int64) Group {
	group := make(Group, 0, len(reactions))

	for _, reaction := range reactions {
		group = append(group,
			Button(
				Class("btn btn-xs gap-1 hover:bg-base-300"),
				Text(reaction.Emoji),
				Text(fmt.Sprintf("%d", reaction.Count)),
				Attr("hx-post", r.Path("messenger.reaction.add", messageID)),
				Attr("hx-vals", fmt.Sprintf(`{"emoji": "%s"}`, reaction.Emoji)),
				Attr("hx-target", fmt.Sprintf("#message-%d", messageID)),
				Attr("hx-swap", "outerHTML"),
				Title("Click to toggle reaction"),
			),
		)
	}

	return group
}

func renderAttachments(r *ui.Request, attachments []FileAttachmentData) Group {
	group := make(Group, 0, len(attachments))

	for _, attachment := range attachments {
		group = append(group, FileAttachment(r, attachment))
	}

	return group
}
