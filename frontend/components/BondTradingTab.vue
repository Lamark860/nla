<template>
  <div class="animate-fade-in">
    <!-- Stat cards -->
    <div class="row g-3 mb-4">
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Статус</div>
          <div class="stat-value" :class="bond.trading_status === 'T' ? 'text-positive' : 'text-neutral'">
            {{ bond.trading_status === 'T' ? 'Торгуется' : 'Нет торгов' }}
          </div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Спрос (Bid)</div>
          <div class="stat-value">{{ fmt.percent(bond.bid) }}</div>
          <div v-if="bond.bid && bond.facevalue" class="stat-sub font-monospace">{{ fmt.priceRub(bond.bid / 100 * bond.facevalue) }} · глуб. {{ fmt.num(bond.biddeptht) }}</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Предложение (Ask)</div>
          <div class="stat-value">{{ fmt.percent(bond.offer) }}</div>
          <div v-if="bond.offer && bond.facevalue" class="stat-sub font-monospace">{{ fmt.priceRub(bond.offer / 100 * bond.facevalue) }} · глуб. {{ fmt.num(bond.offerdeptht) }}</div>
        </div>
      </div>
      <div class="col-6 col-lg-3">
        <div class="stat-card">
          <div class="stat-label">Спред</div>
          <div class="stat-value">{{ spread != null ? spread.toFixed(2) + ' п.п.' : '—' }}</div>
          <div v-if="spreadPct != null" class="stat-sub font-monospace">{{ spreadPct.toFixed(2) }}% от бида</div>
        </div>
      </div>
    </div>

    <!-- Price bars + Order book -->
    <div class="row g-4 mb-4">
      <!-- Prices of the day -->
      <div class="col-lg-6">
        <div class="card p-4">
          <h3 class="section-title mb-4">Цены торгового дня</h3>
          <div class="d-flex flex-column gap-3">
            <PriceBar label="Открытие" :value="bond.open" :min="priceMin" :max="priceMax" variant="secondary" />
            <PriceBar label="Минимум" :value="bond.low" :min="priceMin" :max="priceMax" variant="danger" />
            <PriceBar label="Максимум" :value="bond.high" :min="priceMin" :max="priceMax" variant="success" />
            <PriceBar label="Последняя" :value="bond.last" :min="priceMin" :max="priceMax" variant="primary" />
            <PriceBar label="Спрос (Bid)" :value="bond.bid" :min="priceMin" :max="priceMax" variant="info" />
            <PriceBar label="Предл. (Ask)" :value="bond.offer" :min="priceMin" :max="priceMax" variant="warning" />
          </div>
        </div>
      </div>

      <!-- Depth visualization -->
      <div class="col-lg-6">
        <div class="card p-4">
          <h3 class="section-title mb-4">Глубина стакана</h3>
          <div v-if="bond.bid != null && bond.offer != null">
            <div class="d-flex justify-content-between small fw-semibold mb-2">
              <span class="text-success">↑ Bid</span>
              <span class="text-danger">Ask ↓</span>
            </div>
            <div class="d-flex rounded overflow-hidden" style="height: 28px; background: var(--nla-border)">
              <div class="d-flex align-items-center justify-content-center text-white fw-bold" style="font-size: 10px; background: rgba(var(--bs-success-rgb), 0.7)" :style="{ width: bidRatio + '%' }">
                {{ bidRatio }}%
              </div>
              <div class="d-flex align-items-center justify-content-center text-white fw-bold" style="font-size: 10px; background: rgba(var(--bs-danger-rgb), 0.7)" :style="{ width: (100 - bidRatio) + '%' }">
                {{ 100 - bidRatio }}%
              </div>
            </div>
            <div class="d-flex justify-content-between mt-3">
              <div class="text-center">
                <div class="small text-muted">Покупка (Bid)</div>
                <div class="fs-5 fw-bold font-monospace text-positive">{{ fmt.percent(bond.bid) }}</div>
              </div>
              <div class="text-center">
                <div class="small text-muted">Продажа (Ask)</div>
                <div class="fs-5 fw-bold font-monospace text-negative">{{ fmt.percent(bond.offer) }}</div>
              </div>
            </div>
          </div>
          <div v-else class="text-center text-muted small py-5">
            Нет данных по стакану
          </div>
        </div>
      </div>
    </div>

    <!-- Trading data + Volume stats -->
    <div class="row g-4 mb-4">
      <div class="col-lg-6">
        <div class="card overflow-hidden">
          <div class="panel-header">
            <i class="bi bi-bar-chart"></i>
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
      </div>

      <div class="col-lg-6">
        <div class="card overflow-hidden">
          <div class="panel-header">
            <i class="bi bi-sort-down"></i>
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
    </div>

    <!-- Timestamps -->
    <div v-if="bond.updatetime || bond.tradetime || bond.systime" class="card overflow-hidden">
      <div class="panel-header">
        <i class="bi bi-clock"></i>
        Временные метки
      </div>
      <div class="row g-3 p-3">
        <div v-if="bond.updatetime" class="col-sm-4">
          <div class="stat-label">Время обновления</div>
          <div class="small font-monospace mt-1">{{ fmt.time(bond.updatetime) }}</div>
        </div>
        <div v-if="bond.tradetime" class="col-sm-4">
          <div class="stat-label">Время сделки</div>
          <div class="small font-monospace mt-1">{{ fmt.time(bond.tradetime) }}</div>
        </div>
        <div v-if="bond.systime" class="col-sm-4">
          <div class="stat-label">Системное время</div>
          <div class="small font-monospace mt-1">{{ fmt.dateTime(bond.systime) }}</div>
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
  if (props.bond.bid != null && props.bond.offer != null) return props.bond.offer - props.bond.bid
  return null
})

const spreadPct = computed(() => {
  if (spread.value != null && props.bond.bid != null && props.bond.bid > 0) return (spread.value / props.bond.bid) * 100
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

const priceMin = computed(() => allPrices.value.length ? Math.min(...allPrices.value) - 0.3 : 95)
const priceMax = computed(() => allPrices.value.length ? Math.max(...allPrices.value) + 0.3 : 105)

const bidRatio = computed(() => {
  if (!props.bond.biddeptht && !props.bond.offerdeptht) return 50
  const total = (props.bond.biddeptht ?? 0) + (props.bond.offerdeptht ?? 0)
  return total > 0 ? Math.round(((props.bond.biddeptht ?? 0) / total) * 100) : 50
})
</script>
