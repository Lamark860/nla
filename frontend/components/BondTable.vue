<template>
  <div class="card overflow-hidden">
    <!-- Header with sort controls -->
    <div class="px-5 py-4 border-b border-slate-200/80 dark:border-slate-700/40 flex items-center justify-between gap-4 flex-wrap">
      <div class="flex items-center gap-3">
        <select v-model="currentSort" class="input w-auto text-sm py-2 pl-3 pr-8" @change="$emit('sort', currentSort)">
          <option value="best">Лучшие</option>
          <option value="coupon_desc">Купон ↓</option>
          <option value="coupon_asc">Купон ↑</option>
          <option value="yield_desc">Доходность ↓</option>
          <option value="yield_asc">Доходность ↑</option>
          <option value="maturity_asc">Погашение ↑</option>
          <option value="maturity_desc">Погашение ↓</option>
          <option value="volume_desc">Объём ↓</option>
          <option value="duration_asc">Дюрация ↑</option>
          <option value="duration_desc">Дюрация ↓</option>
        </select>
        <span class="text-sm text-slate-400 dark:text-slate-500">
          Найдено: <span class="font-semibold text-slate-700 dark:text-slate-300">{{ meta.total }}</span>
        </span>
      </div>
    </div>

    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="data-table" style="min-width: 1100px">
        <thead>
          <tr>
            <th class="text-left" style="width: 200px">НАЗВАНИЕ</th>
            <th class="text-right whitespace-nowrap" style="width: 85px">ДОХ.</th>
            <th class="text-right whitespace-nowrap" style="width: 95px">ЦЕНА</th>
            <th class="text-right whitespace-nowrap hidden md:table-cell" style="width: 100px">ИЗМЕНЕНИЕ</th>
            <th class="text-right whitespace-nowrap hidden lg:table-cell" style="width: 110px">ОБЪЕМ</th>
            <th class="text-right whitespace-nowrap hidden lg:table-cell" style="width: 60px">НКД</th>
            <th class="text-right whitespace-nowrap hidden md:table-cell" style="width: 100px">КУПОН</th>
            <th class="text-right whitespace-nowrap hidden lg:table-cell" style="width: 120px">ПОГАШЕНИЕ</th>
            <th class="text-center whitespace-nowrap hidden xl:table-cell" style="width: 120px">СТАТУС</th>
            <th class="text-center whitespace-nowrap hidden xl:table-cell" style="width: 60px">ДЕЙСТВИЯ</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="bond in bonds"
            :key="bond.secid"
            class="cursor-pointer group"
            @click="$router.push(`/bonds/${bond.secid}`)"
          >
            <!-- НАЗВАНИЕ -->
            <td>
              <div class="flex flex-col gap-0.5">
                <div class="flex items-center gap-2">
                  <span :class="categoryBadge(bond.bond_category)" class="badge-sm whitespace-nowrap">
                    {{ categoryLabel(bond.bond_category, bond.is_float, bond.is_indexed) }}
                  </span>
                </div>
                <span class="font-medium text-slate-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">{{ bond.shortname }}</span>
                <span class="text-xs text-slate-400 dark:text-slate-500 font-mono">{{ bond.isin }}</span>
              </div>
            </td>
            <!-- ДОХ. -->
            <td class="text-right font-mono tabular-nums whitespace-nowrap">
              <span :class="yieldColor(bond.yield)" class="font-semibold">
                {{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}
              </span>
            </td>
            <!-- ЦЕНА -->
            <td class="text-right font-mono tabular-nums whitespace-nowrap">
              <div class="flex flex-col items-end">
                <span class="text-slate-900 dark:text-white">{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</span>
                <span v-if="bond.price_rub" class="text-xs text-emerald-600 dark:text-emerald-400">{{ fmt.priceRub(bond.price_rub) }}</span>
              </div>
            </td>
            <!-- ИЗМЕНЕНИЕ -->
            <td class="text-right hidden md:table-cell font-mono tabular-nums whitespace-nowrap">
              <span :class="changeColor(bond.last_change_prcnt)">
                {{ formatChange(bond.last_change_prcnt) }}
              </span>
            </td>
            <!-- ОБЪЕМ -->
            <td class="text-right hidden lg:table-cell font-mono tabular-nums">
              <div class="flex flex-col items-end">
                <span class="text-slate-700 dark:text-slate-300">{{ fmt.volume(bond.value_today_rub) }}</span>
                <span v-if="bond.vol_today > 0" class="text-xs text-slate-400 dark:text-slate-500">штук: {{ fmt.num(bond.vol_today) }}</span>
              </div>
            </td>
            <!-- НКД -->
            <td class="text-right hidden lg:table-cell font-mono tabular-nums text-slate-500 dark:text-slate-400">
              {{ bond.accrued_int > 0 ? fmt.num(bond.accrued_int, 2) : '—' }}
            </td>
            <!-- КУПОН -->
            <td class="text-right hidden md:table-cell font-mono tabular-nums">
              <div class="flex flex-col items-end">
                <div class="flex items-center gap-1">
                  <span class="text-slate-900 dark:text-white">{{ fmt.percent(bond.coupon_display) }}</span>
                  <span v-if="bond.is_float" class="badge-sm bg-cyan-50 text-cyan-700 dark:bg-cyan-500/10 dark:text-cyan-400 !text-[9px] !px-1">F</span>
                  <span v-if="bond.is_indexed" class="badge-sm bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400 !text-[9px] !px-1">I</span>
                </div>
                <span class="text-xs text-slate-400 dark:text-slate-500">{{ bond.next_coupon ? fmt.dateShort(bond.next_coupon) : '' }}</span>
              </div>
            </td>
            <!-- ПОГАШЕНИЕ -->
            <td class="text-right hidden lg:table-cell tabular-nums">
              <div class="flex flex-col items-end">
                <span class="text-slate-700 dark:text-slate-300 font-mono text-xs">{{ fmt.date(bond.matdate) }}</span>
                <span :class="bond.days_to_maturity != null && bond.days_to_maturity < 365 ? 'text-amber-500' : 'text-slate-400 dark:text-slate-500'" class="text-xs">
                  {{ fmt.daysToMaturity(bond.days_to_maturity) }}
                </span>
              </div>
            </td>
            <!-- СТАТУС -->
            <td class="text-center hidden xl:table-cell">
              <span :class="tradingStatusClass(bond.trading_status)" class="badge-sm whitespace-nowrap">
                {{ tradingStatusText(bond.trading_status) }}
              </span>
            </td>
            <!-- ДЕЙСТВИЯ -->
            <td class="text-center hidden xl:table-cell">
              <button
                v-if="auth.isLoggedIn.value"
                class="inline-flex items-center justify-center w-8 h-8 rounded-full transition-colors"
                :class="favorites.isFavorite(bond.secid)
                  ? 'text-amber-500 hover:text-amber-600'
                  : 'text-slate-300 hover:text-amber-400 dark:text-slate-600 dark:hover:text-amber-400'"
                :title="favorites.isFavorite(bond.secid) ? 'Убрать из избранного' : 'В избранное'"
                @click.stop="favorites.toggle(bond.secid)"
              >
                <svg class="w-4 h-4" :fill="favorites.isFavorite(bond.secid) ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                </svg>
              </button>
              <NuxtLink
                v-else
                :to="`/bonds/${bond.secid}`"
                class="inline-flex items-center justify-center w-8 h-8 rounded-full border border-slate-200 dark:border-slate-600 text-slate-400 hover:text-primary-500 hover:border-primary-500 dark:hover:text-primary-400 dark:hover:border-primary-400 transition-colors"
                @click.stop
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="px-5 py-3 border-t border-slate-200/80 dark:border-slate-700/40 flex items-center justify-between">
      <button
        :disabled="meta.page <= 1"
        class="btn-secondary text-sm disabled:opacity-30 disabled:pointer-events-none"
        @click="$emit('page', meta.page - 1)"
      >
        ← Назад
      </button>
      <span class="text-xs text-slate-400 dark:text-slate-500 font-mono tabular-nums">
        {{ (meta.page - 1) * meta.per_page + 1 }}–{{ Math.min(meta.page * meta.per_page, meta.total) }} из {{ meta.total }}
      </span>
      <button
        :disabled="meta.page * meta.per_page >= meta.total"
        class="btn-secondary text-sm disabled:opacity-30 disabled:pointer-events-none"
        @click="$emit('page', meta.page + 1)"
      >
        Вперёд →
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{
  bonds: Bond[]
  meta: { page: number; per_page: number; total: number }
  sort: string
}>()

defineEmits<{
  sort: [value: string]
  page: [value: number]
}>()

const fmt = useFormat()
const auth = useAuth()
const favorites = useFavorites()
const currentSort = ref(props.sort ?? 'best')

watch(() => props.sort, (val) => {
  if (val) currentSort.value = val
})

function yieldColor(y: number | null | undefined): string {
  if (y == null) return 'text-slate-400'
  if (y >= 12) return 'text-emerald-600 dark:text-emerald-400'
  if (y >= 8) return 'text-primary-600 dark:text-primary-400'
  return 'text-slate-500 dark:text-slate-400'
}

function categoryBadge(cat: string | undefined): string {
  switch (cat) {
    case 'ОФЗ': return 'bg-blue-500/15 text-blue-600 dark:text-blue-400'
    default: return 'bg-slate-500/15 text-slate-500 dark:text-slate-400'
  }
}

function categoryLabel(cat: string | undefined, isFloat: boolean, isIndexed: boolean): string {
  if (isFloat) return 'Флоатер'
  if (isIndexed) return 'Индексируемая'
  return 'Фикс с известным купоном'
}

function changeColor(val: number | null | undefined): string {
  if (val == null || val === 0) return 'text-slate-400 dark:text-slate-500'
  return val > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500 dark:text-red-400'
}

function formatChange(val: number | null | undefined): string {
  if (val == null) return '—'
  const sign = val > 0 ? '+' : ''
  return sign + val.toFixed(2) + '%'
}

function tradingStatusClass(status: string | undefined): string {
  switch (status) {
    case 'T': return 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'
    case 'S': return 'bg-amber-500/15 text-amber-600 dark:text-amber-400'
    default: return 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400'
  }
}

function tradingStatusText(status: string | undefined): string {
  switch (status) {
    case 'T': return 'Торги идут'
    case 'S': return 'Приостановлены'
    default: return 'Торги не ведутся'
  }
}
</script>
