# Быстрый старт: Разделение Pagoda и вашего кода

## Шаг 1: Настройка Git репозиториев

```bash
# 1. Если еще не сделали форк, сделайте его на GitHub/GitLab
# 2. Добавьте upstream (оригинальный репозиторий Pagoda)
git remote add upstream https://github.com/mikestefanello/pagoda.git

# 3. Проверьте настройки
git remote -v
```

## Шаг 2: Создание структуры проекта

```bash
# Запустить скрипт создания структуры
./scripts/setup-project-structure.sh
```

Это создаст:
- `app/` - директория для вашего кода
- Все необходимые поддиректории

## Шаг 3: Создание веток

```bash
# Ветка для синхронизации с Pagoda
git checkout -b pagoda-upstream

# Ветка для разработки Slack
git checkout -b slack-development
```

## Шаг 4: Начало разработки

```bash
# Переключиться на рабочую ветку
git checkout slack-development

# Создать первый файл в app/
touch app/handlers/messenger.go
```

## Шаг 5: Синхронизация с Pagoda (когда нужно)

```bash
# Запустить скрипт синхронизации
./scripts/sync-pagoda.sh
```

Или вручную:

```bash
# 1. Переключиться на ветку синхронизации
git checkout pagoda-upstream

# 2. Получить обновления
git fetch upstream
git merge upstream/main

# 3. Вернуться на рабочую ветку
git checkout slack-development

# 4. Слить обновления
git merge pagoda-upstream
```

## Структура вашего кода

```
app/                          # ВАШ КОД
├── handlers/
│   └── messenger.go         # Ваши handlers
├── services/
│   └── notifications.go     # Ваши сервисы
├── ui/
│   ├── components/messenger/
│   ├── pages/messenger/
│   └── forms/messenger/
└── websocket/
    └── hub.go

ent/schema/                   # СХЕМЫ
├── user.go                  # Pagoda (не изменять)
├── passwordtoken.go         # Pagoda (не изменять)
├── workspace.go             # ВАШ КОД
├── channel.go               # ВАШ КОД
└── message.go              # ВАШ КОД

pkg/                          # PAGODA (стараться не изменять)
└── ...
```

## Правила работы

### ✅ Делайте

1. Весь ваш код в `app/`
2. Ваши схемы Ent в `ent/schema/` (новые файлы)
3. Документируйте изменения в `CHANGES.md`
4. Регулярно синхронизируйтесь с upstream

### ❌ Не делайте

1. Не изменяйте напрямую `pkg/` файлы
2. Не изменяйте `ent/schema/user.go` и `passwordtoken.go`
3. Не коммитьте без проверки конфликтов

## Полезные команды

```bash
# Проверить различия с upstream
git diff upstream/main

# Посмотреть измененные файлы
git diff --name-only upstream/main

# Просмотреть историю
git log --oneline --graph --all
```

## Дополнительная информация

См. `SEPARATION_GUIDE.md` для подробной документации.

