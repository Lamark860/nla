<template>
  <div class="animate-fade-in">
    <!-- 4 KPI top -->
    <div class="td-kpis">
      <KPI label="Статус" :tone="bond.trading_status === 'T' ? 'success' : 'muted'">
        <template #default>{{ bond.trading_status === 'T' ? 'Торгуется' : 'Нет торгов' }}</template>
      </KPI>
      <KPI label="Спрос (Bid)" :value="fmt.percent(bond.bid)">
        <template v-if="bond.bid != null && bond.facevalue" #sub>
          <span class="font-monospace">{{ fmt.priceRub(bond.bid / 100 * bond.facevalue) }}</span>
          <template v-if="bond.biddeptht"> · глуб. {{ fmt.num(bond.biddeptht) }}</template>
        </template>
      </KPI>
      <KPI label="Предложение (Ask)" :value="fmt.percent(bond.offer)">
        <template v-if="bond.offer != null && bond.facevalue" #sub>
          <span class="font-monospace">{{ fmt.priceRub(bond.offer / 100 * bond.facevalue) }}</span>
          <template v-if="bond.offerdeptht"> · глуб. {{ fmt.num(bond.offerdeptht) }}</template>
        </template>
      </KPI>
      <KPI label="Спред" :value="spreadStr">
        <template v-if="spreadPct != null" #sub>
          <span class="font-monospace">{{ spreadPct.toFixed(2) }}%</span> от бида
        </template>
      </KPI>
    </div>

    <!-- Backend-блокер: полный orderbook + лента сделок отсутствуют. См. docs/roadmap.md Phase D -->

    <!-- 2 Panel: Prices + Depth -->
    <div class="td-row td-row--2col">
      <Panel title="Цены торгового дня" icon="bar-chart-line">
        <div class="td-pricebars">
          <PriceBar label="Открытие" :value="bond.open" :min="priceMin" :max="priceMax" variant="secondary" />
          <PriceBar label="Минимум" :value="bond.low" :min="priceMin" :max="priceMax" variant="danger" />
          <PriceBar label="Максимум" :value="bond.high" :min="priceMin" :max="priceMax" variant="success" />
          <PriceBar label="Последняя" :value="bond.last" :min="priceMin" :max="priceMax" variant="primary" />
          <PriceBar label="Спрос (Bid)" :value="bond.bid" :min="priceMin" :max="priceMax" variant="info" />
          <PriceBar label="Предл. (Ask)" :value="bond.offer" :min="priceMin" :max="priceMax" variant="warning" />
        </div>
      </Panel>

      <Panel title="Глубина стакана" icon="layers">
        <div v-if="bond.bid != null && bond.offer != null" class="td-depth">
          <div class="td-depth__head">
            <span class="td-depth__bid-lbl">↑ Bid</span>
            <span class="td-depth__ask-lbl">Ask ↓</span>
          </div>
          <div class="td-depth__bar">
            <div class="td-depth__bar-bid" :style="{ width: bidRatio + '%' }">{{ bidRatio }}%</div>
            <div class="td-depth__bar-ask" :style="{ width: (100 - bidRatio) + '%' }">{{ 100 - bidRatio }}%</div>
          </div>
          <div class="td-depth__legend">
            <div class="td-depth__leg">
              <div class="td-depth__leg-lbl">Покупка (Bid)</div>
              <div class="td-depth__leg-val td-depth__leg-val--bid">{{ fmt.percent(bond.bid) }}</div>
            </div>
            <div class="td-depth__leg">
              <div class="td-depth__leg-lbl">Продажа (Ask)</div>
              <div class="td-depth__leg-val td-depth__leg-val--ask">{{ fmt.percent(bond.offer) }}</div>
            </div>
          </div>
        </div>
        <div v-else class="td-depth-empty">
          Нет данных по стакану
        </div>
      </Panel>
    </div>

    <!-- 2 Panel: Trading data + Volumes -->
    <div class="td-row td-row--2col">
      <Panel title="Торговые данные за день" icon="bar-chart">
        <InfoRow label="Открытие" :value="fmtPct(bond.open)" mono />
        <InfoRow label="Минимум" :value="fmtPct(bond.low)" mono />
        <InfoRow label="Максимум" :value="fmtPct(bond.high)" mono />
        <InfoRow label="Закрытие пред." :value="fmtPct(bond.prevprice)" mono />
        <InfoRow label="Изменение за день" :value="changeStr" mono :tone="changeTone" />
        <InfoRow label="WAP (средневзв.)" :value="fmtPct(bond.waprice)" mono />
        <InfoRow label="Изм. от WAP" :value="wapChangeStr" mono />
      </Panel>

      <Panel title="Объёмы и статистика" icon="sort-down">
        <InfoRow v-if="bond.mid_price_pct != null" label="Mid-price" :value="midPriceText" mono tone="primary" />
        <InfoRow label="Кол-во сделок" :value="fmt.num(bond.numtrades)" mono />
        <InfoRow label="Объём, шт." :value="fmt.num(bond.vol_today)" mono />
        <InfoRow label="Объём, ₽" :value="fmt.volume(bond.valtoday)" mono />
        <InfoRow label="Средний размер сделки" :value="avgTradeSize" mono />
        <InfoRow label="Общая глубина" :value="totalDepth" mono />
        <InfoRow label="Bid/Ask ratio" :value="bidAskRatio" mono />
        <InfoRow label="Заявки на покупку" :value="fmt.num(bond.numbids)" mono />
        <InfoRow label="Заявки на продажу" :value="fmt.num(bond.numoffers)" mono />
      </Panel>
    </div>

    <!-- Timestamps -->
    <Panel
      v-if="bond.updatetime || bond.tradetime || bond.systime"
      title="Временные метки"
      icon="clock"
      flush
      class="mt-3"
    >
      <div class="td-times">
        <div v-if="bond.updatetime" class="td-time">
          <div class="td-time__lbl">Время обновления</div>
          <div class="td-time__val">{{ fmt.time(bond.updatetime) }}</div>
        </div>
        <div v-if="bond.tradetime" class="td-time">
          <div class="td-time__lbl">Время сделки</div>
          <div class="td-time__val">{{ fmt.time(bond.tradetime) }}</div>
        </div>
        <div v-if="bond.systime" class="td-time">
          <div class="td-time__lbl">Системное время</div>
          <div class="td-time__val">{{ fmt.dateTime(bond.systime) }}</div>
        </div>
      </div>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{ bond: Bond }>()
const fmt = useFormat()

function fmtPct(v: number | null | undefined): string {
  return v != null ? v.toFixed(2) + '%' : '—'
}

const spread = computed(() => {
  if (props.bond.bid != null && props.bond.offer != null) return props.bond.offer - props.bond.bid
  return null
})
const spreadStr = computed(() => spread.value != null ? spread.value.toFixed(2) + ' п.п.' : '—')
const spreadPct = computed(() => {
  if (spread.value != null && props.bond.bid != null && props.bond.bid > 0) return (spread.value / props.bond.bid) * 100
  return null
})

const changeTone = computed<'success' | 'danger' | undefined>(() => {
  if (props.bond.last_change == null) return undefined
  if (props.bond.last_change > 0) return 'success'
  if (props.bond.last_change < 0) return 'danger'
  return undefined
})

const changeStr = computed(() => {
  if (props.bond.last_change == null) return '—'
  const pct = props.bond.last_change_prcnt
  const sign = props.bond.last_change >= 0 ? '+' : ''
  let s = sign + props.bond.last_change.toFixed(2) + ' п.п.'
  if (pct != null) s += ` (${sign}${pct.toFixed(2)}%)`
  return s
})

const wapChangeStr = computed(() => {
  if (props.bond.last == null || props.bond.waprice == null) return '—'
  const diff = props.bond.last - props.bond.waprice
  const sign = diff >= 0 ? '+' : ''
  return sign + diff.toFixed(3) + ' п.п.'
})

const avgTradeSize = computed(() => {
  if (!props.bond.numtrades || props.bond.numtrades <= 0 || !props.bond.vol_today) return '—'
  return Math.round(props.bond.vol_today / props.bond.numtrades) + ' шт.'
})

const totalDepth = computed(() => {
  const bid = props.bond.biddeptht ?? 0
  const offer = props.bond.offerdeptht ?? 0
  return bid + offer > 0 ? fmt.num(bid + offer) : '—'
})

const bidAskRatio = computed(() => {
  if (!props.bond.biddeptht || !props.bond.offerdeptht || props.bond.offerdeptht === 0) return '—'
  return (props.bond.biddeptht / props.bond.offerdeptht).toFixed(2)
})

const midPriceText = computed(() => {
  const pct = props.bond.mid_price_pct
  const rub = props.bond.mid_price_rub
  if (pct == null) return '—'
  let s = pct.toFixed(2) + '%'
  if (rub != null) s += ` · ${fmt.priceRub(rub)}`
  return s
})

const allPrices = computed(() =>
  [props.bond.bid, props.bond.last, props.bond.offer, props.bond.open, props.bond.low, props.bond.high]
    .filter((v): v is number => v != null)
)

const priceMin = computed(() => allPrices.value.length ? Math.min(...allPrices.value) - 0.3 : 95)
const priceMax = computed(() => allPrices.value.length ? Math.max(...allPrices.value) + 0.3 : 105)

const bidRatio = computed(() => {
  if (!props.bond.biddeptht && !props.bond.offerdeptht) return 50
  const total = (props.bond.biddeptht ?? 0) + (props.bond.offerdeptht ?? 0)
  return total > 0 ? Math.round(((props.bond.biddeptht ?? 0) / total) * 100) : 50
})
</script>

<style scoped>
.td-kpis {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.td-row { margin-bottom: 16px; }
.td-row--2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 992px) {
  .td-kpis { grid-template-columns: repeat(2, 1fr); }
  .td-row--2col { grid-template-columns: 1fr; }
}

/* Price bars container */
.td-pricebars {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 18px;
}

/* Depth visual */
.td-depth {
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.td-depth__head {
  display: flex;
  justify-content: space-between;
  font: 600 11.5px/1 var(--nla-font);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.td-depth__bid-lbl { color: var(--nla-success); }
.td-depth__ask-lbl { color: var(--nla-danger); }

.td-depth__bar {
  display: flex;
  height: 28px;
  border-radius: var(--nla-radius-sm);
  overflow: hidden;
  background: var(--nla-border-light);
}
.td-depth__bar-bid,
.td-depth__bar-ask {
  display: flex;
  align-items: center;
  justify-content: center;
  font: 700 11px/1 var(--nla-font-mono);
  color: #fff;
}
.td-depth__bar-bid { background: var(--nla-success); }
.td-depth__bar-ask { background: var(--nla-danger); }

.td-depth__legend {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
}
.td-depth__leg { text-align: center; }
.td-depth__leg-lbl {
  font: 500 11px/1 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.td-depth__leg-val {
  font: 700 18px/1.2 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  margin-top: 4px;
}
.td-depth__leg-val--bid { color: var(--nla-success); }
.td-depth__leg-val--ask { color: var(--nla-danger); }

.td-depth-empty {
  padding: 40px 18px;
  text-align: center;
  color: var(--nla-text-muted);
  font: 500 12.5px/1.4 var(--nla-font);
}

/* Times row */
.td-times {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}
.td-time {
  padding: 14px 18px;
  border-right: 1px solid var(--nla-border-light);
}
.td-time:last-child { border-right: 0; }
.td-time__lbl {
  font: 600 10.5px/1 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.td-time__val {
  font: 600 13px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  margin-top: 6px;
}

@media (max-width: 768px) {
  .td-times { grid-template-columns: 1fr; }
  .td-time { border-right: 0; border-bottom: 1px solid var(--nla-border-light); }
  .td-time:last-child { border-bottom: 0; }
}
</style>
