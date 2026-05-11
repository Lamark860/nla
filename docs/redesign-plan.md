# Redesign migration plan

Перенос пакета `handoff/` (violet redesign) в `frontend/`. Источник истины — `handoff/MIGRATION.md`. Этот документ — журнал прогресса.

## Правила

- После каждого шага запускать `cd frontend && npm test` + `npx nuxi typecheck` (оба должны быть зелёные) и `cd .. && go test ./internal/... -count=1`
- Каждый шаг — отдельный коммит, чтобы можно было откатить точечно
- Открытые вопросы складываем в `docs/redesign-questions.md`, не блокируем процесс — отвечаем когда удобно

## Статус

| # | Шаг | Статус | Файлы | Заметки |
|---|---|---|---|---|
| 1 | Токены — `:root` + `[data-theme="dark"]` | **done** | `frontend/assets/css/main.css` | + alias `--bs-font-monospace: var(--nla-font-mono)` чтобы BS-утилита `.font-monospace` подхватила JetBrains Mono |
| 2 | JetBrains Mono в `nuxt.config.ts` | **done** | `frontend/nuxt.config.ts` | weights 400;500;600;700 |
| 3 | Системные компоненты | **done** | 8 SFC | `InfoRow` усилен под обратную совместимость с `#value`-named-slot |
| 3a | Семантический баг: `BondAiTab` использовал `<RatingBadge :rating="number">` для AI-скоров | **done** | `BondAiTab.vue` | Заменил на `<AiScore>` |
| 4 | `IssuerCardGrid.vue` | **done** | + фикс runtime-бага с двойным `defineProps`, `DateLine.days` → `PropType<number \| null>` |
| 5 | `BondBasicTab.vue` → `BondInfoBasic.vue` | **done** | + `pages/bonds/[secid].vue` |
| 6 | `BondHero.vue` встроен в `[secid].vue` | **done** | + `copyShareLink()`, маппинг ratings из `IssuerRatingResponse` |
| 7 | `IssuerFilters.vue` + адаптация страниц | **done** | новый `composables/useIssuerFilters.ts` с `matchesBond` / `matchesIssuerRating` / `matchesIssuerAi` — единый источник правды для обеих страниц. По вопросу 1 выбран вариант (в) — диапазоны + новые булевые фильтры. Расширил rating-options компонента: добавил `B_BELOW` и `NONE` |
| 8 | `IssuerProfile.vue` + роут `/issuers/[id]` | **done (частично)** | новый `pages/issuers/[id].vue`. **Пропущено:** новые `BondCouponsTab/BondTradingTab/BondHistoryTab` — их prop-сигнатуры полностью отличаются от существующих (`coupons: Coupon[]`-with-status, `bids/asks/trades` отдельно, chart через slot). Существующие табы используют новые токены — внешне после Шага 1 уже подтянулись |
| 9 | Вынос `YieldBar.vue` и `RangeRow.vue` | **done** | inline-`defineComponent` в `BondYieldsTab`/`BondHistoryTab` удалены |
| 10 | Очистка мёртвого CSS | **skipped** | По вопросу 5 (вариант б): defensive-оверрайды оставлены, ~50 строк, нулевая стоимость хранения |
| 11 | Минимальная адаптация `BondAiTab` | **done** | По вопросу 3 (вариант а): 5 карточек обёрнуты в `<Panel>`, средний AI-балл через `<AiScore>` вместо самописного бейджа. Вся логика (polling, markdown-рендер, удаление анализов) не изменена |
| 12 | `BondHistoryTab` редизайн (гибрид) | **done** | `BondHistoryTab.vue`, `RangeRow.vue`, `assets/css/main.css` | KPI-row, Panel-обёртки, period-toggle с фронт-фильтром, fallback на «Всё». `RangeRow` получил `tone`-проп; добавлен `yield-bar__fill--muted`. |

## Шаг 12 — `BondHistoryTab` редизайн

**Решение:** гибрид — каркас из `handoff/components/BondHistoryTab.vue` + сохраняем рич-инфу из текущей версии в той же визуальной подаче.

**Развилки:**
- Гибрид (выбран) vs 1-в-1 с handoff: handoff теряет ~50% инфы (stat-cards, баланс дней, 2 таблицы статистики). Гибрид сохраняет.
- Mode-toggle Цена/Доходность: **скрыт** — yield-history нет в API, отдельный backend-тикет.
- Periods (1Д/1Н/1М/3М/6М/1Г/Всё): **фронт-фильтр** по дате из полной истории, без backend-параметра.

### Структура (4 блока сверху вниз)

#### Block 1 — Stat-cards row (4 KPI)

```vue
<div class="hist-kpis">
  <KPI label="Цена закрытия" :value="fmt.percent(lastClose)">
    <template #sub v-if="lastCloseRub">{{ fmt.priceRub(lastCloseRub) }}</template>
  </KPI>
  <KPI label="Изменение за период" :value="priceChange" :tone="changeTone">
    <template #sub>{{ priceChangePct }}</template>
  </KPI>
  <KPI label="Волатильность" :value="volatility">
    <template #sub>σ дневных изменений</template>
  </KPI>
  <KPI label="Средний объём" :value="fmt.num(avgVolume)">
    <template #sub>шт./день</template>
  </KPI>
</div>
```
- Source: текущие `stat-card` в `BondHistoryTab.vue` строки 9-37 → удалить, заменить
- Grid: `repeat(4, 1fr)` desktop, `repeat(2, 1fr)` ≤768px
- Tone для «Изменение»: `success` если ≥0, `danger` если <0, `muted` если null

#### Block 2 — Главная chart-панель

```vue
<Panel flush>
  <template #head>
    <div class="hist-head">
      <div class="hist-title">
        <i class="bi bi-graph-up"/>
        <span>История цены</span>
      </div>
      <div class="hist-periods" role="tablist">
        <button v-for="p in periods" :key="p.value"
                class="hist-period" :class="{ active: p.value === activePeriod }"
                @click="activePeriod = p.value">{{ p.label }}</button>
      </div>
    </div>
  </template>
  <ClientOnly>
    <div class="hist-chart"><canvas ref="chartCanvas"/></div>
  </ClientOnly>
  <div class="hist-ranges">
    <RangeRow label="Текущая" :value="lastClose" :min="stats.low" :max="stats.high" tone="primary"/>
    <RangeRow label="Минимум" :value="stats.low" :min="stats.low" :max="stats.high"/>
    <RangeRow label="Максимум" :value="stats.high" :min="stats.low" :max="stats.high"/>
    <RangeRow label="Среднее" :value="stats.avg" :min="stats.low" :max="stats.high" tone="muted"/>
  </div>
</Panel>
```

**Periods state:**
```ts
const periods = [
  { value: '7',   label: '1Н' },
  { value: '30',  label: '1М' },
  { value: '90',  label: '3М' },
  { value: '180', label: '6М' },
  { value: '365', label: '1Г' },
  { value: 'all', label: 'Всё' },
]
const activePeriod = ref('30')

const filteredHistory = computed(() => {
  if (activePeriod.value === 'all') return props.history
  const days = Number(activePeriod.value)
  const cutoff = Date.now() - days * 86400_000
  return props.history.filter(h => new Date(h.date).getTime() >= cutoff)
})
```
- 1Д убрал — данные дневные, отдельная точка не имеет смысла. Минимум 1Н.
- Все computed (`stats`, `lastClose`, `priceChange`, `volatility`, `dayChanges`, `totalVolume`, `avgVolume`) → переписать на `filteredHistory.value` вместо `props.history`
- Edge case: если `filteredHistory.length === 0` (период длиннее данных) → показывать всю историю (`return props.history`) и подсветить активный период как «Всё»? Или просто пусто? **Решение:** если пусто, fallback на всю историю, без подсветки fallback'а

#### Block 3 — Баланс торговых дней

```vue
<Panel title="Баланс торговых дней" icon="bar-chart-steps">
  <div class="hist-balance">
    <div class="bal-row">
      <span class="bal-label">Рост</span>
      <span class="bal-val text-success font-monospace">{{ upDays }} дн. ({{ upPct }}%)</span>
    </div>
    <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--success" :style="{ width: upPct + '%' }"/></div>

    <!-- Падение, Без изменений — аналогично -->
  </div>
</Panel>
```
- Source: текущие строки 64-93 → завернуть в `<Panel>`, убрать `card p-4` обёртку и `section-title`
- Маркап утилит-классов BS оставить (они на токенах через `text-success`/`text-danger`)

#### Block 4 — Две таблицы статистики

```vue
<div class="hist-tables">
  <Panel title="Статистика за период" icon="calculator">
    <InfoRow label="Общий объём, шт." :value="fmt.num(totalVolume)" mono/>
    <!-- ... остальные 6 строк как сейчас -->
  </Panel>
  <Panel title="Изменение цены" icon="graph-up-arrow">
    <InfoRow label="Начальная цена" :value="fmt.percent(firstClose)" mono/>
    <!-- ... остальные 6 строк как сейчас -->
  </Panel>
</div>
```
- Source: текущие строки 97-131 → заменить `<div class="card overflow-hidden">` + `<div class="panel-header">` на `<Panel title icon>`. InfoRow остаются.
- Добавить `mono` где значения числовые (deviation от номинала, разница и т.д.)

### CSS-сетка

```css
.hist-kpis {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 18px;
}
.hist-head { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; width: 100%; }
.hist-title { display: flex; align-items: center; gap: 8px; font: 700 13px/1.4 var(--nla-font); color: var(--nla-text); }
.hist-title i { color: var(--nla-primary); }
.hist-periods { margin-left: auto; display: flex; gap: 2px; padding: 2px;
  background: var(--nla-bg-subtle); border: 1px solid var(--nla-border); border-radius: var(--nla-radius-sm); }
.hist-period { appearance: none; border: 0; background: transparent; padding: 5px 10px;
  font: 500 11.5px/1 var(--nla-font); color: var(--nla-text-secondary); border-radius: 4px; cursor: pointer; }
.hist-period.active { background: var(--nla-bg-card); color: var(--nla-text); font-weight: 600; box-shadow: var(--nla-shadow-sm); }
.hist-chart { height: 360px; padding: 16px; }
.hist-ranges { display: grid; grid-template-columns: repeat(4, 1fr); border-top: 1px solid var(--nla-border); }
/* RangeRow внутри hist-ranges — border-left, padding 12 16 (см. handoff/components/BondHistoryTab.vue:155-162) */

.hist-tables { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-top: 16px; }

@media (max-width: 992px) {
  .hist-kpis { grid-template-columns: repeat(2, 1fr); }
  .hist-tables { grid-template-columns: 1fr; }
  .hist-ranges { grid-template-columns: repeat(2, 1fr); }
}
```

### Что удалить из текущего файла

- `<div class="row g-3 mb-4">` ... `</div>` (stat-cards) — заменить на `<div class="hist-kpis">`
- `<div class="card p-4 mb-4">` (chart wrapper) и `<h3 class="section-title">` — заменить на `<Panel flush>`
- Вся секция «Ценовой диапазон» (`<div class="col-lg-6">` со старыми RangeRow) — **удалить**, RangeRow переехал в `hist-ranges` под чартом
- `<div class="card p-4">` обёртка вокруг балансе дней → `<Panel>`
- `<div class="card overflow-hidden">` + `<div class="panel-header">` × 2 → `<Panel>` × 2

### Чарт — мелкие правки

- `style="background: var(--nla-bg)"` на canvas-обёртке убрать (Panel сам даёт фон)
- `gridColor` / `textColor` оставить как есть — они уже dark/light-aware
- Цвет цены `#3b82f6` → не трогаем, или подменить на `var(--nla-primary)`? **Решение:** оставить пока, в отдельном тикете может стать violet
- Tooltip body weight `'bold'` уже исправлен в v0.11.1, не трогать

### Тесты / контроль

- `npx nuxi typecheck` (должен остаться 0 ошибок)
- `npm test` (vitest, не должен сломаться — компонент не покрыт тестами)
- Visual check через chrome-devtools MCP:
  - Live: `http://localhost:3000/bonds/SU26244RMFS2` (любой ликвидный, history всегда есть)
  - Reference: `file:///Users/maximlomaev/dockers/preprod/src/nla/handoff/preview/bond-detail.html` + переключение на таб «График»

### Файлы которые меняются

- `frontend/components/BondHistoryTab.vue` — full rewrite шаблона + расширение `<script>` на period-state и `filteredHistory`
- `docs/redesign-plan.md` — пометить Шаг 12 как **done**
- `CHANGELOG.md` — отдельная запись в `## Unreleased` или `v0.12.1`

### Workflow после рестарта Claude Code

1. Verify chrome-devtools MCP инструменты доступны
2. Baseline screenshots: live `BondHistoryTab` (Цена, основной таб) + handoff/preview раздел `chart-toolbar/chart-stage/chart-stats`
3. Implement Block 1 → screenshot → diff
4. Implement Block 2 (главная chart-панель + период-фильтр) → screenshot → diff
5. Implement Block 3+4 → screenshot → diff
6. Mobile (≤768px) — отдельный скрин
7. Dark theme — отдельный скрин с `data-theme="dark"`
8. Mark Шаг 12 done, commit

## Контрольные проверки

```bash
# в frontend/
npx nuxi typecheck && npm test

# в корне репо
go test ./internal/... -count=1
```

После финальных шагов всё зелёное: typecheck (0 ошибок), 117 vitest-кейсов, ~145+ Go-тестов.

## Откат

Если что-то ломается — `git revert` коммита соответствующего шага. Все шаги изолированы.

---

## Журнал прогресса

**2026-05-06, сессия 1:** выполнены Шаги 1–5 (+3a, +4a временный фикс).

**2026-05-07, сессия 2:** Шаги 6, 7, 8 (частично — без новых табов), 9, 11. Шаг 10 пропущен по решению (вопрос 5 вариант б). Все тесты зелёные.

**2026-05-08, сессия 3:** Шаг 12 — гибридный `BondHistoryTab`. Старые stat-cards / 2× `card overflow-hidden` / отдельная панель «Ценовой диапазон» удалены, заменены на `<KPI>` × 4, `<Panel flush>` с period-toolbar и `<RangeRow>` × 4 в `hist-ranges`, два `<Panel>` со статистикой. Добавлены period-state (`1Н/1М/3М/6М/1Г/Всё`, дефолт `30`) + `filteredHistory` с fallback на полную историю, если в окне нет точек (подсветка автоматически прыгает на «Всё»). Все computed (stats, lastClose, volatility, dayChanges и т.п.) перевешены на `filteredHistory`, чарт перерисовывается на смену периода. `RangeRow` расширен `tone`-пропом, добавлен `.yield-bar__fill--muted` в `main.css`. Скрины: `handoff/screens-step12/`. Typecheck + 117 vitest зелёные.

**Итог migrationа:**
- 16 новых компонентов + 2 вынесенных SFC
- Новый composable `useIssuerFilters.ts` (единая логика фильтрации)
- Новый роут `/issuers/[id]` с `IssuerProfile`
- Все правки в working tree, не закоммичено — пользователь сам разобьёт на коммиты

**Что осталось от пакета `handoff/` за рамками этой миграции:**
- `BondCouponsTab/BondTradingTab/BondHistoryTab` из handoff — не drop-in. Их использование = переписать chart-логику и переразложить данные. Отдельный тикет, если будем делать
- Skeleton-состояния — не в пакете
- Полный редизайн `BondAiTab` — отдельный тикет когда уточнится AI-флоу

---

## Phase 2 — догнать handoff: Шаги 13-19

После Шага 12 пользователь обратил внимание, что handoff содержит rich-контент на табах, которые в live ещё в старой вёрстке (Купоны, Стакан, AI), и data-слой в карточках/профиле эмитента местами беднее handoff. Этот раздел — план второй волны миграции.

### Inventory map (handoff ↔ live)

| handoff bond-detail таб | live таб | Статус | Шаг |
|---|---|---|---|
| Параметры (`BasicTab`) | Основное (`BondInfoBasic`) | Done в Шаге 5, минор-расхождения | (не отдельным шагом) |
| График (`ChartTab`) | История (`BondHistoryTab`) | Done в Шаге 12 | — |
| Купоны (`CouponsTab`) | Купоны (`BondCouponsTab`) | старая вёрстка | **14** |
| Стакан (`TradesTab`) | Торговля (`BondTradingTab`) | старая вёрстка, multi-level OB+trades нет в API | **17** |
| Эмитент (`IssuerTab`) | стр. `/issuers/[id]` (`IssuerProfile`) | догнать | **15** |
| AI-анализ (`AiTab`) | AI Анализ (`BondAiTab`) | Шаг 11 (минимальная адаптация) | **16** |
| — | Доходности (`BondYieldsTab`) | в handoff аналога нет | **не трогаем** |
| — | Детали (`BondDetailsTab`) | в handoff аналога нет; решено: оставить как есть | **не трогаем** |
| — | Внешние (`BondExternalTab`) | iframes, в handoff нет | **не трогаем** |
| handoff index.html (По эмитентам) | `/bonds/by-issuer` (`IssuerCardGrid`) | Done в Шаге 4, минор-расхождения | **18** |

### API-контракты (что есть, что блокирует)

Подтверждено через `curl http://localhost:8090/api/v1/...` на ликвидной бумаге `RU000A106JC8`:

| Endpoint | Что отдаёт | Что нужно для handoff | Блокер? |
|---|---|---|---|
| `GET /bonds/{secid}` | full Bond (`bid`, `offer`, `numbids`, `numoffers`, `biddeptht`, `offerdeptht`, `bid_offer_ratio`, `total_depth`, `mid_price_pct`) | best bid/offer + агрегаты ✓ | **multi-level orderbook (5+5 уровней) — НЕТ** |
| `GET /bonds/{secid}/coupons` | `[{coupon_date, record_date, start_date, value, value_percent, value_rub}]` | + status (paid/next/put/future/maturity) | вычисляется на фронте ✓ |
| `GET /bonds/{secid}/history` | OHLCV (price + volume) | + yield-history, + OFZ benchmark | **обе кривые отсутствуют** |
| `GET /bonds/{secid}/dohod` | `quality`, `stability`, `quality_balance`, `quality_earnings`, `quality_roe_score`, `liq_*`, `best_score`, `down_risk`, `description`, `credit_rating_text`, `simple_yield`, `current_yield`, `dohod_duration` | quality bars (4 штуки) + сводный балл | 2 из 4 баров есть напрямую (`quality`, `stability`), 2 маппятся из суб-показателей |
| **отсутствует** | — | trades feed (лента сделок) | **НЕТ endpoint, MOEX отдельный запрос** |

**Backend-блокеры — переедут в `docs/roadmap.md`:**
- `MOEX orderbook` endpoint (5+5 уровней цен) → нужен для handoff `TradesTab` (полноценный стакан)
- `MOEX trades` endpoint (лента сделок) → нужен для handoff `TradesTab`
- `yield_history` поле в `/bonds/{secid}/history` → нужен mode-toggle Цена/Доходность в `BondHistoryTab`
- `ofz_benchmark` (опц.) → пунктирная линия "OFZ benchmark" в графике

### Решения на развилках (зафиксировано до начала)

- **Эмитент**: догоняем `/issuers/[id]` под handoff `IssuerTab`. Отдельный таб на bond-detail НЕ добавляем (избежать дубликата контента).
- **Детали**: оставляем как есть. handoff аналога не имеет; полей много, перенос в `<Panel>` дал бы смешанную семантику со старыми. Гигиену сделаем отдельным тикетом если понадобится.
- **Стакан**: редизайн из доступных агрегатов (best bid/offer, depth, ratio, count). Multi-level OB и trades feed — backend-тикет.
- **Доходности**: вне scope handoff, не трогаем.
- **`pages/index.vue`** (флэт-таблица), `login`/`chat`/`tools`/`favorites`, skeleton-состояния — нет в handoff. Не трогаем.

---

### Шаг 13 — Подготовка (без кода)

- 13.1: Зафиксировать в этом документе план Шагов 14-19 (✓ — этот раздел).
- 13.2: Перенести список backend-блокеров (multi-level OB, trades feed, yield_history, OFZ benchmark) в `docs/roadmap.md` отдельной секцией "Phase D — данные для редизайна". Без обязательств по срокам.
- 13.3: Сделать baseline-скрины **всех** табов на ликвидной бумаге (RU000A106JC8): Основное / Торговля / Купоны / Доходности / Детали / AI / Внешние + страница `/issuers/{emitter_id}`. Сохранить в `handoff/screens-phase2/00-baseline-*.png`. Свет + тёмная тема. Это referent для "что было до".
- 13.4: Запустить полные регрессии **до** начала Phase 2 — `npx nuxi typecheck`, `npm test`, `go test ./internal/... -count=1`. Все три зелёные на старте — иначе фиксим перед началом.

**Контроль выхода из Шага 13:** план в этом документе, baseline-скрины на диске, зелёные регрессии. Только после этого начинаем Шаг 14.

### Шаг 14 — `BondCouponsTab` гибрид

**Цель:** перенести handoff `CouponsTab` (4-cell summary + table со status pills), сохранив наши rich-блоки (Параметры купона, Прогноз по годам).

#### Sub-steps

- **14.1** — добавить computed на фронте: для каждого купона из `/bonds/{secid}/coupons` определить `status: 'paid' | 'next' | 'put' | 'future' | 'maturity'`. Логика:
  - `coupon_date < today` → `paid`
  - `coupon_date === bond.next_coupon` → `next`
  - `coupon_date === bond.offerdate` → `put` (PUT-оферта совпадает с купоном)
  - `coupon_date === bond.matdate` → `maturity` (последний купон + номинал)
  - иначе → `future`
- **14.2** — заменить 4 stat-card вверху на `<KPI>` в `.coupons-kpis` grid: «Выплачено» (sum × paid), «Следующий купон» (date + days), «PUT-оферта» (date + days, скрыть если нет offerdate), «К погашению» (sum × ¬paid + facevalue).
- **14.3** — заменить «Ключевые события» (3 строки с цветными иконками) и «Параметры купона» (текущие 4 InfoRow) на единую `<Panel title="Параметры купона" icon="cash-stack">` с `<InfoRow mono>`. Иконку-блок «Ключевые события» удалить — даты уже в KPI.
- **14.4** — заменить таблицу «Прогноз по годам» — обернуть в `<Panel>`, переоформить progress-cell (см. как сейчас в `IssuerProfile.vue` или `RangeRow`).
- **14.5** — главная таблица купонов: переоформить со status pills (`<Pill tone="success">Выплачен</Pill>` / `<Pill tone="primary">Ближайший</Pill>` / `<Pill tone="warning">Оферта</Pill>` / `<Pill tone="default">Ожидается</Pill>` / `<Pill tone="danger">Погашение</Pill>`). Текущая таблица уже есть, нужны новые статус-колонки.

#### Проверки после 14.5
- `npx nuxi typecheck` 0 ошибок
- `npm test` 117/117
- chrome-devtools: live screenshot (десктоп light + dark, mobile 414×896)
- ручная проверка: на бумаге с PUT-офертой (любая корпоративная биржевая) видно `put` статус
- baseline diff vs `00-baseline-coupons.png` — соответствует handoff

### Шаг 15 — `IssuerProfile.vue` догнать handoff `IssuerTab`

**Цель:** карточка эмитента на `/issuers/[id]` показывает то же что handoff `IssuerTab` — header + 4-cell stats + bonds-table + ratings-cards + quality-bars из dohod.

#### Sub-steps

- **15.1** — issuer header: добавить «логотип» (буква в кружке, цвет = акцент) + теги (Системно-значимый / Гос. участие / Уровень 1 — пока только если в данных есть). Кнопка «Все эмитенты» → `/bonds/by-issuer`. Сейчас просто заголовок + meta.
- **15.2** — `<div class="issuer-stats">` 4-cell summary: Облигаций в обращении (с подразделом «N — флоатеры, M — фикс» — считаем на фронте по `is_float`), Общий объём (`sum(facevalue × issuesize_placed)`), Средняя YTM (взвешенная по объёму), AI-рейтинг эмитента (если есть `analysisStats.{secid}` → average).
- **15.3** — bonds-table: текущий `IssuerProfile` уже содержит список бумаг, но в виде blocks. Перевести в `<table class="issuer-bonds-tbl">` с колонками SECID / Название / Погашение / Купон (right) / Цена (right) / YTM (right) и подсветкой "current" если из URL передан `?current=secid`.
- **15.4** — ratings-cards: каждое `IssuerRating` → `<Panel>` сетка 4-кол: agency name, grade, outlook (с иконкой и цветом — Стабильный=neutral, Позитивный=success, Негативный=danger), updated date. Сейчас просто `<RatingBadge>`.
- **15.5** — quality-bars из dohod: 4 строки с `<RangeRow>`-подобным компонентом. Маппинг полей: «Кредитное качество»=`quality` (0–10), «Финансовая стабильность»=`stability`, «Прозрачность»=`quality_balance` (если null — скрываем), «Корп. управление»=`quality_earnings` (если null — скрываем). Бар цвета `--nla-success`. Внизу «Сводный балл» = `best_score` крупно справа.

#### Проверки после 15.5
- typecheck 0
- vitest 117/117
- ручной тест: открыть `/issuers/484` (Сбербанк, есть AAA от АКРА/ЭкспертРА/НКР/НРА), `/issuers/9697` (Роял Капитал, есть dohod-quality), `/issuers/1228` (без dohod) — все три без crash
- chrome-devtools скрин на каждом из трёх эмитентов

### Шаг 16 — `BondAiTab` full redesign

**Цель:** перенести handoff `AiTab` 2-кол layout: левая колонка `form panel + result panel`, правая `history panel + stats panel`. Логика polling/markdown/удаления — без изменений.

#### Sub-steps

- **16.1** — каркас 2-кол grid `.ai-grid` (1fr 380px), на ≤992px — stack. Шаблоны хранят текущую логику `pollAnalysis()`, `submitAnalysis()`, `deleteAnalysis()`.
- **16.2** — form panel: `<Panel title="Анализ облигации через ИИ" icon="stars">` с тремя блоками:
  - JSON preview (сгенерированный объект `bondData` от useApi) в `<pre class="code-block">`. Под ним badge «Dohod.ru данные включены» если `dohodDetails` загружен.
  - textarea «Дополнительный JSON · опционально» — пробросим в submit как extra.
  - submit-button с `<i class="bi-send-fill">` + label, полная ширина.
- **16.3** — result panel: появляется когда выбран анализ. Шапка — `<AiScore>` + время + clipboard/download иконки (clipboard = копирует markdown, download = .md файл). Body — рендеренный markdown через текущий `renderMarkdown()`.
- **16.4** — history panel (правая колонка): `<Panel title="История анализов" icon="clock-history" :meta="`${analyses.length}`">`. Список карточек, активная — `border-color: --nla-primary`. Внутри карточки: AiScore + время + preview (первые 100 символов markdown) + clipboard/delete actions.
- **16.5** — stats panel (правая колонка, под history): 2-cell grid — «Всего анализов» (число), «Средний рейтинг» (`<AiScore>` с avg). Источник `analysisStats[secid]`.

#### Проверки после 16.5
- typecheck 0
- vitest 117/117
- e2e ручной: запустить новый анализ → polling → результат отрисовался → удалить анализ → исчез
- скрины: пустое состояние (нет анализов), один анализ, несколько (активный + история), mobile

### Шаг 17 — `BondTradingTab` редизайн из агрегатов

**Цель:** перенести в Panel/KPI/InfoRow существующий контент, не претендуя на multi-level OB / trades feed (нет в API). Backend-тикет на эти данные — отдельной строкой в roadmap.

#### Sub-steps

- **17.1** — заменить 4 stat-card вверху на `<KPI>`: Статус (tone success/muted), Спрос (Bid + ₽ + глуб.), Предложение (Ask + ₽ + глуб.), Спред (п.п. + % от бида). Структура справа аналог Шага 12.
- **17.2** — «Цены торгового дня» (PriceBar × 6) и «Глубина стакана» (visual bidRatio bar) → 2 `<Panel>` бок-о-бок. Markup PriceBar не трогаем — компонент уже на токенах.
- **17.3** — «Торговые данные за день» / «Объёмы и статистика» → 2 `<Panel title icon>` с `<InfoRow mono>`. Удалить дубли — `numtrades`, `vol_today`, `valtoday` сейчас фигурируют в обоих, оставить только в одном (Объёмы).
- **17.4** — «Временные метки» → `<Panel title="Временные метки" icon="clock">` с тремя `<InfoRow mono>`.

#### Проверки после 17.4
- typecheck 0
- vitest 117/117
- ручной тест: открыть на торгующейся бумаге (статус T) и на нестандартной (статус N) — оба без crash, "Нет данных по стакану" корректно показывается
- скрины

### Шаг 18 — `IssuerCardGrid` точечный catch-up

**Цель:** мелкие расхождения с handoff IssuerCard.

#### Sub-steps

- **18.1** — preview-строка collapsed-карточки: добавить YTM первой бумаги (`{shortname} · {fmt.percent(yield)}`), текущее состояние без yield.
- **18.2** — summary: «N облигаций в обращении» вместо просто «N облигаций».
- **18.3** — AiPill в шапке (на одной строке с rating chips), сейчас `<AiScore>` ниже отдельным блоком.

#### Проверки после 18.3
- typecheck 0
- vitest 117/117
- скрины `/bonds/by-issuer` (с раскрытой карточкой и без)

### Шаг 19 — Финал

- **19.1** — Final regressions: typecheck, vitest, go test
- **19.2** — Обновить `CHANGELOG.md` запись `v0.13.0 — Phase 2 redesign catch-up` с перечислением Шагов 14-18
- **19.3** — Обновить `docs/redesign-plan.md` — пометить Шаги 13-19 как **done** в журнале

---

### Workflow между шагами

После каждого Шага (14, 15, 16, 17, 18) — **отдельный коммит**, чтобы можно было откатить точечно. Не пушим. Между шагами:

1. `npx nuxi typecheck` (зелёный)
2. `npm test` (117/117)
3. chrome-devtools live screenshot (light + dark + mobile)
4. визуальный diff с handoff/preview
5. commit (`feat(redesign): шаг N — XYZ`)

При обнаружении регрессии в одной из проверок — фиксим до коммита, не накапливаем долг.

### Что не входит в Phase 2 (явно out-of-scope)

- `pages/index.vue` (флэт-таблица) — нет дизайна в handoff
- `pages/login.vue`, `chat.vue`, `tools.vue`, `favorites.vue` — нет дизайна
- Skeleton-состояния — нет дизайна
- `BondDetailsTab`, `BondYieldsTab`, `BondExternalTab` — оставляем как есть по решению
- handoff Sidebar (240px nav слева) — отдельная история про лейаут, не таб
- Multi-level OB / trades feed / yield_history / OFZ benchmark — backend-тикеты, в roadmap

---

## Журнал прогресса (продолжение)

**2026-05-08, сессия 5 (Phase 3):** Шаги 20-24 — sidebar + унификация главных. Top-navbar заменён на sticky sidebar 240px (`layouts/default.vue` scoped, drawer на мобиле). `/bonds/by-issuer` и `/bonds/monthly` получили `<PageHead>` + `<ViewToggle>` (handoff page-head). `IssuerFilters` принимает `:stats` и рендерит footer с stat-cells + legend агентств (6 примеров). `/` → 302 redirect на `/bonds/by-issuer`, флэт-список переехал на `/bonds/flat`. typecheck/vitest/go test зелёные.

**2026-05-08, сессия 4 (Phase 2):** все Шаги 13-19 за один заход.
- Шаг 13 — план зафиксирован в этом документе, backend-блокеры в `docs/roadmap.md` Phase D, baseline-скрины 21 шт. в `handoff/screens-phase2/`, регрессии зелёные на старте
- Шаг 14 — `BondCouponsTab` гибрид: 4-cell summary, status pills (paid/next/put/maturity/future), Параметры в Panel/InfoRow. Проверено на двух бумагах (с PUT и без)
- Шаг 15 — `IssuerProfile.vue` догнал handoff `IssuerTab`: header с буквенным логотипом, 4-cell stats (взвеш. YTM, общий объём в трлн/млрд), bonds-table с подсветкой `?current=secid`, ratings cards (без outlook — нет в API), quality bars из dohod (4 строки + сводный балл + description). `BondHero` теперь линкует с `?current=secid`
- Шаг 16 — `BondAiTab` full redesign: 2-кол grid 1fr 380px, form panel + result panel с clipboard/download/close, history sidebar с active-state, stats 2-cell. Логика polling/markdown/delete не тронута
- Шаг 17 — `BondTradingTab` из агрегатов: 4 KPI top, 2 Panel (PriceBars + Глубина), 2 Panel InfoRow (Торговые данные / Объёмы), Временные метки. Multi-level OB + trades feed остаются backend-блокером
- Шаг 18 — `IssuerCardGrid` catch-up: yield первой бумаги в preview, «облигаций в обращении»
- Шаг 19 — финальные регрессии (typecheck 0, vitest 117/117, go test ok), CHANGELOG `v0.13.0`, журнал

Все правки в working tree, не закоммичено.

---

## Phase 3 — Sidebar layout + унификация главных страниц (Шаги 20-24)

**Контекст:** `/bonds/by-issuer` и `/bonds/monthly` — основные «лендинги» приложения, должны жить по handoff `index.html`. Предыдущая главная `/` (флэт-таблица) — рудимент. Решения зафиксированы в чате (вопросы 1-4 от 2026-05-08):

1. Sidebar 240px вместо top-navbar. Наполнение оставляем как сейчас (Облигации / Эмитенты / Купоны / Чат / Инструменты + Избранное + Войти + Theme toggle), не переносим из handoff (там «Каталог/Портфель/Инструменты» — для будущей расширенной навигации)
2. by-issuer + monthly — два роута, в page-head btn-group toggle. AI-страницы у нас отдельно нет
3. `/` → redirect на `/bonds/by-issuer`. Флэт-таблица переезжает на `/bonds/flat`, доступна из сайдбара как «Плоский список power»
4. Сводку «Эмитентов / Облигаций / Всего» переносим из страниц в footer фильтров + легенда рейтинг-агентств 6 примеров (как в handoff)

### Шаг 20 — `layouts/default.vue` → sidebar 240px

- Shell grid `240px 1fr`, sidebar sticky `top:0;height:100vh`, скроллится отдельно
- Сайдбар: brand (буква «N» в кружке + «NLA · bonds»), nav-items, Theme toggle внизу
- Mobile (≤768px): сайдбар скрывается, открывается через hamburger в верхней полоске
- Active state по route — текущая логика `isActive(path)` сохраняется
- Сохраняем `<main>`+`.container` для совместимости со страницами (контентная max-width внутри main)

### Шаг 21 — page-head

- В `pages/bonds/by-issuer.vue` и `pages/bonds/monthly.vue` сверху новый блок:
  - h1 «Облигации по эмитентам» / «Месячные купоны» + sub-line с количеством
  - btn-group [Эмитенты | Купоны] — переключение между роутами через NuxtLink
  - Export-кнопка (заглушка с `disabled`, или onClick → `console.log` пока)
- Старые `<h1>` и summary-line «Эмитентов / Облигаций / Всего» удаляются — переедут в Шаг 22

### Шаг 22 — `IssuerFilters` footer

- В `IssuerFilters.vue` под `.filters__row--chips` добавить `.filters__stats-bar`:
  - 3 stat-cells (эмитентов / облигаций / показано) — props/emits через `stats={ issuers, bonds, shown }`
  - 6 legend-items: AAA(RU)/АКРА, ruAAA/Эксперт РА, AA+.ru/НКР, AA/ДОХОДЪ, Baa1/Moody's, BB.ru/НРА — справочные примеры рейтинг-форматов
- Страницы передают подсчёты через `:stats` prop в IssuerFilters

### Шаг 23 — Redirect `/` → `/bonds/by-issuer` + `/bonds/flat`

- `pages/index.vue`: заменить контент на `definePageMeta({ middleware: () => navigateTo('/bonds/by-issuer') })` (или server-side redirect)
- Создать `pages/bonds/flat.vue` с текущим содержимым `pages/index.vue` (BondTable)
- Добавить пункт «Плоский список» в сайдбар (возможно в отдельную секцию «Инструменты»)

### Шаг 24 — финал

- typecheck, vitest, go test
- CHANGELOG `v0.13.1` (или `v0.14.0` если sidebar считаем major-bump)
- Journal entry

---

## Что ещё осталось вне scope (для будущих фаз)

- `pages/index.vue` — флэт-таблица облигаций, нет дизайна в handoff
- `pages/login.vue`, `pages/chat.vue`, `pages/tools.vue`, `pages/favorites.vue` — нет дизайна
- Skeleton states — нет в handoff
- `BondDetailsTab`, `BondYieldsTab`, `BondExternalTab` — оставлены по решению Phase 2 (handoff аналога нет, миграция в Panel/InfoRow дала бы смешанную семантику)
- handoff Sidebar 240px — отдельная история про общий лейаут приложения, не таб
- Multi-level OB / trades feed / yield_history / OFZ benchmark — backend-тикеты в `docs/roadmap.md` Phase D