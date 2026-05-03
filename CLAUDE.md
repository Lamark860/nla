# CLAUDE.md — NLA Project Context

## Working preferences

- Persistent context lives **in this file (or other repo docs)**, not in private session memory. If something is worth remembering across sessions, write it here.
- Repo is local-only — never pushed to a remote. Don't escalate "secret leaked into git history" findings.
- Conversational replies in Russian; code/identifiers/paths in English.

## Project

**NLA** — анализатор финансовых инструментов. Новый стек, эволюция проекта ASH.
Go API + PostgreSQL + MongoDB + Redis + Nginx.

## Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| API | Go 1.25 + Chi router | REST API, WebSocket (future) |
| Auth | JWT (HS256) + bcrypt | Stateless authentication |
| Relational DB | PostgreSQL 16 | Users, auth, favorites, portfolios |
| Document DB | MongoDB 7 | Bonds, AI analyses, news, chats |
| Cache/Queue | Redis 7 | Cache, job queue, pub/sub |
| Proxy | Nginx | Reverse proxy, static, WebSocket |

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
# Run all tests
make test

# Verbose
make test-v

# With coverage report
make test-cover

# Specific package
go test ./internal/service/... -v
go test ./internal/middleware/... -v
go test ./internal/handler/... -v
```

### Test structure (61 tests, all passing)
- `internal/service/auth_test.go` — auth unit tests (mock repo, 12 tests)
- `internal/service/analysis_test.go` — rating parser (21 subtests: [RATING:XX.X], итоговая оценка, баллы, bold, /100)
- `internal/middleware/auth_test.go` — JWT middleware (9 tests)
- `internal/handler/auth_test.go` — HTTP handler integration (9 tests)
- `internal/handler/test_helpers_test.go` — shared test mocks
- `internal/client/dohod/client_test.go` — dohod.ru Nuxt payload parser (6 tests)

**Note:** MongoDB-dependent code (IssuerRepo, AnalysisRepo, etc.) and MOEX API clients are not unit-tested — they require integration environment. Current tests focus on pure logic (parsing, auth, middleware).

## API Documentation

OpenAPI 3.0 spec: `docs/openapi.yaml`
View with: https://editor.swagger.io (paste YAML) or any OpenAPI viewer

## Ports

| Service | Host Port | Container Port |
|---------|-----------|---------------|
| Nginx | 8090 | 80 |
| Go API | 8085 | 8080 |
| Frontend (Nuxt 3) | 3000 | 3000 |
| PostgreSQL | 5433 | 5432 |
| MongoDB | 27018 | 27017 |
| Redis | 6380 | 6379 |

**UI:** http://localhost:8090/ (через Nginx → Frontend)
**API:** http://localhost:8090/api/v1/ (через Nginx → Go)

## Project Structure

```
nla/
├── cmd/api/main.go              # Entry point, DI wiring, queue worker, issuer sync
├── cmd/sync-ratings/main.go     # CLI tool for bulk rating sync from dohod.ru
├── internal/
│   ├── config/config.go         # Environment config (DB, Redis, OpenAI, JWT)
│   ├── database/                # DB connections (postgres, mongo, redis)
│   │   └── migrations/          # PostgreSQL schema (*.sql, loaded via embed.FS, lex-order)
│   ├── client/
│   │   ├── moex/client.go       # MOEX ISS API client (5 endpoints + GetDisclosure + GetCCIRatings)
│   │   ├── openai/client.go     # OpenAI client (retry, reasoning models)
│   │   └── dohod/client.go      # Dohod.ru Nuxt SSR parser (HTTP + __NUXT_DATA__)
│   ├── handler/
│   │   ├── auth.go              # Auth handlers + Handler struct
│   │   ├── bond.go              # Bond endpoints (list, detail, coupons, history)
│   │   ├── analysis.go          # AI analysis + job polling + delete
│   │   ├── details.go           # Dohod.ru details endpoint
│   │   ├── rating.go            # Credit rating CRUD endpoints
│   │   ├── chat.go              # Chat handler
│   │   └── favorite.go          # Favorites handler (JWT required)
│   ├── middleware/auth.go       # JWT middleware
│   ├── model/                   # Data models (user, bond, job, chat, dohod)
│   ├── mongo/
│   │   ├── analysis.go          # BondAnalysis MongoDB repo
│   │   ├── details.go           # DohodDetails MongoDB repo (30-day TTL cache)
│   │   ├── issuer.go            # BondIssuer repo (CRUD, Upsert, GetAllSecids, UpdateEmitterName)
│   │   ├── rating.go            # IssuerRating repo (credit ratings by emitter_id)
│   │   ├── chat.go              # Chat sessions/messages repo
│   │   └── queue.go             # QueueJob MongoDB repo
│   ├── queue/worker.go          # Background job worker (goroutine)
│   ├── repository/
│   │   ├── user.go              # PostgreSQL user queries
│   │   └── favorite.go          # PostgreSQL favorites queries
│   ├── service/
│   │   ├── auth.go              # Auth business logic
│   │   ├── bond.go              # BondService + public API + getAllBonds (cache/MOEX fetch)
│   │   ├── bond_parse.go        # parseBond, mergeYieldData (MOEX rows → model.Bond)
│   │   ├── bond_calc.go         # calculateFields, calcRiskCategory, sortBonds, bondScore
│   │   ├── bond_sync.go         # SyncMissingIssuers, SyncMissingRatingsFromMoex
│   │   ├── bond_helpers.go      # extractRows, get*/safe* primitives
│   │   ├── analysis.go          # AI analysis + rating parser
│   │   ├── details.go           # Dohod.ru service (cache + retry + save + rating sync + emitter name backfill)
│   │   ├── rating.go            # Credit ratings CRUD + GetAll by emitter_id
│   │   ├── chat.go              # Chat service
│   │   └── queue.go             # Job lifecycle + dedup
│   └── router/router.go         # Chi routes
├── data/prompts/bond_analyst.txt # AI system prompt
├── frontend/                    # Nuxt 3 (Vue 3 + Tailwind + Chart.js)
│   ├── nuxt.config.ts
│   ├── Dockerfile               # Node 20 Alpine multi-stage
│   ├── layouts/default.vue      # Header, nav, dark mode toggle
│   ├── pages/
│   │   ├── index.vue            # Bond list with sort/pagination
│   │   ├── login.vue            # Login/register
│   │   ├── chat.vue             # AI chat
│   │   ├── favorites.vue        # User favorites
│   │   ├── tools.vue            # Markdown formatter + JSON decoder
│   │   ├── bonds/[secid].vue    # Bond detail with 8 tabs
│   │   ├── bonds/monthly.vue    # Monthly coupons
│   │   └── bonds/by-issuer.vue  # Bonds grouped by issuer
│   ├── components/              # BondTable, BondAiTab, IssuerCardGrid, IssuerFilters, charts, etc.
│   └── composables/             # useApi.ts, useFormat.ts, useAuth.ts, useFavorites.ts
├── nginx/nginx.conf             # Reverse proxy (/ → frontend, /api → Go)
├── docker-compose.yml           # 6 services
├── Dockerfile                   # Multi-stage Go build
└── docs/                        # architecture.md, entities.md, api-plan.md, openapi.yaml, roadmap.md (open work)
```

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
              issuer_ratings collection (key: emitter_id + agency)
                        ↓
              backfill emitter_name → bond_issuers.UpdateEmitterName()

MOEX CCI API → SyncMissingRatingsFromMoex() (fallback)
                        ↓
              /iss/cci/rating/companies/ecbd_{EMITTER_ID}.json
                        ↓
              issuer_ratings collection (same format)
```

- Ratings stored by `emitter_id` (int64), NOT by name
- `emitter_id` resolved from `bond_issuers` (MOEX disclosure API)
- **Two scores per record:** `Score` (legacy 1-10, kept for API/frontend filters), `ScoreOrd` (canonical 1-22). Use `ScoreOrd` for any sorting / cross-agency comparison — `Score` collapses BBB- and BB+ into the same bucket
- Normalisation lives in `service/rating_score.go::NormalizeRating(text) → (ord, tier)` — handles АКРА `AAA(RU)`, Эксперт РА `ruAAA`, НКР `AAA.ru`, НРА `AAA|ru|`, Moody's `Aaa/Baa1/...`, S&P/Fitch bare letters, ДОХОДЪ numeric `7`/`7/10`, outlook stripping
- On API startup `RatingService.RecomputeAllScores` re-runs normalisation across all stored ratings (idempotent)
- Agencies: АКРА, Эксперт РА, НКР, НРА, Fitch, Moody's, S&P, ДОХОДЪ
- MOEX CCI adds: НКР, НРА agencies (not available from dohod.ru)
- API `/ratings` returns `map[emitter_id_string]IssuerRatingResponse`
- Frontend maps directly by `emitter_id` (no fuzzy matching)
- AI ratings: 0-100 float64 scale (parsed from `[RATING:XX.X]`)
- `ratingChipStyle()` in useFormat.ts — color by tier: AAA→green, AA→blue, A→teal, BBB→yellow, BB→orange, B/C/D→red

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
                    ┌─ Frontend (Nuxt 3 SSR) ─┐
                    │  pages → composables     │
                    │      ↓ fetch /api/v1     │
                    └──────────────────────────┘
                              ↓ Nginx
Handler → Service → Repository/Repo → Database
   ↓         ↓         ↗
 JSON    Business    Queue Worker
         logic       (goroutine)
              ↓
         OpenAI / MOEX ISS
```

- **Handler**: HTTP layer, request parsing, response writing
- **Service**: Business logic, validation, orchestration
- **Repository**: Database queries only
- **Model**: Data structures, request/response types

## Style

- Go idioms: error wrapping, context propagation, interface-based DI
- No ORM — raw SQL via pgx for PostgreSQL
- Official driver for MongoDB
- Chi router (stdlib-compatible, middleware-friendly)
- Multi-stage Docker builds (builder → alpine runtime)

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
