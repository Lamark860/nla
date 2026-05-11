<!--
  BondHistoryTab.vue — график цены и доходности.
  Тулбар периода + Chart.js canvas + сводка диапазонов.
-->
<template>
  <Panel flush>
    <template #head>
      <div class="hist__head">
        <div class="hist__title">
          <i class="bi bi-graph-up" aria-hidden="true"></i>
          <span>{{ activeMode === 'price' ? 'Цена' : 'Доходность' }}</span>
        </div>
        <div class="hist__modes">
          <button
            v-for="m in ['price', 'yield'] as const"
            :key="m"
            class="hist-mode"
            :class="{ active: m === activeMode }"
            @click="activeMode = m"
          >
            {{ m === 'price' ? 'Цена' : 'Доходность' }}
          </button>
        </div>
        <div class="hist__periods">
          <button
            v-for="p in periods"
            :key="p.value"
            class="hist-period"
            :class="{ active: p.value === activePeriod }"
            @click="activePeriod = p.value; $emit('change', { mode: activeMode, period: activePeriod })"
          >
            {{ p.label }}
          </button>
        </div>
      </div>
    </template>

    <div class="hist__chart">
      <slot name="chart">
        <!-- Чарт прокидывается из родителя (Chart.js) — здесь placeholder -->
        <div class="hist__placeholder">
          <i class="bi bi-graph-up"></i>
          <span>Chart.js canvas — см. родителя</span>
        </div>
      </slot>
    </div>

    <div class="hist__ranges">
      <RangeRow label="Текущая" :value="current" tone="primary" mono />
      <RangeRow label="Минимум за период" :value="min" mono />
      <RangeRow label="Максимум за период" :value="max" mono />
      <RangeRow label="Среднее" :value="avg" tone="muted" mono />
    </div>
  </Panel>
</template>

<script setup lang="ts">
import Panel from './Panel.vue'
import { defineComponent, h } from 'vue'

defineProps<{
  current?: string
  min?: string
  max?: string
  avg?: string
}>()
defineEmits<{ change: [val: { mode: 'price' | 'yield'; period: string }] }>()

const activeMode = ref<'price' | 'yield'>('price')
const activePeriod = ref('30')

const periods = [
  { value: '1', label: '1Д' },
  { value: '7', label: '1Н' },
  { value: '30', label: '1М' },
  { value: '90', label: '3М' },
  { value: '180', label: '6М' },
  { value: '365', label: '1Г' },
  { value: 'all', label: 'Всё' },
]

// inline component: RangeRow — пара label / mono value
export const RangeRow = defineComponent({
  name: 'BH_RangeRow',
  props: { label: String, value: String, tone: String, mono: Boolean },
  setup(props) {
    return () => h('div', { class: 'rrow' }, [
      h('span', { class: 'rrow__lbl' }, props.label),
      h('span', { class: ['rrow__val', props.mono && 'is-mono', props.tone && `is-${props.tone}`] }, props.value || '—')
    ])
  }
})
</script>

<style scoped>
.hist__head { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; width: 100%; }
.hist__title {
  display: flex; align-items: center; gap: 8px;
  font: 700 13px / 1.4 var(--nla-font);
  color: var(--nla-text);
}
.hist__title i { color: var(--nla-primary); font-size: 14px; }

.hist__modes, .hist__periods {
  display: flex;
  gap: 2px;
  padding: 2px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-sm);
}
.hist__periods { margin-left: auto; }

.hist-mode, .hist-period {
  appearance: none; border: 0; background: transparent;
  padding: 5px 10px;
  font: 500 11.5px / 1 var(--nla-font);
  color: var(--nla-text-secondary);
  border-radius: 4px;
  cursor: pointer;
  font-feature-settings: 'tnum';
}
.hist-mode:hover, .hist-period:hover { color: var(--nla-text); }
.hist-mode.active, .hist-period.active {
  background: var(--nla-bg-card);
  color: var(--nla-text);
  font-weight: 600;
  box-shadow: var(--nla-shadow-sm);
}

.hist__chart {
  height: 360px;
  padding: 16px;
}
.hist__placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--nla-text-subtle);
  border: 1px dashed var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 13px / 1 var(--nla-font);
}
.hist__placeholder i { font-size: 34px; }

.hist__ranges {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0;
  border-top: 1px solid var(--nla-border);
}
:deep(.rrow) {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-left: 1px solid var(--nla-border-light);
}
:deep(.rrow:first-child) { border-left: 0; }
:deep(.rrow__lbl) {
  font: 600 10.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
:deep(.rrow__val) {
  font: 600 16px / 1.2 var(--nla-font);
  color: var(--nla-text);
}
:deep(.rrow__val.is-mono)    { font-family: var(--nla-font-mono); font-feature-settings: 'tnum'; }
:deep(.rrow__val.is-primary) { color: var(--nla-primary); }
:deep(.rrow__val.is-muted)   { color: var(--nla-text-muted); }

@media (max-width: 768px) {
  .hist__ranges { grid-template-columns: repeat(2, 1fr); }
  :deep(.rrow:nth-child(3)) { border-left: 0; }
  :deep(.rrow:nth-child(n+3)) { border-top: 1px solid var(--nla-border-light); }
}
</style>
