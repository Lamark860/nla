# STATUS — текущее состояние работы

> Файл предназначен для возобновления работы после ресета Claude-сессии. **Всегда читать первым.**
> Источник истины по плану — `docs/roadmap.md`. По архитектуре — `CLAUDE.md`. Здесь только: где остановились, что в работе, что дальше.

## Сейчас в работе

**Фаза 1 — инфра-упрощение** ✅ done (2026-05-11).

**Фаза 2 — Scoring engine** — backend готов.
**Фаза 3 — UI трёх профилей** — первый раунд собран (sidebar switcher + 3 бейджа на hero + новый таб «Индекс» с breakdown + кнопка «Получить разбор»). Открыто: 3-бейдж-режим на главных страницах (IssuerCardGrid / by-issuer / monthly) — пока не подключён, требует batch-эндпоинта.

### Сделано (2026-05-11, не закоммичено на момент записи)

- `internal/scoring/` — чистая логика (engine + 12 факторов + normalize + 3 пресета + 17 тестов)
- `internal/model/scoring.go` — `ScoringProfile`, `BondScoreExplanation`, `ScoreExplainJobData`. `BondScore` живёт в `internal/repository/scoring.go` (не в model — был бы импорт-цикл model → scoring → model)
- `internal/repository/scoring.go` — `ScoringRepo` с методами `GetProfile`/`ListProfiles`/`GetScoreByID`/`GetLatestScore`/`InsertScore`/`DeleteScoresOlderThan`/`GetExplanationByScoreID`/`InsertExplanation`
- `internal/service/scoring.go` — `ScoringService` собирает `Input` из `BondService.GetBondDetail` + `RatingService.GetByEmitterID` + `DetailsService.GetDetails`. TTL 24h. `ComputeAll`/`ComputeOne`/`Explain`/`GetExplanation`/`ListProfiles`. ОФЗ-бенчмарк пока `nil` (factor #3 уходит в `missing_factors`)
- `internal/handler/scoring.go` — `GET /api/v1/scoring/profiles`, `GET /api/v1/bonds/{secid}/score[?profile=]`, `POST /api/v1/bonds/{secid}/score/explain?profile=X` (enqueue + immediate response с `job_id` + `score_id`)
- `internal/queue/worker.go` — обработчик `JobTypeScoreExplain` (вызывает `scoringSvc.Explain(scoreID)`, помечает job done с `explanation_id`)
- `data/prompts/scoring_explain.txt` — RU-промпт «объясни балл по факторам в 3-4 абзаца, никаких рекомендаций покупать/продавать»
- `internal/client/openai/client.go` — добавлен публичный `Model()` getter (нужен сервису для тэгирования explanation-row)
- `internal/scoring/normalize.go` — нормализатор YTM ужесточён: >100% → 0 (MOEX отдаёт мусорные значения для near-maturity/illiquid бумаг — известная регрессия из CLAUDE.md). Тест на это добавлен
- DI в `cmd/api/main.go` + регистрация роутов в `internal/router/router.go`

**Smoke-проверка (живой стек):**
- `GET /api/v1/scoring/profiles` → 3 пресета с весами из migration 0002 ✓
- `GET /api/v1/bonds/RU000A108KK5/score` → 3 балла (low=58.6 / mid=48.3 / high=43.0), кэш-хит при повторе ✓
- `bond_scores` хранит рассчитанные ряды ✓
- `go test ./internal/...` 19/19 пакетов зелёных ✓

### Следующий шаг — Фаза 3 (UI трёх профилей)

Бэк готов отдавать данные. Дальше работа на фронте (см. `docs/roadmap.md` Фаза 3):

1. Composable `useScoringProfile()` — глобальный профиль в localStorage, дефолт `mid`
2. Глобальный переключатель в `layouts/default.vue` (Sidebar) — `🛡️ Низкий / ⚖️ Средний / 🚀 Повышенный`
3. На `BondHero.vue` / `IssuerCardGrid.vue` — три `<AiScore>` бейджа, активный профиль bold
4. Новый таб `BondScoreTab.vue` (заменит или дополнит `BondAiTab`): 3 балла сверху, раскрывающийся breakdown по факторам, кнопка «Получить разбор» → POST /score/explain → polling `/jobs/{id}` → отрисовка `explanation.text`
5. Сортировки/фильтры на главных страницах пересчитываются под активный профиль

### Открытые backend-долги под Фазу 2

- **ОФЗ-бенчмарк для factor #3 (`ytm_premium`)** — нужно либо подтянуть MOEX yield curve, либо захардкодить bucket-map (≤1y/1-3y/3-5y/5y+) по текущей ключевой ставке ЦБ. Пока для всех бумаг фактор #3 — `missing_factors`
- **Калибровка весов на ~100 бумагах** — фронт + UI dashboard для сравнения score vs ручная оценка

## Зафиксированные решения (не пересматривать без явного слова пользователя)

- Терминология: **«Аналитический индекс»** (не «AI-балл/оценка/рейтинг»). Внутренние идентификаторы (`AiScore.vue`, `analyze` endpoint, `bond_analyses` таблица) **НЕ** переименовываем — только user-facing метки
- Профили скоринга: **Низкий / Средний / Повышенный риск** (внутр. коды `low` / `mid` / `high`, иконки 🛡️ / ⚖️ / 🚀)
- Алерты / email-уведомления / telegram-бот — **out of scope** (инструмент анализа, не торговый)
- Стек после Фазы 1: **Go + Postgres + Nginx**. Mongo, Redis, отдельный Node-frontend контейнер — **уходят**
- Frontend: Nuxt 3 в **SSG-режиме** (`ssr: false`, `nuxt generate`), nginx раздаёт статику
- Скоринг: **12 факторов**, веса в `scoring_profiles.weights JSONB`, калибровка после первого запуска на ~100 бумагах. Список факторов с весами — в `docs/roadmap.md` Фаза 2
- LLM используется **только для текстовых объяснений** (кэширование в `bond_score_explanations`), сам балл считается детерминированно
- Без дизайнера — UI делаем в текущем violet-стиле, новые компоненты используют системные `<Panel>/<KPI>/<InfoRow>/<AiScore>/<Pill>/<Tag>`

## Журнал фаз

| Фаза | Статус | Дата | Заметки |
|---|---|---|---|
| 0 — терминология + доки | ✅ done | 2026-05-11 | `v0.14.8` в CHANGELOG. Не закоммичено |
| 1 — инфра-упрощение | ✅ done | 2026-05-11 | `v0.15.0` в CHANGELOG. Стек: api + postgres + frontend(nginx+SSG). 3295 issuers, 631 ratings, 57 analyses, 47 dohod, 1 chat session/4 messages мигрированы из Mongo |
| 2 — scoring engine | 🟡 backend done | 2026-05-11 | Engine + repo + service + handler + worker job + LLM-prompt. API/profile/score/explain эндпоинты живут. Открыто: ОФЗ-бенчмарк для factor #3 |
| 3 — UI 3 профилей | 🟡 partial | 2026-05-11 | Sidebar switcher + 3 бейджа на BondHero + BondScoreTab (breakdown 12 факторов + кнопка разбора с polling). Открыто: 3-бейджи на главных списках, сортировка по активному профилю |
| 4 — portfolio | ⏳ ждёт | — | — |
| 5 — Tinkoff events | ⏳ ждёт | — | — |
| 6 — мелочи (скринеры, ical, реф) | ⏳ ждёт | — | — |
| 7 — биллинг | ⏳ ждёт | — | — |

## Чек-лист Фазы 1 — все шаги выполнены

- [x] **1.1** DDL — миграция `0002_postgres_full_schema.sql` (13 таблиц + 3 seed-профиля). Применена.
- [x] **1.2** Postgres-репозитории в `internal/repository/`: `analysis.go`, `issuer.go`, `details.go`, `rating.go`, `queue.go`, `chat.go`. Имена методов 1:1 с Mongo-версиями.
- [x] **1.3** `cmd/migrate-mongo-pg/main.go` создан, запущен один раз, перенёс реальные данные. Сам скрипт удалён вместе с Mongo (всё в git history).
- [x] **1.4** DI в `cmd/api/main.go` переключён на Postgres-репо. Стек запущен и проходит smoke (`/health` ok, `/api/v1/bonds` отдаёт).
- [x] **1.5** Удалены: `internal/mongo/*.go`, `internal/database/mongodb.go`, `cmd/sync-ratings/`. Mongo поля из `config.go` убраны. `go mod tidy` отработал.
- [x] **1.6** Redis: `internal/database/redis.go` удалён, `BondService` переведён на `memoryCache` (sync.RWMutex + map с TTL), Redis-поля из конфига убраны. Compose без `redis` сервиса.
- [x] **1.7** Nuxt SSG: `ssr: false` в `nuxt.config.ts`. `frontend/Dockerfile` multi-stage → nginx со статикой. `frontend/nginx.conf` раздаёт статику + проксит `/api`. Отдельный root-уровень `nginx/` удалён.
- [x] **1.8** Финальный compose — 3 контейнера: `api`, `postgres`, `frontend` (nginx со статикой). Volume только `nla-pgdata`. Mongo и Redis volumes удалены.
- [x] **1.9** vitest 117/117 ✓, go test all ok ✓, smoke на новом стеке ✓. `docs/entities.md` переписан под Postgres-схему. CHANGELOG `v0.15.0`.

## Команды восстановления контекста при reset

```bash
git status --short
git log -10 --oneline
cat docs/STATUS.md      # этот файл
cat docs/roadmap.md     # план
ls docs/
```

## Open decisions to revisit later

- Биллинг-провайдер: ЮКасса vs CloudPayments
- Партнёрка с первым брокером: Тинькофф / БКС / Финам
- Конкретные веса скоринга — после калибровки на ~100 бумагах
- Multi-level orderbook / trades feed / yield_history — backend-блокеры для редизайна. Решить когда дойдёт до Фазы 2 — нужны или нет для скоринга
