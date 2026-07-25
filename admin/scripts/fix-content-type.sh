#!/bin/bash

# Скрипт для исправления content-type файлов в бакете

set -e

BUCKET_NAME="${YC_BUCKET_NAME:-workouts-admin}"

echo "🔧 Исправляем content-type файлов..."

# Переходим в папку admin (скрипт запускается из scripts/)
cd "$(dirname "$0")/.."

if [ ! -d "dist" ]; then
    echo "❌ Папка dist не найдена. Сначала выполните сборку: npm run build"
    exit 1
fi

cd dist

# Обновляем метаданные для каждого файла
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

    echo "Обновляем: ${KEY} -> ${CONTENT_TYPE}"

    # Копируем объект сам на себя с новыми метаданными
    yc storage s3api copy-object \
        --bucket ${BUCKET_NAME} \
        --copy-source "${BUCKET_NAME}/${KEY}" \
        --key "${KEY}" \
        --metadata-directive REPLACE \
        --content-type "${CONTENT_TYPE}"
done

cd ..

echo ""
echo "✅ Content-type исправлен для всех файлов!"

