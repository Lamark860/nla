<!--
  IssuerFilters.vue — фильтр-панель главной (по эмитенту).
  Заменяет existing IssuerFilters: чище visual hierarchy, fold/unfold по primary,
  все хардкорные стили вынесены в токены.

  Совместима с текущим контрактом: испускает 'change' с объектом фильтров,
  внутри использует useFormat для нормализации значений.
-->
<template>
  <Panel class="issuer-filters">
    <template #head>
      <div class="filters__head">
        <i class="bi bi-funnel" aria-hidden="true"></i>
        <span class="filters__head-title">Фильтры</span>
        <span v-if="activeCount > 0" class="filters__head-count">{{ activeCount }}</span>
        <button v-if="activeCount > 0" class="filters__reset" @click="reset">Сбросить</button>
      </div>
    </template>

    <div class="filters__body">
      <!-- Row 1: search + advanced toggle -->
      <div class="filters__row filters__row--search">
        <div class="search-input">
          <i class="bi bi-search" aria-hidden="true"></i>
          <input
            v-model="state.search"
            type="search"
            placeholder="Поиск по названию эмитента, ISIN, SECID…"
            class="search-input__field"
            @input="emitChange"
          />
        </div>
        <button
          class="filters__advanced-toggle"
          :class="{ 'is-open': advancedOpen }"
          :aria-expanded="advancedOpen"
          @click="advancedOpen = !advancedOpen"
        >
          <i class="bi bi-sliders" aria-hidden="true"></i>
          <span>Расширенные</span>
          <span v-if="advancedActiveCount > 0" class="filters__advanced-badge">{{ advancedActiveCount }}</span>
          <i :class="advancedOpen ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
        </button>
      </div>

      <!-- Chips — visible always -->
      <div class="filters__row filters__row--chips">
        <Chip v-model="state.tradeable" @change="emitChange">Только торгуемые</Chip>
        <Chip v-model="state.hasRating" @change="emitChange">С рейтингом</Chip>
        <Chip v-model="state.isFloat" @change="emitChange">Только флоатеры</Chip>
        <Chip v-model="state.hideMatured" @change="emitChange">Скрыть погашенные</Chip>
      </div>

      <!-- Advanced: dropdowns + ranges (collapsible) -->
      <div v-show="advancedOpen" class="filters__advanced">
        <div class="filters__row filters__row--selects">
          <Select v-model="state.category" :options="categoryOptions" placeholder="Категория" @change="emitChange" />
          <Select v-model="state.sector" :options="sectorOptions" placeholder="Сектор" @change="emitChange" />
          <Select v-model="state.rating" :options="ratingOptions" placeholder="Рейтинг" @change="emitChange" />
          <Select v-model="state.aiBucket" :options="aiOptions" placeholder="Индекс" @change="emitChange" />
        </div>

        <div class="filters__row filters__row--ranges">
          <RangeField v-model="state.yield" label="Доходность" suffix="%" step="0.1" @change="emitChange" />
          <RangeField v-model="state.coupon" label="Купон" suffix="%" step="0.1" @change="emitChange" />
          <RangeField v-model="state.price" label="Цена" suffix="%" step="0.1" @change="emitChange" />
          <RangeField v-model="state.maturity" label="До погашения" suffix=" дн" @change="emitChange" />
        </div>
      </div>

      <!-- Stats + legend bar -->
      <div v-if="stats" class="filters__stats">
        <div class="filters__stat">
          <div class="filters__stat-num">{{ formatNum(stats.issuers) }}</div>
          <div class="filters__stat-lbl">эмитентов</div>
        </div>
        <div class="filters__stat">
          <div class="filters__stat-num">{{ formatNum(stats.bonds) }}</div>
          <div class="filters__stat-lbl">облигаций</div>
        </div>
        <div class="filters__stat">
          <div class="filters__stat-num">{{ formatNum(stats.shown) }}</div>
          <div class="filters__stat-lbl">показано</div>
        </div>
        <div class="filters__legend">
          <span v-for="ex in legendExamples" :key="ex.agency" class="filters__legend-item">
            <span class="filters__legend-rt" :class="`filters__legend-rt--${ex.tier}`">{{ ex.rating }}</span>
            <span class="filters__legend-agency">{{ ex.agency }}</span>
          </span>
        </div>
      </div>
    </div>
  </Panel>
</template>

<script setup lang="ts">
import Panel from './Panel.vue'

interface FilterStats {
  issuers: number
  bonds: number
  shown: number
}

const props = defineProps<{ stats?: FilterStats | null }>()
const emit = defineEmits<{ change: [value: any] }>()

const advancedOpen = ref(false)

const state = reactive({
  search: '',
  category: '',
  sector: '',
  rating: '',
  aiBucket: '',
  yield: { min: '', max: '' } as { min: string; max: string },
  coupon: { min: '', max: '' },
  price: { min: '', max: '' },
  maturity: { min: '', max: '' },
  tradeable: false,
  hasRating: false,
  isFloat: false,
  hideMatured: true,
})

const categoryOptions = [
  { value: 'corporate', label: 'Корпоративные' },
  { value: 'ofz', label: 'ОФЗ' },
  { value: 'sub', label: 'Субфедеральные' },
  { value: 'mun', label: 'Муниципальные' },
]
const sectorOptions = [
  { value: 'banks', label: 'Банки' },
  { value: 'oilgas', label: 'Нефть и газ' },
  { value: 'metals', label: 'Металлы' },
  { value: 'retail', label: 'Ретейл' },
  { value: 'telecom', label: 'Телеком' },
  { value: 'other', label: 'Прочее' },
]
const ratingOptions = [
  { value: 'AAA', label: 'AAA' },
  { value: 'AA', label: 'AA+ … AA-' },
  { value: 'A', label: 'A+ … A-' },
  { value: 'BBB', label: 'BBB+ … BBB-' },
  { value: 'BB', label: 'BB+ … BB-' },
  { value: 'B_BELOW', label: 'B+ и ниже' },
  { value: 'NONE', label: 'Без рейтинга' },
]
const aiOptions = [
  { value: '80+', label: '80+' },
  { value: '60-80', label: '60–80' },
  { value: '40-60', label: '40–60' },
  { value: '<40', label: 'до 40' },
]

const activeCount = computed(() => {
  let n = 0
  if (state.search) n++
  if (state.category) n++
  if (state.sector) n++
  if (state.rating) n++
  if (state.aiBucket) n++
  if (state.yield.min || state.yield.max) n++
  if (state.coupon.min || state.coupon.max) n++
  if (state.price.min || state.price.max) n++
  if (state.maturity.min || state.maturity.max) n++
  if (state.tradeable) n++
  if (state.hasRating) n++
  if (state.isFloat) n++
  if (state.hideMatured) n++
  return n
})

// Только из advanced-секции — для бейджа на кнопке
const advancedActiveCount = computed(() => {
  let n = 0
  if (state.category) n++
  if (state.sector) n++
  if (state.rating) n++
  if (state.aiBucket) n++
  if (state.yield.min || state.yield.max) n++
  if (state.coupon.min || state.coupon.max) n++
  if (state.price.min || state.price.max) n++
  if (state.maturity.min || state.maturity.max) n++
  return n
})

// Если в advanced что-то активно — раскрываем по умолчанию (на повторных загрузках страницы),
// чтобы пользователь не терял контекст
watch(advancedActiveCount, (n) => {
  if (n > 0 && !advancedOpen.value) advancedOpen.value = true
}, { immediate: true })

function emitChange() { emit('change', { ...state }) }
function reset() {
  state.search = ''
  state.category = state.sector = state.rating = state.aiBucket = ''
  state.yield = { min: '', max: '' }
  state.coupon = { min: '', max: '' }
  state.price = { min: '', max: '' }
  state.maturity = { min: '', max: '' }
  state.tradeable = state.hasRating = state.isFloat = false
  state.hideMatured = true
  emitChange()
}

function formatNum(n: number): string {
  return n.toLocaleString('ru-RU')
}

// Справочные примеры форматов рейтингов от разных агентств — для пользователя
// чтобы понять как читать рейтинги в карточках (handoff index.html stats-bar legend).
const legendExamples = [
  { rating: 'AAA(RU)', agency: 'АКРА',       tier: 'aaa' },
  { rating: 'ruAAA',   agency: 'Эксперт РА', tier: 'aaa' },
  { rating: 'AAA.ru',  agency: 'НКР',        tier: 'aaa' },
  { rating: 'AA',      agency: 'ДОХОДЪ',     tier: 'aa' },
  { rating: 'Baa1',    agency: "Moody's",    tier: 'bbb' },
  { rating: 'BB|ru|',  agency: 'НРА',        tier: 'bb' },
] as const
</script>

<!-- inline helpers, локальные на компонент -->
<script lang="ts">
import { defineComponent, h, computed } from 'vue'

export const Select = defineComponent({
  name: 'IF_Select',
  props: { modelValue: String, placeholder: String, options: Array as any },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    return () => h('select', {
      class: ['if-select', { 'is-active': !!props.modelValue }],
      value: props.modelValue,
      onChange: (e: Event) => {
        const v = (e.target as HTMLSelectElement).value
        emit('update:modelValue', v)
        emit('change')
      }
    }, [
      h('option', { value: '' }, props.placeholder),
      ...(props.options as any[]).map(o => h('option', { value: o.value }, o.label))
    ])
  }
})

export const RangeField = defineComponent({
  name: 'IF_RangeField',
  props: {
    modelValue: Object as any,
    label: String,
    suffix: String,
    // Native <input type=number> step; controls the up/down arrow increment.
    // Use "0.1" for price/yield-style filters where 1-unit jumps are too coarse.
    step: { type: String, default: '1' },
  },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit }) {
    const update = (k: 'min' | 'max') => (e: Event) => {
      const v = (e.target as HTMLInputElement).value
      emit('update:modelValue', { ...props.modelValue, [k]: v })
      emit('change')
    }
    return () => h('div', { class: 'if-range' }, [
      h('span', { class: 'if-range__lbl' }, props.label),
      h('div', { class: 'if-range__pair' }, [
        h('input', { type: 'number', step: props.step, class: 'if-range__field', placeholder: 'от', value: props.modelValue.min, onInput: update('min') }),
        h('span', { class: 'if-range__dash' }, '—'),
        h('input', { type: 'number', step: props.step, class: 'if-range__field', placeholder: 'до', value: props.modelValue.max, onInput: update('max') }),
        props.suffix ? h('span', { class: 'if-range__suffix' }, props.suffix) : null
      ])
    ])
  }
})

export const Chip = defineComponent({
  name: 'IF_Chip',
  props: { modelValue: Boolean },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, slots }) {
    return () => h('button', {
      type: 'button',
      class: ['if-chip', { 'is-on': props.modelValue }],
      onClick: () => { emit('update:modelValue', !props.modelValue); emit('change') }
    }, [
      h('i', { class: ['bi', props.modelValue ? 'bi-check-circle-fill' : 'bi-circle'] }),
      h('span', null, slots.default?.())
    ])
  }
})
</script>

<style scoped>
.filters__head {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}
.filters__head i { color: var(--nla-text-muted); font-size: 14px; }
.filters__head-title { font: 600 13px / 1.4 var(--nla-font); color: var(--nla-text); }
.filters__head-count {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  font: 700 11px / 1 var(--nla-font);
  padding: 3px 7px;
  border-radius: var(--nla-radius-pill);
}
.filters__reset {
  margin-left: auto;
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--nla-text-muted);
  font: 500 12px / 1 var(--nla-font);
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 3px;
}
.filters__reset:hover { color: var(--nla-primary); }

.filters__body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
}
.filters__row { display: flex; gap: 8px; flex-wrap: wrap; }
.filters__row--selects > * { flex: 1 1 180px; }
/* flex-basis ниже реального минимума гарантирует, что все 4 диапазона
   укладываются в одну строку: flex-grow растягивает их по ширине
   контейнера поровну. На ширине ~700px каждый получает ~165px, чего
   хватает на «от/—/до/%» вместе с нативной стрелкой. */
.filters__row--ranges  > * { flex: 1 1 130px; min-width: 0; }

/* search row layout */
.filters__row--search { display: flex; gap: 8px; align-items: stretch; }

/* advanced toggle button */
.filters__advanced-toggle {
  appearance: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 32px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 12.5px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 120ms ease, color 120ms ease, border-color 120ms ease;
}
.filters__advanced-toggle:hover {
  background: var(--nla-bg-subtle);
  color: var(--nla-text);
}
.filters__advanced-toggle.is-open {
  background: var(--nla-primary-light);
  border-color: color-mix(in oklab, var(--nla-primary) 30%, var(--nla-border));
  color: var(--nla-primary-ink);
}
[data-theme="dark"] .filters__advanced-toggle.is-open { color: var(--nla-primary); }
.filters__advanced-toggle .bi { font-size: 13px; }
.filters__advanced-badge {
  background: var(--nla-primary);
  color: #fff;
  font: 700 10px/1 var(--nla-font);
  padding: 2px 6px;
  border-radius: var(--nla-radius-pill);
  letter-spacing: 0;
}
[data-theme="dark"] .filters__advanced-badge { color: #0c0e0d; }

.filters__advanced {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--nla-border-light);
  margin-top: 4px;
}

/* search */
.search-input {
  position: relative;
  flex: 1 1 auto;
}
.search-input i {
  position: absolute;
  left: 11px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--nla-text-muted);
  font-size: 13px;
}
.search-input__field {
  width: 100%;
  height: 32px;
  padding: 0 12px 0 32px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 13px / 1 var(--nla-font);
  color: var(--nla-text);
  transition: border-color 120ms ease, box-shadow 120ms ease;
}
.search-input__field:focus {
  outline: none;
  border-color: var(--nla-primary);
  box-shadow: 0 0 0 3px var(--nla-primary-subtle);
}

/* injected helpers */
:deep(.if-select) {
  height: 32px;
  padding: 0 28px 0 11px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 12.5px / 1 var(--nla-font);
  color: var(--nla-text);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8' viewBox='0 0 12 8'%3E%3Cpath fill='none' stroke='%238a8478' stroke-width='1.6' stroke-linecap='round' stroke-linejoin='round' d='M1.5 1.5 6 6l4.5-4.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  cursor: pointer;
}
:deep(.if-select:focus) {
  outline: none;
  border-color: var(--nla-primary);
  box-shadow: 0 0 0 3px var(--nla-primary-subtle);
}
:deep(.if-select.is-active) {
  border-color: var(--nla-primary);
  background-color: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  font-weight: 600;
}

:deep(.if-range) {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
:deep(.if-range__lbl) {
  font: 600 9.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
:deep(.if-range__pair) {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 6px 0 8px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
}
:deep(.if-range__pair:focus-within) {
  border-color: var(--nla-primary);
  box-shadow: 0 0 0 3px var(--nla-primary-subtle);
}
/* Two fields share the container 50/50 via flex-basis 0 — no more visual
   asymmetry where "до" appeared wider than "от" because the suffix's
   margin-left:auto was eating leftover space on the right. */
:deep(.if-range__field) {
  flex: 1 1 0;
  min-width: 0;
  text-align: center;
  border: 0;
  background: transparent;
  font: 500 12.5px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  outline: none;
}
:deep(.if-range__field::placeholder) { font-size: 11px; }
/* Native number-input spinners stay visible so the user can nudge the
   value by `step` (0.1 for percent ranges) with the arrow keys or
   click-arrows. */
:deep(.if-range__dash)   { color: var(--nla-text-muted); font-size: 11px; flex-shrink: 0; }
:deep(.if-range__suffix) { color: var(--nla-text-muted); font-size: 11px; flex-shrink: 0; padding-left: 2px; }

:deep(.if-chip) {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 26px;
  padding: 0 10px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-pill);
  font: 500 11.5px / 1 var(--nla-font);
  color: var(--nla-text-secondary);
  cursor: pointer;
  transition: all 120ms ease;
}
:deep(.if-chip i) { font-size: 14px; color: var(--nla-text-muted); }
:deep(.if-chip:hover) { border-color: var(--nla-border-strong); color: var(--nla-text); }
:deep(.if-chip.is-on) {
  background: var(--nla-primary-light);
  border-color: color-mix(in oklab, var(--nla-primary) 30%, transparent);
  color: var(--nla-primary-ink);
  font-weight: 600;
}
:deep(.if-chip.is-on i) { color: var(--nla-primary); }

/* Stats + legend bar */
.filters__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(0, 2.6fr);
  gap: 0;
  margin-top: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--nla-border-light);
}
.filters__stat {
  text-align: center;
  padding: 1px 8px;
  border-right: 1px solid var(--nla-border-light);
}
.filters__stat:nth-last-child(2) { border-right: 0; }
.filters__stat-num {
  font: 600 18px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  letter-spacing: -0.01em;
  color: var(--nla-text);
}
.filters__stat-lbl {
  font: 500 9.5px/1 var(--nla-font);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  margin-top: 3px;
}
.filters__legend {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px 14px;
  align-items: center;
  padding: 0 6px 0 16px;
  border-left: 1px solid var(--nla-border-light);
}
.filters__legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font: 500 11px/1 var(--nla-font);
  color: var(--nla-text-muted);
  white-space: nowrap;
}
.filters__legend-rt {
  font: 600 11px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  padding: 4px 7px;
  border-radius: 5px;
  letter-spacing: 0.02em;
  border: 1px solid transparent;
}
.filters__legend-rt--aaa { background: #dff5e8; color: #0d6e3a; border-color: #bfe9d0; }
.filters__legend-rt--aa  { background: #e3f2dc; color: #3d6e0d; border-color: #cce4bd; }
.filters__legend-rt--a   { background: #fff5d8; color: #7a5800; border-color: #f0e2a3; }
.filters__legend-rt--bbb { background: #ffe6d4; color: #8a3d00; border-color: #f5cba8; }
.filters__legend-rt--bb  { background: #ffd7d4; color: #8a1f1a; border-color: #f5b1aa; }
[data-theme="dark"] .filters__legend-rt--aaa { background: #143a23; color: #7ee0a3; border-color: #1f5232; }
[data-theme="dark"] .filters__legend-rt--aa  { background: #1e3812; color: #a8d97c; border-color: #2c4f1d; }
[data-theme="dark"] .filters__legend-rt--a   { background: #3a2f0e; color: #e8c468; border-color: #4f4115; }
[data-theme="dark"] .filters__legend-rt--bbb { background: #3a1f0d; color: #e89b5e; border-color: #4f2c14; }
[data-theme="dark"] .filters__legend-rt--bb  { background: #3a1414; color: #e87070; border-color: #4f1f1f; }

@media (max-width: 992px) {
  .filters__stats { grid-template-columns: repeat(3, 1fr); }
  .filters__legend { grid-column: 1 / -1; grid-template-columns: repeat(2, 1fr); border-left: 0; padding: 12px 0 0; border-top: 1px solid var(--nla-border-light); margin-top: 8px; }
}
@media (max-width: 480px) {
  .filters__legend { grid-template-columns: 1fr; }
}
</style>
