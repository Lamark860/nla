<template>
  <div class="animate-fade-in">
    <!-- Stat cards -->
    <div class="row g-3 mb-4">
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Доходность к погашению</div>
          <div class="stat-value" :class="yieldColor(bond.yield)">{{ fmt.percent(bond.yield) }}</div>
          <div class="stat-sub">YTM</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">К оферте</div>
          <div class="stat-value" :class="yieldColor(bond.yieldtooffer)">{{ fmt.percent(bond.yieldtooffer) }}</div>
          <div class="stat-sub">{{ bond.offerdate && bond.offerdate !== 'None' ? fmt.date(bond.offerdate) : 'нет оферты' }}</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Дюрация</div>
          <div class="stat-value">{{ bond.duration ?? '—' }}</div>
          <div class="stat-sub">{{ modDuration }}</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Купонная доходность</div>
          <div class="stat-value">{{ couponYield }}</div>
          <div class="stat-sub">годовой купон / цена</div>
        </div>
      </div>
    </div>

    <!-- Yield comparison + Decomposition -->
    <div class="row g-4 mb-4">
      <!-- Yield comparison -->
      <div class="col-lg-8">
        <div class="card p-4">
          <h3 class="section-title mb-4">Сравнение доходностей</h3>
          <div class="d-flex flex-column gap-3">
            <YieldBar label="Доходность к погашению (YTM)" :value="bond.yield" :max="maxYieldForBar" />
            <YieldBar label="К оферте" :value="bond.yieldtooffer" :max="maxYieldForBar" />
            <YieldBar label="Эффективная доходность" :value="effectiveYieldNum" :max="maxYieldForBar" />
            <YieldBar label="Купонная доходность" :value="couponYieldNum" :max="maxYieldForBar" />
            <YieldBar label="Текущая доходность" :value="currentYieldNum" :max="maxYieldForBar" />
            <YieldBar label="По WAP (средневзвешенная)" :value="bond.yieldatwaprice" :max="maxYieldForBar" />
            <YieldBar label="По предыдущей WAP" :value="bond.yieldatprevwaprice" :max="maxYieldForBar" />
          </div>
        </div>
      </div>

      <!-- YTM Decomposition -->
      <div class="col-lg-4">
        <div class="card p-4">
          <h3 class="section-title mb-4">Декомпозиция YTM</h3>
          <div v-if="bond.yield != null" class="d-flex flex-column gap-3">
            <div>
              <div class="d-flex justify-content-between mb-1">
                <span class="small text-muted">Купонный доход</span>
                <span class="small fw-semibold font-monospace" :class="couponYieldNum != null && couponYieldNum > 0 ? 'text-success' : 'text-muted'">{{ couponYield }}</span>
              </div>
              <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--success" :style="{ width: couponPctOfYtm + '%' }"></div></div>
            </div>
            <div>
              <div class="d-flex justify-content-between mb-1">
                <span class="small text-muted">Ценовой {{ priceYieldNum >= 0 ? 'доход' : 'убыток' }}</span>
                <span class="small fw-semibold font-monospace" :class="priceYieldNum >= 0 ? 'text-primary' : 'text-danger'">{{ priceYieldNum.toFixed(2) }}%</span>
              </div>
              <div class="yield-bar"><div :class="priceYieldNum >= 0 ? 'yield-bar__fill--primary' : 'yield-bar__fill--danger'" class="yield-bar__fill" :style="{ width: pricePctOfYtm + '%' }"></div></div>
            </div>
            <div class="pt-3 border-top">
              <div class="d-flex justify-content-between">
                <span class="small fw-medium">Итого YTM</span>
                <span class="small fw-bold font-monospace" :class="yieldColor(bond.yield)">{{ fmt.percent(bond.yield) }}</span>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-muted small py-5">Нет данных</div>
        </div>
      </div>
    </div>

    <!-- Yield indicators table -->
    <div class="card overflow-hidden mb-4">
      <div class="panel-header">
        <i class="bi bi-bar-chart"></i>
        Показатели доходности
      </div>
      <table class="data-table">
        <thead><tr><th class="text-start">Показатель</th><th class="text-end">Значение</th></tr></thead>
        <tbody>
          <tr><td>По WAP (текущая)</td><td class="text-end font-monospace">{{ fmt.percent(bond.yieldatwaprice) }}</td></tr>
          <tr><td>Изменение от пред. дня</td><td class="text-end font-monospace" :class="changeColor(bond.yieldtoprevyield)">{{ fmtChange(bond.yieldtoprevyield) }}</td></tr>
          <tr><td>По предыдущей WAP</td><td class="text-end font-monospace">{{ fmt.percent(bond.yieldatprevwaprice) }}</td></tr>
          <tr><td>К оферте</td><td class="text-end font-monospace">{{ fmt.percent(bond.yieldtooffer) }}</td></tr>
          <tr><td>По последнему купону</td><td class="text-end font-monospace">{{ fmt.percent(bond.yieldlastcoupon) }}</td></tr>
          <tr><td>CALL-опцион доходность</td><td class="text-end font-monospace">{{ fmt.percent(bond.calloptionyield) }}</td></tr>
          <tr><td>Закрытия</td><td class="text-end font-monospace">{{ fmt.percent(bond.closeyield) }}</td></tr>
          <tr><td>Дата расчёта</td><td class="text-end font-monospace">{{ bond.prevdate ? fmt.date(bond.prevdate) : '—' }}</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Spreads and indices -->
    <div class="card overflow-hidden mb-4">
      <div class="panel-header">
        <i class="bi bi-arrows-expand"></i>
        Спреды и индексы
      </div>
      <table class="data-table">
        <thead><tr><th class="text-start">Показатель</th><th class="text-end">Значение</th></tr></thead>
        <tbody>
          <tr><td>Z-spread</td><td class="text-end font-monospace">{{ fmtBps(bond.zspread) }}</td></tr>
          <tr><td>Z-spread по WAP</td><td class="text-end font-monospace">{{ fmtBps(bond.zspreadatwaprice) }}</td></tr>
          <tr><td>Bid/Ask спред</td><td class="text-end font-monospace">{{ bond.spread != null ? bond.spread.toFixed(2) + ' п.п.' : '—' }}</td></tr>
          <tr><td>IR (индекс ИПЦ)</td><td class="text-end font-monospace">{{ fmtBps(bond.iricpiclose) }}</td></tr>
          <tr><td>BEI (инфл. ожидания)</td><td class="text-end font-monospace">{{ fmtBps(bond.beiclose) }}</td></tr>
          <tr><td>CBR (ключ. ставка)</td><td class="text-end font-monospace">{{ fmtBps(bond.cbrclose) }}</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Coupon parameters -->
    <div class="card p-4 mb-4">
      <h3 class="section-title mb-4">Параметры купона</h3>
      <div class="row g-4">
        <div class="col-6 col-lg-3"><div class="stat-label">Ставка купона</div><div class="stat-value mt-1">{{ fmt.percent(bond.coupon_percent) }}</div></div>
        <div class="col-6 col-lg-3"><div class="stat-label">Сумма купона</div><div class="stat-value mt-1">{{ fmt.priceRub(bond.coupon_value) }}</div></div>
        <div class="col-6 col-lg-3"><div class="stat-label">Период купона</div><div class="stat-value mt-1">{{ bond.coupon_period }} <span class="small fw-normal text-muted">дн.</span></div></div>
        <div class="col-6 col-lg-3"><div class="stat-label">НКД</div><div class="stat-value mt-1">{{ fmt.priceRub(bond.accrued_int) }}</div></div>
      </div>
    </div>

    <!-- Timestamps -->
    <div v-if="bond.updatetime || bond.tradetime || bond.systime" class="card p-4">
      <h3 class="section-title mb-3">Временные метки</h3>
      <div class="row g-3">
        <div v-if="bond.updatetime" class="col-sm-4"><div class="stat-label">Обновление данных</div><div class="small font-monospace mt-1">{{ bond.updatetime }}</div></div>
        <div v-if="bond.tradetime" class="col-sm-4"><div class="stat-label">Последняя сделка</div><div class="small font-monospace mt-1">{{ bond.tradetime }}</div></div>
        <div v-if="bond.systime" class="col-sm-4"><div class="stat-label">Системное время</div><div class="small font-monospace mt-1">{{ bond.systime }}</div></div>
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
  return `≈ ${(props.bond.duration / 365).toFixed(2)} лет`
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
  const vals = [props.bond.yield, couponYieldNum.value, currentYieldNum.value, effectiveYieldNum.value, props.bond.yieldatwaprice, props.bond.yieldatprevwaprice, props.bond.yieldtooffer]
    .filter((v): v is number => v != null && v > 0)
  return vals.length ? Math.max(...vals) * 1.2 : 20
})

function yieldColor(y: number | null | undefined): string {
  if (y == null) return 'text-muted'
  if (y >= 15) return 'text-success'
  if (y >= 10) return 'text-primary'
  return ''
}

function changeColor(v: number | null | undefined): string {
  if (v == null) return ''
  return v > 0 ? 'text-success' : v < 0 ? 'text-danger' : ''
}

function fmtChange(v: number | null | undefined): string {
  if (v == null) return '—'
  return (v > 0 ? '+' : '') + v.toFixed(2) + ' п.п.'
}

function fmtBps(v: number | null | undefined): string {
  if (v == null || v === 0) return '—'
  return v.toFixed(0) + ' б.п.'
}

// YieldBar inline component
const YieldBar = defineComponent({
  props: { label: String, value: { type: Number, default: null }, max: { type: Number, default: 20 } },
  setup(p) {
    const pct = computed(() => p.value == null || p.max <= 0 ? 0 : Math.max(0, Math.min(100, (p.value / p.max) * 100)))
    const cls = computed(() => {
      if (p.value == null) return 'yield-bar__fill--primary'
      if (p.value >= 20) return 'yield-bar__fill--danger'
      if (p.value >= 15) return 'yield-bar__fill--warning'
      if (p.value >= 10) return 'yield-bar__fill--success'
      return 'yield-bar__fill--primary'
    })
    return { pct, cls }
  },
  template: `<div>
    <div class="d-flex justify-content-between mb-1"><span class="small text-muted fw-medium">{{ label }}</span><span class="small fw-semibold font-monospace">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span></div>
    <div class="yield-bar"><div :class="cls" class="yield-bar__fill" :style="{ width: pct + '%' }"></div></div>
  </div>`,
})
</script>
