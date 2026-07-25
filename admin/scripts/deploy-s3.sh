#!/bin/bash

# Скрипт для деплоя admin панели в Yandex Cloud Object Storage + CDN

set -e

BUCKET_NAME="${YC_BUCKET_NAME:-workouts-admin}"
FOLDER_ID="${YC_FOLDER_ID:-}"

echo "🚀 Начинаем деплой в Object Storage..."

# Проверка наличия YC CLI
if ! command -v yc &> /dev/null; then
    echo "❌ YC CLI не установлен. Установите: https://cloud.yandex.ru/docs/cli/quickstart"
    exit 1
fi

# Проверка авторизации
if ! yc config list &> /dev/null; then
    echo "❌ Не авторизованы в YC CLI. Выполните: yc init"
    exit 1
fi

# Получаем folder_id если не задан
if [ -z "$FOLDER_ID" ]; then
    echo "📁 Получаем folder_id..."
    FOLDER_ID=$(yc config get folder-id)
    if [ -z "$FOLDER_ID" ]; then
        echo "❌ folder-id не найден. Установите: yc config set folder-id <folder-id>"
        exit 1
    fi
fi

# Переходим в папку admin (скрипт запускается из scripts/)
cd "$(dirname "$0")/.."

# Проверка и настройка переменных окружения для сборки
echo "🔧 Проверяем переменные окружения..."
if [ -z "$VITE_API_URL" ]; then
    echo "⚠️  VITE_API_URL не задана. Используется значение из .env файла (если есть)"
    echo "   Для production задайте: export VITE_API_URL=https://your-backend.com/api"
else
    echo "✅ VITE_API_URL установлена: $VITE_API_URL"
    export VITE_API_URL
fi

# Сборка проекта
echo "🔨 Собираем проект с переменными окружения..."
npm run build

if [ ! -d "dist" ]; then
    echo "❌ Папка dist не найдена. Сборка не удалась."
    exit 1
fi

# Убираем crossorigin для совместимости с Safari
echo "🔧 Убираем crossorigin атрибуты для Safari..."
if [ -f "dist/index.html" ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' 's/ crossorigin//g' dist/index.html
    else
        # Linux
        sed -i 's/ crossorigin//g' dist/index.html
    fi
fi

# Проверяем существование бакета
echo "🔍 Проверяем существование бакета..."
EXISTING=$(yc storage bucket get --name ${BUCKET_NAME} 2>/dev/null || echo "")

if [ -z "$EXISTING" ]; then
    echo "📦 Создаем новый бакет..."
    yc storage bucket create \
        --name ${BUCKET_NAME} \
        --folder-id ${FOLDER_ID} \
        --max-size 10737418240 \
        --public-read

    echo "✅ Бакет создан: ${BUCKET_NAME}"
else
    echo "✅ Бакет уже существует: ${BUCKET_NAME}"
fi

# Настраиваем статический хостинг
echo "🌐 Настраиваем статический хостинг..."
yc storage bucket update \
    --name ${BUCKET_NAME} \
    --website-settings '{"index":"index.html","error":"index.html"}'

# Загружаем файлы
echo "📤 Загружаем файлы в бакет..."
cd dist

# Загружаем файлы с правильными content-type через s3api
for file in $(find . -type f); do
    # Определяем content-type по расширению
    case "${file##*.}" in
        html)
            CONTENT_TYPE="text/html; charset=utf-8"
            ;;
        css)
            CONTENT_TYPE="text/css"
            ;;
        js)
            CONTENT_TYPE="application/javascript"
            ;;
        json)
            CONTENT_TYPE="application/json; charset=utf-8"
            ;;
        png)
            CONTENT_TYPE="image/png"
            ;;
        jpg|jpeg)
            CONTENT_TYPE="image/jpeg"
            ;;
        svg)
            CONTENT_TYPE="image/svg+xml"
            ;;
        ico)
            CONTENT_TYPE="image/x-icon"
            ;;
        woff|woff2)
            CONTENT_TYPE="font/woff2"
            ;;
        *)
            CONTENT_TYPE="application/octet-stream"
            ;;
    esac

    KEY="${file#./}"

    # Загружаем файл через s3api с правильным content-type
    yc storage s3api put-object \
        --bucket ${BUCKET_NAME} \
        --key "${KEY}" \
        --body "${file}" \
        --content-type "${CONTENT_TYPE}"
done

cd ..

# Получаем URL бакета
echo "🌐 Получаем URL бакета..."
BUCKET_URL="https://${BUCKET_NAME}.website.yandexcloud.net"

echo ""
echo "✅ Деплой завершен!"
echo "📍 URL сайта: ${BUCKET_URL}"
echo ""
echo "📝 Следующие шаги:"
echo "1. Настройте CDN для ускорения:"
echo "   ./scripts/setup-cdn.sh"
echo ""
echo "2. Настройте домен (если есть):"
echo "   yc cdn resource update <resource-id> --secondary-hostnames <your-domain.com>"
echo ""
echo "3. Настройте SSL сертификат для домена через Certificate Manager"

