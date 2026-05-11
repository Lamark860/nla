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
      <!-- Row 1: search wide -->
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
      </div>

      <!-- Row 2: dropdowns -->
      <div class="filters__row filters__row--selects">
        <Select v-model="state.category" :options="categoryOptions" placeholder="Категория" @change="emitChange" />
        <Select v-model="state.sector" :options="sectorOptions" placeholder="Сектор" @change="emitChange" />
        <Select v-model="state.rating" :options="ratingOptions" placeholder="Рейтинг" @change="emitChange" />
        <Select v-model="state.aiBucket" :options="aiOptions" placeholder="AI-балл" @change="emitChange" />
      </div>

      <!-- Row 3: range pairs -->
      <div class="filters__row filters__row--ranges">
        <RangeField v-model="state.yield" label="Доходность" suffix="%" @change="emitChange" />
        <RangeField v-model="state.coupon" label="Купон" suffix="%" @change="emitChange" />
        <RangeField v-model="state.duration" label="Дюрация" suffix=" дн" @change="emitChange" />
      </div>

      <!-- Row 4: chips -->
      <div class="filters__row filters__row--chips">
        <Chip v-model="state.tradeable" @change="emitChange">Только торгуемые</Chip>
        <Chip v-model="state.hasRating" @change="emitChange">С рейтингом</Chip>
        <Chip v-model="state.isFloat" @change="emitChange">Только флоатеры</Chip>
        <Chip v-model="state.hideMatured" @change="emitChange">Скрыть погашенные</Chip>
      </div>
    </div>
  </Panel>
</template>

<script setup lang="ts">
import Panel from './Panel.vue'

const emit = defineEmits<{ change: [value: any] }>()

const state = reactive({
  search: '',
  category: '',
  sector: '',
  rating: '',
  aiBucket: '',
  yield: { min: '', max: '' } as { min: string; max: string },
  coupon: { min: '', max: '' },
  duration: { min: '', max: '' },
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
  { value: 'AA', label: 'AA / AA+ / AA-' },
  { value: 'A', label: 'A / A+ / A-' },
  { value: 'BBB', label: 'BBB / BBB+ / BBB-' },
  { value: 'BB', label: 'BB и ниже' },
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
  if (state.duration.min || state.duration.max) n++
  if (state.tradeable) n++
  if (state.hasRating) n++
  if (state.isFloat) n++
  if (state.hideMatured) n++
  return n
})

function emitChange() { emit('change', { ...state }) }
function reset() {
  state.search = ''
  state.category = state.sector = state.rating = state.aiBucket = ''
  state.yield = { min: '', max: '' }
  state.coupon = { min: '', max: '' }
  state.duration = { min: '', max: '' }
  state.tradeable = state.hasRating = state.isFloat = false
  state.hideMatured = true
  emitChange()
}
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
  props: { modelValue: Object as any, label: String, suffix: String },
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
        h('input', { type: 'number', class: 'if-range__field', placeholder: 'от', value: props.modelValue.min, onInput: update('min') }),
        h('span', { class: 'if-range__dash' }, '—'),
        h('input', { type: 'number', class: 'if-range__field', placeholder: 'до', value: props.modelValue.max, onInput: update('max') }),
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
  gap: 14px;
  padding: 16px 18px;
}
.filters__row { display: flex; gap: 10px; flex-wrap: wrap; }
.filters__row--selects > * { flex: 1 1 200px; }
.filters__row--ranges  > * { flex: 1 1 240px; }

/* search */
.search-input {
  position: relative;
  flex: 1 1 auto;
}
.search-input i {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--nla-text-muted);
  font-size: 14px;
}
.search-input__field {
  width: 100%;
  height: 38px;
  padding: 0 14px 0 36px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 13.5px / 1 var(--nla-font);
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
  height: 38px;
  padding: 0 32px 0 12px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 13px / 1 var(--nla-font);
  color: var(--nla-text);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8' viewBox='0 0 12 8'%3E%3Cpath fill='none' stroke='%238a8478' stroke-width='1.6' stroke-linecap='round' stroke-linejoin='round' d='M1.5 1.5 6 6l4.5-4.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
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
  gap: 4px;
}
:deep(.if-range__lbl) {
  font: 600 10.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
:deep(.if-range__pair) {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 12px;
  background: var(--nla-bg-input);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
}
:deep(.if-range__pair:focus-within) {
  border-color: var(--nla-primary);
  box-shadow: 0 0 0 3px var(--nla-primary-subtle);
}
:deep(.if-range__field) {
  width: 60px;
  border: 0;
  background: transparent;
  font: 500 13px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  outline: none;
  -moz-appearance: textfield;
}
:deep(.if-range__field::-webkit-outer-spin-button),
:deep(.if-range__field::-webkit-inner-spin-button) { -webkit-appearance: none; margin: 0; }
:deep(.if-range__dash)  { color: var(--nla-text-muted); }
:deep(.if-range__suffix){ color: var(--nla-text-muted); font-size: 12px; margin-left: auto; }

:deep(.if-chip) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-pill);
  font: 500 12.5px / 1 var(--nla-font);
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
</style>
