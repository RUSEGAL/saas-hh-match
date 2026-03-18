# ТЗ: Интеграционное тестирование трёх сервисов

## 1. Overview

Проверить корректную работу всей системы: API + AI Service + Telegram Bot.

## 2. Компоненты

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│ Telegram    │      │   API       │      │ AI Service  │
│   Bot       │ ───▶│             │ ───▶│             │
│             │      │             │ ←───│             │
└─────────────┘      └──────┬──────┘      └─────────────┘
                             │
                             ▼
                       ┌─────────────┐
                       │   NATS      │
                       └─────────────┘
```

## 3. Сценарии тестирования

### 3.1 Аутентификация

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: `/start` | Запрос username |
| 2 | Bot: отправить username | Создание user в API, генерация token |
| 3 | API: `POST /auth/getToken` | Возврат token |
| 4 | Bot: сохранить token | Token доступен для запросов |

### 3.2 Резюме (полный цикл)

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: `/resumes` → Создать | Запрос названия |
| 2 | Bot: ввести название | Запрос текста |
| 3 | Bot: ввести текст | Отправка в API |
| 4 | API: `POST /api/resumes` | Создание в БД, publish в NATS |
| 5 | AI Service: получить job | Обработка resume.analyze |
| 6 | AI Service: `POST /ai/webhook/analyze` | Обновление resume с tags, score |
| 7 | Bot: уведомление | "Резюме проанализировано! Score: 85%" |

### 3.3 Поиск вакансий (полный цикл)

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: `/vacancies` → Начать поиск | Список резюме для выбора |
| 2 | Bot: выбрать резюме | Запрос поискового запроса |
| 3 | Bot: ввести "golang developer" | Запрос фильтров (опционально) |
| 4 | Bot: пропустить/выбрать фильтры | Отправка в API |
| 5 | API: `POST /api/vacancies/match` | Создание job, publish vacancy.match |
| 6 | AI Service: получить job | Вызов hh.ru API |
| 7 | AI Service: семантическое сравнение | Фильтрация >70% match |
| 8 | AI Service: `POST /ai/webhook/matches` | Сохранение результатов |
| 9 | Bot: уведомление | "Найдено N вакансий" |
| 10 | Bot: показать результаты | Inline-кнопки с вакансиями |

### 3.4 Подписка

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: `/payment` | Проверка статуса через API |
| 2 | API: `GET /api/user/payment-status` | `is_active: false` |
| 3 | Bot: Оплатить | Генерация ссылки на оплату |
| 4 | Оплата через ЮKassa | Webhook в API |
| 5 | API: `POST /api/payments` | Создание платежа |
| 6 | Bot: уведомление | "Подписка активирована на 30 дней" |
| 7 | Bot: `/payment` повторно | `is_active: true`, days_remaining: 30 |

### 3.5 Ограничение без подписки

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: истекла подписка | `days_remaining: 0` |
| 2 | Bot: любые команды (кроме /payment) | "❌ Подписка неактивна. /payment" |

### 3.6 Статистика

| # | Шаг | Ожидаемый результат |
|---|-----|---------------------|
| 1 | Bot: `/stats` | Запрос в API |
| 2 | API: `GET /api/user/stats` | Корректные данные |
| 3 | Bot: отобразить | Резюме, поиски, вакансии, платежи |

## 4. Тестовые данные

### Пользователи
```json
{
  "username": "test_user_1",
  "telegram_id": 123456789
}
```

### Резюме
```json
{
  "title": "Senior Go Developer",
  "content": "5 years experience with Go, PostgreSQL, Redis, microservices..."
}
```

### Ожидаемые tags от AI
`["golang", "postgresql", "redis", "microservices", "docker", "kubernetes"]`

### Ожидаемый score
`0.7 - 1.0`

## 5. NATS Queues

### Подписки AI Service
```
resume.analyze     → AI обработка резюме
vacancy.match     → AI поиск вакансий
```

### JetStream streams
- `ANALYSIS` — для resume.analyze
- `VACANCY` — для vacancy.match

## 6. API Endpoints для проверки

```
POST /auth/getToken
GET  /api/user/stats
GET  /api/user/payment-status
POST /api/resumes
GET  /api/resumes/me
POST /api/vacancies/match
GET  /api/vacancies/matches/:id
POST /api/payments
POST /ai/webhook/analyze
POST /ai/webhook/matches
```

## 7. Проверки

### HTTP коды
- [ ] 200 для успешных операций
- [ ] 201 для создания ресурсов
- [ ] 401 для неавторизованных
- [ ] 403 для чужих данных

### Данные в БД
- [ ] User создан с telegram_id
- [ ] Resume содержит tags и score после AI
- [ ] VacancyMatchResults сохраняются
- [ ] Payment имеет статус completed

### NATS
- [ ] Jobs публикуются в правильные queues
- [ ] Jobs не теряются (JetStream)
- [ ] Consumer получает все jobs

### Redis
- [ ] Token cache работает
- [ ] Rate limiting срабатывает
- [ ] Cache инвалидируется

## 8. Стек для тестов

```yaml
services:
  api:
    build: ./go-api
    ports: ["8080:8080"]
    
  ai-service:
    build: ./ai-service
    ports: ["8081:8081"]
    
  telegram-bot:
    build: ./tg-bot
    environment:
      - BOT_TOKEN=test_token
      - API_URL=http://api:8080
      
  nats:
    image: nats:2.10-alpine
    command: ["-js"]  # JetStream enabled
    
  postgres:
    image: postgres:16-alpine
    
  redis:
    image: redis:7-alpine
```

## 9. Test Runner

Использовать `testcontainers-go` или ручной запуск через docker-compose.

## 10. Acceptance Criteria

- [ ] Все 3 сервиса стартуют без ошибок
- [ ] Telegram Bot получает ответы от API
- [ ] AI Service обрабатывает NATS jobs
- [ ] Webhooks от AI приходят в API
- [ ] Данные сохраняются корректно
- [ ] Подписочная модель работает
- [ ] Rate limiting не блокирует нормальные запросы
