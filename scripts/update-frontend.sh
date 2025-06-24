#!/bin/bash

# Скрипт для обновления frontend из репозитория

set -e  # Останавливаться при ошибках

echo "Начинаю обновление frontend..."

# Удаляем старый бекап если существует
if [ -d "frontend.bak" ]; then
    echo "Удаляю старый бекап frontend.bak..."
    rm -rf frontend.bak
fi

# Переименовываем текущий frontend в frontend.bak
if [ -d "frontend" ]; then
    echo "Создаю бекап текущего frontend..."
    mv frontend frontend.bak
fi

# Клонируем весь репозиторий из ветки stable
echo "Загружаю репозиторий из ветки 'stable'..."
git clone --depth 1 --branch stable --single-branch git@github.com:Forlyongame/forly-shard_legends.git temp_repo

# Создаем каталог frontend и копируем все файлы из корня репозитория
if [ -d "temp_repo" ]; then
    # Удаляем .git из временного репозитория
    rm -rf temp_repo/.git
    # Перемещаем все содержимое в каталог frontend
    mv temp_repo frontend
    echo "Frontend успешно обновлен!"
    
    # Восстанавливаем node_modules из бекапа если есть
    if [ -d "frontend.bak/node_modules" ]; then
        echo "Восстанавливаю node_modules из бекапа..."
        mv frontend.bak/node_modules frontend/
        echo "node_modules восстановлен"
    fi
    
    # Устанавливаем зависимости
    echo "Устанавливаю зависимости..."
    cd frontend && npm install
    
    # Собираем проект
    echo "Собираю проект..."
    npm run build
    
    cd ..
    echo "Установка и сборка завершены!"
    
else
    echo "Ошибка: не удалось склонировать репозиторий"
    # Восстанавливаем из бекапа если что-то пошло не так
    if [ -d "frontend.bak" ]; then
        mv frontend.bak frontend
        echo "Восстановлен из бекапа"
    fi
    rm -rf temp_repo
    exit 1
fi

# Удаляем временный репозиторий
rm -rf temp_repo

echo "Готово!"