<!--
  Pill.vue — пилл-статус с цветной точкой.
  Используется в ленте сделок (BUY/SELL), в купонной таблице (выплачен/ожидается),
  в шапке торгов (T — торгуется).
-->
<template>
  <span class="pill" :class="`pill--${tone}`">
    <span v-if="dot" class="pill__dot"></span>
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  tone?: 'default' | 'success' | 'danger' | 'warning' | 'primary' | 'live'
  /** показывать цветную точку слева — true по умолчанию */
  dot?: boolean
}>(), { tone: 'default', dot: true })
</script>

<style scoped>
.pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px 3px 8px;
  font: 600 11px / 1.5 var(--nla-font);
  border-radius: var(--nla-radius-pill);
  background: var(--nla-bg-subtle);
  color: var(--nla-text-secondary);
  letter-spacing: 0.01em;
  white-space: nowrap;
}
.pill__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  flex-shrink: 0;
}
.pill--success { background: var(--nla-success-light); color: var(--nla-success); }
.pill--danger  { background: var(--nla-danger-light);  color: var(--nla-danger); }
.pill--warning { background: var(--nla-warning-light); color: var(--nla-warning); }
.pill--primary { background: var(--nla-primary-light); color: var(--nla-primary-ink); }

/* live — пульсирующая зелёная точка для статуса торгов */
.pill--live { background: var(--nla-success-light); color: var(--nla-success); }
.pill--live .pill__dot { animation: pillPulse 1.6s ease-out infinite; }

@keyframes pillPulse {
  0%, 100% { box-shadow: 0 0 0 0 currentColor; opacity: 1; }
  60%      { box-shadow: 0 0 0 4px transparent; opacity: 0.6; }
}
</style>
