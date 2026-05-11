<!--
  AiScore.vue — AI-балл 0..100, с цветной заливкой по порогам из useFormat().
  Иконка sparkles + значение. Когда compact — без подписи AI.
-->
<template>
  <span class="ai-score" :class="{ 'ai-score--compact': compact }" :style="badgeStyle">
    <i class="bi bi-stars ai-score__icon" aria-hidden="true"></i>
    <span class="ai-score__lbl" v-if="!compact">AI</span>
    <span class="ai-score__val">{{ Math.round(value) }}</span>
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  value: number
  compact?: boolean
}>()
const fmt = useFormat()
const badgeStyle = computed(() => fmt.aiRatingStyleSoft(props.value))
</script>

<style scoped>
.ai-score {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px 3px 7px;
  border-radius: var(--nla-radius-pill);
  font: 600 11.5px / 1.4 var(--nla-font);
  white-space: nowrap;
}
.ai-score__icon { font-size: 12px; opacity: 0.85; }
.ai-score__lbl  { font-size: 9.5px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; opacity: 0.7; }
.ai-score__val  {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0;
}
.ai-score--compact { padding: 2px 7px; }
</style>
