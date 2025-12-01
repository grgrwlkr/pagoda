# Изменения в Slack-клоне

Этот файл отслеживает все изменения, сделанные поверх оригинального проекта Pagoda.

## Структура проекта

### Добавленные директории

- `app/` - весь пользовательский код для Slack-клона
  - `app/handlers/` - HTTP handlers для мессенджера
  - `app/services/` - сервисы (notifications, websocket)
  - `app/ui/` - UI компоненты, страницы, формы
  - `app/websocket/` - WebSocket инфраструктура
  - `app/middleware/` - кастомные middleware

## Модели данных (Ent)

### Добавленные сущности

- `ent/schema/workspace.go` - рабочее пространство
- `ent/schema/channel.go` - канал
- `ent/schema/message.go` - сообщение
- `ent/schema/directmessage.go` - прямое сообщение
- `ent/schema/directmessagecontent.go` - содержимое DM
- `ent/schema/reaction.go` - реакция на сообщение
- `ent/schema/attachment.go` - вложение файла
- `ent/schema/channelmember.go` - участник канала
- `ent/schema/workspacemember.go` - участник workspace
- `ent/schema/notification.go` - уведомление

### Модифицированные сущности

- `ent/schema/user.go` - НЕ ИЗМЕНЯЕТСЯ напрямую
  - Используется связь с UserProfile (если нужно расширение)

## Конфигурация

### Модифицированные файлы

- `config/config.yaml` - добавлена секция `messenger:`

### Добавленные файлы

- `config/app.yaml` - только пользовательские настройки (опционально)

## Handlers

### Добавленные handlers

- `app/handlers/messenger.go` - основной handler для мессенджера

### Модифицированные handlers

- `pkg/handlers/router.go` - добавлен WebSocket endpoint (если нужно)

## UI

### Добавленные компоненты

- `app/ui/components/messenger/` - все компоненты мессенджера
- `app/ui/pages/messenger/` - страницы мессенджера
- `app/ui/forms/messenger/` - формы мессенджера
- `app/ui/layouts/messenger.go` - layout для мессенджера

## Сервисы

### Добавленные сервисы

- `app/services/notifications.go` - сервис уведомлений
- `app/websocket/hub.go` - WebSocket hub
- `app/websocket/connection.go` - обработка соединений

## Middleware

### Добавленные middleware

- `app/middleware/workspace.go` - проверка членства в workspace
- `app/middleware/channel.go` - проверка членства в канале

## Зависимости

### Добавленные зависимости

- (добавить по мере необходимости)

## История изменений

### [Unreleased]

- Начало разработки Slack-клона
- Создана структура проекта
- Настроена синхронизация с Pagoda upstream

### [2024-11-30] Фаза 1: Модели данных

- ✅ Созданы все схемы Ent для Slack-клона:
  - `workspace.go` - рабочее пространство
  - `channel.go` - канал
  - `message.go` - сообщение
  - `directmessage.go` - прямое сообщение
  - `directmessagecontent.go` - содержимое DM
  - `reaction.go` - реакция на сообщение
  - `attachment.go` - вложение файла
  - `channelmember.go` - участник канала
  - `workspacemember.go` - участник workspace
  - `userprofile.go` - профиль пользователя (расширение User)
  - `notification.go` - уведомление
- ✅ Сгенерирован код Ent: `make ent-gen`
- ✅ Все связи (edges) настроены
- ✅ Добавлены валидации полей
- ✅ Добавлен уникальный индекс для DirectMessage (user1_id, user2_id)

### [2024-11-30] Фаза 2: WebSocket инфраструктура (начало)

- ✅ Создан WebSocket Hub (`app/websocket/hub.go`)
  - Управление подключениями по user ID
  - Broadcast механизм
  - Отправка сообщений конкретным пользователям
  - Отправка сообщений в каналы
- ✅ Создан WebSocket Connection (`app/websocket/connection.go`)
  - Обработка чтения/записи сообщений
  - Ping/pong для поддержания соединения
  - Обработка событий от клиента
- ✅ Созданы типы событий (`app/websocket/events.go`)
  - События от клиента: join_channel, typing_start, message_send и др.
  - События от сервера: message_new, user_online, reaction_added и др.
- ✅ Созданы типы сообщений (`app/websocket/messages.go`)
  - Структуры для сообщений, реакций, статусов
- ✅ Создан WebSocket Handler (`app/handlers/websocket.go`)
  - Endpoint `/ws` с аутентификацией
  - Интеграция с Echo
  - Автоматическая регистрация через handlers system
- ✅ Добавлена зависимость: `github.com/gorilla/websocket`
- ✅ Добавлен route name: `WebSocket` в `pkg/routenames/names.go`

### [2024-12-01] Фаза 4: Обработчики и API (базовая версия)

- ✅ Создан основной handler (`app/handlers/messenger.go`)
  - Все основные endpoints для workspace, channel, message, DM, reactions, attachments
  - Интеграция с Ent ORM
  - Обработка ошибок
- ✅ Созданы middleware (`app/middleware/`):
  - `workspace.go` - LoadWorkspace, RequireWorkspaceMember
  - `channel.go` - LoadChannel, RequireChannelMember
- ✅ Добавлены route names (`app/routenames/names.go`)
  - Все route names для messenger endpoints
- ✅ Интеграция middleware в routes
  - Все защищенные routes проверяют членство
  - Workspace и Channel загружаются в контекст

### [2024-12-01] Фаза 3: Базовый UI и навигация

- ✅ Создан Messenger layout (`app/ui/layouts/messenger.go`)
  - Трехпанельный layout (sidebar, content, right panel)
  - Интеграция с HTMX и Alpine.js
- ✅ Созданы UI компоненты (`app/ui/components/messenger/`):
  - `sidebar.go` - боковая панель с workspace, channels, DMs
  - `channel_list.go` - список каналов
  - `message_list.go` - список сообщений с реакциями
  - `message_input.go` - форма ввода сообщения
  - `user_avatar.go` - аватар пользователя с инициалами
  - `typing_indicator.go` - индикатор набора текста
- ✅ Созданы страницы (`app/ui/pages/messenger/`):
  - `workspace.go` - главная страница workspace
  - `channel.go` - страница канала
  - `direct_message.go` - страница прямого сообщения

### [2024-12-01] Фаза 2: WebSocket инфраструктура (завершение)

- ✅ Реализованы все обработчики событий в `connection.go`:
  - `handleJoinChannel` - проверка членства и уведомление участников
  - `handleLeaveChannel` - проверка членства и уведомление участников
  - `handleTypingStart` - валидация и broadcast индикатора набора
  - `handleTypingStop` - валидация членства
  - `handleMessageSend` - сохранение в БД и broadcast сообщения
  - `handleMarkRead` - обновление last_read_at в ChannelMember
- ✅ Оптимизирован `SendToChannel` в `hub.go`:
  - Получение участников канала из БД
  - Отправка сообщений только участникам канала
  - Fallback на broadcast при ошибке получения участников
- ✅ Улучшен `CheckOrigin` в `websocket.go`:
  - Проверка origin из запроса
  - Поддержка localhost для разработки
  - TODO: добавление конфигурации для production
- ✅ Добавлено логирование:
  - Логирование всех WebSocket событий через slog
  - Логирование ошибок и важных операций
  - Передача logger через context
- ✅ Интеграция с БД:
  - Connection теперь имеет доступ к ORM и context
  - Все операции сохраняются в базу данных
  - Валидация членства в каналах перед операциями

---

## Примечания

- Все изменения в Pagoda core файлах должны быть задокументированы здесь
- При синхронизации с upstream проверяйте этот файл на актуальность
- Используйте маркеры `CUSTOM CODE START/END` в коде для обозначения ваших изменений

