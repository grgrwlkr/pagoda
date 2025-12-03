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
	ID        int64
	Content   string
	UserID    int64
	UserName  string
	CreatedAt time.Time
	EditedAt  *time.Time
	Reactions []ReactionData
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
		UserAvatar(r, msg.UserID, msg.UserName, "sm"),
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
			Div(
				Class("text-base-content/90 whitespace-pre-wrap break-words"),
				Text(msg.Content),
			),
			// Reactions (if any)
			If(len(msg.Reactions) > 0,
				Div(
					Class("flex flex-wrap gap-1 mt-2"),
					Group(renderReactions(r, msg.Reactions, msg.ID)),
				),
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
