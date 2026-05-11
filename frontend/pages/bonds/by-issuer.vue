<template>
  <div>
    <PageHead title="Облигации по эмитентам">
      <template #sub>
        Сгруппировано по эмитентам ·
        <b>{{ totalIssuers }}</b> компаний ·
        <b>{{ totalBonds }}</b> {{ pluralBond(totalBonds) }}
      </template>
      <template #actions>
        <ViewToggle :options="viewOptions" />
      </template>
    </PageHead>

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
      <IssuerFilters class="mb-4" :stats="filterStats" @change="onFiltersChange" />
      <IssuerCardGrid :issuers="filteredIssuers" :ratings="ratingsMap" :ai-stats="aiStatsMap" />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { IssuerGroup, IssuerRatingResponse, AnalysisStats } from '~/composables/useApi'
import {
  emptyIssuerFilterState,
  matchesBond,
  matchesIssuerRating,
  matchesIssuerAi,
  type IssuerFilterState,
} from '~/composables/useIssuerFilters'

const api = useApi()

const filters = ref<IssuerFilterState>(emptyIssuerFilterState())

const { data: groupedData, pending, error, refresh } = useAsyncData('bonds-grouped', () => api.getBondsGrouped())

const ratingsMap = ref<Record<string, IssuerRatingResponse>>({})
const { data: ratingsData } = useAsyncData('ratings', () => api.getRatings(), { default: () => ({}) })
watch(ratingsData, (v) => { if (v) ratingsMap.value = v }, { immediate: true })

const aiStatsMap = ref<Record<string, AnalysisStats>>({})
const { data: aiStatsData } = useAsyncData('ai-bulk-stats', () => api.getBulkAnalysisStats(), { default: () => ({}) })
watch(aiStatsData, (v) => { if (v) aiStatsMap.value = v }, { immediate: true })

const filteredIssuers = computed(() => {
  if (!groupedData.value) return []
  const f = filters.value
  return groupedData.value.groups
    .map(group => {
      const bonds = group.bonds.filter(b => matchesBond(b, f))
      if (bonds.length === 0) return null
      return { ...group, bonds, bond_count: bonds.length }
    })
    .filter((g): g is IssuerGroup => g !== null)
    .filter(group => {
      const ratings = ratingsMap.value[String(group.emitter_id)]?.ratings
      if (!matchesIssuerRating(ratings, f.rating, f.hasRating)) return false
      if (!matchesIssuerAi(group.bonds, aiStatsMap.value, f.aiBucket)) return false
      return true
    })
})

const totalIssuers = computed(() => groupedData.value?.total_issuers ?? groupedData.value?.groups.length ?? 0)
const totalBonds = computed(() =>
  groupedData.value?.groups.reduce((acc, g) => acc + g.bond_count, 0) ?? 0
)
const filteredBondCount = computed(() => filteredIssuers.value.reduce((acc, g) => acc + g.bond_count, 0))

const filterStats = computed(() => ({
  issuers: totalIssuers.value,
  bonds: totalBonds.value,
  shown: filteredBondCount.value,
}))

const viewOptions = [
  { path: '/bonds/by-issuer', label: 'Эмитенты',        icon: 'collection' },
  { path: '/bonds/monthly',   label: 'Месячные купоны', icon: 'calendar3' },
]

function onFiltersChange(value: IssuerFilterState) {
  filters.value = value
}

function pluralBond(n: number): string {
  const m10 = n % 10, m100 = n % 100
  if (m100 >= 11 && m100 <= 19) return 'выпусков'
  if (m10 === 1) return 'выпуск'
  if (m10 >= 2 && m10 <= 4) return 'выпуска'
  return 'выпусков'
}

useHead({ title: 'По эмитентам — NLA' })
</script>
