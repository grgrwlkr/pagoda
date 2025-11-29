# Руководство по разделению кода Pagoda и пользовательских наработок

## Обзор стратегии

Данное руководство описывает подход к разделению оригинального кода Pagoda и ваших наработок (Slack-клон), позволяющий:
- ✅ Обновлять Pagoda по мере его развития
- ✅ Применять обновления Pagoda в свой проект
- ✅ Чётко разграничить свой код и код репозитория
- ✅ Избежать конфликтов при слиянии

---

## Стратегия 1: Git Fork + Upstream (Рекомендуется)

### Шаг 1: Настройка репозиториев

```bash
# 1. Форкнуть оригинальный репозиторий на GitHub/GitLab
# (через веб-интерфейс или команду ниже)

# 2. Клонировать свой форк
git clone git@github.com:YOUR_USERNAME/pagoda-slack.git
cd pagoda-slack

# 3. Добавить оригинальный репозиторий как upstream
git remote add upstream https://github.com/mikestefanello/pagoda.git

# 4. Проверить настройки
git remote -v
# origin    git@github.com:YOUR_USERNAME/pagoda-slack.git (fetch)
# origin    git@github.com:YOUR_USERNAME/pagoda-slack.git (push)
# upstream  https://github.com/mikestefanello/pagoda.git (fetch)
# upstream  https://github.com/mikestefanello/pagoda.git (push)
```

### Шаг 2: Создание структуры веток

```bash
# 1. Создать ветку для разработки Slack
git checkout -b slack-development

# 2. Создать ветку для синхронизации с upstream
git checkout -b pagoda-upstream
```

### Шаг 3: Создание структуры директорий для пользовательского кода

Создайте следующую структуру для четкого разделения:

```
project/
├── app/                    # ВАШ КОД - приложение Slack
│   ├── handlers/
│   │   └── messenger.go
│   ├── services/
│   │   └── notifications.go
│   ├── ui/
│   │   ├── components/
│   │   │   └── messenger/
│   │   ├── pages/
│   │   │   └── messenger/
│   │   └── forms/
│   │       └── messenger/
│   └── websocket/
│       ├── hub.go
│       └── ...
│
├── ent/
│   └── schema/
│       ├── user.go         # Pagoda
│       ├── passwordtoken.go # Pagoda
│       ├── workspace.go    # ВАШ КОД
│       ├── channel.go      # ВАШ КОД
│       └── message.go      # ВАШ КОД
│
├── pkg/                    # Pagoda (старайтесь не изменять)
│   ├── handlers/           # Pagoda handlers
│   ├── services/           # Pagoda services
│   └── ...
│
├── config/
│   ├── config.yaml         # Pagoda + ваши настройки
│   └── app.yaml            # ВАШ КОД - только ваши настройки
│
└── cmd/
    ├── web/
    │   └── main.go         # Pagoda (можно расширить)
    └── admin/
        └── main.go         # Pagoda
```

### Шаг 4: Настройка .gitattributes для отслеживания изменений

Создайте `.gitattributes`:

```gitattributes
# Файлы Pagoda (оригинальные)
/pkg/** linguist-vendored=false
/config/config.go linguist-vendored=false
/cmd/** linguist-vendored=false
/ent/schema/user.go linguist-vendored=false
/ent/schema/passwordtoken.go linguist-vendored=false

# Ваш код (не оригинальный)
/app/** linguist-vendored=false
/ent/schema/workspace.go linguist-vendored=false
/ent/schema/channel.go linguist-vendored=false
/ent/schema/message.go linguist-vendored=false
```

---

## Стратегия 2: Структура с префиксами и маркерами

### Маркеры в коде

Используйте комментарии для обозначения вашего кода:

```go
// ============================================================================
// CUSTOM CODE START - Slack Messenger Feature
// ============================================================================
// This section contains custom code for Slack clone functionality.
// Do not modify Pagoda core files above this marker.
// ============================================================================

package messenger

// ... ваш код ...

// ============================================================================
// CUSTOM CODE END
// ============================================================================
```

### Префиксы в именах

Используйте префиксы для ваших файлов/пакетов:
- `app/` вместо `pkg/` для вашего кода
- Или `pkg/slack/` для Slack-специфичного кода

---

## Workflow для синхронизации с Pagoda

### Получение обновлений от Pagoda

```bash
# 1. Переключиться на ветку для синхронизации
git checkout pagoda-upstream

# 2. Получить последние изменения от upstream
git fetch upstream

# 3. Слить изменения в ветку синхронизации
git merge upstream/main

# 4. Переключиться на вашу рабочую ветку
git checkout slack-development

# 5. Слить обновления Pagoda в вашу ветку
git merge pagoda-upstream

# 6. Разрешить конфликты (если есть)
# Git автоматически покажет конфликты между вашим кодом и Pagoda
```

### Автоматизация через скрипт

Создайте `scripts/sync-pagoda.sh`:

```bash
#!/bin/bash
set -e

echo "🔄 Синхронизация с Pagoda upstream..."

# Переключиться на ветку синхронизации
git checkout pagoda-upstream

# Получить обновления
git fetch upstream

# Слить изменения
git merge upstream/main || {
    echo "⚠️  Конфликты при слиянии. Разрешите вручную."
    exit 1
}

# Вернуться на рабочую ветку
git checkout slack-development

# Слить обновления
git merge pagoda-upstream || {
    echo "⚠️  Конфликты при слиянии в рабочую ветку. Разрешите вручную."
    exit 1
}

echo "✅ Синхронизация завершена!"
```

Сделайте скрипт исполняемым:
```bash
chmod +x scripts/sync-pagoda.sh
```

---

## Разделение конфигурации

### config/config.yaml (Pagoda + ваши настройки)

```yaml
# ============================================================================
# PAGODA CORE CONFIGURATION
# ============================================================================
http:
  port: 8000
  # ... остальная конфигурация Pagoda

# ============================================================================
# SLACK MESSENGER CONFIGURATION (CUSTOM)
# ============================================================================
messenger:
  maxFileSize: 10485760
  allowedFileTypes: ["image/*", "application/pdf"]
  messageLimit: 50
  typingTimeout: "3s"
  onlineTimeout: "5m"
```

### config/app.yaml (только ваши настройки)

Создайте отдельный файл для ваших настроек и загружайте его дополнительно:

```go
// В config/config.go добавить:
func LoadAppConfig() (*AppConfig, error) {
    // Загрузка app.yaml
}
```

---

## Разделение handlers

### Подход 1: Отдельный пакет

Создайте `app/handlers/` вместо изменения `pkg/handlers/`:

```go
// app/handlers/messenger.go
package handlers

import (
    "github.com/mikestefanello/pagoda/pkg/services"
    "github.com/labstack/echo/v4"
)

type Messenger struct{}

func init() {
    // Регистрация вашего handler
    handlers.Register(new(Messenger))
}

func (h *Messenger) Init(c *services.Container) error {
    return nil
}

func (h *Messenger) Routes(g *echo.Group) {
    // Ваши routes
}
```

### Подход 2: Расширение существующих handlers

Если нужно расширить существующий handler, создайте wrapper:

```go
// app/handlers/pages_extended.go
package handlers

import (
    pagodaHandlers "github.com/mikestefanello/pagoda/pkg/handlers"
)

// Расширение существующего handler
type PagesExtended struct {
    *pagodaHandlers.Pages
}

func (h *PagesExtended) Routes(g *echo.Group) {
    // Вызвать оригинальные routes
    h.Pages.Routes(g)
    
    // Добавить свои routes
    g.GET("/slack", h.SlackHome)
}
```

---

## Разделение UI компонентов

### Структура

```
pkg/ui/                    # Pagoda компоненты
├── components/
│   ├── alerts.go         # Pagoda
│   └── nav.go            # Pagoda
└── layouts/
    └── primary.go        # Pagoda

app/ui/                    # Ваши компоненты
├── components/
│   └── messenger/        # Slack компоненты
└── layouts/
    └── messenger.go      # Slack layout
```

### Использование в коде

```go
// app/ui/pages/messenger/channel.go
package messenger

import (
    pagodaUI "github.com/mikestefanello/pagoda/pkg/ui"
    "github.com/mikestefanello/pagoda/app/ui/components/messenger"
)

func ChannelPage(ctx echo.Context) error {
    r := pagodaUI.NewRequest(ctx)  // Используем Pagoda
    
    content := messenger.ChannelList(r)  // Используем свой компонент
    
    return r.Render(messenger.Layout, content)
}
```

---

## Разделение схем Ent

### Структура

```
ent/schema/
├── user.go              # Pagoda - НЕ ИЗМЕНЯТЬ напрямую
├── passwordtoken.go     # Pagoda - НЕ ИЗМЕНЯТЬ напрямую
├── workspace.go         # ВАШ КОД
├── channel.go           # ВАШ КОД
└── message.go           # ВАШ КОД
```

### Расширение User схемы

Если нужно расширить User, создайте отдельный файл:

```go
// ent/schema/user_extensions.go
package schema

import (
    "github.com/mikestefanello/pagoda/ent/schema"
)

// UserExtensions добавляет поля к User через хук
func UserExtensions() {
    // Используйте Ent hooks для добавления полей
    // Или создайте отдельную сущность UserProfile
}
```

**Лучше:** Создать отдельную сущность `UserProfile` с отношением к `User`:

```go
// ent/schema/userprofile.go
type UserProfile struct {
    ent.Schema
}

func (UserProfile) Fields() []ent.Field {
    return []ent.Field{
        field.String("avatar_url"),
        field.String("status"),
        field.String("status_message"),
        // ...
    }
}

func (UserProfile) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("user", User.Type).Unique().Required(),
    }
}
```

---

## Обработка конфликтов при слиянии

### Стратегия разрешения конфликтов

1. **Конфликты в Pagoda файлах:**
   - Обычно принимайте версию из upstream (Pagoda)
   - Ваши изменения переносите в отдельные файлы

2. **Конфликты в ваших файлах:**
   - Разрешайте вручную, сохраняя свою логику

3. **Конфликты в конфигурации:**
   - Сливайте обе версии
   - Используйте отдельные секции

### Git merge strategy

```bash
# Настроить стратегию слияния для Pagoda файлов
git config merge.ours.driver true

# В .gitattributes
/pkg/** merge=ours
/config/config.go merge=ours
```

---

## Документирование изменений

### Файл CHANGES.md

Создайте файл для отслеживания ваших изменений:

```markdown
# Изменения в Slack-клоне

## Добавленные файлы
- app/handlers/messenger.go
- app/services/notifications.go
- ent/schema/workspace.go
- ...

## Модифицированные файлы Pagoda
- config/config.yaml (добавлена секция messenger)
- cmd/web/main.go (добавлена регистрация WebSocket)

## Зависимости
- Добавлены: gorilla/websocket (если нужно)
```

---

## CI/CD для проверки синхронизации

### GitHub Actions workflow

Создайте `.github/workflows/sync-check.yml`:

```yaml
name: Check Pagoda Sync

on:
  schedule:
    - cron: '0 0 * * 0'  # Каждое воскресенье
  workflow_dispatch:

jobs:
  check-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Check upstream sync
        run: |
          git fetch upstream
          git checkout pagoda-upstream
          git merge upstream/main || exit 1
          echo "✅ Upstream синхронизирован"
```

---

## Рекомендации

### ✅ Делайте

1. **Используйте отдельные директории** (`app/`) для вашего кода
2. **Документируйте изменения** в Pagoda файлах
3. **Регулярно синхронизируйтесь** с upstream (раз в неделю/месяц)
4. **Тестируйте после синхронизации**
5. **Используйте ветки** для изоляции изменений

### ❌ Не делайте

1. **Не изменяйте напрямую** Pagoda core файлы без необходимости
2. **Не коммитьте** изменения в Pagoda файлы без маркеров
3. **Не игнорируйте** конфликты при слиянии
4. **Не удаляйте** оригинальные файлы Pagoda

---

## Быстрый старт

```bash
# 1. Настроить upstream
git remote add upstream https://github.com/mikestefanello/pagoda.git

# 2. Создать структуру
mkdir -p app/{handlers,services,ui/{components/messenger,pages/messenger,forms/messenger},websocket}

# 3. Создать ветки
git checkout -b slack-development
git checkout -b pagoda-upstream

# 4. Начать разработку
git checkout slack-development
# ... ваш код в app/
```

---

## Полезные команды

```bash
# Проверить различия с upstream
git diff upstream/main

# Посмотреть файлы, измененные относительно upstream
git diff --name-only upstream/main

# Создать патч с вашими изменениями
git format-patch upstream/main

# Просмотреть историю синхронизаций
git log --oneline --graph --all
```

---

## Альтернативный подход: Go Modules

Если хотите использовать Pagoda как зависимость:

```go
// go.mod
module github.com/yourusername/slack-clone

require (
    github.com/mikestefanello/pagoda v0.1.0
    // ...
)
```

Но это менее гибко для кастомизации.

---

## Заключение

Рекомендуемый подход: **Git Fork + Upstream + отдельная структура директорий (`app/`)**.

Это позволит:
- ✅ Легко получать обновления Pagoda
- ✅ Чётко разделять код
- ✅ Избегать большинства конфликтов
- ✅ Поддерживать проект в актуальном состоянии

