# CLAUDE.md — NLA Project Context

## Working preferences

- Persistent context lives **in this file (or other repo docs)**, not in private session memory. If something is worth remembering across sessions, write it here.
- Repo is mirrored to GitHub (`git@github.com-personal:Lamark860/nla.git`, branch `main`) via personal SSH alias. Treat secrets in git history seriously — history is published.
- Conversational replies in Russian; code/identifiers/paths in English.

## Project

**NLA** — анализатор облигаций MOEX для квалифицированных частных инвесторов. Эволюция проекта ASH.
Продуктовое ядро (см. `docs/roadmap.md`) — детерминированный scoring-движок с тремя профилями риска
(Низкий / Средний / Повышенный) + LLM как объяснитель факторов. Пользовательский термин — **«Аналитический индекс»** (не «AI-оценка»).

## Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| API | Go 1.25 + Chi router | REST API |
| Auth | JWT (HS256) + bcrypt | Stateless authentication |
| Relational DB | PostgreSQL 16 | Users, favorites, (Фаза 1: всё остальное переезжает сюда) |
| Document DB | MongoDB 7 | bond_analyses, bond_issuers, dohod_details, issuer_ratings, queue_jobs, chat — **в процессе миграции на Postgres JSONB, см. Фазу 1** |
| Cache | Redis 7 | Кэш списка облигаций (2 ключа) — **в Фазе 1 заменяется на in-memory** |
| Proxy | Nginx | Reverse proxy + статика фронта (после Фазы 1 — без отдельного Node-контейнера) |

## Commands

```bash
# Start everything
cd /Users/maximlomaev/dockers/preprod/src/nla
docker compose up -d

# Rebuild after code changes (API only)
docker compose up -d --build api

# Rebuild frontend
docker compose up -d --build frontend

# Rebuild everything
docker compose up -d --build

# Logs
docker compose logs -f api

# Stop
docker compose down

# Full reset (including volumes)
docker compose down -v
```

## Tests

```bash
# Backend (Go) — all tests
make test          # or: go test ./internal/... -count=1

# Verbose / coverage
make test-v
make test-cover

# Specific package
go test ./internal/service/... -v
go test ./internal/middleware/... -v
go test ./internal/handler/... -v

# Frontend (TS) — vitest
cd frontend && npm test            # one-shot
cd frontend && npm run test:watch  # watch mode
```

### Test structure
**Backend (Go, 145+ subtests):**
- `internal/service/auth_test.go` — auth unit tests (mock repo)
- `internal/service/analysis_test.go` — AI rating parser ([RATING:XX.X], итоговая оценка, баллы, bold, /100)
- `internal/service/rating_score_test.go` — `NormalizeRating` table tests across 8 agencies + cross-agency ordering + ДОХОДЪ legacy round-trip
- `internal/middleware/auth_test.go` — JWT middleware
- `internal/handler/auth_test.go` — HTTP handler integration
- `internal/handler/test_helpers_test.go` — shared test mocks
- `internal/client/dohod/client_test.go` — dohod.ru Nuxt payload parser

**Frontend (TS, vitest, 117 cases):**
- `frontend/composables/useRating.spec.ts` — mirrors `rating_score_test.go` to keep TS port behaviourally identical to Go (any drift breaks tests on both sides). Add new agency-format cases to **both** files

**Note:** Repo-layer (`internal/repository/*.go`) и MOEX API clients не покрыты unit-тестами — требуют интеграционной среды (testcontainers с Postgres). Текущие тесты — pure logic: парсинг, auth, middleware, normalisation.

## API Documentation

OpenAPI 3.0 spec: `docs/openapi.yaml`
View with: https://editor.swagger.io (paste YAML) or any OpenAPI viewer

## Ports

| Service | Host Port | Container Port |
|---------|-----------|---------------|
| Frontend (nginx + SSG SPA + /api proxy) | 8090 | 80 |
| Go API | (internal only) | 8080 |
| PostgreSQL | 5433 | 5432 |

**UI:** http://localhost:8090/ (статика отдаётся frontend-контейнером)
**API:** http://localhost:8090/api/v1/ (через тот же nginx → Go API)
**Postgres (для psql/dump):** localhost:5433 (user `nla`, db `nla`)

## Project Structure

```
nla/
├── cmd/api/main.go              # Entry point, DI wiring, queue worker, issuer sync
├── internal/
│   ├── config/config.go         # Environment config (Postgres DSN, OpenAI, JWT)
│   ├── database/                # Postgres pool + embed-based migrations loader
│   │   └── migrations/          # 0001_init.sql + 0002_postgres_full_schema.sql (см. entities.md)
│   ├── client/
│   │   ├── moex/client.go       # MOEX ISS API client (5 endpoints + GetDisclosure + GetCCIRatings)
│   │   ├── openai/client.go     # OpenAI client (retry, reasoning models)
│   │   └── dohod/client.go      # Dohod.ru Nuxt SSR parser (HTTP + __NUXT_DATA__)
│   ├── handler/                 # HTTP layer (auth, bond, analysis, details, rating, chat, favorite)
│   ├── middleware/auth.go       # JWT middleware
│   ├── model/                   # Data structs (bond, dohod, chat, job, user)
│   ├── repository/              # pgx-based persistence layer (ВСЁ persistent живёт тут)
│   │   ├── user.go              # users
│   │   ├── favorite.go          # favorites
│   │   ├── analysis.go          # bond_analyses (UUID PK)
│   │   ├── issuer.go            # bond_issuers
│   │   ├── details.go           # dohod_details (JSONB body, 30d TTL)
│   │   ├── rating.go            # issuer_ratings (composite PK emitter_id+agency)
│   │   ├── queue.go             # queue_jobs (atomic FetchPending via SKIP LOCKED)
│   │   └── chat.go              # chat_sessions + chat_messages (CASCADE)
│   ├── queue/worker.go          # Background job worker (goroutine)
│   ├── service/
│   │   ├── auth.go              # Auth business logic
│   │   ├── bond.go              # BondService + in-memory cache (sync.RWMutex + map с TTL)
│   │   ├── bond_parse.go        # parseBond, mergeYieldData (MOEX rows → model.Bond)
│   │   ├── bond_calc.go         # calculateFields, calcRiskCategory, sortBonds, bondScore
│   │   ├── bond_sync.go         # SyncMissingIssuers, SyncMissingRatingsFromMoex
│   │   ├── bond_helpers.go      # extractRows, get*/safe* primitives
│   │   ├── analysis.go          # AI analysis + rating parser
│   │   ├── details.go           # Dohod.ru service (cache + retry + save + rating sync)
│   │   ├── rating.go            # Credit ratings CRUD + GetAll by emitter_id
│   │   ├── rating_score.go      # NormalizeRating: 8 agencies → 22-level ordinal scale
│   │   ├── chat.go              # Chat service
│   │   └── queue.go             # Job lifecycle + dedup
│   └── router/router.go         # Chi routes
├── data/prompts/bond_analyst.txt # System prompt для аналитического индекса
├── frontend/                    # Nuxt 3 (Vue 3 + Bootstrap 5 + Chart.js, custom CSS in assets/css/main.css)
│   ├── nuxt.config.ts           # ssr: false → SSG SPA
│   ├── Dockerfile               # Multi-stage: builder (node) → runtime (nginx со статикой)
│   ├── nginx.conf               # Раздаёт статику + проксит /api/, /health → api контейнер
│   ├── layouts/default.vue      # Sidebar layout, dark mode toggle, footer с дисклеймером
│   ├── pages/                   # index (→ redirect), login, chat, favorites, tools, bonds/*, issuers/[id]
│   ├── components/              # System (Panel/KPI/InfoRow/Tag/Pill/TabBar/RatingBadge/AiScore/YieldBar/RangeRow/PageHead/ViewToggle), Domain (BondHero/BondInfoBasic/BondTable/Bond*Tab/IssuerCardGrid/IssuerFilters/IssuerProfile)
│   └── composables/             # useApi, useFormat, useRating(+.spec), useIssuerFilters, useAuth, useFavorites
├── docker-compose.yml           # 3 services: api, postgres, frontend
├── Dockerfile                   # Multi-stage Go build (alpine runtime)
├── handoff/                     # Violet redesign package — DESIGN.md, preview/, tokens.css
└── docs/                        # STATUS.md, roadmap.md, entities.md, openapi.yaml, redesign-plan.md, redesign-questions.md
```

**Источник истины:**
- по плану работ — `docs/roadmap.md`
- по текущему статусу между сессиями — `docs/STATUS.md` (читать первым после ресета)
- по схеме БД — `internal/database/migrations/*.sql` (живёт в коде)

## API Endpoints

```
GET  /health                          — Health check

# Auth
POST /api/v1/auth/register            — Register user
POST /api/v1/auth/login               — Login, returns JWT
GET  /api/v1/auth/me                  — Current user (JWT required)

# Bonds (public)
GET  /api/v1/bonds                    — List bonds (page, per_page, sort)
GET  /api/v1/bonds/monthly            — Monthly coupon bonds
GET  /api/v1/bonds/{secid}            — Bond detail
GET  /api/v1/bonds/{secid}/coupons    — Coupon schedule
GET  /api/v1/bonds/{secid}/history    — Price history (candles)

# Dohod.ru Details (public, async)
GET  /api/v1/bonds/{secid}/dohod      — Dohod.ru analytics (cached or enqueue)

# AI Analysis (public)
POST /api/v1/bonds/{secid}/analyze    — Start AI analysis (async)
GET  /api/v1/bonds/{secid}/analyses   — List analyses for bond
GET  /api/v1/bonds/{secid}/analysis-stats — Analysis statistics
GET  /api/v1/analyses/{id}            — Single analysis by ID

# AI Analysis (public)
DELETE /api/v1/analyses/{id}          — Delete analysis

# Jobs & Queue
GET  /api/v1/jobs/{id}                — Job status (polling)
GET  /api/v1/queue/stats              — Queue statistics

# Ratings (public)
GET  /api/v1/ratings                  — All ratings grouped by emitter_id
GET  /api/v1/ratings/{emitter_id}     — Ratings for specific emitter

# Favorites (JWT required)
GET    /api/v1/favorites              — List user's favorites
POST   /api/v1/favorites/toggle       — Toggle favorite (add/remove)
GET    /api/v1/favorites/check        — Check multiple secids (?secids=A,B)
POST   /api/v1/favorites/{secid}      — Add to favorites
DELETE /api/v1/favorites/{secid}      — Remove from favorites
```

## Credit Ratings Architecture

```
dohod.ru fetch → DetailsService.FetchAndSave()
                        ↓
              updateRatingsFromDohod()
                        ↓
              bond_issuers.GetBySecid(secid) → emitter_id
                        ↓
              issuer_ratings table  (PK: emitter_id + agency)
                        ↓
              backfill emitter_name → bond_issuers.UpdateEmitterName()

MOEX CCI API → SyncMissingRatingsFromMoex() (fallback)
                        ↓
              /iss/cci/rating/companies/ecbd_{EMITTER_ID}.json
                        ↓
              issuer_ratings table (same shape)
```

- Ratings stored by `emitter_id` (int64), NOT by name
- `emitter_id` resolved from `bond_issuers` (MOEX disclosure API)
- **Two scores per record:** `Score` (legacy 1-10, kept for API/frontend filters), `ScoreOrd` (canonical 1-22). Use `ScoreOrd` for any sorting / cross-agency comparison — `Score` collapses BBB- and BB+ into the same bucket
- Normalisation lives in `service/rating_score.go::NormalizeRating(text) → (ord, tier)` — handles АКРА `AAA(RU)`, Эксперт РА `ruAAA`, НКР `AAA.ru`, НРА `AAA|ru|`, Moody's `Aaa/Baa1/...`, S&P/Fitch bare letters, ДОХОДЪ numeric `7`/`7/10`, outlook stripping. **TS port lives in `frontend/composables/useRating.ts`** — keep both implementations in sync, mirrored test tables in `rating_score_test.go` and `useRating.spec.ts`
- On API startup `RatingService.RecomputeAllScores` re-runs normalisation across all stored ratings (idempotent)
- Agencies: АКРА, Эксперт РА, НКР, НРА, Fitch, Moody's, S&P, ДОХОДЪ
- MOEX CCI adds: НКР, НРА agencies (not available from dohod.ru)
- API `/ratings` returns `map[emitter_id_string]IssuerRatingResponse`
- Frontend maps directly by `emitter_id` (no fuzzy matching)
- Аналитический индекс (Phase 2, в работе): 0-100 float64. Сейчас в `bond_analyses.rating` живёт legacy LLM-балл `[RATING:XX.X]`; в Phase 2 рядом появятся `bond_scores` (детерминированные баллы по 3 профилям)
- `ratingChipStyle()` in `useFormat.ts` (delegates to `useRating.ratingTierStyle`) — colours by canonical tier so Moody's `Baa1` is BBB-yellow, not B-red. Tiers: AAA→green, AA→blue, A→teal, BBB→yellow, BB→orange, B→red, CCC/CC/C/D→deep red
- Issuer rating filters use 22-level ordinal via `useRating.maxRatingOrd(rating.ratings)` + `ordMatchesBucket`. Buckets: `aaa`, `aa`, `a`, `bbb`, `bb`, `b_below`, `none`

## Issuer Sync Architecture

```
API startup → goroutine:
  1. BondService.SyncMissingIssuers() (2 min timeout)
     getAllBonds() vs IssuerRepo.GetAllSecids() → MOEX GetDisclosure() → Upsert
  2. BondService.SyncMissingRatingsFromMoex() (3 min timeout)
     IssuerRepo.GetAll() vs RatingRepo.GetDistinctEmitterIDs() → MOEX CCI API → Upsert ratings
```

## Architecture Pattern

```
              ┌─ Frontend (Nuxt 3 SSG SPA) ─┐
              │  pages → composables          │
              │      ↓ fetch /api/v1          │
              └────────────┬──────────────────┘
                           ↓ same container nginx
                           ↓ /api/* → upstream api
              Handler → Service → Repository → Postgres
                 ↓        ↓
               JSON    Business     ↘ Queue worker (goroutine)
                       logic        ↘ MOEX in-memory cache
                                    ↘ OpenAI / MOEX ISS / dohod.ru
```

- **Handler**: HTTP layer, request parsing, response writing
- **Service**: Business logic, validation, orchestration. Кэши тоже здесь (in-process для MOEX в `BondService.cache`)
- **Repository** (`internal/repository/`): Postgres queries only via pgx
- **Model**: Data structures, request/response types

## Style

- Go idioms: error wrapping, context propagation
- No ORM — raw SQL via pgx for PostgreSQL
- JSONB для гибких payload'ов (dohod_details.data, bond_analyses.json_data, queue_jobs.data, scoring_profiles.weights)
- Chi router (stdlib-compatible, middleware-friendly)
- Multi-stage Docker builds (builder → alpine runtime для API, nginx-alpine для фронта)

## ASH → NLA Migration Status

Original project: `/Users/maximlomaev/dockers/src/ash` (Yii2 PHP)
Reference UI: `http://postroika.test:8081/bond`

### Feature Parity

| Feature | ASH | NLA | Status |
|---------|-----|-----|--------|
| Bond list (table, sort, paginate) | ✅ | ✅ | Done |
| Bond detail — 8 tabs | ✅ 8 tabs | ✅ 8 tabs | Done |
| Basic: Params + Финансы | ✅ | ✅ | Done |
| Basic: Даты + прогресс жизни | ✅ | ✅ | Done |
| Basic: Купонные параметры | ✅ | ✅ | Done |
| Trading: Stat cards + price bars | ✅ | ✅ | Done |
| Trading: Глубина стакана | ✅ | ✅ | Done |
| Trading: Торговые данные за день | ✅ | ✅ | Done (v0.6) |
| History: Chart + stats | ✅ combo | ✅ combo | Done — volume bars + nominal (v0.6) |
| Yields: 4 cards + comparison | ✅ | ✅ | Done — 5 types + decomposition (v0.6) |
| Coupons: Schedule table | ✅ | ✅ | Done |
| Coupons: Прогноз по годам | ✅ | ✅ | Done (v0.6) |
| AI Analysis: Send + poll + history | ✅ | ✅ | Done |
| External: iframes | ✅ | ✅ | Done |
| Details tab (dohod.ru data) | ✅ | ✅ | Done — async fetch + cache 30d |
| By issuer: cards + filters | ✅ | ✅ | Done |
| Monthly coupons | ✅ | ✅ | Done |
| Credit ratings | ✅ | ✅ | Done |
| Dark theme | ✅ | ✅ | Done |
| Chat | ✅ | ✅ | Done |
| Favorites/Watchlist | ✅ | ✅ | Done (v0.7) — full stack |
| Trading status badges | ✅ | ✅ | Done (existing) |
| Auth UI (login/register) | ✅ | ✅ | Done (v0.7) |
| Tools page | ✅ | ✅ | Done — Markdown formatter + JSON decoder (`pages/tools.vue`) |

### Known Issues
- Bondization endpoint was using wrong MOEX ISS path (fixed: `securities/{secid}/bondization.json`)
- Some bonds show >999% yield (capped at >999% in display)
- Panel-header SVGs needed explicit `w-4 h-4` classes (CSS `@apply` not sufficient in all cases)
- ~34 emitters still show "Эмитент #NNN" (no dohod.ru data fetched yet, will auto-fill)

## Key Features Added (v0.9+)

### Violet redesign (v0.12.0)
- Палитра / spacing / shadow / radius — все через токены `--nla-*` в `frontend/assets/css/main.css`. Меняем значения, не имена. См. `handoff/DESIGN.md` для словаря и `handoff/MIGRATION.md` для пошагового переноса
- Системные SFC под общую систему: `Panel.vue`, `KPI.vue`, `InfoRow.vue`, `Tag.vue`, `Pill.vue`, `TabBar.vue`, `RatingBadge.vue` (credit-rating string), `AiScore.vue` (AI score 0-100). `RatingBadge` делегирует в `useRating.ratingTierStyle`, `AiScore` в `useFormat.aiRatingStyleSoft`
- `composables/useIssuerFilters.ts` — единая логика фильтрации эмитентов / облигаций для `pages/bonds/by-issuer.vue` и `monthly.vue`. Тип `IssuerFilterState` + чистые функции `matchesBond/matchesIssuerRating/matchesIssuerAi`. Не дублируем в страницах
- `BondHero.vue` — шапка карточки бумаги, потребляет `useFormat`/`useAuth`. Кнопка «AI-анализ» эмитит `analyze` (страница переключает таб на `'ai'`)
- `pages/issuers/[id].vue` — карточки эмитентов на главной ссылаются сюда. Профиль через `IssuerProfile.vue`

### MOEX CCI Rating Integration (v0.10)
- `GetCCIRatings()` in moex/client.go — fetches from `/iss/cci/rating/companies/ecbd_{ID}.json`
- `SyncMissingRatingsFromMoex()` — finds unrated emitters, fetches via CCI API (200ms delay)
- Runs on API startup after SyncMissingIssuers (3 min timeout)
- Added agencies: НКР, НРА (not available from dohod.ru)
- Stats: 474 emitters with ratings, 587 records total

### Rating Display Redesign (v0.10)
- `ratingChipStyle()` moved to useFormat.ts — shared across all components
- IssuerCardGrid: compact rating-only badges (no agency text, agency in tooltip)
- [secid].vue: ratings inline with ISIN/Код/Тип/Валюта row (single flex line)
- IssuerFilters: static rating legend — agency signature examples (AAA(RU)/АКРА, ruA+/Эксперт РА, etc.)
- NULL ratings filtered from display
- "—" badge for emitters without ratings

### Issuer Auto-Sync
- `SyncMissingIssuers()` runs on API startup — new MOEX bonds auto-added to `bond_issuers`
- Emitter names auto-backfilled from dohod.ru via `updateRatingsFromDohod()`
- `IssuerRepo.Upsert()` / `GetAllSecids()` / `UpdateEmitterName()` methods

### AI Analysis Improvements
- `BondAnalysis.Rating` changed from `*int` to `*float64` (supports `[RATING:77.5]`)
- Delete analysis: `DELETE /api/v1/analyses/{id}`
- Markdown table rendering in AI responses (`renderMarkdown()` → `<table>`)

### Queue Reliability
- `ResetStaleJobs()` on worker start — resets "running" jobs older than 3 minutes to "pending"

### Frontend
- Unified rating colors: `aiRatingStyle()`, `aiRatingStyleSoft()`, `ratingChipStyle()`, `issuerRatingBg()` in useFormat.ts
- Markdown tables in BondAiTab and Chat with dark mode support
- Text readability: adjusted `--nla-text-muted` (light: `#64748b`, dark: `#8b9bb5`)
- Background softened: `--nla-bg` light → `#f8f9fb`
