# Production Deployment

## Quick Start

1. Copy environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your credentials:
   - `BOT_TOKEN` - Telegram bot token
   - `AI_API_KEY` - DeepSeek API key
   - `DB_PASSWORD` - Database password
   - `JWT_SECRET` - Random string for JWT

3. Start services:
```bash
docker-compose up -d
```

4. Check status:
```bash
docker-compose ps
docker-compose logs -f api
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| nginx | 80 | Load balancer & reverse proxy |
| api | 8080 | Go API (3 replicas) |
| tg-bot | - | Telegram bot |
| ai-service | - | AI analysis service |
| postgres | 5432 | Database |
| redis | 6379 | Cache |
| nats | 4222 | Message queue |

## Scaling

```bash
# Scale API
docker-compose up -d --scale api=10

# Scale bot (for high load)
docker-compose up -d --scale tg-bot=3

# Scale AI workers
RESUME_WORKERS=8 VACANCY_WORKERS=4 docker-compose up -d --build ai-service
```

## Monitoring

```bash
# API health
curl http://localhost/health

# NATS monitoring
curl http://localhost:8222/healthz

# Redis
redis-cli -h localhost ping
```

## Stopping

```bash
docker-compose down

# With volumes (clean start)
docker-compose down -v
```

## Production Checklist

- [ ] Change all passwords in `.env`
- [ ] Set `JWT_SECRET` to random 64-char string
- [ ] Configure SSL certificates in `nginx/ssl/`
- [ ] Set proper `WEBHOOK_URL` for bot
- [ ] Set resource limits for production server
- [ ] Setup monitoring (Prometheus, Grafana)
- [ ] Setup backups for PostgreSQL
