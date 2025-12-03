package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	"github.com/mikestefanello/pagoda/pkg/ui/forms/messenger"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ChannelCreateModal renders a modal for creating a new channel
func ChannelCreateModal(r *ui.Request, workspaceID int64, form *messenger.ChannelForm) Node {
	return Div(
		ID("channel-create-modal"),
		Class("modal"),
		Div(
			Class("modal-box"),
			Form(
				Method("POST"),
				Action(r.Path("messenger.channel.create", workspaceID)),
				Attr("hx-post", r.Path("messenger.channel.create", workspaceID)),
				Attr("hx-target", "body"),
				Attr("hx-swap", "outerHTML"),
				// CSRF token
				If(r.CSRF != "", Input(
					Type("hidden"),
					Name("csrf"),
					Value(r.CSRF),
				)),
				// Header
				Div(
					Class("flex items-center justify-between mb-4"),
					H3(
						Class("text-lg font-bold"),
						Text("Create Channel"),
					),
					Button(
						Type("button"),
						Class("btn btn-sm btn-circle btn-ghost"),
						Attr("onclick", "channel_create_modal.close()"),
						Text("✕"),
					),
				),
				// Form fields
				Div(
					Class("space-y-4"),
					// Name field
					Div(
						Label(
							Class("label"),
							Span(
								Class("label-text"),
								Text("Channel Name"),
							),
						),
						Input(
							Type("text"),
							Name("name"),
							Class("input input-bordered w-full"),
							Placeholder("general"),
							Required(),
							If(form != nil, Value(form.Name)),
						),
						If(form != nil && form.FieldHasErrors("Name"), Group(
							func() Group {
								errs := form.GetFieldErrors("Name")
								g := make(Group, len(errs))
								for i, err := range errs {
									g[i] = Div(
										Class("label"),
										Span(
											Class("label-text-alt text-error"),
											Text(err),
										),
									)
								}
								return g
							}(),
						)),
					),
					// Description field
					Div(
						Label(
							Class("label"),
							Span(
								Class("label-text"),
								Text("Description"),
							),
						),
						Textarea(
							Name("description"),
							Class("textarea textarea-bordered w-full"),
							Placeholder("What's this channel about?"),
							Rows("3"),
							If(form != nil, Text(form.Description)),
						),
					),
					// Privacy toggle
					Div(
						Class("form-control"),
						Label(
							Class("label cursor-pointer justify-start gap-3"),
							Input(
								Type("checkbox"),
								Name("is_private"),
								Class("checkbox checkbox-primary"),
								If(form != nil && form.IsPrivate, Checked()),
							),
							Span(
								Class("label-text"),
								Text("Make this channel private"),
							),
						),
					),
				),
				// Actions
				Div(
					Class("modal-action"),
					Button(
						Type("button"),
						Class("btn btn-ghost"),
						Attr("onclick", "channel_create_modal.close()"),
						Text("Cancel"),
					),
					Button(
						Type("submit"),
						Class("btn btn-primary"),
						Text("Create Channel"),
					),
				),
			),
		),
		Form(
			Method("dialog"),
			Class("modal-backdrop"),
			Button(Text("close")),
		),
		Attr("onclick", "if(event.target === this) this.close()"),
	)
}

// WorkspaceCreateModal renders a modal for creating a new workspace
func WorkspaceCreateModal(r *ui.Request, form *messenger.WorkspaceForm) Node {
	return Div(
		ID("workspace-create-modal"),
		Class("modal"),
		Div(
			Class("modal-box"),
			Form(
				Method("POST"),
				Action(r.Path("messenger.workspace.create")),
				Attr("hx-post", r.Path("messenger.workspace.create")),
				Attr("hx-target", "body"),
				Attr("hx-swap", "outerHTML"),
				// CSRF token
				If(r.CSRF != "", Input(
					Type("hidden"),
					Name("csrf"),
					Value(r.CSRF),
				)),
				// Header
				Div(
					Class("flex items-center justify-between mb-4"),
					H3(
						Class("text-lg font-bold"),
						Text("Create Workspace"),
					),
					Button(
						Type("button"),
						Class("btn btn-sm btn-circle btn-ghost"),
						Attr("onclick", "workspace_create_modal.close()"),
						Text("✕"),
					),
				),
				// Form fields
				Div(
					Class("space-y-4"),
					// Name field
					Div(
						Label(
							Class("label"),
							Span(
								Class("label-text"),
								Text("Workspace Name"),
							),
						),
						Input(
							Type("text"),
							Name("name"),
							Class("input input-bordered w-full"),
							Placeholder("My Workspace"),
							Required(),
							If(form != nil, Value(form.Name)),
						),
						If(form != nil && form.FieldHasErrors("Name"), Group(
							func() Group {
								errs := form.GetFieldErrors("Name")
								g := make(Group, len(errs))
								for i, err := range errs {
									g[i] = Div(
										Class("label"),
										Span(
											Class("label-text-alt text-error"),
											Text(err),
										),
									)
								}
								return g
							}(),
						)),
					),
					// Slug field
					Div(
						Label(
							Class("label"),
							Span(
								Class("label-text"),
								Text("URL Slug"),
							),
						),
						Input(
							Type("text"),
							Name("slug"),
							Class("input input-bordered w-full"),
							Placeholder("my-workspace"),
							Required(),
							If(form != nil, Value(form.Slug)),
						),
						Div(
							Class("label"),
							Span(
								Class("label-text-alt text-base-content/60"),
								Text("Used in the workspace URL"),
							),
						),
						// Error messages would be shown here if form validation fails
					),
					// Description field
					Div(
						Label(
							Class("label"),
							Span(
								Class("label-text"),
								Text("Description"),
							),
						),
						Textarea(
							Name("description"),
							Class("textarea textarea-bordered w-full"),
							Placeholder("What's this workspace about?"),
							Rows("3"),
							If(form != nil, Text(form.Description)),
						),
					),
				),
				// Actions
				Div(
					Class("modal-action"),
					Button(
						Type("button"),
						Class("btn btn-ghost"),
						Attr("onclick", "workspace_create_modal.close()"),
						Text("Cancel"),
					),
					Button(
						Type("submit"),
						Class("btn btn-primary"),
						Text("Create Workspace"),
					),
				),
			),
		),
		Form(
			Method("dialog"),
			Class("modal-backdrop"),
			Button(Text("close")),
		),
		Attr("onclick", "if(event.target === this) this.close()"),
	)
}
