package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SidebarData содержит данные для отображения боковой панели
// Боковая панель показывает workspace, каналы и прямые сообщения
type SidebarData struct {
	WorkspaceID     int64         // ID текущего workspace
	WorkspaceName   string        // Название workspace
	Channels        []ChannelData // Список каналов, где пользователь является участником
	DirectMessages  []DMData      // Список прямых сообщений пользователя
	ActiveChannelID *int64        // ID активного канала (nil если нет активного канала)
	ActiveDMID      *int64        // ID активного DM (nil если нет активного DM)
}

// Sidebar рендерит левую боковую панель с workspace, каналами и прямыми сообщениями
// Боковая панель - это основной элемент навигации в приложении
// Параметры:
//   - r: объект запроса с контекстом
//   - data: данные для отображения (каналы, DM, активные элементы)
//
// Возвращает HTML-узел с боковой панелью
func Sidebar(r *ui.Request, data SidebarData) Node {
	// Помечаем активные каналы (для подсветки в UI)
	channels := make([]ChannelData, len(data.Channels))
	for i, ch := range data.Channels {
		channels[i] = ch
		// Если это активный канал, помечаем его
		if data.ActiveChannelID != nil && ch.ID == *data.ActiveChannelID {
			channels[i].IsActive = true
		}
	}

	// Помечаем активные DM (для подсветки в UI)
	dms := make([]DMData, len(data.DirectMessages))
	for i, dm := range data.DirectMessages {
		dms[i] = dm
		// Если это активный DM, помечаем его
		if data.ActiveDMID != nil && dm.ID == *data.ActiveDMID {
			dms[i].IsActive = true
		}
	}

	return Div(
		Class("w-64 bg-base-200 border-r border-base-300 flex flex-col"), // w-64: фиксированная ширина; bg-base-200: цвет фона; border-r: правая граница; flex flex-col: вертикальное расположение

		// Заголовок с профилем пользователя
		Div(
			Class("p-4 border-b border-base-300"), // p-4: отступы; border-b: нижняя граница
			userProfileSection(r),                 // Секция с аватаром и именем пользователя
		),

		// Поисковая строка
		SearchBox(r, data.WorkspaceID),

		// Селектор workspace (название текущего workspace)
		Div(
			Class("p-4 border-b border-base-300"),
			H3(
				Class("text-sm font-semibold uppercase text-base-content/70 mb-2"), // text-sm: маленький размер; uppercase: заглавные буквы; mb-2: нижний отступ
				Text("Workspace"),
			),
			Div(
				Class("text-base font-medium"), // text-base: базовый размер; font-medium: средняя жирность
				Text(data.WorkspaceName),
			),
		),

		// Секция каналов
		Div(
			Class("flex-1 overflow-y-auto p-4"),        // flex-1: занимает всё доступное пространство; overflow-y-auto: прокрутка при переполнении
			ChannelList(r, channels, data.WorkspaceID), // Список каналов
		),

		// Секция прямых сообщений
		Div(
			Class("p-4 border-t border-base-300"), // border-t: верхняя граница (разделитель)
			DirectMessagesList(r, dms),            // Список DM
		),
	)
}

// userProfileSection рендерит секцию профиля пользователя в верхней части sidebar
// Показывает аватар, имя и email текущего пользователя
func userProfileSection(r *ui.Request) Node {
	// Проверяем, авторизован ли пользователь
	// Если нет, показываем сообщение "Not authenticated"
	if !r.IsAuth || r.AuthUser == nil {
		return Div(Text("Not authenticated"))
	}

	return Div(
		Class("flex items-center gap-3"), // flex: горизонтальное расположение; items-center: вертикальное выравнивание; gap-3: отступ между элементами
		// Аватар пользователя
		// nil для isOnline означает, что статус онлайн неизвестен (будет обновлён через WebSocket)
		UserAvatar(r, int64(r.AuthUser.ID), r.AuthUser.Name, "sm", nil),

		// Информация о пользователе
		Div(
			Class("flex-1 min-w-0"), // flex-1: занимает всё доступное пространство; min-w-0: позволяет тексту обрезаться
			Div(
				Class("font-medium truncate"), // font-medium: средняя жирность; truncate: обрезать длинные имена
				Text(r.AuthUser.Name),
			),
			Div(
				Class("text-xs text-base-content/60"), // text-xs: очень маленький размер; text-base-content/60: цвет с прозрачностью
				Text(r.AuthUser.Email),
			),
		),
	)
}
