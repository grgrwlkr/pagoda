# Нарушения правил разработки (Development Guide Violations)

Этот файл содержит список всех найденных нарушений правил из `DEVELOPMENT_GUIDE.md`, которые требуют исправления.

## Критические нарушения: Использование чистого JavaScript

### 1. `pkg/ui/components/htmx.go`
**Проблема:** Использование чистого JavaScript для обработки HTMX событий
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 12-48

**Текущий код:**
```go
const htmxErr = `
    document.body.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.xhr.status >= 400){
            evt.detail.shouldSwap = true;
            evt.detail.target = htmx.find("body");
        }
    });
`
```

**Рекомендация:** 
- Для обработки ошибок использовать HTMX атрибуты `hx-on::response-error` на формах
- Для CSRF токена использовать Go компонент `components.CSRFInput(r)` вместо JavaScript
- Для модальных окон использовать Alpine.js или HTMX атрибуты

---

### 2. `pkg/ui/components/messenger/modals.go`
**Проблема:** Использование `onclick` и `document.getElementById()` для работы с модальными окнами
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 65, 83, 181, 202, 241, 259, 365, 384

**Текущий код:**
```go
Attr("onclick", "document.getElementById('channel-create-modal').close()")
Attr("hx-on::after-request", "if(event.detail.xhr.status === 200) { const modal = document.getElementById('channel-create-modal'); if(modal) modal.close(); }")
```

**Рекомендация:**
- Использовать Alpine.js для управления состоянием модальных окон
- Использовать HTMX атрибуты `hx-on::after-request` с Alpine.js директивами
- Пример: `Attr("x-data", "{ open: false }")` и `Attr("@click", "open = false")`

---

### 3. `pkg/ui/pages/messenger/channel.go`
**Проблема:** Обширное использование чистого JavaScript для WebSocket, DOM манипуляций, fetch()
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 53-54, 101, 108, 126, 154, 170, 186, 226-262, 275-299, 320-357, 374-525

**Основные проблемы:**
1. **WebSocket подключение** (строки 53-54):
   ```javascript
   const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
   ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
   ```
   **Рекомендация:** Использовать HTMX WebSocket расширение или Alpine.js для управления WebSocket

2. **Использование fetch()** (строка 126):
   ```javascript
   fetch('/message/' + threadId + '/thread/panel', {
       headers: { 'HX-Request': 'true' }
   })
   ```
   **Рекомендация:** Использовать HTMX атрибуты `hx-get`, `hx-target` вместо fetch()

3. **Использование htmx.ajax()** (строка 114):
   ```javascript
   htmx.ajax('GET', '/message/' + threadId + '/thread/panel', {
   ```
   **Рекомендация:** Использовать HTMX атрибуты напрямую в HTML

4. **DOM манипуляции** (множество строк):
   - `document.getElementById()`
   - `querySelector()`
   - `createElement()`
   - `innerHTML`
   - `addEventListener()`
   
   **Рекомендация:** 
   - Использовать HTMX для динамических обновлений
   - Использовать Alpine.js для условного рендеринга
   - Использовать Go компоненты для генерации HTML

5. **Глобальные функции** (строки 396-525):
   - `window.openThreadPanel()`
   - `window.closeThreadPanel()`
   - `window.switchThreadPanel()`
   - `window.scrollToNewReply()`
   
   **Рекомендация:** Использовать Alpine.js компоненты или HTMX атрибуты

---

### 4. `pkg/ui/pages/messenger/direct_message.go`
**Проблема:** Обширное использование чистого JavaScript для WebSocket и DOM манипуляций
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 80-81, 140-208

**Основные проблемы:**
1. **WebSocket подключение** (строки 80-81) - аналогично channel.go
2. **DOM манипуляции** (строки 140-208) - аналогично channel.go

**Рекомендация:** Те же, что и для channel.go

---

### 5. `pkg/ui/components/messenger/message_input.go`
**Проблема:** Использование чистого JavaScript для обработки файлов и форм
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 56, 84, 128-220

**Текущий код:**
```go
Attr("hx-on::after-request", "this.querySelector('textarea').value = ''; ...")
Attr("@keydown.enter", "if(!event.shiftKey) { event.preventDefault(); document.getElementById('message-form').requestSubmit(); }")
// Большой блок JavaScript для обработки файлов (строки 128-220)
```

**Рекомендация:**
- Для очистки формы использовать HTMX атрибуты или Alpine.js
- Для обработки файлов использовать HTMX `hx-encoding="multipart/form-data"`
- Для Enter использовать Alpine.js директивы

---

### 6. `pkg/ui/components/messenger/message_list.go`
**Проблема:** Использование чистого JavaScript для DOM манипуляций
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 151, 183, 221, 263

**Текущий код:**
```go
var rightPanel = document.getElementById('right-panel');
Attr("onclick", fmt.Sprintf("document.getElementById('thread-reply-form-%d').classList.add('hidden');", messageID))
```

**Рекомендация:**
- Использовать Alpine.js для условного отображения: `Attr("x-show", "!hidden")`
- Использовать HTMX для загрузки контента в панель

---

### 7. `pkg/ui/components/messenger/thread_panel.go`
**Проблема:** Использование чистого JavaScript для обработки ошибок и DOM манипуляций
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 73, 136-137, 163-169, 192-240

**Текущий код:**
```go
Attr("hx-on::htmx:response-error", `
    var errorDiv = document.createElement('div');
    // ... большой блок JavaScript
`)
```

**Рекомендация:**
- Использовать HTMX `hx-swap` для замены контента при ошибках
- Использовать Alpine.js для условного отображения сообщений об ошибках
- Генерировать HTML ошибок на сервере через Go компоненты

---

### 8. `pkg/ui/components/messenger/search.go`
**Проблема:** Использование чистого JavaScript для DOM манипуляций
**Нарушение:** Правило о минимизации JavaScript
**Строки:** 37-39

**Текущий код:**
```go
document.getElementById('search-results').classList.remove('hidden');
document.getElementById('search-results').classList.add('hidden');
```

**Рекомендация:**
- Использовать Alpine.js: `Attr("x-show", "showResults")` и `Attr("@click", "showResults = true")`

---

### 9. `pkg/ui/components/messenger/file_attachment.go`
**Проблема:** Использование `onclick` для открытия файлов
**Нарушение:** Правило о минимизации JavaScript
**Строка:** 67

**Текущий код:**
```go
Attr("onclick", fmt.Sprintf("window.open('%s', '_blank')", attachment.URL))
```

**Рекомендация:**
- Использовать обычную ссылку с `target="_blank"` или HTMX атрибут `hx-target="_blank"`

---

## Использование Tailwind CSS классов напрямую

### 10. `pkg/ui/components/messenger/search.go`
**Проблема:** Использование Tailwind CSS классов `bg-warning` и `text-warning-content` напрямую
**Нарушение:** Правило об использовании только DaisyUI компонентов
**Строка:** 202

**Текущий код:**
```go
result += escapedText[start:idx] + "<mark class='bg-warning text-warning-content'>" + matchedText + "</mark>"
```

**Рекомендация:**
- Проверить документацию DaisyUI для компонента `mark` или использовать DaisyUI классы
- Если DaisyUI не предоставляет такие классы, использовать альтернативу через DaisyUI компоненты

---

### 11. `pkg/ui/pages/messenger/channel.go` и `pkg/ui/pages/messenger/direct_message.go`
**Проблема:** Использование Tailwind CSS классов в `innerHTML`
**Нарушение:** Правило об использовании только DaisyUI компонентов
**Строки:** 
- channel.go: 277, 309
- direct_message.go: 142, 168

**Текущий код:**
```javascript
container.innerHTML = '<div id="typing-indicator" class="px-4 py-2 text-sm text-base-content/60 italic">' + ...
```

**Рекомендация:**
- Генерировать HTML через Go компоненты (Gomponents) вместо JavaScript `innerHTML`
- Использовать DaisyUI классы вместо прямых Tailwind CSS классов
- Использовать HTMX для обновления контента вместо `innerHTML`

---

## Использование fetch() и XMLHttpRequest вместо HTMX

### 12. `pkg/ui/pages/messenger/channel.go`
**Проблема:** Использование `fetch()` и `htmx.ajax()` вместо HTMX атрибутов
**Нарушение:** Правило об использовании HTMX для всех AJAX запросов
**Строки:** 114, 126

**Текущий код:**
```javascript
htmx.ajax('GET', '/message/' + threadId + '/thread/panel', {
fetch('/message/' + threadId + '/thread/panel', {
```

**Рекомендация:**
- Использовать HTMX атрибуты: `Attr("hx-get", r.Path("messenger.message.thread.panel", threadID))`
- Использовать `Attr("hx-target", "#right-panel")` для указания цели

---

## Резюме нарушений

### По типам:

1. **Чистый JavaScript для DOM манипуляций:** ~50+ случаев
2. **Чистый JavaScript для WebSocket:** 2 файла (channel.go, direct_message.go)
3. **Использование fetch()/htmx.ajax():** 2 случая
4. **Использование onclick/onchange:** 10+ случаев
5. **Прямое использование Tailwind CSS классов:** 5 случаев
6. **Использование innerHTML:** 4 случая

### Приоритет исправлений:

**Высокий приоритет:**
1. WebSocket подключения (channel.go, direct_message.go) - можно заменить на HTMX WebSocket
2. Использование fetch() - заменить на HTMX атрибуты
3. Глобальные функции window.* - заменить на Alpine.js компоненты

**Средний приоритет:**
1. DOM манипуляции через getElementById/querySelector - заменить на HTMX/Alpine.js
2. Обработчики событий через addEventListener - заменить на HTMX/Alpine.js атрибуты
3. Использование innerHTML - заменить на Go компоненты + HTMX

**Низкий приоритет:**
1. Мелкие onclick обработчики - заменить на Alpine.js или HTMX атрибуты
2. Прямое использование Tailwind CSS - проверить документацию DaisyUI и заменить

---

## Рекомендации по исправлению

### Общий подход:

1. **WebSocket:** Использовать HTMX WebSocket расширение или Alpine.js для управления соединением
2. **AJAX запросы:** Всегда использовать HTMX атрибуты (`hx-get`, `hx-post`, `hx-target`, `hx-swap`)
3. **Интерактивность:** Использовать Alpine.js (`x-data`, `x-show`, `@click`, `@input`)
4. **Условный рендеринг:** Использовать Alpine.js или серверный рендеринг через Go
5. **Генерация HTML:** Всегда использовать Go компоненты (Gomponents), никогда не использовать `innerHTML`
6. **Стилизация:** Использовать только DaisyUI компоненты и классы, Tailwind CSS только если разрешено документацией DaisyUI

### Примеры правильных решений:

**Вместо:**
```javascript
document.getElementById('modal').close();
```

**Использовать:**
```go
Attr("x-data", "{ open: false }"),
Attr("@click", "open = false"),
Attr("x-show", "open"),
```

**Вместо:**
```javascript
fetch('/api/data').then(response => response.text()).then(html => {
    document.getElementById('target').innerHTML = html;
});
```

**Использовать:**
```go
Attr("hx-get", r.Path("api.data")),
Attr("hx-target", "#target"),
Attr("hx-swap", "innerHTML"),
```

**Вместо:**
```javascript
const ws = new WebSocket('ws://localhost/ws');
ws.onmessage = function(event) { ... };
```

**Использовать:**
```go
// HTMX WebSocket или Alpine.js компонент
Attr("hx-ws", "connect:/ws"),
```

---

**Дата проверки:** 2025-01-XX
**Версия DEVELOPMENT_GUIDE.md:** Текущая

---

## Поэтапный план устранения нарушений

### Этап 1: Критические нарушения - WebSocket и AJAX запросы (Высокий приоритет)

**Цель:** Устранить использование чистого JavaScript для WebSocket и fetch() запросов

#### 1.1. Замена WebSocket подключений на HTMX WebSocket

**Файлы:**
- `pkg/ui/pages/messenger/channel.go` (строки 53-54, 80-525)
- `pkg/ui/pages/messenger/direct_message.go` (строки 80-81, 140-208)

**Шаги:**
1. Изучить документацию HTMX WebSocket расширения
2. Создать Go компонент для WebSocket подключения через HTMX
3. Заменить `new WebSocket()` на HTMX атрибуты `hx-ws="connect:/ws"`
4. Заменить обработчики `ws.onmessage`, `ws.onopen`, `ws.onclose` на HTMX события
5. Протестировать WebSocket функциональность
6. Удалить весь JavaScript код, связанный с WebSocket

**Ожидаемый результат:** WebSocket работает через HTMX, без чистого JavaScript

**Время:** 4-6 часов

---

#### 1.2. Замена fetch() и htmx.ajax() на HTMX атрибуты

**Файлы:**
- `pkg/ui/pages/messenger/channel.go` (строки 114, 126)

**Шаги:**
1. Найти все места использования `fetch()` и `htmx.ajax()`
2. Заменить на HTMX атрибуты:
   - `fetch('/message/...')` → `Attr("hx-get", r.Path("messenger.message.thread.panel", threadID))`
   - `htmx.ajax('GET', ...)` → `Attr("hx-get", ...)`
3. Указать правильные `hx-target` и `hx-swap` атрибуты
4. Протестировать функциональность загрузки панели треда
5. Удалить JavaScript код

**Ожидаемый результат:** Все AJAX запросы выполняются через HTMX атрибуты

**Время:** 2-3 часа

---

### Этап 2: DOM манипуляции и глобальные функции (Высокий приоритет)

**Цель:** Заменить все DOM манипуляции на HTMX/Alpine.js/Go компоненты

#### 2.1. Замена глобальных функций window.* на Alpine.js компоненты

**Файлы:**
- `pkg/ui/pages/messenger/channel.go` (строки 396-525)

**Функции для замены:**
- `window.openThreadPanel(threadId)`
- `window.closeThreadPanel()`
- `window.switchThreadPanel(newThreadId)`
- `window.scrollToNewReply()`

**Шаги:**
1. Создать Alpine.js компонент для управления состоянием панели треда:
   ```go
   Attr("x-data", `{
       openThread(threadId) { ... },
       closeThread() { ... },
       switchThread(newThreadId) { ... },
       scrollToNewReply() { ... }
   }`)
   ```
2. Заменить все вызовы `window.openThreadPanel()` на Alpine.js методы
3. Использовать HTMX для загрузки контента панели
4. Использовать Alpine.js `x-show` для условного отображения
5. Протестировать открытие/закрытие/переключение панели треда
6. Удалить глобальные функции

**Ожидаемый результат:** Панель треда управляется через Alpine.js и HTMX

**Время:** 4-5 часов

---

#### 2.2. Замена DOM манипуляций на HTMX/Alpine.js

**Файлы:**
- `pkg/ui/pages/messenger/channel.go` (строки 101, 108, 154, 170, 186, 226-262, 275-299, 320-357)
- `pkg/ui/pages/messenger/direct_message.go` (строки 140-208)
- `pkg/ui/components/messenger/message_list.go` (строки 151, 183, 263)
- `pkg/ui/components/messenger/search.go` (строки 37-39)

**Шаги:**
1. **Замена getElementById/querySelector:**
   - Использовать HTMX `hx-target` для обновления элементов
   - Использовать Alpine.js `x-ref` для ссылок на элементы
   - Использовать HTMX `hx-swap` для замены контента

2. **Замена createElement/innerHTML:**
   - Генерировать HTML через Go компоненты (Gomponents)
   - Использовать HTMX для вставки сгенерированного HTML
   - Никогда не использовать `innerHTML` в JavaScript

3. **Замена addEventListener:**
   - Использовать HTMX атрибуты (`hx-on::event`)
   - Использовать Alpine.js директивы (`@click`, `@input`, `@change`)

4. **Конкретные замены:**
   - `document.getElementById('message-list')` → HTMX `hx-target="#message-list"`
   - `element.innerHTML = html` → HTMX `hx-swap="innerHTML"` + Go компонент
   - `element.classList.add/remove()` → Alpine.js `x-show`, `x-bind:class`
   - `addEventListener('input', ...)` → Alpine.js `@input` или HTMX `hx-on::input`

5. Протестировать все функции после замены
6. Удалить весь JavaScript код для DOM манипуляций

**Ожидаемый результат:** Все DOM манипуляции выполняются через HTMX/Alpine.js/Go

**Время:** 8-10 часов

---

### Этап 3: Обработка форм и файлов (Средний приоритет)

**Цель:** Заменить JavaScript обработку форм на HTMX/Alpine.js

#### 3.1. Замена обработки файлов

**Файлы:**
- `pkg/ui/components/messenger/message_input.go` (строки 128-220)

**Шаги:**
1. Использовать HTMX `hx-encoding="multipart/form-data"` для загрузки файлов
2. Использовать Alpine.js для превью файлов:
   ```go
   Attr("x-data", `{
       files: [],
       addFile(file) { ... },
       removeFile(index) { ... },
       showPreview(file) { ... }
   }`)
   ```
3. Использовать Alpine.js `@change` для обработки выбора файлов
4. Генерировать превью через Go компоненты или Alpine.js
5. Протестировать загрузку файлов
6. Удалить JavaScript код для обработки файлов

**Ожидаемый результат:** Загрузка файлов работает через HTMX + Alpine.js

**Время:** 3-4 часа

---

#### 3.2. Замена обработки форм

**Файлы:**
- `pkg/ui/components/messenger/message_input.go` (строки 56, 84)
- `pkg/ui/components/messenger/thread_panel.go` (строки 136-137)
- `pkg/ui/components/messenger/message_list.go` (строка 221)

**Шаги:**
1. Заменить очистку формы:
   - Вместо `this.querySelector('textarea').value = ''` использовать HTMX `hx-on::after-request`
   - Или использовать Alpine.js для управления состоянием формы

2. Заменить обработку Enter:
   - Вместо `@keydown.enter` с `preventDefault()` использовать HTMX `hx-trigger="submit"`
   - Или использовать Alpine.js `@keydown.enter.prevent`

3. Протестировать отправку форм
4. Удалить JavaScript код для обработки форм

**Ожидаемый результат:** Все формы обрабатываются через HTMX

**Время:** 2-3 часа

---

### Этап 4: Модальные окна и интерактивные элементы (Средний приоритет)

**Цель:** Заменить onclick и DOM манипуляции на Alpine.js

#### 4.1. Замена модальных окон

**Файлы:**
- `pkg/ui/components/messenger/modals.go` (строки 65, 83, 181, 202, 241, 259, 365, 384)

**Шаги:**
1. Создать Alpine.js компонент для модальных окон:
   ```go
   Attr("x-data", `{
       open: false,
       show() { this.open = true; },
       close() { this.open = false; }
   }`)
   Attr("x-show", "open"),
   Attr("@click.away", "close()"),
   ```

2. Заменить `onclick="document.getElementById('modal').close()"` на Alpine.js методы

3. Заменить закрытие модального окна после HTMX запроса:
   - Вместо JavaScript в `hx-on::after-request` использовать Alpine.js
   - Или использовать HTMX `hx-on::htmx:after-request` с Alpine.js методом

4. Протестировать открытие/закрытие модальных окон
5. Удалить JavaScript код для модальных окон

**Ожидаемый результат:** Модальные окна управляются через Alpine.js

**Время:** 2-3 часа

---

#### 4.2. Замена onclick обработчиков

**Файлы:**
- `pkg/ui/components/messenger/file_attachment.go` (строка 67)
- `pkg/ui/components/messenger/message_list.go` (строка 263)

**Шаги:**
1. Заменить `onclick="window.open(...)"` на обычную ссылку с `target="_blank"`
2. Заменить `onclick="element.classList.add('hidden')"` на Alpine.js `x-show` или `x-bind:class`
3. Протестировать функциональность
4. Удалить onclick атрибуты

**Ожидаемый результат:** Все интерактивные элементы используют Alpine.js или нативные HTML атрибуты

**Время:** 1-2 часа

---

### Этап 5: Обработка ошибок (Средний приоритет)

**Цель:** Заменить JavaScript обработку ошибок на серверный рендеринг

#### 5.1. Замена обработки ошибок в формах

**Файлы:**
- `pkg/ui/components/messenger/thread_panel.go` (строки 163-169, 192-240)

**Шаги:**
1. Создать Go компонент для отображения ошибок:
   ```go
   func ErrorMessage(message string) Node {
       return Div(
           Class("alert alert-error mt-2"),
           Text(message),
       )
   }
   ```

2. На сервере генерировать HTML ошибки через Go компоненты
3. Использовать HTMX `hx-swap` для замены контента при ошибках
4. Использовать Alpine.js для условного отображения ошибок
5. Протестировать отображение ошибок
6. Удалить JavaScript код для обработки ошибок

**Ожидаемый результат:** Ошибки отображаются через серверный рендеринг + HTMX

**Время:** 2-3 часа

---

### Этап 6: HTMX слушатели событий (Низкий приоритет)

**Цель:** Минимизировать JavaScript в HTMX слушателях

#### 6.1. Оптимизация HTMX слушателей

**Файлы:**
- `pkg/ui/components/htmx.go` (строки 12-48)

**Шаги:**
1. **Обработка ошибок (строки 12-18):**
   - Убрать глобальный обработчик `htmx:beforeSwap`
   - Использовать `hx-on::response-error` на конкретных формах
   - Использовать HTMX `hx-swap` для замены контента при ошибках

2. **CSRF токен (строки 21-27):**
   - ✅ Уже исправлено: используется Go компонент `components.CSRFInput(r)`
   - Убедиться, что глобальный обработчик не нужен

3. **Модальные окна (строки 29-39):**
   - Заменить на Alpine.js компонент
   - Использовать Alpine.js `x-show` для управления видимостью
   - Использовать HTMX `hx-on::after-swap` с Alpine.js методом

4. Протестировать все функции
5. Минимизировать или удалить JavaScript в HTMX слушателях

**Ожидаемый результат:** Минимум JavaScript в HTMX слушателях, максимум через HTMX атрибуты и Alpine.js

**Время:** 2-3 часа

---

### Этап 7: Замена Tailwind CSS на DaisyUI (Низкий приоритет)

**Цель:** Использовать только DaisyUI компоненты и классы

#### 7.1. Проверка и замена Tailwind CSS классов

**Файлы:**
- `pkg/ui/components/messenger/search.go` (строка 202)
- `pkg/ui/pages/messenger/channel.go` (строки 277, 309)
- `pkg/ui/pages/messenger/direct_message.go` (строки 142, 168)

**Шаги:**
1. Проверить документацию DaisyUI для каждого используемого класса:
   - `bg-warning`, `text-warning-content` → проверить DaisyUI альтернативы
   - `px-4`, `py-2`, `text-sm`, `text-base-content/60`, `italic` → проверить DaisyUI классы

2. Заменить классы на DaisyUI эквиваленты:
   - Если DaisyUI предоставляет класс → использовать его
   - Если нет → проверить, разрешает ли документация DaisyUI использование Tailwind CSS в этом случае

3. Заменить `innerHTML` с классами на Go компоненты:
   - Генерировать HTML через Gomponents с DaisyUI классами
   - Использовать HTMX для обновления контента

4. Протестировать стилизацию
5. Удалить использование `innerHTML` с Tailwind CSS классами

**Ожидаемый результат:** Все стили используют только DaisyUI классы (или Tailwind CSS, если разрешено документацией DaisyUI)

**Время:** 2-3 часа

---

## Общий план выполнения

### Фаза 1: Критические нарушения (Неделя 1)
- ✅ Этап 1.1: WebSocket через HTMX (4-6 часов)
- ✅ Этап 1.2: Замена fetch() на HTMX (2-3 часа)
- ✅ Этап 2.1: Глобальные функции на Alpine.js (4-5 часов)

**Итого:** 10-14 часов

### Фаза 2: DOM манипуляции (Неделя 2)
- ✅ Этап 2.2: DOM манипуляции на HTMX/Alpine.js (8-10 часов)
- ✅ Этап 3.1: Обработка файлов (3-4 часа)
- ✅ Этап 3.2: Обработка форм (2-3 часа)

**Итого:** 13-17 часов

### Фаза 3: Интерактивность и ошибки (Неделя 3)
- ✅ Этап 4.1: Модальные окна (2-3 часа)
- ✅ Этап 4.2: onclick обработчики (1-2 часа)
- ✅ Этап 5.1: Обработка ошибок (2-3 часа)
- ✅ Этап 6.1: HTMX слушатели (2-3 часа)

**Итого:** 7-11 часов

### Фаза 4: Стилизация (Неделя 4)
- ✅ Этап 7.1: Замена Tailwind CSS на DaisyUI (2-3 часа)
- ✅ Финальное тестирование всех функций (4-6 часов)
- ✅ Рефакторинг и оптимизация (2-4 часа)

**Итого:** 8-13 часов

---

## Общее время выполнения

**Минимум:** 38 часов (≈ 5 рабочих дней)
**Максимум:** 55 часов (≈ 7 рабочих дней)

---

## Критерии успешного завершения

1. ✅ Нет использования `document.*`, `window.*` (кроме необходимых для HTMX/Alpine.js)
2. ✅ Нет использования `fetch()`, `XMLHttpRequest`, `htmx.ajax()`
3. ✅ Нет использования `innerHTML`, `createElement()`, `appendChild()`
4. ✅ Нет использования `addEventListener()` (кроме HTMX/Alpine.js)
5. ✅ Нет использования `onclick`, `onchange`, `onsubmit` (кроме Alpine.js директив)
6. ✅ Все AJAX запросы через HTMX атрибуты
7. ✅ Все интерактивные элементы через Alpine.js
8. ✅ Все стили через DaisyUI (или Tailwind CSS, если разрешено документацией)
9. ✅ Все HTML генерируется через Go компоненты (Gomponents)
10. ✅ Все тесты проходят
11. ✅ Функциональность не нарушена

---

## Рекомендации по выполнению

1. **Начинать с критических нарушений** (Этап 1-2) - они влияют на архитектуру
2. **Тестировать после каждого этапа** - не накапливать проблемы
3. **Создавать переиспользуемые компоненты** - для Alpine.js и Go
4. **Документировать изменения** - обновлять комментарии в коде
5. **Использовать TDD подход** - писать тесты перед рефакторингом
6. **Проверять документацию** - HTMX, Alpine.js, DaisyUI перед использованием
7. **Не смешивать подходы** - использовать один подход для одной задачи

---

## Полезные ресурсы

- [HTMX Documentation](https://htmx.org/docs/)
- [HTMX WebSocket Extension](https://htmx.org/extensions/web-sockets/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [DaisyUI Components](https://daisyui.com/components/)
- [DaisyUI Utilities](https://daisyui.com/docs/utilities/)
- [Gomponents Documentation](https://github.com/maragudk/gomponents)
