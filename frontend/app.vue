<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup lang="ts">
const auth = useAuth()
const favorites = useFavorites()

onMounted(() => {
  auth.init()
  // Load favorites once auth is ready
  watch(() => auth.isLoggedIn.value, (loggedIn) => {
    if (loggedIn) favorites.load()
  }, { immediate: true })
})
</script>
