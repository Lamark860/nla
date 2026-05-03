# CHANGELOG — NLA (ASH → NLA Migration)

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
