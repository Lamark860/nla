## API Plan — Полный список эндпоинтов NLA

### Соглашения

- Базовый путь: `/api/v1`
- Формат: JSON
- Auth: `Authorization: Bearer <jwt>` для защищённых эндпоинтов
- Ошибки: `{"error": "описание"}`
- Пагинация: `?page=1&per_page=20` → `{"data": [...], "meta": {"page": 1, "per_page": 20, "total": 500}}`

---

### Auth

| Метод | Путь | Auth | Описание | Статус |
|-------|------|------|----------|--------|
| POST | /auth/register | — | Регистрация | ✅ Готово |
| POST | /auth/login | — | Логин → JWT | ✅ Готово |
| GET | /auth/me | JWT | Текущий пользователь | ✅ Готово |

---

### Bonds (облигации)

| Метод | Путь | Auth | Описание | Статус |
|-------|------|------|----------|--------|
| GET | /bonds | — | Список облигаций (пагинация, сортировка) | ✅ Готово |
| GET | /bonds/:secid | — | Детали облигации | ✅ Готово |
| GET | /bonds/:secid/coupons | — | Купонный календарь | ✅ Готово |
| GET | /bonds/:secid/history | — | История цен (180 дней OHLC) | ✅ Готово |
| GET | /bonds/monthly | — | Облигации с ежемесячным купоном | ✅ Готово |
| GET | /bonds/by-issuer | — | Все облигации по эмитентам | — |

**Query params для GET /bonds:**
```
sort=best|yield_desc|yield_asc|maturity_asc|maturity_desc|volume_desc|coupon_desc|coupon_asc
page=1
per_page=20
```

---

### AI Analysis (анализ)

| Метод | Путь | Auth | Описание | Статус |
|-------|------|------|----------|--------|
| POST | /bonds/:secid/analyze | — | Запустить AI-анализ (async → job_id) | ✅ Готово |
| GET | /bonds/:secid/analyses | — | История анализов по облигации | ✅ Готово |
| GET | /bonds/:secid/analysis-stats | — | Статистика анализов (avg rating, total) | ✅ Готово |
| GET | /analyses/:id | — | Конкретный анализ | ✅ Готово |

**POST /bonds/:secid/analyze response:**
```json
{
  "job_id": "uuid",
  "status": "pending"
}
```

---

### Jobs (фоновые задачи)

| Метод | Путь | Auth | Описание | Статус |
|-------|------|------|----------|--------|
| GET | /jobs/:id | — | Статус задачи (polling) | ✅ Готово |
| GET | /queue/stats | — | Статистика очереди | ✅ Готово |

**Response:**
```json
{
  "job_id": "uuid",
  "type": "ai_analysis",
  "status": "done",
  "result": { "rating": 72, "analysis_id": "..." },
  "error": null,
  "created_at": "...",
  "finished_at": "..."
}
```

---

### Chat (нейрочат)

| Метод | Путь | Auth | Описание | ASH-аналог |
|-------|------|------|----------|-----------|
| GET | /chat/sessions | — | Список сессий | actionIndex |
| POST | /chat/sessions | — | Создать сессию | actionCreate |
| GET | /chat/sessions/:id | — | Получить сессию с сообщениями | actionView |
| DELETE | /chat/sessions/:id | — | Удалить сессию | actionDelete |
| POST | /chat/sessions/:id/messages | — | Отправить сообщение (sync AI ответ) | actionSend |

**POST /chat/sessions request:**
```json
{
  "agent_type": "analyst",
  "title": "Обсуждение ОФЗ"
}
```

**POST /chat/sessions/:id/messages request:**
```json
{
  "content": "Что думаешь про ОФЗ-26238?"
}
```

**Response:**
```json
{
  "user_message": {
    "session_id": "uuid",
    "role": "user",
    "content": "Что думаешь про ОФЗ-26238?",
    "created_at": "..."
  },
  "assistant_message": {
    "session_id": "uuid",
    "role": "assistant",
    "content": "ОФЗ-26238 — это...",
    "created_at": "..."
  }
}
```

---

### Chat Agents (информация об агентах)

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | /chat/agents | — | Список доступных агентов |

**Response:**
```json
[
  {
    "type": "analyst",
    "name": "Финансовый аналитик",
    "description": "Обсуждение облигаций, образование",
    "icon": "bi-graph-up"
  },
  ...
]
```

---

### Issuers (эмитенты)

| Метод | Путь | Auth | Описание | ASH-аналог |
|-------|------|------|----------|-----------|
| GET | /issuers | — | Список эмитентов (с кол-вом облигаций) | getAllIssuers |
| GET | /issuers/:id | — | Данные эмитента | getEmitterDetails |
| GET | /issuers/:id/bonds | — | Облигации эмитента | getBySecids filtered |

---

### Health

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | /health | — | Статус сервиса | ✅ Готово |

---

### Порядок реализации

```
Фаза 1:  Auth ✅ → Bonds ✅ → MOEX client ✅
Фаза 2:  AI Analysis ✅ → Queue worker ✅ → Job polling ✅
Фаза 2.5: Frontend (Nuxt 3) ✅ — список, детали, AI-таб, купоны
Фаза 3:  Chat → Agents → Prompts
Фаза 4:  Issuers → Parsers (Python integration)
Фаза 5:  Тесты на весь функционал
```
