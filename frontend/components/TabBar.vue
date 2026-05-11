<!--
  TabBar.vue — сегментный переключатель табов.
  Заменяет nav-tabs из Bootstrap. Поддерживает:
   - счётчик внутри таба (count)
   - sticky-режим (`sticky` prop) для длинных страниц
   - v-model для текущего таба
-->
<template>
  <nav class="tabs" :class="{ 'tabs--sticky': sticky }" role="tablist">
    <button
      v-for="t in tabs"
      :key="t.key"
      class="tab"
      :class="{ active: t.key === modelValue }"
      role="tab"
      :aria-selected="t.key === modelValue"
      @click="$emit('update:modelValue', t.key)"
    >
      <i v-if="t.icon" :class="`bi bi-${t.icon}`" class="tab__icon" aria-hidden="true"></i>
      <span class="tab__label">{{ t.label }}</span>
      <span v-if="t.count != null" class="count">{{ t.count }}</span>
    </button>
  </nav>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string
  tabs: Array<{ key: string; label: string; icon?: string; count?: number | string }>
  sticky?: boolean
}>()
defineEmits<{ 'update:modelValue': [key: string] }>()
</script>

<style scoped>
.tabs {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  overflow-x: auto;
  scrollbar-width: none;
}
.tabs::-webkit-scrollbar { display: none; }
.tabs--sticky {
  position: sticky;
  top: 64px;
  z-index: 10;
  backdrop-filter: blur(8px);
  background: color-mix(in oklab, var(--nla-bg-subtle) 90%, transparent);
}

.tab {
  appearance: none;
  border: 0;
  background: transparent;
  padding: 8px 14px;
  font: 500 13px / 1 var(--nla-font);
  color: var(--nla-text-secondary);
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
  white-space: nowrap;
}
.tab:hover { color: var(--nla-text); background: var(--nla-bg-hover); }
.tab.active {
  background: var(--nla-bg-card);
  color: var(--nla-text);
  font-weight: 600;
  box-shadow: var(--nla-shadow-sm);
}
.tab.active .tab__icon { color: var(--nla-primary); }

.tab__icon {
  font-size: 14px;
  color: var(--nla-text-muted);
}
.count {
  display: inline-flex;
  align-items: center;
  min-width: 18px;
  padding: 0 5px;
  height: 16px;
  border-radius: 8px;
  background: var(--nla-bg-hover);
  font: 600 10.5px / 1 var(--nla-font);
  color: var(--nla-text-muted);
  font-feature-settings: 'tnum';
}
.tab.active .count {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
}

.tab:focus-visible {
  outline: 2px solid var(--nla-primary);
  outline-offset: 2px;
}
</style>
