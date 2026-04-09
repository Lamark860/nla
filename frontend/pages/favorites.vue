<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <svg class="w-6 h-6 text-amber-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
          </svg>
          Избранное
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Облигации, которые вы отслеживаете
        </p>
      </div>
      <span v-if="bonds.length" class="text-sm text-slate-400 font-mono">{{ bonds.length }} шт.</span>
    </div>

    <!-- Not logged in -->
    <div v-if="!auth.isLoggedIn.value" class="card p-12 text-center">
      <svg class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
      </svg>
      <p class="text-slate-500 dark:text-slate-400 mb-4">Войдите, чтобы сохранять облигации в избранное</p>
      <NuxtLink to="/login" class="btn-primary text-sm">Войти</NuxtLink>
    </div>

    <!-- Loading -->
    <div v-else-if="loading" class="card p-12 text-center">
      <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full mx-auto"></div>
      <p class="text-sm text-slate-400 mt-3">Загрузка...</p>
    </div>

    <!-- Empty -->
    <div v-else-if="bonds.length === 0" class="card p-12 text-center">
      <svg class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
      </svg>
      <p class="text-slate-500 dark:text-slate-400 mb-2">Пока тут пусто</p>
      <p class="text-xs text-slate-400 dark:text-slate-500 mb-4">Нажмите ★ на любой облигации, чтобы добавить</p>
      <NuxtLink to="/" class="btn-secondary text-sm">К списку облигаций</NuxtLink>
    </div>

    <!-- Favorites table -->
    <div v-else class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="data-table" style="min-width: 800px">
          <thead>
            <tr>
              <th class="text-left" style="width: 240px">НАЗВАНИЕ</th>
              <th class="text-right" style="width: 85px">ДОХ.</th>
              <th class="text-right" style="width: 95px">ЦЕНА</th>
              <th class="text-right" style="width: 100px">КУПОН</th>
              <th class="text-right" style="width: 120px">ПОГАШЕНИЕ</th>
              <th class="text-center" style="width: 60px">★</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="bond in bonds"
              :key="bond.secid"
              class="cursor-pointer group"
              @click="$router.push(`/bonds/${bond.secid}`)"
            >
              <td>
                <div class="flex flex-col gap-0.5">
                  <span class="font-medium text-slate-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">{{ bond.shortname }}</span>
                  <span class="text-xs text-slate-400 dark:text-slate-500 font-mono">{{ bond.isin }}</span>
                </div>
              </td>
              <td class="text-right font-mono tabular-nums">
                <span :class="yieldColor(bond.yield)" class="font-semibold">{{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}</span>
              </td>
              <td class="text-right font-mono tabular-nums">
                <div class="flex flex-col items-end">
                  <span class="text-slate-900 dark:text-white">{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</span>
                  <span v-if="bond.price_rub" class="text-xs text-emerald-600 dark:text-emerald-400">{{ fmt.priceRub(bond.price_rub) }}</span>
                </div>
              </td>
              <td class="text-right font-mono tabular-nums">
                <span class="text-slate-900 dark:text-white">{{ fmt.percent(bond.coupon_display) }}</span>
              </td>
              <td class="text-right tabular-nums">
                <div class="flex flex-col items-end">
                  <span class="text-slate-700 dark:text-slate-300 font-mono text-xs">{{ fmt.date(bond.matdate) }}</span>
                  <span class="text-xs text-slate-400 dark:text-slate-500">{{ fmt.daysToMaturity(bond.days_to_maturity) }}</span>
                </div>
              </td>
              <td class="text-center">
                <button
                  class="inline-flex items-center justify-center w-8 h-8 rounded-full text-amber-500 hover:text-red-500 transition-colors"
                  title="Убрать из избранного"
                  @click.stop="removeFav(bond.secid)"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const auth = useAuth()
const favorites = useFavorites()
const api = useApi()
const fmt = useFormat()

const bonds = ref<Bond[]>([])
const loading = ref(true)

async function loadFavorites() {
  if (!auth.isLoggedIn.value) {
    loading.value = false
    return
  }

  loading.value = true
  try {
    await favorites.load()
    const secids = [...favorites.favoriteSecids.value]
    if (secids.length === 0) {
      bonds.value = []
      return
    }

    // Fetch each bond — parallel
    const results = await Promise.allSettled(
      secids.map(secid => api.getBond(secid))
    )
    bonds.value = results
      .filter((r): r is PromiseFulfilledResult<Bond> => r.status === 'fulfilled')
      .map(r => r.value)
  } catch {
    bonds.value = []
  } finally {
    loading.value = false
  }
}

async function removeFav(secid: string) {
  await favorites.toggle(secid)
  bonds.value = bonds.value.filter(b => b.secid !== secid)
}

function yieldColor(y: number | null | undefined): string {
  if (y == null) return 'text-slate-400'
  if (y >= 12) return 'text-emerald-600 dark:text-emerald-400'
  if (y >= 8) return 'text-primary-600 dark:text-primary-400'
  return 'text-slate-500 dark:text-slate-400'
}

onMounted(loadFavorites)

watch(() => auth.isLoggedIn.value, (v) => {
  if (v) loadFavorites()
  else { bonds.value = []; loading.value = false }
})
</script>
