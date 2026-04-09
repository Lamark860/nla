<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Coupon parameters stat-cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="stat-card">
        <div class="stat-label">Ставка купона</div>
        <div class="stat-value">{{ fmt.percent(bond.coupon_percent) }}</div>
        <div class="stat-sub">годовых</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Сумма купона</div>
        <div class="stat-value">{{ fmt.priceRub(bond.coupon_value) }}</div>
        <div class="stat-sub">за выплату</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Период</div>
        <div class="stat-value">{{ bond.coupon_period }} <span class="text-sm font-normal text-slate-400">дн.</span></div>
        <div class="stat-sub">{{ formatPeriodName(bond.coupon_period) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Следующий купон</div>
        <div class="stat-value text-base">{{ bond.next_coupon ? fmt.dateShort(bond.next_coupon) : '—' }}</div>
        <div v-if="daysToNextCoupon != null" class="stat-sub">через {{ daysToNextCoupon }} дн.</div>
      </div>
    </div>

    <!-- Key events timeline -->
    <div class="card p-6">
      <h3 class="section-title mb-5">Ключевые события</h3>
      <div class="space-y-4">
        <div v-if="bond.next_coupon" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-primary-50 dark:bg-primary-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Следующий купон</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.next_coupon) }}</span>
            </div>
            <div class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">через {{ daysToNextCoupon ?? '?' }} дн.</div>
          </div>
        </div>
        <!-- PUT offer -->
        <div v-if="bond.offerdate && bond.offerdate !== 'None'" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-amber-50 dark:bg-amber-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9 8.25H7.5a2.25 2.25 0 00-2.25 2.25v9a2.25 2.25 0 002.25 2.25h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25H15m0-3l-3-3m0 0l-3 3m3-3V15"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">PUT-оферта</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.offerdate) }}</span>
            </div>
            <div class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">{{ bond.yieldtooffer != null ? 'Доходность к оферте: ' + bond.yieldtooffer.toFixed(2) + '%' : '' }}</div>
          </div>
        </div>
        <!-- PUT option -->
        <div v-if="bond.putoptiondate && bond.putoptiondate !== 'None'" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-amber-50 dark:bg-amber-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">PUT-опцион</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.putoptiondate) }}</span>
            </div>
          </div>
        </div>
        <!-- CALL option -->
        <div v-if="bond.calloptiondate && bond.calloptiondate !== 'None'" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-cyan-50 dark:bg-cyan-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-cyan-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 6.75c0 8.284 6.716 15 15 15h2.25a2.25 2.25 0 002.25-2.25v-1.372c0-.516-.351-.966-.852-1.091l-4.423-1.106c-.44-.11-.902.055-1.173.417l-.97 1.293c-.282.376-.769.542-1.21.38a12.035 12.035 0 01-7.143-7.143c-.162-.441.004-.928.38-1.21l1.293-.97c.363-.271.527-.734.417-1.173L6.963 3.102a1.125 1.125 0 00-1.091-.852H4.5A2.25 2.25 0 002.25 4.5v2.25z"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">CALL-опцион</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.calloptiondate) }}</span>
            </div>
            <div v-if="bond.calloptionyield != null" class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">Доходность: {{ bond.calloptionyield.toFixed(2) }}%{{ bond.calloptionduration ? ', дюрация: ' + bond.calloptionduration + ' дн.' : '' }}</div>
          </div>
        </div>
        <!-- Buyback -->
        <div v-if="bond.buybackdate && bond.buybackdate !== '0000-00-00' && bond.buybackdate !== 'None'" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-violet-50 dark:bg-violet-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-violet-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Buyback</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.buybackdate) }}</span>
            </div>
            <div v-if="bond.buybackprice" class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">Цена: {{ bond.buybackprice }}%</div>
          </div>
        </div>
        <!-- Maturity -->
        <div v-if="bond.matdate" class="flex items-center gap-4">
          <div class="w-10 h-10 rounded-lg bg-red-50 dark:bg-red-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Погашение</span>
              <span class="text-sm font-mono text-slate-900 dark:text-white">{{ fmt.dateShort(bond.matdate) }}</span>
            </div>
            <div class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">через {{ bond.days_to_maturity ?? '?' }} дн. · {{ couponsRemaining }} купонов осталось</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Coupon parameters table -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z"/><path stroke-linecap="round" stroke-linejoin="round" d="M6 6h.008v.008H6V6z"/></svg>
        Параметры купона
      </div>
      <table class="data-table">
        <tbody>
          <tr>
            <td class="font-medium">Ставка купона</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.coupon_percent) }}</td>
            <td class="font-medium">Номинал</td>
            <td class="text-right font-mono">{{ fmt.priceRub(bond.facevalue) }}</td>
          </tr>
          <tr>
            <td class="font-medium">Сумма купона</td>
            <td class="text-right font-mono">{{ fmt.priceRub(bond.coupon_value) }}</td>
            <td class="font-medium">Лот</td>
            <td class="text-right font-mono">{{ bond.lotsize ? bond.lotsize + ' шт.' : '—' }}</td>
          </tr>
          <tr>
            <td class="font-medium">Период купона</td>
            <td class="text-right font-mono">{{ periodText }}</td>
            <td class="font-medium">Тип</td>
            <td class="text-right font-mono">{{ couponTypeText }}</td>
          </tr>
          <tr>
            <td class="font-medium">НКД</td>
            <td class="text-right font-mono">{{ fmt.priceRub(bond.accrued_int) }}</td>
            <td class="font-medium">Купонная доходность</td>
            <td class="text-right font-mono">{{ couponYieldText }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Forecast by years -->
    <div v-if="yearlyForecast.length > 0" class="card p-6">
      <h3 class="section-title mb-1">Прогноз купонного дохода</h3>
      <p class="text-xs text-slate-400 dark:text-slate-500 mb-5">{{ bond.is_float ? '⚠️ Облигация с плавающей ставкой — прогноз по текущему купону' : 'На основе фиксированного купона' }}</p>

      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div class="stat-card">
          <div class="stat-label">Купонов осталось</div>
          <div class="stat-value">{{ couponsRemaining }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Итого до погашения</div>
          <div class="stat-value text-base">{{ fmt.priceRub(totalCouponIncome) }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">В среднем в год</div>
          <div class="stat-value text-base">{{ fmt.priceRub(avgYearlyIncome) }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Тип</div>
          <div class="stat-value text-base">{{ bond.is_float ? 'Флоатер' : 'Фикс' }}</div>
          <div class="stat-sub">{{ fmt.priceRub(bond.coupon_value) }} / выплата</div>
        </div>
      </div>

      <table class="data-table">
        <thead>
          <tr>
            <th class="text-left">Год</th>
            <th class="text-right">Выплат</th>
            <th class="text-right">Сумма, ₽</th>
            <th class="text-left w-1/3">Доля</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in yearlyForecast" :key="row.year">
            <td class="font-mono font-medium">{{ row.year }}</td>
            <td class="text-right font-mono">{{ row.count }}</td>
            <td class="text-right font-mono font-medium">{{ fmt.priceRub(row.total) }}</td>
            <td>
              <div class="flex items-center gap-2">
                <div class="flex-1 h-2 bg-slate-100 dark:bg-surface-800 rounded-full overflow-hidden">
                  <div class="h-full bg-primary-500 dark:bg-primary-400 rounded-full" :style="{ width: row.pct + '%' }"></div>
                </div>
                <span class="text-xs text-slate-400 font-mono w-10 text-right">{{ row.pct }}%</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Coupon schedule table -->
    <div class="card overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-200/80 dark:border-slate-700/30 flex items-center justify-between">
        <h3 class="section-title">График купонных выплат</h3>
        <span class="text-sm text-slate-400 dark:text-slate-500 font-mono">{{ coupons.length }} купонов</span>
      </div>

      <div v-if="coupons.length === 0" class="p-8 text-center text-slate-500 dark:text-slate-400">
        Нет данных о купонах
      </div>

      <div v-else class="overflow-x-auto">
        <table class="data-table">
          <thead>
            <tr>
              <th class="text-left">Дата выплаты</th>
              <th class="text-left">Начало периода</th>
              <th class="text-right">Ставка, %</th>
              <th class="text-right">Сумма, ₽</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(c, i) in coupons"
              :key="i"
              :class="isPastCoupon(c.coupon_date) ? 'opacity-40' : ''"
            >
              <td>
                <div class="flex items-center gap-2">
                  <span :class="isNextCoupon(c.coupon_date) ? 'w-2 h-2 bg-primary-500 rounded-full animate-pulse-soft' : 'w-2 h-2'" />
                  <span class="font-mono">{{ fmt.dateShort(c.coupon_date) }}</span>
                </div>
              </td>
              <td class="font-mono">{{ fmt.dateShort(c.start_date) }}</td>
              <td class="text-right font-mono tabular-nums">{{ fmt.percent(c.value_percent) }}</td>
              <td class="text-right font-mono tabular-nums font-medium">{{ fmt.priceRub(c.value_rub || c.value) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Coupon, Bond } from '~/composables/useApi'

const props = defineProps<{
  coupons: Coupon[]
  bond: Bond
}>()
const fmt = useFormat()

const today = new Date().toISOString().slice(0, 10)

function isPastCoupon(date: string): boolean {
  return date < today
}

function isNextCoupon(date: string): boolean {
  if (date < today) return false
  const future = props.coupons.filter(c => c.coupon_date >= today)
  return future.length > 0 && future[0].coupon_date === date
}

const daysToNextCoupon = computed(() => {
  if (!props.bond.next_coupon) return null
  const diff = new Date(props.bond.next_coupon).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
})

const futureCoupons = computed(() => props.coupons.filter(c => c.coupon_date >= today))
const couponsRemaining = computed(() => futureCoupons.value.length)
const totalCouponIncome = computed(() => futureCoupons.value.reduce((acc, c) => acc + (c.value_rub || c.value || 0), 0))

const avgYearlyIncome = computed(() => {
  if (!props.bond.days_to_maturity || props.bond.days_to_maturity <= 0) return 0
  const years = props.bond.days_to_maturity / 365
  return years > 0 ? totalCouponIncome.value / years : 0
})

interface YearRow {
  year: number
  count: number
  total: number
  pct: number
}

const yearlyForecast = computed<YearRow[]>(() => {
  if (futureCoupons.value.length === 0) return []
  const map = new Map<number, { count: number; total: number }>()
  for (const c of futureCoupons.value) {
    const year = parseInt(c.coupon_date.slice(0, 4))
    const entry = map.get(year) || { count: 0, total: 0 }
    entry.count++
    entry.total += c.value_rub || c.value || 0
    map.set(year, entry)
  }
  const maxTotal = Math.max(...[...map.values()].map(v => v.total))
  return [...map.entries()]
    .sort(([a], [b]) => a - b)
    .map(([year, data]) => ({
      year,
      count: data.count,
      total: data.total,
      pct: maxTotal > 0 ? Math.round((data.total / maxTotal) * 100) : 0,
    }))
})

function formatPeriodName(days: number): string {
  if (days >= 27 && days <= 33) return 'Ежемесячный'
  if (days >= 85 && days <= 95) return 'Ежеквартальный'
  if (days >= 175 && days <= 190) return 'Полугодовой'
  if (days >= 355 && days <= 370) return 'Годовой'
  return days + ' дней'
}

const periodText = computed(() => {
  const p = props.bond.coupon_period
  if (!p) return '—'
  return `${p} дн. (${formatPeriodName(p)})`
})

const couponTypeText = computed(() => {
  if (props.bond.is_float) return 'Плавающий'
  if (props.bond.is_indexed) return 'Индексируемый'
  return 'Фиксированный'
})

const couponYieldText = computed(() => {
  if (!props.bond.coupon_value || !props.bond.price_rub || props.bond.price_rub <= 0) return '—'
  const cpy = props.bond.coupon_period > 0 ? 365 / props.bond.coupon_period : 2
  return ((props.bond.coupon_value * cpy / props.bond.price_rub) * 100).toFixed(2) + '%'
})
</script>
