<template>
  <div class="animate-fade-in">
    <div v-if="!history || history.length === 0" class="card py-5 text-center text-muted">
      Нет данных об истории торгов
    </div>

    <template v-else>
      <!-- Block 1 — KPI row -->
      <div class="hist-kpis">
        <KPI label="Цена закрытия" :value="fmt.percent(lastClose)">
          <template v-if="lastCloseRub != null" #sub>
            <span class="font-monospace">{{ fmt.priceRub(lastCloseRub) }}</span>
          </template>
        </KPI>
        <KPI label="Изменение за период" :tone="changeTone" :sub-tone="changeTone">
          <template #default>{{ priceChange }}</template>
          <template #sub>{{ priceChangePct }}</template>
        </KPI>
        <KPI label="Волатильность" :value="volatility" sub="σ дневных изменений" />
        <KPI label="Средний объём" :value="fmt.num(avgVolume)" sub="шт./день" />
      </div>

      <!-- Block 2 — Chart panel + period toggle + ranges -->
      <Panel flush>
        <template #head>
          <div class="hist-head">
            <div class="hist-title">
              <i class="bi bi-graph-up" aria-hidden="true"></i>
              <span>История цены</span>
            </div>
            <div class="hist-volscale" role="tablist" aria-label="Шкала объёмов">
              <button
                v-for="v in volScales"
                :key="v.value"
                class="hist-period"
                :class="{ active: v.value === volScale }"
                role="tab"
                :aria-selected="v.value === volScale"
                :title="`Шкала объёмов: ${v.title}`"
                @click="volScale = v.value"
              >
                <i class="bi bi-bar-chart-fill" aria-hidden="true"></i>
                <span>{{ v.label }}</span>
              </button>
            </div>
            <div class="hist-periods" role="tablist">
              <button
                v-for="p in periods"
                :key="p.value"
                class="hist-period"
                :class="{ active: p.value === effectivePeriod }"
                role="tab"
                :aria-selected="p.value === effectivePeriod"
                @click="activePeriod = p.value"
              >
                {{ p.label }}
              </button>
            </div>
          </div>
        </template>

        <ClientOnly>
          <div class="hist-chart">
            <canvas ref="chartCanvas" />
          </div>
        </ClientOnly>

        <div class="hist-ranges">
          <RangeRow label="Текущая" :value="lastClose" :min="rangeMin" :max="rangeMax" tone="primary" />
          <RangeRow label="Минимум" :value="stats.low" :min="rangeMin" :max="rangeMax" tone="muted" />
          <RangeRow label="Максимум" :value="stats.high" :min="rangeMin" :max="rangeMax" tone="muted" />
          <RangeRow label="Среднее" :value="stats.avg" :min="rangeMin" :max="rangeMax" tone="muted" />
        </div>
      </Panel>

      <!-- Block 3 — Trading day balance -->
      <Panel title="Баланс торговых дней" icon="bar-chart-steps" class="mt-3">
        <div class="hist-balance">
          <div class="bal-row">
            <div class="d-flex justify-content-between mb-1">
              <span class="small text-muted">Рост</span>
              <span class="small fw-semibold text-success font-monospace">{{ upDays }} дн. ({{ upPct }}%)</span>
            </div>
            <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--success" :style="{ width: upPct + '%' }"></div></div>
          </div>
          <div class="bal-row">
            <div class="d-flex justify-content-between mb-1">
              <span class="small text-muted">Падение</span>
              <span class="small fw-semibold text-danger font-monospace">{{ downDays }} дн. ({{ downPct }}%)</span>
            </div>
            <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--danger" :style="{ width: downPct + '%' }"></div></div>
          </div>
          <div class="bal-row">
            <div class="d-flex justify-content-between mb-1">
              <span class="small text-muted">Без изменений</span>
              <span class="small fw-semibold text-muted font-monospace">{{ flatDays }} дн. ({{ flatPct }}%)</span>
            </div>
            <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--muted" :style="{ width: flatPct + '%' }"></div></div>
          </div>
        </div>
      </Panel>

      <!-- Block 4 — Two stats panels -->
      <div class="hist-tables">
        <Panel title="Статистика за период" icon="calculator">
          <InfoRow label="Общий объём, шт." :value="fmt.num(totalVolume)" mono />
          <InfoRow label="Общая сумма, ₽" :value="fmt.volume(totalValue)" mono />
          <InfoRow label="Средний объём/день" :value="fmt.num(avgVolume)" mono />
          <InfoRow label="Средняя цена" :value="fmt.percent(stats.avg)" mono />
          <InfoRow label="Мин. цена" :value="fmt.percent(stats.low)" mono />
          <InfoRow label="Макс. цена" :value="fmt.percent(stats.high)" mono />
          <InfoRow label="Размах" :value="rangeText" mono />
        </Panel>
        <Panel title="Изменение цены" icon="graph-up-arrow">
          <InfoRow label="Начальная цена" :value="fmt.percent(firstClose)" mono />
          <InfoRow label="Конечная цена" :value="fmt.percent(lastClose)" mono />
          <InfoRow label="Общее изменение" :value="priceChange" mono :tone="changeTone === 'default' ? undefined : changeTone" />
          <InfoRow label="Изменение, %" :value="priceChangePct" mono :tone="changeTone === 'default' ? undefined : changeTone" />
          <InfoRow label="Волатильность (σ)" :value="volatility" mono />
          <InfoRow label="Номинал" :value="bond.facevalue ? fmt.priceRub(bond.facevalue) : '—'" mono />
          <InfoRow label="Отклонение от номинала" :value="deviationText" mono />
        </Panel>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Chart, registerables } from 'chart.js'
import type { OHLC, Bond } from '~/composables/useApi'

Chart.register(...registerables)

const props = defineProps<{ history: OHLC[]; bond: Bond }>()
const fmt = useFormat()
const chartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

// — Period toggle —
type PeriodValue = '7' | '30' | '90' | '180' | '365' | 'all'
const periods: { value: PeriodValue; label: string }[] = [
  { value: '7',   label: '1Н' },
  { value: '30',  label: '1М' },
  { value: '90',  label: '3М' },
  { value: '180', label: '6М' },
  { value: '365', label: '1Г' },
  { value: 'all', label: 'Всё' },
]
const activePeriod = ref<PeriodValue>('30')

// Шкала объёмов: компакт = bars прижаты к низу (max × 5), норма = средне (× 3), широкий = bars высокие (× 1.3)
type VolScale = 'compact' | 'normal' | 'wide'
const volScales: { value: VolScale; label: string; title: string }[] = [
  { value: 'compact', label: '↓',   title: 'компакт' },
  { value: 'normal',  label: '·',   title: 'норма' },
  { value: 'wide',    label: '↑',   title: 'широкий' },
]
const volScale = ref<VolScale>('normal')
const volScaleMultiplier: Record<VolScale, number> = {
  compact: 5,
  normal: 3,
  wide: 1.3,
}

const filteredHistory = computed(() => {
  if (activePeriod.value === 'all') return props.history
  const days = Number(activePeriod.value)
  const cutoff = Date.now() - days * 86400_000
  const slice = props.history.filter(h => new Date(h.date).getTime() >= cutoff)
  // fallback: если период длиннее данных, показываем всю историю
  return slice.length === 0 ? props.history : slice
})

const fellbackToAll = computed(() =>
  activePeriod.value !== 'all'
    && filteredHistory.value === props.history
    && props.history.length > 0
)

const effectivePeriod = computed<PeriodValue>(() =>
  fellbackToAll.value ? 'all' : activePeriod.value
)

// — Stats over filtered window —
const stats = computed(() => {
  const closes = filteredHistory.value.map(h => h.close).filter((v): v is number => v != null && Number.isFinite(v))
  return {
    high: closes.length ? Math.max(...closes) : null,
    low: closes.length ? Math.min(...closes) : null,
    avg: closes.length ? closes.reduce((a, b) => a + b, 0) / closes.length : null,
  }
})

const lastClose = computed(() => {
  const arr = filteredHistory.value
  return arr.length ? arr[arr.length - 1].close : null
})
const firstClose = computed(() => {
  const arr = filteredHistory.value
  return arr.length ? arr[0].close : null
})
const lastCloseRub = computed(() => {
  if (lastClose.value == null || !props.bond.facevalue) return null
  return (lastClose.value / 100) * props.bond.facevalue
})

const priceChange = computed(() => {
  if (firstClose.value == null || lastClose.value == null) return '—'
  const diff = lastClose.value - firstClose.value
  return (diff >= 0 ? '+' : '') + diff.toFixed(2) + ' п.п.'
})

const priceChangePct = computed(() => {
  if (firstClose.value == null || lastClose.value == null || firstClose.value === 0) return '—'
  const pct = ((lastClose.value - firstClose.value) / firstClose.value) * 100
  return (pct >= 0 ? '+' : '') + pct.toFixed(2) + '%'
})

const changeTone = computed<'success' | 'danger' | 'default'>(() => {
  if (firstClose.value == null || lastClose.value == null) return 'default'
  if (lastClose.value > firstClose.value) return 'success'
  if (lastClose.value < firstClose.value) return 'danger'
  return 'default'
})

const volatility = computed(() => {
  const arr = filteredHistory.value
  if (arr.length < 2) return '—'
  const changes: number[] = []
  for (let i = 1; i < arr.length; i++) {
    if (arr[i].close != null && arr[i - 1].close != null) changes.push(arr[i].close - arr[i - 1].close)
  }
  if (changes.length === 0) return '—'
  const mean = changes.reduce((a, b) => a + b, 0) / changes.length
  const variance = changes.reduce((a, c) => a + Math.pow(c - mean, 2), 0) / changes.length
  return Math.sqrt(variance).toFixed(3)
})

const totalVolume = computed(() => filteredHistory.value.reduce((a, h) => a + (h.volume || 0), 0))
const totalValue = computed(() => filteredHistory.value.reduce((a, h) => a + (h.value || 0), 0))
const avgVolume = computed(() => filteredHistory.value.length ? Math.round(totalVolume.value / filteredHistory.value.length) : 0)

const dayChanges = computed(() => {
  const arr = filteredHistory.value
  let up = 0, down = 0, flat = 0
  for (let i = 1; i < arr.length; i++) {
    const diff = arr[i].close - arr[i - 1].close
    if (diff > 0) up++; else if (diff < 0) down++; else flat++
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

// Range guard: если low===high (одна точка), искусственно растягиваем шкалу,
// чтобы прогресс-бар не схлопнулся в 0.
const rangeMin = computed(() => {
  if (stats.value.low == null) return 0
  if (stats.value.high != null && stats.value.high !== stats.value.low) return stats.value.low
  return stats.value.low - 1
})
const rangeMax = computed(() => {
  if (stats.value.high == null) return 100
  if (stats.value.low != null && stats.value.high !== stats.value.low) return stats.value.high
  return stats.value.high + 1
})

const rangeText = computed(() =>
  stats.value.high != null && stats.value.low != null
    ? (stats.value.high - stats.value.low).toFixed(2) + ' п.п.'
    : '—'
)
const deviationText = computed(() =>
  lastClose.value != null ? (lastClose.value - 100).toFixed(2) + ' п.п.' : '—'
)

function renderChart() {
  if (!chartCanvas.value) return
  if (chartInstance) { chartInstance.destroy(); chartInstance = null }
  const arr = filteredHistory.value
  if (!arr.length) return

  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  const gridColor = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(148,163,184,0.15)'
  const textColor = isDark ? '#64748b' : '#94a3b8'
  const volColors = arr.map((h, i) => {
    if (i === 0) return isDark ? 'rgba(96,165,250,0.4)' : 'rgba(59,130,246,0.3)'
    return h.close >= arr[i - 1].close
      ? (isDark ? 'rgba(52,211,153,0.5)' : 'rgba(16,185,129,0.4)')
      : (isDark ? 'rgba(248,113,113,0.5)' : 'rgba(239,68,68,0.4)')
  })

  chartInstance = new Chart(chartCanvas.value, {
    type: 'bar',
    data: {
      labels: arr.map(h => h.date),
      datasets: [
        { type: 'line', label: 'Цена (%)', data: arr.map(h => h.close), borderColor: isDark ? '#60a5fa' : '#3b82f6', backgroundColor: isDark ? 'rgba(96,165,250,0.08)' : 'rgba(59,130,246,0.08)', fill: true, tension: 0.35, pointRadius: 0, pointHitRadius: 10, borderWidth: 2, yAxisID: 'yPrice', order: 1 },
        { type: 'line', label: 'Номинал', data: arr.map(() => 100), borderColor: isDark ? 'rgba(148,163,184,0.3)' : 'rgba(148,163,184,0.4)', borderDash: [6, 4], borderWidth: 1, pointRadius: 0, fill: false, yAxisID: 'yPrice', order: 2 },
        { type: 'bar', label: 'Объём (шт.)', data: arr.map(h => h.volume || 0), backgroundColor: volColors, yAxisID: 'yVolume', order: 3, barPercentage: 0.8 },
      ],
    },
    options: {
      responsive: true, maintainAspectRatio: false, interaction: { intersect: false, mode: 'index' },
      plugins: { legend: { display: false }, tooltip: { backgroundColor: isDark ? '#1a2332' : '#fff', titleColor: textColor, bodyColor: isDark ? '#e2e8f0' : '#0f172a', titleFont: { family: 'JetBrains Mono, monospace', size: 11 }, bodyFont: { family: 'JetBrains Mono, monospace', size: 12, weight: 'bold' }, borderColor: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.06)', borderWidth: 1, padding: 12, cornerRadius: 10, callbacks: { label: (ctx) => { const y = ctx.parsed.y ?? 0; if (ctx.datasetIndex === 0) return ` Цена: ${y.toFixed(2)}%`; if (ctx.datasetIndex === 1) return ''; return ` Объём: ${y.toLocaleString('ru-RU')} шт.` } } } },
      scales: {
        x: { ticks: { color: textColor, font: { family: 'JetBrains Mono', size: 10 }, maxRotation: 0, autoSkip: true, maxTicksLimit: 12 }, grid: { color: gridColor } },
        yPrice: { position: 'left', ticks: { color: textColor, font: { family: 'JetBrains Mono', size: 11 }, callback: (v: any) => v + '%' }, grid: { color: gridColor } },
        yVolume: { position: 'right', ticks: { color: textColor, font: { size: 10 } }, grid: { display: false }, max: Math.max(...arr.map(h => h.volume || 0)) * volScaleMultiplier[volScale.value] },
      },
    },
  })
}

onMounted(() => nextTick(renderChart))
watch(() => props.history, () => nextTick(renderChart), { deep: true })
watch(filteredHistory, () => nextTick(renderChart))
watch(volScale, () => nextTick(renderChart))
onUnmounted(() => { if (chartInstance) chartInstance.destroy() })
</script>

<style scoped>
.hist-kpis {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 18px;
}

.hist-head { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; width: 100%; }
.hist-title { display: flex; align-items: center; gap: 8px; font: 700 13px/1.4 var(--nla-font); color: var(--nla-text); }
.hist-title i { color: var(--nla-primary); font-size: 14px; }

.hist-periods,
.hist-volscale {
  display: flex;
  gap: 2px;
  padding: 2px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-sm);
}
.hist-periods { margin-left: auto; }
.hist-volscale .hist-period {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 8px;
}
.hist-volscale .bi { font-size: 10px; opacity: 0.7; }
.hist-period {
  appearance: none;
  border: 0;
  background: transparent;
  padding: 5px 10px;
  font: 500 11.5px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  border-radius: 4px;
  cursor: pointer;
  font-feature-settings: 'tnum';
}
.hist-period:hover { color: var(--nla-text); }
.hist-period.active {
  background: var(--nla-bg-card);
  color: var(--nla-text);
  font-weight: 600;
  box-shadow: var(--nla-shadow-sm);
}

.hist-chart {
  height: 360px;
  padding: 16px;
}

.hist-ranges {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border-top: 1px solid var(--nla-border);
}
.hist-ranges > * {
  padding: 14px 16px;
  border-left: 1px solid var(--nla-border-light);
}
.hist-ranges > *:first-child { border-left: 0; }

.hist-balance { display: flex; flex-direction: column; gap: 12px; }

.hist-tables {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 16px;
}

@media (max-width: 992px) {
  .hist-kpis { grid-template-columns: repeat(2, 1fr); }
  .hist-tables { grid-template-columns: 1fr; }
  .hist-ranges { grid-template-columns: repeat(2, 1fr); }
  .hist-ranges > *:nth-child(3) { border-left: 0; }
  .hist-ranges > *:nth-child(n+3) { border-top: 1px solid var(--nla-border-light); }
}

@media (max-width: 480px) {
  .hist-chart { height: 280px; padding: 12px; }
}
</style>
