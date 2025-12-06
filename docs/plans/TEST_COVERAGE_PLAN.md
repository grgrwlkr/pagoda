# План покрытия тестами проекта Pagoda

## Обзор

Этот документ описывает план тестов для всех handlers в проекте и текущее состояние покрытия тестами.

## Структура тестов

Все тесты следуют принципам TDD из `DEVELOPMENT_GUIDE.md`:
- **Arrange-Act-Assert (AAA)** паттерн
- Тестирование happy path, edge cases, error cases, corner cases
- Использование тестовых хелперов из `router_test.go`

## Текущее покрытие тестами

### ✅ Auth Handler (`pkg/handlers/auth_test.go`)

**Реализованные тесты:**
1. ✅ `TestAuthLoginPage` - отображение страницы входа
2. ✅ `TestAuthLoginSubmit_Success` - успешный вход
3. ✅ `TestAuthLoginSubmit_InvalidCredentials` - вход с неверными учетными данными
4. ✅ `TestAuthLoginSubmit_UserNotFound` - вход несуществующего пользователя
5. ✅ `TestAuthLoginSubmit_ValidationErrors` - ошибки валидации при входе
6. ✅ `TestAuthRegisterPage` - отображение страницы регистрации
7. ✅ `TestAuthRegisterSubmit_Success` - успешная регистрация
8. ✅ `TestAuthRegisterSubmit_DuplicateEmail` - регистрация с дублирующимся email
9. ✅ `TestAuthRegisterSubmit_ValidationErrors` - ошибки валидации при регистрации
10. ✅ `TestAuthLogout` - выход из системы
11. ✅ `TestAuthForgotPasswordPage` - страница восстановления пароля
12. ✅ `TestAuthForgotPasswordSubmit_Success` - успешная отправка запроса на восстановление
13. ✅ `TestAuthForgotPasswordSubmit_UserNotFound` - запрос для несуществующего пользователя
14. ✅ `TestAuthForgotPasswordSubmit_ValidationErrors` - ошибки валидации
15. ✅ `TestAuthVerifyEmail_InvalidToken` - верификация с неверным токеном

**Не реализованные тесты (требуют дополнительной настройки):**
- `TestAuthResetPasswordPage` - требует валидный токен сброса пароля
- `TestAuthResetPasswordSubmit_Success` - требует валидный токен
- `TestAuthResetPasswordSubmit_InvalidToken` - тест с невалидным токеном
- `TestAuthVerifyEmail_Success` - требует валидный токен верификации

### ✅ Messenger Handler (`pkg/handlers/messenger_test.go`)

**Реализованные тесты:**
1. ✅ `TestMessengerRootRedirect` - редирект на главную
2. ✅ `TestMessengerWorkspaceList` - список workspace
3. ✅ `TestMessengerWorkspaceCreatePage` - форма создания workspace
4. ✅ `TestMessengerWorkspaceCreate` - создание workspace
5. ✅ `TestMessengerWorkspaceView` - просмотр workspace
6. ✅ `TestMessengerChannelList` - список каналов
7. ✅ `TestMessengerChannelCreateForm` - форма создания канала
8. ✅ `TestMessengerWorkspaceCreate_ValidationErrors` - ошибки валидации при создании workspace
9. ✅ `TestMessengerWorkspaceCreate_DuplicateSlug` - дублирующийся slug
10. ✅ `TestMessengerWorkspaceView_Unauthorized` - просмотр без доступа
11. ✅ `TestMessengerChannelCreate` - создание канала
12. ✅ `TestMessengerChannelView` - просмотр канала
13. ✅ `TestMessengerChannelView_Unauthorized` - просмотр без доступа
14. ✅ `TestMessengerChannelMessages` - получение сообщений канала
15. ✅ `TestMessengerMessageCreate` - создание сообщения
16. ✅ `TestMessengerDirectMessageList` - список прямых сообщений

**Не реализованные тесты (требуют дополнительной работы):**
- `TestMessengerWorkspaceUpdate` - обновление workspace
- `TestMessengerWorkspaceDelete` - удаление workspace
- `TestMessengerWorkspaceAddMember` - добавление участника
- `TestMessengerWorkspaceRemoveMember` - удаление участника
- `TestMessengerChannelUpdate` - обновление канала
- `TestMessengerChannelDelete` - удаление канала
- `TestMessengerChannelAddMember` - добавление участника в канал
- `TestMessengerChannelRemoveMember` - удаление участника из канала
- `TestMessengerMessageUpdate` - обновление сообщения
- `TestMessengerMessageDelete` - удаление сообщения
- `TestMessengerMessageReplies` - получение ответов на сообщение
- `TestMessengerMessageReply` - создание ответа
- `TestMessengerMessageThread` - просмотр треда
- `TestMessengerDirectMessageView` - просмотр прямого сообщения
- `TestMessengerDirectMessageCreate` - создание прямого сообщения
- `TestMessengerDirectMessageMessages` - получение сообщений DM
- `TestMessengerDirectMessageMessageCreate` - создание сообщения в DM
- `TestMessengerReactionAdd` - добавление реакции
- `TestMessengerReactionRemove` - удаление реакции
- `TestMessengerAttachmentUpload` - загрузка вложения
- `TestMessengerDirectMessageAttachmentUpload` - загрузка вложения в DM
- `TestMessengerAttachmentView` - просмотр вложения
- `TestMessengerAttachmentDelete` - удаление вложения

### ✅ Search Handler (`pkg/handlers/search_test.go`)

**Реализованные тесты:**
1. ✅ `TestSearchPage` - страница поиска
2. ✅ `TestSearchPage_Unauthorized` - поиск без доступа к workspace
3. ✅ `TestSearchMessages` - поиск сообщений
4. ✅ `TestSearchMessages_NoQuery` - поиск без параметра запроса
5. ✅ `TestSearchMessages_Unauthorized` - поиск без доступа
6. ✅ `TestSearchUsers` - поиск пользователей
7. ✅ `TestSearchUsers_NoQuery` - поиск без параметра запроса
8. ✅ `TestSearchUsers_NoWorkspaceID` - поиск без workspace_id
9. ✅ `TestSearchUsers_Unauthorized` - поиск без доступа

**Дополнительные тесты (опционально):**
- Тесты с фильтрами по датам
- Тесты с фильтрами по каналам
- Тесты пагинации

### ✅ WebSocket Handler (`pkg/handlers/websocket_test.go`)

**Реализованные тесты:**
1. ✅ `TestWebSocket_Unauthenticated` - подключение без аутентификации
2. ✅ `TestWebSocket_Authenticated` - подключение с аутентификацией
3. ✅ `TestWebSocket_OriginCheck` - проверка Origin заголовка

**Дополнительные тесты (требуют WebSocket клиента):**
- Тесты отправки сообщений через WebSocket
- Тесты получения сообщений через WebSocket
- Тесты обработки ошибок соединения
- Тесты переподключения

## Статистика покрытия

### По handlers:
- **Auth**: 15/19 тестов (79%)
- **Messenger**: 16/40+ тестов (40%)
- **Search**: 9/9 основных тестов (100%)
- **WebSocket**: 3/3 базовых тестов (100%)

### По типам тестов:
- **Happy path**: ✅ Покрыто для основных сценариев
- **Error cases**: ✅ Покрыто для большинства handlers
- **Edge cases**: ✅ Частично покрыто
- **Corner cases**: ⚠️ Требует дополнительной работы

## Рекомендации по дальнейшему развитию

1. **Приоритет 1**: Добавить тесты для всех CRUD операций Messenger handler
2. **Приоритет 2**: Добавить тесты для ResetPassword и VerifyEmail с валидными токенами
3. **Приоритет 3**: Добавить интеграционные тесты для WebSocket с реальными сообщениями
4. **Приоритет 4**: Добавить тесты производительности для поиска

## Соответствие DEVELOPMENT_GUIDE.md

Все написанные тесты соответствуют требованиям из `DEVELOPMENT_GUIDE.md`:
- ✅ Используют паттерн Arrange-Act-Assert
- ✅ Используют тестовые хелперы (`createTestUser`, `authenticateUser`, `request`)
- ✅ Тестируют happy path, edge cases, error cases
- ✅ Проверяют HTTP статус коды
- ✅ Проверяют содержимое HTML через goquery
- ✅ Правильно очищают тестовые данные (defer cleanup)

## Запуск тестов

```bash
# Все тесты handlers
go test ./pkg/handlers/... -v

# Конкретный handler
go test ./pkg/handlers/... -run TestAuth

# С покрытием
go test ./pkg/handlers/... -cover
```
