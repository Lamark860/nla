<!--
  Panel.vue — стандартный контейнер с шапкой.
  Использует токены --nla-bg-card, --nla-border, --nla-shadow-sm.
  Слот #head — кастомная шапка вместо стандартной title+icon.
  Слот по умолчанию — содержимое (paddings задаёшь сам — иногда таблице нужно без отступа).
-->
<template>
  <section class="panel" :class="{ 'panel--flush': flush }">
    <header v-if="$slots.head || title" class="panel-head">
      <slot name="head">
        <i v-if="icon" :class="`bi bi-${icon}`" class="panel-head__icon" aria-hidden="true"></i>
        <span class="panel-head__title">{{ title }}</span>
        <span v-if="meta" class="panel-head__meta">{{ meta }}</span>
        <span v-if="$slots.headRight" class="panel-head__right"><slot name="headRight" /></span>
      </slot>
    </header>
    <div :class="bodyClass">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  icon?: string
  meta?: string
  /** убрать padding в body (для таблиц / списков inset) */
  flush?: boolean
}>()

const bodyClass = computed(() => '')
</script>

<style scoped>
.panel {
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-lg);
  box-shadow: var(--nla-shadow-sm);
  overflow: hidden;
}
.panel-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--nla-border);
  background: var(--nla-bg-elevated);
  font: 600 13px / 1.4 var(--nla-font);
  color: var(--nla-text);
  letter-spacing: -0.005em;
}
.panel-head__icon {
  color: var(--nla-text-muted);
  font-size: 14px;
  flex-shrink: 0;
}
.panel-head__title {
  flex: 1 1 auto;
  min-width: 0;
}
.panel-head__meta {
  font: 500 11px / 1 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.panel-head__right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}
.panel--flush > div {
  padding: 0;
}
</style>
