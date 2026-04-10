<template>
  <span v-if="value != null" class="badge font-monospace fw-bold" :style="{ backgroundColor: color, color: '#fff' }">
    {{ value.toFixed(1) }}/10
  </span>
  <span v-else class="text-muted">—</span>
</template>

<script setup lang="ts">
const props = defineProps<{ value: number | null | undefined }>()

const color = computed(() => {
  if (props.value == null) return '#6c757d'
  if (props.value >= 8) return '#198754'
  if (props.value >= 6) return '#0dcaf0'
  if (props.value >= 4) return '#0d6efd'
  if (props.value >= 2) return '#fd7e14'
  return '#dc3545'
})
</script>
