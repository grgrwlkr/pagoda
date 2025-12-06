# Test Helpers Documentation

Этот документ описывает структуру тестовых хелперов в пакете `handlers`.

## Структура файлов

Тестовые хелперы разделены на отдельные файлы по функциональности:

### 1. `test_setup.go` и `test_setup_test.go`
**Назначение**: Инициализация тестового окружения

- **`test_setup.go`**: Содержит глобальные переменные `srv` (HTTP сервер) и `c` (контейнер сервисов)
- **`test_setup_test.go`**: Содержит `TestMain`, который инициализирует тестовое окружение перед запуском всех тестов

**Использование**: Не требует прямого использования, автоматически выполняется при запуске тестов.

### 2. `auth_helpers_test.go`
**Назначение**: Хелперы для работы с аутентификацией и пользователями

#### Функции:

- **`createTestUser(t *testing.T, email, name, password string) *ent.User`**
  - Создает тестового пользователя в базе данных
  - Возвращает созданного пользователя
  - **Важно**: Вызывающий код должен самостоятельно очистить пользователя после теста

  ```go
  usr := createTestUser(t, "test@example.com", "Test User", "password123")
  defer func() {
      ctx := context.Background()
      c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
  }()
  ```

- **`authenticateUser(t *testing.T, email, password string) *http.Client`**
  - Аутентифицирует пользователя и возвращает HTTP клиент с установленными сессионными cookies
  - Автоматически обрабатывает получение CSRF токена и отправку формы логина
  - Возвращает готовый к использованию аутентифицированный клиент

  ```go
  client := authenticateUser(t, "test@example.com", "password123")
  resp, err := client.Get(someURL)
  ```

### 3. `http_helpers_test.go`
**Назначение**: Хелперы для выполнения HTTP запросов в тестах

#### Типы:

- **`httpRequest`**: Строитель HTTP запросов с fluent API
- **`httpResponse`**: Обертка над HTTP ответом с удобными методами для проверок

#### Функции и методы:

- **`request(t *testing.T) *httpRequest`**
  - Создает новый строитель HTTP запросов с чистым cookie jar
  - Точка входа для создания HTTP запросов в тестах

- **`httpRequest.setRoute(route string, params ...any) *httpRequest`**
  - Устанавливает маршрут для запроса по имени маршрута и опциональным параметрам
  - Автоматически добавляет URL тестового сервера

- **`httpRequest.setBody(body url.Values) *httpRequest`**
  - Устанавливает тело формы для POST/PUT запросов

- **`httpRequest.get() *httpResponse`**
  - Выполняет GET запрос и возвращает ответ

- **`httpRequest.post() *httpResponse`**
  - Выполняет POST запрос с автоматической обработкой CSRF токена
  - Сначала делает GET запрос на тот же маршрут для получения CSRF токена
  - Затем включает токен в тело POST запроса

- **`httpRequest.postAuthenticated(client *http.Client) *httpResponse`**
  - Выполняет POST запрос используя аутентифицированный HTTP клиент
  - Полезно для выполнения аутентифицированных POST запросов
  - Автоматически обрабатывает CSRF токен

- **`httpResponse.assertStatusCode(code int) *httpResponse`**
  - Проверяет, что ответ имеет ожидаемый статус код
  - Возвращает себя для цепочки вызовов

- **`httpResponse.assertRedirect(t *testing.T, route string, params ...any) *httpResponse`**
  - Проверяет, что ответ является редиректом на ожидаемый маршрут
  - Возвращает себя для цепочки вызовов

- **`httpResponse.toDoc() *goquery.Document`**
  - Парсит тело ответа как HTML и возвращает goquery документ
  - Автоматически закрывает тело ответа после парсинга

## Примеры использования

### Пример 1: Простой GET запрос

```go
func TestSomething(t *testing.T) {
    resp := request(t).
        setRoute(routenames.SomeRoute).
        get()
    
    resp.assertStatusCode(http.StatusOK)
    doc := resp.toDoc()
    // Работа с документом...
}
```

### Пример 2: POST запрос с автоматическим CSRF

```go
func TestCreateSomething(t *testing.T) {
    body := url.Values{}
    body.Set("name", "Test")
    
    resp := request(t).
        setRoute(routenames.CreateSomething).
        setBody(body).
        post()
    
    resp.assertStatusCode(http.StatusOK)
}
```

### Пример 3: Аутентифицированный POST запрос

```go
func TestAuthenticatedAction(t *testing.T) {
    usr := createTestUser(t, "test@example.com", "Test", "password123")
    defer func() {
        ctx := context.Background()
        c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
    }()
    
    client := authenticateUser(t, "test@example.com", "password123")
    
    body := url.Values{}
    body.Set("data", "value")
    
    resp := request(t).
        setRoute(routenames.SomeAction).
        setBody(body).
        postAuthenticated(client)
    
    resp.assertStatusCode(http.StatusOK)
}
```

### Пример 4: Создание и использование тестового пользователя

```go
func TestUserAction(t *testing.T) {
    ctx := context.Background()
    
    // Создаем пользователя
    usr := createTestUser(t, "user@example.com", "User", "password123")
    
    // Cleanup
    defer func() {
        // Удаляем связанные данные
        // ...
        // Удаляем пользователя
        c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
    }()
    
    // Аутентифицируем пользователя
    client := authenticateUser(t, "user@example.com", "password123")
    
    // Используем клиент для запросов
    resp, err := client.Get(someURL)
    // ...
}
```

## Преимущества новой структуры

1. **Разделение ответственности**: Каждый файл отвечает за свою область
2. **Переиспользование**: Хелперы можно использовать во всех тестовых файлах
3. **Читаемость**: Тестовые файлы содержат только тесты, без вспомогательного кода
4. **Документация**: Каждая функция имеет подробные комментарии
5. **Унификация**: Единый подход к написанию тестов во всем пакете

## Миграция существующих тестов

Все существующие тесты уже используют эти хелперы. Если вы добавляете новые тесты:

1. Используйте `createTestUser` для создания пользователей
2. Используйте `authenticateUser` для получения аутентифицированного клиента
3. Используйте `request(t)` для создания HTTP запросов
4. Используйте методы `httpResponse` для проверок

## Примечания

- Все хелперы автоматически используют глобальные переменные `srv` и `c`
- CSRF токены обрабатываются автоматически в методах `post()` и `postAuthenticated()`
- Не забывайте очищать тестовые данные в `defer` блоках
- Порядок удаления данных важен: сначала зависимые записи, затем основные
