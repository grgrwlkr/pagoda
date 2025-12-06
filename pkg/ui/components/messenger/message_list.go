package messenger

import (
	"fmt"
	"time"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageList рендерит список сообщений в канале или прямом сообщении
// Принимает:
//   - r: объект запроса с контекстом и данными пользователя
//   - messages: массив сообщений для отображения
//
// Возвращает HTML-узел (Node) с контейнером списка сообщений
func MessageList(r *ui.Request, messages []MessageData) Node {
	return Div(
		ID("message-list"), // ID для HTMX - позволяет обновлять список через hx-target
		Class("flex-1 overflow-y-auto p-4 space-y-4"), // flex-1: занимает всё доступное пространство; overflow-y-auto: прокрутка при переполнении; p-4: отступы; space-y-4: вертикальные отступы между элементами
		Group(renderMessages(r, messages)),            // Group объединяет несколько узлов в один
	)
}

// MessageData структура данных для одного сообщения
// Используется для передачи данных между handlers и UI компонентами
type MessageData struct {
	ID          int64                // Уникальный идентификатор сообщения в базе данных
	Content     string               // Текст сообщения
	UserID      int64                // ID пользователя, который отправил сообщение
	UserName    string               // Имя пользователя для отображения
	CreatedAt   time.Time            // Время создания сообщения
	EditedAt    *time.Time           // Время последнего редактирования (nil если не редактировалось)
	Reactions   []ReactionData       // Массив реакций (эмодзи) на сообщение
	Attachments []FileAttachmentData // Массив вложенных файлов
	ReplyCount  int                  // Количество ответов в потоке (thread)
}

// ReactionData структура данных для реакции на сообщение
type ReactionData struct {
	Emoji   string  // Эмодзи реакции (например, "👍", "❤️")
	Count   int     // Количество пользователей, поставивших эту реакцию
	UserIDs []int64 // Массив ID пользователей, поставивших реакцию (для проверки, поставил ли текущий пользователь)
}

// renderMessages преобразует массив MessageData в массив HTML-узлов
// Создаёт Group (группу узлов) для эффективного рендеринга
func renderMessages(r *ui.Request, messages []MessageData) Group {
	// Создаём Group с предварительно выделенной ёмкостью (оптимизация памяти)
	group := make(Group, 0, len(messages))

	// Итерируемся по всем сообщениям и добавляем HTML-узел для каждого
	for _, msg := range messages {
		group = append(group, messageItem(r, msg))
	}

	return group
}

// MessageItem рендерит одно сообщение (экспортируется для использования в HTMX)
// HTMX может запросить только этот компонент для обновления конкретного сообщения
// без перезагрузки всей страницы
func MessageItem(r *ui.Request, msg MessageData) Node {
	return messageItem(r, msg)
}

// messageItem создаёт HTML-разметку для одного сообщения
// Включает: аватар, имя пользователя, время, текст, вложения, реакции, действия с потоком
func messageItem(r *ui.Request, msg MessageData) Node {
	return Div(
		// Основной контейнер сообщения
		Class("flex gap-3 hover:bg-base-200/50 p-2 rounded-lg transition-colors"), // flex: горизонтальное расположение; gap-3: отступ между элементами; hover: эффект при наведении; p-2: внутренние отступы; rounded-lg: скруглённые углы; transition-colors: плавная смена цвета
		ID(fmt.Sprintf("message-%d", msg.ID)),                                     // Уникальный ID для HTMX обновлений и JavaScript манипуляций

		// Аватар пользователя
		// nil для isOnline означает, что статус онлайн неизвестен (будет обновлён через WebSocket)
		UserAvatar(r, msg.UserID, msg.UserName, "sm", nil),

		// Контейнер с содержимым сообщения
		Div(
			Class("flex-1 min-w-0"), // flex-1: занимает всё доступное пространство; min-w-0: позволяет тексту переноситься

			// Заголовок с именем пользователя и временем
			Div(
				Class("flex items-center gap-2 mb-1"), // flex: горизонтальное расположение; items-center: вертикальное выравнивание по центру; gap-2: отступ между элементами; mb-1: нижний отступ
				Span(
					Class("font-semibold"), // Полужирный шрифт для имени
					Text(msg.UserName),
				),
				Span(
					Class("text-xs text-base-content/60"), // text-xs: маленький размер; text-base-content/60: цвет с прозрачностью 60%
					Text(msg.CreatedAt.Format("15:04")),   // Форматируем время в формат ЧЧ:ММ
				),
				// Показываем метку "(edited)" только если сообщение редактировалось
				If(msg.EditedAt != nil,
					Span(
						Class("text-xs text-base-content/40 italic"), // italic: курсив
						Text("(edited)"),
					),
				),
			),

			// Текст сообщения
			// Показываем только если есть содержимое (может быть только файл без текста)
			If(msg.Content != "",
				Div(
					Class("text-base-content/90 whitespace-pre-wrap break-words"), // whitespace-pre-wrap: сохраняет переносы строк; break-words: переносит длинные слова
					Text(msg.Content),
				),
			),

			// Вложения (файлы, изображения)
			// Показываем только если есть вложения
			If(len(msg.Attachments) > 0,
				Div(
					Class("mt-2 space-y-2"),                      // mt-2: верхний отступ; space-y-2: вертикальные отступы между вложениями
					Group(renderAttachments(r, msg.Attachments)), // Рендерим все вложения
				),
			),

			// Реакции (эмодзи)
			// Показываем только если есть реакции
			If(len(msg.Reactions) > 0,
				Div(
					Class("flex flex-wrap gap-1 mt-2"),               // flex-wrap: перенос на новую строку при нехватке места; gap-1: отступ между реакциями
					Group(renderReactions(r, msg.Reactions, msg.ID)), // Рендерим все реакции
				),
			),

			// Действия с потоком (thread)
			Div(
				Class("flex items-center gap-2 mt-2 text-sm"), // text-sm: маленький размер текста

				// Кнопка "Ответить" - открывает панель треда справа
				Button(
					Type("button"),                      // Важно: кнопка не должна отправлять форму
					Class("btn btn-ghost btn-sm gap-1"), // btn-ghost: прозрачная кнопка; btn-sm: маленький размер; gap-1: отступ между иконкой и текстом
					// hx-get: HTMX запрос GET для открытия панели треда
					Attr("hx-get", r.Path("messenger.message.thread.panel", msg.ID)),
					// hx-target: куда вставить панель треда (правая панель)
					Attr("hx-target", "#right-panel"),
					// hx-swap: как вставлять (innerHTML - заменить содержимое)
					Attr("hx-swap", "innerHTML"),
					// hx-on::after-request: показать панель после загрузки с анимацией
					// Use Alpine.js store method instead of global function
					Attr("hx-on::after-request", fmt.Sprintf(`
						if (window.Alpine && window.Alpine.store && window.Alpine.store('threadPanel')) {
							window.Alpine.store('threadPanel').openThreadPanel('%d');
						} else {
							// Fallback if Alpine.js store not available
							var rightPanel = document.getElementById('right-panel');
							if (rightPanel) {
								rightPanel.classList.remove('hidden');
								rightPanel.setAttribute('data-thread-id', '%d');
							}
						}
					`, msg.ID, msg.ID)),
					// Предотвращаем стандартное поведение кнопки
					Attr("@click", "event.preventDefault(); return false;"),
					Text("💬 Reply"),
				),

				// Счётчик ответов (показываем только если есть ответы) - открывает панель треда
				If(msg.ReplyCount > 0,
					Button(
						Type("button"), // Важно: кнопка не должна отправлять форму
						Class("btn btn-ghost btn-sm text-base-content/60 hover:text-base-content"), // hover: изменение цвета при наведении
						// hx-get: HTMX запрос GET для открытия панели треда
						Attr("hx-get", r.Path("messenger.message.thread.panel", msg.ID)),
						// hx-target: куда вставить панель треда (правая панель)
						Attr("hx-target", "#right-panel"),
						// hx-swap: как вставлять (innerHTML - заменить содержимое)
						Attr("hx-swap", "innerHTML"),
						// hx-on::after-request: показать панель после загрузки с анимацией
						// Используем Alpine.js store для плавного переключения между тредами
						Attr("hx-on::after-request", fmt.Sprintf(`
							if (window.Alpine && window.Alpine.store && window.Alpine.store('threadPanel')) {
								window.Alpine.store('threadPanel').switchThreadPanel('%d');
							} else {
								// Fallback if Alpine.js store not available
								var rightPanel = document.getElementById('right-panel');
								if (rightPanel) {
									rightPanel.classList.remove('hidden');
									rightPanel.setAttribute('data-thread-id', '%d');
								}
							}
						`, msg.ID, msg.ID)),
						// Предотвращаем стандартное поведение кнопки через Alpine.js
						Attr("@click.prevent", ""),
						Text(fmt.Sprintf("%d %s", msg.ReplyCount, pluralize(msg.ReplyCount, "reply", "replies"))), // Правильное склонение (1 reply, 2 replies)
					),
				),
			),
		),
	)
}

// pluralize возвращает правильную форму слова в зависимости от количества
// Используется для корректного отображения "1 reply" vs "2 replies"
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular // "1 reply"
	}
	return plural // "2 replies", "0 replies"
}

// renderThreadReplyForm создаёт форму для ответа в потоке
// Форма отправляется через HTMX, ответ добавляется в поток без перезагрузки страницы
func renderThreadReplyForm(r *ui.Request, messageID int64) Node {
	return Form(
		Class("flex gap-2"), // flex: горизонтальное расположение элементов формы
		Method("POST"),      // HTTP метод для отправки
		Action(r.Path("messenger.message.reply", messageID)), // URL для отправки (fallback для браузеров без JS)
		// HTMX атрибуты для AJAX отправки
		Attr("hx-post", r.Path("messenger.message.reply", messageID)), // HTMX POST запрос
		Attr("hx-target", fmt.Sprintf("#thread-%d", messageID)),       // Куда вставить ответ
		Attr("hx-swap", "beforeend"),                                  // Вставить в конец контейнера
		// hx-on::after-request: выполнить после успешной отправки (очистить textarea)
		Attr("hx-on::after-request", "this.querySelector('textarea').value = ''; this.querySelector('textarea').style.height = 'auto';"),

		// CSRF токен для защиты от подделки запросов
		If(r.CSRF != "", Input(
			Type("hidden"), // Скрытое поле
			Name("csrf"),
			Value(r.CSRF),
		)),

		// Контейнер для textarea
		Div(
			Class("flex-1"), // Занимает всё доступное пространство
			Textarea(
				Name("content"), // Имя поля для отправки на сервер
				Class("textarea textarea-bordered w-full resize-none text-sm"), // textarea: стили DaisyUI; resize-none: запрет изменения размера; text-sm: маленький размер
				Placeholder("Write a reply..."),                                // Подсказка в пустом поле
				Rows("2"),                                                      // Начальная высота (2 строки)
				Required(),                                                     // Обязательное поле
				// Alpine.js для автоматического изменения высоты textarea при вводе
				Attr("x-data", `{
					resize() {
						this.$el.style.height = "auto"; // Сброс высоты
						this.$el.style.height = this.$el.scrollHeight + "px"; // Установка высоты по содержимому
					}
				}`),
				Attr("@input", "resize()"), // Вызывать resize при каждом вводе
			),
		),

		// Контейнер для кнопок
		Div(
			Class("flex flex-col gap-2"), // flex-col: вертикальное расположение; gap-2: отступ между кнопками
			// Кнопка отправки
			Button(
				Type("submit"),                  // Отправляет форму
				Class("btn btn-primary btn-sm"), // btn-primary: основная кнопка; btn-sm: маленький размер
				Text("Reply"),
			),
			// Кнопка отмены (скрывает форму) - используем Alpine.js @click
			Button(
				Type("button"),                // Не отправляет форму
				Class("btn btn-ghost btn-sm"), // btn-ghost: прозрачная кнопка
				Attr("@click", fmt.Sprintf("document.getElementById('thread-reply-form-%d').classList.add('hidden');", messageID)), // Скрываем форму через Alpine.js
				Text("Cancel"),
			),
		),
	)
}

// renderReactions создаёт кнопки для всех реакций на сообщение
// Каждая кнопка показывает эмодзи и количество, при клике переключает реакцию
func renderReactions(r *ui.Request, reactions []ReactionData, messageID int64) Group {
	group := make(Group, 0, len(reactions))

	for _, reaction := range reactions {
		group = append(group,
			Button(
				Class("btn btn-xs gap-1 hover:bg-base-300"), // btn-xs: очень маленький размер; gap-1: отступ между эмодзи и числом
				Text(reaction.Emoji),                        // Эмодзи реакции
				Text(fmt.Sprintf("%d", reaction.Count)),     // Количество
				// HTMX для переключения реакции
				Attr("hx-post", r.Path("messenger.reaction.add", messageID)),    // POST запрос для добавления/удаления реакции
				Attr("hx-vals", fmt.Sprintf(`{"emoji": "%s"}`, reaction.Emoji)), // Передаём эмодзи в теле запроса
				Attr("hx-target", fmt.Sprintf("#message-%d", messageID)),        // Обновляем всё сообщение
				Attr("hx-swap", "outerHTML"),                                    // Заменяем весь элемент сообщения
				Title("Click to toggle reaction"),                               // Подсказка при наведении
			),
		)
	}

	return group
}

// renderAttachments создаёт HTML-узлы для всех вложений сообщения
// Использует компонент FileAttachment для отображения каждого файла
func renderAttachments(r *ui.Request, attachments []FileAttachmentData) Group {
	group := make(Group, 0, len(attachments))

	for _, attachment := range attachments {
		// FileAttachment определяет тип файла и рендерит соответствующий UI
		group = append(group, FileAttachment(r, attachment))
	}

	return group
}
