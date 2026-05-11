<!--
  RatingBadge.vue — кредитный рейтинг (AAA, AA-, BB+, ...).
  Совместим со старым RatingBadge: использует ratingChipStyle/ratingTierStyle
  из useFormat() / useRating(). Фоллбэк на softRatingChipStyle, если есть.
  Шрифт — mono, fw 700; рядом — agency маленьким текстом.
-->
<template>
  <span v-if="rating != null && rating !== 'NULL'" class="rating-badge" :style="badgeStyle">
    {{ rating }}<span v-if="agency" class="rating-badge__ag">{{ agency }}</span>
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{ rating?: string | null; agency?: string }>()
const fmt = useFormat()

const badgeStyle = computed(() => {
  // используем существующий хелпер — единый источник правды для tier-цветов
  return props.rating ? fmt.ratingChipStyle(props.rating) : undefined
})
</script>

<style scoped>
.rating-badge {
  display: inline-flex;
  align-items: baseline;
  gap: 5px;
  padding: 3px 8px;
  border-radius: var(--nla-radius-sm);
  border: 1px solid currentColor;
  font: 700 11px / 1.4 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  letter-spacing: 0;
  white-space: nowrap;
  /* цвета через :style из ratingChipStyle (background + color); border берёт currentColor */
}
.rating-badge__ag {
  font: 600 9.5px / 1 var(--nla-font);
  letter-spacing: 0.05em;
  text-transform: uppercase;
  opacity: 0.7;
}
</style>
