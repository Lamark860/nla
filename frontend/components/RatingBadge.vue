<template>
  <span class="badge font-mono" :style="badgeStyle">
    <template v-if="rating != null">{{ rating }}<span style="opacity: 0.6">/100</span></template>
    <template v-else>—</template>
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  rating: number | null | undefined
  size?: 'sm' | 'md'
}>()

const fmt = useFormat()
const badgeStyle = computed(() => fmt.aiRatingStyle(props.rating))
</script>
