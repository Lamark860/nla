# Roadmap — план развития NLA

Документ — единая точка для текущего плана работ. Перезаписывается целиком при смене направления.

Pet-инструмент с потенциалом превращения в SaaS для квалифицированных частных инвесторов в облигации. Продуктовое ядро — детерминированный **scoring-движок с тремя профилями риска** (Низкий / Средний / Повышенный) + LLM как объяснитель факторов. Алерты и трейдинг-сигналы — out of scope (инструмент анализа, не торговый).

## Терминология (с Фазы 0)

- ✅ «Аналитический индекс» — 0..100, нейтрально, регуляторно безопасно
- ❌ «AI-балл», «AI-оценка», «AI-рейтинг», «оценка для покупки» — вне UI
- Профили: «Низкий риск» / «Средний риск» / «Повышенный риск» (внутр. коды `low` / `mid` / `high`)
- Дисклеймер обязателен на каждой странице с индексом

## Сделано (для контекста)

- **v0.10.1 (2026-05-03)** — уборка артефактов, миграции в `embed.FS`, распил `service/bond.go` 1073→354
- **v0.11.0 (2026-05-03)** — нормализация рейтингов на 22-уровневую шкалу, BBB-/BB+ collapse исправлен, миграция 614 рейтингов
- **v0.11.1 (2026-05-06)** — TS-нормализатор `useRating.ts`, vitest 117 кейсов, typecheck 0
- **v0.12.0 (2026-05-07)** — Violet redesign: токены, JetBrains Mono, системные SFC
- **v0.13.0 (2026-05-08)** — Phase 2 catch-up: BondCouponsTab/IssuerProfile/BondAiTab/BondTradingTab/IssuerCardGrid
- **v0.14.0–14.7 (2026-05-08–09)** — Sidebar layout, page-head, collapse длинных списков, Hero расширен

---

## Фаза 0 — терминология + доки (~1 день)

**Цель:** убрать «AI-оценка» из UI, удалить устаревшие доки, зафиксировать новый план.

- [x] Удалить `docs/architecture.md` и `docs/api-plan.md` (устарели: Redis-queue/Selenium-парсеры в коде нет)
- [x] Подчистить `docs/entities.md` — удалены несуществующие коллекции `bond_details` и `emitter_details`
- [x] Переписать этот `docs/roadmap.md`
- [x] Frontend rename `AI-оценка`/`AI-балл`/`AI Анализ` → `Аналитический индекс`/`Индекс` в:
  - `BondHero.vue`, `BondAiTab.vue`, `IssuerCardGrid.vue`, `IssuerProfile.vue`, `IssuerFilters.vue`, `pages/bonds/[secid].vue`, `nuxt.config.ts`
- [x] Дисклеймер в `layouts/default.vue` footer
- [ ] Обновить CLAUDE.md — fix описание Mongo/Redis, убрать ссылки на удалённые доки

## Фаза 1 — инфра-упрощение (3-4 дня)

**Цель:** один Postgres вместо Mongo+Redis, статичный фронт. 3 контейнера вместо 6.

**Шаги:**

1. **Mongo → Postgres JSONB.** Новые таблицы (`bond_analyses`, `bond_issuers`, `dohod_details`, `issuer_ratings`, `queue_jobs`, `chat_sessions`, `chat_messages`) — реляционные ключи + jsonb для произвольных полей. TTL заменить на nightly `DELETE WHERE created_at < now() - interval '30 days'`. Один скрипт переноса `cmd/migrate-mongo-pg/`.
2. **Переписать `internal/mongo/*.go`** (~840 строк) на pgx. Имена методов сохранить, чтобы сервисы не трогать.
3. **Заложить таблицы под Фазу 2:**
   ```sql
   scoring_profiles (id, code, name, weights JSONB, is_preset, user_id NULL)
   bond_scores (id, secid, profile_code, score, breakdown JSONB, computed_at)
   bond_score_explanations (id, bond_score_id, llm_model, text, created_at)
   ```
4. **Redis выкинуть.** Заменить два call-site'а в `service/bond.go` на `sync.Map`/`ristretto` + `time.AfterFunc`. Убрать `database/redis.go`, `cfg.RedisAddr`, redis-сервис из compose.
5. **Nuxt SSR → SSG.** `ssr: false` + `nuxt generate`. Nginx раздаёт статику. Контейнер frontend исчезает из runtime.
6. **docker-compose**: `api` + `postgres` + `nginx`.
7. Обновить `entities.md` под Postgres-схему.

## Фаза 2 — Scoring engine (3-4 дня) ★ ключевая

**Цель:** детерминированный движок, считающий аналитический индекс 0..100 по 12 факторам, с тремя пресетами весов.

### Структура

```
internal/scoring/
├── factors.go        # 12 extract-функций: (bond, issuer, dohod) → float64
├── normalize.go      # value → 0..100 по табличным диапазонам
├── engine.go         # Compute(bond, profile) → ScoreResult
├── presets.go        # Low / Mid / High pre-defined weight sets
└── engine_test.go    # покрытие всех факторов + 3 пресета
```

### 12 факторов (стартовые веса)

| # | Фактор | Источник | Low | Mid | High |
|---|---|---|---|---|---|
| 1 | Кредитный рейтинг (`score_ord` 0-22) | `issuer_ratings.max(score_ord)` | 0.40 | 0.25 | 0.10 |
| 2 | YTM | `bond.effective_yield` | 0.05 | 0.20 | 0.30 |
| 3 | Премия YTM к ОФЗ той же дюрации | OFZ benchmark | 0.05 | 0.15 | 0.25 |
| 4 | Дюрация (короче = лучше для low) | `bond.duration` | 0.15 | 0.10 | 0.05 |
| 5 | Ликвидность (avg оборот 30 дн) | history aggregate | 0.15 | 0.10 | 0.05 |
| 6 | Категория (ОФЗ/корп/ВДО/субфед) | derived | 0.05 | 0.05 | 0.05 |
| 7 | PUT-оферта <90 дн (штраф) | `bond.offerdate` | −0.05 | 0 | 0 |
| 8 | Размер эмиссии (отсечка для high) | `bond.issuesize_placed × facevalue` | 0.05 | 0.05 | 0.05 |
| 9 | Купонный тип (фикс vs флоат) | `is_float`/`is_indexed` | 0.05 | 0.05 | 0.05 |
| 10 | Возраст рейтинга (старее 6 мес — штраф) | `issuer_ratings.updated_at` | 0.05 | 0.05 | 0.05 |
| 11 | dohod.quality | `dohod_details.quality` | 0 | 0 | 0.05 |
| 12 | dohod.stability | `dohod_details.stability` | 0 | 0 | 0.05 |

**Калибровка** — после первого запуска на ~100 бумагах. Веса пишутся в `scoring_profiles.weights JSONB`, легко править без редеплоя.

### API

```
GET  /api/v1/bonds/{secid}/score                    — все три индекса с breakdown
GET  /api/v1/bonds/{secid}/score?profile=mid        — конкретный профиль
POST /api/v1/bonds/{secid}/score/explain?profile=X  — запросить LLM-объяснение (async, как сейчас)
GET  /api/v1/scoring/profiles                       — пресеты + кастомные профили пользователя
```

Реал-тайм пересчёт **дёшев** (без LLM) — кэш в `bond_scores` на сутки, дальше пересчёт. LLM-объяснение асинхронно через очередь, кэшируется в `bond_score_explanations`.

## Фаза 3 — UI трёх профилей (~неделя)

В рамках текущего violet-стиля, без новых макетов:

1. **Глобальный переключатель** профиля в Sidebar — `🛡️ Низкий / ⚖️ Средний / 🚀 Повышенный`. State в localStorage + `useScoringProfile()` composable
2. **Три бейджа на карточке** (`BondHero`, `IssuerCardGrid`) — три `<AiScore>` с tone. Активный профиль bold + подсветка
3. **Новый `BondScoreTab`** (заменяет `BondAiTab` либо рядом):
   - 3 индекса сверху + радар-чарт (Chart.js radar) сравнения
   - Раскрывающийся breakdown по каждому профилю (12 строк с факторами и вкладом в балл)
   - Кнопка «Получить разбор» → LLM-объяснение (кэшируется)
4. **Compare bonds** — `/compare?secids=A,B,C`, до 3 бумаг бок-о-бок с теми же 3 индексами на каждой
5. Сортировки/фильтры на главных страницах пересчитываются под активный профиль

## Фаза 4 — Portfolio (1-2 дня)

**Таблицы:**
```sql
portfolio_positions (id, user_id, secid, qty, price_in, date_in, note)
```

**API:**
```
GET    /api/v1/portfolio                  — все позиции пользователя
POST   /api/v1/portfolio                  — добавить
PUT    /api/v1/portfolio/{id}             — обновить qty/price_in
DELETE /api/v1/portfolio/{id}             — удалить
GET    /api/v1/portfolio/summary          — взвеш. YTM, дюрация, cash flow 12 мес
GET    /api/v1/portfolio/coupons.ics      — iCalendar экспорт выплат
```

**Frontend:** `/portfolio` с таблицей позиций + bar chart cash flow + pie диверсификация по эмитентам/рейтингам + **профильный балл портфеля** (взвешенный по объёму).

## Фаза 5 — Tinkoff events + AI news agent (~3-4 дня)

**Источник:** Tinkoff Invest API (REST/gRPC, токен в `TINKOFF_API_TOKEN`).

**Таблицы:**
```sql
ALTER TABLE bond_issuers ADD COLUMN inn TEXT, external_ids JSONB;
CREATE TABLE issuer_events (id, emitter_id, source, event_type, title,
  summary, url, published_at, raw JSONB, UNIQUE(source, external_id));
CREATE TABLE issuer_news_analyses (emitter_id, summary, hash, updated_at);
```

**События:** coupon / dividend / corporate_action / news. AI-агент новостей (`data/prompts/news_analyst.txt`) даёт структурированное резюме рисков. **Сильные негативные события автоматически снижают балл скоринга** (rating downgrade → штраф к фактору 1, новая просрочка → штраф к фактору 11).

**Frontend:** новая секция «События» в `IssuerProfile.vue` (прошедшие/ожидаемые) или таб в bond-detail.

## Фаза 6 — мелочи (~3-4 дня)

- **Сохранённые скринеры** — `saved_screeners(user_id, name, filters JSONB, is_public)` + UI кнопка «💾 Сохранить» в `IssuerFilters`
- **ical-экспорт купонов** — `GET /api/v1/bonds/{secid}/coupons.ics` + кнопка в `BondCouponsTab` (и в портфеле)
- **Реф-ссылки на брокеров** — `BuyButton.vue` с конфигом 2-3 брокеров, редирект через `/r/{broker}?secid=X` с логированием
- **Markdown/CSV экспорт** — расширить существующий download AI-разбора на breakdown скоринга и таблицы

## Фаза 7 — биллинг (~неделя, last)

- ЮКасса или CloudPayments (решить когда дойдёт)
- `subscriptions(user_id, plan, status, started_at, expires_at)`, `payment_events`
- Middleware `RequirePlan(plan)` на Pro-эндпоинты (`/score/explain`, `/screeners`, `/portfolio/*`)
- Gated UI с тултипом «Доступно в Pro»
- **Кастомный профиль** (12 ползунков весов) — Pro Plus

---

## Open product questions

- **Регуляторика:** financial disclaimer wording — проверить с юристом перед запуском подписки
- **Партнёрка с брокером:** какой первый (Тинькофф / БКС / Финам) — зависит от ставок и условий
- **Mobile UX:** не углублялись, ждёт отдельной итерации
- **Outlook у рейтингов:** нет в API сейчас, добавить парсер если важно для скоринга
- **Multi-level orderbook / trades feed / yield_history:** в Фазе 2 redesign были зафиксированы как backend-блокеры под дизайн. Для текущего scoring-направления **не критично**, можно вечно откладывать

## Open tech debt

- Тесты на репо-слой (после Mongo→Postgres все становятся обычными unit/integration testcontainers, гораздо проще)
- `cmd/sync-ratings/main.go` — CLI tool, после миграции данных через jobs может стать не нужен
- handoff/screens-* — Phase 1/2 baselines уже свою функцию выполнили, можно архивировать
