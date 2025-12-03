package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
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
		actionRoute = r.Path("messenger.dm.message.create", channelIDOrDMID)
	} else {
		// Для каналов используем обычный route
		actionRoute = r.Path("messenger.message.create", channelIDOrDMID)
	}

	return Div(
		Class("border-t border-base-300 p-4 bg-base-100"), // border-t: верхняя граница; p-4: отступы; bg-base-100: цвет фона

		// Контейнер для превью загружаемых файлов (изначально скрыт)
		// Показывается через JavaScript при выборе файлов
		Div(
			ID("file-preview-container"),              // ID для JavaScript манипуляций
			Class("mb-2 flex flex-wrap gap-2 hidden"), // mb-2: нижний отступ; flex-wrap: перенос на новую строку; gap-2: отступ между превью; hidden: скрыт по умолчанию
		),

		// Основная форма для отправки сообщения
		Form(
			Class("flex gap-2"),                    // flex: горизонтальное расположение; gap-2: отступ между элементами
			ID("message-form"),                     // ID для JavaScript и HTMX
			Method("POST"),                         // HTTP метод отправки
			Action(actionRoute),                    // URL для отправки (fallback для браузеров без JS)
			Attr("enctype", "multipart/form-data"), // Необходимо для загрузки файлов

			// HTMX атрибуты для AJAX отправки без перезагрузки страницы
			Attr("hx-post", actionRoute),               // HTMX POST запрос
			Attr("hx-target", "#message-list"),         // Куда вставить новое сообщение
			Attr("hx-swap", "beforeend"),               // Вставить в конец списка сообщений
			Attr("hx-encoding", "multipart/form-data"), // Кодировка для файлов
			// hx-on::after-request: выполнить после успешной отправки
			// Очищаем textarea, сбрасываем высоту, очищаем превью файлов и input файлов
			Attr("hx-on::after-request", "this.querySelector('textarea').value = ''; this.querySelector('textarea').style.height = 'auto'; document.getElementById('file-preview-container').innerHTML = ''; document.getElementById('file-preview-container').classList.add('hidden'); document.getElementById('file-input').value = '';"),

			// CSRF токен для защиты от подделки запросов
			// Показываем только если токен есть (для неавторизованных пользователей может не быть)
			If(r.CSRF != "", Input(
				Type("hidden"), // Скрытое поле
				Name("csrf"),
				Value(r.CSRF),
			)),

			// Контейнер для textarea (занимает всё доступное пространство)
			Div(
				Class("flex-1"), // flex-1: занимает всё доступное пространство
				Textarea(
					ID("message-input"), // ID для JavaScript манипуляций
					Name("content"),     // Имя поля для отправки на сервер
					Class("textarea textarea-bordered w-full resize-none"),    // textarea: стили DaisyUI; w-full: полная ширина; resize-none: запрет изменения размера
					Placeholder("Type a message... (drag & drop files here)"), // Подсказка в пустом поле
					Rows("1"), // Начальная высота (1 строка)

					// Alpine.js для автоматического изменения высоты textarea при вводе
					Attr("x-data", `{
						resize() {
							// Сбрасываем высоту на auto для пересчёта
							this.$el.style.height = "auto";
							// Устанавливаем высоту равную содержимому
							this.$el.style.height = this.$el.scrollHeight + "px";
						}
					}`),
					Attr("@input", "resize()"), // Вызывать resize при каждом вводе текста

					// Обработка нажатия Enter
					// Shift+Enter = новая строка, Enter = отправить сообщение
					Attr("@keydown.enter", "if(!event.shiftKey) { event.preventDefault(); document.getElementById('message-form').requestSubmit(); }"),

					// Обработка drag & drop файлов
					Attr("ondrop", "handleFileDrop(event); return false;"),      // ondrop: когда файл отпущен над textarea
					Attr("ondragover", "event.preventDefault(); return false;"), // ondragover: предотвращаем стандартное поведение браузера
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
						ID("file-input"), // ID для JavaScript манипуляций
						Name("files"),    // Имя поля для отправки на сервер (множественное)
						Class("hidden"),  // Скрываем стандартный input (используем кастомную кнопку)
						Attr("multiple"), // Разрешаем выбор нескольких файлов
						Attr("onchange", "handleFileSelect(event)"), // Обработчик выбора файлов через диалог
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

		// JavaScript код для обработки файлов
		// Встроенный скрипт для работы с drag & drop и превью файлов
		Script(
			Raw(`
			// Обработчик выбора файлов через диалог (кнопка "📎")
			function handleFileSelect(event) {
				const files = event.target.files; // Получаем выбранные файлы
				showFilePreviews(files); // Показываем превью
			}
			
			// Обработчик drag & drop файлов на textarea
			function handleFileDrop(event) {
				event.preventDefault(); // Предотвращаем стандартное поведение браузера (открытие файла)
				const files = event.dataTransfer.files; // Получаем перетащенные файлы
				const fileInput = document.getElementById('file-input'); // Находим скрытый input
				
				// Создаём DataTransfer для программной установки файлов в input
				// Это необходимо, чтобы файлы были доступны при отправке формы
				const dataTransfer = new DataTransfer();
				for (let i = 0; i < files.length; i++) {
					dataTransfer.items.add(files[i]); // Добавляем каждый файл
				}
				fileInput.files = dataTransfer.files; // Устанавливаем файлы в input
				showFilePreviews(files); // Показываем превью
			}
			
			// Функция отображения превью выбранных файлов
			function showFilePreviews(files) {
				const container = document.getElementById('file-preview-container');
				container.innerHTML = ''; // Очищаем предыдущие превью
				
				// Если файлов нет, скрываем контейнер
				if (files.length === 0) {
					container.classList.add('hidden');
					return;
				}
				
				// Показываем контейнер
				container.classList.remove('hidden');
				
				// Создаём превью для каждого файла
				for (let i = 0; i < files.length; i++) {
					const file = files[i];
					const div = document.createElement('div');
					div.className = 'relative inline-block p-2 border border-base-300 rounded-lg bg-base-200';
					
					// Для изображений показываем миниатюру
					if (file.type.startsWith('image/')) {
						const img = document.createElement('img');
						img.src = URL.createObjectURL(file); // Создаём временный URL для превью
						img.className = 'max-w-20 max-h-20 object-cover rounded'; // Ограничиваем размер
						div.appendChild(img);
					} else {
						// Для других файлов показываем иконку
						const icon = document.createElement('div');
						icon.className = 'text-2xl';
						icon.textContent = '📎';
						div.appendChild(icon);
					}
					
					// Показываем имя файла
					const name = document.createElement('div');
					name.className = 'text-xs truncate max-w-20'; // text-xs: маленький размер; truncate: обрезать длинные имена
					name.textContent = file.name;
					div.appendChild(name);
					
					// Кнопка удаления файла из списка
					const removeBtn = document.createElement('button');
					removeBtn.type = 'button'; // Не отправляет форму
					removeBtn.className = 'absolute -top-1 -right-1 btn btn-xs btn-circle btn-error'; // Абсолютное позиционирование в правом верхнем углу
					removeBtn.textContent = '×';
					// Обработчик удаления файла
					removeBtn.onclick = function() {
						removeFile(i); // Вызываем функцию удаления с индексом файла
					};
					div.appendChild(removeBtn);
					
					container.appendChild(div); // Добавляем превью в контейнер
				}
			}
			
			// Функция удаления файла из списка перед отправкой
			function removeFile(index) {
				const fileInput = document.getElementById('file-input');
				// Создаём новый DataTransfer с файлами, исключая удаляемый
				const dataTransfer = new DataTransfer();
				for (let i = 0; i < fileInput.files.length; i++) {
					if (i !== index) { // Пропускаем файл с указанным индексом
						dataTransfer.items.add(fileInput.files[i]);
					}
				}
				fileInput.files = dataTransfer.files; // Обновляем список файлов
				showFilePreviews(fileInput.files); // Обновляем превью
			}
			`),
		),
	)
}
