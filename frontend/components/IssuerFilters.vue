<template>
  <div class="card filter-panel mb-4">
    <div class="card-body">
      <!-- Row 1: search + selects -->
      <div class="row g-2 mb-2">
        <div class="col-12 col-md-4">
          <div class="input-group">
            <span class="input-group-text"><i class="bi bi-search"></i></span>
            <input :value="filters.search" @input="emit('update', { ...filters, search: ($event.target as HTMLInputElement).value })" type="text" class="form-control" placeholder="Поиск по названию..." />
          </div>
        </div>
        <div class="col-6 col-md-2">
          <select :value="filters.couponType" @change="emit('update', { ...filters, couponType: ($event.target as HTMLSelectElement).value })" class="form-select">
            <option value="">Тип купона</option>
            <option value="fixed">Фиксированный</option>
            <option value="float">Плавающий</option>
            <option value="indexed">Индексируемый</option>
          </select>
        </div>
        <div class="col-6 col-md-2">
          <select :value="filters.rating" @change="emit('update', { ...filters, rating: ($event.target as HTMLSelectElement).value })" class="form-select">
            <option value="">Рейтинг</option>
            <option value="aaa">AAA (9-10)</option>
            <option value="aa">AA (7-8)</option>
            <option value="a">A (5-6)</option>
            <option value="bbb">BBB и ниже</option>
            <option value="none">Без рейтинга</option>
          </select>
        </div>
        <div v-if="showPeriod" class="col-6 col-md-2">
          <select :value="filters.period" @change="emit('update', { ...filters, period: ($event.target as HTMLSelectElement).value })" class="form-select">
            <option value="">Период купона</option>
            <option value="monthly">Ежемесячный</option>
            <option value="quarterly">Ежеквартальный</option>
            <option value="semiannual">Полугодовой</option>
            <option value="annual">Годовой</option>
          </select>
        </div>
        <div v-if="showCategory" class="col-6 col-md-2">
          <select :value="filters.category" @change="emit('update', { ...filters, category: ($event.target as HTMLSelectElement).value })" class="form-select">
            <option value="">Категория</option>
            <option value="Корпоративная">Корпоративные</option>
            <option value="ОФЗ">ОФЗ</option>
            <option value="Субфедеральная">Субфедеральные</option>
            <option value="Муниципальная">Муниципальные</option>
          </select>
        </div>
      </div>

      <!-- Row 2: range filters -->
      <div class="row g-2 align-items-end">
        <div class="col-6 col-md">
          <label class="filter-label">Доходность, %</label>
          <div class="input-group input-group-sm">
            <input :value="filters.yieldMin" @input="emit('update', { ...filters, yieldMin: parseNum($event) })" type="number" step="0.5" class="form-control text-center" placeholder="от" />
            <span class="input-group-text">—</span>
            <input :value="filters.yieldMax" @input="emit('update', { ...filters, yieldMax: parseNum($event) })" type="number" step="0.5" class="form-control text-center" placeholder="до" />
          </div>
        </div>
        <div class="col-6 col-md">
          <label class="filter-label">Купон, % год.</label>
          <div class="input-group input-group-sm">
            <input :value="filters.couponMin" @input="emit('update', { ...filters, couponMin: parseNum($event) })" type="number" step="0.5" class="form-control text-center" placeholder="от" />
            <span class="input-group-text">—</span>
            <input :value="filters.couponMax" @input="emit('update', { ...filters, couponMax: parseNum($event) })" type="number" step="0.5" class="form-control text-center" placeholder="до" />
          </div>
        </div>
        <div class="col-6 col-md">
          <label class="filter-label">Погашение, дней</label>
          <div class="input-group input-group-sm">
            <input :value="filters.maturityMin" @input="emit('update', { ...filters, maturityMin: parseNum($event) })" type="number" class="form-control text-center" placeholder="от" />
            <span class="input-group-text">—</span>
            <input :value="filters.maturityMax" @input="emit('update', { ...filters, maturityMax: parseNum($event) })" type="number" class="form-control text-center" placeholder="до" />
          </div>
        </div>
        <div class="col-6 col-md">
          <label class="filter-label">Цена макс., %</label>
          <input :value="filters.priceMax" @input="emit('update', { ...filters, priceMax: parseNum($event) })" type="number" step="0.1" class="form-control form-control-sm" placeholder="напр. 100" />
        </div>
        <div class="col-auto">
          <button @click="emit('reset')" class="btn btn-outline-secondary btn-sm">✕ Сбросить</button>
        </div>
      </div>
    </div>

    <!-- Stats footer -->
    <div class="card-footer">
      <div class="d-flex align-items-center">
        <div class="flex-fill text-center">
          <div class="h5 mb-0 fw-bold" style="font-variant-numeric: tabular-nums">{{ stats.issuers }}</div>
          <small class="text-muted text-uppercase" style="letter-spacing: 0.05em; font-size: 0.7rem">эмитентов</small>
        </div>
        <div class="flex-fill text-center">
          <div class="h5 mb-0 fw-bold" style="font-variant-numeric: tabular-nums">{{ stats.bonds }}</div>
          <small class="text-muted text-uppercase" style="letter-spacing: 0.05em; font-size: 0.7rem">облигаций</small>
        </div>
        <div class="flex-fill text-center">
          <div class="h5 mb-0 fw-bold" style="font-variant-numeric: tabular-nums">{{ stats.total }}</div>
          <small class="text-muted text-uppercase" style="letter-spacing: 0.05em; font-size: 0.7rem">показано</small>
        </div>
        <div class="border-start py-1 ps-3" style="width: 38%; flex-shrink: 0; display: grid; grid-template-columns: 1fr 1fr; gap: 3px 14px; align-content: center">
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('AAA(RU)')" style="font-size: 0.7rem; padding: 2px 6px">AAA(RU)</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">АКРА</span></div>
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('ruA+')" style="font-size: 0.7rem; padding: 2px 6px">ruA+</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">Эксперт РА</span></div>
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('BB+.ru')" style="font-size: 0.7rem; padding: 2px 6px">BB+.ru</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">НКР</span></div>
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('BBB|ru|')" style="font-size: 0.7rem; padding: 2px 6px">BBB|ru|</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">НРА</span></div>
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('AA')" style="font-size: 0.7rem; padding: 2px 6px">AA</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">ДОХОДЪ</span></div>
          <div class="d-flex align-items-center gap-1"><span class="badge font-monospace text-nowrap" :style="ratingChipStyle('Baa1')" style="font-size: 0.7rem; padding: 2px 6px">Baa1</span><span class="text-muted text-nowrap" style="font-size: 0.7rem">Moody's</span></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface IssuerFilterValues {
  search: string
  couponType: string
  rating: string
  period: string
  category: string
  yieldMin: number | null
  yieldMax: number | null
  couponMin: number | null
  couponMax: number | null
  maturityMin: number | null
  maturityMax: number | null
  priceMax: number | null
}

const fmt = useFormat()
const { ratingChipStyle } = fmt

defineProps<{
  filters: IssuerFilterValues
  stats: { issuers: number; bonds: number; total: number }
  showPeriod?: boolean
  showCategory?: boolean
}>()

const emit = defineEmits<{
  update: [filters: IssuerFilterValues]
  reset: []
}>()

function parseNum(e: Event): number | null {
  const v = (e.target as HTMLInputElement).value
  return v === '' ? null : Number(v)
}
</script>
