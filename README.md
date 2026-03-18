# Resume Bot

Telegram бот для поиска работы с AI-анализом резюме.

## Микросервисная архитектура

```
┌─────────────┐    ┌─────────┐    ┌────────────┐
│  tg-bot     │───▶│  go-api │───▶│ ai-service│
│  (Telegram) │    │   API   │    │   (AI)    │
└─────────────┘    └────┬────┘    └────────────┘
                      │
                ┌─────┴─────┐
                │PostgreSQL │
                │   Redis   │
                │   NATS    │
                └───────────┘
```

## Быстрый старт

### Development

```bash
# 1. Настройка переменных окружения
cp .env.example .env
# Заполните BOT_TOKEN, AI_API_KEY

# 2. Запуск с hot reload
make dev

# Или вручную:
docker-compose -f docker-compose.dev.yml up -d
```

### Production

```bash
# 1. Настройка
cp .env.example .env
# Заполните все переменные

# 2. Production запуск
make up

# Проверить статус
make ps
make logs
```

## Сервисы

| Сервис | Описание | Порт |
|--------|----------|------|
| tg-bot | Telegram бот | - |
| go-api | REST API | 8080 |
| ai-service | AI анализ резюме | - |
| nginx | Load balancer | 80 |
| postgres | База данных | 5432 |
| redis | Кэш | 6379 |
| nats | Message queue | 4222 |

## Команды Makefile

```bash
make dev          # Development с hot reload
make dev-down     # Остановить dev
make up           # Production запуск
make down         # Остановить production
make logs         # Логи всех сервисов
make logs-api     # Логи API
make test         # Тесты
make lint         # Линтер
make build        # Собрать образы
```

## CI/CD

| Workflow | Описание |
|----------|----------|
| CI | Тесты, lint, Docker build |
| Build | Сборка и пуш образов |
| Deploy | Деплой на сервер |
| Security | Проверка уязвимостей |
| Backup | Ежедневный бэкап БД |

## Переменные окружения

```env
# Обязательные
BOT_TOKEN=           # Telegram bot token
AI_API_KEY=          # DeepSeek API key
DB_PASSWORD=         # PostgreSQL password
JWT_SECRET=          # JWT signing secret

# Опциональные
RESUME_WORKERS=4     # AI workers
VACANCY_WORKERS=2    # Vacancy workers
```

## Структура проекта

```
.
├── go-api/         # REST API (Go + Gin)
├── ai-service/     # AI анализ (Go)
├── tg-bot/         # Telegram бот (Go)
├── nginx/          # Load balancer
├── docker-compose.yml     # Production
└── docker-compose.dev.yml # Development
```

## Лицензия

MIT
