package messenger

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Thread renders the thread view page
func Thread(ctx echo.Context, channelID int64, parent messengerComponents.MessageData, replies []messengerComponents.MessageData) error {
	r := ui.NewRequest(ctx)
	r.Title = "Thread"

	content := Div(
		Class("flex flex-col h-full"),
		// Thread header
		Div(
			Class("border-b border-base-300 p-4 bg-base-100"),
			Div(
				Class("flex items-center gap-2 mb-2"),
				A(
					Href(r.Path(routenames.MessengerChannelView, channelID)),
					Class("btn btn-sm btn-ghost"),
					Text("← Back to channel"),
				),
				H1(
					Class("text-xl font-bold"),
					Text("Thread"),
				),
			),
		),
		// Thread content
		Div(
			Class("flex-1 overflow-y-auto p-4"),
			Div(
				Class("max-w-4xl mx-auto space-y-4"),
				// Parent message
				Div(
					Class("bg-base-200 rounded-lg p-4"),
					messengerComponents.MessageItem(r, parent),
				),
				// Replies section
				Div(
					ID("thread-replies"),
					Class("ml-8 border-l-2 border-base-300 pl-4 space-y-4"),
					If(len(replies) > 0,
						Div(
							Class("text-sm font-semibold text-base-content/70 mb-2"),
							Text(fmt.Sprintf("%d %s", len(replies), pluralize(len(replies), "reply", "replies"))),
						),
					),
					Group(renderThreadReplies(r, replies, parent.ID)),
				),
				// Reply form
				Div(
					Class("ml-8 mt-4"),
					renderThreadReplyForm(r, parent.ID),
				),
			),
		),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func renderThreadReplies(r *ui.Request, replies []messengerComponents.MessageData, parentID int64) Group {
	group := make(Group, 0, len(replies))
	for _, reply := range replies {
		group = append(group,
			Div(
				Class("bg-base-100 rounded-lg p-3 mb-2"),
				messengerComponents.MessageItem(r, reply),
			),
		)
	}
	return group
}

func renderThreadReplyForm(r *ui.Request, messageID int64) Node {
	return Form(
		Class("flex gap-2"),
		Method("POST"),
		Action(r.Path(routenames.MessengerMessageReply, messageID)),
		Attr("hx-post", r.Path(routenames.MessengerMessageReply, messageID)),
		Attr("hx-target", "#thread-replies"),
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
				Class("textarea textarea-bordered w-full resize-none"),
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
