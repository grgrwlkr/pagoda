# 🚀 Быстрый старт Slack-клона

## Запуск приложения

### Для разработки Slack-приложения:

```bash
# Запустить с автоперезагрузкой
make watch-slack
```

### Для разработки Pagoda (оригинального приложения):

```bash
# Запустить с автоперезагрузкой
make watch
```

## Переключение между приложениями

Проект поддерживает два режима работы:

1. **Pagoda** - оригинальное веб-приложение (по умолчанию)
2. **Slack** - Slack-клон мессенджер

### Команды для Slack:

```bash
make run-slack      # Запустить Slack приложение
make watch-slack    # Запустить Slack с автоперезагрузкой
make build-slack    # Собрать Slack приложение
```

### Команды для Pagoda:

```bash
make run            # Запустить Pagoda приложение
make watch          # Запустить Pagoda с автоперезагрузкой
make build          # Собрать Pagoda приложение
```

## Как это работает

- **Точка входа Pagoda**: `cmd/web/main.go` → устанавливает `PAGODA_APP_MODE=pagoda`
- **Точка входа Slack**: `cmd/slack/main.go` → устанавливает `PAGODA_APP_MODE=slack`

Handlers регистрируются автоматически в зависимости от режима:
- **Pagoda handlers**: Pages, Auth, Admin, Search, Contact, Files, Cache, Task
- **Slack handlers**: Messenger, WebSocket

Подробнее см. [APP_MODE_SWITCHING.md](./APP_MODE_SWITCHING.md)

---

**Приложение будет доступно на**: `http://localhost:8000`

