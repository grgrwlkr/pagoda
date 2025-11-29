#!/bin/bash
# Скрипт для синхронизации с Pagoda upstream
# Использование: ./scripts/sync-pagoda.sh

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔄 Синхронизация с Pagoda upstream...${NC}"

# Проверка наличия upstream
if ! git remote | grep -q upstream; then
    echo -e "${YELLOW}⚠️  Upstream не настроен. Добавляю...${NC}"
    git remote add upstream https://github.com/mikestefanello/pagoda.git
fi

# Сохранить текущую ветку
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${GREEN}📌 Текущая ветка: ${CURRENT_BRANCH}${NC}"

# Создать или переключиться на ветку синхронизации
if git show-ref --verify --quiet refs/heads/pagoda-upstream; then
    echo -e "${GREEN}✓ Ветка pagoda-upstream существует${NC}"
    git checkout pagoda-upstream
else
    echo -e "${YELLOW}⚠️  Создаю ветку pagoda-upstream${NC}"
    git checkout -b pagoda-upstream
fi

# Получить обновления
echo -e "${GREEN}📥 Получение обновлений от upstream...${NC}"
git fetch upstream

# Слить изменения
echo -e "${GREEN}🔀 Слияние изменений...${NC}"
if git merge upstream/main; then
    echo -e "${GREEN}✅ Слияние pagoda-upstream успешно${NC}"
else
    echo -e "${RED}❌ Конфликты при слиянии pagoda-upstream. Разрешите вручную.${NC}"
    echo -e "${YELLOW}💡 После разрешения конфликтов выполните:${NC}"
    echo -e "   git add ."
    echo -e "   git commit"
    echo -e "   git checkout ${CURRENT_BRANCH}"
    echo -e "   git merge pagoda-upstream"
    exit 1
fi

# Вернуться на рабочую ветку
echo -e "${GREEN}🔄 Возврат на ветку ${CURRENT_BRANCH}...${NC}"
git checkout "${CURRENT_BRANCH}"

# Слить обновления в рабочую ветку
echo -e "${GREEN}🔀 Слияние обновлений в рабочую ветку...${NC}"
if git merge pagoda-upstream; then
    echo -e "${GREEN}✅ Синхронизация завершена успешно!${NC}"
    echo -e "${GREEN}📊 Статистика изменений:${NC}"
    git log --oneline pagoda-upstream^..HEAD | head -10
else
    echo -e "${RED}❌ Конфликты при слиянии в рабочую ветку.${NC}"
    echo -e "${YELLOW}💡 Разрешите конфликты и выполните:${NC}"
    echo -e "   git add ."
    echo -e "   git commit"
    exit 1
fi

echo -e "${GREEN}✨ Готово!${NC}"

