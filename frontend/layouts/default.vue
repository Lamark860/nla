<template>
  <div class="app-shell" :class="{ 'is-mobile-open': mobileNavOpen }">
    <!-- Mobile top bar (visible <= 768px) -->
    <div class="mobile-bar">
      <button class="mobile-bar__toggle" :aria-expanded="mobileNavOpen" aria-label="Меню" @click="mobileNavOpen = !mobileNavOpen">
        <i class="bi" :class="mobileNavOpen ? 'bi-x-lg' : 'bi-list'"></i>
      </button>
      <NuxtLink to="/" class="mobile-bar__brand">
        <span class="mobile-bar__mark">N</span>
        <span class="mobile-bar__title">NLA</span>
      </NuxtLink>
      <button class="mobile-bar__theme" @click="toggleDark" :title="isDark ? 'Светлая тема' : 'Тёмная тема'">
        <i :class="isDark ? 'bi bi-sun-fill' : 'bi bi-moon-fill'"></i>
      </button>
    </div>

    <!-- Backdrop (mobile only) -->
    <div v-if="mobileNavOpen" class="app-shell__backdrop" @click="mobileNavOpen = false"></div>

    <!-- Sidebar -->
    <aside class="app-sidebar">
      <NuxtLink to="/" class="app-sidebar__brand" @click="mobileNavOpen = false">
        <span class="app-sidebar__mark">N</span>
        <span class="app-sidebar__title">
          NLA <span class="app-sidebar__title-meta">· bonds</span>
        </span>
      </NuxtLink>

      <div class="app-sidebar__section">Каталог</div>
      <NuxtLink
        v-for="item in catalogNav"
        :key="item.path"
        :to="item.path"
        class="app-sidebar__item"
        :class="{ 'is-active': isActive(item.path) }"
        @click="mobileNavOpen = false"
      >
        <i :class="`bi bi-${item.icon}`" aria-hidden="true"></i>
        <span>{{ item.label }}</span>
      </NuxtLink>

      <div class="app-sidebar__section">Инструменты</div>
      <NuxtLink
        v-for="item in toolsNav"
        :key="item.path"
        :to="item.path"
        class="app-sidebar__item"
        :class="{ 'is-active': isActive(item.path) }"
        @click="mobileNavOpen = false"
      >
        <i :class="`bi bi-${item.icon}`" aria-hidden="true"></i>
        <span>{{ item.label }}</span>
      </NuxtLink>

      <div class="app-sidebar__bottom">
        <NuxtLink
          v-if="auth.isLoggedIn.value"
          to="/favorites"
          class="app-sidebar__item"
          :class="{ 'is-active': isActive('/favorites') }"
          @click="mobileNavOpen = false"
        >
          <i class="bi bi-star-fill"></i>
          <span>Избранное</span>
          <span v-if="favorites.count.value > 0" class="app-sidebar__badge">{{ favorites.count.value }}</span>
        </NuxtLink>

        <div v-if="auth.isLoggedIn.value" class="app-sidebar__user">
          <i class="bi bi-person-circle"></i>
          <span class="app-sidebar__user-name">{{ auth.user.value?.name || auth.user.value?.email }}</span>
          <button @click="auth.logout()" class="app-sidebar__logout" title="Выход">
            <i class="bi bi-box-arrow-right"></i>
          </button>
        </div>
        <NuxtLink v-else to="/login" class="app-sidebar__item" :class="{ 'is-active': isActive('/login') }" @click="mobileNavOpen = false">
          <i class="bi bi-box-arrow-in-right"></i>
          <span>Войти</span>
        </NuxtLink>

        <button @click="toggleDark" class="app-sidebar__theme" :title="isDark ? 'Светлая тема' : 'Тёмная тема'">
          <i :class="isDark ? 'bi bi-sun-fill' : 'bi bi-moon-fill'"></i>
          <span>{{ isDark ? 'Светлая тема' : 'Тёмная тема' }}</span>
        </button>
      </div>
    </aside>

    <!-- Main -->
    <main class="app-main">
      <div class="app-main__inner">
        <slot />
      </div>

      <footer class="app-footer">
        <div class="app-footer__main">
          <span>NLA · Данные MOEX ISS</span>
          <span>{{ year }}</span>
        </div>
        <div class="app-footer__disclaimer">
          Информация на сайте носит аналитический характер и не является индивидуальной инвестиционной рекомендацией.
        </div>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const colorMode = useColorMode()
const auth = useAuth()
const favorites = useFavorites()

const isDark = computed(() => colorMode.value === 'dark')
const year = new Date().getFullYear()
const mobileNavOpen = ref(false)

const catalogNav = [
  { path: '/bonds/by-issuer', label: 'Эмитенты',          icon: 'collection' },
  { path: '/bonds/monthly',   label: 'Месячные купоны',   icon: 'calendar3' },
  { path: '/chat',            label: 'AI-чат',            icon: 'stars' },
] as const

const toolsNav = [
  { path: '/bonds/flat',  label: 'Плоский список', icon: 'table' },
  { path: '/tools',       label: 'Утилиты',        icon: 'wrench-adjustable' },
] as const

function isActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(path + '/')
}

function toggleDark() {
  colorMode.preference = isDark.value ? 'light' : 'dark'
}

// Close mobile nav on route change
watch(() => route.path, () => { mobileNavOpen.value = false })
</script>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: 240px 1fr;
  min-height: 100vh;
  background: var(--nla-bg);
  color: var(--nla-text);
}

/* Sidebar */
.app-sidebar {
  background: var(--nla-bg-elevated);
  border-right: 1px solid var(--nla-border);
  padding: 20px 14px;
  position: sticky;
  top: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  z-index: 100;
}

.app-sidebar__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 4px 28px;
  font: 700 15px/1 var(--nla-font);
  letter-spacing: -0.01em;
  color: var(--nla-text);
  text-decoration: none;
}
.app-sidebar__mark {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--nla-primary);
  color: #fff;
  display: grid;
  place-items: center;
  font: 700 12px/1 var(--nla-font);
  box-shadow: var(--nla-shadow-sm);
}
[data-theme="dark"] .app-sidebar__mark { color: #0c0e0d; }
.app-sidebar__title-meta {
  color: var(--nla-text-muted);
  font-weight: 400;
}

.app-sidebar__section {
  font: 600 10px/1 var(--nla-font);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  margin: 14px 8px 6px;
}

.app-sidebar__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 8px;
  color: var(--nla-text-secondary);
  font: 500 13px/1.2 var(--nla-font);
  text-decoration: none;
  transition: background 120ms ease, color 120ms ease;
}
.app-sidebar__item i {
  font-size: 15px;
  width: 16px;
  display: inline-flex;
  justify-content: center;
}
.app-sidebar__item:hover {
  background: var(--nla-bg-subtle);
  color: var(--nla-text);
}
.app-sidebar__item.is-active {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  font-weight: 600;
}
[data-theme="dark"] .app-sidebar__item.is-active {
  color: var(--nla-primary);
}
.app-sidebar__badge {
  margin-left: auto;
  font: 600 11px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text-muted);
  padding: 2px 6px;
  background: var(--nla-bg-card);
  border-radius: 4px;
}

.app-sidebar__bottom {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-top: 12px;
  border-top: 1px solid var(--nla-border-light);
}
.app-sidebar__user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  font: 500 12px/1.2 var(--nla-font);
  color: var(--nla-text-muted);
}
.app-sidebar__user-name {
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.app-sidebar__logout {
  background: transparent;
  border: 0;
  padding: 4px;
  color: var(--nla-text-muted);
  cursor: pointer;
  border-radius: var(--nla-radius-sm);
}
.app-sidebar__logout:hover { color: var(--nla-danger); background: var(--nla-bg-subtle); }

.app-sidebar__theme {
  appearance: none;
  border: 1px solid var(--nla-border);
  background: var(--nla-bg-card);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 8px;
  color: var(--nla-text-secondary);
  font: 500 12.5px/1 var(--nla-font);
  cursor: pointer;
  margin-top: 6px;
  transition: background 120ms ease, color 120ms ease;
}
.app-sidebar__theme:hover { background: var(--nla-primary-light); color: var(--nla-primary-ink); border-color: color-mix(in oklab, var(--nla-primary) 25%, var(--nla-border)); }
.app-sidebar__theme i { font-size: 14px; }

/* Main */
.app-main {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  min-width: 0;
}
.app-main__inner {
  flex: 1 1 auto;
  padding: 28px 36px 60px;
  max-width: 1480px;
  width: 100%;
}
.app-footer {
  padding: 16px 36px;
  border-top: 1px solid var(--nla-border);
  background: var(--nla-bg-elevated);
  display: flex;
  flex-direction: column;
  gap: 6px;
  font: 500 12px/1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.app-footer__main { display: flex; justify-content: space-between; }
.app-footer__disclaimer { font-size: 11px; color: var(--nla-text-subtle, var(--nla-text-muted)); line-height: 1.5; }

/* Mobile top bar */
.mobile-bar {
  display: none;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--nla-border);
  background: var(--nla-bg-elevated);
  position: sticky;
  top: 0;
  z-index: 90;
}
.mobile-bar__toggle,
.mobile-bar__theme {
  appearance: none;
  border: 0;
  background: transparent;
  padding: 0;
  border-radius: var(--nla-radius-sm);
  color: var(--nla-text);
  cursor: pointer;
  font-size: 18px;
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
}
.mobile-bar__toggle:hover,
.mobile-bar__theme:hover { background: var(--nla-bg-subtle); }
.mobile-bar__brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font: 700 15px/1 var(--nla-font);
  color: var(--nla-text);
  text-decoration: none;
  flex: 1 1 auto;
}
.mobile-bar__mark {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: var(--nla-primary);
  color: #fff;
  display: grid;
  place-items: center;
  font: 700 11px/1 var(--nla-font);
}
[data-theme="dark"] .mobile-bar__mark { color: #0c0e0d; }

.app-shell__backdrop {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 95;
}

@media (max-width: 768px) {
  .app-shell {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }
  .mobile-bar { display: flex; }
  .app-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 260px;
    height: 100vh;
    transform: translateX(-100%);
    transition: transform 220ms ease;
  }
  .is-mobile-open .app-sidebar { transform: translateX(0); }
  .is-mobile-open .app-shell__backdrop { display: block; }
  .app-main__inner { padding: 16px 14px 40px; }
  .app-footer { padding: 14px 16px; flex-direction: column; gap: 4px; }
}
</style>
