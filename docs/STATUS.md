# STATUS — текущее состояние работы

> Файл предназначен для возобновления работы после ресета Claude-сессии. **Всегда читать первым.**
> Источник истины по плану — `docs/roadmap.md`. По архитектуре — `CLAUDE.md`. Здесь только: где остановились, что в работе, что дальше.

## Сейчас в работе

**Фаза 1 — инфра-упрощение** ✅ done (2026-05-11).

**Фаза 2 — Scoring engine** — в работе.

### Сделано (2026-05-11, после v0.15.0, не закоммичено)

- `internal/scoring/` пакет создан, чистая логика без БД-зависимостей:
  - `engine.go` — типы `Input`/`Profile`/`BreakdownItem`/`ScoreResult`, `Compute(in, profile) → ScoreResult`, constants `FactorXxx` совпадают с JSONB-ключами в `scoring_profiles.weights` (миграция 0002, lines 185-192)
  - `factors.go` — 12 extract-функций. Сигнатура `(Input) → (raw float64, hasData bool)`. Источники: `model.Bond` (MOEX поля), `[]model.IssuerRating`, `*model.DohodBondData`, опц. `BenchmarkYieldPct *float64`
  - `normalize.go` — таблично-кусочные мапы raw → 0..100. Missing-data policy: positive weight → 50 (нейтрально), negative weight → 0 (штраф не срабатывает)
  - `presets.go` — `PresetLow`/`PresetMid`/`PresetHigh` 1:1 с seed-строками БД, плюс `Presets map[string]Profile`
  - `engine_test.go` — 17 кейсов (extract happy/missing, normalize anchors, Compute end-to-end по 3 пресетам, missing-factors tracking, lock-in весов)
- `go test ./internal/scoring/...` 17/17 ✓, `go test ./internal/...` весь пакет зелёный

### Следующий шаг — `internal/repository/scoring.go` + `internal/service/scoring.go`

1. **Репо-слой** — pgx-методы:
   - `ScoringProfileRepo.GetByCode(code)` / `ListPresets()` / `ListForUser(userID)` / `Upsert(profile)` → таблица `scoring_profiles`
   - `BondScoreRepo.GetLatest(secid, profile)` / `Insert(score)` → `bond_scores`. TTL 24h: при `computed_at < now()-interval '1 day'` пересчёт
   - `BondScoreExplanationRepo.GetByScoreID(id)` / `Insert(...)` → `bond_score_explanations`
2. **Сервис-слой** `internal/service/scoring.go` — `ComputeAll(secid)` (3 профиля) / `ComputeOne(secid, profile)` с автокэшем; собирает `Input` из `BondService.GetBySecid` + `RatingService.GetByEmitterID` + `DetailsService.GetBySecid`
3. **Handler + routes** — `GET /api/v1/bonds/{secid}/score[?profile=]`, `GET /api/v1/scoring/profiles`. LLM-объяснение через очередь (как нынешний `analyze`), endpoint `POST .../score/explain?profile=X`, новый `JobType="score_explain"` в `queue/worker.go`
4. **ОФЗ-бенчмарк** — пока заглушка `nil`; решить когда брать (MOEX yield curve или захардкоженный buckets-map по дюрации). Без бенчмарка фактор #3 будет в `missing_factors`, остальное считается

После этого можно мерджить в `v0.16.0` и переходить на Фазу 3 (UI трёх профилей).

**Следующая фаза:** Фаза 3 — UI трёх профилей (после API).

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
| 2 — scoring engine | 🟡 in progress | 2026-05-11 | `internal/scoring/` чистая логика готова (17 тестов ✓). Осталось: repo/service/handler-слой + ОФЗ-бенчмарк + LLM-объяснение через очередь |
| 3 — UI 3 профилей | ⏳ ждёт | — | — |
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
