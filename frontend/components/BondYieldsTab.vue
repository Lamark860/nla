<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Stat cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="stat-card">
        <div class="stat-label">Доходность к погашению</div>
        <div class="stat-value" :class="yieldColor(bond.yield)">{{ fmt.percent(bond.yield) }}</div>
        <div class="stat-sub">YTM</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">К оферте</div>
        <div class="stat-value" :class="yieldColor(bond.yieldtooffer)">{{ fmt.percent(bond.yieldtooffer) }}</div>
        <div class="stat-sub">{{ bond.offerdate && bond.offerdate !== 'None' ? fmt.date(bond.offerdate) : 'нет оферты' }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Дюрация</div>
        <div class="stat-value">{{ bond.duration ?? '—' }}</div>
        <div class="stat-sub">{{ modDuration }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Купонная доходность</div>
        <div class="stat-value">{{ couponYield }}</div>
        <div class="stat-sub">годовой купон / цена</div>
      </div>
    </div>

    <!-- Yield comparison + Decomposition -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Yield comparison (col-span-2) -->
      <div class="card p-6 lg:col-span-2">
        <h3 class="section-title mb-5">Сравнение доходностей</h3>
        <div class="space-y-5">
          <YieldRow label="Доходность к погашению (YTM)" :value="bond.yield" :max="maxYieldForBar" />
          <YieldRow label="К оферте" :value="bond.yieldtooffer" :max="maxYieldForBar" />
          <YieldRow label="Эффективная доходность" :value="effectiveYieldNum" :max="maxYieldForBar" />
          <YieldRow label="Купонная доходность" :value="couponYieldNum" :max="maxYieldForBar" />
          <YieldRow label="Текущая доходность" :value="currentYieldNum" :max="maxYieldForBar" />
          <YieldRow label="По WAP (средневзвешенная)" :value="bond.yieldatwaprice" :max="maxYieldForBar" />
          <YieldRow label="По предыдущей WAP" :value="bond.yieldatprevwaprice" :max="maxYieldForBar" />
        </div>
      </div>

      <!-- YTM Decomposition -->
      <div class="card p-6">
        <h3 class="section-title mb-5">Декомпозиция YTM</h3>
        <div v-if="bond.yield != null" class="space-y-5">
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs text-slate-500 dark:text-slate-400">Купонный доход</span>
              <span class="text-sm font-semibold font-mono" :class="couponYieldNum != null && couponYieldNum > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400'">{{ couponYield }}</span>
            </div>
            <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
              <div class="h-full bg-emerald-500 dark:bg-emerald-400 rounded-lg" :style="{ width: couponPctOfYtm + '%' }"></div>
            </div>
          </div>
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs text-slate-500 dark:text-slate-400">Ценовой {{ priceYieldNum >= 0 ? 'доход' : 'убыток' }}</span>
              <span class="text-sm font-semibold font-mono" :class="priceYieldNum >= 0 ? 'text-blue-600 dark:text-blue-400' : 'text-red-600 dark:text-red-400'">{{ priceYieldNum.toFixed(2) }}%</span>
            </div>
            <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
              <div :class="priceYieldNum >= 0 ? 'bg-blue-500 dark:bg-blue-400' : 'bg-red-500 dark:bg-red-400'" class="h-full rounded-lg" :style="{ width: pricePctOfYtm + '%' }"></div>
            </div>
          </div>
          <div class="pt-3 border-t border-slate-200/60 dark:border-slate-700/30">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium text-slate-700 dark:text-slate-300">Итого YTM</span>
              <span class="text-sm font-bold font-mono" :class="yieldColor(bond.yield)">{{ fmt.percent(bond.yield) }}</span>
            </div>
          </div>
        </div>
        <div v-else class="text-center text-sm text-slate-400 dark:text-slate-500 py-6">Нет данных</div>
      </div>
    </div>

    <!-- Yield indicators table -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"/></svg>
        Показатели доходности
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th class="text-left">Показатель</th>
            <th class="text-right">Значение</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>По WAP (текущая)</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.yieldatwaprice) }}</td>
          </tr>
          <tr>
            <td>Изменение от пред. дня</td>
            <td class="text-right font-mono" :class="changeColor(bond.yieldtoprevyield)">{{ fmtChange(bond.yieldtoprevyield) }}</td>
          </tr>
          <tr>
            <td>По предыдущей WAP</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.yieldatprevwaprice) }}</td>
          </tr>
          <tr>
            <td>К оферте</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.yieldtooffer) }}</td>
          </tr>
          <tr>
            <td>По последнему купону</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.yieldlastcoupon) }}</td>
          </tr>
          <tr>
            <td>CALL-опцион доходность</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.calloptionyield) }}</td>
          </tr>
          <tr>
            <td>Закрытия</td>
            <td class="text-right font-mono">{{ fmt.percent(bond.closeyield) }}</td>
          </tr>
          <tr>
            <td>Дата расчёта</td>
            <td class="text-right font-mono">{{ bond.prevdate ? fmt.date(bond.prevdate) : '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Spreads and indices -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"/></svg>
        Спреды и индексы
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th class="text-left">Показатель</th>
            <th class="text-right">Значение</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Z-spread</td>
            <td class="text-right font-mono">{{ fmtBps(bond.zspread) }}</td>
          </tr>
          <tr>
            <td>Z-spread по WAP</td>
            <td class="text-right font-mono">{{ fmtBps(bond.zspreadatwaprice) }}</td>
          </tr>
          <tr>
            <td>Bid/Ask спред</td>
            <td class="text-right font-mono">{{ bond.spread != null ? bond.spread.toFixed(2) + ' п.п.' : '—' }}</td>
          </tr>
          <tr>
            <td>IR (индекс ИПЦ)</td>
            <td class="text-right font-mono">{{ fmtBps(bond.iricpiclose) }}</td>
          </tr>
          <tr>
            <td>BEI (инфл. ожидания)</td>
            <td class="text-right font-mono">{{ fmtBps(bond.beiclose) }}</td>
          </tr>
          <tr>
            <td>CBR (ключ. ставка)</td>
            <td class="text-right font-mono">{{ fmtBps(bond.cbrclose) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Coupon parameters -->
    <div class="card p-6">
      <h3 class="section-title mb-5">Параметры купона</h3>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-6">
        <div>
          <div class="stat-label">Ставка купона</div>
          <div class="stat-value mt-1">{{ fmt.percent(bond.coupon_percent) }}</div>
        </div>
        <div>
          <div class="stat-label">Сумма купона</div>
          <div class="stat-value mt-1">{{ fmt.priceRub(bond.coupon_value) }}</div>
        </div>
        <div>
          <div class="stat-label">Период купона</div>
          <div class="stat-value mt-1">{{ bond.coupon_period }} <span class="text-sm font-normal text-slate-400">дн.</span></div>
        </div>
        <div>
          <div class="stat-label">НКД</div>
          <div class="stat-value mt-1">{{ fmt.priceRub(bond.accrued_int) }}</div>
        </div>
      </div>
    </div>

    <!-- Timestamps -->
    <div v-if="bond.updatetime || bond.tradetime || bond.systime" class="card p-6">
      <h3 class="section-title mb-4">Временные метки</h3>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div v-if="bond.updatetime">
          <div class="stat-label">Обновление данных</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text);">{{ bond.updatetime }}</div>
        </div>
        <div v-if="bond.tradetime">
          <div class="stat-label">Последняя сделка</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text);">{{ bond.tradetime }}</div>
        </div>
        <div v-if="bond.systime">
          <div class="stat-label">Системное время</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text);">{{ bond.systime }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{ bond: Bond }>()
const fmt = useFormat()

const effectiveYieldNum = computed<number | null>(() => {
  if (props.bond.yield == null) return null
  const couponsPerYear = props.bond.coupon_period > 0 ? 365 / props.bond.coupon_period : 2
  return (Math.pow(1 + (props.bond.yield / 100) / couponsPerYear, couponsPerYear) - 1) * 100
})

const modDuration = computed(() => {
  if (props.bond.duration == null || props.bond.yield == null) return '—'
  const years = props.bond.duration / 365
  return `≈ ${years.toFixed(2)} лет`
})

const couponYieldNum = computed<number | null>(() => {
  if (props.bond.coupon_value <= 0 || props.bond.price_rub == null || props.bond.price_rub <= 0) return null
  const couponsPerYear = props.bond.coupon_period > 0 ? 365 / props.bond.coupon_period : 2
  return (props.bond.coupon_value * couponsPerYear / props.bond.price_rub) * 100
})

const couponYield = computed(() => couponYieldNum.value != null ? couponYieldNum.value.toFixed(2) + '%' : '—')

const currentYieldNum = computed<number | null>(() => {
  if (props.bond.coupon_value <= 0 || props.bond.price_rub == null || props.bond.price_rub <= 0) return null
  const couponsPerYear = props.bond.coupon_period > 0 ? 365 / props.bond.coupon_period : 2
  const totalPrice = props.bond.price_rub + (props.bond.accrued_int ?? 0)
  return (props.bond.coupon_value * couponsPerYear / totalPrice) * 100
})

// YTM Decomposition
const priceYieldNum = computed(() => {
  if (props.bond.yield == null || couponYieldNum.value == null) return 0
  return props.bond.yield - couponYieldNum.value
})

const couponPctOfYtm = computed(() => {
  if (props.bond.yield == null || props.bond.yield === 0 || couponYieldNum.value == null) return 0
  return Math.max(0, Math.min(100, Math.abs(couponYieldNum.value / props.bond.yield) * 100))
})

const pricePctOfYtm = computed(() => {
  if (props.bond.yield == null || props.bond.yield === 0) return 0
  return Math.max(0, Math.min(100, Math.abs(priceYieldNum.value / props.bond.yield) * 100))
})

const maxYieldForBar = computed(() => {
  const vals = [
    props.bond.yield, couponYieldNum.value, currentYieldNum.value, effectiveYieldNum.value,
    props.bond.yieldatwaprice, props.bond.yieldatprevwaprice, props.bond.yieldtooffer,
  ].filter((v): v is number => v != null && v > 0)
  return vals.length ? Math.max(...vals) * 1.2 : 20
})

function yieldColor(y: number | null | undefined): string {
  if (y == null) return 'text-slate-400'
  if (y >= 15) return 'text-emerald-600 dark:text-emerald-400'
  if (y >= 10) return 'text-primary-600 dark:text-primary-400'
  return 'text-slate-700 dark:text-slate-300'
}

function changeColor(v: number | null | undefined): string {
  if (v == null) return ''
  return v > 0 ? 'text-emerald-600 dark:text-emerald-400' : v < 0 ? 'text-red-600 dark:text-red-400' : ''
}

function fmtChange(v: number | null | undefined): string {
  if (v == null) return '—'
  const sign = v > 0 ? '+' : ''
  return sign + v.toFixed(2) + ' п.п.'
}

function fmtBps(v: number | null | undefined): string {
  if (v == null || v === 0) return '—'
  return v.toFixed(0) + ' б.п.'
}

const YieldRow = defineComponent({
  props: {
    label: String,
    value: { type: Number, default: null },
    max: { type: Number, default: 20 },
  },
  setup(props) {
    const pct = computed(() => {
      if (props.value == null || props.max <= 0) return 0
      return Math.max(0, Math.min(100, (props.value / props.max) * 100))
    })
    const color = computed(() => {
      if (props.value == null) return 'bg-slate-300 dark:bg-slate-600'
      if (props.value >= 20) return 'bg-red-500 dark:bg-red-400'
      if (props.value >= 15) return 'bg-amber-500 dark:bg-amber-400'
      if (props.value >= 10) return 'bg-emerald-500 dark:bg-emerald-400'
      return 'bg-primary-500 dark:bg-primary-400'
    })
    return { pct, color, fmt: useFormat() }
  },
  template: `
    <div>
      <div class="flex items-center justify-between mb-1.5">
        <span class="text-xs text-slate-500 dark:text-slate-400 font-medium">{{ label }}</span>
        <span class="text-sm font-semibold text-slate-900 dark:text-white tabular-nums font-mono">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span>
      </div>
      <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
        <div :class="color" class="h-full rounded-lg transition-all duration-500 ease-out" :style="{ width: pct + '%' }"></div>
      </div>
    </div>
  `,
})
</script>
