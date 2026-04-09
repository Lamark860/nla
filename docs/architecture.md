## Архитектура NLA

### Обзор

NLA — анализатор облигаций MOEX. Миграция проекта ASH (Yii2/PHP) на Go-стек с улучшенной архитектурой.

### Стек

| Слой | Технология | Роль |
|------|-----------|------|
| API | Go 1.25 + Chi | REST API, WebSocket (будущее) |
| Auth | JWT (HS256) + bcrypt | Аутентификация |
| Relational DB | PostgreSQL 16 | Пользователи, избранное, портфели |
| Document DB | MongoDB 7 | Облигации, анализы, чаты, кэши |
| Cache / Queue | Redis 7 | Кэш MOEX, очередь задач, pub/sub |
| Proxy | Nginx | Reverse proxy, статика, WebSocket |
| Парсеры | Python 3 + Selenium | Parsing dohod.ru, эмитенты |

### Архитектурный паттерн

```
HTTP Request
  │
  ▼
Handler (internal/handler/)         ← HTTP: парсинг запроса, формирование ответа
  │
  ▼
Service (internal/service/)         ← Бизнес-логика, оркестрация, валидация
  │
  ├──▶ Repository (internal/repository/)  ← SQL (PostgreSQL) / MongoDB запросы
  ├──▶ Client (internal/client/)          ← Внешние API (MOEX ISS, OpenAI)
  └──▶ Queue (internal/queue/)            ← Фоновые задачи через Redis
```

Каждый слой зависит только от слоя ниже. Handler НЕ обращается к Repository напрямую.

### Структура проекта

```
nla/
├── cmd/
│   └── api/main.go                 # Точка входа, DI, graceful shutdown
├── internal/
│   ├── config/                     # Конфигурация из ENV
│   ├── database/                   # Коннекторы: postgres, mongo, redis
│   ├── model/                      # Структуры данных (доменные модели)
│   ├── repository/                 # PostgreSQL запросы (pgx)
│   ├── mongo/                      # MongoDB операции (коллекции)
│   ├── client/                     # Внешние HTTP-клиенты
│   │   ├── moex/                   # MOEX ISS API
│   │   └── openai/                 # OpenAI API
│   ├── service/                    # Бизнес-логика
│   │   ├── auth.go
│   │   ├── bond.go
│   │   ├── analysis.go
│   │   ├── chat.go
│   │   ├── issuer.go
│   │   ├── parser.go
│   │   └── queue.go
│   ├── handler/                    # HTTP handlers
│   ├── middleware/                  # JWT, logging, recovery
│   ├── queue/                      # Job definitions + worker
│   │   ├── worker.go
│   │   ├── ai_analysis_job.go
│   │   ├── parse_bond_job.go
│   │   ├── parse_emitter_job.go
│   │   └── sync_issuer_job.go
│   └── router/                     # Маршрутизация
├── data/
│   └── prompts/                    # Системные промпты агентов
├── python/                         # Selenium-парсеры
├── migrations/                     # SQL-миграции PostgreSQL
├── docs/                           # Документация
├── nginx/                          # Конфиг Nginx
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

### Модули и их ответственность

#### Bond Service
- Загрузка облигаций из MOEX ISS API
- Пересчёт цен (% → рубли)
- Сортировка, фильтрация, пагинация
- Кэширование в Redis (24 часа)
- Группировка по эмитентам

#### AI Analysis Service
- Подготовка данных облигации для OpenAI
- Отправка на анализ (через очередь)
- Парсинг рейтинга из ответа AI
- CRUD анализов в MongoDB

#### Chat Service
- CRUD сессий и сообщений в MongoDB
- Загрузка системного промпта по типу агента
- Формирование контекста (последние N сообщений)
- Вызов OpenAI для генерации ответа

#### Issuer Service
- Маппинг SECID ↔ emitter_id через MOEX disclosure API
- Кэширование данных эмитента (30 дней TTL)
- Кэширование данных облигации с dohod.ru (30 дней TTL)

#### Queue Service
- Трекинг статусов задач в MongoDB
- Дедупликация (не создавать повторную задачу)
- Типы задач: ai_analysis, parse_bond, parse_emitter, sync_issuer
- Статусы: pending → running → done | error

#### Parser Service
- Вызов Python-скриптов (dohod-parser.py, emit-parser.py)
- Парсинг JSON-вывода
- Сохранение результатов через IssuerService

### Потоки данных

#### AI-анализ облигации (async)

```
POST /api/v1/bonds/{secid}/analyze
  │
  ▼
BondHandler.Analyze()
  ├── QueueService.CreateJob("ai_analysis", secid)
  └── Redis queue ← push AiAnalysisJob
  │
  ▼ Return {job_id, status: "pending"}

Worker (goroutine):
  ├── QueueService.MarkRunning(jobId)
  ├── BondService.GetBondFullDetails(secid)
  ├── OpenAI client → AI response
  ├── Parse rating [RATING:XX]
  ├── AnalysisService.Save(secid, response, rating)
  └── QueueService.MarkDone(jobId, result)

Frontend polling:
  GET /api/v1/jobs/{id} → {status: "done", result: {...}}
```

#### Чат (sync → будет async через WebSocket)

```
POST /api/v1/chat/{sessionId}/messages
  │
  ▼
ChatHandler.SendMessage()
  ├── ChatService.AddMessage(sessionId, "user", content)
  ├── ChatService.GetContextMessages(sessionId, 20)
  ├── PromptService.GetAgentPrompt(agentType)
  ├── OpenAI client → AI response
  ├── ChatService.AddMessage(sessionId, "assistant", response)
  └── Return {userMessage, assistantMessage}
```

### Отличия от ASH (Yii2)

| Аспект | ASH (старый) | NLA (новый) |
|--------|-------------|-------------|
| Очередь | Файловая (runtime/queue/) | Redis (asynq или BullMQ-подобная) |
| MOEX кэш | FileCache (24ч) | Redis (24ч, атомарные операции) |
| Параллелизм | Нет (PHP блокирующий) | Go goroutines (параллельные API-вызовы) |
| Auth | Нет (гостевой доступ) | JWT + PostgreSQL users |
| API | Смешанный HTML/JSON | Чистый REST JSON API |
| Worker | Один процесс, умирает | Встроенный в API или отдельный, устойчивый |
| Конфиг | Хардкод в params.php | ENV переменные |
| БД | Только MongoDB | PostgreSQL (реляционное) + MongoDB (документы) |
