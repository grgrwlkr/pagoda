package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	"github.com/mikestefanello/pagoda/pkg/ui/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageInput рендерит форму ввода сообщения внизу чата
// Параметры:
//   - r: объект запроса с контекстом
//   - channelIDOrDMID: ID канала (для каналов) или ID прямого сообщения (для DM)
//   - isDM: true если это прямое сообщение, false если канал
//
// Возвращает HTML-узел с формой ввода сообщения
func MessageInput(r *ui.Request, channelIDOrDMID int64, isDM bool) Node {
	var actionRoute string
	// Определяем URL для отправки сообщения в зависимости от типа (канал или DM)
	if isDM {
		// Для прямых сообщений используем специальный route
		actionRoute = r.Path("messenger.direct_message.message.create", channelIDOrDMID)
	} else {
		// Для каналов используем обычный route
		actionRoute = r.Path("messenger.message.create", channelIDOrDMID)
	}

	return Div(
		Class("border-t border-base-300 p-4 bg-base-100"), // border-t: верхняя граница; p-4: отступы; bg-base-100: цвет фона

		// Контейнер для превью загружаемых файлов (управляется через Alpine.js)
		// File API требует JavaScript для работы с файлами
		FilePreviewContainer(),

		// Основная форма для отправки сообщения
		Form(
			Class("flex gap-2"),                    // flex: горизонтальное расположение; gap-2: отступ между элементами
			ID("message-form"),                     // ID для HTMX
			Method("POST"),                         // HTTP метод отправки
			Action(actionRoute),                    // URL для отправки (fallback для браузеров без JS)
			Attr("enctype", "multipart/form-data"), // Необходимо для загрузки файлов

			// HTMX атрибуты для AJAX отправки без перезагрузки страницы
			Attr("hx-post", actionRoute),               // HTMX POST запрос
			Attr("hx-target", "#message-list"),         // Куда вставить новое сообщение
			Attr("hx-swap", "beforeend"),               // Вставить в конец списка сообщений
			Attr("hx-encoding", "multipart/form-data"), // Кодировка для файлов
			Attr("hx-trigger", "submit"),               // Явно указываем триггер submit
			// HTMX автоматически предотвращает стандартную отправку формы
			// hx-on::after-request: выполнить после успешной отправки
			// Очищаем textarea, сбрасываем высоту, очищаем превью файлов
			Attr("hx-on::after-request", `
				if(event.detail.xhr.status === 200) {
					// Clear textarea
					const textarea = this.querySelector('textarea');
					if (textarea) {
						textarea.value = '';
						textarea.style.height = 'auto';
					}
					// Clear file previews via custom event
					const previewContainer = document.getElementById('file-preview-container');
					if (previewContainer) {
						previewContainer.dispatchEvent(new CustomEvent('clear-files'));
					}
				}
			`),

			// CSRF токен для защиты от подделки запросов
			components.CSRFInput(r),

			// Контейнер для textarea (занимает всё доступное пространство)
			Div(
				Class("flex-1"), // flex-1: занимает всё доступное пространство
				Textarea(
					ID("message-input"), // ID для связи с FilePreviewContainer
					Name("content"),     // Имя поля для отправки на сервер
					Class("textarea textarea-bordered w-full resize-none"),    // textarea: стили DaisyUI; w-full: полная ширина; resize-none: запрет изменения размера
					Placeholder("Type a message... (drag & drop files here)"), // Подсказка в пустом поле
					Rows("1"), // Начальная высота (1 строка)

					// Автоматическое изменение высоты textarea при вводе (Alpine.js для textarea resize)
					Attr("x-data", `{
						resize() {
							this.$el.style.height = "auto";
							this.$el.style.height = this.$el.scrollHeight + "px";
						}
					}`),
					Attr("@input", "resize()"),

					// Обработка нажатия Enter - Shift+Enter = новая строка, Enter = отправить сообщение
					Attr("@keydown.enter", "if(!event.shiftKey) { event.preventDefault(); document.getElementById('message-form').requestSubmit(); }"),

					// Обработка drag & drop файлов
					// Связываемся с FilePreviewContainer через custom events
					Attr("@drop.prevent", `
						const container = document.getElementById('file-preview-container');
						if (container) {
							container.dispatchEvent(new CustomEvent('file-drop', {
								detail: { files: event.dataTransfer.files }
							}));
						}
					`),
					Attr("@dragover.prevent", ""), // Предотвращаем стандартное поведение dragover
				),
			),

			// Контейнер для кнопок (вертикальное расположение)
			Div(
				Class("flex flex-col gap-2"), // flex-col: вертикальное расположение; gap-2: отступ между кнопками

				// Кнопка загрузки файлов
				Label(
					Class("btn btn-circle btn-ghost cursor-pointer"), // btn-circle: круглая кнопка; btn-ghost: прозрачная; cursor-pointer: курсор-указатель
					Title("Upload file"),                             // Подсказка при наведении
					Input(
						Type("file"),     // Поле для выбора файлов
						ID("file-input"), // ID для связи с FilePreviewContainer
						Name("files"),    // Имя поля для отправки на сервер (множественное)
						Class("hidden"),  // Скрываем стандартный input (используем кастомную кнопку)
						Attr("multiple"), // Разрешаем выбор нескольких файлов
						// File selection handled by Alpine.js @change - dispatch event to FilePreviewContainer
						Attr("@change", `
							const container = document.getElementById('file-preview-container');
							if (container && event.target.files) {
								container.dispatchEvent(new CustomEvent('file-input-change', {
									detail: { files: event.target.files }
								}));
							}
						`),
					),
					Text("📎"), // Иконка скрепки
				),

				// Кнопка отправки сообщения
				Button(
					Type("submit"),                      // Отправляет форму
					ID("message-send"),                  // ID для JavaScript манипуляций
					Class("btn btn-primary btn-circle"), // btn-primary: основная кнопка; btn-circle: круглая
					Title("Send message"),               // Подсказка при наведении
					Text("➤"),                           // Иконка стрелки вправо
				),
			),
		),
		// File handling is now done via Alpine.js in FilePreviewContainer component
		// File API requires JavaScript for file preview functionality
	)
}
