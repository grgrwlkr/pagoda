package messenger

import (
	"fmt"
	"strings"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FileAttachmentData представляет данные о вложенном файле
// Используется для передачи информации о файле между handlers и UI компонентами
type FileAttachmentData struct {
	ID       int64  // Уникальный идентификатор вложения в базе данных
	Filename string // Оригинальное имя файла
	MimeType string // MIME тип файла (например, "image/png", "application/pdf")
	FileSize int64  // Размер файла в байтах
	URL      string // URL для доступа к файлу через HTTP
}

// FileAttachment рендерит вложение файла в зависимости от его типа
// Разные типы файлов отображаются по-разному:
//   - Изображения: показывается превью с возможностью открыть в полном размере
//   - PDF: иконка и кнопка "Open"
//   - Текстовые файлы: иконка и кнопка "View"
//   - Другие: иконка по расширению и кнопка "Download"
//
// Параметры:
//   - r: объект запроса с контекстом
//   - attachment: данные о вложении
//
// Возвращает HTML-узел с отображением файла
func FileAttachment(r *ui.Request, attachment FileAttachmentData) Node {
	// Определяем тип файла по MIME типу
	isImage := strings.HasPrefix(attachment.MimeType, "image/") // Все изображения начинаются с "image/"
	isPDF := attachment.MimeType == "application/pdf"           // PDF файлы
	isText := strings.HasPrefix(attachment.MimeType, "text/")   // Все текстовые файлы начинаются с "text/"

	// Форматируем размер файла в человекочитаемый формат (KB, MB, GB)
	sizeStr := formatFileSize(attachment.FileSize)

	// Рендерим соответствующий компонент в зависимости от типа файла
	if isImage {
		return imageAttachment(r, attachment, sizeStr)
	} else if isPDF {
		return pdfAttachment(r, attachment, sizeStr)
	} else if isText {
		return textAttachment(r, attachment, sizeStr)
	} else {
		return genericAttachment(r, attachment, sizeStr)
	}
}

// imageAttachment рендерит вложение изображения с превью
// Изображения показываются как миниатюры, при клике открываются в полном размере
func imageAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 rounded-lg overflow-hidden border border-base-300"), // mt-2: верхний отступ; rounded-lg: скруглённые углы; overflow-hidden: скрывает переполнение; border: граница
		Div(
			Class("relative"), // relative: для абсолютного позиционирования (если понадобится)
			// Превью изображения
			Img(
				Src(attachment.URL),                       // URL изображения
				Alt(attachment.Filename),                  // Альтернативный текст для доступности
				Class("max-w-full h-auto cursor-pointer"), // max-w-full: максимальная ширина 100%; h-auto: автоматическая высота; cursor-pointer: курсор-указатель
				Attr("onclick", fmt.Sprintf("window.open('%s', '_blank')", attachment.URL)), // При клике открываем изображение в новой вкладке
				Title("Click to view full size"),                                            // Подсказка при наведении
			),
		),
		// Информация о файле (имя и размер)
		Div(
			Class("p-2 bg-base-200 flex items-center justify-between text-xs"), // p-2: отступы; bg-base-200: цвет фона; flex: горизонтальное расположение; justify-between: пространство между элементами
			Span(
				Class("text-base-content/70 truncate"), // truncate: обрезать длинные имена
				Text(attachment.Filename),
			),
			Span(
				Class("text-base-content/50 ml-2"), // ml-2: левый отступ
				Text(sizeStr),                      // Размер файла
			),
		),
	)
}

// pdfAttachment рендерит вложение PDF файла
// PDF файлы показываются с иконкой и кнопкой для открытия
func pdfAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"), // flex items-center: вертикальное выравнивание по центру; gap-3: отступ между элементами
		Div(
			Class("text-3xl"), // text-3xl: очень большой размер текста (для эмодзи)
			Text("📄"),         // Иконка документа
		),
		Div(
			Class("flex-1 min-w-0"), // flex-1: занимает всё доступное пространство; min-w-0: позволяет тексту обрезаться
			Div(
				Class("font-medium truncate"), // font-medium: средняя жирность; truncate: обрезать длинные имена
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),  // text-xs: очень маленький размер
				Text(fmt.Sprintf("PDF • %s", sizeStr)), // Показываем тип и размер
			),
		),
		// Кнопка для открытия PDF
		A(
			Href(attachment.URL),            // URL файла
			Target("_blank"),                // Открывать в новой вкладке
			Class("btn btn-sm btn-primary"), // btn-sm: маленькая кнопка; btn-primary: основная кнопка
			Text("Open"),
		),
	)
}

// textAttachment рендерит вложение текстового файла
// Текстовые файлы показываются с иконкой и кнопкой для просмотра
func textAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"),
		Div(
			Class("text-3xl"),
			Text("📝"), // Иконка текстового файла
		),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(fmt.Sprintf("Text • %s", sizeStr)), // Показываем тип и размер
			),
		),
		// Кнопка для просмотра текстового файла
		A(
			Href(attachment.URL),
			Target("_blank"),
			Class("btn btn-sm btn-primary"),
			Text("View"),
		),
	)
}

// genericAttachment рендерит вложение файла общего типа
// Используется для файлов, которые не являются изображениями, PDF или текстовыми
// Показывает иконку в зависимости от расширения файла
func genericAttachment(r *ui.Request, attachment FileAttachmentData, sizeStr string) Node {
	// Получаем расширение файла для определения иконки
	ext := getFileExtension(attachment.Filename)
	icon := getFileIcon(ext) // Получаем эмодзи иконку по расширению

	return Div(
		Class("mt-2 p-3 rounded-lg border border-base-300 bg-base-200 flex items-center gap-3"),
		Div(
			Class("text-3xl"),
			Text(icon), // Иконка файла
		),
		Div(
			Class("flex-1 min-w-0"),
			Div(
				Class("font-medium truncate"),
				Text(attachment.Filename),
			),
			Div(
				Class("text-xs text-base-content/60"),
				Text(fmt.Sprintf("%s • %s", strings.ToUpper(ext), sizeStr)), // Показываем расширение (заглавными) и размер
			),
		),
		// Кнопка для скачивания файла
		A(
			Href(attachment.URL),
			Target("_blank"),
			Class("btn btn-sm btn-primary"),
			Text("Download"),
		),
	)
}

// formatFileSize форматирует размер файла в человекочитаемый формат
// Преобразует байты в KB, MB, GB и т.д.
// Примеры:
//   - 1024 байта -> "1.0 KB"
//   - 1048576 байт -> "1.0 MB"
//   - 1073741824 байт -> "1.0 GB"
//
// Параметры:
//   - bytes: размер файла в байтах
//
// Возвращает:
//   - string: отформатированный размер (например, "1.5 MB")
func formatFileSize(bytes int64) string {
	const unit = 1024 // Базовый множитель (1024 байта = 1 KB)
	if bytes < unit {
		// Меньше 1 KB - показываем в байтах
		return fmt.Sprintf("%d B", bytes)
	}

	// Определяем степень (KB, MB, GB, TB, PB, EB)
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Форматируем с одной десятичной цифрой
	// "KMGTPE" - массив символов для единиц измерения
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getFileExtension извлекает расширение файла из имени
// Примеры:
//   - "document.pdf" -> "pdf"
//   - "image.png" -> "png"
//   - "file.tar.gz" -> "gz" (возвращает последнее расширение)
//   - "file" -> "" (нет расширения)
//
// Параметры:
//   - filename: имя файла
//
// Возвращает:
//   - string: расширение файла (в нижнем регистре) или пустую строку
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		// Нет точки в имени - нет расширения
		return ""
	}
	// Возвращаем последнюю часть после точки (для "file.tar.gz" вернёт "gz")
	return parts[len(parts)-1]
}

// getFileIcon возвращает эмодзи иконку в зависимости от расширения файла
// Используется для визуального отображения типа файла
// Параметры:
//   - ext: расширение файла (например, "pdf", "doc", "zip")
//
// Возвращает:
//   - string: эмодзи иконка или "📎" по умолчанию
func getFileIcon(ext string) string {
	ext = strings.ToLower(ext) // Приводим к нижнему регистру для сравнения

	// Маппинг расширений на иконки
	icons := map[string]string{
		"doc":  "📄", // Word документы
		"docx": "📄",
		"xls":  "📊", // Excel таблицы
		"xlsx": "📊",
		"ppt":  "📽️", // PowerPoint презентации
		"pptx": "📽️",
		"zip":  "📦", // Архивы
		"rar":  "📦",
		"7z":   "📦",
		"mp3":  "🎵", // Аудио файлы
		"mp4":  "🎬", // Видео файлы
		"avi":  "🎬",
		"mov":  "🎬",
		"exe":  "⚙️", // Исполняемые файлы
		"dmg":  "💿",  // Диски
		"iso":  "💿",
	}

	// Если есть иконка для этого расширения, возвращаем её
	if icon, ok := icons[ext]; ok {
		return icon
	}

	// По умолчанию возвращаем иконку скрепки
	return "📎"
}
