<template>
  <div class="card overflow-hidden mb-6">
    <!-- Row 1: search + selects -->
    <div class="p-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3">
        <div class="lg:col-span-2">
          <div class="relative">
            <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none" style="color: var(--nla-text-muted)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path stroke-linecap="round" d="m21 21-4.35-4.35"/></svg>
            <input :value="filters.search" @input="emit('update', { ...filters, search: ($event.target as HTMLInputElement).value })" type="text" placeholder="Поиск по названию..." class="input pl-10" />
          </div>
        </div>
        <div>
          <select :value="filters.couponType" @change="emit('update', { ...filters, couponType: ($event.target as HTMLSelectElement).value })" class="input">
            <option value="">Тип купона</option>
            <option value="fixed">Фиксированный</option>
            <option value="float">Плавающий</option>
            <option value="indexed">Индексируемый</option>
          </select>
        </div>
        <div>
          <select :value="filters.rating" @change="emit('update', { ...filters, rating: ($event.target as HTMLSelectElement).value })" class="input">
            <option value="">Рейтинг</option>
            <option value="aaa">AAA (9-10)</option>
            <option value="aa">AA (7-8)</option>
            <option value="a">A (5-6)</option>
            <option value="bbb">BBB и ниже</option>
            <option value="none">Без рейтинга</option>
          </select>
        </div>
        <div v-if="showPeriod">
          <select :value="filters.period" @change="emit('update', { ...filters, period: ($event.target as HTMLSelectElement).value })" class="input">
            <option value="">Период купона</option>
            <option value="monthly">Ежемесячный</option>
            <option value="quarterly">Ежеквартальный</option>
            <option value="semiannual">Полугодовой</option>
            <option value="annual">Годовой</option>
          </select>
        </div>
        <div v-if="showCategory">
          <select :value="filters.category" @change="emit('update', { ...filters, category: ($event.target as HTMLSelectElement).value })" class="input">
            <option value="">Категория</option>
            <option value="Корпоративная">Корпоративные</option>
            <option value="ОФЗ">ОФЗ</option>
            <option value="Субфедеральная">Субфедеральные</option>
            <option value="Муниципальная">Муниципальные</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Row 2: range filters -->
    <div class="px-4 pb-4" style="background: var(--nla-bg); border-top: 1px solid var(--nla-border)">
      <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-5 gap-3 pt-4">
        <div>
          <label class="filter-label">Доходность, %</label>
          <div class="flex items-center gap-1.5">
            <input :value="filters.yieldMin" @input="emit('update', { ...filters, yieldMin: parseNum($event) })" type="number" step="0.5" placeholder="от" class="input text-center" />
            <span style="color: var(--nla-text-muted)" class="text-xs">—</span>
            <input :value="filters.yieldMax" @input="emit('update', { ...filters, yieldMax: parseNum($event) })" type="number" step="0.5" placeholder="до" class="input text-center" />
          </div>
        </div>
        <div>
          <label class="filter-label">Купон, % год.</label>
          <div class="flex items-center gap-1.5">
            <input :value="filters.couponMin" @input="emit('update', { ...filters, couponMin: parseNum($event) })" type="number" step="0.5" placeholder="от" class="input text-center" />
            <span style="color: var(--nla-text-muted)" class="text-xs">—</span>
            <input :value="filters.couponMax" @input="emit('update', { ...filters, couponMax: parseNum($event) })" type="number" step="0.5" placeholder="до" class="input text-center" />
          </div>
        </div>
        <div>
          <label class="filter-label">Погашение, дней</label>
          <div class="flex items-center gap-1.5">
            <input :value="filters.maturityMin" @input="emit('update', { ...filters, maturityMin: parseNum($event) })" type="number" placeholder="от" class="input text-center" />
            <span style="color: var(--nla-text-muted)" class="text-xs">—</span>
            <input :value="filters.maturityMax" @input="emit('update', { ...filters, maturityMax: parseNum($event) })" type="number" placeholder="до" class="input text-center" />
          </div>
        </div>
        <div>
          <label class="filter-label">Цена макс., %</label>
          <input :value="filters.priceMax" @input="emit('update', { ...filters, priceMax: parseNum($event) })" type="number" placeholder="напр. 100" class="input" />
        </div>
        <div class="flex items-end">
          <button @click="emit('reset')" class="btn-secondary w-full text-sm">✕ Сбросить</button>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="flex items-center justify-center gap-12 px-4 py-5" style="border-top: 1px solid var(--nla-border)">
      <div class="text-center">
        <div class="text-3xl font-bold font-mono" style="color: var(--nla-text)">{{ stats.issuers }}</div>
        <div class="text-xs mt-1 uppercase tracking-wider" style="color: var(--nla-text-muted)">эмитентов</div>
      </div>
      <div class="text-center">
        <div class="text-3xl font-bold font-mono" style="color: var(--nla-text)">{{ stats.bonds }}</div>
        <div class="text-xs mt-1 uppercase tracking-wider" style="color: var(--nla-text-muted)">облигаций</div>
      </div>
      <div class="text-center">
        <div class="text-3xl font-bold font-mono" style="color: var(--nla-text)">{{ stats.total }}</div>
        <div class="text-xs mt-1 uppercase tracking-wider" style="color: var(--nla-text-muted)">показано</div>
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
