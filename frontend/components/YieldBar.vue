<template>
  <div>
    <div class="d-flex justify-content-between mb-1">
      <span class="small text-muted fw-medium">{{ label }}</span>
      <span class="small fw-semibold font-monospace">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span>
    </div>
    <div class="yield-bar">
      <div :class="fillClass" class="yield-bar__fill" :style="{ width: pct + '%' }"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  label?: string
  value?: number | null
  max?: number
}>(), {
  label: '',
  value: null,
  max: 20,
})

const pct = computed(() =>
  props.value == null || props.max <= 0
    ? 0
    : Math.max(0, Math.min(100, (props.value / props.max) * 100))
)

const fillClass = computed(() => {
  if (props.value == null) return 'yield-bar__fill--primary'
  if (props.value >= 20)   return 'yield-bar__fill--danger'
  if (props.value >= 15)   return 'yield-bar__fill--warning'
  if (props.value >= 10)   return 'yield-bar__fill--success'
  return 'yield-bar__fill--primary'
})
</script>
