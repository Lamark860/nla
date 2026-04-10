<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Облигации по эмитентам</h1>

    <!-- Loading -->
    <div v-if="pending" class="card p-5 text-center">
      <div class="spinner-border" role="status"><span class="visually-hidden">Загрузка…</span></div>
      <p class="mt-3 small text-muted">Загрузка облигаций…</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="card p-5 text-center">
      <p class="text-danger small">{{ error.message || 'Ошибка загрузки' }}</p>
      <button class="btn btn-primary btn-sm mt-3" @click="refresh()">Повторить</button>
    </div>

    <template v-else-if="groupedData">
      <IssuerFilters
        :filters="filters"
        :stats="{ issuers: filteredIssuers.length, bonds: filteredBondCount, total: groupedData.total_issuers ?? 0 }"
        :show-period="true"
        :show-category="true"
        @update="filters = $event"
        @reset="resetFilters"
      />
      <IssuerCardGrid :issuers="filteredIssuers" :ratings="ratingsMap" :ai-stats="aiStatsMap" />
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

const { data: groupedData, pending, error, refresh } = useAsyncData('bonds-grouped', () => api.getBondsGrouped())

const ratingsMap = ref<Record<string, IssuerRatingResponse>>({})
const { data: ratingsData } = useAsyncData('ratings', () => api.getRatings(), { default: () => ({}) })
watch(ratingsData, (v) => { if (v) ratingsMap.value = v }, { immediate: true })

const aiStatsMap = ref<Record<string, AnalysisStats>>({})
const { data: aiStatsData } = useAsyncData('ai-bulk-stats', () => api.getBulkAnalysisStats(), { default: () => ({}) })
watch(aiStatsData, (v) => { if (v) aiStatsMap.value = v }, { immediate: true })

const filteredIssuers = computed(() => {
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

const filteredBondCount = computed(() => filteredIssuers.value.reduce((acc, g) => acc + g.bond_count, 0))

function matchesBond(bond: Bond): boolean {
  const f = filters.value
  if (f.search) {
    const q = f.search.toLowerCase()
    if (!bond.shortname.toLowerCase().includes(q) && !bond.secname.toLowerCase().includes(q) && !bond.isin.toLowerCase().includes(q) && !bond.secid.toLowerCase().includes(q)) return false
  }
  if (f.category && bond.bond_category !== f.category) return false
  if (f.couponType === 'float' && !bond.is_float) return false
  if (f.couponType === 'indexed' && !bond.is_indexed) return false
  if (f.couponType === 'fixed' && (bond.is_float || bond.is_indexed)) return false
  if (f.period) {
    const p = bond.coupon_period
    if (f.period === 'monthly' && (p < 27 || p > 33)) return false
    if (f.period === 'quarterly' && (p < 85 || p > 95)) return false
    if (f.period === 'semiannual' && (p < 175 || p > 190)) return false
    if (f.period === 'annual' && (p < 355 || p > 370)) return false
  }
  if (f.yieldMin != null && (bond.yield == null || bond.yield < f.yieldMin)) return false
  if (f.yieldMax != null && (bond.yield == null || bond.yield > f.yieldMax)) return false
  if (f.couponMin != null && (bond.coupon_percent == null || bond.coupon_percent < f.couponMin)) return false
  if (f.couponMax != null && (bond.coupon_percent == null || bond.coupon_percent > f.couponMax)) return false
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
  const key = String(emitterId)
  return ratingsMap.value[key] ?? null
}

function resetFilters() {
  filters.value = { search: '', couponType: '', rating: '', period: '', category: '', yieldMin: null, yieldMax: null, couponMin: null, couponMax: null, maturityMin: null, maturityMax: null, priceMax: null }
}

useHead({ title: 'По эмитентам — NLA' })
</script>
