# Организация пакета handlers

## 📁 Структура файлов

Все файлы находятся в корне `pkg/handlers/`, так как в Go файлы одного пакета должны быть в одной директории.

### Группировка по префиксам

Файлы организованы логически через именование с префиксами:

#### 🔐 Auth (Аутентификация)
- `auth.go` - основной хендлер
- `auth_test.go` - тесты
- `auth_helpers_test.go` - тестовые хелперы для auth

#### 💬 Messenger (Мессенджер)
- `messenger.go` - основной хендлер
- `messenger_test.go` - основные тесты
- `messenger_additional_test.go` - дополнительные тесты

#### 🔍 Search (Поиск)
- `search.go` - основной хендлер
- `search_test.go` - тесты

#### 🔌 WebSocket
- `websocket.go` - основной хендлер
- `websocket_test.go` - тесты

#### 🛠 Common (Общие)
- `handlers.go` - интерфейс Handler, Register, GetHandlers, fail
- `handlers_test.go` - тесты
- `router.go` - BuildRouter
- `error.go` - обработчик ошибок

#### 🧪 Test Helpers (Тестовые хелперы)
- `test_setup.go` - глобальные переменные
- `test_setup_test.go` - TestMain
- `http_helpers_test.go` - HTTP хелперы

#### 🗑 Deprecated (Устаревшие)
- `admin.go`, `cache.go`, `contact.go`, `files.go`, `task.go`

## 🎯 Почему не используются отдельные папки?

В Go файлы одного пакета **должны** находиться в одной директории. Создание отдельных папок потребовало бы:
- Отдельных пакетов для каждого хендлера
- Экспортирования всех общих функций
- Обновления всех импортов
- Усложнения без реальной пользы

Текущая организация через префиксы обеспечивает:
- ✅ Логическую группировку (файлы группируются по префиксам в IDE)
- ✅ Простоту навигации
- ✅ Соответствие стандартам Go
- ✅ Легкость поддержки

## 📝 Навигация в IDE

Большинство IDE автоматически группируют файлы по префиксам, что делает навигацию удобной:

```
📁 handlers/
  📁 auth*
    📄 auth.go
    📄 auth_test.go
    📄 auth_helpers_test.go
  📁 messenger*
    📄 messenger.go
    📄 messenger_test.go
    📄 messenger_additional_test.go
  📁 search*
    📄 search.go
    📄 search_test.go
  📁 websocket*
    📄 websocket.go
    📄 websocket_test.go
  📁 handlers*, router*, error*
    📄 handlers.go
    📄 router.go
    📄 error.go
  📁 test_*, *_helpers*
    📄 test_setup.go
    📄 http_helpers_test.go
```

## ✅ Итог

Структура организована логически через именование файлов, что соответствует стандартам Go и обеспечивает удобную навигацию.

