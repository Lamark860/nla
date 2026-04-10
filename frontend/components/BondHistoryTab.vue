<template>
  <div class="animate-fade-in">
    <div v-if="!history || history.length === 0" class="card py-5 text-center text-muted">
      Нет данных об истории торгов
    </div>

    <template v-else>
      <!-- Stats -->
      <div class="row g-3 mb-4">
        <div class="col-6 col-sm-3">
          <div class="stat-card">
            <div class="stat-label">Цена закрытия</div>
            <div class="stat-value">{{ fmt.percent(lastClose) }}</div>
            <div v-if="bond.facevalue && lastClose" class="stat-sub font-monospace">{{ fmt.priceRub(lastClose / 100 * bond.facevalue) }}</div>
          </div>
        </div>
        <div class="col-6 col-sm-3">
          <div class="stat-card">
            <div class="stat-label">Изменение за период</div>
            <div class="stat-value" :class="priceChangeColor">{{ priceChange }}</div>
            <div class="stat-sub font-monospace">{{ priceChangePct }}</div>
          </div>
        </div>
        <div class="col-6 col-sm-3">
          <div class="stat-card">
            <div class="stat-label">Волатильность</div>
            <div class="stat-value">{{ volatility }}</div>
            <div class="stat-sub">σ дневных изменений</div>
          </div>
        </div>
        <div class="col-6 col-sm-3">
          <div class="stat-card">
            <div class="stat-label">Средний объём</div>
            <div class="stat-value">{{ fmt.num(avgVolume) }}</div>
            <div class="stat-sub">шт./день</div>
          </div>
        </div>
      </div>

      <!-- Chart -->
      <div class="card p-4 mb-4">
        <h3 class="section-title mb-4">История цены и объёмов</h3>
        <ClientOnly>
          <div class="rounded overflow-hidden p-2" style="height: 24rem; background: var(--nla-bg)">
            <canvas ref="chartCanvas" />
          </div>
        </ClientOnly>
      </div>

      <!-- Price range + trading days -->
      <div class="row g-4 mb-4">
        <div class="col-lg-6">
          <div class="card p-4">
            <h3 class="section-title mb-4">Ценовой диапазон</h3>
            <div class="d-flex flex-column gap-3">
              <RangeRow label="Минимум" :value="stats.low" :min="stats.low!" :max="stats.high!" />
              <RangeRow label="Средняя" :value="stats.avg" :min="stats.low!" :max="stats.high!" />
              <RangeRow v-if="lastClose" label="Текущая" :value="lastClose" :min="stats.low!" :max="stats.high!" />
              <RangeRow label="Максимум" :value="stats.high" :min="stats.low!" :max="stats.high!" />
              <RangeRow v-if="bond.facevalue" label="Номинал (%)" :value="100" :min="stats.low!" :max="stats.high!" />
            </div>
          </div>
        </div>

        <div class="col-lg-6">
          <div class="card p-4">
            <h3 class="section-title mb-4">Баланс торговых дней</h3>
            <div class="d-flex flex-column gap-3">
              <div>
                <div class="d-flex justify-content-between mb-1">
                  <span class="small text-muted">Рост</span>
                  <span class="small fw-semibold text-success font-monospace">{{ upDays }} дн. ({{ upPct }}%)</span>
                </div>
                <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--success" :style="{ width: upPct + '%' }"></div></div>
              </div>
              <div>
                <div class="d-flex justify-content-between mb-1">
                  <span class="small text-muted">Падение</span>
                  <span class="small fw-semibold text-danger font-monospace">{{ downDays }} дн. ({{ downPct }}%)</span>
                </div>
                <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--danger" :style="{ width: downPct + '%' }"></div></div>
              </div>
              <div>
                <div class="d-flex justify-content-between mb-1">
                  <span class="small text-muted">Без изменений</span>
                  <span class="small fw-semibold text-muted font-monospace">{{ flatDays }} дн. ({{ flatPct }}%)</span>
                </div>
                <div class="yield-bar"><div class="yield-bar__fill" style="background: var(--nla-text-muted)" :style="{ width: flatPct + '%' }"></div></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Statistics tables -->
      <div class="row g-4">
        <div class="col-lg-6">
          <div class="card overflow-hidden">
            <div class="panel-header">
              <i class="bi bi-calculator"></i>
              Статистика за период
            </div>
            <div>
              <InfoRow label="Общий объём, шт." :value="fmt.num(totalVolume)" />
              <InfoRow label="Общая сумма, ₽" :value="fmt.volume(totalValue)" />
              <InfoRow label="Средний объём/день" :value="fmt.num(avgVolume)" />
              <InfoRow label="Средняя цена" :value="fmt.percent(stats.avg)" />
              <InfoRow label="Мин. цена" :value="fmt.percent(stats.low)" />
              <InfoRow label="Макс. цена" :value="fmt.percent(stats.high)" />
              <InfoRow label="Размах" :value="stats.high && stats.low ? (stats.high - stats.low).toFixed(2) + ' п.п.' : '—'" />
            </div>
          </div>
        </div>
        <div class="col-lg-6">
          <div class="card overflow-hidden">
            <div class="panel-header">
              <i class="bi bi-graph-up-arrow"></i>
              Изменение цены
            </div>
            <div>
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

const stats = computed(() => {
  const closes = props.history.map(h => h.close).filter(Boolean)
  return {
    high: closes.length ? Math.max(...closes) : null,
    low: closes.length ? Math.min(...closes) : null,
    avg: closes.length ? closes.reduce((a, b) => a + b, 0) / closes.length : null,
  }
})

const lastClose = computed(() => props.history.length ? props.history[props.history.length - 1].close : null)
const firstClose = computed(() => props.history.length ? props.history[0].close : null)

const priceChange = computed(() => {
  if (!firstClose.value || !lastClose.value) return '—'
  const diff = lastClose.value - firstClose.value
  return (diff >= 0 ? '+' : '') + diff.toFixed(2) + ' п.п.'
})

const priceChangePct = computed(() => {
  if (!firstClose.value || !lastClose.value || firstClose.value === 0) return '—'
  const pct = ((lastClose.value - firstClose.value) / firstClose.value) * 100
  return (pct >= 0 ? '+' : '') + pct.toFixed(2) + '%'
})

const priceChangeColor = computed(() => {
  if (!firstClose.value || !lastClose.value) return ''
  return lastClose.value >= firstClose.value ? 'text-success' : 'text-danger'
})

const volatility = computed(() => {
  if (props.history.length < 2) return '—'
  const changes: number[] = []
  for (let i = 1; i < props.history.length; i++) {
    if (props.history[i].close && props.history[i - 1].close) changes.push(props.history[i].close - props.history[i - 1].close)
  }
  if (changes.length === 0) return '—'
  const mean = changes.reduce((a, b) => a + b, 0) / changes.length
  const variance = changes.reduce((a, c) => a + Math.pow(c - mean, 2), 0) / changes.length
  return Math.sqrt(variance).toFixed(3)
})

const totalVolume = computed(() => props.history.reduce((a, h) => a + (h.volume || 0), 0))
const totalValue = computed(() => props.history.reduce((a, h) => a + (h.value || 0), 0))
const avgVolume = computed(() => props.history.length ? Math.round(totalVolume.value / props.history.length) : 0)

const dayChanges = computed(() => {
  let up = 0, down = 0, flat = 0
  for (let i = 1; i < props.history.length; i++) {
    const diff = props.history[i].close - props.history[i - 1].close
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

function renderChart() {
  if (!chartCanvas.value || !props.history.length) return
  if (chartInstance) chartInstance.destroy()

  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  const gridColor = isDark ? 'rgba(255,255,255,0.04)' : 'rgba(148,163,184,0.15)'
  const textColor = isDark ? '#64748b' : '#94a3b8'
  const volColors = props.history.map((h, i) => {
    if (i === 0) return isDark ? 'rgba(96,165,250,0.4)' : 'rgba(59,130,246,0.3)'
    return h.close >= props.history[i - 1].close
      ? (isDark ? 'rgba(52,211,153,0.5)' : 'rgba(16,185,129,0.4)')
      : (isDark ? 'rgba(248,113,113,0.5)' : 'rgba(239,68,68,0.4)')
  })

  chartInstance = new Chart(chartCanvas.value, {
    type: 'bar',
    data: {
      labels: props.history.map(h => h.date),
      datasets: [
        { type: 'line', label: 'Цена (%)', data: props.history.map(h => h.close), borderColor: isDark ? '#60a5fa' : '#3b82f6', backgroundColor: isDark ? 'rgba(96,165,250,0.08)' : 'rgba(59,130,246,0.08)', fill: true, tension: 0.35, pointRadius: 0, pointHitRadius: 10, borderWidth: 2, yAxisID: 'yPrice', order: 1 },
        { type: 'line', label: 'Номинал', data: props.history.map(() => 100), borderColor: isDark ? 'rgba(148,163,184,0.3)' : 'rgba(148,163,184,0.4)', borderDash: [6, 4], borderWidth: 1, pointRadius: 0, fill: false, yAxisID: 'yPrice', order: 2 },
        { type: 'bar', label: 'Объём (шт.)', data: props.history.map(h => h.volume || 0), backgroundColor: volColors, yAxisID: 'yVolume', order: 3, barPercentage: 0.8 },
      ],
    },
    options: {
      responsive: true, maintainAspectRatio: false, interaction: { intersect: false, mode: 'index' },
      plugins: { legend: { display: false }, tooltip: { backgroundColor: isDark ? '#1a2332' : '#fff', titleColor: textColor, bodyColor: isDark ? '#e2e8f0' : '#0f172a', titleFont: { family: 'JetBrains Mono, monospace', size: 11 }, bodyFont: { family: 'JetBrains Mono, monospace', size: 12, weight: '600' }, borderColor: isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.06)', borderWidth: 1, padding: 12, cornerRadius: 10, callbacks: { label: (ctx) => { if (ctx.datasetIndex === 0) return ` Цена: ${ctx.parsed.y.toFixed(2)}%`; if (ctx.datasetIndex === 1) return ''; return ` Объём: ${ctx.parsed.y.toLocaleString('ru-RU')} шт.` } } } },
      scales: {
        x: { ticks: { color: textColor, font: { family: 'JetBrains Mono', size: 10 }, maxRotation: 0, autoSkip: true, maxTicksLimit: 12 }, grid: { color: gridColor } },
        yPrice: { position: 'left', ticks: { color: textColor, font: { family: 'JetBrains Mono', size: 11 }, callback: (v: any) => v + '%' }, grid: { color: gridColor } },
        yVolume: { position: 'right', ticks: { color: textColor, font: { size: 10 } }, grid: { display: false }, max: Math.max(...props.history.map(h => h.volume || 0)) * 3 },
      },
    },
  })
}

onMounted(() => nextTick(renderChart))
watch(() => props.history, () => nextTick(renderChart), { deep: true })
onUnmounted(() => { if (chartInstance) chartInstance.destroy() })

// RangeRow inline component
const RangeRow = defineComponent({
  props: { label: String, value: { type: Number, default: null }, min: { type: Number, default: 95 }, max: { type: Number, default: 105 } },
  setup(p) {
    const pct = computed(() => p.value == null || p.max <= p.min ? 0 : Math.max(0, Math.min(100, ((p.value - p.min) / (p.max - p.min)) * 100)))
    const f = useFormat()
    return { pct, f }
  },
  template: `<div>
    <div class="d-flex justify-content-between mb-1"><span class="small text-muted">{{ label }}</span><span class="small fw-semibold font-monospace">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span></div>
    <div class="yield-bar"><div class="yield-bar__fill yield-bar__fill--primary" :style="{ width: pct + '%' }"></div></div>
  </div>`,
})
</script>
