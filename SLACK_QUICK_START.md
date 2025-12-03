# 🚀 Быстрый старт Slack-клона

Это приложение представляет собой Slack-клон мессенджер, построенный на базе Pagoda Go starter kit.

## Запуск приложения

### Для разработки:

```bash
# Запустить с автоперезагрузкой
make watch
```

### Для продакшена:

```bash
# Собрать приложение
make build

# Запустить собранное приложение
./tmp/main
```

### Примечание о CSS

Если вы видите предупреждение `Warning: tailwindcss not found, skipping CSS build`, это нормально. CSS файл уже собран и находится в `public/static/main.css`. Если нужно пересобрать CSS, установите tailwindcss:

```bash
make tailwind-install
make css
```

## Структура приложения

- **Точка входа**: `cmd/web/main.go`
- **Handlers**: 
  - `pkg/handlers/messenger.go` - основной handler для мессенджера
  - `pkg/handlers/websocket.go` - WebSocket handler для real-time коммуникации
  - `pkg/handlers/auth.go` - аутентификация

## Функциональность

- ✅ Workspaces (рабочие пространства)
- ✅ Channels (каналы)
- ✅ Direct Messages (личные сообщения)
- ✅ Real-time сообщения через WebSocket
- ✅ Реакции на сообщения
- ✅ Потоки сообщений (threads)
- ✅ Вложения файлов
- ✅ Аутентификация и регистрация

## Разработка

Приложение использует:
- **Go** с Echo framework
- **Ent ORM** для работы с базой данных
- **Gomponents** для UI компонентов
- **HTMX** для динамических обновлений
- **Alpine.js** для клиентской логики
- **Tailwind CSS + DaisyUI** для стилизации
- **WebSocket** для real-time коммуникации
