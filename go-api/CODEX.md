# Codebase Index

## Overview
Go API с Clean Architecture (Handlers → Services → Repositories → Database)

## Tech Stack
- **Language:** Go 1.26.1
- **Framework:** Gin v1.12.0
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Queue:** NATS 2.10
- **Logging:** zerolog

## Directory Structure
```
cmd/api/main.go          # Entry point
internal/
├── config/db/           # Database setup
├── handlers/            # HTTP handlers (auth, payments, resumes, admin)
├── helpers/             # Utilities (Respond, GetUserID, RandStr)
├── logger/              # Centralized logging (zerolog wrapper)
├── middleware/          # Auth middleware (Auth, RequireAdmin)
├── repository/          # Data access layer
├── routes/             # Route definitions
├── service/            # Business logic
└── types/              # Data types (internal/external)
```

## API Routes

### Public
| Method | Endpoint | Handler |
|--------|----------|---------|
| POST | `/auth/getToken` | GetToken |

### Protected (Bearer Token)
| Method | Endpoint | Handler |
|--------|----------|---------|
| POST | `/api/payments` | CreatePayment |
| PATCH | `/api/payments/:id` | UpdatePayment |
| GET | `/api/payments/me` | GetMyPayments |
| POST | `/api/resumes` | AddResume |
| PATCH | `/api/resumes/:id` | UpdateResume |
| DELETE | `/api/resumes/:id` | DeleteResume |
| GET | `/api/resumes/me` | GetMyResumes |
| GET | `/api/resumes/:id` | GetResumeByID |
| POST | `/api/vacancies/match` | MatchVacancies |
| GET | `/api/vacancies/matches/me` | GetUserMatches |
| GET | `/api/vacancies/matches/:id` | GetMatchResults |
| GET | `/api/user/stats` | GetMyStats |
| GET | `/api/user/payment-status` | GetMyPaymentStatus |

### AI Webhooks
| Method | Endpoint | Handler |
|--------|----------|---------|
| POST | `/ai/webhook/analyze` | WebhookAnalyzeResume |
| POST | `/ai/webhook/matches` | WebhookVacancyMatches |

### Admin (Bearer Token + is_admin=1)
| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/api/admin/users` | GetUsers | Список всех пользователей |
| GET | `/api/admin/stats` | GetStats | Статистика (все или ?user_id=N) |
| GET | `/api/admin/users/:id/resumes` | GetUserResumes | Резюме конкретного юзера |
| GET | `/api/admin/users/:id/payments` | GetUserPayments | Платежи конкретного юзера |

**Response Stats:**
```json
{
  "user_id": 1,
  "username": "Egor",
  "resumes": 3,
  "payments": 5,
  "total_amount": 15000
}
```

## Database Schema
- **users:** id, username (UNIQUE), telegram_id (UNIQUE), is_admin (BOOLEAN, default 0)
- **tokens:** id, user_id (FK), token (indexed), created_at
- **payments:** id, user_id (FK), amount, status, provider, created_at
- **resumes:** id, user_id (FK), title, content, tags (JSON), score (REAL), created_at
- **vacancy_matches:** id, user_id (FK), resume_id (FK), query, status (pending/completed), created_at
- **vacancy_match_results:** id, match_id (FK), vacancy_title, company, score, url, salary, excerpt

## Middleware
- `middleware.Auth()` — проверка Bearer токена
- `middleware.RequireAdmin()` — проверка is_admin=1

## Key Patterns
- Auth: Token in `Authorization: Bearer <token>` header
- User ID from context: `helpers.GetUserID(c)`
- Consistent responses: `helpers.Respond(c, status, "error_type", "message")`
- Logging: `logger.Info()/Error()/Debug()` via `internal/logger`
- Max 5 tokens per user
- Auto user creation on first token request

## Logger (`internal/logger`)
```go
logger.Init()           // Initialize (call in main)
logger.Info()           // Info level
logger.Error()          // Error level
logger.Debug()          // Debug level
logger.Warn()           // Warn level
logger.LogRequest(c)()  // HTTP request logging with user_id
```

## Server Config
- HTTP: `:8080`
- PProf: `localhost:6060`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Redis: `localhost:6379` (env: `REDIS_ADDR`)
- NATS: `localhost:4222` (env: `NATS_URL`)

## Database Config (PostgreSQL)

### Connection via PgBouncer
```
DB_HOST=pgbouncer  (not postgres directly)
DB_PORT=5432
```

### Configurable via env
```env
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME_MIN=5
```

### Read Replica (optional)
```env
DB_REPLICA_HOST=postgres-replica
```

### PgBouncer Settings (docker-compose)
- Pool Mode: `transaction`
- Max Client Conn: `1000`
- Default Pool Size: `25`
- Min Pool Size: `5`

### PostgreSQL Tuning
- max_connections: 200
- shared_buffers: 256MB
- effective_cache_size: 1GB
- work_mem: 4MB

## Optimizations

### Rate Limiting
- `/auth/*` — 10 req/min per IP
- `/api/*` — 100 req/min per IP

### Caching (Redis)
| Data | TTL | Key Pattern |
|------|-----|-------------|
| Token validation | 15 min | `token:{token}` |
| User stats | 5 min | `user:stats:{id}` |
| Payment status | 5 min | `user:payment:{id}` |
| Admin stats | 2-3 min | `admin:stats:*` |
| Admin users | 5 min | `admin:users` |

### SQL Optimizations
- Combined queries (N+1 fix) in GetUserStats
- Batch inserts in SaveVacancyMatchResults
- Proper indexes on all foreign keys

## Cache (Redis)
| Endpoint | TTL | Key Pattern |
|----------|-----|-------------|
| GET /api/admin/stats | 2 min | `admin:stats:all` |
| GET /api/admin/stats?user_id=N | 3 min | `admin:stats:user:{id}` |
| GET /api/admin/users | 5 min | `admin:users` |
| GET /api/admin/users/:id/resumes | 5 min | `admin:resumes:{id}` |
| GET /api/admin/users/:id/payments | 5 min | `admin:payments:{id}` |

**Invalidation:** `admin.InvalidateUserCache(userID)` after POST/PATCH

## AI Integration (NATS)

### Queues

| Queue | Trigger | Payload |
|-------|---------|---------|
| `resume.analyze` | POST /api/resumes | `{resume_id, user_id, title}` |
| `vacancy.match` | POST /api/vacancies/match | `{match_id, job: {resume_id, user_id, query, limit, exclude_words, employment_types, work_formats}}` |

### Vacancy Matching Filters

| Field | Type | Description |
|-------|------|-------------|
| `exclude_words` | `[]string` | Исключить вакансии с этими словами |
| `employment_types` | `[]string` | Тип оформления: `"трудоустройство"`, `"самозанятость"`, `"ИП"`, `"договор ГПХ"` |
| `work_formats` | `[]string` | Формат работы: `"удалёнка"`, `"офис"`, `"гибрид"` |

### Webhooks

| Endpoint | Purpose |
|----------|---------|
| `POST /ai/webhook/analyze` | AI returns analyzed resume |
| `POST /ai/webhook/matches` | AI returns vacancy matches |

## AI Service (separate repo)

**Endpoints used:**
- Subscribe: `resume.analyze` (NATS)
- Call: `POST http://localhost:8080/ai/webhook/analyze`

## Telegram Bot

### Commands
- `/start` — Приветствие
- `/menu` — Главное меню
- `/resumes` — Управление резюме
- `/vacancies` — Поиск вакансий
- `/schedule` — Расписание поиска
- `/stats` — Статистика
- `/payment` — Оплата

### Bot API Endpoints
| Endpoint | Purpose |
|----------|---------|
| `GET /api/user/stats` | Статистика пользователя |
| `GET /api/user/payment-status` | Статус подписки |
| `DELETE /api/resumes/:id` | Удаление резюме |

### User Response Examples

**Stats Response:**
```json
{
  "user_id": 1,
  "resumes_count": 3,
  "searches_count": 12,
  "matches_count": 47
}
```

**Payment Status Response:**
```json
{
  "is_active": true,
  "expires_at": "2026-04-15",
  "days_remaining": 28
}
```

## Docker Compose

```bash
# Basic (dev)
docker-compose up -d

# With scaling (replicas)
docker-compose --profile scaling up -d
```

## Build & Dev
```bash
# Generate swagger
swag init -g cmd/api/main.go

# Run
go run cmd/api/main.go
```
