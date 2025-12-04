# Руководство по разработке

Этот документ описывает обязательные шаги и лучшие практики при разработке новых функций в проекте Pagoda.

## Содержание

- [Технологии](#технологии)
- [Структурные подходы](#структурные-подходы)
- [TDD - Test-Driven Development](#tdd---test-driven-development)
- [Документирование API](#документирование-api)
- [Стиль кода](#стиль-кода)
- [Процесс разработки](#процесс-разработки)

## Технологии

Проект построен на следующих технологиях:

### Backend

- **[Go](https://go.dev/)** - основной язык программирования
- **[Echo](https://echo.labstack.com/)** - высокопроизводительный, расширяемый, минималистичный веб-фреймворк для Go
- **[Ent](https://entgo.io/)** - простой, но мощный ORM для моделирования и запросов данных
- **[Gomponents](https://github.com/maragudk/gomponents)** - HTML компоненты, написанные на чистом Go

### Frontend

- **[HTMX](https://htmx.org/)** - доступ к AJAX, CSS Transitions, WebSockets и Server Sent Events напрямую в HTML
- **[Alpine.js](https://alpinejs.dev/)** - минималистичный инструмент для композиции поведения в разметке
- **[DaisyUI](https://daisyui.com/)** - компонентная библиотека для Tailwind CSS
- **[Tailwind CSS](https://tailwindcss.com/)** - утилитарный CSS фреймворк

### Storage

- **[SQLite](https://sqlite.org/)** - легковесная, быстрая, самодостаточная SQL база данных

### Дополнительные библиотеки

- **[Gorilla Sessions](https://github.com/gorilla/sessions)** - управление сессиями
- **[Validator](https://github.com/go-playground/validator)** - валидация данных
- **[Viper](https://github.com/spf13/viper)** - управление конфигурацией
- **[Testify](https://github.com/stretchr/testify)** - библиотека для тестирования
- **[Goquery](https://github.com/PuerkitoBio/goquery)** - парсинг HTML в тестах

## Структурные подходы

### Service Container

Все сервисы приложения находятся в контейнере (`pkg/services/container.go`). Контейнер обеспечивает:

- **Dependency Injection** - легкая инъекция зависимостей во все части приложения
- **Единая точка инициализации** - все сервисы создаются и инициализируются в одном месте
- **Graceful shutdown** - корректное завершение работы всех сервисов

Сервисы, доступные в контейнере:
- Authentication
- Cache
- Configuration
- Database
- Files
- Mail
- ORM
- Tasks
- Validator
- Web (Echo router)

### Handlers

Обработчики маршрутов (`pkg/handlers`) следуют единому паттерну:

1. **Определение типа handler:**
```go
type Example struct {
    orm *ent.Client
}
```

2. **Автоматическая регистрация:**
```go
func init() {
    Register(new(Example))
}
```

3. **Инициализация с зависимостями:**
```go
func (e *Example) Init(c *services.Container) error {
    e.orm = c.ORM
    return nil
}
```

4. **Регистрация маршрутов:**
```go
func (e *Example) Routes(g *echo.Group) {
    g.GET("/example", e.Page).Name = routenames.Example
    g.POST("/example", e.PageSubmit).Name = routenames.ExampleSubmit
}
```

**Важно:** Все маршруты должны иметь имена, определенные в пакете `routenames`. Это позволяет генерировать URL динамически и избегать хардкода путей.

### Forms

Формы (`pkg/ui/forms`) используют встроенную валидацию:

```go
type Guestbook struct {
    Message    string `form:"message" validate:"required"`
    form.Submission
}
```

Обработка формы:
```go
err := form.Submit(ctx, &input)
switch err.(type) {
case nil:
    // Успешная валидация
case validator.ValidationErrors:
    // Ошибки валидации - перерендерить форму
    return e.Page(ctx)
default:
    // Другая ошибка
    return err
}
```

### Pages и Components

- **Pages** (`pkg/ui/pages`) - полные страницы приложения
- **Components** (`pkg/ui/components`) - переиспользуемые UI компоненты
- **Layouts** (`pkg/ui/layouts`) - шаблоны макетов страниц

Все UI строится с помощью Gomponents, что обеспечивает типобезопасность и возможность переиспользования компонентов.

### Middleware

Middleware (`pkg/middleware`) используется для:
- Аутентификации (`RequireAuthentication`, `RequireNoAuthentication`)
- Проверки прав доступа (`RequireAdmin`)
- Загрузки данных в контекст (`LoadAuthenticatedUser`)
- Логирования (`SetLogger`, `LogRequest`)
- Кэширования (`CacheControl`)

## TDD - Test-Driven Development

**Мы пропагандируем подход TDD (Test-Driven Development).**

### Основные принципы

1. **Тесты пишутся ПЕРВЫМИ** - перед написанием кода функционала
2. **Тесты обязательны** - каждый новый функционал должен быть покрыт тестами
3. **Тесты запускаются после разработки** - перед коммитом необходимо убедиться, что все тесты проходят

### Процесс разработки с TDD

#### Шаг 1: Составление тест-плана

Перед началом разработки нового функционала необходимо:

1. **Определить границы функционала** - что именно должно быть реализовано
2. **Составить список всех тест-кейсов**, включая:
   - **Happy path** - успешные сценарии
   - **Edge cases** - граничные случаи
   - **Error cases** - случаи ошибок
   - **Corner cases** - неочевидные сценарии

3. **Пример тест-плана:**

```
Функционал: Создание workspace

Тест-кейсы:
1. Happy path:
   - Успешное создание workspace с валидными данными
   - Проверка сохранения в БД
   - Проверка создания связи с пользователем

2. Edge cases:
   - Создание workspace с максимальной длиной имени
   - Создание workspace с минимальной длиной имени
   - Создание workspace с граничными значениями slug

3. Error cases:
   - Попытка создать workspace без имени
   - Попытка создать workspace с невалидным slug
   - Попытка создать workspace с дублирующимся slug
   - Попытка создать workspace без аутентификации

4. Corner cases:
   - Создание workspace с специальными символами в имени
   - Создание workspace с пробелами в slug
   - Одновременное создание workspace несколькими пользователями
```

#### Шаг 2: Написание тестов

Напишите все тесты из тест-плана **до** реализации функционала:

```go
func TestWorkspaceCreate(t *testing.T) {
    // Arrange
    usr := createTestUser(t, "test@example.com", "Test User", "password123")
    defer cleanupUser(t, usr.ID)
    
    client := authenticateUser(t, "test@example.com", "password123")
    
    // Act
    resp, err := client.PostForm(createURL, formData)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusFound, resp.StatusCode)
    // Проверка создания в БД
    workspace, err := c.ORM.Workspace.Query()...
    assert.NoError(t, err)
    assert.Equal(t, "Test Workspace", workspace.Name)
}
```

#### Шаг 3: Запуск тестов (они должны упасть)

Запустите тесты - они должны упасть, так как функционал еще не реализован:

```bash
go test ./pkg/handlers/... -v
```

#### Шаг 4: Реализация функционала

Реализуйте минимально необходимый код, чтобы тесты прошли.

#### Шаг 5: Рефакторинг

После того, как тесты проходят, можно провести рефакторинг кода, сохраняя зеленые тесты.

#### Шаг 6: Финальная проверка

Перед коммитом обязательно:

```bash
# Запустить все тесты
go test ./...

# Или тесты конкретного пакета
go test ./pkg/handlers/... -v

# Проверить покрытие (опционально)
go test ./... -cover
```

### Структура тестов

Тесты должны следовать паттерну **Arrange-Act-Assert (AAA)**:

```go
func TestExample(t *testing.T) {
    // Arrange - подготовка данных
    ctx := c.Web.NewContext(nil, nil).Request().Context()
    usr := createTestUser(t, "test@example.com", "Test", "password")
    defer cleanup(t, usr.ID)
    
    // Act - выполнение действия
    resp, err := client.Get(url)
    
    // Assert - проверка результата
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Тестовые хелперы

Используйте существующие тестовые хелперы из `pkg/handlers/router_test.go`:

- `createTestUser()` - создание тестового пользователя
- `authenticateUser()` - аутентификация пользователя для тестов
- `request()` - создание HTTP запроса
- `httpRequest`, `httpResponse` - хелперы для HTTP запросов

### Тестирование HTTP handlers

Для тестирования handlers используется реальный HTTP сервер:

```go
func TestHandler_Get(t *testing.T) {
    doc := request(t).
        setRoute("route_name").
        get().
        assertStatusCode(http.StatusOK).
        toDoc()
    
    // Использование goquery для проверки HTML
    h1 := doc.Find("h1.title")
    assert.Len(t, h1.Nodes, 1)
    assert.Equal(t, "Expected Title", h1.Text())
}
```

### Тестирование форм

При тестировании форм необходимо:

1. Получить CSRF токен из GET запроса
2. Включить токен в POST запрос
3. Проверить валидацию и обработку ошибок

```go
// Получение CSRF токена
resp, err := client.Get(formURL)
doc, _ := goquery.NewDocumentFromReader(resp.Body)
csrf := doc.Find(`input[name="csrf"]`).First()
token, _ := csrf.Attr("value")

// POST запрос с токеном
body.Set("csrf", token)
body.Set("field", "value")
resp, err = client.PostForm(submitURL, body)
```

## Документирование API

### Публичное API должно быть документировано

Все публичные функции, методы, типы и константы должны иметь комментарии в формате Go doc.

### Формат документации

```go
// FunctionName выполняет описание функции.
// 
// Parameters:
//   - param1: описание параметра 1
//   - param2: описание параметра 2
//
// Returns:
//   - result: описание возвращаемого значения
//   - error: описание ошибки, если она может возникнуть
//
// Example:
//   result, err := FunctionName(ctx, "value1", "value2")
//   if err != nil {
//       return err
//   }
func FunctionName(ctx echo.Context, param1 string, param2 int) (result string, err error) {
    // реализация
}
```

### Что документировать

1. **Все публичные функции и методы** (с заглавной буквы)
2. **Типы и структуры** - описание назначения
3. **Поля структур**, если они являются частью публичного API
4. **Константы** - описание значения
5. **Пакеты** - описание в `package` комментарии

### Примеры хорошей документации

```go
// Auth обрабатывает HTTP запросы, связанные с аутентификацией пользователей.
// Поддерживает регистрацию, вход, выход, восстановление пароля и верификацию email.
type Auth struct {
    config *config.Config
    auth   *services.AuthClient
    mail   *services.MailClient
    orm    *ent.Client
}

// Init инициализирует handler с зависимостями из контейнера.
// Parameters:
//   - c: контейнер сервисов со всеми зависимостями
//
// Returns:
//   - error: ошибка инициализации, если какая-либо зависимость отсутствует
func (h *Auth) Init(c *services.Container) error {
    h.config = c.Config
    h.orm = c.ORM
    h.auth = c.Auth
    h.mail = c.Mail
    return nil
}

// LoginSubmit обрабатывает отправку формы входа.
// Проверяет учетные данные пользователя и создает сессию при успешной аутентификации.
// Parameters:
//   - ctx: Echo context с данными формы (email, password)
//
// Returns:
//   - error: ошибка аутентификации или ошибка рендеринга страницы
func (h *Auth) LoginSubmit(ctx echo.Context) error {
    // реализация
}
```

### Что НЕ документировать

- Приватные функции (с маленькой буквы) - только если логика сложная и требует пояснения
- Очевидные геттеры/сеттеры без дополнительной логики
- Внутренние вспомогательные функции

## Стиль кода

### Общие принципы

1. **Следуйте стандартам Go** - используйте `gofmt` и `golint`
2. **Именование** - следуйте Go conventions:
   - Публичные идентификаторы - с заглавной буквы
   - Приватные - с маленькой
   - Короткие имена для локальных переменных, длинные для глобальных
3. **Длина строк** - рекомендуется не более 100-120 символов
4. **Комментарии** - пишите на русском языке для внутренней документации

### Структура файлов

```go
package handlers

import (
    // Стандартная библиотека
    "fmt"
    "net/http"
    
    // Сторонние библиотеки
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    
    // Внутренние пакеты
    "github.com/mikestefanello/pagoda/pkg/services"
)
```

### Обработка ошибок

Всегда обрабатывайте ошибки явно:

```go
// Плохо
user, _ := c.ORM.User.Create().Save(ctx)

// Хорошо
user, err := c.ORM.User.Create().Save(ctx)
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

### Возврат ошибок из handlers

Используйте `echo.HTTPError` для HTTP ошибок:

```go
if err != nil {
    return echo.NewHTTPError(http.StatusInternalServerError, "описание ошибки")
}
```

### Контекст

Всегда передавайте контекст из запроса:

```go
ctx := c.Web.NewContext(nil, nil).Request().Context()
user, err := c.ORM.User.Query().Where(...).Only(ctx)
```

### Валидация

Используйте struct tags для валидации:

```go
type CreateWorkspace struct {
    Name string `form:"name" validate:"required,min=3,max=100"`
    Slug string `form:"slug" validate:"required,alphanum,min=3,max=50"`
    form.Submission
}
```

### Логирование

Используйте структурированное логирование:

```go
log.Ctx(ctx).Info("workspace created",
    "workspace_id", workspace.ID,
    "user_id", user.ID,
)
```

## Процесс разработки

### Обязательные шаги при разработке новой функции

1. **Планирование**
   - Определить требования к функционалу
   - Составить тест-план со всеми кейсами (happy path, edge cases, error cases, corner cases)
   - Определить необходимые изменения в БД (если нужны)

2. **Написание тестов (TDD)**
   - Написать все тесты из тест-плана
   - Убедиться, что тесты падают (функционал еще не реализован)

3. **Реализация**
   - Реализовать минимально необходимый функционал
   - Следовать структурным подходам проекта (handlers, services, forms, pages)
   - Использовать dependency injection через Container

4. **Документирование**
   - Добавить Go doc комментарии ко всем публичным функциям и типам
   - Описать параметры и возвращаемые значения
   - Добавить примеры использования для сложных функций

5. **Тестирование**
   - Запустить все тесты: `go test ./...`
   - Убедиться, что все тесты проходят
   - Проверить покрытие кода (опционально)

6. **Рефакторинг**
   - Улучшить читаемость кода
   - Удалить дублирование
   - Оптимизировать производительность (если необходимо)
   - Убедиться, что тесты все еще проходят

7. **Проверка перед коммитом**
   - Запустить `go fmt ./...`
   - Запустить `go vet ./...`
   - Запустить все тесты
   - Проверить, что нет неиспользуемых импортов

### Чек-лист перед коммитом

- [ ] Все тесты написаны и проходят
- [ ] Публичное API документировано
- [ ] Код следует стилю проекта
- [ ] Использован `gofmt`
- [ ] Использован `go vet`
- [ ] Нет неиспользуемых импортов
- [ ] Ошибки обрабатываются явно
- [ ] Логирование добавлено где необходимо
- [ ] Тест-план выполнен полностью

### Работа с БД

При добавлении новых сущностей:

1. Создать schema в `ent/schema/`
2. Запустить `make ent-gen` для генерации кода
3. Миграции выполняются автоматически при старте приложения
4. Для тестов используется отдельная тестовая БД (in-memory SQLite по умолчанию)

### Работа с формами

1. Определить форму в `pkg/ui/forms/`
2. Добавить валидацию через struct tags
3. Создать метод `Render()` для формы
4. Обработать submission в POST handler

### Работа с UI

1. Создать page в `pkg/ui/pages/`
2. Использовать существующие components из `pkg/ui/components/`
3. Выбрать подходящий layout
4. Использовать Gomponents для построения HTML

## Дополнительные ресурсы

- [README.md](./README.md) - полная документация проекта
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - стандарты кода Go
- [Effective Go](https://go.dev/doc/effective_go) - лучшие практики Go
- [Echo Documentation](https://echo.labstack.com/guide/) - документация Echo
- [Ent Documentation](https://entgo.io/docs/getting-started) - документация Ent ORM

---

**Помните:** Качество кода и тестов важнее скорости разработки. Потратьте время на правильную архитектуру и тестирование - это окупится в будущем.
