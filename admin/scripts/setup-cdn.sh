#!/bin/bash

# Скрипт для настройки CDN для Object Storage

set -e

BUCKET_NAME="${YC_BUCKET_NAME:-workouts-admin}"
ORIGIN_DOMAIN="${BUCKET_NAME}.website.yandexcloud.net"
FOLDER_ID="${YC_FOLDER_ID:-}"

echo "🌐 Настраиваем CDN..."

# Проверка наличия YC CLI
if ! command -v yc &> /dev/null; then
    echo "❌ YC CLI не установлен"
    exit 1
fi

# Получаем folder_id если не задан
if [ -z "$FOLDER_ID" ]; then
    FOLDER_ID=$(yc config get folder-id)
fi

# Создаем CDN ресурс
echo "📦 Создаем CDN ресурс..."
# CNAME должен быть FQDN, используем формат с доменом
CNAME_FQDN="${BUCKET_NAME}-cdn.cdn.yandexcloud.net"
RESOURCE_ID=$(yc cdn resource create \
    "${CNAME_FQDN}" \
    --origin-bucket-source="${ORIGIN_DOMAIN}" \
    --origin-bucket-name="${BUCKET_NAME}" \
    --origin-protocol HTTPS \
    --folder-id="${FOLDER_ID}" \
    --format json | jq -r '.id')

if [ -z "$RESOURCE_ID" ] || [ "$RESOURCE_ID" = "null" ]; then
    echo "❌ Не удалось создать CDN ресурс"
    exit 1
fi

echo "✅ CDN ресурс создан: ${RESOURCE_ID}"

# Получаем CNAME
CNAME_VALUE=$(yc cdn resource get ${RESOURCE_ID} --format json | jq -r '.cname')

echo ""
echo "✅ CDN настроен!"
echo "📍 CNAME для настройки DNS: ${CNAME_VALUE}"
echo "📍 Resource ID: ${RESOURCE_ID}"
echo ""
echo "📝 Для настройки домена:"
echo "1. Добавьте CNAME запись в DNS: ${CNAME_VALUE}"
echo "2. Обновите CDN ресурс с доменом:"
echo "   yc cdn resource update ${RESOURCE_ID} --secondary-hostnames <your-domain.com>"
echo "3. Настройте SSL сертификат через Certificate Manager"

