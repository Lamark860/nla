<!--
  Tag.vue — небольшая прямоугольная метка (категория, период купона, признак "Флоатер").
  Не нести в одной странице больше 3 разных тонов.
-->
<template>
  <span class="tag" :class="`tag--${tone}`">
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  /** default — нейтральный, primary — фиолет, accent — заливка primary, success/warning/danger */
  tone?: 'default' | 'primary' | 'accent' | 'success' | 'warning' | 'danger'
}>(), { tone: 'default' })
</script>

<style scoped>
.tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font: 500 11px / 1.4 var(--nla-font);
  border-radius: var(--nla-radius-sm);
  border: 1px solid var(--nla-border);
  background: var(--nla-bg-card);
  color: var(--nla-text-secondary);
  white-space: nowrap;
  letter-spacing: 0;
}
.tag--primary {
  border-color: color-mix(in oklab, var(--nla-primary) 25%, transparent);
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  font-weight: 500;
}
.tag--accent {
  border-color: var(--nla-primary);
  background: var(--nla-primary);
  color: var(--nla-text-inverse);
  font-weight: 600;
}
.tag--success {
  border-color: color-mix(in oklab, var(--nla-success) 25%, transparent);
  background: var(--nla-success-light);
  color: var(--nla-success);
}
.tag--warning {
  border-color: color-mix(in oklab, var(--nla-warning) 25%, transparent);
  background: var(--nla-warning-light);
  color: var(--nla-warning);
}
.tag--danger {
  border-color: color-mix(in oklab, var(--nla-danger) 25%, transparent);
  background: var(--nla-danger-light);
  color: var(--nla-danger);
}
</style>
