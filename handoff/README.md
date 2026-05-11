# NLA · Visual Redesign Handoff

Дизайн-пакет для переноса в `frontend/`.

Коротко: **тёплый off-white в светлой теме, тёплый почти-чёрный в тёмной, единый violet-акцент `#5b3aa8` / `#b89cff`**. Контракт `--nla-*` сохранён — все 30+ компонентов, читающих эти токены, продолжат работать.

---

## С чего начать

1. **`screens/index.html`** — открой в браузере, посмотри пары light/dark всех 7 экранов
2. **`DESIGN.md`** — обзор системы (палитра, типографика, spacing, что не трогаем)
3. **`MIGRATION.md`** — пошаговый чеклист, 9 шагов в порядке безопасности
4. **`tokens.css`** — drop-in замена для блоков `:root` / `[data-theme="dark"]` в `main.css`
5. **`components/`** — Vue 3 SFC, готовые к копированию в `frontend/components/`

---

## Структура

```
handoff/
├── README.md                    ← этот файл
├── DESIGN.md                    ← обзор системы
├── MIGRATION.md                 ← пошаговый чеклист
├── tokens.css                   ← :root + [data-theme="dark"]
├── screens/                     ← PNG light/dark пары + index.html
│   ├── index.html
│   ├── by-issuer-{light,dark}.png
│   ├── bond-basic-{light,dark}.png
│   ├── bond-chart-{light,dark}.png
│   ├── bond-coupons-{light,dark}.png
│   ├── bond-trades-{light,dark}.png
│   ├── bond-issuer-{light,dark}.png
│   └── bond-ai-{light,dark}.png
└── components/
    ├── Panel.vue                ← системный контейнер (заменяет .card + .panel-header)
    ├── KPI.vue                  ← метрика для hero
    ├── InfoRow.vue              ← пара label-value (replace existing)
    ├── Tag.vue                  ← маленькая прямоугольная метка
    ├── Pill.vue                 ← пилл-статус с цветной точкой
    ├── TabBar.vue               ← сегментный переключатель табов
    ├── RatingBadge.vue          ← кредитный рейтинг (replace existing)
    ├── AiScore.vue              ← AI-балл 0..100
    ├── BondHero.vue             ← шапка карточки облигации с 5 KPI
    ├── BondInfoBasic.vue        ← таб «Параметры» (replaces BondBasicTab)
    ├── BondCouponsTab.vue       ← таб «Купоны» (новый)
    ├── BondTradingTab.vue       ← таб «Торги» — стакан + лента (новый)
    ├── BondHistoryTab.vue       ← таб «История» — график (новый)
    ├── IssuerCardGrid.vue       ← грид карточек эмитентов (replace existing)
    ├── IssuerFilters.vue        ← фильтр-панель главной (replace existing)
    └── IssuerProfile.vue        ← шапка профиля эмитента (новый)
```

---

## Принципы (1 минута чтения)

1. **Данные — главный визуал.** Цифры, рейтинги, графики держат ритм. Иллюстрации не несут смысла.
2. **Один акцент.** Violet. Зелёный/красный — только семантика (доходность, цена, статус).
3. **Контракт `--nla-*` неизменен.** Меняем значения, добавляем под тем же префиксом — не переименовываем.
4. **Темы равноправны.** AA-контраст в обеих, `--nla-text-muted` в dark поднят до `#919a91`.
5. **Числа — JetBrains Mono с tabular-nums.** Запятая как разделитель, пробел как тысячный (`13,42%`, `1 000 000 ₽`).

---

## Самые быстрые победы (1 PR ≈ 1 час)

Если у твоей команды нет времени на весь редизайн сразу, возьми только это и выкати как «PR #1»:

1. Скопируй содержимое `tokens.css` в `main.css` (заменив старые `:root` и `[data-theme="dark"]` блоки)
2. Подключи JetBrains Mono в `nuxt.config.ts`
3. Задеплой

После этого:
- акцент сменится на violet
- фоны станут тёплыми
- числа автоматически выровнены, потому что `font-monospace` теперь указывает на JetBrains Mono с feature-settings

Это 80% визуальной переработки. Дальше можно подключать компоненты по одному — `MIGRATION.md` написан так, что каждый шаг безопасен и независим.

---

## Что **не** в пакете

- Skeleton-состояния (если нужны — отдельный пакет)
- BondAiTab — большой, переписывать его без уточнения флоу AI смысла нет
- chat.vue, страницы аутентификации — вне скоупа

---

## Контакт по вопросам

Дизайн собран на основе текущего фронта (Nuxt 3 + Vue 3 + Bootstrap 5 + @nuxtjs/color-mode). Все компоненты совместимы с этим стеком, никаких новых рантайм-зависимостей.

Если что-то ломается при переносе — `MIGRATION.md` → раздел «Если что-то пошло не так».
