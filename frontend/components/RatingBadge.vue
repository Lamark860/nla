<template>
  <span :class="['badge font-mono', colorClass]">
    <template v-if="rating != null">{{ rating }}<span class="opacity-60">/100</span></template>
    <template v-else>—</template>
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  rating: number | null | undefined
}>()

const fmt = useFormat()
const colorClass = computed(() => fmt.ratingColor(props.rating))
</script>
