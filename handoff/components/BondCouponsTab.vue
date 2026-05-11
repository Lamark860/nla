<!--
  BondCouponsTab.vue — таб "Купоны".
  Сводка сверху + таблица всех купонных выплат.
  Pill для статуса выплаты, mono для денег и дат.
-->
<template>
  <div class="coupons-tab">
    <!-- Summary row -->
    <div class="coupons-summary">
      <div class="cs-stat">
        <div class="cs-stat__lbl">Всего купонов</div>
        <div class="cs-stat__val">{{ total }}</div>
      </div>
      <div class="cs-stat">
        <div class="cs-stat__lbl">Выплачено</div>
        <div class="cs-stat__val cs-stat__val--success">{{ paid }}</div>
        <div class="cs-stat__sub">{{ fmt.priceRub(paidSum) }}</div>
      </div>
      <div class="cs-stat">
        <div class="cs-stat__lbl">Ожидается</div>
        <div class="cs-stat__val">{{ pending }}</div>
        <div class="cs-stat__sub">{{ fmt.priceRub(pendingSum) }}</div>
      </div>
      <div class="cs-stat">
        <div class="cs-stat__lbl">Следующий</div>
        <div class="cs-stat__val cs-stat__val--primary">{{ nextDate ? fmt.dateShort(nextDate) : '—' }}</div>
        <div v-if="nextDays != null" class="cs-stat__sub">через {{ nextDays }} дн</div>
      </div>
    </div>

    <Panel title="График купонов" icon="cash-stack" flush>
      <table class="coupons-table">
        <thead>
          <tr>
            <th class="t-num">№</th>
            <th>Дата</th>
            <th class="t-right">Купон, ₽</th>
            <th class="t-right">Ставка</th>
            <th class="t-right">Период, дн</th>
            <th>Статус</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(c, i) in coupons" :key="i" :class="{ 'is-next': c === next }">
            <td class="t-num">{{ i + 1 }}</td>
            <td class="t-mono">{{ fmt.dateShort(c.coupondate) }}</td>
            <td class="t-right t-mono">{{ fmt.priceRub(c.value_rub) }}</td>
            <td class="t-right t-mono">{{ c.value_pct != null ? fmt.percent(c.value_pct) : '—' }}</td>
            <td class="t-right t-mono t-muted">{{ c.coupon_period || '—' }}</td>
            <td>
              <Pill v-if="c.status === 'paid'" tone="success">Выплачен</Pill>
              <Pill v-else-if="c === next" tone="primary">Следующий</Pill>
              <Pill v-else tone="default" :dot="false">Ожидается</Pill>
            </td>
          </tr>
        </tbody>
      </table>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import Panel from './Panel.vue'
import Pill from './Pill.vue'

interface Coupon {
  coupondate: string
  value_rub: number
  value_pct?: number
  coupon_period?: number
  status?: 'paid' | 'pending'
}

const props = defineProps<{ coupons: Coupon[] }>()
const fmt = useFormat()

const total = computed(() => props.coupons.length)
const paidList = computed(() => props.coupons.filter(c => c.status === 'paid'))
const pendingList = computed(() => props.coupons.filter(c => c.status !== 'paid'))
const paid = computed(() => paidList.value.length)
const pending = computed(() => pendingList.value.length)
const paidSum = computed(() => paidList.value.reduce((s, c) => s + (c.value_rub || 0), 0))
const pendingSum = computed(() => pendingList.value.reduce((s, c) => s + (c.value_rub || 0), 0))
const next = computed(() => pendingList.value[0])
const nextDate = computed(() => next.value?.coupondate)
const nextDays = computed(() => {
  if (!nextDate.value) return null
  const d = new Date(nextDate.value)
  const now = new Date()
  const days = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  return days > 0 ? days : null
})
</script>

<style scoped>
.coupons-tab { display: flex; flex-direction: column; gap: 16px; }

.coupons-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1px;
  background: var(--nla-border);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-lg);
  overflow: hidden;
}
.cs-stat {
  background: var(--nla-bg-card);
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.cs-stat__lbl {
  font: 600 10.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
.cs-stat__val {
  font: 600 22px / 1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  letter-spacing: -0.02em;
}
.cs-stat__val--success { color: var(--nla-success); }
.cs-stat__val--primary { color: var(--nla-primary); }
.cs-stat__sub {
  font: 500 12px / 1.2 var(--nla-font);
  font-feature-settings: 'tnum';
  color: var(--nla-text-muted);
}

@media (max-width: 768px) {
  .coupons-summary { grid-template-columns: repeat(2, 1fr); }
}

/* Coupons table */
.coupons-table {
  width: 100%;
  border-collapse: collapse;
  font: 500 13px / 1.4 var(--nla-font);
  color: var(--nla-text);
}
.coupons-table thead th {
  font: 600 10.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  text-align: left;
  padding: 12px 14px;
  background: var(--nla-bg-elevated);
  border-bottom: 1px solid var(--nla-border);
  position: sticky;
  top: 0;
}
.coupons-table tbody td {
  padding: 11px 14px;
  border-top: 1px solid var(--nla-border-light);
  vertical-align: middle;
}
.coupons-table tbody tr:hover td { background: var(--nla-bg-hover); }
.coupons-table tbody tr.is-next td {
  background: var(--nla-primary-light);
  font-weight: 600;
}
.t-mono { font-family: var(--nla-font-mono); font-feature-settings: 'tnum'; }
.t-num { font-family: var(--nla-font-mono); color: var(--nla-text-muted); width: 50px; }
.t-right { text-align: right; }
.t-muted { color: var(--nla-text-muted); }
</style>
