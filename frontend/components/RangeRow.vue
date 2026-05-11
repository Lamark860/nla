<template>
  <div>
    <div class="d-flex justify-content-between mb-1">
      <span class="small text-muted">{{ label }}</span>
      <span class="small fw-semibold font-monospace">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span>
    </div>
    <div class="yield-bar">
      <div class="yield-bar__fill" :class="`yield-bar__fill--${tone}`" :style="{ width: pct + '%' }"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  label?: string
  value?: number | null
  min?: number
  max?: number
  tone?: 'primary' | 'success' | 'danger' | 'muted'
}>(), {
  label: '',
  value: null,
  min: 95,
  max: 105,
  tone: 'primary',
})

const pct = computed(() =>
  props.value == null || props.max <= props.min
    ? 0
    : Math.max(0, Math.min(100, ((props.value - props.min) / (props.max - props.min)) * 100))
)
</script>
