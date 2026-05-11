<template>
  <div>
    <!-- Back link -->
    <NuxtLink to="/" class="d-inline-flex align-items-center gap-1 small text-muted text-decoration-none mb-4">
      <i class="bi bi-chevron-left"></i>
      Все облигации
    </NuxtLink>

    <!-- Loading -->
    <div v-if="pending" class="card p-5 text-center">
      <div class="spinner-border" role="status"><span class="visually-hidden">Загрузка…</span></div>
      <p class="mt-3 small text-muted">Загрузка данных…</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="card p-5 text-center">
      <p class="text-danger small">{{ error.message || 'Облигация не найдена' }}</p>
      <NuxtLink to="/" class="btn btn-primary btn-sm mt-3">Назад к списку</NuxtLink>
    </div>

    <!-- Bond data -->
    <template v-else-if="data">
      <BondHero
        :bond="data"
        :ratings="heroRatings"
        :ai-score="analysisStats?.avg_rating ?? null"
        :scores="scores ?? null"
        :issuer-name="issuerRating?.issuer"
        :is-favorite="favorites.isFavorite(data.secid)"
        class="mb-4"
        @toggle-favorite="favorites.toggle(data.secid)"
        @share="copyShareLink"
        @analyze="activeTab = 'ai'"
        @open-score="activeTab = 'score'"
      />

      <!-- Tabs (pill-style) -->
      <nav class="bond-tabs mb-4">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          :class="activeTab === tab.id ? 'bond-tab bond-tab--active' : 'bond-tab'"
          @click="activeTab = tab.id"
        >
          <i :class="tab.icon" class="bi"></i>
          <span>{{ tab.label }}</span>
        </button>
      </nav>

      <!-- Tab content -->
      <div>
        <BondInfoBasic v-if="activeTab === 'basic'" :bond="data" />
        <BondTradingTab v-else-if="activeTab === 'trading'" :bond="data" />
        <BondCouponsTab v-else-if="activeTab === 'coupons'" :coupons="coupons ?? []" :bond="data" />
        <BondYieldsTab v-else-if="activeTab === 'yields'" :bond="data" />
        <BondHistoryTab v-else-if="activeTab === 'history'" :history="history ?? []" :bond="data" />
        <BondDetailsTab v-else-if="activeTab === 'details'" :bond="data" :dohod="dohod" :dohod-loading="dohodLoading" />
        <BondScoreTab v-else-if="activeTab === 'score'" :secid="secid" :scores="scores ?? null" @refresh="refreshScores" />
        <BondAiTab v-else-if="activeTab === 'ai'" :secid="secid" :bond="data" :dohod="dohod" :analyses="analyses" :job-id="currentJobId" @analysis-complete="onAnalysisComplete" @analysis-deleted="refreshAnalyses" @start-analysis="startAnalysis" />
        <BondExternalTab v-else-if="activeTab === 'external'" :secid="secid" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { IssuerRatingResponse, DohodBondData, ScoreResponse } from '~/composables/useApi'

const route = useRoute()
const api = useApi()
const favorites = useFavorites()

const secid = route.params.secid as string
const activeTab = ref('basic')
const currentJobId = ref<string | null>(null)

const tabs = [
  { id: 'basic', label: 'Основное', icon: 'bi-info-circle' },
  { id: 'trading', label: 'Торговля', icon: 'bi-bar-chart' },
  { id: 'history', label: 'История', icon: 'bi-calendar3' },
  { id: 'yields', label: 'Доходности', icon: 'bi-graph-up' },
  { id: 'coupons', label: 'Купоны', icon: 'bi-cash-stack' },
  { id: 'details', label: 'Детали', icon: 'bi-list-columns-reverse' },
  { id: 'score', label: 'Индекс', icon: 'bi-shield-shaded' },
  { id: 'ai', label: 'Разбор LLM', icon: 'bi-stars' },
  { id: 'external', label: 'Внешние', icon: 'bi-box-arrow-up-right' },
]

const { data, pending, error } = useAsyncData(`bond-${secid}`, () => api.getBond(secid))

const { data: coupons } = useAsyncData(`coupons-${secid}`, () => api.getBondCoupons(secid), { lazy: true })
const { data: history } = useAsyncData(`history-${secid}`, () => api.getBondHistory(secid), { lazy: true })

const { data: analyses, refresh: refreshAnalyses } = useAsyncData(`analyses-${secid}`, () => api.getAnalyses(secid))
const { data: analysisStats } = useAsyncData(`stats-${secid}`, () => api.getAnalysisStats(secid))

// Phase 3 — three deterministic scores (low/mid/high). Lazy so the page
// renders without waiting for the 3-profile compute; the tab + hero
// badges show a loading state until this resolves.
const { data: scores, refresh: refreshScores } = useAsyncData<ScoreResponse[]>(
  `scores-${secid}`,
  () => api.getScore(secid),
  { lazy: true },
)

// Fetch issuer credit rating by emitter_id (from all ratings map)
const issuerRating = ref<IssuerRatingResponse | null>(null)
watch(data, async (bond) => {
  if (!bond) return
  try {
    const ratings = await api.getRatings()
    if (bond.emitter_id) {
      const r = ratings[String(bond.emitter_id)]
      if (r && r.ratings?.length) issuerRating.value = r
    }
  } catch { /* no ratings */ }
}, { immediate: true })

// Auto-fetch dohod.ru data
const dohod = ref<DohodBondData | null>(null)
const dohodLoading = ref(false)
watch(data, async (bond) => {
  if (!bond?.secid) return
  dohodLoading.value = true
  try {
    const res = await api.getDohodDetails(bond.secid) as any
    if (res.job_id) {
      // Async job enqueued — poll for completion
      pollDohodJob(res.job_id, bond.secid)
    } else if (res.isin) {
      dohod.value = res as DohodBondData
      dohodLoading.value = false
    } else {
      dohodLoading.value = false
    }
  } catch {
    dohodLoading.value = false
  }
}, { immediate: true })

let dohodPollTimer: ReturnType<typeof setTimeout> | null = null
async function pollDohodJob(jobId: string, bondSecid: string) {
  try {
    const job = await api.getJobStatus(jobId)
    if (job.status === 'done') {
      const res = await api.getDohodDetails(bondSecid) as any
      if (res.isin) dohod.value = res as DohodBondData
      dohodLoading.value = false
    } else if (job.status === 'error') {
      dohodLoading.value = false
    } else {
      dohodPollTimer = setTimeout(() => pollDohodJob(jobId, bondSecid), 2000)
    }
  } catch {
    dohodLoading.value = false
  }
}
onUnmounted(() => { if (dohodPollTimer) clearTimeout(dohodPollTimer) })

async function startAnalysis() {
  try {
    const res = await api.startAnalysis(secid)
    currentJobId.value = res.job_id
    activeTab.value = 'ai'
  } catch (e: any) {
    alert('Ошибка: ' + (e.message || 'Не удалось запустить анализ'))
  }
}

function onAnalysisComplete() {
  refreshAnalyses()
  currentJobId.value = null
}

const heroRatings = computed(() =>
  (issuerRating.value?.ratings ?? [])
    .filter(r => r.rating && r.rating !== 'NULL')
    .map(r => ({ agency: r.agency, rating: r.rating }))
)

async function copyShareLink() {
  try {
    await navigator.clipboard.writeText(window.location.href)
  } catch {
    /* clipboard unavailable */
  }
}





useHead({ title: computed(() => data.value ? `${data.value.shortname} — NLA` : 'Загрузка…') })
</script>
