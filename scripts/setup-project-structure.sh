#!/bin/bash
# Скрипт для создания структуры директорий проекта
# Использование: ./scripts/setup-project-structure.sh

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}📁 Создание структуры директорий для Slack-клона...${NC}"

# Создать директории для пользовательского кода
mkdir -p app/handlers
mkdir -p app/services
mkdir -p app/ui/components/messenger
mkdir -p app/ui/pages/messenger
mkdir -p app/ui/forms/messenger
mkdir -p app/ui/layouts
mkdir -p app/websocket
mkdir -p app/middleware

# Создать директории для документации
mkdir -p docs

# Создать .gitkeep файлы
touch app/handlers/.gitkeep
touch app/services/.gitkeep
touch app/ui/components/messenger/.gitkeep
touch app/ui/pages/messenger/.gitkeep
touch app/ui/forms/messenger/.gitkeep
touch app/ui/layouts/.gitkeep
touch app/websocket/.gitkeep
touch app/middleware/.gitkeep

echo -e "${GREEN}✅ Структура директорий создана:${NC}"
echo ""
echo "app/"
echo "├── handlers/          # Ваши HTTP handlers"
echo "├── services/          # Ваши сервисы"
echo "├── ui/"
echo "│   ├── components/messenger/  # UI компоненты мессенджера"
echo "│   ├── pages/messenger/      # Страницы мессенджера"
echo "│   ├── forms/messenger/      # Формы мессенджера"
echo "│   └── layouts/              # Ваши layouts"
echo "├── websocket/        # WebSocket инфраструктура"
echo "└── middleware/       # Ваши middleware"
echo ""
echo -e "${YELLOW}💡 Теперь можно начинать разработку!${NC}"

