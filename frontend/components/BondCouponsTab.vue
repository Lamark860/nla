<template>
  <div class="animate-fade-in">
    <!-- Coupon parameters stat-cards -->
    <div class="row g-3 mb-4">
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Ставка купона</div>
          <div class="stat-value">{{ fmt.percent(bond.coupon_percent) }}</div>
          <div class="stat-sub">годовых</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Сумма купона</div>
          <div class="stat-value">{{ fmt.priceRub(bond.coupon_value) }}</div>
          <div class="stat-sub">за выплату</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Период</div>
          <div class="stat-value">{{ bond.coupon_period }} <span class="small fw-normal text-muted">дн.</span></div>
          <div class="stat-sub">{{ formatPeriodName(bond.coupon_period) }}</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Следующий купон</div>
          <div class="stat-value" style="font-size: 1rem">{{ bond.next_coupon ? fmt.dateShort(bond.next_coupon) : '—' }}</div>
          <div v-if="daysToNextCoupon != null" class="stat-sub">через {{ daysToNextCoupon }} дн.</div>
        </div>
      </div>
    </div>

    <!-- Key events -->
    <div class="card p-4 mb-4">
      <h3 class="section-title mb-4">Ключевые события</h3>
      <div class="d-flex flex-column gap-3">
        <div v-if="bond.next_coupon" class="d-flex align-items-center gap-3">
          <div class="rounded d-flex align-items-center justify-content-center flex-shrink-0" style="width: 40px; height: 40px; background: var(--nla-primary-subtle)">
            <i class="bi bi-clock text-primary"></i>
          </div>
          <div class="flex-grow-1">
            <div class="d-flex justify-content-between"><span class="small fw-medium">Следующий купон</span><span class="small font-monospace">{{ fmt.dateShort(bond.next_coupon) }}</span></div>
            <div class="small text-muted">через {{ daysToNextCoupon ?? '?' }} дн.</div>
          </div>
        </div>
        <div v-if="bond.offerdate && bond.offerdate !== 'None'" class="d-flex align-items-center gap-3">
          <div class="rounded d-flex align-items-center justify-content-center flex-shrink-0" style="width: 40px; height: 40px; background: var(--nla-warning-light)">
            <i class="bi bi-arrow-down-up text-warning"></i>
          </div>
          <div class="flex-grow-1">
            <div class="d-flex justify-content-between"><span class="small fw-medium">PUT-оферта</span><span class="small font-monospace">{{ fmt.dateShort(bond.offerdate) }}</span></div>
            <div v-if="bond.yieldtooffer != null" class="small text-muted">Доходность к оферте: {{ bond.yieldtooffer.toFixed(2) }}%</div>
          </div>
        </div>
        <div v-if="bond.matdate" class="d-flex align-items-center gap-3">
          <div class="rounded d-flex align-items-center justify-content-center flex-shrink-0" style="width: 40px; height: 40px; background: var(--nla-danger-light)">
            <i class="bi bi-x-circle text-danger"></i>
          </div>
          <div class="flex-grow-1">
            <div class="d-flex justify-content-between"><span class="small fw-medium">Погашение</span><span class="small font-monospace">{{ fmt.dateShort(bond.matdate) }}</span></div>
            <div class="small text-muted">через {{ bond.days_to_maturity ?? '?' }} дн. · {{ couponsRemaining }} купонов осталось</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Coupon parameters table -->
    <div class="card overflow-hidden mb-4">
      <div class="panel-header">
        <i class="bi bi-tag"></i>
        Параметры купона
      </div>
      <table class="data-table">
        <tbody>
          <tr><td class="fw-medium">Ставка купона</td><td class="text-end font-monospace">{{ fmt.percent(bond.coupon_percent) }}</td><td class="fw-medium">Номинал</td><td class="text-end font-monospace">{{ fmt.priceRub(bond.facevalue) }}</td></tr>
          <tr><td class="fw-medium">Сумма купона</td><td class="text-end font-monospace">{{ fmt.priceRub(bond.coupon_value) }}</td><td class="fw-medium">Лот</td><td class="text-end font-monospace">{{ bond.lotsize ? bond.lotsize + ' шт.' : '—' }}</td></tr>
          <tr><td class="fw-medium">Период купона</td><td class="text-end font-monospace">{{ periodText }}</td><td class="fw-medium">Тип</td><td class="text-end font-monospace">{{ couponTypeText }}</td></tr>
          <tr><td class="fw-medium">НКД</td><td class="text-end font-monospace">{{ fmt.priceRub(bond.accrued_int) }}</td><td class="fw-medium">Купонная доходность</td><td class="text-end font-monospace">{{ couponYieldText }}</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Forecast by years -->
    <div v-if="yearlyForecast.length > 0" class="card p-4 mb-4">
      <h3 class="section-title mb-1">Прогноз купонного дохода</h3>
      <p class="small text-muted mb-4">{{ bond.is_float ? '⚠️ Облигация с плавающей ставкой — прогноз по текущему купону' : 'На основе фиксированного купона' }}</p>

      <div class="row g-3 mb-4">
        <div class="col-6 col-lg-3"><div class="stat-card"><div class="stat-label">Купонов осталось</div><div class="stat-value">{{ couponsRemaining }}</div></div></div>
        <div class="col-6 col-lg-3"><div class="stat-card"><div class="stat-label">Итого до погашения</div><div class="stat-value" style="font-size: 1rem">{{ fmt.priceRub(totalCouponIncome) }}</div></div></div>
        <div class="col-6 col-lg-3"><div class="stat-card"><div class="stat-label">В среднем в год</div><div class="stat-value" style="font-size: 1rem">{{ fmt.priceRub(avgYearlyIncome) }}</div></div></div>
        <div class="col-6 col-lg-3"><div class="stat-card"><div class="stat-label">Тип</div><div class="stat-value" style="font-size: 1rem">{{ bond.is_float ? 'Флоатер' : 'Фикс' }}</div><div class="stat-sub">{{ fmt.priceRub(bond.coupon_value) }} / выплата</div></div></div>
      </div>

      <table class="data-table">
        <thead><tr><th class="text-start">Год</th><th class="text-end">Выплат</th><th class="text-end">Сумма, ₽</th><th class="text-start" style="width: 33%">Доля</th></tr></thead>
        <tbody>
          <tr v-for="row in yearlyForecast" :key="row.year">
            <td class="font-monospace fw-medium">{{ row.year }}</td>
            <td class="text-end font-monospace">{{ row.count }}</td>
            <td class="text-end font-monospace fw-medium">{{ fmt.priceRub(row.total) }}</td>
            <td>
              <div class="d-flex align-items-center gap-2">
                <div class="flex-grow-1 nla-progress">
                  <div class="nla-progress__bar" :style="{ width: row.pct + '%' }"></div>
                </div>
                <span class="small text-muted font-monospace" style="width: 2.5rem; text-align: right">{{ row.pct }}%</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Coupon schedule table -->
    <div class="card overflow-hidden">
      <div class="panel-header justify-content-between">
        <div class="d-flex align-items-center gap-2">
          <i class="bi bi-list-ul"></i>
          График купонных выплат
        </div>
        <span class="small text-muted font-monospace">{{ coupons.length }} купонов</span>
      </div>

      <div v-if="coupons.length === 0" class="p-5 text-center text-muted">
        Нет данных о купонах
      </div>

      <div v-else class="table-responsive">
        <table class="data-table">
          <thead><tr><th class="text-start">Дата выплаты</th><th class="text-start">Начало периода</th><th class="text-end">Ставка, %</th><th class="text-end">Сумма, ₽</th></tr></thead>
          <tbody>
            <tr v-for="(c, i) in coupons" :key="i" :class="isPastCoupon(c.coupon_date) ? 'opacity-25' : ''">
              <td>
                <div class="d-flex align-items-center gap-2">
                  <span v-if="isNextCoupon(c.coupon_date)" class="rounded-circle bg-primary animate-pulse-soft" style="width: 8px; height: 8px"></span>
                  <span v-else style="width: 8px"></span>
                  <span class="font-monospace">{{ fmt.dateShort(c.coupon_date) }}</span>
                </div>
              </td>
              <td class="font-monospace">{{ fmt.dateShort(c.start_date) }}</td>
              <td class="text-end font-monospace">{{ fmt.percent(c.value_percent) }}</td>
              <td class="text-end font-monospace fw-medium">{{ fmt.priceRub(c.value_rub || c.value) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Coupon, Bond } from '~/composables/useApi'

const props = defineProps<{ coupons: Coupon[]; bond: Bond }>()
const fmt = useFormat()

const today = new Date().toISOString().slice(0, 10)

function isPastCoupon(date: string): boolean { return date < today }
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

interface YearRow { year: number; count: number; total: number; pct: number }
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
  return [...map.entries()].sort(([a], [b]) => a - b).map(([year, data]) => ({
    year, count: data.count, total: data.total, pct: maxTotal > 0 ? Math.round((data.total / maxTotal) * 100) : 0,
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
  return p ? `${p} дн. (${formatPeriodName(p)})` : '—'
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
