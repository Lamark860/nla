# Migration · Чеклист переноса в существующий код

Цель — внедрить редизайн в `frontend/` без переписывания приложения. Все шаги изолированные, можно мерджить по одному.

---

## Порядок (рекомендуемый)

1. **Шаг 1** · Заменить токены в `main.css` → ничего не сломается, но визуал поменяется сразу
2. **Шаг 2** · Подключить шрифт JetBrains Mono в `nuxt.config.ts`
3. **Шаг 3** · Скопировать системные компоненты (`Panel`, `KPI`, `InfoRow`, `Tag`, `Pill`, `TabBar`, `RatingBadge`, `AiScore`)
4. **Шаг 4** · Заменить `IssuerCardGrid.vue` → главная сразу преобразится
5. **Шаг 5** · Заменить `BondBasicTab.vue` → `BondInfoBasic.vue`
6. **Шаг 6** · Внедрить `BondHero.vue` в `pages/bonds/[secid].vue` (вместо текущей шапки)
7. **Шаг 7** · `IssuerFilters.vue` → новый
8. **Шаг 8** · Профиль эмитента, остальные табы
9. **Шаг 9** · Удалить мёртвый CSS

Каждый шаг безопасно делать отдельным PR.

---

## Шаг 1 · Токены

**Файл:** `frontend/assets/css/main.css`

**Действие:** Заменить блоки `:root { … }` (строки **6–63**) и `[data-theme="dark"] { … }` (строки **66–107**) на содержимое `tokens.css`.

```diff
- :root {
-     --nla-font: 'Inter', -apple-system, ...;
-     --nla-primary: #6366f1;
-     ...
- }
- [data-theme="dark"] {
-     --nla-primary: #818cf8;
-     ...
- }
+ /* skopirovat' soderzhimoe handoff/tokens.css */
```

**Что проверить после:**
- Главная и страница бумаги визуально стали тёплее, акцент сменился с indigo на violet
- Тёмная тема: фон стал почти-чёрным тёплым, акцент — светлый violet
- Чарты: цвета 1..6 поменялись, но семантика (1=primary, 2=success) сохранилась
- **Контраст текста**: `--nla-text-muted` в dark теперь `#919a91` — это намеренное усиление, AA на новом фоне

**Что не делать:** не переименовывать токены, не удалять «ненужные», не выкидывать `[data-theme]`-блок. Колор-моде Nuxt'а зависит от этого атрибута.

---

## Шаг 2 · Шрифт JetBrains Mono

**Файл:** `frontend/nuxt.config.ts`

В `app.head.link[]` уже есть Inter. Добавить:

```diff
  link: [
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
    { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap' },
+   { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap' },
  ]
```

В `tokens.css` уже зашит `--nla-font-mono: 'JetBrains Mono', ...`. Старый класс `.font-monospace` от Bootstrap теперь рендерится этим шрифтом — числа автоматически станут tabular.

---

## Шаг 3 · Системные компоненты

Скопировать **в `frontend/components/`**:

| Откуда (handoff/components/) | Куда |
|---|---|
| `Panel.vue` | `frontend/components/Panel.vue` |
| `KPI.vue` | `frontend/components/KPI.vue` |
| `InfoRow.vue` | `frontend/components/InfoRow.vue` *(перезаписать существующий)* |
| `Tag.vue` | `frontend/components/Tag.vue` |
| `Pill.vue` | `frontend/components/Pill.vue` |
| `TabBar.vue` | `frontend/components/TabBar.vue` |
| `RatingBadge.vue` | `frontend/components/RatingBadge.vue` *(перезаписать)* |
| `AiScore.vue` | `frontend/components/AiScore.vue` *(новый)* |

> Nuxt 3 автоимпортирует компоненты из `components/` — никаких `import` не нужно. Если у тебя `components.dirs` со специфичной структурой — проверь.

**Совместимость:**
- `RatingBadge.vue` сохраняет старый prop `rating: string`, добавляет опциональный `agency: string`. Существующие вызовы продолжат работать.
- `InfoRow.vue` сохраняет `label, value`, добавляет `mono`, `tone`. Старые вызовы — без изменений.

---

## Шаг 4 · Грид эмитентов

**Файл:** `frontend/components/IssuerCardGrid.vue` → перезаписать содержимым `handoff/components/IssuerCardGrid.vue`.

**Контракт пропсов не меняется:**
```ts
defineProps<{
  issuers: IssuerGroup[]
  ratings: Record<string, IssuerRatingResponse>
  aiStats?: Record<string, AnalysisStats>
}>()
```

**Что изменилось внутри:**
- `.row.g-3 + .col-md-6.col-lg-4` → `.issuer-grid` (CSS grid с `auto-fill, minmax(360px, 1fr)`)
- Bootstrap badge'ы → `RatingBadge`, `AiScore`, `Tag`
- `.issuer-bond-metrics`, `.metric-label`, `.metric-value` → inline mini-component `Metric` внутри файла (`:deep(.metric)` стилизует)
- `.issuer-bond-dates` → inline `DateLine`

**Что в `main.css` можно удалить позже** (Шаг 9):
- `.issuer-bond-metrics`, `.metric-label`, `.metric-value` (если других вызовов нет — `grep` подтвердит)
- `.issuer-bond-dates`, `.date-label`

---

## Шаг 5 · Таб «Параметры»

**Файл:** `frontend/components/BondBasicTab.vue` → переименовать в `BondInfoBasic.vue`, скопировать содержимое.

**Где вызывается:**
```bash
grep -r "BondBasicTab" frontend/
```
Скорее всего в `frontend/pages/bonds/[secid].vue` — заменить на `<BondInfoBasic :bond="bond" />`.

**Изменения:**
- 4 панели → `<Panel>` вместо `.card + .panel-header`
- Все `<InfoRow>` получили `:mono` где значение число/дата
- Прогресс-бар жизни облигации стал ярче, добавлен маркер
- Иконка для «Финансовые параметры» сменена с `bi-currency-dollar` на `bi-cash-coin` (рублёвая семантика)

---

## Шаг 6 · BondHero

**Файл:** `frontend/pages/bonds/[secid].vue`

Найти блок «шапка бумаги» (то, что сейчас собрано из `<h1>`, `<RatingBadge>`-цепочки, бейджей, кнопок «избранное»). Заменить на:

```vue
<BondHero
  :bond="bond"
  :ratings="issuerRatings"
  :ai-score="aiAvg"
  :issuer-name="bond.emitter_name"
  :is-favorite="favorites.isFavorite(bond.secid)"
  @toggle-favorite="favorites.toggle(bond.secid)"
  @share="copyShareLink"
  @analyze="$router.push({ hash: '#ai' })"
/>
```

`BondHero` сам собирает 5 KPI из полей `bond.*`. Если каких-то полей у тебя нет (например, `duration` или `value_today_rub`), KPI просто не покажет sub-строку.

**Старая шапка** удаляется целиком, включая её inline-стили.

---

## Шаг 7 · IssuerFilters

**Файл:** `frontend/components/IssuerFilters.vue` → перезаписать.

**Предупреждение:** новый компонент **не подключён** к API текущей страницы. Он испускает `change` с объектом `{ search, category, sector, rating, aiBucket, yield, coupon, duration, tradeable, hasRating, isFloat, hideMatured }`. Тебе нужно адаптировать обработчик в `pages/bonds/by-issuer.vue` под этот объект (или поменять имена ключей в фильтре под существующий API).

**Если страница уже умеет принимать что-то другое** — вторая стратегия: оставить старый IssuerFilters рабочим, новый положить рядом как `IssuerFiltersNew.vue` и переключить на флажок, пока не привяжется.

---

## Шаг 8 · Остальные

| Компонент | Куда |
|---|---|
| `BondCouponsTab.vue` | `frontend/components/BondCouponsTab.vue` (новый) |
| `BondTradingTab.vue` | `frontend/components/BondTradingTab.vue` (новый) |
| `BondHistoryTab.vue` | `frontend/components/BondHistoryTab.vue` (новый — Chart.js прокидывается через `<slot name="chart">`) |
| `IssuerProfile.vue` | `frontend/components/IssuerProfile.vue` (для `pages/issuers/[id].vue`) |

Эти файлы — **не drop-in replacement**, потому что у тебя сейчас табов нет в этом виде. Их использование — отдельный кусок работы по сборке нового layout страницы бумаги.

---

## Шаг 9 · Очистка `main.css`

После всех замен:

```bash
grep -r "panel-header" frontend/
grep -r "issuer-bond-metrics" frontend/
grep -r "metric-label" frontend/
```

Если упоминаний нет — удалять соответствующие блоки в `main.css`. Если есть — оставлять.

**Точно можно удалить** (после Шагов 4–6):
- блок переопределения `.bg-light`, `.bg-dark` (мёртвый — не используется в шаблонах)
- секцию `.text-success !important { color: ... }` и аналоги (новый дизайн использует `tone`-пропы, не Bootstrap-утилиты)

**Не трогать:**
- bootstrap-импорт целиком
- классы `.btn`, `.form-control`, `.modal`, `.dropdown-*`, `.nav-tabs` (используются BS5)
- `.data-table` — на этом классе таблицы во многих местах

---

## Если что-то пошло не так

| Симптом | Причина | Что делать |
|---|---|---|
| Тёмная тема стала «горячей» (фон не чёрный, а серый) | Не заменён `[data-theme="dark"]` блок целиком | Перепроверь, что заменил **оба** блока в `main.css` |
| Бейджи рейтингов стали бесцветные | Сломался `useFormat().ratingChipStyle` или его не подключили в новом RatingBadge | Открыть `RatingBadge.vue`, убедиться, что `useFormat()` импортируется и `ratingChipStyle` доступен |
| Контраст текста плохой в light | В `tokens.css` есть `--nla-text-subtle` — он намеренно низкоконтрастный, **только** для текста ≥18px. Не использовать его на body. | Проверить, не использован ли в шаблонах |
| Чарт.js перекрашен только на 50% | Конфиг чарта читает HEX напрямую, а не `var()` | Заменить хардкод-цвета в конфиге на `getComputedStyle(document.documentElement).getPropertyValue('--nla-chart-1')` |
| Кнопки/инпуты бутстрапа не подхватили акцент | Это нормально — bootstrap.scss имеет свой `$primary`. Если хочешь и его перебить — переопредели `$primary: #5b3aa8` в `assets/css/_bootstrap-overrides.scss` до импорта BS. | По желанию |

---

## Что **не** в этом пакете

Намеренно оставлено:
- `BondAiTab.vue` — большой компонент со своим состоянием. Можно оборачивать в `Panel`, использовать `Pill` для статусов. Менять структуру не вижу смысла, пока флоу AI-анализа не уточнён.
- `chat.vue`, `favorites.vue`, страницы auth — за пределами скоупа аудита.
- Skeleton-состояния — добавлять параллельно с заменой компонентов, я могу выкатить пакет `<SkeletonRow>` отдельно, если нужно.
