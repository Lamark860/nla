# Сущности и хранилище данных

Источник истины по схеме — `internal/database/migrations/*.sql`. Этот файл — обзорная карта.

После Фазы 1 (см. `roadmap.md`) единственное persistent-хранилище — **PostgreSQL 16**. MOEX-кэш — in-process (`memoryCache` в `service/bond.go`). MongoDB и Redis удалены полностью.

## Обзор таблиц

```
PostgreSQL — единое хранилище
┌──────────────────────────────────────────────────────────────────┐
│ Auth/User-owned:    users, favorites, portfolio_positions        │
│ Reference (MOEX):   bond_issuers, issuer_ratings                 │
│ External cache:     dohod_details (TTL 30d, cleanup nightly)     │
│ AI / scoring:       bond_analyses, scoring_profiles, bond_scores,│
│                     bond_score_explanations                       │
│ Async jobs:         queue_jobs                                    │
│ Chat:               chat_sessions, chat_messages (CASCADE)        │
└──────────────────────────────────────────────────────────────────┘
```

## users

| Поле | Тип | Описание |
|---|---|---|
| id | BIGSERIAL PK | Автоинкремент |
| email | VARCHAR(255) UNIQUE | Email (логин) |
| password | VARCHAR(255) | bcrypt hash |
| name | VARCHAR(255) | Имя пользователя |
| created_at / updated_at | TIMESTAMPTZ | Даты |

## favorites

`(user_id, secid) UNIQUE`. CASCADE delete по user_id. Используется `/api/v1/favorites/*`.

## bond_issuers

Маппинг `secid → emitter_id + emitter_name`. PK по `secid`. Поля `inn` и `external_ids JSONB` зарезервированы под Фазу 5 (Tinkoff API).

| Поле | Тип | Использование |
|---|---|---|
| secid | VARCHAR(64) PK | MOEX security ID |
| emitter_id | BIGINT | Из MOEX disclosure API |
| emitter_name | TEXT | Подтягивается из dohod.ru при первом запросе |
| inn | TEXT (NULL) | Phase 5 — Tinkoff `Asset.Brand.companyInn` |
| external_ids | JSONB (NULL) | Phase 5 — `{"tinkoff": "BBG..."}` для маппинга на внешние ID |
| is_hidden, needs_sync | BOOLEAN | Флаги UI |

## issuer_ratings

Композитный PK `(emitter_id, agency)`. Два скора: `score` (legacy 1-10) и `score_ord` (canonical 1-22). Для сортировок и cross-agency сравнений — всегда `score_ord`. Подробности нормализации — `internal/service/rating_score.go::NormalizeRating`, TS-зеркало в `frontend/composables/useRating.ts`.

## dohod_details

`isin TEXT PK` + индекс по `secid`. Полные ~80 полей парсера dohod.ru хранятся в `data JSONB`. TTL — 30 дней (поле `updated_at`); чистка ночным DELETE'ом (см. `roadmap.md` Фаза 1.6).

## bond_analyses

UUID PK (`gen_random_uuid()`). История AI-анализов по бумаге. Содержит сам текст ответа LLM, парсенный балл 0..100 и опциональный `user_id` (FK → users, ON DELETE SET NULL). Гибкие поля `json_data` и `custom_json` — JSONB.

Парсинг рейтинга (`service/analysis.go::parseFloatRating`) уважает порядок:

1. `[RATING:72.5]` — машиночитаемый тег (приоритет)
2. «Итоговая оценка: **72/100**» — markdown bold + контекст
3. «Итоговая оценка: 72 баллов» — словоформа
4. Просто `**72/100**` standalone
5. Последний `XX/100` в тексте — fallback

## queue_jobs

UUID PK. Атомарный claim через `UPDATE ... FOR UPDATE SKIP LOCKED RETURNING` в `QueueRepo.FetchPending`. Дедупликация — partial index по `(type, secid)` для статусов `pending|running`. На старте worker'а — `ResetStaleJobs(3 min)`.

Типы задач: `ai_analysis`, `parse_bond`, `parse_emitter`, `sync_issuer`, `parse_dohod`. Поля `data` и `result` — JSONB.

## chat_sessions / chat_messages

`session_id VARCHAR(64)` как естественный ключ. Сообщения — `BIGSERIAL` PK + FK на сессию с `ON DELETE CASCADE`. Индекс `(session_id, created_at)` для timeline.

Типы агентов (`agent_type`): `analyst`, `report`, `support` — промпты в `data/prompts/<type>.txt`.

## scoring_profiles (Фаза 2)

`code` как PK. Три preset-профиля (`low`, `mid`, `high`) — `is_preset = TRUE`, `user_id` NULL. Пользовательские профили (Pro Plus в Фазе 7) — `is_preset = FALSE`, `user_id` указывает владельца, `code` — synthetic.

`weights JSONB` — словарь `{factor_name: weight_float}`. Список факторов и стартовые значения — в `roadmap.md` Фаза 2.

## bond_scores (Фаза 2)

Хранит **историю** баллов: `(secid, profile_code, score, breakdown, computed_at)`. Latest-balance по `(secid, profile_code)` достаётся `ORDER BY computed_at DESC LIMIT 1`. `breakdown JSONB` — массив объектов `{factor, raw_value, normalised, contribution}` для каждого из 12 факторов.

## bond_score_explanations (Фаза 2)

LLM-объяснение конкретного снимка скоринга. FK на `bond_scores(id)` с CASCADE. Дорогая операция (вызов OpenAI), вызывается только по явному запросу.

## portfolio_positions (Фаза 4)

`user_id` FK с CASCADE. Поля: `secid`, `qty`, `price_in`, `date_in`, `note`. На основе этой таблицы — расчёт cash-flow на год, средневзвешенной YTM, дюрации портфеля и профильного балла портфеля.

---

## Не хранится локально

MOEX ISS API кэшируется **только в памяти процесса** (`memoryCache` в `service/bond.go`). Два ключа: `bonds:list` и `bonds:{secid}`. TTL 24 часа. При рестарте API — пустой кэш, прогрев одним запросом.

Это допустимо для one-instance деплоя. При горизонтальном масштабировании потребуется внешний кэш — на этом этапе можно вернуть Redis или использовать `pg_advisory_lock` + таблицу-кэш.

---

## Бизнес-правила

### Парсинг AI-рейтинга

См. `bond_analyses` выше.

### Типы купонов

```
IS_FLOAT    = BONDTYPE содержит "Флоатер" или "Float"
IS_INDEXED  = SECTYPE === "6" или BONDTYPE содержит "Индексируемые"

Отображение купона:
  Флоатер:  (COUPONMIN + COUPONMAX) / 2
  Обычный:  COUPONPERCENT
  Fallback: (COUPONVALUE / FACEVALUE) × 100
```

### Месячные купоны

`COUPONPERIOD >= 27 AND COUPONPERIOD <= 33`.

### Сортировки облигаций (`?sort=`)

`best`, `yield_desc`, `yield_asc`, `maturity_asc`, `maturity_desc`, `volume_desc`, `coupon_desc`, `coupon_asc`.

### Дедупликация AI-задач

Перед созданием task'a — `QueueRepo.FindPending(type, secid)`. Если есть `pending`/`running` с тем же ключом — возвращается существующий `job_id`.

### Система чат-агентов

| Тип | Название | Промпт |
|---|---|---|
| analyst | Финансовый аналитик | `data/prompts/analyst.txt` |
| report | Генератор отчётов | `data/prompts/report.txt` |
| support | Анализ поддержки | `data/prompts/support.txt` |

### AI-анализ — 4-блочная оценка (используется в Фазе 2 как один из факторов скоринга)

```
Кредитное качество:       /45 баллов
Ликвидность:              /25 баллов
Адекватность доходности:  /15 баллов
Структура выпуска:        /15 баллов
─────────────────────────────────────
Итого:                    /100 → [RATING:XX]
```

Промпт: `data/prompts/bond_analyst.txt`.
