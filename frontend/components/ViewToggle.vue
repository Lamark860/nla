<!--
  ViewToggle.vue — segmented controls для переключения вью.
  Использует NuxtLink, активный пункт по route.path.
-->
<template>
  <div class="view-toggle">
    <NuxtLink
      v-for="opt in options"
      :key="opt.path"
      :to="opt.path"
      class="view-toggle__seg"
      :class="{ 'is-active': isActive(opt.path) }"
    >
      <i v-if="opt.icon" :class="`bi bi-${opt.icon}`" aria-hidden="true"></i>
      <span>{{ opt.label }}</span>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
interface Option {
  path: string
  label: string
  icon?: string
}
defineProps<{ options: Option[] }>()

const route = useRoute()
function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<style scoped>
.view-toggle {
  display: inline-flex;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border);
  border-radius: 9px;
  padding: 2px;
}
.view-toggle__seg {
  height: 26px;
  padding: 0 12px;
  border-radius: 7px;
  font: 500 12px/1 var(--nla-font);
  color: var(--nla-text-muted);
  background: transparent;
  border: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  text-decoration: none;
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
}
.view-toggle__seg:hover { color: var(--nla-text); }
.view-toggle__seg.is-active {
  background: var(--nla-bg-card);
  color: var(--nla-text);
  box-shadow: var(--nla-shadow-sm);
}
.view-toggle__seg i { font-size: 13px; }
</style>
