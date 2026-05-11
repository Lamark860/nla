<!--
  KPI.vue — метрика на hero-блоке: лейбл / значение / sub.
  Значение всегда mono + tabular-nums.
  tone: 'default' | 'success' | 'danger' | 'primary' — окрашивает значение.
-->
<template>
  <div class="kpi">
    <div class="kpi-lbl">{{ label }}</div>
    <div class="kpi-val" :class="`kpi-val--${tone}`">
      <slot>{{ value }}</slot>
      <span v-if="unit" class="kpi-val__unit">{{ unit }}</span>
    </div>
    <div v-if="sub || $slots.sub" class="kpi-sub" :class="subToneClass">
      <slot name="sub">{{ sub }}</slot>
    </div>
  </div>
</template>

<script setup lang="ts">
type Tone = 'default' | 'success' | 'danger' | 'primary' | 'muted'

const props = withDefaults(defineProps<{
  label: string
  value?: string | number | null
  unit?: string
  sub?: string
  tone?: Tone
  subTone?: Tone
}>(), { tone: 'default', subTone: 'muted' })

const subToneClass = computed(() => `kpi-sub--${props.subTone}`)
</script>

<style scoped>
.kpi {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.kpi-lbl {
  font: 600 10.5px / 1.2 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
.kpi-val {
  font: 600 24px / 1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  color: var(--nla-text);
  letter-spacing: -0.02em;
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.kpi-val__unit {
  font: 500 13px / 1 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0;
}
.kpi-val--success { color: var(--nla-success); }
.kpi-val--danger  { color: var(--nla-danger);  }
.kpi-val--primary { color: var(--nla-primary); }
.kpi-val--muted   { color: var(--nla-text-muted); }

.kpi-sub {
  font: 500 12px / 1.3 var(--nla-font);
  font-feature-settings: 'tnum';
}
.kpi-sub--default { color: var(--nla-text); }
.kpi-sub--muted   { color: var(--nla-text-muted); }
.kpi-sub--success { color: var(--nla-success); }
.kpi-sub--danger  { color: var(--nla-danger);  }
.kpi-sub--primary { color: var(--nla-primary); }
</style>
