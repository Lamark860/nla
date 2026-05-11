<template>
  <div>
    <NuxtLink to="/bonds/by-issuer" class="d-inline-flex align-items-center gap-1 small text-muted text-decoration-none mb-3">
      <i class="bi bi-chevron-left"></i>
      К списку эмитентов
    </NuxtLink>

    <div v-if="pending" class="card p-5 text-center">
      <div class="spinner-border" role="status"><span class="visually-hidden">Загрузка…</span></div>
    </div>

    <IssuerProfile
      v-else-if="group"
      :emitter-id="emitterId"
      :name="group.emitter_name"
      :sector="dohodData?.issuer_sector || dohodData?.sector_text"
      :bonds="group.bonds"
      :ratings="profileRatings"
      :ai-stats="aiStatsData"
      :dohod="dohodData"
      :current-secid="currentSecid"
    />

    <div v-else class="card p-5 text-center">
      <p class="text-muted small">Эмитент с ID {{ emitterId }} не найден</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { IssuerGroup, IssuerRatingResponse, AnalysisStats, DohodBondData } from '~/composables/useApi'
import { normalizeRating } from '~/composables/useRating'

const route = useRoute()
const api = useApi()

const emitterId = Number(route.params.id)
const currentSecid = computed(() => (route.query.current as string) || undefined)

const { data: groupedData, pending } = useAsyncData(`issuer-${emitterId}`, () => api.getBondsGrouped())
const { data: ratingsData } = useAsyncData(`issuer-ratings-${emitterId}`, () => api.getRatings(), { default: () => ({} as Record<string, IssuerRatingResponse>) })
const { data: aiStatsRaw } = useAsyncData(`issuer-ai-${emitterId}`, () => api.getBulkAnalysisStats(), { default: () => ({} as Record<string, AnalysisStats>) })

const aiStatsData = computed<Record<string, AnalysisStats>>(() => aiStatsRaw.value ?? {})

const group = computed<IssuerGroup | null>(() =>
  groupedData.value?.groups.find(g => g.emitter_id === emitterId) ?? null
)

const profileRatings = computed(() => {
  const r = ratingsData.value?.[String(emitterId)]
  if (!r) return []
  return r.ratings
    .filter(x => x.rating && x.rating !== 'NULL')
    .map(x => ({
      agency: x.agency,
      rating: x.rating,
      updated_at: x.updated_at,
      tier: normalizeRating(x.rating).tier,
    }))
})

// Try to fetch dohod for any of the issuer's bonds (issuer-level fields are same).
// Pick first bond — if no data, try second, etc. Limit to 3 attempts.
const { data: dohodData } = useAsyncData<DohodBondData | null>(
  `issuer-dohod-${emitterId}`,
  async () => {
    const bonds = group.value?.bonds ?? []
    if (!bonds.length) return null
    for (const b of bonds.slice(0, 3)) {
      try {
        const res = await api.getDohodDetails(b.secid)
        // job_id response means async fetch in progress — skip
        if (res && typeof res === 'object' && 'isin' in res) {
          return res as DohodBondData
        }
      } catch { /* try next */ }
    }
    return null
  },
  { watch: [group], default: () => null },
)

useHead({ title: computed(() => group.value ? `${group.value.emitter_name} — NLA` : 'Эмитент — NLA') })
</script>
