<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Stat cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="stat-card">
        <div class="stat-label">Статус</div>
        <div class="stat-value" :class="bond.trading_status === 'T' ? 'text-positive' : 'text-neutral'">
          {{ bond.trading_status === 'T' ? 'Торгуется' : 'Нет торгов' }}
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Спрос (Bid)</div>
        <div class="stat-value">{{ fmt.percent(bond.bid) }}</div>
        <div v-if="bond.bid && bond.facevalue" class="stat-sub font-mono">{{ fmt.priceRub(bond.bid / 100 * bond.facevalue) }} · глуб. {{ fmt.num(bond.biddeptht) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Предложение (Ask)</div>
        <div class="stat-value">{{ fmt.percent(bond.offer) }}</div>
        <div v-if="bond.offer && bond.facevalue" class="stat-sub font-mono">{{ fmt.priceRub(bond.offer / 100 * bond.facevalue) }} · глуб. {{ fmt.num(bond.offerdeptht) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Спред</div>
        <div class="stat-value">
          {{ spread != null ? spread.toFixed(2) + ' п.п.' : '—' }}
        </div>
        <div v-if="spreadPct != null" class="stat-sub font-mono">{{ spreadPct.toFixed(2) }}% от бида</div>
      </div>
    </div>

    <!-- Price bars + Order book -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Prices of the day -->
      <div class="card p-6">
        <h3 class="section-title mb-5">Цены торгового дня</h3>
        <div class="space-y-4">
          <PriceBar label="Открытие" :value="bond.open" :min="priceMin" :max="priceMax" color="bg-slate-400 dark:bg-slate-500" />
          <PriceBar label="Минимум" :value="bond.low" :min="priceMin" :max="priceMax" color="bg-red-500 dark:bg-red-400" />
          <PriceBar label="Максимум" :value="bond.high" :min="priceMin" :max="priceMax" color="bg-emerald-500 dark:bg-emerald-400" />
          <PriceBar label="Последняя" :value="bond.last" :min="priceMin" :max="priceMax" color="bg-primary-500 dark:bg-primary-400" />
          <PriceBar label="Спрос (Bid)" :value="bond.bid" :min="priceMin" :max="priceMax" color="bg-blue-500 dark:bg-blue-400" />
          <PriceBar label="Предл. (Ask)" :value="bond.offer" :min="priceMin" :max="priceMax" color="bg-amber-500 dark:bg-amber-400" />
        </div>
      </div>

      <!-- Depth visualization -->
      <div class="card p-6">
        <h3 class="section-title mb-5">Глубина стакана</h3>
        <div v-if="bond.bid != null && bond.offer != null" class="space-y-4">
          <div class="flex items-center justify-between text-xs font-semibold">
            <span class="text-emerald-500">↑ Bid</span>
            <span class="text-red-500">Ask ↓</span>
          </div>
          <div class="h-7 rounded-lg overflow-hidden flex" style="background: var(--nla-border-light)">
            <div class="h-full bg-emerald-500/70 flex items-center justify-center text-[10px] font-bold text-white" :style="{ width: bidRatio + '%' }">
              {{ bidRatio }}%
            </div>
            <div class="h-full bg-red-500/70 flex items-center justify-center text-[10px] font-bold text-white" :style="{ width: (100 - bidRatio) + '%' }">
              {{ 100 - bidRatio }}%
            </div>
          </div>
          <div class="flex items-center justify-between">
            <div class="text-center">
              <div class="text-xs" style="color: var(--nla-text-muted)">Покупка (Bid)</div>
              <div class="text-lg font-bold font-mono text-positive">{{ fmt.percent(bond.bid) }}</div>
            </div>
            <div class="text-center">
              <div class="text-xs" style="color: var(--nla-text-muted)">Продажа (Ask)</div>
              <div class="text-lg font-bold font-mono text-negative">{{ fmt.percent(bond.offer) }}</div>
            </div>
          </div>
        </div>
        <div v-else class="text-center text-sm py-6" style="color: var(--nla-text-muted)">
          Нет данных по стакану
        </div>
      </div>
    </div>

    <!-- Trading data + Volume stats -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Trading day data -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75z"/></svg>
          Торговые данные за день
        </div>
        <div>
          <InfoRow label="Открытие" :value="fmtPct(bond.open)" />
          <InfoRow label="Минимум" :value="fmtPct(bond.low)" />
          <InfoRow label="Максимум" :value="fmtPct(bond.high)" />
          <InfoRow label="Закрытие пред." :value="fmtPct(bond.prevprice)" />
          <InfoRow label="Изменение за день" :value="changeStr" />
          <InfoRow label="WAP (средневзв.)" :value="fmtPct(bond.waprice)" />
          <InfoRow label="Изм. от WAP" :value="wapChangeStr" />
        </div>
      </div>

      <!-- Volume and stats -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h5.25m5.25-.75L17.25 9m0 0L21 12.75M17.25 9v12"/></svg>
          Объёмы и статистика
        </div>
        <div>
          <InfoRow label="Кол-во сделок" :value="fmt.num(bond.numtrades)" />
          <InfoRow label="Объём, шт." :value="fmt.num(bond.vol_today)" />
          <InfoRow label="Объём, ₽" :value="fmt.volume(bond.valtoday)" />
          <InfoRow label="Средний размер сделки" :value="avgTradeSize" />
          <InfoRow label="Общая глубина" :value="totalDepth" />
          <InfoRow label="Bid/Ask ratio" :value="bidAskRatio" />
          <InfoRow label="Заявки на покупку" :value="fmt.num(bond.numbids)" />
          <InfoRow label="Заявки на продажу" :value="fmt.num(bond.numoffers)" />
        </div>
      </div>
    </div>

    <!-- Timestamps -->
    <div v-if="bond.updatetime || bond.tradetime || bond.systime" class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
        Временные метки
      </div>
      <div class="grid grid-cols-2 sm:grid-cols-3 gap-4 p-4">
        <div v-if="bond.updatetime">
          <div class="stat-label">Время обновления</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text)">{{ bond.updatetime }}</div>
        </div>
        <div v-if="bond.tradetime">
          <div class="stat-label">Время сделки</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text)">{{ bond.tradetime }}</div>
        </div>
        <div v-if="bond.systime">
          <div class="stat-label">Системное время</div>
          <div class="text-sm font-mono mt-1" style="color: var(--nla-text)">{{ bond.systime }}</div>
        </div>
      </div>
    </div>
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
  if (props.bond.bid != null && props.bond.offer != null) {
    return props.bond.offer - props.bond.bid
  }
  return null
})

const spreadPct = computed(() => {
  if (spread.value != null && props.bond.bid != null && props.bond.bid > 0) {
    return (spread.value / props.bond.bid) * 100
  }
  return null
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

const allPrices = computed(() => {
  return [props.bond.bid, props.bond.last, props.bond.offer, props.bond.open, props.bond.low, props.bond.high]
    .filter((v): v is number => v != null)
})

const priceMin = computed(() => allPrices.value.length ? Math.min(...allPrices.value) - 0.5 : 90)
const priceMax = computed(() => allPrices.value.length ? Math.max(...allPrices.value) + 0.5 : 110)

const bidRatio = computed(() => {
  if (props.bond.bid == null || props.bond.offer == null) return 50
  const total = props.bond.bid + props.bond.offer
  return total > 0 ? Math.round((props.bond.bid / total) * 100) : 50
})

const PriceBar = defineComponent({
  props: {
    label: String,
    value: { type: Number, default: null },
    min: { type: Number, default: 90 },
    max: { type: Number, default: 110 },
    color: { type: String, default: 'bg-primary-500' },
  },
  setup(props) {
    const pct = computed(() => {
      if (props.value == null) return 0
      return Math.max(0, Math.min(100, ((props.value - props.min) / (props.max - props.min)) * 100))
    })
    return { pct, fmt: useFormat() }
  },
  template: `
    <div class="flex items-center gap-3">
      <span class="text-xs text-slate-400 dark:text-slate-500 w-28 shrink-0 font-medium">{{ label }}</span>
      <div class="flex-1 h-6 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden relative">
        <div :class="color" class="h-full rounded-lg transition-all duration-500 ease-out" :style="{ width: pct + '%' }"></div>
      </div>
      <span class="text-sm font-semibold text-slate-900 dark:text-white w-16 text-right tabular-nums font-mono">{{ value != null ? fmt.percent(value) : '—' }}</span>
    </div>
  `,
})
</script>
