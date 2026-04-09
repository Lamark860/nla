<template>
  <div class="flex flex-col gap-0.5">
    <span class="stat-label">{{ label }}</span>
    <span :class="['text-base font-semibold tabular-nums font-mono', highlight ? 'text-primary-600 dark:text-primary-400' : 'text-slate-900 dark:text-white']">
      {{ value }}
    </span>
    <span v-if="sub" class="text-sm text-slate-400 dark:text-slate-500">{{ sub }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: string
  sub?: string
  highlight?: boolean
}>()
</script>
