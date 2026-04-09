<template>
  <div class="space-y-6 animate-fade-in">
    <div v-if="!history || history.length === 0" class="card py-12 text-center text-slate-400 dark:text-slate-500">
      Нет данных об истории торгов
    </div>

    <template v-else>
      <!-- Stats -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div class="stat-card">
          <div class="stat-label">Цена закрытия</div>
          <div class="stat-value">{{ fmt.percent(lastClose) }}</div>
          <div v-if="bond.facevalue && lastClose" class="stat-sub font-mono">{{ fmt.priceRub(lastClose / 100 * bond.facevalue) }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Изменение за период</div>
          <div class="stat-value" :class="priceChangeColor">{{ priceChange }}</div>
          <div class="stat-sub font-mono">{{ priceChangePct }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Волатильность</div>
          <div class="stat-value">{{ volatility }}</div>
          <div class="stat-sub">σ дневных изменений</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Средний объём</div>
          <div class="stat-value">{{ fmt.num(avgVolume) }}</div>
          <div class="stat-sub">шт./день</div>
        </div>
      </div>

      <!-- Chart -->
      <div class="card p-6">
        <h3 class="section-title mb-5">История цены и объёмов</h3>
        <ClientOnly>
          <div class="h-72 sm:h-96 rounded-xl overflow-hidden bg-slate-50 dark:bg-surface-800 p-3">
            <canvas ref="chartCanvas" />
          </div>
        </ClientOnly>
      </div>

      <!-- Price range + trading days -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Price range -->
        <div class="card p-6">
          <h3 class="section-title mb-5">Ценовой диапазон</h3>
          <div class="space-y-4">
            <RangeRow label="Минимум" :value="stats.low" :min="stats.low!" :max="stats.high!" />
            <RangeRow label="Средняя" :value="stats.avg" :min="stats.low!" :max="stats.high!" />
            <RangeRow v-if="lastClose" label="Текущая" :value="lastClose" :min="stats.low!" :max="stats.high!" />
            <RangeRow label="Максимум" :value="stats.high" :min="stats.low!" :max="stats.high!" />
            <RangeRow v-if="bond.facevalue" label="Номинал (%)" :value="100" :min="stats.low!" :max="stats.high!" />
          </div>
        </div>

        <!-- Trading days balance -->
        <div class="card p-6">
          <h3 class="section-title mb-5">Баланс торговых дней</h3>
          <div class="space-y-5">
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs text-slate-500 dark:text-slate-400">Рост</span>
                <span class="text-sm font-semibold text-emerald-600 dark:text-emerald-400 font-mono">{{ upDays }} дн. ({{ upPct }}%)</span>
              </div>
              <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
                <div class="h-full bg-emerald-500 dark:bg-emerald-400 rounded-lg" :style="{ width: upPct + '%' }"></div>
              </div>
            </div>
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs text-slate-500 dark:text-slate-400">Падение</span>
                <span class="text-sm font-semibold text-red-600 dark:text-red-400 font-mono">{{ downDays }} дн. ({{ downPct }}%)</span>
              </div>
              <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
                <div class="h-full bg-red-500 dark:bg-red-400 rounded-lg" :style="{ width: downPct + '%' }"></div>
              </div>
            </div>
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs text-slate-500 dark:text-slate-400">Без изменений</span>
                <span class="text-sm font-semibold text-slate-500 font-mono">{{ flatDays }} дн. ({{ flatPct }}%)</span>
              </div>
              <div class="h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden">
                <div class="h-full bg-slate-400 dark:bg-slate-500 rounded-lg" :style="{ width: flatPct + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Statistics tables -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="card p-6">
          <h3 class="section-title mb-4">Статистика за период</h3>
          <div class="space-y-3">
            <InfoRow label="Общий объём, шт." :value="fmt.num(totalVolume)" />
            <InfoRow label="Общая сумма, ₽" :value="fmt.volume(totalValue)" />
            <InfoRow label="Средний объём/день" :value="fmt.num(avgVolume)" />
            <InfoRow label="Средняя цена" :value="fmt.percent(stats.avg)" />
            <InfoRow label="Мин. цена" :value="fmt.percent(stats.low)" />
            <InfoRow label="Макс. цена" :value="fmt.percent(stats.high)" />
            <InfoRow label="Размах" :value="stats.high && stats.low ? (stats.high - stats.low).toFixed(2) + ' п.п.' : '—'" />
          </div>
        </div>
        <div class="card p-6">
          <h3 class="section-title mb-4">Изменение цены</h3>
          <div class="space-y-3">
            <InfoRow label="Начальная цена" :value="fmt.percent(firstClose)" />
            <InfoRow label="Конечная цена" :value="fmt.percent(lastClose)" />
            <InfoRow label="Общее изменение" :value="priceChange" />
            <InfoRow label="Изменение, %" :value="priceChangePct" />
            <InfoRow label="Волатильность (σ)" :value="volatility" />
            <InfoRow label="Номинал" :value="bond.facevalue ? fmt.priceRub(bond.facevalue) : '—'" />
            <InfoRow label="Отклонение от номинала" :value="lastClose ? (lastClose - 100).toFixed(2) + ' п.п.' : '—'" />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Chart, registerables } from 'chart.js'
import type { OHLC, Bond } from '~/composables/useApi'

Chart.register(...registerables)

const props = defineProps<{
  history: OHLC[]
  bond: Bond
}>()

const fmt = useFormat()
const chartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

const stats = computed(() => {
  const closes = props.history.map(h => h.close).filter(Boolean)
  return {
    high: closes.length ? Math.max(...closes) : null,
    low: closes.length ? Math.min(...closes) : null,
    avg: closes.length ? closes.reduce((a, b) => a + b, 0) / closes.length : null,
  }
})

const lastClose = computed(() => {
  const h = props.history
  return h.length ? h[h.length - 1].close : null
})

const firstClose = computed(() => {
  return props.history.length ? props.history[0].close : null
})

const priceChange = computed(() => {
  if (!firstClose.value || !lastClose.value) return '—'
  const diff = lastClose.value - firstClose.value
  const sign = diff >= 0 ? '+' : ''
  return sign + diff.toFixed(2) + ' п.п.'
})

const priceChangePct = computed(() => {
  if (!firstClose.value || !lastClose.value || firstClose.value === 0) return '—'
  const pct = ((lastClose.value - firstClose.value) / firstClose.value) * 100
  const sign = pct >= 0 ? '+' : ''
  return sign + pct.toFixed(2) + '%'
})

const priceChangeColor = computed(() => {
  if (!firstClose.value || !lastClose.value) return ''
  return lastClose.value >= firstClose.value
    ? 'text-emerald-600 dark:text-emerald-400'
    : 'text-red-600 dark:text-red-400'
})

const volatility = computed(() => {
  if (props.history.length < 2) return '—'
  const changes: number[] = []
  for (let i = 1; i < props.history.length; i++) {
    if (props.history[i].close && props.history[i - 1].close) {
      changes.push(props.history[i].close - props.history[i - 1].close)
    }
  }
  if (changes.length === 0) return '—'
  const mean = changes.reduce((a, b) => a + b, 0) / changes.length
  const variance = changes.reduce((a, c) => a + Math.pow(c - mean, 2), 0) / changes.length
  return Math.sqrt(variance).toFixed(3)
})

const totalVolume = computed(() => props.history.reduce((a, h) => a + (h.volume || 0), 0))
const totalValue = computed(() => props.history.reduce((a, h) => a + (h.value || 0), 0))
const avgVolume = computed(() => props.history.length ? Math.round(totalVolume.value / props.history.length) : 0)

// Trading days balance
const dayChanges = computed(() => {
  let up = 0, down = 0, flat = 0
  for (let i = 1; i < props.history.length; i++) {
    const diff = props.history[i].close - props.history[i - 1].close
    if (diff > 0) up++
    else if (diff < 0) down++
    else flat++
  }
  return { up, down, flat }
})

const upDays = computed(() => dayChanges.value.up)
const downDays = computed(() => dayChanges.value.down)
const flatDays = computed(() => dayChanges.value.flat)
const totalDays = computed(() => upDays.value + downDays.value + flatDays.value || 1)
const upPct = computed(() => Math.round((upDays.value / totalDays.value) * 100))
const downPct = computed(() => Math.round((downDays.value / totalDays.value) * 100))
const flatPct = computed(() => 100 - upPct.value - downPct.value)

function renderChart() {
  if (!chartCanvas.value || !props.history.length) return
  if (chartInstance) chartInstance.destroy()

  const isDark = document.documentElement.classList.contains('dark')
  const gridColor = isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(148, 163, 184, 0.15)'
  const textColor = isDark ? '#64748b' : '#94a3b8'

  // Volume bar colors: green for up, red for down
  const volColors = props.history.map((h, i) => {
    if (i === 0) return isDark ? 'rgba(96,165,250,0.4)' : 'rgba(59,130,246,0.3)'
    return h.close >= props.history[i - 1].close
      ? (isDark ? 'rgba(52,211,153,0.5)' : 'rgba(16,185,129,0.4)')
      : (isDark ? 'rgba(248,113,113,0.5)' : 'rgba(239,68,68,0.4)')
  })

  const maxVol = Math.max(...props.history.map(h => h.volume || 0))

  chartInstance = new Chart(chartCanvas.value, {
    type: 'bar',
    data: {
      labels: props.history.map(h => h.date),
      datasets: [
        {
          type: 'line',
          label: 'Цена (%)',
          data: props.history.map(h => h.close),
          borderColor: isDark ? '#60a5fa' : '#3b82f6',
          backgroundColor: isDark ? 'rgba(96, 165, 250, 0.08)' : 'rgba(59, 130, 246, 0.08)',
          fill: true,
          tension: 0.35,
          pointRadius: 0,
          pointHitRadius: 10,
          borderWidth: 2,
          yAxisID: 'yPrice',
          order: 1,
        },
        {
          type: 'line',
          label: 'Номинал',
          data: props.history.map(() => 100),
          borderColor: isDark ? 'rgba(148,163,184,0.3)' : 'rgba(148,163,184,0.4)',
          borderDash: [6, 4],
          borderWidth: 1,
          pointRadius: 0,
          fill: false,
          yAxisID: 'yPrice',
          order: 2,
        },
        {
          type: 'bar',
          label: 'Объём (шт.)',
          data: props.history.map(h => h.volume || 0),
          backgroundColor: volColors,
          yAxisID: 'yVolume',
          order: 3,
          barPercentage: 0.8,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { intersect: false, mode: 'index' },
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: isDark ? '#1a2332' : '#ffffff',
          titleColor: isDark ? '#94a3b8' : '#64748b',
          bodyColor: isDark ? '#e2e8f0' : '#0f172a',
          titleFont: { family: 'JetBrains Mono, monospace', size: 11 },
          bodyFont: { family: 'JetBrains Mono, monospace', size: 12, weight: '600' },
          borderColor: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.06)',
          borderWidth: 1,
          padding: 12,
          cornerRadius: 10,
          callbacks: {
            label: (ctx) => {
              if (ctx.datasetIndex === 0) return ` Цена: ${ctx.parsed.y.toFixed(2)}%`
              if (ctx.datasetIndex === 1) return ''
              return ` Объём: ${ctx.parsed.y.toLocaleString('ru-RU')} шт.`
            },
          },
        },
      },
      scales: {
        x: {
          grid: { color: gridColor },
          ticks: { color: textColor, maxTicksLimit: 8, font: { size: 10, family: 'JetBrains Mono, monospace' } },
        },
        yPrice: {
          position: 'left',
          grid: { color: gridColor },
          ticks: { color: textColor, font: { size: 10, family: 'JetBrains Mono, monospace' }, callback: (v) => v + '%' },
        },
        yVolume: {
          position: 'right',
          grid: { display: false },
          max: maxVol * 3.5,
          ticks: { color: textColor, font: { size: 10, family: 'JetBrains Mono, monospace' }, callback: (v) => v > 0 ? (Number(v) / 1000).toFixed(0) + 'K' : '' },
        },
      },
    },
  })
}

onMounted(() => { nextTick(renderChart) })
watch(() => props.history, () => { nextTick(renderChart) })

// RangeRow helper
const RangeRow = defineComponent({
  props: {
    label: String,
    value: { type: Number, default: null },
    min: { type: Number, default: 90 },
    max: { type: Number, default: 110 },
  },
  setup(props) {
    const pct = computed(() => {
      if (props.value == null || props.max === props.min) return 0
      return Math.max(0, Math.min(100, ((props.value - props.min) / (props.max - props.min)) * 100))
    })
    return { pct, fmt: useFormat() }
  },
  template: `
    <div class="flex items-center gap-3">
      <span class="text-xs text-slate-400 dark:text-slate-500 w-28 shrink-0">{{ label }}</span>
      <div class="flex-1 h-2.5 bg-slate-100 dark:bg-surface-800 rounded-lg overflow-hidden relative">
        <div class="absolute h-full w-1 bg-primary-500 rounded" :style="{ left: pct + '%' }"></div>
      </div>
      <span class="text-sm font-semibold text-slate-900 dark:text-white w-16 text-right font-mono">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span>
    </div>
  `,
})
</script>
