# Чеклист разработки Slack-клона

## ✅ Фаза 1: Базовая инфраструктура и модели данных

### Модели данных (Ent)
- [ ] Workspace сущность
- [ ] Channel сущность
- [ ] Message сущность
- [ ] DirectMessage сущность
- [ ] DirectMessageContent сущность
- [ ] Reaction сущность
- [ ] Attachment сущность
- [ ] ChannelMember сущность
- [ ] WorkspaceMember сущность
- [ ] Расширение User сущности
- [ ] Генерация кода: `make ent-gen`
- [ ] Проверка миграций

## ✅ Фаза 2: WebSocket инфраструктура

- [ ] Создать `pkg/websocket/hub.go`
- [ ] Создать `pkg/websocket/connection.go`
- [ ] Создать `pkg/websocket/messages.go`
- [ ] Создать `pkg/websocket/events.go`
- [ ] Добавить WebSocket endpoint в роутер
- [ ] Middleware для аутентификации WebSocket
- [ ] Интеграция с HTMX WebSocket

## ✅ Фаза 3: Базовый UI и навигация

### Layout
- [ ] Создать `pkg/ui/layouts/messenger.go`
- [ ] Левая панель (workspaces, channels, DMs)
- [ ] Центральная панель (сообщения)
- [ ] Правая панель (опционально)

### Компоненты
- [ ] `pkg/ui/components/messenger/sidebar.go`
- [ ] `pkg/ui/components/messenger/channel_list.go`
- [ ] `pkg/ui/components/messenger/message_list.go`
- [ ] `pkg/ui/components/messenger/message_input.go`
- [ ] `pkg/ui/components/messenger/user_avatar.go`
- [ ] `pkg/ui/components/messenger/typing_indicator.go`
- [ ] `pkg/ui/components/messenger/reaction_picker.go`
- [ ] `pkg/ui/components/messenger/file_upload.go`

### Страницы
- [ ] `pkg/ui/pages/messenger/workspace.go`
- [ ] `pkg/ui/pages/messenger/channel.go`
- [ ] `pkg/ui/pages/messenger/direct_message.go`
- [ ] `pkg/ui/pages/messenger/channel_settings.go`

## ✅ Фаза 4: Обработчики и API

### Handlers
- [ ] `pkg/handlers/messenger.go` - основной handler
- [ ] Workspace endpoints (CRUD + members)
- [ ] Channel endpoints (CRUD + members + messages)
- [ ] Message endpoints (CRUD + replies)
- [ ] Direct Message endpoints
- [ ] Reaction endpoints
- [ ] File/Attachment endpoints
- [ ] WebSocket endpoint

### Forms
- [ ] `pkg/ui/forms/messenger/workspace.go`
- [ ] `pkg/ui/forms/messenger/channel.go`
- [ ] `pkg/ui/forms/messenger/message.go`
- [ ] `pkg/ui/forms/messenger/invite.go`

### Middleware
- [ ] `pkg/middleware/workspace.go`
- [ ] `pkg/middleware/channel.go`

### Route names
- [ ] Добавить все route names в `pkg/routenames/names.go`

## ✅ Фаза 5: Реальное время

- [ ] Реализовать WebSocket события от клиента
- [ ] Реализовать WebSocket события от сервера
- [ ] Интеграция отправки сообщений с WebSocket
- [ ] Отслеживание статуса пользователей
- [ ] Отметка о прочтении сообщений
- [ ] Индикатор набора текста
- [ ] Индикатор онлайн статуса
- [ ] Badge непрочитанных сообщений

## ✅ Фаза 6: Поиск и фильтрация

- [ ] Расширить поиск сообщений
- [ ] Поиск пользователей
- [ ] UI компонент поиска
- [ ] Страница результатов поиска
- [ ] Подсветка найденного текста

## ✅ Фаза 7: Файлы и вложения

- [ ] Расширить загрузку файлов
- [ ] Множественная загрузка
- [ ] Предпросмотр изображений
- [ ] Прогресс-бар загрузки
- [ ] Организация хранения файлов
- [ ] Компоненты отображения файлов
- [ ] Галерея файлов канала

## ✅ Фаза 8: Потоки (Threads)

- [ ] Добавить счетчик ответов в Message
- [ ] UI кнопка "Ответить"
- [ ] Отображение количества ответов
- [ ] Развертывание/сворачивание потока
- [ ] Страница просмотра потока
- [ ] API endpoints для потоков

## ✅ Фаза 9: Реакции

- [ ] UI кнопка добавления реакции
- [ ] Отображение реакций
- [ ] Список пользователей с реакциями
- [ ] Emoji picker
- [ ] WebSocket события для реакций

## ✅ Фаза 10: Уведомления

- [ ] Создать Notification сущность
- [ ] `pkg/services/notifications.go`
- [ ] Типы уведомлений
- [ ] UI иконка уведомлений
- [ ] Dropdown уведомлений
- [ ] Страница всех уведомлений
- [ ] Email уведомления
- [ ] Настройки уведомлений

## ✅ Фаза 11: Профили и настройки

- [ ] Страница профиля пользователя
- [ ] Редактирование профиля
- [ ] Загрузка аватара
- [ ] Статус и статус-сообщение
- [ ] Настройки workspace
- [ ] Управление участниками workspace
- [ ] Настройки канала

## ✅ Фаза 12: Оптимизация и производительность

- [ ] Кэширование списков каналов
- [ ] Кэширование участников
- [ ] Кэширование последних сообщений
- [ ] Пагинация сообщений
- [ ] Индексы в БД
- [ ] Оптимизация WebSocket broadcast
- [ ] Мониторинг и метрики

## ✅ Фаза 13: Тестирование

- [ ] Unit тесты для handlers
- [ ] Unit тесты для WebSocket
- [ ] Unit тесты для сервисов
- [ ] Integration тесты API
- [ ] Integration тесты WebSocket
- [ ] E2E тесты основных сценариев

## ✅ Фаза 14: Дополнительные функции (опционально)

- [ ] Голосовые/видео звонки
- [ ] Webhooks
- [ ] API для ботов
- [ ] Мобильное API
- [ ] Push уведомления
- [ ] Экспорт данных

## Конфигурация

- [ ] Добавить настройки messenger в `config/config.yaml`
- [ ] Обновить документацию

## Документация

- [ ] Документировать API endpoints
- [ ] Обновить README с инструкциями
- [ ] Примеры использования

