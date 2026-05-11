<!--
  PageHead.vue — шапка страницы по handoff index.html.
  Title + sub-line + actions area (slot).
-->
<template>
  <header class="page-head">
    <div class="page-head__main">
      <h1 class="page-head__title">{{ title }}</h1>
      <p v-if="sub || $slots.sub" class="page-head__sub">
        <slot name="sub">{{ sub }}</slot>
      </p>
    </div>

    <div v-if="$slots.actions" class="page-head__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  sub?: string
}>()
</script>

<style scoped>
.page-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 22px;
  flex-wrap: wrap;
}
.page-head__main {
  min-width: 0;
}
.page-head__title {
  font: 700 24px/1.2 var(--nla-font);
  letter-spacing: -0.02em;
  margin: 0 0 4px;
  color: var(--nla-text);
}
.page-head__sub {
  margin: 0;
  font: 500 13px/1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.page-head__sub :deep(b),
.page-head__sub :deep(strong) {
  color: var(--nla-text-secondary);
  font-weight: 500;
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.page-head__actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .page-head { gap: 12px; }
  .page-head__title { font-size: 20px; }
}
</style>
