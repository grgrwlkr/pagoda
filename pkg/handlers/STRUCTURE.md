# Структура пакета handlers

Этот документ описывает организацию файлов в пакете `handlers`.

> **Важно**: В Go все файлы одного пакета должны находиться в одной директории. Поэтому все файлы находятся в корне `pkg/handlers/`, но организованы логически через именование и группировку по префиксам.

## 📁 Логическая организация файлов

Файлы сгруппированы по префиксам для легкой навигации:

### 🔐 Auth Handler (Аутентификация)
**Префикс**: `auth*`
- `auth.go` - обработчик аутентификации (логин, регистрация, восстановление пароля, верификация email)
- `auth_test.go` - тесты для auth handler

### 💬 Messenger Handler (Мессенджер)
**Префикс**: `messenger*`
- `messenger.go` - обработчик мессенджера (workspaces, channels, messages, direct messages, reactions, attachments)
- `messenger_test.go` - основные тесты для messenger handler
- `messenger_additional_test.go` - дополнительные тесты (update, delete, reactions, attachments)

### 🔍 Search Handler (Поиск)
**Префикс**: `search*`
- `search.go` - обработчик поиска (сообщения, пользователи)
- `search_test.go` - тесты для search handler

### 🔌 WebSocket Handler (WebSocket)
**Префикс**: `websocket*`
- `websocket.go` - обработчик WebSocket соединений
- `websocket_test.go` - тесты для websocket handler

### 🛠 Common (Общие файлы)
**Префиксы**: `handlers*`, `router*`, `error*`
- `handlers.go` - интерфейс Handler, функции Register/GetHandlers, fail()
- `handlers_test.go` - тесты для общих функций
- `router.go` - BuildRouter() - построение роутера и регистрация всех хендлеров
- `error.go` - обработчик ошибок

### 🧪 Test Helpers (Тестовые хелперы)
**Префиксы**: `test_*`, `*_helpers*`
- `test_setup.go` - глобальные переменные для тестов (srv, c)
- `test_setup_test.go` - TestMain для инициализации тестового окружения
- `auth_helpers_test.go` - хелперы для работы с пользователями и аутентификацией
  - `createTestUser()` - создание тестового пользователя
  - `authenticateUser()` - аутентификация и получение HTTP клиента
- `http_helpers_test.go` - хелперы для HTTP запросов
  - `request()` - строитель HTTP запросов
  - `httpRequest` - тип для построения запросов
  - `httpResponse` - тип для работы с ответами

### 🗑 Deprecated (Устаревшие/неиспользуемые)
**Префиксы**: `admin*`, `cache*`, `contact*`, `files*`, `task*`

Эти хендлеры помечены как "removed - this is now a Slack-only application" и не используются:
- `admin.go` - админ панель
- `cache.go` - кеширование
- `contact.go` - контакты
- `files.go` - файлы
- `task.go` - задачи

## 📊 Статистика

- **Активных хендлеров**: 4 (auth, messenger, search, websocket)
- **Общих файлов**: 4 (handlers, router, error, handlers_test)
- **Тестовых файлов**: 7
- **Тестовых хелперов**: 4 файла
- **Устаревших хендлеров**: 5 (admin, cache, contact, files, task)
- **Всего файлов**: 22

## 🎯 Принципы организации

### 1. Группировка по префиксам
Файлы сгруппированы по префиксам имени:
- `{handler_name}.go` - основной файл хендлера
- `{handler_name}_test.go` - тесты хендлера
- `{handler_name}_additional_test.go` - дополнительные тесты
- `test_*` - тестовые хелперы (setup)
- `*_helpers_test.go` - специализированные хелперы

### 2. Логическая группировка в IDE
При просмотре файлов в IDE они автоматически группируются по префиксам:
```
auth.go
auth_test.go
auth_helpers_test.go

messenger.go
messenger_test.go
messenger_additional_test.go

search.go
search_test.go

websocket.go
websocket_test.go

handlers.go
handlers_test.go
router.go
error.go

test_setup.go
test_setup_test.go
http_helpers_test.go
```

### 3. Именование
- Основные файлы: `{handler_name}.go`
- Тесты: `{handler_name}_test.go`
- Дополнительные тесты: `{handler_name}_additional_test.go`
- Хелперы: `test_{purpose}.go` или `{purpose}_helpers_test.go`

## 🔄 Добавление нового хендлера

1. Создать файл `{handler_name}.go` в корне `pkg/handlers/`
2. Реализовать интерфейс `Handler` (методы `Init` и `Routes`)
3. Зарегистрировать через `handlers.Register()` в функции `init()`
4. Создать тесты в `{handler_name}_test.go`
5. Использовать хелперы из `*_helpers_test.go` для тестов

## 📝 Пример структуры файлов

```
pkg/handlers/
├── handlers.go                    # Интерфейс Handler, Register, GetHandlers, fail
├── handlers_test.go               # Тесты для handlers.go
├── router.go                      # BuildRouter, регистрация всех хендлеров
├── error.go                       # Обработчик ошибок
│
├── auth.go                        # Auth handler
├── auth_test.go                   # Тесты для auth
│
├── messenger.go                   # Messenger handler
├── messenger_test.go              # Основные тесты для messenger
├── messenger_additional_test.go   # Дополнительные тесты
│
├── search.go                      # Search handler
├── search_test.go                 # Тесты для search
│
├── websocket.go                   # WebSocket handler
├── websocket_test.go              # Тесты для websocket
│
├── test_setup.go                  # Глобальные переменные для тестов
├── test_setup_test.go             # TestMain
├── auth_helpers_test.go           # Хелперы для аутентификации
├── http_helpers_test.go           # Хелперы для HTTP запросов
│
├── admin.go                       # Устаревший (не используется)
├── cache.go                       # Устаревший (не используется)
├── contact.go                     # Устаревший (не используется)
├── files.go                       # Устаревший (не используется)
└── task.go                        # Устаревший (не используется)
```

## ✅ Преимущества текущей структуры

1. **Go-идиоматичность**: Соответствует стандартам Go (один пакет = одна директория)
2. **Простота навигации**: Файлы логически сгруппированы по префиксам
3. **Тесты рядом с кодом**: Тесты находятся рядом с соответствующими хендлерами
4. **Переиспользование**: Хелперы доступны во всех тестах пакета
5. **Легко найти**: IDE автоматически группирует файлы по префиксам
6. **Чистота**: Все файлы в одном месте, но логически организованы

## 🔍 Поиск файлов

### По функциональности:
- **Аутентификация**: `auth*`
- **Мессенджер**: `messenger*`
- **Поиск**: `search*`
- **WebSocket**: `websocket*`
- **Общие**: `handlers*`, `router*`, `error*`
- **Тесты**: `*_test.go`, `test_*`
- **Хелперы**: `*_helpers*`

### В IDE:
Большинство IDE автоматически группируют файлы по префиксам, что делает навигацию удобной даже при большом количестве файлов.

## 📚 Дополнительная документация

- `TEST_HELPERS_README.md` - подробная документация по использованию тестовых хелперов
- `STRUCTURE.md` - этот файл, описание структуры пакета

## ⚠️ Почему не используются отдельные папки?

В Go файлы одного пакета должны находиться в одной директории. Создание отдельных папок для каждого хендлера потребовало бы:
- Создания отдельных пакетов (handlers/auth, handlers/messenger и т.д.)
- Экспортирования всех общих функций
- Обновления всех импортов
- Усложнения структуры без реальной пользы

Текущая организация через префиксы обеспечивает:
- ✅ Логическую группировку
- ✅ Простоту навигации
- ✅ Соответствие стандартам Go
- ✅ Легкость поддержки
