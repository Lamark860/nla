# CHANGELOG — NLA (ASH → NLA Migration)

## v0.15.0 (2026-05-11) — Фаза 1: один Postgres, ноль Mongo/Redis, статичный фронт

Полный перенос на единое хранилище (PostgreSQL 16) и упрощение стека до 3 контейнеров. Все слои переписаны без изменения публичного API.

### Schema (`internal/database/migrations/0002_postgres_full_schema.sql`)

13 новых таблиц + 3 seed-профиля скоринга:

- **Перенос Mongo-коллекций**: `bond_issuers`, `issuer_ratings`, `dohod_details` (JSONB для ~80 полей dohod.ru), `bond_analyses` (UUID PK), `queue_jobs` (UUID + partial index по `(type,secid)` для дедупа, partial index по `status=running` для stale-reset), `chat_sessions` + `chat_messages` (CASCADE)
- **Заложены под Фазу 2 (scoring engine)**: `scoring_profiles` с PK по `code`, `bond_scores` с историей, `bond_score_explanations` (FK CASCADE). Сразу засеяны три preset-профиля `low`/`mid`/`high` с весами по 12 факторам из `docs/roadmap.md`
- **Заложена под Фазу 4 (portfolio)**: `portfolio_positions`
- Общая trigger-функция `set_updated_at()` на 7 таблицах с `updated_at`

### Repositories (`internal/repository/`)

Новые pg-реализации с **теми же именами методов** что в удалённом `internal/mongo/`, чтобы сервисы не правились:
- `analysis.go` — `Save/GetBySecid/GetByID/Delete/GetStats/GetLatestRatings/GetBulkStats`. Stats считаются нативной SQL-аггрегацией вместо ручного фолдинга, dedup latest через `DISTINCT ON (secid)` ORDER BY timestamp DESC
- `issuer.go` — 9 методов с пары `Upsert`/`GetAllSecids`/`GetOneSecidPerEmitter` (последний через `DISTINCT ON (emitter_id)`)
- `details.go` — Get/GetBySecid/Upsert с 30d TTL check в коде. Тело dohod-данных хранится в `data JSONB`
- `rating.go` — `BulkUpsert` через `pgx.Batch` (одно сетевое roundtrip на N записей)
- `queue.go` — `FetchPending` через atomic `UPDATE ... FOR UPDATE SKIP LOCKED RETURNING` (готов к multi-worker)
- `chat.go` — простая реляционная схема, CASCADE убирает messages при удалении сессии

### Data migration

Одноразовый скрипт `cmd/migrate-mongo-pg/main.go` (закоммичен и удалён, в `git log`) перенёс реальные данные:

| Сущность | Mongo | Postgres |
|---|---|---|
| bond_issuers | 3295 | 3295 |
| issuer_ratings | 631 | 631 |
| bond_analyses | 57 | 57 |
| chat_sessions / messages | 1 / 4 | 1 / 4 |
| dohod_details | 577 | 47 (только свежие <30d; остальные stale были бы перезаписаны TTL) |
| queue_jobs | — | пропущено (ephemeral) |

### Service layer

Импорты `nla/internal/mongo` → `nla/internal/repository` в `service/{analysis,queue,details,rating,bond,chat}.go`. Конкретные типы `*mongorepo.X` → `*repository.X`.

### BondService — Redis убран

Заменён на in-process `memoryCache` (sync.RWMutex + `map[string]{value, expiresAt}` + `Delete`/`DeletePrefix`). Два ключа (`bonds:list`, `bonds:{secid}`) с TTL 24h. Прогрев холодного кэша — один запрос к MOEX при первом обращении. Для горизонтального масштаба позже можно вернуть Redis — точечной правкой.

### Frontend → SSG SPA

- `nuxt.config.ts`: `ssr: false`. `npm run generate` собирает 11 routes + статика SPA
- `frontend/Dockerfile` multi-stage: builder (node 20-alpine) → runtime (nginx:alpine), копирует `.output/public/` в `/usr/share/nginx/html`. Node на runtime больше нет
- `frontend/nginx.conf` (новый, локальный для фронт-контейнера): раздаёт статику, проксит `/api/` и `/health` на api контейнер, SPA-fallback `try_files $uri $uri/ /index.html` для client-side routing
- Корневой `nginx/nginx.conf` удалён — отдельный nginx-контейнер больше не нужен

### docker-compose

3 контейнера вместо 6:

```
nla-api        Go API                  expose 8080 (internal)
nla-postgres   PostgreSQL 16-alpine    5433 → 5432
nla-frontend   nginx + статика         8090 → 80
```

Удалены сервисы `mongo`, `redis`, отдельный `nginx`. Volumes — только `nla-pgdata`.

### Cleanup

- Удалены пакеты `internal/mongo/` (6 файлов, ~840 строк), `internal/database/mongodb.go`, `internal/database/redis.go`, `cmd/sync-ratings/` (CLI терял смысл — функционал в автосинке `SyncMissingRatingsFromMoex` на старте API)
- Из `internal/config/config.go` удалены `MongoURI`, `MongoDB`, `RedisAddr`, `RedisPassword`, `RedisDB`
- `go mod tidy` снял зависимости `go.mongodb.org/mongo-driver` и `github.com/redis/go-redis/v9`

### Tests

- vitest: 117/117 ✓
- go test ./... ✓ (включая `database`, `handler`, `middleware`, `service`, `client/dohod`)
- Smoke вручную: `GET /health` → 200, `GET /api/v1/bonds?per_page=1` → данные, фронт `GET /` → SPA shell 200, API при старте применяет миграции и грузит pre-existing данные

### Progress tracking

Новый файл `docs/STATUS.md` — карта прогресса между сессиями (на случай ресета Claude). Обновляется при переходе фаз.

---

## v0.14.8 (2026-05-11) — Фаза 0: терминология и доки

Переход с pet-просмотровой утилиты на продуктовый «инструмент для квалифицированных частников». Перед переделкой инфры и фичей — переименование пользовательских меток и расчистка устаревшей документации.

### Терминология (UI)

Пользовательский термин стал **«Аналитический индекс»**. «AI-балл», «AI-оценка», «AI-рейтинг», «AI Анализ» как user-facing метки удалены. Технически компонент `AiScore` и API endpoint `analyze` сохранены — internal identifiers не трогаем, регрессия минимальна.

Конкретные правки:

- `frontend/nuxt.config.ts` — meta description: «AI-оценкой» → «аналитическим индексом»
- `frontend/components/BondHero.vue` — кнопка «AI-анализ» → «Анализ»
- `frontend/components/BondAiTab.vue` — Panel title «Анализ облигации через ИИ» → «Аналитический индекс»; «Отправить в ИИ» → «Рассчитать индекс»; «AI анализирует…» → «Считаем индекс…»
- `frontend/components/IssuerCardGrid.vue` — `issuer-bond__ai-lbl` «AI рейтинг» → «Индекс»
- `frontend/components/IssuerProfile.vue` — «AI-рейтинг эмитента» → «Индекс эмитента»
- `frontend/components/IssuerFilters.vue` — Select placeholder «AI-балл» → «Индекс»
- `frontend/pages/bonds/[secid].vue` — таб «AI Анализ» → «Индекс»

### Дисклеймер

В `layouts/default.vue` footer теперь содержит формулировку «Информация на сайте носит аналитический характер и не является индивидуальной инвестиционной рекомендацией». Стили `.app-footer` адаптированы под двухстрочный layout.

### Документация

- **Удалены** `docs/architecture.md` и `docs/api-plan.md` — устарели (упоминали Redis-queue, Selenium-парсеры, отдельный worker-контейнер, ничего из этого в коде нет)
- **Подчищен** `docs/entities.md` — удалены секции `bond_details` и `emitter_details` (несуществующие коллекции). Добавлена сноска о том, что Redis больше не очередь, а кэш на 2 ключа
- **Переписан** `docs/roadmap.md` целиком под новый план: Фазы 0-7 (инфра-упрощение → scoring engine → UI 3 профилей → portfolio → Tinkoff events → мелочи → биллинг). Алерты явно out of scope. Backend-блокеры под редизайн (multi-level OB, trades feed) понижены в приоритете
- **CLAUDE.md** — обновлены Stack table и описание структуры. Mongo/Redis помечены как «в процессе миграции». Добавлен прицельный пункт про терминологию

### Tests

- vitest: 117/117
- go test: все пакеты ok

---

## v0.14.7 (2026-05-09) — UX: длинные списки сворачиваются, фильтры компактнее, шкала volume (Шаг 31)

Серия мелких UX-правок чтобы экранов не было километровых.

### `BondHistoryTab` — переключатель шкалы объёма

- Слева от period-toolbar добавлен seg-control из 3 кнопок: `↓` (компакт, multiplier × 5) / `·` (норма, × 3) / `↑` (широкий, × 1.3). Активная влияет на `yVolume.max` графика
- На «широкой» шкале гистограмма объёмов занимает больше вертикали → видны мелкие бары на тихих днях. На «компактной» наоборот — bars прижаты к низу, цена доминирует
- Дефолт `normal` (× 3) — как раньше

### `BondCouponsTab` — collapse длинных таблиц

- Если купонов > 7 — по умолчанию показываем окно `next ± 4 + maturity` (≈7 строк). Остальное скрыто
- Внизу таблицы кнопка `▼ Показать все 18 купонов · скрыто 11`. Клик раскрывает все, повторный — сворачивает
- Если купонов ≤ 7 — кнопка не появляется, всё видно сразу
- Полезно для длинных бумаг (10+ лет, 20+ купонов) — раньше прокрутка табы убивала

### `IssuerCardGrid` expanded — collapse списка бумаг

- В развёрнутой карточке эмитента показываем максимум 10 бумаг
- Если у эмитента больше — внизу expanded-блока кнопка `▼ Показать все 265 бумаг · скрыто 255`. Состояние раскрытия хранится в `Set<emitter_id>` отдельно от состояния раскрытия карточки эмитента
- Sber (265 бумаг) перестал ломать ритм страницы

### `IssuerProfile` (`/issuers/[id]`) — collapse таблицы бумаг

- То же самое для таблицы «Облигации эмитента» — лимит 10. Кнопка раскрытия в подвале Panel
- Кнопка использует тот же стиль что в Купонах

### `IssuerFilters` — collapsible advanced

- Search и chips теперь видны постоянно (2 ряда вместо 4)
- Dropdowns (Категория/Сектор/Рейтинг/AI-балл) и ranges (Доходность/Купон/Дюрация) свернуты под кнопку **«⇋ Расширенные ▾»** справа от поиска
- Кнопка показывает бейдж с количеством активных advanced-фильтров (например `⇋ Расширенные 2 ▾`)
- Если в advanced что-то активно при загрузке — открывается автоматически (через `watch(immediate)` на advancedActiveCount)
- Фильтры на by-issuer теперь занимают ~3 строки вместо ~5 — больше места для карточек выше fold

### Tests

- typecheck 0
- vitest 117/117

### Скрины

- `31-light-filters-collapsed.png` — фильтры компактные, advanced свернут
- `31-light-bonds-toggle-btn.png` — кнопка «Показать все 265 бумаг» в expanded Sber
- `31-light-coupons-collapsed.png` — окно `next ± 4 + maturity` на бумаге с 18 купонами
- `31-light-history-volscale-wide.png` — bars высоки на «широкой» шкале

---

## v0.14.6 (2026-05-09) — Hero KPI: добор Купон ₽ / Изменение п.п. / «Скоро оферта» (Шаг 30)

Ещё проход по Hero — добавлены поля, которые быстро дают «полный контекст» бумаги без открытия табов.

- **KPI «Купон» sub** теперь содержит и дату следующего купона, и его рублёвую сумму: `след. 08.07.2026 · 37,40 ₽`. Раньше только дата
- **KPI «Изменение» sub** теперь и в п.п. и оборот: `-0,50 п.п. · оборот 368 тыс`. Раньше только оборот, не было числа в пунктах. Полезно для облигаций где % движения не очевиден без знания цены
- **Pill «🚩 Скоро оферта»** в tags row если `is_near_offer === true` (бэк ставит флаг для PUT <90 дней). Tone warning. Сильный сигнал что бумага скоро может быть выкуплена эмитентом

### Tests

- typecheck 0
- vitest 117/117

### Скрин

- `30-light-hero.png` — `RU000A106JC8` (без оферты): купон ₽, изменение п.п.
- `30-light-hero-near-offer.png` — `RU000A103372` (PUT через 3 дня): pill «Скоро оферта» + категория «До оферты (put)»

---

## v0.14.5 (2026-05-09) — Hero: risk-pill + mid-price в Цене (Шаг 29)

Дополнительный проход по Hero — добавлены поля, которые в bond-detail у нас уже есть в табах, но в шапке отсутствовали.

- **Risk-pill в tags row** — после статуса торгов добавлен бейдж `Риск: <tone>` с цветом по уровню (`safe/stable` зелёный, `moderate` violet, `speculative/risky` warning, `toxic/junk` danger). Позволяет понять про бумагу одной строкой без открытия Основного
- **Mid-price в sub под KPI «Цена»** — добавлено `· mid 20.34%` рядом с рублёвой ценой если `mid_price_pct != null`. Полезно когда есть значимый спред bid/offer

### Skipped intentionally

- Доходности `closeyield/effectiveyield/yieldatwaprice/yieldtoprevyield` — уже отображаются в `BondYieldsTab` (под вопросом, не трогаем). В `BondTradingTab` дублировать смысла не было

### Tests

- typecheck 0
- vitest 117/117

### Скрин

`handoff/screens-phase2/29-light-hero.png` — Hero на ВДО `RU000A106JC8` с красной pill «Риск: токсичный» и mid в sub под Ценой.

---

## v0.14.4 (2026-05-09) — добор пропущенных полей в bond-detail табах (Шаг 28)

Прошлись по всем активным табам страницы конкретной облигации, добавили поля которые были в Bond API но нигде не отрисовывались.

### Основное (`BondInfoBasic`)

В «Параметры выпуска» добавлены строки:

- **ISIN** (mono) — раньше виден только в шапке Hero
- **Гос. рег. номер** (`bond.regnumber`, mono) — например `4B02-09-00382-R-001P`
- **Категория риска** — pill с цветом по уровню (`safe/stable` зелёный, `moderate` фиолетовый, `speculative/risky` жёлтый, `toxic/junk` красный, `unknown` серый). Лейблы: «Низкий / Стабильный / Умеренный / Спекулятивный / Высокий / Токсичный / Мусорный»
- **Площадка торгов** (`bond.boardname` или fallback на `boardid`) — `Т+: Облигации - безадрес.`

### Купоны (`BondCouponsTab`)

- **Доходность к оферте** (`bond.yieldtooffer`) добавлена в Panel «Параметры купона». Рендерится только когда есть PUT-оферта И значение != null. Tone primary как соседняя «Купонная доходность»

### Торговля (`BondTradingTab`)

- **Mid-price** добавлена строка в «Объёмы и статистика» (первая, чтобы не теряться). Формат `20.34% · 203,40 ₽` (mid_price_pct + mid_price_rub). Tone primary

### Tests

- typecheck 0
- vitest 117/117

### Скрины

- `28-light-basic.png` — Параметры выпуска с ISIN/regnumber/risk-pill/boardname
- `28-light-trading.png` — Mid-price в Объёмах
- `28-light-coupon-yto.png` — Доходность к оферте 49.92% в Параметрах купона на бумаге с PUT

---

## v0.14.3 (2026-05-09) — даты-каркас всегда + фильтры компактнее (Шаг 27)

- **Оферта / Колл-опцион / Погашение** теперь рендерятся всегда. Если даты нет — рендерим в значении приглушённый прочерк (как handoff `BondRow`). Так список бумаг визуально выровнен по строкам, и сразу видно «у этой бумаги нет PUT-оферты» вместо угадывания. `DateLine` стал устойчив к `date=null`
- **Фильтры** уменьшены без перестройки layout:
  - inputs/selects/range-pair: 38px → 32px высота, шрифты 13.5 → 12.5
  - chips: 32px → 26px, шрифт 12.5 → 11.5
  - filters body padding: 16/18 → 12/14, row gap: 14 → 8
  - stats numbers: 22 → 18, labels мельче. Stats-bar сверху отжимает на 4px меньше
- Все остальные пропорции сохранены, цвета/spacing-tokens не трогали

### Tests

- typecheck 0
- vitest 117/117

### Скрин

`handoff/screens-phase2/27-light-final.png` — сравнить высоту фильтров и видимость «Оферта/Колл-опцион —» в карточках Sber.

---

## v0.14.2 (2026-05-08) — расширили инфу expanded BondRow (Шаг 26)

Внутри развёрнутой карточки эмитента бумаги теперь показывают больше полезных полей:

- **Relative-даты** во всех `<DateLine>` рендерятся через `fmt.daysToMaturity()` — `5 мес`, `1 г 4 мес`, `7 г 10 мес` вместо тупого `+N дн`. Soon-warn (<200 дн) сохранился — подсвечивает близкие события жёлтым
- **Новая date-row «След. купон»** — `next_coupon` + `days_to_next_coupon` сверху списка дат. Полезно для понимания «когда платят»
- **Купон-метрика** получила `:sub` проп — рублёвая сумма купона под процентом (например `13,50%` сверху, `67,31 ₽` снизу мелким моноширинным)
- **Дюрация-tag** в заголовке бумаги — компактно `1833 дн` / `3,5 г`. Помогает быстро отличить короткие выпуски от длинных при пробежке по списку

### Components

- `Metric` (inline-defineComponent внутри `IssuerCardGrid`) расширен пропом `sub: string | null`, рендерит `.metric__sub` под значением
- `DateLine` (там же) — `relText` теперь делегирует в `useFormat.daysToMaturity` вместо ручного формирования
- `durationLabel(days)` helper в скрипте — `<365 дн → N мес`, `≥365 дн → N,N г`. Выдаёт пустую строку если null/0 (Tag не рендерится)

### Tests

- typecheck 0
- vitest 117/117

### Скрин

`handoff/screens-phase2/26-light-bondrow-extended.png` — теперь у каждой бумаги внутри Sber видно: имя + secid + период + дюрация-tag, 4 metrics с купоном ₽ снизу, 5 date-rows (След. купон / Оферта / Колл / Погашение / AI рейтинг).

---

## v0.14.1 (2026-05-08) — IssuerCardGrid expanded BondRow догнал handoff (Шаг 25)

Развёрнутые карточки бумаг внутри эмитента приведены к виду handoff `IssuerCard.BondRow`:

- **Card chevron** теперь `bi-plus-lg` (свернутая) / `bi-dash-lg` (развёрнутая) — handoff-ритм. Раньше был `bi-chevron-down/up`
- **Bond row** обёрнут в `<NuxtLink>` (вся карточка кликабельна), внутри:
  - Head: имя + secid + period tag + флоатер/индекс tag, **arrow-icon `bi-arrow-up-right` справа** (handoff bond-link). При hover карточки иконка подсвечивается primary
  - Hover bg → `--nla-bg-card`
- **Metrics**: 4-cell **boxed grid** с разделителями (border 1px + radius 8 + bg-card) вместо плоской сетки. Порядок переиздан Цена / Доходн. / НКД / Купон (handoff)
- **Dates**: новый ритм 3-кол `84px / 1fr / auto`:
  - Лейбл UPPERCASE мелкий
  - Дата моно
  - Relative `+138 дн` со знаком (вместо «183 дн»)
  - **Soon-warn**: если 0 ≤ days < 200 → `--nla-warning` цвет + bold (handoff `.date-rel.soon`)
  - Порядок: Оферта → Колл-опцион → Погашение → AI рейтинг (последняя строка)
- **AI рейтинг** — больше не отдельный блок снизу, а 4-я date-row в общем ритме. Pill в primary tone (`--nla-primary-light` фон, dot + число + `/100` приглушённый)
- **Style**: `.issuer-card__bonds` теперь грид без отступов, прижимается к границам карточки `margin: 12px -20px -16px` с `border-radius` снизу — выглядит как один сплошной список вместо разрозненных «коробочек»

### Other

- `daysFromTodayTo()` helper в `IssuerCardGrid` для расчёта `+N дн` от putoptiondate / calloptiondate (раньше `days_to_put` / `days_to_call` часто null'ы из API)

### Tests

- `npx nuxi typecheck` — 0 ошибок
- `npm test` — 117/117

### Скрин

`handoff/screens-phase2/25-light-expanded-bonds-detail.png` — полный вид раскрытой Sber-карточки с 7 бумагами.

---

## v0.14.0 (2026-05-08) — Sidebar layout + унификация главных страниц (Шаги 20-24, Phase 3)

Главные «лендинги» приложения (`/bonds/by-issuer` + `/bonds/monthly`) выровнены под `handoff/preview/index.html`. Sidebar 240px вместо top-navbar, page-head с view-toggle между двумя роутами, stats+legend в подвал фильтров, redirect старого `/`.

### Layout (Шаг 20) — `layouts/default.vue`

- **App shell**: CSS-grid `240px 1fr`, sticky sidebar `top:0;height:100vh` (свой скролл), main column 1fr
- **Сайдбар**:
  - Brand: квадрат «N» с акцентом + «NLA · bonds»
  - Section labels: КАТАЛОГ (Эмитенты / Месячные купоны / AI-чат), ИНСТРУМЕНТЫ (Плоский список / Утилиты)
  - Bottom-cluster: Избранное (+ счётчик) / user-info / logout / theme-toggle
  - Active state по `route.path` через `is-active` класс с `--nla-primary-light`
- **Mobile (≤768px)**: sidebar превращается в drawer, открывается через hamburger в `mobile-bar` (sticky сверху). Backdrop затемняет content. Закрытие по route change или клику на backdrop
- Старые правила `#header .navbar`, `#footer`, `.theme-toggle`, `main > .container` из `assets/css/main.css` удалены (заменены scoped-стилями в layout)

### Page-head на by-issuer/monthly (Шаги 21-22)

- **`PageHead.vue`** — переиспользуемый компонент с props `title`, `sub`, slots `sub`/`actions`. Структура `flex+wrap`, на мобиле stack
- **`ViewToggle.vue`** — segmented control из `<NuxtLink>`, активный пункт по route.path. Используется в actions slot для переключения по-эмитентам/месячные
- В `pages/bonds/by-issuer.vue` и `pages/bonds/monthly.vue`:
  - Старая шапка с h1 + summary-line удалена
  - Новый PageHead с sub-line «Сгруппировано по эмитентам · N компаний · M выпусков» / «Только бумаги с ежемесячным купоном · …»
  - ViewToggle справа [Эмитенты | Месячные купоны]
  - Сводка перенесена внутрь `IssuerFilters` через `:stats` prop

### `IssuerFilters` — stats-bar + legend в footer (Шаг 22)

- Принимает `stats={ issuers, bonds, shown }` через props
- Под `.filters__row--chips` рендерится `.filters__stats` — grid `repeat(3, 1fr) 2.6fr`:
  - 3 stat-cells (эмитентов / облигаций / показано) с моно-числом и uppercase-лейблом
  - 4-я ячейка — `.filters__legend` 3-кол сетка с 6 чипсами справочных рейтингов: AAA(RU)/АКРА, ruAAA/Эксперт РА, AAA.ru/НКР, AA/ДОХОДЪ, Baa1/Moody's, BB|ru|/НРА. Цвета бэйджей по канонической tier-шкале (aaa зелёный → bb красный)
- На ≤992px legend wrap'ится на новую строку

### Redirect / + флэт-список (Шаг 23)

- **`pages/index.vue`** заменён на redirect: SSR `await navigateTo('/bonds/by-issuer', { redirectCode: 302 })` + onMounted-фолбэк на клиенте
- **`pages/bonds/flat.vue`** — старый функционал `pages/index.vue` (BondTable с пагинацией/сортировкой). Перешёл на новый `<PageHead>`. Доступен из сайдбара пунктом «Плоский список»
- Внешние ссылки на `/` теперь автоматически уезжают на `/bonds/by-issuer`

### Tests / контроль

- `npx nuxi typecheck` — 0 ошибок
- `npm test` — 117/117 vitest
- `go test ./internal/... -count=1` — все пакеты ok

### Скрины

`handoff/screens-phase2/`: 20-light-sidebar / 20-dark-sidebar / 20-mobile-sidebar-open + closed / 21-22-light-by-issuer / 21-22-light-monthly / 23-light-flat.

### Migration notes

- **Если у вас в коде есть ссылки на `/`** — они продолжат работать (302 redirect), но при возможности обновите на `/bonds/by-issuer` чтобы избежать лишнего hop
- **`/bonds/flat`** — новая ручка для старого Bootstrap-табличного вида (раньше был `/`)
- Layout-стили теперь scoped в `layouts/default.vue`. Если понадобится переопределение — через `--nla-*` токены

---

## v0.13.0 (2026-05-08) — Phase 2 redesign catch-up (Шаги 13-19)

Полный pass по табам и страницам, которые остались в старой вёрстке после Phase 1. Сделано в одной сессии: план в `docs/redesign-plan.md` (раздел Phase 2), backend-блокеры в `docs/roadmap.md` (Phase D), baseline-скрины в `handoff/screens-phase2/`.

### `BondCouponsTab` гибрид (Шаг 14)

Перенос `handoff/preview/bond-detail.html → CouponsTab` с сохранением «Параметров купона» и «Прогноза по годам».

- **4-cell summary** в шапке таблицы: Выплачено (sum × paid), Следующий купон (date + days + сумма), PUT-оферта (date + days), К погашению (sum × ¬paid + facevalue)
- **Status pills** в таблице через `<Pill>` — статус каждого купона вычисляется на фронте по `coupon_date`, `bond.next_coupon`, `bond.offerdate`, `bond.matdate`. Состояния: `paid` (мутед серый), `next` (primary, violet), `put` (warning, оранжевый), `maturity` (success, зелёный), `future` (outline, без точки)
- Активная строка `next` подсвечена `--nla-primary-light`
- Удалена секция «Ключевые события» — её данные переехали в 4-cell summary
- «Параметры купона» обёрнуты в `<Panel title icon>` с `<InfoRow mono>`. Раздел «Прогноз по годам» — в `<Panel>` с `<KPI>` × 4 + таблица progress

### `IssuerProfile.vue` догнать handoff `IssuerTab` (Шаг 15)

Полная переработка карточки эмитента на странице `/issuers/[id]`. Сейчас содержит ровно ту же информационную плотность, что handoff `IssuerTab`:

- **Header** с буквенным логотипом (детерминированный hue по hash имени), name + sub-line (ID + сектор + ИНН + ОГРН), теги, кнопка «Все эмитенты»
- **4-cell stats**: Облигаций в обращении (с подразделом N флоатеров / M фикс.), Общий объём (`sum(facevalue × issuesize_placed)` в трлн/млрд/млн ₽), Средняя YTM (взвешенная по объёму, фильтрация выбросов >100%), AI-рейтинг через `<AiScore>`
- **Bonds-table** с колонками SECID / Название / Погашение / Купон / Цена / YTM, подсветка текущей бумаги через `?current=secid` (использует `--nla-primary-light`). `BondHero` теперь генерирует ссылку с этим query-параметром
- **Кредитные рейтинги** — карточки 2-кол с agency / grade / updated date. `outlook` отсутствует в `IssuerRating` модели, поэтому колонку пропустили (документировано в Phase 2 план)
- **Качество эмитента (dohod.ru)** — quality bars 4 строки (Кредитное качество = `dohod.quality`, Финансовая стабильность = `stability`, Баланс = `quality_balance`, Прибыльность = `quality_earnings`) + Сводный балл = `best_score`. Если поле null — скрываем строку. Если dohod не загружен — скрываем весь блок. + `description` снизу
- В `pages/issuers/[id].vue` добавлена fetch dohod-данных по первому из бумаг эмитента (issuer-level fields в dohod одинаковы для всех бумаг одного эмитента)

### `BondAiTab` full redesign (Шаг 16)

Перенос `handoff → AiTab` 2-колоночного layout. Логика polling/markdown/удаления не тронута.

- **`.ai-grid` 2-кол** (1fr 380px) на десктопе, stack ≤992px
- **Form panel** с тремя блоками: JSON-preview облигации (с message-индикатором о наличии dohod), textarea для дополнительного JSON, violet submit-кнопка
- **Job polling** — анимированный spinner + текст-статус (pending / running / error)
- **Result panel** — кастомный `<template #head>` с `<AiScore>` + время + clipboard / download / close icon-кнопки. Body — рендеренный markdown. Clipboard копирует raw markdown, download сохраняет `.md` файл с именем `{secid}-{date}.md`
- **History sidebar** — `<ul>`-список с custom `.ai-hist-item` (active state с `--nla-primary-light`, hover, preview-текст 160 chars)
- **Stats sidebar** — 2-cell с числом анализов + `<AiScore>` среднего рейтинга

### `BondTradingTab` редизайн из агрегатов (Шаг 17)

Multi-level orderbook + лента сделок отсутствуют в API (зафиксировано в `docs/roadmap.md` Phase D). Редизайн использует только то что отдаётся через `/bonds/{secid}` (best bid/offer + biddeptht/offerdeptht + numbids/numoffers + bid_offer_ratio).

- **4 KPI top**: Статус (tone success/muted), Bid (% + ₽ + глубина), Ask (% + ₽ + глубина), Спред (п.п. + % от бида)
- **2 Panel side-by-side**: «Цены торгового дня» с `<PriceBar>` × 6 + «Глубина стакана» (visual bid/ask баланс с цветными барами и dual values)
- **2 Panel** с `<InfoRow mono>`: «Торговые данные за день» (open/low/high/prevprice/change/wap) + «Объёмы и статистика» (numtrades/vol/value/avgsize/totaldepth/ratio/numbids/numoffers). Дублей нет — каждое поле отдаётся ровно в одном месте
- **«Временные метки»** в `<Panel flush>` 3-cell на десктопе, stack на мобиле

### `IssuerCardGrid` точечный catch-up (Шаг 18)

Мелкие расхождения с handoff `IssuerCard`:

- Preview-строка collapsed карточки: добавлен YTM первой бумаги (`{shortname} +{yield}%` в success-tone), фильтрация выбросов
- Summary: «N облигаций в обращении» вместо «N облигаций»

### Components

- **`RangeRow.vue`** — раньше расширен `tone` пропом (Шаг 12)
- **`assets/css/main.css`** — `.yield-bar__fill--muted` (Шаг 12)
- **Pill** уже использовался — расширений не потребовалось
- **KPI / InfoRow / Panel** — без изменений, использованы как есть

### Tests / контроль

- `npx nuxi typecheck` — 0 ошибок (зелёный после каждого шага 14-18)
- `npm test` — 117/117 зелёные на каждом шаге
- `go test ./internal/... -count=1` — все пакеты ok

### Скрины

`handoff/screens-phase2/` — 21 baseline-скрин (Шаг 13) + по каждому Шагу 14-18 light/dark/mobile + спец-кейсы (PUT-оферта, non-trading bond, два эмитента: Royal Capital с dohod и Sber без dohod).

### Open work (зафиксировано в roadmap.md Phase D)

- MOEX orderbook endpoint (5+5 уровней) — для multi-level стакана handoff
- MOEX trades feed — для ленты сделок handoff
- yield_history в `/bonds/{secid}/history` — для mode-toggle Цена/Доходность в `BondHistoryTab`
- OFZ benchmark линия в графике (опционально)

---

## v0.12.1 (2026-05-08) — BondHistoryTab гибрид (Шаг 12)

### Frontend
- **`BondHistoryTab.vue` rewrite** — структура из `handoff/components/BondHistoryTab.vue`, рич-инфа из текущей версии сохранена. Сверху вниз:
  1. KPI-row × 4 (`<KPI>`): цена закрытия, изменение за период (с tone success/danger), волатильность, средний объём
  2. `<Panel flush>` с period-toolbar (1Н / 1М / 3М / 6М / 1Г / Всё, дефолт 1М) и `RangeRow` × 4 в подвале (`hist-ranges` grid 4-кол с border-left)
  3. `<Panel>` «Баланс торговых дней» — три yield-bar'a (рост/падение/без изменений)
  4. `.hist-tables` grid 2-кол: `<Panel>` «Статистика за период» + `<Panel>` «Изменение цены», все значения через `<InfoRow mono>`
- **Period-фильтр** — фронт-only, без backend-параметра. `filteredHistory` отрезает по дате, все computed (`stats`, `lastClose`, `volatility`, `dayChanges`, `totalVolume`, `avgVolume`) и сам Chart.js перерисовываются на смену периода. Если в окне нет точек (период длиннее данных) — fallback на полную историю + подсветка автоматически прыгает на «Всё» через `effectivePeriod` вычисляемое
- Mode-toggle Цена/Доходность из handoff отложен — yield-history не отдаётся API (отдельный backend-тикет)

### Components
- **`RangeRow.vue` extended** — добавлен опциональный `tone` проп (`primary | success | danger | muted`, дефолт `primary`); цвет fill вычисляется через `:class="`yield-bar__fill--${tone}`"`. Существующие 5 вызовов в `BondYieldsTab` остаются на дефолтном `primary`
- **`assets/css/main.css`** — добавлен `.yield-bar__fill--muted` (через `--nla-text-muted`) под новый tone

### Tests
- `npx nuxi typecheck` — 0 ошибок
- `npm test` — 117/117 vitest зелёные
- Visual regression: `handoff/screens-step12/` — baseline (live + handoff/preview), результат на dark/light/desktop/mobile (414×896), включая переключение периодов 1М ↔ 3М

---

## v0.12.0 (2026-05-07) — Violet redesign

Перенос дизайн-пакета `handoff/` в `frontend/`. Контракт `--nla-*` сохранён, Bootstrap 5 не выпиливался.

### Visual
- **Палитра violet** — `--nla-primary` `#5b3aa8` (light) / `#b89cff` (dark), тёплый off-white фон light-темы (`#faf8f5`), тёплый почти-чёрный dark (`#0c0e0d`)
- **Новые токены** под тем же префиксом: `--nla-primary-ink`, `--nla-bg-elevated`, `--nla-bg-subtle`, `--nla-text-subtle`, `--nla-border-strong`, `--nla-shadow-xl`, `--nla-radius-2xl`, `--nla-radius-pill`, `--nla-space-1..10`
- **JetBrains Mono** для чисел / кодов через `var(--nla-font-mono)` + alias `--bs-font-monospace` чтобы BS-утилита `.font-monospace` подхватилась без правок шаблонов

### Components
- **8 системных:** `Panel`, `KPI`, `Tag`, `Pill`, `TabBar`, `RatingBadge` (credit-rating string), `AiScore` (AI-балл 0-100), `InfoRow` (slot+prop под обратную совместимость с 24+ вызовами через `#value`)
- **Доменные:** `BondHero` (шапка карточки бумаги с 5 KPI), `BondInfoBasic` (заменил `BondBasicTab`), `IssuerCardGrid` (полный rewrite, CSS grid вместо BS row/col), `IssuerFilters` (новая модель: search/category/sector/rating/aiBucket/yield/coupon/duration/tradeable/hasRating/isFloat/hideMatured), `IssuerProfile`
- **Вынесены из inline-`defineComponent` в SFC:** `YieldBar.vue`, `RangeRow.vue`
- **Минимальная адаптация `BondAiTab`:** 5 карточек обёрнуты в `<Panel>`, средний балл через `<AiScore>`. Логика (polling, markdown-рендер, удаление) не тронута

### Routes / pages
- **Новый роут `/issuers/[id]`** — страница профиля эмитента. Карточки `IssuerCardGrid` теперь ссылаются туда

### Logic
- **`composables/useIssuerFilters.ts`** — единая логика фильтрации эмитентов и облигаций для `pages/bonds/by-issuer.vue` и `pages/bonds/monthly.vue`. Тип `IssuerFilterState`, функции `matchesBond`, `matchesIssuerRating`, `matchesIssuerAi`. Фильтрация по 22-уровневой шкале через `useRating.maxRatingOrd` + `ordMatchesBucket`

### Bug fixes during migration
- `IssuerCardGrid` runtime-баг: двойной `defineProps` (`const props = defineProps as any`) перетирал ссылку на пропсы — `props.ratings` возвращал функцию вместо объекта, бейджи рейтингов не рендерились
- `BondAiTab` использовал `<RatingBadge :rating="number">`, но новый `RatingBadge` принимает `string` (credit-rating). Старый компонент был перегружен — заменил на `<AiScore>` для AI-скоров

### Skipped (out of scope)
- Новые `BondCouponsTab` / `BondTradingTab` / `BondHistoryTab` из пакета — их prop-сигнатуры полностью отличаются от существующих (статусы купонов, отдельные `bids`/`asks`/`trades`, chart через `<slot name="chart">`). Существующие табы используют токены и автоматически перерендерились в новой палитре после Шага 1
- Очистка мёртвого CSS-оверрайдов в `main.css` — defensive-фолбэки оставлены, стоимость хранения нулевая

### Tests
- Без регрессий: 117 vitest + ~145 Go-тестов зелёные после каждого шага
- `npx nuxi typecheck` — 0 ошибок (было 37 до прошлой сессии)

---

## v0.11.1 (2026-05-06) — Frontend rating alignment + typecheck debt cleanup

### Bug fixes
- **`useFormat.ratingChipStyle()` coloured Moody's `Baa1` red** — string-prefix match hit `r.startsWith('b')` for `Baa1`/`Ba1`/`Ba2`/`Ba3` (lower investment grade and BB tier). Now delegates to `useRating.ratingTierStyle()` which normalises to a canonical tier first
- **Issuer rating filters in `monthly.vue` / `by-issuer.vue` collapsed BBB+BB+B+CCC into one bucket** — used legacy 1-10 score where BBB- and BB+ both map to 3. Filter now uses 22-level ordinal via `maxRatingOrd(rating.ratings)` + `ordMatchesBucket`. New buckets: AAA / AA / A / BBB / BB / B+ ниже / Без рейтинга
- **`InfoRow` silently dropped `#value` slot content** — template only rendered the `value` prop, so rows in `BondDetailsTab` that supplied a slot (badges, custom HTML) showed an empty value cell. Now renders the slot when present, falls back to the prop otherwise. `value` prop is now optional

### New
- **`composables/useRating.ts`** — TypeScript port of `internal/service/rating_score.go`. `normalizeRating`, `legacyScore10`, `ratingTierStyle`, `maxRatingOrd`, `ordMatchesBucket`. Pure functions, mirrors Go behaviour 1:1 — both test suites now share the same agency-format table
- **`IssuerRating.score_ord` field** added to frontend type (`composables/useApi.ts`) so the new filters can read the canonical ordinal directly from the API
- **Vitest** added to `frontend/`. 117 cases in `composables/useRating.spec.ts` mirror `internal/service/rating_score_test.go`; run via `npm test` (or `npm run test:watch`)

### Typecheck
- **`npx nuxi typecheck` is clean** (was: 37 errors across 3 files). Fixes:
  - `InfoRow` slot/prop dual-mode (24 errors in `BondDetailsTab.vue`)
  - `YieldBar` / `RangeRow` inline components — `value: { type: Number as PropType<number | null>, default: null }` so chart inputs typed `number | null` are accepted (10 errors in `BondYieldsTab.vue` / `BondHistoryTab.vue`)
  - Chart.js tooltip: `weight: '600'` → `'bold'`, null-safe `ctx.parsed.y` (3 errors)

---

## v0.11.0 (2026-05-03) — Cross-agency rating normalisation

### Bug fixes
- **`ratingToScore` collapsed BBB- and BB+ into the same score (3)** — losing the most important boundary in credit risk (investment grade ↔ speculative). Same for A/A-, BBB+/BBB, BB/BB- pairs
- **НКР (`BBB.ru`), НРА (`BBB|ru|`), Moody's (`Baa1`)** were not parsed at all — silently scored 0 and lost during sorting
- **`dohodScore` overwrote per-agency scores** (`details.go:173-179`) — if АКРА=AAA + Эксперт=BB and dohod gave 5, both became 5

### New
- **`service/rating_score.go::NormalizeRating(text) → (ord, tier)`** — single normaliser for all 8 agencies (АКРА, Эксперт РА, НКР, НРА, S&P, Fitch, Moody's, ДОХОДЪ) onto a 22-level ordinal scale (AAA=22 ... D=1, 0=unrated). Handles outlook stripping (`Stable`/`Negative`/`развивающийся`/...) and parenthetical suffixes
- **`LegacyScore10(ord)`** — maps 22-level ordinal back to legacy 1-10 so existing API/frontend filters keep working unchanged
- **`IssuerRating.ScoreOrd int`** new bson field (alongside legacy `Score`) — write paths fill both via `service.fillScores`
- **`RatingService.RecomputeAllScores(ctx)`** runs at API startup to migrate previously stored records; one-shot 614 records updated on first run
- **`bond.go::GetBondsGroupedByIssuer`** sort now uses `ScoreOrd` — issuer cards on main page render in correct credit order across agencies

### Tests
- `service/rating_score_test.go` — 96 cases across all 8 agencies, plus targeted regression tests:
  - `TestNormalizeRating_InvestmentVsSpeculativeBoundary` (BBB- > BB+)
  - `TestNormalizeRating_CrossAgencyOrdering` (АКРА AAA > Эксперт A, etc.)
  - `TestDohodLegacyRoundTrip` (ДОХОДЪ n → ord → legacy = n for n in 1..10)

### Other
- `mongo/RatingRepo.Upsert/BulkUpsert` clear `_id` before `$set` — Mongo treats `_id` as immutable on update; previously broke `RecomputeAllScores`

---

## v0.10.1 (2026-05-03) — Cleanup & bond.go refactor

### Cleanup
- Untracked build artifacts (`api`, `coverage.out`) removed from repo, `.gitignore` fixed (`/api`, `coverage.out`, `*.test`)
- PostgreSQL schema вынесена из inline-строк в `internal/database/migrations/0001_init.sql`, грузится через `embed.FS` в лексикографическом порядке
- Hardcoded `OPENAI_PROXY` IP в `config.go` возвращён на `getEnv`
- `CLAUDE.md` — секция Working preferences: контекст хранить в репо-документах, не в session memory

### Refactor
- **`internal/service/bond.go`** распилен 1073 → 354 строк (поведение не менялось, 61 тест зелёный):
  - `bond_parse.go` — `parseBond`, `mergeYieldData`
  - `bond_calc.go` — `calculateFields`, `calcRiskCategory`, `sortBonds`, `bondScore`
  - `bond_sync.go` — `SyncMissingIssuers`, `SyncMissingRatingsFromMoex`
  - `bond_helpers.go` — `extractRows`, `get*`/`safe*` примитивы

### Tests
- `internal/database/postgres_test.go` — guard на embed.FS миграций (файлы видны и не пустые)

---

## v0.10.0 (2026-04-10) — MOEX CCI Ratings & Rating Display Redesign

### MOEX CCI Rating Integration (Backend)
- **`GetCCIRatings()`** — новый метод в `moex/client.go`, fetch `/iss/cci/rating/companies/ecbd_{EMITTER_ID}.json`
- **`CCIRating` struct** — AgencyName, RatingValue, RatingDate из extended JSON формата
- **`SyncMissingRatingsFromMoex()`** — находит эмитентов без рейтингов, подтягивает из MOEX CCI (200ms delay, 3 min timeout)
- Запускается на старте API после SyncMissingIssuers (последовательно в одной горутине)
- Добавлены агентства: НКР (11 рейтингов), НРА (8) — ранее только АКРА, Эксперт РА, ДОХОДЪ
- Итого: 474 эмитента с рейтингами, 587 записей (Эксперт РА: 219, АКРА: 209, ДОХОДЪ: 138, НКР: 11, НРА: 8)

### Rating Display Redesign (Frontend)
- **`ratingChipStyle()`** — вынесен в `useFormat.ts` (из IssuerCardGrid и [secid].vue)
- Цвета по тирам: AAA→зелёный, AA→синий, A→teal, BBB→жёлтый, BB→оранжевый, B/C/D→красный, Отозван→серый
- **IssuerCardGrid** — компактные бейджи только с рейтингом (агентство в tooltip), font 11px monospace
- **[secid].vue** — рейтинги в одну строку с ISIN/Код/Тип/Валюта (без отдельного блока)
- **IssuerFilters** — статическая легенда агентств (AAA(RU)/АКРА, ruA+/Эксперт РА, BB+.ru/НКР, BBB|ru|/НРА, AA/ДОХОДЪ, Baa1/Moody's)
- NULL-рейтинги отфильтрованы из отображения
- «—» бейдж для эмитентов без рейтингов
- «Рейтинг не присвоен» на детальной странице если нет рейтингов

### Cleanup
- Удалены неиспользуемые CSS классы: `.rating-grid`, `.rating-chip*` из main.css
- Удалён дублированный `ratingChipStyle()` из IssuerCardGrid.vue и [secid].vue

---

## v0.9.0 (2026-04-10) — Issuer Auto-Sync, Ratings & AI Improvements

### Issuer Auto-Sync
- **`SyncMissingIssuers()`** — при старте API фоновая горутина сверяет бонды MOEX с `bond_issuers`, для недостающих запрашивает `EMITTER_ID` через MOEX description API (200ms delay, 2 min timeout)
- **`IssuerRepo.Upsert()`** — создание/обновление записей `bond_issuers` по secid
- **`IssuerRepo.GetAllSecids()`** — быстрая загрузка всех существующих secid для сверки
- **`IssuerRepo.UpdateEmitterName()`** — обновление имени эмитента для всех бондов по `emitter_id`
- **Emitter name backfill** — `updateRatingsFromDohod()` теперь автоматически обновляет `emitter_name` в `bond_issuers` при получении имени из dohod.ru

### AI Analysis
- **Float rating** — `BondAnalysis.Rating` изменён с `*int` на `*float64` (поддержка `[RATING:77.5]`)
- **`parseFloatRating()`** — парсинг десятичных рейтингов, округление до 1 знака
- **Delete analysis** — `DELETE /api/v1/analyses/{id}` (handler + service + mongo repo)
- **Markdown tables** — `renderMarkdown()` в BondAiTab.vue теперь парсит `| col | col |` формат в `<table>` с Bootstrap-стилями и dark mode

### Queue Reliability
- **`ResetStaleJobs()`** — при старте worker сбрасывает "running" задачи старше 3 минут в "pending"

### Frontend
- **Unified rating colors** — `aiRatingStyle()`, `aiRatingStyleSoft()`, `issuerRatingBg()` в useFormat.ts
- **AI rating badge** на странице детали облигации (из `analysisStats`)
- **Markdown tables** в Chat с dark mode (`color: var(--nla-text)`)
- **Readability** — `--nla-text-muted` light: `#64748b`, dark: `#8b9bb5`; `--nla-bg` light: `#f8f9fb`
- **Rating colors** — score 8: `#17a2b8`, score 7: `#5bc0de`

### Credit Ratings
- Agency ДОХОДЪ добавлен (внутренний рейтинг dohod.ru credit_rating)
- **310 из 359** эмитентов без имён получили имена через backfill из `issuer_ratings`

### Тесты
- 61 тест (включая 21 subtest rating parser с decimal), все проходят
- Новые тесты не требуются — добавленный код (IssuerRepo, SyncMissingIssuers) зависит от MongoDB/MOEX и тестируется только интеграционно

---

## Оценка объёма работ

**Общий прогресс: ~50% от полной паритетности с ASH**

| Область | Готовность | Комментарий |
|---------|-----------|-------------|
| Бэкенд API | ~80% | Bonds, auth, analysis, queue, chat, issuer grouping, marketdata, favorites. Нет: details-данных (dohod.ru), оферт, invest_score |
| Список облигаций | ~75% | 10 колонок, сортировка, пагинация, ★ избранное. Нет: колонка ОЦЕНКА, badge BOARDNAME захардкожен TQCB, нет min-max range для флоатеров |
| Детальная страница | ~55% | Header ✅, 7 из 8 табов. Но: таб "Детали" отсутствует полностью, Основное упрощено (нет оферт/buyback/сектора), Доходности нет спредов/индексов, Купоны нет оферт в timeline |
| Группировка по эмитентам | ~85% | Серверная группировка по emitter_id. Визуал отличается от ASH (карточки вместо таблицы) |
| Месячные купоны | ~85% | Серверная группировка ?monthly=true. Аналогично |
| Чат | ~90% | Работает, 3 агента |
| Визуал (CSS) | ~70% | CSS Custom Properties (--nla-*) = зеркало ASH. Navbar, tabs, cards, stat-cards, tables — портированы. Но мелкие расхождения в деталях |
| Авторизация UI | ~90% | Логин/регистрация, хедер. В ASH этого нет — NLA опережает |
| Избранное | ~90% | PostgreSQL + API + фронтенд. В ASH нет — NLA опережает |
| Инструменты | 0% | Нет (в ASH есть /tools) |
| Homepage | 0% | Нет hero-секции (в ASH есть feature cards + quick links). NLA сразу показывает BondTable |
| Тесты | ~50% | 61 тестов (auth, middleware, handler, rating parser incl. float). Нет интеграционных тестов на bond/issuer service |

### Что есть в NLA, но нет в ASH
- Auth UI (login/register/logout)
- Favorites full stack (список, toggle, badge в навбаре)
- Credit rating badge в заголовке облигации
- Trading status badge в заголовке
- Расширенная сортировка (best, duration)
- Ценовой диапазон в истории (RangeRow)

### Что есть в ASH, но нет в NLA
1. **Вкладка "Детали"** — полностью отсутствует (dohod.ru, кредитные рейтинги по агентствам, финансы эмитента, расчётные показатели)
2. **INVEST_SCORE** — колонка в таблице списка
3. **Страница "Инструменты"** (/tools)
4. **Hero homepage** (feature cards, quick links)
5. **Доходности**: "К оферте", таблицы "Показатели доходности", "Спреды и индексы", временные метки
6. **Основное**: даты оферт/buyback/размещения, английское название, сектор, уровень листинга, объём эмиссии
7. **Купоны**: PUT/CALL оферты в timeline, таблица параметров купона
8. **Quick Actions** card внизу detail page
9. **Купонная колонка**: min-max range для флоатеров, badge "I" для индексируемых

---

## v0.8.0 — CSS-система и визуальная синхронизация с ASH

### CSS Custom Properties
- **Полная система CSS-переменных** (`--nla-*`): зеркалирует ASH `--ash-*` палитру
- Light + Dark тема через `:root` / `.dark` (те же hex-значения что в ASH)
- Переменные для: primary, bg, text, border, shadow, radius, navbar

### Компоненты (plain CSS вместо Tailwind @apply)
- **Navbar**: glassmorphism (`backdrop-filter: blur(12px)`), `.nla-navbar`, `.nla-brand`, `.nla-nav-link` — точные размеры ASH (0.9rem, padding 0.5rem 0.85rem, radius 6px)
- **Tabs**: pill-style `.bond-tabs` / `.bond-tab` / `.bond-tab--active` (radius 10px, active = primary bg + white text, transition 0.15s ease)
- **Tab icons**: `.bond-tab__icon` wrapper — SVG alignment fix (inline-flex)
- **Stat cards**: `.stat-card` с `border-left: 3px solid primary`, min-height 90px, цветовые модификаторы (success/danger/warning/info)
- **Data table**: `.data-table` со sticky header, hover primary-subtle, uppercase th
- **Buttons**: `.btn-primary` / `.btn-secondary` / `.btn-ghost` / `.btn-danger` через CSS vars
- **Cards**: `.card` / `.card-hover` через CSS vars (border, shadow, radius) 
- **Inputs**: `.input` через CSS vars, focus ring primary
- **Прочее**: badge, badge-sm, filter-label, panel-header, section-title, code-block

### Страницы исправлены
- **by-issuer.vue**: убрана фейковая пагинация "Показать ещё" (в ASH нет), рендерятся все эмитенты с JS-фильтрацией
- **monthly.vue**: аналогично убрана пагинация, рендерятся все
- **[secid].vue**: табы переписаны на `.bond-tab` классы
- **default.vue**: header переписан на `nla-navbar` / `nla-nav-link` классы, footer через CSS vars

---

## v0.7.0 (2026-04-09) — Избранное / Watchlist

### Бэкенд
- **Таблица `favorites`** в PostgreSQL: user_id + secid, UNIQUE constraint, CASCADE delete
- **FavoriteRepository**: Add, Remove, ListByUser, GetSecIDs, IsFavorite, CheckMultiple, Count
- **FavoriteHandler**: 5 эндпоинтов (list, toggle, check, add, delete)
- Миграция автоматическая через RunMigrations

### API
```
GET    /api/v1/favorites              — список избранного (JWT)
POST   /api/v1/favorites/toggle       — добавить/убрать (JWT)
GET    /api/v1/favorites/check?secids= — проверить несколько (JWT)
POST   /api/v1/favorites/{secid}      — добавить (JWT)
DELETE /api/v1/favorites/{secid}      — убрать (JWT)
```

### Фронтенд
- **useAuth composable**: JWT в localStorage, login/register/logout, authHeaders
- **useFavorites composable**: load, toggle, isFavorite, реактивный count
- **Страница /login**: вход + регистрация, валидация, редирект
- **Страница /favorites**: таблица избранных с кнопкой удаления
- **Header**: кнопка «Войти» / имя пользователя + выход, ★ со счётчиком
- **BondTable**: ★ toggle (видна только для авторизованных)
- **Bond detail [secid]**: ★ кнопка рядом с названием

---

## v0.6.0 (2025-07-25) — Все вкладки детальной страницы

### Бэкенд
- **13 новых полей marketdata** в Bond: Open, Low, High, WAPrice, NumTrades, ValToday, BidDepth, OfferDepth, NumBids, NumOffers, UpdateTime, TradeTime, SysTime, PrevPrice
- Новый helper `getInt64Ptr` в bond service для парсинга целочисленных полей MOEX
- Bond interface в useApi.ts расширен всеми новыми полями

### Купоны (BondCouponsTab)
- 4 stat-карточки: ставка %, сумма ₽, период (с названием), следующий купон (дата + дней)
- Таймлайн ключевых событий: следующий купон + погашение
- Прогноз по годам: суммарные карточки + таблица с progress-барами
- Табличка расписания купонов (была, сохранена)

### История (BondHistoryTab)
- Композитный Chart.js: линия цены + столбцы объёма (зелёные/красные) + пунктирная номинал на 100%
- 4 stat-карточки: цена закрытия, изменение за период, волатильность, средний объём
- Визуализация диапазона цен (мин/сред/текущая/макс/номинал)
- Баланс торговых дней (рост/падение/без изменений)
- 2 таблицы статистики: за период (7 строк) + изменение цены (7 строк)

### Торговля (BondTradingTab)
- Статус торгов по TRADINGSTATUS вместо эвристики vol_today
- Ценовые уровни дня: Open/Low/High/Last/Bid/Ask с визуальными барами
- Таблица торговых данных (7 строк): Open, Low, High, Prev close, Day change, WAP, WAP change
- Статистика объёмов (8 строк): сделки, объём шт/₽, средний размер, глубина, bid/ask ratio
- Секция таймстемпов: время обновления, торгов, системное

### Доходности (BondYieldsTab)
- 5 видов доходности с барами: YTM, эффективная, купонная, текущая, по WAP
- **Декомпозиция YTM**: купонный доход + ценовой доход/убыток с progress-барами
- Цветовая шкала баров: indigo → emerald → amber → red (по уровню доходности)

### Прочее
- Доходность >999% отображается как ">999%" (было >9999%)
- 48 тестов проходят без изменений

### Всё ещё отсутствует vs ASH ⚠️
- Доходности: нет "К оферте", нет таблиц "Показатели доходности" и "Спреды и индексы", нет временных меток
- Основное: упрощённые даты (нет оферт, buyback, размещения), нет англ. названия, сектора, листинга, объёма эмиссии
- Купоны: нет PUT/CALL оферт в timeline, нет отдельной таблицы параметров купона

---

## v0.5.0 (2026-04-09) — Исправление группировки по эмитентам

### Сделано ✅
- **Миграция issuer данных**: 3175 записей `bond_issuers` экспортированы из ASH MongoDB → NLA MongoDB, индексы на `secid` (unique) и `emitter_id`
- **IssuerRepo** (`internal/mongo/issuer.go`): MongoDB-репозиторий с GetAll, GetBySecid, GetBySecids
- **API endpoint `/api/v1/bonds/grouped`**: серверная группировка по `emitter_id`, поддержка `?monthly=true`
- **Переписан `by-issuer.vue`**: использует серверный API вместо regex на secname, badge с ID эмитента, фильтр по рейтингу (AAA/AA/A/BBB/Без), фильтры по купону %, цене макс.
- **Переписан `monthly.vue`**: использует серверный API `?monthly=true`, аналогичные фильтры и badges
- **useApi.ts**: добавлены типы `IssuerGroup`, `IssuerGroupResponse` и метод `getBondsGrouped(monthly?)`
- **DI wiring**: IssuerRepo проброшен через BondService → handler → router
- **Тесты**: 48 тестов проходят без регрессий

### Статистика API
- `GET /api/v1/bonds/grouped` → 508 эмитентов, 3068 облигаций
- `GET /api/v1/bonds/grouped?monthly=true` → 325 эмитентов, 1078 облигаций

---

## v0.4.0 (2026-04-09) — Визуальная синхронизация

### Сделано ✅
- **Список облигаций**: 10 колонок как в ASH (НАЗВАНИЕ, ДОХ., ЦЕНА, ИЗМЕНЕНИЕ, ОБЪЕМ, НКД, КУПОН, ПОГАШЕНИЕ, СТАТУС, ДЕЙСТВИЯ)
- **Новые поля API**: `last_change`, `last_change_prcnt`, `trading_status` из MOEX marketdata
- **Заголовок**: "Облигации Московской биржи" + кнопки "По эмитентам" / "Месячные купоны"
- **Сортировка**: select в стиле ASH + "Найдено: N"
- **Колонка СТАТУС**: badge "Торги идут" / "Торги не ведутся" по TRADINGSTATUS из MOEX
- **Колонка ИЗМЕНЕНИЕ**: цветной % (зелёный/красный) из LASTCHANGEPRCNT
- **Колонка ДЕЙСТВИЯ**: кнопка ⓘ
- **Дедупликация**: удалены дубли облигаций (MOEX отдаёт одну бумагу на нескольких boardId)
- **Детальная страница**: header переделан — price card справа (цена%, ₽, изменение, доходность), badges ("T+: Облигации - безадрес.", "TQCB"), info row (ISIN, Код, Тип, Валюта)
- **Рейтинговые бейджи**: точные hex-цвета ASH (10→#198754, 9→#198754, ..., 0→#000000) в monthly.vue и by-issuer.vue
- **CSS**: primary indigo #6366f1, card radius 10px, button/input radius 6px, stat-card left border, tab active pill, data-table padding 13px/3

### Нерешённые проблемы (на тот момент) ⚠️
- ~~**Группировка по эмитентам**: ИСПРАВЛЕНО в v0.5.0~~
- ~~**Месячные купоны**: ИСПРАВЛЕНО в v0.5.0~~
- ~~**Вкладка Торговля**: нет таблицы торговых данных — ИСПРАВЛЕНО в v0.6.0~~
- ~~**Вкладка История**: нет volume bars — ИСПРАВЛЕНО в v0.6.0~~
- ~~**Вкладка Купоны**: нет stat-cards, timeline — ИСПРАВЛЕНО в v0.6.0~~
- ~~**Вкладка Доходности**: нет декомпозиции — ИСПРАВЛЕНО в v0.6.0~~
- **Badge типа бумаги в таблице**: всегда "Фикс с известным купоном" — не всегда верно
- **> 9999% доходность**: некоторые бумаги показывают ">999%" — проблема качества данных MOEX
- **Badge BOARDNAME**: захардкожен "TQCB", должен приходить из API
- **Header info row**: нет рег. номера (в ASH есть)

### Нужно сделать ❌

#### P0 — Критично
1. ~~Миграция issuer данных из ASH~~ ✅
2. ~~API endpoint для issuer grouping~~ ✅
3. ~~Переписать by-issuer.vue~~ ✅
4. ~~Переписать monthly.vue~~ ✅
5. **Вкладка "Детали"** — полностью отсутствует (dohod.ru данные, кредитные рейтинги по агентствам, финансы эмитента, расчётные показатели)

#### P1 — Важно
6. **Доходности**: добавить "К оферте", таблицы показателей доходности, спреды и индексы (Z-spread, G-spread, BEI и т.д.)
7. **Основное**: добавить даты оферт/buyback/размещения, английское название, сектор, уровень листинга, объём эмиссии
8. **Купоны**: добавить оферты в timeline, таблицу параметров купона
9. **Колонка ОЦЕНКА** в таблице (средний AI score — INVEST_SCORE)
10. **Страница "Инструменты"** (/tools)
11. **Hero homepage** (feature cards — в ASH есть)
12. **Quick Actions** card внизу detail page
13. **API endpoint для issuer sync**: worker для обновления данных

#### P2 — Мелкие правки
14. Badge BOARDNAME из API вместо хардкода
15. Рег. номер в info row заголовка
16. Купонная колонка: min-max range для флоатеров
17. Текст "Показаны записи X-Y из Z" в пагинации

#### P3 — Тесты
18. Тесты на bond service (parseBond, calculateFields, sortBonds, dedup)
19. Тесты на новые поля (trading_status, last_change)
20. Тесты на issuer grouping API

---

## v0.3.0 (ранее) — Базовый функционал

### Сделано
- Go API + JWT auth + PostgreSQL users
- MOEX ISS API client (5 endpoints)
- Bond list with sort/pagination
- Bond detail (7 tabs: basic, trading, history, yields, coupons, ai, external)
- AI analysis (OpenAI + queue worker + poll)
- Credit ratings (MongoDB + admin API)
- Chat (3 agents, MongoDB sessions)
- Monthly coupons page
- By-issuer page (с кривой группировкой)
- Nuxt 3 frontend + Tailwind + dark mode
- Docker compose (6 services)
- 48 tests passing
