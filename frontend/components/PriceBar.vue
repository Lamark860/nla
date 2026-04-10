<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-1">
      <span class="small text-muted">{{ label }}</span>
      <span class="small fw-semibold font-monospace">{{ value != null ? value.toFixed(2) + '%' : '—' }}</span>
    </div>
    <div class="progress" style="height: 6px">
      <div
        class="progress-bar"
        :class="`bg-${variant}`"
        role="progressbar"
        :style="{ width: pct + '%' }"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string
  value: number | null | undefined
  min: number
  max: number
  variant?: string
}>()

const pct = computed(() => {
  if (props.value == null || props.max <= props.min) return 0
  return Math.max(0, Math.min(100, ((props.value - props.min) / (props.max - props.min)) * 100))
})
</script>
