# Workouts Backend

REST API для приложения Workouts на Go (Gin).

## Требования

- Go 1.24 или выше
- PostgreSQL
- Утилита миграций в `../migrate` (для `make migrate-*`)

## Быстрый старт

```bash
cd backend
cp .env.example .env
make deps
make wire
```

Поднять PostgreSQL локально:

```bash
docker compose -f docker/docker-compose.yml up postgres -d
make migrate-up
make run-env
```

Перед первым запуском задайте в `.env` обязательный `JWT_SECRET`.
Если в БД нет активного admin, любой успешный `POST /api/auth/login` создаёт (или повышает) admin.

Проверка:

```bash
curl http://localhost:8080/api/health
```

Логин:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"your-password"}'
```

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make run` | Запуск сервера |
| `make run-env` | Запуск с переменными из `.env` |
| `make build` | Сборка бинарника |
| `make deps` | Установка зависимостей |
| `make wire` | Генерация Wire DI |
| `make test` | Запуск тестов |
| `make migrate-up` | Применить миграции |
| `make docker-login` | Логин в Yandex Container Registry |
| `make docker-build` | Сборка Docker-образа (linux/amd64) |
| `make docker-push` | Push образа в `cr.yandex` |
| `make deploy-revision` | Новая ревизия Serverless Container |
| `make deploy` | Полный деплой: build + push + revision |
| `make help` | Список всех команд |

## Деплой (Yandex Cloud)

По паттерну choker_market: образ в Container Registry, приложение в Serverless Containers.

```bash
yc init
make docker-login
make deploy
# или по шагам:
# make docker-push
# make deploy-revision
```

Опции: `REGISTRY_ID`, `CONTAINER_NAME` (default `workouts-backend`), `SERVICE_ACCOUNT_NAME` (default `workouts-backend-admin`), `SERVICE_ACCOUNT_ID`.

Сохранить конфиг текущей ревизии:

```bash
make save-container-config
```

## API

Базовый путь: `/api`. OpenAPI: [api/schema.yaml](api/schema.yaml).

| Method | Path | Описание |
|--------|------|----------|
| GET | `/api/health` | Health check (проверяет доступность БД) |
| POST | `/api/auth/login` | Логин (JWT) |
| GET/POST/PUT | `/api/users` | Управление пользователями (только admin) |

## Структура проекта

```
backend/
├── api/schema.yaml
├── cmd/                  # main, wire
├── migrations/
├── src/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── router/
├── docker/
├── scripts/              # deploy.sh, deploy-revision.sh
├── Makefile
└── go.mod
```

## Wire

После изменения провайдеров в `cmd/wire.go`:

```bash
make wire
```
