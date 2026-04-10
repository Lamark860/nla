## Сущности и хранилище данных

### Обзор хранилищ

```
PostgreSQL (реляционные данные)     MongoDB (документы и кэши)         Redis (ephemeral)
┌────────────────────────────┐     ┌──────────────────────────────┐   ┌────────────────────┐
│ users                      │     │ bond_analyses                │   │ bonds:list (24h)   │
│ favorites                  │     │ bond_issuers                 │   │ bonds:{secid} (24h)│
│                            │     │ dohod_details (TTL 30d)      │   │ jobs:queue          │
│                            │     │ issuer_ratings               │   │ sessions (JWT)      │
│                            │     │ queue_jobs                   │   │                    │
│                            │     │ chat_sessions                │   │                    │
│                            │     │ chat_messages                │   │                    │
└────────────────────────────┘     └──────────────────────────────┘   └────────────────────┘
```

---

### PostgreSQL

#### users

| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL PK | Автоинкремент |
| email | VARCHAR(255) UNIQUE | Email (логин) |
| password | VARCHAR(255) | bcrypt hash |
| name | VARCHAR(255) | Имя пользователя |
| created_at | TIMESTAMPTZ | Дата создания |
| updated_at | TIMESTAMPTZ | Дата обновления |

**Индексы:** email (unique)

---

### MongoDB

#### bond_analyses — AI-анализы облигаций

```json
{
  "_id": ObjectId,
  "secid": "RU000A105KV7",
  "response": "Полный текст ответа AI...",
  "rating": 72,
  "json_data": { /* исходные данные облигации */ },
  "custom_json": { /* пользовательские данные */ },
  "timestamp": ISODate,
  "user_id": null,
  "saved_at": ISODate,
  "tags": []
}
```

**Индексы:** secid, timestamp  
**Время жизни:** постоянное хранение  
**Бизнес-правила:**
- rating: 0-100, парсится из текста ответа AI
- Один SECID может иметь много анализов (история)

---

#### bond_issuers — маппинг облигация ↔ эмитент

```json
{
  "_id": ObjectId,
  "secid": "RU000A105KV7",
  "emitter_id": 7720,
  "emitter_name": "ПАО Сбербанк",
  "is_hidden": false,
  "needs_sync": false,
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Индексы:** secid (unique), emitter_id  
**Источник:** MOEX ISS disclosure API  
**Обновление:** через SyncIssuerJob (фоновая задача)

---

#### bond_details — парсинг облигации с dohod.ru

```json
{
  "_id": ObjectId,
  "secid": "RU000A105KV7",
  "data": {
    "credit_rating": "A+",
    "sector": "Финансы",
    "risk_assessment": "...",
    /* произвольная структура от парсера */
  },
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Индексы:** secid (unique)  
**TTL:** 30 дней (auto-delete)  
**Источник:** Python dohod-parser.py

---

#### emitter_details — парсинг эмитента

```json
{
  "_id": ObjectId,
  "emitter_id": 7720,
  "data": {
    "full_name": "ПАО Сбербанк",
    "inn": "7707083893",
    "sector": "Банки",
    "ratings": { "expert_ra": "ruAAA", "acra": "AAA(RU)" },
    /* произвольная структура */
  },
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Индексы:** emitter_id (unique)  
**TTL:** 30 дней  
**Источник:** Python emit-parser.py

---

#### queue_jobs — трекинг фоновых задач

```json
{
  "_id": ObjectId,
  "job_id": "uuid-v4",
  "type": "ai_analysis",
  "reference_id": "RU000A105KV7",
  "status": "done",
  "data": {},
  "result": { "rating": 72, "analysis_id": "..." },
  "error": null,
  "attempts": 1,
  "max_attempts": 3,
  "created_at": ISODate,
  "started_at": ISODate,
  "finished_at": ISODate
}
```

**Индексы:** job_id (unique), status, reference_id  
**Статусы:** pending → running → done | error  
**Типы:** ai_analysis, parse_bond, parse_emitter, sync_issuer  
**Дедупликация:** перед созданием проверяется findPendingJob(type, referenceId)

---

#### issuer_ratings — кредитные рейтинги эмитентов

```json
{
  "_id": ObjectId,
  "emitter_id": 712,
  "issuer": "ОАО \"РЖД\"",
  "agency": "АКРА",
  "rating": "AA+(RU)",
  "score": 9,
  "updated_at": ISODate
}
```

**Индексы:** emitter_id, (emitter_id + agency) unique  
**Источник:** Автоматически из dohod.ru при запросе `/bonds/{secid}/dohod`  
**Обновление:** При каждом FetchAndSave — извлекаются АКРА, Эксперт РА, Fitch, Moody's, S&P  
**Score:** Берётся из `credit_rating` поля dohod.ru (1-10 шкала)  
**Связь:** `emitter_id` из `bond_issuers` коллекции (MOEX disclosure API)

---

#### dohod_details — данные облигации с dohod.ru

```json
{
  "_id": ObjectId,
  "isin": "RU000A0JSGV0",
  "secid": "RU000A0JSGV0",
  "issuer_name": "ОАО \"РЖД\"",
  "credit_rating": 9,
  "credit_rating_text": "AA",
  "akra": "AA+(RU)",
  "expert_ra": "ruAAA",
  "quality": 4.0,
  "description": "Описание эмитента...",
  "event": "право продать (put)",
  "coupon_rate": 9.8,
  "simple_yield": 15.15,
  "fetched_at": ISODate,
  "updated_at": ISODate
}
```

**Индексы:** isin (unique), secid  
**TTL:** 30 дней  
**Источник:** HTTP парсинг `__NUXT_DATA__` с analytics.dohod.ru  
**Содержит:** ~80 полей (рейтинги, качество, финансы эмитента, параметры облигации)

---

#### chat_sessions — чат-сессии

```json
{
  "_id": ObjectId,
  "session_id": "uuid-v4",
  "title": "Анализ ОФЗ",
  "agent_type": "analyst",
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Индексы:** session_id (unique), updated_at  
**Типы агентов:** analyst, report, support

---

#### chat_messages — сообщения чата

```json
{
  "_id": ObjectId,
  "session_id": "uuid-v4",
  "role": "user",
  "content": "Проанализируй ОФЗ-26238",
  "created_at": ISODate
}
```

**Индексы:** session_id, created_at  
**Роли:** user, assistant, system

---

### Внешние данные (не хранятся, кэшируются в Redis)

#### MOEX ISS — Облигация

Ответ MOEX ISS API (кэш 24 часа):

```json
{
  "SECID": "RU000A105KV7",
  "SHORTNAME": "Сбер Б БО37",
  "SECNAME": "Сбербанк ПАО БО-37",
  "ISIN": "RU000A105KV7",
  "FACEVALUE": 1000,
  "MATDATE": "2027-03-15",
  "COUPONPERIOD": 182,
  "COUPONVALUE": 36.90,
  "COUPONPERCENT": 7.38,
  "NEXTCOUPON": "2026-09-15",
  "ACCRUEDINT": 12.45,
  "BONDTYPE": "Корпоративная",

  "LAST": 98.50,
  "BID": 98.45,
  "OFFER": 98.60,
  "YIELD": 8.12,
  "DURATION": 485,
  "VOLTODAY": 15000,

  "_calculated": {
    "PRICE_RUB": 985.00,
    "VALUE_TODAY_RUB": 14775000,
    "DAYS_TO_MATURITY": 341,
    "IS_FLOAT": false,
    "IS_INDEXED": false,
    "BOND_CATEGORY": "Корпоративная",
    "RISK_CATEGORY": "moderate"
  }
}
```

**Критично:** Цены MOEX — в ПРОЦЕНТАХ от номинала, не в рублях!
- `PRICE_RUB = (LAST / 100) × FACEVALUE`
- `VALUE_TODAY_RUB = VOLTODAY × PRICE_RUB`

---

### Бизнес-правила

#### Парсинг AI-рейтинга (приоритет)

```
1. [RATING:72]                     — машиночитаемый тег (высший приоритет)
2. "Итоговая оценка: **72/100**"   — контекст + markdown bold
3. "Итоговая оценка: 72 баллов"    — контекст + словоформа
4. "Оценка: 72/100"                — альтернативный контекст
5. **72/100** (standalone)          — только bold
6. Последний 72/100 в тексте        — fallback
```

#### Типы купонов

```
IS_FLOAT    = BONDTYPE содержит "Флоатер" или "Float"
IS_INDEXED  = SECTYPE === "6" или BONDTYPE содержит "Индексируемые"

Отображение купона:
  Флоатер: (COUPONMIN + COUPONMAX) / 2
  Обычный: COUPONPERCENT
  Fallback: (COUPONVALUE / FACEVALUE) × 100
```

#### Месячные купоны

```
COUPONPERIOD >= 27 AND COUPONPERIOD <= 33 → ежемесячный купон
```

#### Сортировка облигаций

```
best          — комплексный рейтинг (доходность + объём + срок)
yield_desc    — YIELD убывание
yield_asc     — YIELD возрастание
maturity_asc  — MATDATE ближайшее погашение
maturity_desc — MATDATE дальнее погашение
volume_desc   — объём торгов убывание
coupon_desc   — купонная ставка убывание
coupon_asc    — купонная ставка возрастание
```

#### TTL кэшей

| Данные | TTL | Хранилище |
|--------|-----|-----------|
| Список облигаций MOEX | 24 часа | Redis |
| Детали облигации MOEX | 24 часа | Redis |
| Купоны и история MOEX | 24 часа | Redis |
| Парсинг dohod.ru | 30 дней | MongoDB (TTL index) |
| Парсинг эмитента | 30 дней | MongoDB (TTL index) |
| AI-анализы | ∞ | MongoDB |
| Чаты и сообщения | ∞ | MongoDB |

#### Retry-логика OpenAI

```
Макс. попыток: 3
Backoff: 2с, 4с (экспоненциальный)
Retry при: 429 (rate limit), 5xx
Успех при: 2xx
Отмена при: 4xx (кроме 429)
Таймаут: 200с на запрос
```

#### Дедупликация задач

```
Перед созданием задачи:
  findPendingJob(type, referenceId)
  → Если есть pending/running задача с тем же типом и referenceId
  → Вернуть существующий job_id вместо создания нового
```

### Система агентов (чат)

| Тип | Название | Промпт-файл | Описание |
|-----|---------|-------------|----------|
| analyst | Финансовый аналитик | analyst.txt | Обсуждение облигаций, образование |
| report | Генератор отчётов | report.txt | Структурирование заметок в документы |
| support | Анализ поддержки | support.txt | Анализ тикетов, root cause |

Промпты загружаются из `data/prompts/{type}.txt`.

### AI-анализ: 4-блочная система оценки

```
Кредитное качество:       /45 баллов
Ликвидность и доступность: /25 баллов
Адекватность доходности:   /15 баллов
Структура выпуска:         /15 баллов
────────────────────────────────────
Итого:                     /100 баллов → [RATING:XX]
```

Промпт: `data/prompts/bond_analyst.txt`
