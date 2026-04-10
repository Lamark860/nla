<template>
  <div class="d-flex flex-column min-vh-100" style="background-color: var(--nla-bg); color: var(--nla-text);">
    <!-- Header (ASH navbar) -->
    <header id="header">
      <nav class="navbar navbar-expand-md fixed-top">
        <div class="container">
          <NuxtLink to="/" class="navbar-brand">
            <i class="bi bi-bar-chart-line me-1"></i> NLA
          </NuxtLink>
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
            <span class="navbar-toggler-icon"></span>
          </button>
          <div class="collapse navbar-collapse" id="navbarNav">
            <ul class="navbar-nav me-auto">
              <li v-for="item in navItems" :key="item.path" class="nav-item">
                <NuxtLink :to="item.path" :class="['nav-link', isActive(item.path) && 'active']">
                  {{ item.label }}
                </NuxtLink>
              </li>
            </ul>
            <div class="d-flex align-items-center gap-2">
              <!-- Favorites -->
              <NuxtLink
                v-if="auth.isLoggedIn.value"
                to="/favorites"
                :class="['nav-link', isActive('/favorites') && 'active']"
              >
                ★ <span v-if="favorites.count.value > 0">{{ favorites.count.value }}</span><span v-else>Избранное</span>
              </NuxtLink>

              <!-- Auth -->
              <template v-if="auth.isLoggedIn.value">
                <span class="small text-muted d-none d-sm-inline">{{ auth.user.value?.name || auth.user.value?.email }}</span>
                <button @click="auth.logout()" class="theme-toggle" title="Выход">
                  <i class="bi bi-box-arrow-right"></i>
                </button>
              </template>
              <NuxtLink v-else to="/login" class="nav-link">Войти</NuxtLink>

              <!-- Theme toggle -->
              <button @click="toggleDark" class="theme-toggle" :title="isDark ? 'Светлая тема' : 'Тёмная тема'">
                <i :class="isDark ? 'bi bi-sun-fill' : 'bi bi-moon-fill'"></i>
              </button>
            </div>
          </div>
        </div>
      </nav>
    </header>

    <!-- Main -->
    <main id="main" class="flex-shrink-0">
      <div class="container">
        <slot />
      </div>
    </main>

    <!-- Footer -->
    <footer id="footer" class="mt-auto">
      <div class="container">
        <div class="d-flex justify-content-between align-items-center">
          <p class="mb-0">NLA · Данные MOEX ISS · AI OpenAI</p>
          <p class="mb-0">{{ new Date().getFullYear() }}</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const colorMode = useColorMode()
const auth = useAuth()
const favorites = useFavorites()

const isDark = computed(() => colorMode.value === 'dark')

const navItems = [
  { path: '/', label: 'Облигации' },
  { path: '/bonds/by-issuer', label: 'Эмитенты' },
  { path: '/bonds/monthly', label: 'Купоны' },
  { path: '/chat', label: 'Чат' },
  { path: '/tools', label: 'Инструменты' },
]

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(path + '/')
}

function toggleDark() {
  colorMode.preference = isDark.value ? 'light' : 'dark'
}
</script>
