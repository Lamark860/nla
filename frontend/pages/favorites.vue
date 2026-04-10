<template>
  <div>
    <!-- Header -->
    <div class="d-flex align-items-center justify-content-between mb-4">
      <div>
        <h1 class="h4 fw-bold d-flex align-items-center gap-2 mb-1">
          <i class="bi bi-star-fill text-warning"></i>
          Избранное
        </h1>
        <p class="small text-muted mb-0">Облигации, которые вы отслеживаете</p>
      </div>
      <span v-if="bonds.length" class="small text-muted font-monospace">{{ bonds.length }} шт.</span>
    </div>

    <!-- Not logged in -->
    <div v-if="!auth.isLoggedIn.value" class="card text-center py-5 px-4">
      <i class="bi bi-person-circle text-muted d-block mb-3" style="font-size: 48px"></i>
      <p class="text-muted mb-3">Войдите, чтобы сохранять облигации в избранное</p>
      <div>
        <NuxtLink to="/login" class="btn btn-primary btn-sm">Войти</NuxtLink>
      </div>
    </div>

    <!-- Loading -->
    <div v-else-if="loading" class="card text-center py-5">
      <div class="spinner-border text-primary mx-auto mb-3" role="status"></div>
      <p class="small text-muted">Загрузка...</p>
    </div>

    <!-- Empty -->
    <div v-else-if="bonds.length === 0" class="card text-center py-5 px-4">
      <i class="bi bi-star text-muted d-block mb-3" style="font-size: 48px"></i>
      <p class="text-muted mb-1">Пока тут пусто</p>
      <p class="small text-muted mb-3">Нажмите ★ на любой облигации, чтобы добавить</p>
      <div>
        <NuxtLink to="/" class="btn btn-outline-secondary btn-sm">К списку облигаций</NuxtLink>
      </div>
    </div>

    <!-- Favorites table -->
    <div v-else class="card overflow-hidden">
      <div class="table-responsive">
        <table class="data-table" style="min-width: 800px">
          <thead>
            <tr>
              <th class="text-start" style="width: 240px">НАЗВАНИЕ</th>
              <th class="text-end" style="width: 85px">ДОХ.</th>
              <th class="text-end" style="width: 95px">ЦЕНА</th>
              <th class="text-end" style="width: 100px">КУПОН</th>
              <th class="text-end" style="width: 120px">ПОГАШЕНИЕ</th>
              <th class="text-center" style="width: 60px">★</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="bond in bonds"
              :key="bond.secid"
              style="cursor: pointer"
              @click="$router.push(`/bonds/${bond.secid}`)"
            >
              <td>
                <div>
                  <span class="fw-medium">{{ bond.shortname }}</span>
                  <div class="small text-muted font-monospace">{{ bond.isin }}</div>
                </div>
              </td>
              <td class="text-end font-monospace">
                <span :class="yieldColor(bond.yield)" class="fw-semibold">{{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}</span>
              </td>
              <td class="text-end font-monospace">
                <div>{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</div>
                <div v-if="bond.price_rub" class="small text-success">{{ fmt.priceRub(bond.price_rub) }}</div>
              </td>
              <td class="text-end font-monospace">
                {{ fmt.percent(bond.coupon_display) }}
              </td>
              <td class="text-end">
                <div class="font-monospace small">{{ fmt.date(bond.matdate) }}</div>
                <div class="small text-muted">{{ fmt.daysToMaturity(bond.days_to_maturity) }}</div>
              </td>
              <td class="text-center">
                <button
                  class="btn btn-link p-0 text-warning"
                  title="Убрать из избранного"
                  @click.stop="removeFav(bond.secid)"
                >
                  <i class="bi bi-star-fill"></i>
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
  if (y == null) return 'text-muted'
  if (y >= 12) return 'text-success'
  if (y >= 8) return 'text-primary'
  return ''
}

onMounted(loadFavorites)

watch(() => auth.isLoggedIn.value, (v) => {
  if (v) loadFavorites()
  else { bonds.value = []; loading.value = false }
})
</script>
