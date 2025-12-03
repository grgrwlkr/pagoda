package messenger

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// UserAvatar рендерит аватар пользователя с опциональным индикатором статуса онлайн
// Параметры:
//   - r: объект запроса с контекстом
//   - userID: ID пользователя (для уникального ID элемента)
//   - userName: имя пользователя (для генерации инициалов)
//   - size: размер аватара ("xs", "sm", "md", "lg")
//   - isOnline: опциональный статус онлайн (nil = неизвестен, true = онлайн, false = офлайн)
//
// Возвращает HTML-узел с аватаром
func UserAvatar(r *ui.Request, userID int64, userName string, size string, isOnline *bool) Node {
	// Генерируем инициалы из имени пользователя
	// Например, "John Doe" -> "JD", "Alice" -> "A"
	initials := getInitials(userName)

	// Классы размеров для аватара
	// Каждый размер имеет свою ширину, высоту и размер текста
	sizeClasses := map[string]string{
		"xs": "w-6 h-6 text-xs",     // Очень маленький (24x24px)
		"sm": "w-8 h-8 text-sm",     // Маленький (32x32px)
		"md": "w-12 h-12 text-base", // Средний (48x48px) - по умолчанию
		"lg": "w-16 h-16 text-lg",   // Большой (64x64px)
	}

	sizeClass := sizeClasses[size]
	if sizeClass == "" {
		// Если размер не указан или неверный, используем средний
		sizeClass = sizeClasses["md"]
	}

	// Размер индикатора статуса онлайн зависит от размера аватара
	statusSize := "w-2 h-2" // По умолчанию маленький индикатор
	if size == "xs" {
		statusSize = "w-1.5 h-1.5" // Для очень маленького аватара - ещё меньше
	} else if size == "lg" {
		statusSize = "w-3 h-3" // Для большого аватара - больше
	}

	// TODO: Добавить поддержку URL аватара, когда будет реализован UserProfile
	// Пока используем placeholder с инициалами

	// Строим индикатор статуса онлайн, если isOnline не nil
	var statusIndicator Node
	if isOnline != nil {
		var statusClass string
		var statusValue string
		if *isOnline {
			// Пользователь онлайн - зелёный индикатор
			statusClass = "bg-success"
			statusValue = "true"
		} else {
			// Пользователь офлайн - серый индикатор
			statusClass = "bg-base-300"
			statusValue = "false"
		}
		// Индикатор позиционируется абсолютно в правом нижнем углу аватара
		statusIndicator = Div(
			Class(fmt.Sprintf("absolute bottom-0 right-0 %s rounded-full border-2 border-base-100 %s", statusSize, statusClass)), // absolute: абсолютное позиционирование; bottom-0 right-0: правый нижний угол; rounded-full: круглый; border-2: граница
			Attr("data-online", statusValue), // data-атрибут для JavaScript (можно обновлять через JS при изменении статуса)
		)
	}

	return Div(
		Class(fmt.Sprintf("avatar placeholder %s relative", sizeClass)), // avatar placeholder: классы DaisyUI для аватара; relative: для абсолютного позиционирования индикатора
		ID(fmt.Sprintf("user-avatar-%d", userID)),                       // Уникальный ID для JavaScript манипуляций (обновление статуса онлайн)
		Div(
			Class("bg-neutral text-neutral-content rounded-full flex items-center justify-center font-semibold"), // bg-neutral: нейтральный цвет фона; text-neutral-content: контрастный цвет текста; rounded-full: круглый; flex items-center justify-center: центрирование инициалов; font-semibold: полужирный шрифт
			Text(initials), // Инициалы пользователя
		),
		// Индикатор статуса онлайн (показывается только если isOnline не nil)
		If(statusIndicator != nil, statusIndicator),
	)
}

// getInitials извлекает инициалы из имени пользователя
// Правила:
//   - Если имя пустое, возвращает "?"
//   - Если одно слово, возвращает первую букву
//   - Если несколько слов, возвращает первую букву первого слова и первую букву последнего слова
//
// Примеры:
//   - "John Doe" -> "JD"
//   - "Alice" -> "A"
//   - "Mary Jane Watson" -> "MW"
//   - "" -> "?"
func getInitials(name string) string {
	if len(name) == 0 {
		return "?"
	}

	// Преобразуем строку в руны (для корректной работы с Unicode)
	words := []rune(name)
	if len(words) == 1 {
		// Одно слово - возвращаем первую букву
		return string(words[0:1])
	}

	// Получаем первую букву первого слова
	first := string(words[0])

	// Ищем первую букву последнего слова (пропускаем пробелы в конце)
	last := ""
	for i := len(words) - 1; i >= 0; i-- {
		if words[i] != ' ' {
			last = string(words[i])
			break
		}
	}

	// Если не нашли последнее слово (только пробелы), возвращаем только первую букву
	if last == "" {
		return first
	}

	// Возвращаем инициалы: первая буква первого слова + первая буква последнего слова
	return first + last
}
