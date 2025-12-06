# План разработки копии Slack на базе Pagoda

## Обзор проекта

Создание полнофункционального мессенджера для команд с каналами, прямыми сообщениями, файлами, уведомлениями и другими функциями Slack.

---

## Фаза 1: Базовая инфраструктура и модели данных (2-3 недели)

### 1.1 Расширение моделей данных (Ent)

**Создать новые сущности:**

1. **Workspace** (Рабочее пространство)
   - Поля: `name`, `slug`, `description`, `owner_id`, `created_at`, `updated_at`
   - Edges: `owner` (User), `members` (User), `channels` (Channel)

2. **Channel** (Канал)
   - Поля: `name`, `slug`, `description`, `is_private`, `workspace_id`, `created_by`, `created_at`, `updated_at`
   - Edges: `workspace` (Workspace), `creator` (User), `members` (User), `messages` (Message)

3. **Message** (Сообщение)
   - Поля: `content`, `message_type` (text, file, thread_reply), `channel_id`, `user_id`, `thread_id` (nullable), `created_at`, `updated_at`, `edited_at` (nullable)
   - Edges: `channel` (Channel), `user` (User), `thread` (Message), `replies` (Message), `reactions` (Reaction), `attachments` (Attachment)

4. **DirectMessage** (Прямое сообщение)
   - Поля: `user1_id`, `user2_id`, `last_message_at`, `created_at`
   - Edges: `user1` (User), `user2` (User), `messages` (DirectMessageContent)

5. **DirectMessageContent** (Содержимое прямого сообщения)
   - Поля: `content`, `dm_id`, `user_id`, `created_at`, `edited_at` (nullable)
   - Edges: `dm` (DirectMessage), `user` (User), `attachments` (Attachment)

6. **Reaction** (Реакция)
   - Поля: `emoji`, `message_id`, `user_id`, `created_at`
   - Edges: `message` (Message), `user` (User)

7. **Attachment** (Вложение)
   - Поля: `filename`, `filepath`, `file_size`, `mime_type`, `message_id` (nullable), `dm_content_id` (nullable), `uploaded_by`, `created_at`
   - Edges: `message` (Message), `dm_content` (DirectMessageContent), `uploader` (User)

8. **ChannelMember** (Участник канала)
   - Поля: `channel_id`, `user_id`, `joined_at`, `last_read_at`
   - Edges: `channel` (Channel), `user` (User)

9. **WorkspaceMember** (Участник рабочего пространства)
   - Поля: `workspace_id`, `user_id`, `role` (owner, admin, member), `joined_at`
   - Edges: `workspace` (Workspace), `user` (User)

10. **User** (Расширение существующей)
    - Добавить поля: `avatar_url`, `status` (online, away, busy, offline), `status_message`, `timezone`
    - Добавить edges: `workspaces`, `channels`, `messages`, `direct_messages_sent`, `direct_messages_received`

**Задачи:**
- Создать схемы в `ent/schema/`
- Настроить валидацию полей
- Определить все связи (edges)
- Сгенерировать код: `make ent-gen`

### 1.2 Миграции и тестовая БД
- Проверить автоматические миграции
- Настроить тестовую БД для разработки

---

## Фаза 2: WebSocket инфраструктура (1-2 недели)

### 2.1 WebSocket сервер
**Создать пакет `pkg/websocket/`:**

1. **Hub** (Центральный узел)
   - Управление подключениями
   - Маршрутизация сообщений
   - Broadcast механизм

2. **Connection** (Подключение)
   - Обертка над WebSocket соединением
   - Привязка к пользователю
   - Обработка сообщений

3. **Message types** (Типы сообщений)
   - Структуры для различных типов событий
   - Сериализация/десериализация JSON

**Файлы:**
- `pkg/websocket/hub.go` - центральный хаб
- `pkg/websocket/connection.go` - обработка соединений
- `pkg/websocket/messages.go` - типы сообщений
- `pkg/websocket/events.go` - события (message_sent, user_typing, user_online, etc.)

### 2.2 Интеграция с Echo
- Добавить WebSocket endpoint в роутер
- Middleware для аутентификации WebSocket
- Обработка upgrade соединений

### 2.3 HTMX WebSocket поддержка
- Настроить HTMX для работы с WebSocket
- Компоненты для подключения к WebSocket

---

## Фаза 3: Базовый UI и навигация (2 недели)

### 3.1 Новый layout для мессенджера
**Создать `pkg/ui/layouts/messenger.go`:**

- Левая панель:
  - Список рабочих пространств
  - Список каналов текущего workspace
  - Список прямых сообщений
  - Кнопка создания канала
  - Профиль пользователя

- Центральная панель:
  - Заголовок канала/диалога
  - Область сообщений (scrollable)
  - Форма ввода сообщения
  - Кнопка загрузки файлов

- Правая панель (опционально):
  - Информация о канале
  - Участники
  - Файлы

### 3.2 Компоненты UI
**Создать в `pkg/ui/components/messenger/`:**

- `sidebar.go` - боковая панель
- `channel_list.go` - список каналов
- `message_list.go` - список сообщений
- `message_input.go` - поле ввода
- `user_avatar.go` - аватар пользователя
- `typing_indicator.go` - индикатор набора текста
- `reaction_picker.go` - выбор реакции
- `file_upload.go` - загрузка файлов

### 3.3 Страницы
**Создать в `pkg/ui/pages/messenger/`:**

- `workspace.go` - главная страница workspace
- `channel.go` - страница канала
- `direct_message.go` - страница прямого сообщения
- `channel_settings.go` - настройки канала

---

## Фаза 4: Обработчики и API (2-3 недели)

### 4.1 Handlers
**Создать `pkg/handlers/messenger.go`:**

**Workspace endpoints:**
- `GET /workspace` - список workspace пользователя
- `GET /workspace/:id` - информация о workspace
- `POST /workspace` - создание workspace
- `PUT /workspace/:id` - обновление workspace
- `DELETE /workspace/:id` - удаление workspace
- `POST /workspace/:id/members` - добавление участника
- `DELETE /workspace/:id/members/:user_id` - удаление участника

**Channel endpoints:**
- `GET /workspace/:workspace_id/channels` - список каналов
- `GET /channel/:id` - информация о канале
- `POST /workspace/:workspace_id/channels` - создание канала
- `PUT /channel/:id` - обновление канала
- `DELETE /channel/:id` - удаление канала
- `POST /channel/:id/members` - добавление участника
- `DELETE /channel/:id/members/:user_id` - удаление участника
- `GET /channel/:id/messages` - получение сообщений (с пагинацией)

**Message endpoints:**
- `POST /channel/:channel_id/messages` - отправка сообщения
- `PUT /message/:id` - редактирование сообщения
- `DELETE /message/:id` - удаление сообщения
- `GET /message/:id/replies` - получение ответов в треде
- `POST /message/:id/replies` - ответ в треде

**Direct Message endpoints:**
- `GET /dms` - список прямых сообщений
- `GET /dm/:id` - информация о DM
- `POST /dm` - создание/получение DM с пользователем
- `GET /dm/:id/messages` - получение сообщений
- `POST /dm/:id/messages` - отправка сообщения

**Reaction endpoints:**
- `POST /message/:id/reactions` - добавление реакции
- `DELETE /message/:id/reactions/:emoji` - удаление реакции

**File endpoints:**
- `POST /message/:id/attachments` - загрузка файла
- `GET /attachment/:id` - скачивание файла
- `DELETE /attachment/:id` - удаление файла

**WebSocket endpoint:**
- `WS /ws` - WebSocket соединение

### 4.2 Формы
**Создать в `pkg/ui/forms/messenger/`:**

- `workspace.go` - форма создания/редактирования workspace
- `channel.go` - форма создания/редактирования канала
- `message.go` - форма отправки сообщения
- `invite.go` - форма приглашения пользователя

### 4.3 Middleware
- `RequireWorkspaceMember` - проверка членства в workspace
- `RequireChannelMember` - проверка членства в канале
- `LoadWorkspace` - загрузка workspace в контекст
- `LoadChannel` - загрузка канала в контекст

---

## Фаза 5: Реальное время (1-2 недели)

### 5.1 WebSocket события

**События от клиента:**
- `join_channel` - присоединение к каналу
- `leave_channel` - выход из канала
- `typing_start` - начало набора текста
- `typing_stop` - окончание набора текста
- `message_send` - отправка сообщения (через WebSocket)
- `mark_read` - отметка о прочтении

**События от сервера:**
- `message_new` - новое сообщение
- `message_edited` - сообщение отредактировано
- `message_deleted` - сообщение удалено
- `user_typing` - пользователь печатает
- `user_online` - пользователь онлайн
- `user_offline` - пользователь офлайн
- `reaction_added` - добавлена реакция
- `reaction_removed` - удалена реакция
- `channel_updated` - канал обновлен
- `member_joined` - участник присоединился
- `member_left` - участник покинул

### 5.2 Интеграция с базой данных
- При создании сообщения через HTTP API отправлять событие через WebSocket
- Отслеживание статуса пользователей (online/offline)
- Отметка о прочтении сообщений

### 5.3 Индикаторы
- Индикатор набора текста
- Индикатор онлайн статуса
- Непрочитанные сообщения (badge)

---

## Фаза 6: Поиск и фильтрация (1 неделя)

### 6.1 Поиск сообщений
**Расширить `pkg/handlers/search.go`:**

- Поиск по содержимому сообщений
- Фильтрация по каналу, пользователю, дате
- Полнотекстовый поиск (FTS для SQLite)

### 6.2 Поиск пользователей
- Поиск пользователей в workspace
- Поиск по имени, email

### 6.3 UI для поиска
- Компонент поиска в sidebar
- Страница результатов поиска
- Подсветка найденного текста

---

## Фаза 7: Файлы и вложения (1-2 недели)

### 7.1 Загрузка файлов
- Расширить существующий `pkg/handlers/files.go`
- Поддержка множественной загрузки
- Предпросмотр изображений
- Прогресс-бар загрузки

### 7.2 Хранение файлов
- Использовать существующий `afero.Fs` из Container
- Организация по workspace/channel
- Ограничение размера файлов
- Проверка MIME-типов

### 7.3 Отображение файлов
- Компоненты для разных типов файлов
- Встроенный просмотр изображений
- Скачивание файлов
- Галерея файлов канала

---

## Фаза 8: Потоки (Threads) (1 неделя)

### 8.1 Модель данных
- Уже есть поле `thread_id` в Message
- Добавить счетчик ответов в Message

### 8.2 UI для потоков
- Кнопка "Ответить" под сообщением
- Отображение количества ответов
- Развертывание/сворачивание потока
- Страница просмотра потока

### 8.3 API
- Endpoint для получения ответов
- Endpoint для отправки ответа

---

## Фаза 9: Реакции (1 неделя)

### 9.1 UI для реакций
- Кнопка добавления реакции
- Отображение реакций под сообщением
- Список пользователей, поставивших реакцию
- Быстрые реакции (emoji picker)

### 9.2 API
- Уже созданы endpoints в Фазе 4
- WebSocket события для реакций

---

## Фаза 10: Уведомления (1-2 недели)

### 10.1 Система уведомлений
**Создать `pkg/services/notifications.go`:**

- Хранение уведомлений в БД
- Типы уведомлений:
  - Упоминание в сообщении
  - Упоминание в канале
  - Приглашение в канал
  - Прямое сообщение
  - Реакция на сообщение

### 10.2 Модель данных
**Создать сущность `Notification`:**

- Поля: `user_id`, `type`, `title`, `content`, `link`, `read`, `created_at`
- Edges: `user` (User)

### 10.3 UI уведомлений
- Иконка уведомлений в header
- Dropdown со списком уведомлений
- Страница всех уведомлений
- Отметка о прочтении

### 10.4 Email уведомления
- Использовать существующий `MailClient`
- Настройки уведомлений пользователя
- Шаблоны email

---

## Фаза 11: Профили и настройки (1 неделя)

### 11.1 Профиль пользователя
- Страница профиля
- Редактирование профиля
- Загрузка аватара
- Статус и статус-сообщение

### 11.2 Настройки workspace
- Настройки workspace (владелец/админ)
- Управление участниками
- Роли и права доступа

### 11.3 Настройки канала
- Описание канала
- Управление участниками
- Архивирование канала

---

## Фаза 12: Оптимизация и производительность (1-2 недели)

### 12.1 Кэширование
- Кэширование списков каналов
- Кэширование участников
- Кэширование последних сообщений

### 12.2 Пагинация
- Использовать существующий `pkg/pager`
- Бесконечная прокрутка для сообщений
- Виртуализация списков (если нужно)

### 12.3 Оптимизация запросов
- Eager loading для связанных сущностей
- Индексы в БД
- Оптимизация WebSocket broadcast

### 12.4 Мониторинг
- Логирование WebSocket событий
- Метрики производительности
- Мониторинг очередей задач

---

## Фаза 13: Тестирование (1-2 недели)

### 13.1 Unit тесты
- Тесты для handlers
- Тесты для WebSocket hub
- Тесты для сервисов

### 13.2 Integration тесты
- Тесты API endpoints
- Тесты WebSocket соединений
- Тесты с реальной БД

### 13.3 E2E тесты
- Тесты основных сценариев
- Тесты отправки сообщений
- Тесты создания каналов

---

## Фаза 14: Дополнительные функции (опционально)

### 14.1 Голосовые/видео звонки
- Интеграция с WebRTC
- Или использование внешнего сервиса

### 14.2 Интеграции
- Webhooks
- API для ботов
- Интеграция с внешними сервисами

### 14.3 Мобильное приложение
- REST API для мобильных клиентов
- Push уведомления

### 14.4 Экспорт данных
- Экспорт сообщений канала
- Экспорт всех данных workspace

---

## Технические детали

### Зависимости для добавления

```go
// WebSocket (Echo уже поддерживает, но может понадобиться)
// gorilla/websocket - если нужна более продвинутая функциональность

// Для полнотекстового поиска
// github.com/blevesearch/bleve - если нужен более мощный поиск

// Для обработки изображений (опционально)
// github.com/disintegration/imaging
```

### Структура файлов

```
ent/schema/
  ├── workspace.go
  ├── channel.go
  ├── message.go
  ├── directmessage.go
  ├── reaction.go
  ├── attachment.go
  └── notification.go

pkg/
  ├── websocket/
  │   ├── hub.go
  │   ├── connection.go
  │   ├── messages.go
  │   └── events.go
  ├── handlers/
  │   └── messenger.go
  ├── services/
  │   └── notifications.go
  ├── ui/
  │   ├── components/
  │   │   └── messenger/
  │   ├── pages/
  │   │   └── messenger/
  │   └── forms/
  │       └── messenger/
  └── middleware/
      ├── workspace.go
      └── channel.go
```

### Конфигурация

Добавить в `config/config.yaml`:

```yaml
messenger:
  maxFileSize: 10485760  # 10MB
  allowedFileTypes: ["image/*", "application/pdf", "text/*"]
  messageLimit: 50  # сообщений за раз
  typingTimeout: "3s"
  onlineTimeout: "5m"
```

---

## Оценка времени

**Общее время разработки: 16-24 недели** (4-6 месяцев при работе 1 разработчика)

**Разбивка по фазам:**
- Фаза 1: 2-3 недели
- Фаза 2: 1-2 недели
- Фаза 3: 2 недели
- Фаза 4: 2-3 недели
- Фаза 5: 1-2 недели
- Фаза 6: 1 неделя
- Фаза 7: 1-2 недели
- Фаза 8: 1 неделя
- Фаза 9: 1 неделя
- Фаза 10: 1-2 недели
- Фаза 11: 1 неделя
- Фаза 12: 1-2 недели
- Фаза 13: 1-2 недели
- Фаза 14: опционально

---

## Приоритеты разработки

**MVP (Минимально жизнеспособный продукт):**
1. Фаза 1: Модели данных
2. Фаза 2: WebSocket базовая инфраструктура
3. Фаза 3: Базовый UI
4. Фаза 4: Основные handlers
5. Фаза 5: Реальное время (базовое)

**После MVP:**
- Остальные фазы по приоритету

---

## Полезные ресурсы

- [Echo WebSocket](https://echo.labstack.com/cookbook/websocket)
- [HTMX WebSocket](https://htmx.org/docs/#websockets)
- [Ent Documentation](https://entgo.io/docs/getting-started)
- [Gomponents Examples](https://www.gomponents.com/)

---

## Примечания

- Использовать существующие паттерны проекта
- Следовать структуре handlers, forms, pages
- Использовать существующие компоненты где возможно
- Тестировать каждую фазу перед переходом к следующей
- Документировать API endpoints

