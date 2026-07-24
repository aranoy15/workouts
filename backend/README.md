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

Перед первым запуском задайте в `.env` обязательный `JWT_SECRET` и `ADMIN_PASSWORD`
(если таблица пользователей пуста, при старте создаётся начальный admin).

Проверка:

```bash
curl http://localhost:8080/api/health
```

Логин:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"change-me-admin-password"}'
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
| `make help` | Список всех команд |

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
├── Makefile
└── go.mod
```

## Wire

После изменения провайдеров в `cmd/wire.go`:

```bash
make wire
```
