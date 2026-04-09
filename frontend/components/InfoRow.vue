<template>
  <div class="info-row">
    <span class="info-row__label">{{ label }}</span>
    <span class="info-row__value">{{ value }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: string
}>()
</script>
