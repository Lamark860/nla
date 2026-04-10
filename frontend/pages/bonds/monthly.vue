<template>
  <div>
    <!-- Header -->
    <div class="d-flex align-items-center justify-content-between mb-4">
      <h1 class="h4 fw-bold mb-0">Месячные купоны по эмитентам</h1>
      <NuxtLink to="/bonds/by-issuer" class="btn btn-outline-secondary btn-sm">
        <i class="bi bi-grid me-1"></i>Все по эмитентам
      </NuxtLink>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="card p-5 text-center">
      <div class="spinner-border" role="status"><span class="visually-hidden">Загрузка…</span></div>
      <p class="mt-3 small text-muted">Загрузка…</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="card p-5 text-center">
      <p class="text-danger small">{{ error.message }}</p>
      <button class="btn btn-primary btn-sm mt-3" @click="refresh()">Повторить</button>
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
      <IssuerCardGrid :issuers="filteredGroups" :ratings="ratingsMap" :ai-stats="aiStatsMap" />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { Bond, IssuerGroup, IssuerRatingResponse, AnalysisStats } from '~/composables/useApi'
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

const aiStatsMap = ref<Record<string, AnalysisStats>>({})
const { data: aiStatsData } = useAsyncData('monthly-ai-stats', () => api.getBulkAnalysisStats(), { default: () => ({}) })
watch(aiStatsData, (v) => { if (v) aiStatsMap.value = v }, { immediate: true })

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
  const rating = getIssuerRating(group.emitter_id)
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

function getIssuerRating(emitterId: number): IssuerRatingResponse | null {
  if (!emitterId) return null
  return ratingsMap.value[String(emitterId)] ?? null
}

function resetFilters() {
  filters.value = { search: '', couponType: '', rating: '', period: '', category: '', yieldMin: null, yieldMax: null, couponMin: null, couponMax: null, maturityMin: null, maturityMax: null, priceMax: null }
}

useHead({ title: 'Месячные купоны по эмитентам — NLA' })
</script>
