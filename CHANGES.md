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

---

## Примечания

- Все изменения в Pagoda core файлах должны быть задокументированы здесь
- При синхронизации с upstream проверяйте этот файл на актуальность
- Используйте маркеры `CUSTOM CODE START/END` в коде для обозначения ваших изменений

