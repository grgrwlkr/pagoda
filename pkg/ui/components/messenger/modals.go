package messenger

import (
	"fmt"
	"strings"

	"github.com/mikestefanello/pagoda/pkg/ui"
	"github.com/mikestefanello/pagoda/pkg/ui/forms/messenger"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ChannelCreateModal renders a modal for creating a new channel
// Параметры:
//   - r: объект запроса с контекстом (должен быть не nil)
//   - workspaceID: ID workspace для создания канала
//   - form: форма с данными (может быть nil для новой формы)
func ChannelCreateModal(r *ui.Request, workspaceID int64, form *messenger.ChannelForm) Node {
	// Проверяем, что r не nil (защита от panic)
	if r == nil {
		// Возвращаем пустой div с ошибкой, если r nil
		return Div(Text("Error: Request is nil"))
	}

	// Проверяем, что workspaceID валидный (больше 0)
	if workspaceID <= 0 {
		// Если workspaceID невалидный, возвращаем ошибку
		return Div(
			ID("channel-create-modal"),
			Class("modal"),
			Div(
				Class("modal-box"),
				Div(Text("Error: Invalid workspace ID")),
			),
		)
	}

	// Формируем путь для создания канала
	// r.Path() уже имеет встроенную защиту от паники и fallback
	createPath := r.Path("messenger.channel.create", workspaceID)

	// Создаем CSRF input заранее, чтобы избежать проблем с nil
	var csrfInput Node
	if r != nil && r.CSRF != "" {
		csrfInput = Input(
			Type("hidden"),
			Name("csrf"),
			Value(r.CSRF),
		)
	}

	// Возвращаем полное модальное окно с id, так как вставляем его в body
	// Используем <dialog> элемент для поддержки showModal() метода (рекомендуемый метод DaisyUI)
	// Согласно документации DaisyUI: https://daisyui.com/components/modal/
	// Dialog элемент можно открыть через ID.showModal() и закрыть через ID.close()
	// Используем Raw для всего dialog элемента, так как gomponents не имеет встроенной поддержки dialog
	modalContent := Div(
		Class("modal-box"),
		Form(
			Method("POST"),
			Action(createPath),
			Attr("hx-post", createPath),
			Attr("hx-target", "body"),
			Attr("hx-swap", "outerHTML"),
			// Close modal after successful creation - use Alpine.js event
			Attr("hx-on::after-request", `
				if(event.detail.xhr.status === 200) {
					const modal = document.getElementById('channel-create-modal');
					if(modal) modal.close();
				}
			`),
			// CSRF token
			func() Node {
				if csrfInput != nil {
					return csrfInput
				}
				return nil
			}(),
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
					Attr("@click", "document.getElementById('channel-create-modal').close()"),
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
					func() Node {
						inputAttrs := []Node{
							Type("text"),
							Name("name"),
							Class("input input-bordered w-full"),
							Placeholder("general"),
							Required(),
						}
						if form != nil && form.Name != "" {
							inputAttrs = append(inputAttrs, Value(form.Name))
						}
						return Input(inputAttrs...)
					}(),
					func() Node {
						if form != nil && form.FieldHasErrors("Name") {
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
							return Group(g)
						}
						return nil
					}(),
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
					func() Node {
						textareaAttrs := []Node{
							Name("description"),
							Class("textarea textarea-bordered w-full"),
							Placeholder("What's this channel about?"),
							Rows("3"),
						}
						if form != nil && form.Description != "" {
							textareaAttrs = append(textareaAttrs, Text(form.Description))
						}
						return Textarea(textareaAttrs...)
					}(),
				),
				// Privacy toggle
				Div(
					Class("form-control"),
					Label(
						Class("label cursor-pointer justify-start gap-3"),
						func() Node {
							checkboxAttrs := []Node{
								Type("checkbox"),
								Name("is_private"),
								Class("checkbox checkbox-primary"),
							}
							if form != nil && form.IsPrivate {
								checkboxAttrs = append(checkboxAttrs, Checked())
							}
							return Input(checkboxAttrs...)
						}(),
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
					Attr("@click", "document.getElementById('channel-create-modal').close()"),
					Text("Cancel"),
				),
				Button(
					Type("submit"),
					Class("btn btn-primary"),
					Text("Create Channel"),
				),
			),
		),
	)

	backdropForm := Form(
		Method("dialog"),
		Class("modal-backdrop"),
		Button(Text("close")),
	)

	// Возвращаем dialog элемент через Raw, так как gomponents не поддерживает dialog напрямую
	// Формируем HTML вручную для dialog элемента
	// Используем Alpine.js @click вместо onclick для закрытия при клике на backdrop
	return Raw(fmt.Sprintf(`<dialog id="channel-create-modal" class="modal" @click="if(event.target === this) this.close()">%s%s</dialog>`,
		renderNodeToString(modalContent),
		renderNodeToString(backdropForm),
	))
}

// renderNodeToString рендерит Node в строку HTML
func renderNodeToString(node Node) string {
	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		return ""
	}
	return buf.String()
}

// WorkspaceCreateModal renders a modal for creating a new workspace
func WorkspaceCreateModal(r *ui.Request, form *messenger.WorkspaceForm) Node {
	// Формируем путь для создания workspace
	// r.Path() уже имеет встроенную защиту от паники и fallback
	createPath := r.Path("messenger.workspace.create")

	// Создаем CSRF input заранее
	var csrfInput Node
	if r != nil && r.CSRF != "" {
		csrfInput = Input(
			Type("hidden"),
			Name("csrf"),
			Value(r.CSRF),
		)
	}

	modalContent := Div(
		Class("modal-box"),
		Form(
			Method("POST"),
			Action(createPath),
			Attr("hx-post", createPath),
			Attr("hx-target", "body"),
			Attr("hx-swap", "outerHTML"),
			// Close modal after successful creation - use Alpine.js event
			Attr("hx-on::after-request", `
				if(event.detail.xhr.status === 200) {
					const modal = document.getElementById('workspace-create-modal');
					if(modal) modal.close();
				}
			`),
			// CSRF token
			func() Node {
				if csrfInput != nil {
					return csrfInput
				}
				return nil
			}(),
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
					Attr("@click", "document.getElementById('workspace-create-modal').close()"),
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
					func() Node {
						inputAttrs := []Node{
							Type("text"),
							Name("name"),
							Class("input input-bordered w-full"),
							Placeholder("My Workspace"),
							Required(),
						}
						if form != nil && form.Name != "" {
							inputAttrs = append(inputAttrs, Value(form.Name))
						}
						return Input(inputAttrs...)
					}(),
					func() Node {
						if form != nil && form.FieldHasErrors("Name") {
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
							return Group(g)
						}
						return nil
					}(),
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
					func() Node {
						inputAttrs := []Node{
							Type("text"),
							Name("slug"),
							Class("input input-bordered w-full"),
							Placeholder("my-workspace"),
							Required(),
						}
						if form != nil && form.Slug != "" {
							inputAttrs = append(inputAttrs, Value(form.Slug))
						}
						return Input(inputAttrs...)
					}(),
					Div(
						Class("label"),
						Span(
							Class("label-text-alt text-base-content/60"),
							Text("Used in the workspace URL"),
						),
					),
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
					func() Node {
						textareaAttrs := []Node{
							Name("description"),
							Class("textarea textarea-bordered w-full"),
							Placeholder("What's this workspace about?"),
							Rows("3"),
						}
						if form != nil && form.Description != "" {
							textareaAttrs = append(textareaAttrs, Text(form.Description))
						}
						return Textarea(textareaAttrs...)
					}(),
				),
			),
			// Actions
			Div(
				Class("modal-action"),
				Button(
					Type("button"),
					Class("btn btn-ghost"),
					Attr("@click", "document.getElementById('workspace-create-modal').close()"),
					Text("Cancel"),
				),
				Button(
					Type("submit"),
					Class("btn btn-primary"),
					Text("Create Workspace"),
				),
			),
		),
	)

	backdropForm := Form(
		Method("dialog"),
		Class("modal-backdrop"),
		Button(Text("close")),
	)

	// Возвращаем dialog элемент через Raw, так как gomponents не поддерживает dialog напрямую
	// Используем Alpine.js @click вместо onclick для закрытия при клике на backdrop
	return Raw(fmt.Sprintf(`<dialog id="workspace-create-modal" class="modal" @click="if(event.target === this) this.close()">%s%s</dialog>`,
		renderNodeToString(modalContent),
		renderNodeToString(backdropForm),
	))
}
