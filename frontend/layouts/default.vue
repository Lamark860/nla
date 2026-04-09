<template>
  <div class="min-h-screen flex flex-col" style="background-color: var(--nla-bg); color: var(--nla-text);">
    <!-- Header (glassmorphism, matches ASH navbar) -->
    <header class="nla-navbar sticky top-0 z-40">
      <div class="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-14">
          <!-- Logo -->
          <NuxtLink to="/" class="nla-brand">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" /></svg>
            NLA
          </NuxtLink>

          <!-- Nav -->
          <nav class="flex items-center gap-1">
            <NuxtLink
              v-for="item in navItems"
              :key="item.path"
              :to="item.path"
              :class="['nla-nav-link', isActive(item.path) && 'nla-nav-link--active']"
            >
              {{ item.label }}
            </NuxtLink>

            <!-- Favorites (if logged in) -->
            <NuxtLink
              v-if="auth.isLoggedIn.value"
              to="/favorites"
              :class="['nla-nav-link', isActive('/favorites') && 'nla-nav-link--active']"
            >
              ★ <span v-if="favorites.count.value > 0" class="tabular-nums">{{ favorites.count.value }}</span><span v-else>Избранное</span>
            </NuxtLink>

            <!-- Auth -->
            <template v-if="auth.isLoggedIn.value">
              <span class="text-xs ml-2 hidden sm:inline" style="color: var(--nla-text-muted);">{{ auth.user.value?.name || auth.user.value?.email }}</span>
              <button
                @click="auth.logout()"
                class="nla-theme-toggle ml-1"
                title="Выход"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </template>
            <NuxtLink v-else to="/login" class="nla-nav-link">Войти</NuxtLink>

            <!-- Theme toggle -->
            <button
              @click="toggleDark"
              class="nla-theme-toggle ml-1"
              :title="isDark ? 'Светлая тема' : 'Тёмная тема'"
            >
              <svg v-if="isDark" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
              </svg>
            </button>
          </nav>
        </div>
      </div>
    </header>

    <!-- Main -->
    <main class="flex-1">
      <div class="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <slot />
      </div>
    </main>

    <!-- Footer (ASH matched) -->
    <footer style="background: var(--nla-bg-card); border-top: 1px solid var(--nla-border); color: var(--nla-text-muted); font-size: 0.85rem; padding: 1.5rem 0;">
      <div class="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between">
          <p>NLA · Данные MOEX ISS · AI OpenAI</p>
          <p>{{ new Date().getFullYear() }}</p>
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
]

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(path + '/')
}

function toggleDark() {
  colorMode.preference = isDark.value ? 'light' : 'dark'
}
</script>
