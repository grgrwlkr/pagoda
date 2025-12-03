package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/message"
)

// Константы для настройки WebSocket соединения
const (
	// writeWait - время, отведённое на запись сообщения клиенту
	// Если запись не завершится за это время, соединение считается "мёртвым"
	writeWait = 10 * time.Second

	// pongWait - время ожидания следующего pong сообщения от клиента
	// Pong - это ответ на ping, используется для проверки, что соединение живо
	// Если pong не приходит, соединение закрывается
	pongWait = 60 * time.Second

	// pingPeriod - период отправки ping сообщений клиенту
	// Должен быть меньше pongWait, чтобы у клиента было время ответить
	// Обычно 9/10 от pongWait
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize - максимальный размер сообщения от клиента
	// Защита от слишком больших сообщений, которые могут забить память
	maxMessageSize = 512 * 1024 // 512KB
)

// Connection представляет одно WebSocket соединение
// Это посредник между WebSocket соединением и Hub
// Каждое соединение имеет два goroutine: ReadPump (чтение) и WritePump (запись)
type Connection struct {
	// WebSocket соединение (из библиотеки gorilla/websocket)
	WS *websocket.Conn

	// Буферизованный канал исходящих сообщений
	// Hub отправляет сообщения в этот канал, WritePump читает и отправляет клиенту
	Send chan []byte

	// ID пользователя, связанного с этим соединением
	// Используется для идентификации пользователя и маршрутизации сообщений
	UserID int64

	// Ссылка на Hub для отправки событий
	Hub *Hub

	// Контекст для операций с базой данных
	// Используется при обработке событий, требующих доступа к БД
	Ctx context.Context

	// ORM клиент для операций с базой данных
	// Используется для проверки прав доступа (например, является ли пользователь участником канала)
	ORM *ent.Client

	// Логгер для логирования событий соединения
	Logger *slog.Logger

	// onceClose гарантирует, что Close вызывается только один раз
	onceClose sync.Once
}

// ReadPump читает сообщения из WebSocket соединения и отправляет их в Hub
// Запускается в отдельной goroutine для каждого соединения
// Работает до закрытия соединения или ошибки чтения
func (c *Connection) ReadPump() {
	if c.Logger != nil {
		c.Logger.Info("WebSocket read pump started", "user_id", c.UserID)
	}

	// defer выполнится при выходе из функции (нормальном или из-за ошибки)
	defer func() {
		if c.Logger != nil {
			c.Logger.Info("WebSocket read pump closing", "user_id", c.UserID)
		}

		// Отменяем регистрацию соединения в Hub ПЕРЕД отправкой событий
		// Это предотвращает отправку событий самому себе
		c.Hub.unregister <- c

		// Отправляем событие "пользователь офлайн" после отмены регистрации
		// Это позволяет другим пользователям видеть, что пользователь отключился
		// Но не отправляем самому себе, так как соединение уже удалено из map
		offlineEvent := UserOfflineEvent(c.UserID)
		c.Hub.Broadcast(offlineEvent.ToJSON())

		// Закрываем WebSocket соединение
		c.WS.Close()
	}()

	// Устанавливаем таймаут для чтения (pongWait)
	// Если за это время не будет получено сообщение, соединение считается "мёртвым"
	c.WS.SetReadDeadline(time.Now().Add(pongWait))
	// Устанавливаем максимальный размер сообщения
	c.WS.SetReadLimit(maxMessageSize)
	// Устанавливаем обработчик pong сообщений
	// Pong - это ответ на ping, используется для keep-alive
	c.WS.SetPongHandler(func(string) error {
		// При получении pong, обновляем таймаут чтения
		c.WS.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Бесконечный цикл чтения сообщений
	for {
		// Читаем сообщение из WebSocket
		// ReadMessage блокируется до получения сообщения или ошибки
		_, message, err := c.WS.ReadMessage()
		if err != nil {
			// Проверяем, является ли ошибка неожиданным закрытием
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Неожиданное закрытие - логируем как предупреждение
				if c.Logger != nil {
					c.Logger.Warn("WebSocket unexpected close", "user_id", c.UserID, "error", err)
				}
			} else if c.Logger != nil {
				// Ожидаемое закрытие (нормальное отключение) - логируем как debug
				c.Logger.Debug("WebSocket read error (normal close)", "user_id", c.UserID, "error", err)
			}
			break // Выходим из цикла при любой ошибке
		}

		// Парсим входящее сообщение (ожидается JSON)
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			// Не удалось распарсить JSON - отправляем ошибку клиенту
			if c.Logger != nil {
				c.Logger.Warn("Failed to parse WebSocket message", "user_id", c.UserID, "error", err, "message_length", len(message))
			}
			c.SendError("invalid_message_format", "Failed to parse message")
			continue // Продолжаем чтение следующих сообщений
		}

		// Логируем входящее событие
		if c.Logger != nil {
			c.Logger.Info("WebSocket event received", "user_id", c.UserID, "event_type", event.Type)
		}

		// Обрабатываем событие
		c.handleEvent(&event)
	}
}

// WritePump отправляет сообщения из канала Send клиенту через WebSocket
// Запускается в отдельной goroutine для каждого соединения
// Работает до закрытия канала Send или ошибки записи
func (c *Connection) WritePump() {
	if c.Logger != nil {
		c.Logger.Info("WebSocket write pump started", "user_id", c.UserID)
	}

	// Создаём тикер для периодической отправки ping сообщений
	// Ping используется для keep-alive - проверки, что соединение живо
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop() // Останавливаем тикер при выходе
		if c.Logger != nil {
			c.Logger.Info("WebSocket write pump closing", "user_id", c.UserID)
		}
		c.WS.Close() // Закрываем WebSocket соединение
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// Устанавливаем таймаут для записи
			c.WS.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub закрыл канал - отправляем close сообщение и выходим
				c.WS.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Получаем writer для отправки текстового сообщения
			w, err := c.WS.NextWriter(websocket.TextMessage)
			if err != nil {
				if c.Logger != nil {
					c.Logger.Warn("WebSocket write error", "user_id", c.UserID, "error", err)
				}
				return // Ошибка получения writer - выходим
			}
			// Записываем сообщение
			w.Write(message)

			// Добавляем накопленные в канале сообщения к текущему сообщению
			// Это оптимизация - отправляем несколько сообщений одним пакетом
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'}) // Разделитель между сообщениями
				w.Write(<-c.Send)     // Читаем следующее сообщение из канала
			}

			// Закрываем writer (отправляет данные клиенту)
			if err := w.Close(); err != nil {
				if c.Logger != nil {
					c.Logger.Warn("WebSocket writer close error", "user_id", c.UserID, "error", err)
				}
				return // Ошибка закрытия - выходим
			}

		case <-ticker.C:
			// Время отправить ping сообщение
			c.WS.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WS.WriteMessage(websocket.PingMessage, nil); err != nil {
				if c.Logger != nil {
					c.Logger.Warn("WebSocket ping error", "user_id", c.UserID, "error", err)
				}
				return // Ошибка отправки ping - выходим
			}
		}
	}
}

// handleEvent обрабатывает входящие события от клиента
// События могут быть: join_channel, leave_channel, typing_start, typing_stop, message_send, mark_read
// Параметры:
//   - event: событие от клиента
func (c *Connection) handleEvent(event *Event) {
	switch event.Type {
	case EventTypeJoinChannel:
		// Пользователь присоединился к каналу
		c.handleJoinChannel(event)
	case EventTypeLeaveChannel:
		// Пользователь покинул канал
		c.handleLeaveChannel(event)
	case EventTypeTypingStart:
		// Пользователь начал печатать
		c.handleTypingStart(event)
	case EventTypeTypingStop:
		// Пользователь перестал печатать
		c.handleTypingStop(event)
	case EventTypeMessageSend:
		// Пользователь отправил сообщение через WebSocket
		c.handleMessageSend(event)
	case EventTypeMarkRead:
		// Пользователь прочитал сообщения (отметил как прочитанные)
		c.handleMarkRead(event)
	default:
		// Неизвестный тип события - отправляем ошибку клиенту
		c.SendError("unknown_event_type", "Unknown event type: "+string(event.Type))
	}
}

// handleJoinChannel обрабатывает событие присоединения пользователя к каналу
// Проверяет, является ли пользователь участником канала, и уведомляет других участников
func (c *Connection) handleJoinChannel(event *Event) {
	// Извлекаем channel_id из данных события
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Проверяем, является ли пользователь участником канала
	// Это важно для безопасности - нельзя присоединиться к каналу, где ты не участник
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		// Пользователь не является участником канала
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Уведомляем других участников канала о присоединении (но не самого пользователя)
	response := MemberJoinedEvent(chID, c.UserID)
	c.Hub.SendToChannelExcluding(chID, c.UserID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("User joined channel", "user_id", c.UserID, "channel_id", chID)
	}
}

// handleLeaveChannel обрабатывает событие выхода пользователя из канала
// Проверяет права доступа и уведомляет других участников
func (c *Connection) handleLeaveChannel(event *Event) {
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Проверяем, является ли пользователь участником канала
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Уведомляем других участников канала о выходе
	response := &Event{
		Type: EventTypeMemberLeft,
		Data: map[string]interface{}{
			"channel_id": chID,
			"user_id":    c.UserID,
		},
	}
	c.Hub.SendToChannel(chID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("User left channel", "user_id", c.UserID, "channel_id", chID)
	}
}

// handleTypingStart обрабатывает событие начала печати пользователя
// Отправляет индикатор печати другим участникам канала
func (c *Connection) handleTypingStart(event *Event) {
	// Транслируем индикатор печати участникам канала
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Проверяем, является ли пользователь участником канала
	// Тихие ошибки - если не участник, просто игнорируем событие
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil || !exists {
		// Тихие ошибки для typing событий - не отправляем ошибку клиенту
		return
	}

	// Отправляем событие "пользователь печатает" всем участникам канала
	response := UserTypingEvent(c.UserID, chID)
	c.Hub.SendToChannel(chID, response.ToJSON())
}

// handleTypingStop обрабатывает событие окончания печати пользователя
// Обычно обрабатывается на клиенте, но мы валидируем права доступа
func (c *Connection) handleTypingStop(event *Event) {
	// Остановка печати обычно обрабатывается на клиенте
	// Мы можем реализовать это, если нужно для серверной очистки
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		return
	}

	chID := int64(channelID)

	// Проверяем, является ли пользователь участником канала
	// Тихие ошибки - если не участник, просто игнорируем событие
	_, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		// Тихие ошибки для typing stop
		return
	}

	// Можно отправить typing_stop событие, если нужно
	// Пока просто валидируем и возвращаем
}

// handleMessageSend обрабатывает отправку сообщения через WebSocket
// Создаёт сообщение в базе данных и транслирует его участникам канала
// Примечание: Обычно сообщения отправляются через HTTP (MessageCreate), но WebSocket тоже поддерживается
func (c *Connection) handleMessageSend(event *Event) {
	// Извлекаем данные сообщения из события
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	content, ok := event.Data["content"].(string)
	if !ok || content == "" {
		c.SendError("invalid_content", "Message content is required")
		return
	}

	chID := int64(channelID)

	// Проверяем, является ли пользователь участником канала
	exists, err := c.ORM.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		Exist(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to check channel membership", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to verify channel membership")
		return
	}

	if !exists {
		c.SendError("not_member", "You are not a member of this channel")
		return
	}

	// Создаём сообщение в базе данных
	msg, err := c.ORM.Message.
		Create().
		SetContent(content).
		SetMessageType(message.MessageTypeText).
		SetChannelID(int(chID)).
		SetUserID(int(c.UserID)).
		Save(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to save message", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to save message")
		return
	}

	// Транслируем сообщение всем участникам канала
	response := MessageNewEvent(int64(msg.ID), chID, c.UserID, content)
	c.Hub.SendToChannel(chID, response.ToJSON())

	if c.Logger != nil {
		c.Logger.Info("Message sent", "user_id", c.UserID, "channel_id", chID, "message_id", msg.ID)
	}
}

// handleMarkRead обрабатывает событие отметки сообщений как прочитанных
// Обновляет last_read_at для пользователя в канале
func (c *Connection) handleMarkRead(event *Event) {
	channelID, ok := event.Data["channel_id"].(float64)
	if !ok {
		c.SendError("invalid_channel_id", "Invalid channel ID")
		return
	}

	chID := int64(channelID)

	// Обновляем last_read_at для пользователя в этом канале
	// Это используется для подсчёта непрочитанных сообщений
	_, err := c.ORM.ChannelMember.
		Update().
		Where(channelmember.ChannelIDEQ(int(chID))).
		Where(channelmember.UserIDEQ(int(c.UserID))).
		SetLastReadAt(time.Now()).
		Save(c.Ctx)

	if err != nil {
		if c.Logger != nil {
			c.Logger.Error("Failed to update last_read_at", "user_id", c.UserID, "channel_id", chID, "error", err)
		}
		c.SendError("database_error", "Failed to update read status")
		return
	}

	if c.Logger != nil {
		c.Logger.Debug("Read receipt updated", "user_id", c.UserID, "channel_id", chID)
	}
}

// SendError отправляет сообщение об ошибке клиенту
// Используется для уведомления клиента о проблемах (неверные данные, ошибки БД и т.д.)
// Параметры:
//   - code: код ошибки (например, "invalid_channel_id")
//   - message: текстовое описание ошибки
func (c *Connection) SendError(code, message string) {
	response := &Event{
		Type: EventTypeError,
		Data: map[string]interface{}{
			"error_code":    code,
			"error_message": message,
		},
	}
	c.Send <- response.ToJSON()
}

// Close закрывает соединение
// Закрывает WebSocket соединение и канал Send
// Использует sync.Once для гарантии, что закрытие происходит только один раз
func (c *Connection) Close() {
	c.onceClose.Do(func() {
		if c.WS != nil {
			c.WS.Close()
		}
		if c.Send != nil {
			close(c.Send)
		}
	})
}
