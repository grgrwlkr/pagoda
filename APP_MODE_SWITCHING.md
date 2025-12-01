# Переключение между Pagoda и Slack приложением

## Обзор

Проект поддерживает два режима работы:
- **Pagoda** - оригинальное веб-приложение Pagoda (по умолчанию)
- **Slack** - Slack-клон мессенджер

Переключение между режимами происходит через переменную окружения `PAGODA_APP_MODE` или через отдельные точки входа.

## Точки входа

### Pagoda (оригинальное приложение)
- **Точка входа**: `cmd/web/main.go`
- **Режим**: `pagoda` (по умолчанию)
- **Handlers**: Все Pagoda handlers (Pages, Auth, Admin, Search, Contact, Files, Cache, Task)

### Slack (мессенджер)
- **Точка входа**: `cmd/slack/main.go`
- **Режим**: `slack`
- **Handlers**: Только Slack handlers (Messenger, WebSocket)

## Команды Makefile

### Pagoda приложение

```bash
# Запустить Pagoda приложение
make run

# Запустить Pagoda с автоперезагрузкой (использует .air.toml)
make watch

# Собрать Pagoda приложение
make build
```

### Slack приложение

```bash
# Запустить Slack приложение
make run-slack

# Запустить Slack с автоперезагрузкой (использует .air.slack.toml)
make watch-slack

# Собрать Slack приложение
make build-slack
```

### Справка по командам

```bash
# Показать все доступные команды
make help
```

## Переключение через переменную окружения

Вы также можете использовать переменную окружения для переключения режима:

```bash
# Запустить Pagoda
PAGODA_APP_MODE=pagoda go run cmd/web/main.go

# Запустить Slack
PAGODA_APP_MODE=slack go run cmd/web/main.go
```

## Как это работает

### Регистрация handlers

Handlers регистрируются автоматически через `init()` функции, но регистрируются в router только если они соответствуют текущему режиму приложения.

### Интерфейс ModeHandler

Handlers могут реализовать интерфейс `ModeHandler` для указания, в каком режиме они должны быть зарегистрированы:

```go
type ModeHandler interface {
    Handler
    ShouldRegister(appMode string) bool
}
```

### Примеры

**Pagoda handler** (регистрируется только в режиме "pagoda"):
```go
func (h *Pages) ShouldRegister(appMode string) bool {
    return appMode == "pagoda"
}
```

**Slack handler** (регистрируется только в режиме "slack"):
```go
func (h *Messenger) ShouldRegister(appMode string) bool {
    return appMode == "slack"
}
```

### Handlers без ShouldRegister

Handlers, которые не реализуют `ModeHandler`, регистрируются всегда (для обратной совместимости).

## Конфигурация Air

Для каждого режима есть отдельный конфигурационный файл:
- `.air.toml` - для Pagoda (использует `make build`)
- `.air.slack.toml` - для Slack (использует `make build-slack`)

## Структура файлов

```
cmd/
├── web/
│   └── main.go          # Точка входа для Pagoda
└── slack/
    └── main.go          # Точка входа для Slack

app/handlers/
├── messenger.go         # Slack handler (ShouldRegister -> slack)
└── websocket.go         # Slack handler (ShouldRegister -> slack)

pkg/handlers/
├── pages.go             # Pagoda handler (ShouldRegister -> pagoda)
├── auth.go              # Pagoda handler (ShouldRegister -> pagoda)
├── admin.go             # Pagoda handler (ShouldRegister -> pagoda)
└── ...                  # Другие Pagoda handlers
```

## Рекомендации

1. **Для разработки Slack**: используйте `make watch-slack`
2. **Для разработки Pagoda**: используйте `make watch`
3. **Для продакшена**: используйте соответствующий `make build` или `make build-slack`

## Примечания

- Оба приложения используют одну и ту же базу данных
- Оба приложения используют одни и те же middleware и сервисы
- Различие только в регистрации handlers и точках входа
- Error handler регистрируется всегда (не имеет ShouldRegister)
- В Slack приложении корневой путь "/" перенаправляет на первый workspace пользователя

## Быстрое переключение

Самый простой способ переключиться между приложениями:

```bash
# Для Slack
make watch-slack

# Для Pagoda  
make watch
```

Оба команды используют air для автоматической перезагрузки при изменениях кода.

---

**Дата создания**: 2024-12-01

