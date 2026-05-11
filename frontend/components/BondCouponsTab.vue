<template>
  <div class="animate-fade-in">
    <!-- Block 1 — Coupon summary (4 KPI cells) -->
    <Panel flush>
      <template #head>
        <div class="cp-head">
          <div class="cp-title">
            <i class="bi bi-calendar3" aria-hidden="true"></i>
            <span>График купонных выплат</span>
          </div>
          <span class="cp-meta">{{ coupons.length }} купонов</span>
        </div>
      </template>

      <div class="cp-summary">
        <div class="cp-cell">
          <div class="cp-lbl">Выплачено</div>
          <div class="cp-val">{{ fmt.priceRub(paidTotal) }}</div>
          <div class="cp-sub">{{ paidCount }} из {{ coupons.length }} купонов</div>
        </div>
        <div class="cp-cell">
          <div class="cp-lbl">Следующий купон</div>
          <div class="cp-val cp-val--up">{{ nextCoupon ? fmt.dateShort(nextCoupon.coupon_date) : '—' }}</div>
          <div v-if="nextCoupon" class="cp-sub">
            через {{ daysToNextCoupon }} дн.
            <template v-if="nextCoupon.value_rub != null"> · {{ fmt.priceRub(nextCoupon.value_rub) }}</template>
          </div>
          <div v-else class="cp-sub">все купоны выплачены</div>
        </div>
        <div class="cp-cell">
          <div class="cp-lbl">PUT-оферта</div>
          <div class="cp-val">{{ offerDate ? fmt.dateShort(offerDate) : '—' }}</div>
          <div class="cp-sub">
            <template v-if="offerDate && daysToOffer != null">через {{ daysToOffer }} дн.</template>
            <template v-else>не предусмотрена</template>
          </div>
        </div>
        <div class="cp-cell">
          <div class="cp-lbl">К погашению</div>
          <div class="cp-val">{{ fmt.priceRub(toMaturityTotal) }}</div>
          <div class="cp-sub">купоны + номинал</div>
        </div>
      </div>

      <div v-if="coupons.length === 0" class="p-5 text-center text-muted">
        Нет данных о купонах
      </div>

      <div v-else class="table-responsive">
        <table class="coupon-tbl">
          <thead>
            <tr>
              <th class="cp-tbl__num">№</th>
              <th>Дата выплаты</th>
              <th>Период</th>
              <th class="right">Сумма</th>
              <th class="right">Дней</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(c, i) in displayedCoupons"
              :key="i"
              :class="{ paid: c.status === 'paid', next: c.status === 'next' }"
            >
              <td class="cp-tbl__num">{{ String(c.originalIndex + 1).padStart(2, '0') }}</td>
              <td class="cp-tbl__date">{{ fmt.dateShort(c.coupon_date) }}</td>
              <td class="cp-tbl__period">{{ periodCellText(c) }}</td>
              <td class="right cp-tbl__sum">
                {{ fmt.priceRub(c.value_rub || c.value || 0) }}
                <span v-if="c.status === 'maturity'" class="cp-tbl__plus-nominal">+ номинал</span>
              </td>
              <td class="right cp-tbl__days">{{ relativeDays(c.coupon_date) }}</td>
              <td>
                <Pill v-if="c.status === 'paid'" tone="default"><i class="bi bi-check-lg"/>Выплачен</Pill>
                <Pill v-else-if="c.status === 'next'" tone="primary"><i class="bi bi-arrow-right-short"/>Ближайший</Pill>
                <Pill v-else-if="c.status === 'put'" tone="warning"><i class="bi bi-flag-fill"/>Оферта</Pill>
                <Pill v-else-if="c.status === 'maturity'" tone="success"><i class="bi bi-flag-fill"/>Погашение</Pill>
                <Pill v-else :dot="false" tone="default">Ожидается</Pill>
              </td>
            </tr>
          </tbody>
        </table>

        <button v-if="hasHidden" class="cp-toggle" @click="showAllCoupons = !showAllCoupons">
          <i :class="showAllCoupons ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
          <span v-if="!showAllCoupons">Показать все {{ couponsWithStatus.length }} купонов · скрыто {{ hiddenCount }}</span>
          <span v-else>Свернуть · показывать только окно</span>
        </button>
      </div>
    </Panel>

    <!-- Block 2 — Coupon parameters -->
    <Panel title="Параметры купона" icon="cash-stack" class="mt-3">
      <InfoRow label="Ставка купона" :value="fmt.percent(bond.coupon_percent)" mono />
      <InfoRow label="Сумма купона" :value="fmt.priceRub(bond.coupon_value)" mono />
      <InfoRow label="Период" :value="periodText" />
      <InfoRow label="Тип купона" :value="couponTypeText" />
      <InfoRow label="НКД" :value="fmt.priceRub(bond.accrued_int)" mono />
      <InfoRow label="Купонная доходность" :value="couponYieldText" mono tone="primary" />
      <InfoRow
        v-if="bond.yieldtooffer != null && offerDate"
        label="Доходность к оферте"
        :value="fmt.percent(bond.yieldtooffer)"
        mono
        tone="primary"
      />
      <InfoRow label="Номинал" :value="fmt.priceRub(bond.facevalue)" mono />
      <InfoRow label="Лот" :value="bond.lotsize ? bond.lotsize + ' шт.' : '—'" mono />
    </Panel>

    <!-- Block 3 — Forecast -->
    <Panel
      v-if="yearlyForecast.length > 0"
      title="Прогноз купонного дохода"
      icon="graph-up-arrow"
      :meta="bond.is_float ? 'плавающая ставка' : 'фикс. купон'"
      class="mt-3"
    >
      <div class="cp-forecast-kpis">
        <KPI label="Купонов осталось" :value="String(couponsRemaining)" />
        <KPI label="Итого до погашения" :value="fmt.priceRub(totalCouponIncome)" tone="primary" />
        <KPI label="В среднем в год" :value="fmt.priceRub(avgYearlyIncome)" />
        <KPI label="Тип" :value="bond.is_float ? 'Флоатер' : 'Фикс'" :sub="fmt.priceRub(bond.coupon_value) + ' / выплата'" />
      </div>

      <table class="data-table cp-forecast-tbl">
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
    </Panel>
  </div>
</template>

<script setup lang="ts">
import type { Coupon, Bond } from '~/composables/useApi'

const props = defineProps<{ coupons: Coupon[]; bond: Bond }>()
const fmt = useFormat()

const today = new Date().toISOString().slice(0, 10)

type CouponStatus = 'paid' | 'next' | 'put' | 'maturity' | 'future'

interface CouponWithStatus extends Coupon {
  status: CouponStatus
  originalIndex: number
}

function isValidDate(v?: string | null): boolean {
  return !!v && v !== '0000-00-00' && v !== 'None' && v !== ''
}

const offerDate = computed<string | null>(() => {
  const d = props.bond.offerdate || props.bond.putoptiondate
  return isValidDate(d) ? d! : null
})

const matDate = computed<string | null>(() => isValidDate(props.bond.matdate) ? props.bond.matdate : null)

const couponsWithStatus = computed<CouponWithStatus[]>(() => {
  if (props.coupons.length === 0) return []
  const sorted = [...props.coupons].sort((a, b) => a.coupon_date.localeCompare(b.coupon_date))
  const future = sorted.filter(c => c.coupon_date >= today)
  const nextDate = future.length ? future[0].coupon_date : null
  const lastDate = sorted[sorted.length - 1]?.coupon_date

  return sorted.map((c, originalIndex) => {
    let status: CouponStatus = 'future'
    if (c.coupon_date < today) status = 'paid'
    else if (offerDate.value && c.coupon_date === offerDate.value) status = 'put'
    else if (matDate.value && c.coupon_date === matDate.value) status = 'maturity'
    else if (c.coupon_date === lastDate && c.coupon_date >= today) status = 'maturity'
    else if (c.coupon_date === nextDate) status = 'next'
    return { ...c, status, originalIndex }
  })
})

// Окно по умолчанию: последний выплаченный + next + 4 следующих + maturity (всегда)
// Если всего <= 7 купонов — показываем всё.
const showAllCoupons = ref(false)
const COLLAPSE_THRESHOLD = 7
const displayedCoupons = computed(() => {
  const all = couponsWithStatus.value
  if (showAllCoupons.value || all.length <= COLLAPSE_THRESHOLD) return all

  const nextIdx = all.findIndex(c => c.status === 'next')
  const matIdx = all.findIndex(c => c.status === 'maturity')

  // Окно: 1 paid до next + next + 4 после, плюс maturity если она вне окна
  const start = nextIdx >= 0 ? Math.max(0, nextIdx - 1) : 0
  const end = nextIdx >= 0 ? Math.min(all.length, nextIdx + 5) : Math.min(5, all.length)
  const window = all.slice(start, end)

  // Add maturity if not already in window
  if (matIdx >= 0 && (matIdx < start || matIdx >= end)) {
    window.push(all[matIdx])
  }
  return window
})

const hasHidden = computed(() => couponsWithStatus.value.length > displayedCoupons.value.length)
const hiddenCount = computed(() => couponsWithStatus.value.length - displayedCoupons.value.length)

const paidCoupons = computed(() => couponsWithStatus.value.filter(c => c.status === 'paid'))
const futureCoupons = computed(() => couponsWithStatus.value.filter(c => c.status !== 'paid'))

const paidCount = computed(() => paidCoupons.value.length)
const paidTotal = computed(() => paidCoupons.value.reduce((s, c) => s + (c.value_rub || c.value || 0), 0))

const nextCoupon = computed(() => couponsWithStatus.value.find(c => c.status === 'next' || c.status === 'maturity') ?? null)

const daysToNextCoupon = computed(() => {
  if (!nextCoupon.value) return null
  return daysFromToday(nextCoupon.value.coupon_date)
})

const daysToOffer = computed(() => offerDate.value ? daysFromToday(offerDate.value) : null)

const toMaturityTotal = computed(() => {
  const couponsSum = futureCoupons.value.reduce((s, c) => s + (c.value_rub || c.value || 0), 0)
  return couponsSum + (props.bond.facevalue || 0)
})

const couponsRemaining = computed(() => futureCoupons.value.length)
const totalCouponIncome = computed(() => futureCoupons.value.reduce((s, c) => s + (c.value_rub || c.value || 0), 0))

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

function daysFromToday(date: string): number {
  const diff = new Date(date).getTime() - Date.now()
  return Math.round(diff / 86400_000)
}

function relativeDays(date: string): string {
  const d = daysFromToday(date)
  if (d === 0) return 'сегодня'
  return d > 0 ? `+${d}` : String(d)
}

function periodCellText(c: Coupon): string {
  // Длина периода в днях между start_date и coupon_date
  if (!c.start_date) return '—'
  const ms = new Date(c.coupon_date).getTime() - new Date(c.start_date).getTime()
  if (!Number.isFinite(ms) || ms <= 0) return '—'
  return Math.round(ms / 86400_000) + ' дн.'
}

function formatPeriodName(days: number): string {
  if (days >= 27 && days <= 33) return 'ежемесячный'
  if (days >= 85 && days <= 95) return 'ежеквартальный'
  if (days >= 175 && days <= 190) return 'полугодовой'
  if (days >= 355 && days <= 370) return 'годовой'
  return ''
}

const periodText = computed(() => {
  const p = props.bond.coupon_period
  if (!p) return '—'
  const name = formatPeriodName(p)
  return name ? `${p} дн. · ${name}` : `${p} дн.`
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

<style scoped>
/* Header */
.cp-head {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}
.cp-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font: 700 13px/1.4 var(--nla-font);
  color: var(--nla-text);
}
.cp-title i { color: var(--nla-primary); font-size: 14px; }
.cp-meta {
  margin-left: auto;
  font: 500 11px/1 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  font-feature-settings: 'tnum';
}

/* Summary 4-cell */
.cp-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border-bottom: 1px solid var(--nla-border);
}
.cp-cell {
  padding: 14px 18px;
  border-right: 1px solid var(--nla-border-light);
}
.cp-cell:last-child { border-right: 0; }
.cp-lbl {
  font: 600 10.5px/1.2 var(--nla-font);
  color: var(--nla-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 5px;
}
.cp-val {
  font: 600 18px/1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  color: var(--nla-text);
}
.cp-val--up { color: var(--nla-success); }
.cp-sub {
  font: 500 11px/1.3 var(--nla-font);
  color: var(--nla-text-muted);
  margin-top: 4px;
  font-feature-settings: 'tnum';
}

@media (max-width: 768px) {
  .cp-summary { grid-template-columns: repeat(2, 1fr); }
  .cp-cell:nth-child(2) { border-right: 0; }
  .cp-cell:nth-child(n+3) { border-top: 1px solid var(--nla-border-light); }
}

/* Coupons table */
.coupon-tbl {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}
.coupon-tbl th {
  background: var(--nla-bg-elevated);
  text-align: left;
  padding: 9px 14px;
  font: 600 10.5px/1 var(--nla-font);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--nla-text-muted);
  border-bottom: 1px solid var(--nla-border);
  white-space: nowrap;
}
.coupon-tbl th.right { text-align: right; }
.coupon-tbl td {
  padding: 11px 14px;
  border-bottom: 1px solid var(--nla-border-light);
  color: var(--nla-text-secondary);
  vertical-align: middle;
}
.coupon-tbl td.right {
  text-align: right;
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
}

.cp-tbl__num,
.cp-tbl__days,
.cp-tbl__date,
.cp-tbl__sum {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.cp-tbl__num { color: var(--nla-text-muted); width: 50px; text-align: right; }
.cp-tbl__days { color: var(--nla-text-muted); width: 70px; }
.cp-tbl__period { color: var(--nla-text-muted); width: 90px; }
.cp-tbl__plus-nominal {
  font-family: var(--nla-font);
  font-weight: 400;
  color: var(--nla-text-muted);
  font-size: 11px;
  margin-left: 6px;
}

.cp-toggle {
  appearance: none;
  width: 100%;
  border: 0;
  border-top: 1px solid var(--nla-border-light);
  background: var(--nla-bg-elevated);
  padding: 11px 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font: 500 12.5px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
}
.cp-toggle:hover {
  background: var(--nla-bg-subtle);
  color: var(--nla-primary);
}
.cp-toggle .bi { font-size: 13px; }

/* Row tones */
.coupon-tbl tr.paid td:not(:last-child) { color: var(--nla-text-muted); }
.coupon-tbl tr.paid td.right { color: var(--nla-text-muted); }
.coupon-tbl tr.next {
  background: var(--nla-primary-light);
}
.coupon-tbl tr.next td:not(:last-child) {
  color: var(--nla-primary-ink);
  font-weight: 600;
}
.coupon-tbl tr.next td.right { color: var(--nla-primary-ink); }

[data-theme="dark"] .coupon-tbl tr.next td:not(:last-child),
[data-theme="dark"] .coupon-tbl tr.next td.right {
  color: var(--nla-primary);
}

/* Pill icon spacing */
:deep(.pill .bi) {
  font-size: 11px;
  line-height: 1;
}

/* Forecast inside Panel */
.cp-forecast-kpis {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  padding: 16px 18px;
  border-bottom: 1px solid var(--nla-border-light);
}
@media (max-width: 768px) {
  .cp-forecast-kpis { grid-template-columns: repeat(2, 1fr); }
}
.cp-forecast-tbl { margin-bottom: 0; }

@media (max-width: 480px) {
  .coupon-tbl th, .coupon-tbl td { padding: 8px 8px; font-size: 12px; }
  .cp-tbl__period { display: none; }
  .cp-tbl__num { width: 32px; }
}
</style>
