<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold" style="color: var(--nla-text)">Месячные купоны по эмитентам</h1>
      <div class="flex items-center gap-2">
        <NuxtLink to="/bonds/by-issuer" class="btn-secondary text-sm">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>
          Все по эмитентам
        </NuxtLink>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="card p-16 text-center">
      <div class="inline-block w-6 h-6 border-2 border-primary-200 border-t-primary-600 dark:border-primary-800 dark:border-t-primary-400 rounded-full animate-spin"></div>
      <p class="mt-4 text-xs" style="color: var(--nla-text-muted)">Загрузка…</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="card p-10 text-center">
      <p class="text-red-600 dark:text-red-400 text-sm">{{ error.message }}</p>
      <button class="btn-primary mt-4 text-sm" @click="refresh()">Повторить</button>
    </div>

    <template v-else-if="groupedData">
      <IssuerFilters
        :filters="filters"
        :stats="{ issuers: filteredGroups.length, bonds: filteredBondCount, total: groupedData.total_issuers ?? 0 }"
        :show-period="false"
        :show-category="false"
        @update="filters = $event"
        @reset="resetFilters"
      />
      <IssuerCardGrid :issuers="filteredGroups" :ratings="ratingsMap" />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { Bond, IssuerGroup, IssuerRatingResponse } from '~/composables/useApi'
import type { IssuerFilterValues } from '~/components/IssuerFilters.vue'

const api = useApi()

const filters = ref<IssuerFilterValues>({
  search: '', couponType: '', rating: '', period: '', category: '',
  yieldMin: null, yieldMax: null, couponMin: null, couponMax: null,
  maturityMin: null, maturityMax: null, priceMax: null,
})

const { data: groupedData, pending, error, refresh } = useAsyncData('monthly-grouped', () => api.getBondsGrouped(true))

const ratingsMap = ref<Record<string, IssuerRatingResponse>>({})
const { data: ratingsData } = useAsyncData('monthly-ratings', () => api.getRatings(), { default: () => ({}) })
watch(ratingsData, (v) => { if (v) ratingsMap.value = v }, { immediate: true })

const filteredGroups = computed(() => {
  if (!groupedData.value) return []
  return groupedData.value.groups
    .map(group => {
      const bonds = group.bonds.filter(matchesBond)
      if (bonds.length === 0) return null
      return { ...group, bonds, bond_count: bonds.length }
    })
    .filter((g): g is IssuerGroup => g !== null)
    .filter(matchesIssuerRating)
})

const filteredBondCount = computed(() => filteredGroups.value.reduce((acc, g) => acc + g.bond_count, 0))

function matchesBond(bond: Bond): boolean {
  const f = filters.value
  if (f.search) {
    const q = f.search.toLowerCase()
    if (!bond.shortname.toLowerCase().includes(q) && !bond.secname.toLowerCase().includes(q) && !bond.secid.toLowerCase().includes(q) && !(bond.isin && bond.isin.toLowerCase().includes(q))) return false
  }
  if (f.couponType === 'float' && !bond.is_float) return false
  if (f.couponType === 'indexed' && !bond.is_indexed) return false
  if (f.couponType === 'fixed' && (bond.is_float || bond.is_indexed)) return false
  if (f.yieldMin != null && (bond.yield == null || bond.yield < f.yieldMin)) return false
  if (f.yieldMax != null && (bond.yield == null || bond.yield > f.yieldMax)) return false
  if (f.couponMin != null && (bond.coupon_display == null || bond.coupon_display < f.couponMin)) return false
  if (f.couponMax != null && (bond.coupon_display == null || bond.coupon_display > f.couponMax)) return false
  if (f.maturityMin != null && (bond.days_to_maturity == null || bond.days_to_maturity < f.maturityMin)) return false
  if (f.maturityMax != null && (bond.days_to_maturity == null || bond.days_to_maturity > f.maturityMax)) return false
  if (f.priceMax != null && (bond.last == null || bond.last > f.priceMax)) return false
  return true
}

function matchesIssuerRating(group: IssuerGroup): boolean {
  const r = filters.value.rating
  if (!r) return true
  const rating = getIssuerRating(group.emitter_name)
  const score = rating?.score ?? -1
  switch (r) {
    case 'aaa': return score >= 9
    case 'aa': return score >= 7 && score <= 8
    case 'a': return score >= 5 && score <= 6
    case 'bbb': return score >= 1 && score <= 4
    case 'none': return score < 0
    default: return true
  }
}

function getIssuerRating(name: string): IssuerRatingResponse | null {
  if (ratingsMap.value[name]) return ratingsMap.value[name]
  const lower = name.toLowerCase()
  for (const [key, val] of Object.entries(ratingsMap.value)) {
    if (lower.includes(key.toLowerCase()) || key.toLowerCase().includes(lower)) return val
  }
  return null
}

function resetFilters() {
  filters.value = { search: '', couponType: '', rating: '', period: '', category: '', yieldMin: null, yieldMax: null, couponMin: null, couponMax: null, maturityMin: null, maturityMax: null, priceMax: null }
}

useHead({ title: 'Месячные купоны по эмитентам — NLA' })
</script>
