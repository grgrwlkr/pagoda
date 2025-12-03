package websocket

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
)

// Hub поддерживает набор активных WebSocket соединений и транслирует сообщения между ними
// Hub - это центральный компонент для real-time коммуникации в приложении
// Он управляет всеми подключениями и маршрутизирует сообщения нужным пользователям
type Hub struct {
	// Зарегистрированные соединения, ключ - user ID
	// Каждый пользователь может иметь только одно активное соединение
	// Если пользователь открывает новое соединение, старое закрывается
	connections map[int64]*Connection

	// Входящие сообщения от соединений для широковещательной рассылки
	// Канал для передачи сообщений всем подключенным пользователям
	broadcast chan []byte

	// Запросы на регистрацию новых соединений
	// Когда клиент подключается, он отправляется в этот канал
	register chan *Connection

	// Запросы на отмену регистрации соединений
	// Когда клиент отключается, он отправляется в этот канал
	unregister chan *Connection

	// Мьютекс для потокобезопасного доступа к connections map
	// Go map не является потокобезопасной, поэтому нужна синхронизация
	mu sync.RWMutex

	// ORM клиент для операций с базой данных
	// Используется для получения списков участников каналов/workspace
	ORM *ent.Client
}

// Глобальный экземпляр hub (устанавливается WebSocket handler'ом)
// Это singleton паттерн - один hub на всё приложение
var globalHub *Hub

// GetHub возвращает глобальный экземпляр hub
// Используется другими handlers для отправки WebSocket событий
func GetHub() *Hub {
	return globalHub
}

// SetHub устанавливает глобальный экземпляр hub
// Вызывается при инициализации WebSocket handler'а
func SetHub(hub *Hub) {
	globalHub = hub
}

// NewHub создаёт новый экземпляр Hub
// Параметры:
//   - orm: клиент Ent ORM для работы с базой данных
//
// Возвращает:
//   - *Hub: новый экземпляр hub
func NewHub(orm *ent.Client) *Hub {
	return &Hub{
		connections: make(map[int64]*Connection), // Инициализируем пустую map для соединений
		broadcast:   make(chan []byte, 256),      // Буферизованный канал на 256 сообщений (предотвращает блокировку)
		register:    make(chan *Connection),      // Небуферизованный канал (блокирует до обработки)
		unregister:  make(chan *Connection),      // Небуферизованный канал (блокирует до обработки)
		ORM:         orm,                         // Сохраняем ORM клиент
	}
}

// Run запускает главный цикл hub
// Этот метод должен запускаться в отдельной goroutine
// Он обрабатывает регистрацию/отмену регистрации соединений и рассылку сообщений
// Работает бесконечно до завершения приложения
func (h *Hub) Run() {
	slog.Info("WebSocket hub started")
	for {
		select {
		case conn := <-h.register:
			// Новое соединение регистрируется
			h.mu.Lock() // Блокируем для записи
			// Если у пользователя уже есть соединение, закрываем старое
			// Это предотвращает множественные соединения от одного пользователя
			if oldConn, exists := h.connections[conn.UserID]; exists {
				slog.Info("Closing existing connection for user", "user_id", conn.UserID)
				oldConn.Close() // Закрываем старое соединение
			}
			// Регистрируем новое соединение
			h.connections[conn.UserID] = conn
			connectionCount := len(h.connections)
			h.mu.Unlock() // Разблокируем
			slog.Info("Connection registered", "user_id", conn.UserID, "total_connections", connectionCount)

		case conn := <-h.unregister:
			// Соединение отменяется (клиент отключился)
			h.mu.Lock() // Блокируем для записи
			if _, ok := h.connections[conn.UserID]; ok {
				// Удаляем соединение из map
				delete(h.connections, conn.UserID)
				// Закрываем канал отправки сообщений
				// Это сигнализирует WritePump, что нужно закрыть соединение
				close(conn.Send)
				connectionCount := len(h.connections)
				slog.Info("Connection unregistered", "user_id", conn.UserID, "total_connections", connectionCount)
			}
			h.mu.Unlock() // Разблокируем

		case message := <-h.broadcast:
			// Сообщение для широковещательной рассылки всем подключенным пользователям
			h.mu.RLock() // Блокируем для чтения (RWMutex позволяет множественное чтение)
			connectionCount := len(h.connections)
			sentCount := 0
			// Отправляем сообщение всем активным соединениям
			for _, conn := range h.connections {
				select {
				case conn.Send <- message:
					// Сообщение успешно отправлено в канал соединения
					sentCount++
				default:
					// Канал переполнен или закрыт - соединение "мёртвое"
					// Закрываем соединение и удаляем из map
					close(conn.Send)
					delete(h.connections, conn.UserID)
				}
			}
			h.mu.RUnlock() // Разблокируем
			if connectionCount > 0 {
				slog.Debug("Message broadcast", "total_connections", connectionCount, "sent_to", sentCount)
			}
		}
	}
}

// Broadcast отправляет сообщение всем подключенным клиентам
// Используется для глобальных уведомлений (например, системные сообщения)
// Параметры:
//   - message: JSON байты сообщения для отправки
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser отправляет сообщение конкретному пользователю
// Используется для приватных сообщений (например, DM, личные уведомления)
// Параметры:
//   - userID: ID пользователя-получателя
//   - message: JSON байты сообщения для отправки
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()         // Блокируем для чтения
	defer h.mu.RUnlock() // Разблокируем при выходе из функции

	// Проверяем, есть ли активное соединение у пользователя
	if conn, ok := h.connections[userID]; ok {
		select {
		case conn.Send <- message:
			// Сообщение успешно отправлено
			slog.Debug("Message sent to user", "user_id", userID)
		default:
			// Канал переполнен - соединение "мёртвое"
			// Закрываем и удаляем (но не можем удалить здесь из-за RLock)
			slog.Warn("Failed to send message to user, closing connection", "user_id", userID)
			close(conn.Send)
			delete(h.connections, userID)
		}
	} else {
		// Пользователь не подключен - сообщение теряется
		// В production можно добавить очередь сообщений для офлайн пользователей
		slog.Debug("User not connected, message not sent", "user_id", userID)
	}
}

// SendToWorkspace отправляет сообщение всем пользователям в workspace
// Используется для workspace-wide уведомлений
// Параметры:
//   - ctx: контекст для запросов к базе данных
//   - workspaceID: ID workspace
//   - message: JSON байты сообщения для отправки
func (h *Hub) SendToWorkspace(ctx context.Context, workspaceID int64, message []byte) {
	slog.Info("Sending message to workspace", "workspace_id", workspaceID)

	h.mu.RLock()         // Блокируем для чтения
	defer h.mu.RUnlock() // Разблокируем при выходе

	// Получаем список участников workspace из базы данных
	// Это необходимо, так как не все участники могут быть онлайн
	slog.Debug("Loading workspace members", "workspace_id", workspaceID)
	members, err := h.ORM.WorkspaceMember.
		Query().                                                // Начинаем запрос
		Where(workspacemember.WorkspaceIDEQ(int(workspaceID))). // Фильтр: только участники этого workspace
		WithUser().                                             // Загружаем связанных пользователей через edge
		All(ctx)                                                // Получаем всех участников

	if err != nil {
		// Если не удалось получить участников, используем fallback - рассылаем всем
		// Это не идеально, но лучше чем потерять сообщение
		slog.Error("Failed to get workspace members for SendToWorkspace", "workspace_id", workspaceID, "error", err)
		slog.Warn("Falling back to broadcast to all connections", "workspace_id", workspaceID)
		sentCount := 0
		for _, conn := range h.connections {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				// Соединение "мёртвое" - закрываем и удаляем
				close(conn.Send)
				delete(h.connections, conn.UserID)
			}
		}
		slog.Info("Message broadcasted to all (fallback)", "workspace_id", workspaceID, "sent_to", sentCount)
		return
	}
	slog.Info("Workspace members loaded", "workspace_id", workspaceID, "members_count", len(members))

	// Создаём set (map) ID пользователей, которые являются участниками
	// Это позволяет быстро проверить, является ли пользователь участником
	memberUserIDs := make(map[int64]bool)
	for _, member := range members {
		memberUserIDs[int64(member.UserID)] = true
	}

	// Отправляем сообщение только участникам workspace, которые онлайн
	sentCount := 0
	for userID, conn := range h.connections {
		if memberUserIDs[userID] {
			// Пользователь является участником workspace и онлайн
			select {
			case conn.Send <- message:
				sentCount++
			default:
				// Соединение "мёртвое"
				slog.Warn("Failed to send message to workspace member, closing connection", "workspace_id", workspaceID, "user_id", userID)
				close(conn.Send)
				delete(h.connections, userID)
			}
		}
	}
	slog.Info("Message sent to workspace members", "workspace_id", workspaceID, "sent_to", sentCount, "total_members", len(members))
}

// SendToChannel отправляет сообщение всем пользователям в канале
// Используется для сообщений в каналах (новые сообщения, события канала)
// Параметры:
//   - channelID: ID канала
//   - message: JSON байты сообщения для отправки
func (h *Hub) SendToChannel(channelID int64, message []byte) {
	slog.Info("Sending message to channel", "channel_id", channelID)

	// Получаем список участников канала из базы данных
	slog.Debug("Loading channel members", "channel_id", channelID)
	members, err := h.ORM.ChannelMember.
		Query().                                          // Начинаем запрос
		Where(channelmember.ChannelIDEQ(int(channelID))). // Фильтр: только участники этого канала
		All(context.Background())                         // Получаем всех участников (используем Background context, так как это не HTTP запрос)

	if err != nil {
		// Если не удалось получить участников, используем fallback - рассылаем всем
		// Это не идеально, но лучше чем потерять сообщение
		slog.Error("Failed to get channel members, falling back to broadcast", "channel_id", channelID, "error", err)
		h.mu.RLock() // Блокируем для чтения
		sentCount := 0
		for _, conn := range h.connections {
			select {
			case conn.Send <- message:
				sentCount++
			default:
				// Соединение "мёртвое" - закрываем и удаляем
				close(conn.Send)
				delete(h.connections, conn.UserID)
			}
		}
		h.mu.RUnlock() // Разблокируем
		slog.Info("Message broadcasted to all (fallback)", "channel_id", channelID, "sent_to", sentCount)
		return
	}
	slog.Info("Channel members loaded", "channel_id", channelID, "members_count", len(members))

	// Создаём set (map) ID пользователей, которые являются участниками канала
	memberUserIDs := make(map[int64]bool)
	for _, member := range members {
		memberUserIDs[int64(member.UserID)] = true
	}

	// Отправляем сообщение только участникам канала, которые онлайн
	h.mu.RLock() // Блокируем для чтения
	sentCount := 0
	for userID, conn := range h.connections {
		if memberUserIDs[userID] {
			// Пользователь является участником канала и онлайн
			select {
			case conn.Send <- message:
				sentCount++
			default:
				// Соединение "мёртвое"
				slog.Warn("Failed to send message to channel member, closing connection", "channel_id", channelID, "user_id", userID)
				close(conn.Send)
				delete(h.connections, userID)
			}
		}
	}
	h.mu.RUnlock() // Разблокируем
	slog.Info("Message sent to channel members", "channel_id", channelID, "sent_to", sentCount, "total_members", len(members))
}

// Register регистрирует новое соединение
// Вызывается при подключении нового клиента
// Параметры:
//   - conn: новое WebSocket соединение
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// GetConnectionCount возвращает количество активных соединений
// Используется для мониторинга и статистики
// Возвращает:
//   - int: количество активных соединений
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()         // Блокируем для чтения
	defer h.mu.RUnlock() // Разблокируем при выходе
	return len(h.connections)
}

// IsUserOnline проверяет, онлайн ли пользователь
// Используется для отображения статуса онлайн/офлайн
// Параметры:
//   - userID: ID пользователя для проверки
//
// Возвращает:
//   - bool: true если пользователь онлайн, false если офлайн
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()         // Блокируем для чтения
	defer h.mu.RUnlock() // Разблокируем при выходе
	_, ok := h.connections[userID]
	return ok
}
