<!--
  BondTradingTab.vue — стакан + лента сделок.
  Левая колонка: стакан bid/ask. Правая: лента с BUY/SELL pills.
-->
<template>
  <div class="trading-tab">
    <Panel title="Стакан" icon="bar-chart-steps" flush>
      <template #headRight>
        <span class="trading__spread">
          spread <strong>{{ spread }}</strong>
        </span>
      </template>

      <div class="orderbook">
        <!-- ASK side (red), top = best ask, перевёрнуто -->
        <div class="ob-side ob-side--ask">
          <div v-for="row in asks" :key="`a${row.price}`" class="ob-row">
            <span class="ob-row__bar" :style="{ width: barWidth(row.qty, maxQty) + '%' }"></span>
            <span class="ob-row__qty">{{ fmt.num(row.qty) }}</span>
            <span class="ob-row__price ob-row__price--ask">{{ row.price.toFixed(2) }}</span>
          </div>
        </div>

        <!-- mid -->
        <div class="ob-mid">
          <div class="ob-mid__last">{{ last?.toFixed(2) ?? '—' }}</div>
          <div class="ob-mid__lbl">последняя цена</div>
        </div>

        <!-- BID side (green), top = best bid -->
        <div class="ob-side ob-side--bid">
          <div v-for="row in bids" :key="`b${row.price}`" class="ob-row">
            <span class="ob-row__price ob-row__price--bid">{{ row.price.toFixed(2) }}</span>
            <span class="ob-row__qty">{{ fmt.num(row.qty) }}</span>
            <span class="ob-row__bar" :style="{ width: barWidth(row.qty, maxQty) + '%' }"></span>
          </div>
        </div>
      </div>
    </Panel>

    <Panel title="Лента сделок" icon="lightning-charge" flush>
      <template #headRight>
        <Pill tone="live">live</Pill>
      </template>
      <table class="trades-table">
        <thead>
          <tr>
            <th>Время</th>
            <th class="t-right">Цена</th>
            <th class="t-right">Кол-во</th>
            <th class="t-right">Сумма</th>
            <th>Сторона</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in trades" :key="t.id">
            <td class="t-mono t-muted">{{ t.time }}</td>
            <td class="t-mono t-right">{{ t.price.toFixed(2) }}</td>
            <td class="t-mono t-right">{{ fmt.num(t.qty) }}</td>
            <td class="t-mono t-right">{{ fmt.priceRub(t.price * t.qty * 10) }}</td>
            <td>
              <Pill :tone="t.side === 'B' ? 'success' : 'danger'">
                {{ t.side === 'B' ? 'BUY' : 'SELL' }}
              </Pill>
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

interface Level { price: number; qty: number }
interface Trade { id: string | number; time: string; price: number; qty: number; side: 'B' | 'S' }

const props = defineProps<{
  bids: Level[]
  asks: Level[]
  trades: Trade[]
  last?: number | null
}>()
const fmt = useFormat()

const maxQty = computed(() => Math.max(
  ...props.bids.map(b => b.qty),
  ...props.asks.map(a => a.qty),
  1
))
function barWidth(qty: number, max: number) { return Math.min(100, (qty / max) * 100) }

const spread = computed(() => {
  const a = props.asks[0]?.price
  const b = props.bids[0]?.price
  if (a == null || b == null) return '—'
  return (a - b).toFixed(2)
})
</script>

<style scoped>
.trading-tab {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 16px;
}
@media (max-width: 992px) {
  .trading-tab { grid-template-columns: 1fr; }
}

.trading__spread {
  font: 500 12px / 1 var(--nla-font);
  color: var(--nla-text-muted);
}
.trading__spread strong {
  font-family: var(--nla-font-mono);
  color: var(--nla-text);
  font-weight: 600;
  margin-left: 4px;
}

/* Orderbook */
.orderbook { display: flex; flex-direction: column; }
.ob-side { display: flex; flex-direction: column; }
.ob-side--ask { flex-direction: column-reverse; }

.ob-row {
  position: relative;
  display: grid;
  grid-template-columns: 80px 1fr 80px;
  align-items: center;
  height: 28px;
  padding: 0 14px;
  font: 500 12px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.ob-row__bar {
  position: absolute;
  inset: 2px auto 2px auto;
  pointer-events: none;
  border-radius: 4px;
  opacity: 0.18;
}
.ob-side--ask .ob-row__bar { right: 14px; background: var(--nla-danger); }
.ob-side--bid .ob-row__bar { left: 14px;  background: var(--nla-success); }

.ob-row__qty   { color: var(--nla-text-muted); text-align: right; padding: 0 8px; position: relative; }
.ob-row__price { font-weight: 600; position: relative; }
.ob-side--ask .ob-row__qty   { order: 2; }
.ob-side--ask .ob-row__price { order: 3; text-align: right; }
.ob-row__price--ask { color: var(--nla-danger); }
.ob-row__price--bid { color: var(--nla-success); }

.ob-mid {
  background: var(--nla-bg-elevated);
  border-top: 1px solid var(--nla-border);
  border-bottom: 1px solid var(--nla-border);
  padding: 12px 14px;
  text-align: center;
}
.ob-mid__last {
  font: 700 22px / 1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  letter-spacing: -0.02em;
  color: var(--nla-primary);
}
.ob-mid__lbl {
  font: 600 10px / 1 var(--nla-font);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  margin-top: 2px;
}

/* Trades */
.trades-table { width: 100%; border-collapse: collapse; }
.trades-table thead th {
  font: 600 10.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  text-align: left;
  padding: 12px 14px;
  background: var(--nla-bg-elevated);
  border-bottom: 1px solid var(--nla-border);
}
.trades-table tbody td {
  padding: 9px 14px;
  border-top: 1px solid var(--nla-border-light);
  font: 500 12.5px / 1.2 var(--nla-font);
}
.trades-table tbody tr:hover td { background: var(--nla-bg-hover); }
.t-mono { font-family: var(--nla-font-mono); font-feature-settings: 'tnum'; }
.t-right { text-align: right; }
.t-muted { color: var(--nla-text-muted); }
</style>
