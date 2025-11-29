# Разделение кода: Pagoda vs Ваш код

## 📋 Обзор

Этот проект использует Pagoda как основу и добавляет функциональность Slack-клона. Для правильной работы с обновлениями Pagoda необходимо четко разделять код.

## 🗂️ Структура

```
pagoda/
├── app/                    # 👤 ВАШ КОД
│   ├── handlers/          # Ваши HTTP handlers
│   ├── services/          # Ваши сервисы
│   ├── ui/                # Ваши UI компоненты
│   ├── websocket/         # WebSocket инфраструктура
│   └── routenames/        # Ваши route names
│
├── pkg/                   # 📦 PAGODA (стараться не изменять)
│   ├── handlers/          # Pagoda handlers
│   ├── services/          # Pagoda services
│   └── ...
│
├── ent/schema/            # 📊 СХЕМЫ
│   ├── user.go           # Pagoda (не изменять)
│   ├── passwordtoken.go  # Pagoda (не изменять)
│   ├── workspace.go      # 👤 ВАШ КОД
│   └── channel.go        # 👤 ВАШ КОД
│
├── config/
│   └── config.yaml       # Pagoda + ваши настройки
│
└── scripts/              # 🔧 УТИЛИТЫ
    ├── sync-pagoda.sh    # Синхронизация с Pagoda
    └── setup-project-structure.sh
```

## 🚀 Быстрый старт

1. **Настройка Git:**
   ```bash
   git remote add upstream https://github.com/mikestefanello/pagoda.git
   ```

2. **Создание структуры:**
   ```bash
   ./scripts/setup-project-structure.sh
   ```

3. **Создание веток:**
   ```bash
   git checkout -b pagoda-upstream
   git checkout -b slack-development
   ```

4. **Начало разработки:**
   ```bash
   git checkout slack-development
   # Ваш код в app/
   ```

## 🔄 Синхронизация с Pagoda

```bash
./scripts/sync-pagoda.sh
```

Или вручную:
```bash
git checkout pagoda-upstream
git fetch upstream
git merge upstream/main
git checkout slack-development
git merge pagoda-upstream
```

## 📚 Документация

- **SEPARATION_GUIDE.md** - подробное руководство
- **QUICK_START.md** - быстрый старт
- **CHANGES.md** - отслеживание изменений

## ✅ Правила

### Делайте:
- ✅ Весь ваш код в `app/`
- ✅ Новые схемы Ent в `ent/schema/` (новые файлы)
- ✅ Документируйте изменения в `CHANGES.md`
- ✅ Используйте маркеры `CUSTOM CODE START/END`

### Не делайте:
- ❌ Не изменяйте напрямую `pkg/` файлы
- ❌ Не изменяйте `ent/schema/user.go` и `passwordtoken.go`
- ❌ Не коммитьте без проверки конфликтов

## 🔍 Полезные команды

```bash
# Проверить различия с upstream
git diff upstream/main

# Посмотреть измененные файлы
git diff --name-only upstream/main

# Просмотреть историю
git log --oneline --graph --all
```

## 📝 Примеры

См. `app/handlers/example.go` для примера правильного оформления кода.

