<!--
  InfoRow.vue — пара "лейбл — значение" в панели.
  Совместим со старым InfoRow по props (label, value).
  Добавлены: mono (моно-шрифт значения), tone (окраска значения), :slot value.
-->
<template>
  <div class="info-row">
    <div class="info-lbl">{{ label }}</div>
    <div class="info-val" :class="[{ 'info-val--mono': mono }, tone && `info-val--${tone}`]">
      <slot>{{ value ?? '—' }}</slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value?: string | number | null
  /** дать значению моноширинный шрифт + tabular-nums */
  mono?: boolean
  tone?: 'success' | 'danger' | 'primary' | 'muted'
}>()
</script>

<style scoped>
.info-row {
  display: grid;
  grid-template-columns: minmax(140px, 40%) 1fr;
  gap: 16px;
  padding: 11px 18px;
  border-top: 1px solid var(--nla-border-light);
  align-items: baseline;
}
.info-row:first-child { border-top: 0; }

.info-lbl {
  font: 500 12.5px / 1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.info-val {
  font: 500 13.5px / 1.4 var(--nla-font);
  color: var(--nla-text);
  text-align: right;
  word-break: break-word;
}
.info-val--mono {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  font-weight: 500;
}
.info-val--success { color: var(--nla-success); }
.info-val--danger  { color: var(--nla-danger); }
.info-val--primary { color: var(--nla-primary); }
.info-val--muted   { color: var(--nla-text-muted); }

@media (max-width: 480px) {
  .info-row { grid-template-columns: 1fr auto; gap: 10px; padding: 10px 14px; }
}
</style>
