<template>
  <!-- Redirect-страница: главная теперь /bonds/by-issuer (Шаг 23). -->
  <div class="redirect-pad">
    <div class="spinner-border" role="status" style="color: var(--nla-primary)"><span class="visually-hidden">Перенаправление…</span></div>
  </div>
</template>

<script setup lang="ts">
// SSR redirect — страница не рендерится, серверный 302 уходит сразу.
if (import.meta.server) {
  await navigateTo('/bonds/by-issuer', { redirectCode: 302 })
}
onMounted(() => {
  navigateTo('/bonds/by-issuer', { replace: true })
})
useHead({ title: 'NLA' })
</script>

<style scoped>
.redirect-pad {
  min-height: 60vh;
  display: grid;
  place-items: center;
}
</style>
