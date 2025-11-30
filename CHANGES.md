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

### [2024-11-30] Фаза 2: WebSocket инфраструктура

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

---

## Примечания

- Все изменения в Pagoda core файлах должны быть задокументированы здесь
- При синхронизации с upstream проверяйте этот файл на актуальность
- Используйте маркеры `CUSTOM CODE START/END` в коде для обозначения ваших изменений

