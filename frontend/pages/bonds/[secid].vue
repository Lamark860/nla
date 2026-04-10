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
      <!-- Header card -->
      <div class="mb-4">
        <div class="row g-4 align-items-stretch">
          <!-- Left: name + badges + info row -->
          <div class="col-lg-7">
            <div class="card p-3 h-100">
              <div class="d-flex align-items-center gap-2 flex-wrap mb-3">
                <h1 class="h4 fw-bold mb-0">{{ data.shortname }}</h1>
                <!-- Favorite toggle -->
                <button
                  v-if="auth.isLoggedIn.value"
                  class="btn btn-link p-1"
                  :class="favorites.isFavorite(data.secid) ? 'text-warning' : 'text-muted'"
                  :title="favorites.isFavorite(data.secid) ? 'Убрать из избранного' : 'В избранное'"
                  @click="favorites.toggle(data.secid)"
                >
                  <i :class="favorites.isFavorite(data.secid) ? 'bi-star-fill' : 'bi-star'" class="bi fs-5"></i>
                </button>
                <span v-if="data.trading_status === 'T'" class="badge bg-success">{{ data.boardname || 'Торги идут' }}</span>
                <span v-if="data.trading_status === 'N'" class="badge bg-secondary">Торги не ведутся</span>
                <span class="badge bg-light text-dark border font-monospace">{{ data.boardid || 'TQCB' }}</span>
              </div>
              <!-- Info row: ISIN | Код | Тип | Валюта | Ratings -->
              <div class="d-flex align-items-center gap-4 flex-wrap small" style="color: var(--nla-text)">
                <div>
                  <span class="text-muted">ISIN</span>
                  <span class="ms-1 fw-semibold font-monospace">{{ data.isin }}</span>
                </div>
                <div>
                  <span class="text-muted">Код</span>
                  <span class="ms-1 fw-semibold font-monospace">{{ data.secid }}</span>
                </div>
                <div v-if="data.regnumber">
                  <span class="text-muted">Рег. №</span>
                  <span class="ms-1 fw-semibold font-monospace">{{ data.regnumber }}</span>
                </div>
                <div>
                  <span class="text-muted">Тип</span>
                  <span class="ms-1 fw-semibold">{{ data.bond_category }}</span>
                </div>
                <div>
                  <span class="text-muted">Валюта</span>
                  <span class="ms-1 fw-semibold">{{ data.currencyid === 'SUR' ? 'RUB' : data.currencyid || 'RUB' }}</span>
                </div>
                <span
                  v-if="analysisStats?.avg_rating"
                  class="badge fw-semibold"
                  :style="fmt.aiRatingStyleSoft(analysisStats.avg_rating)"
                  :title="`AI: средний балл ${analysisStats.avg_rating.toFixed(1)}, анализов ${analysisStats.total}`"
                >🤖 {{ Math.round(analysisStats.avg_rating) }}</span>
                <!-- Inline credit ratings -->
                <template v-if="issuerRating && issuerRating.ratings.length">
                  <span
                    v-for="r in issuerRating.ratings.filter(x => x.rating && x.rating !== 'NULL')"
                    :key="r.agency"
                    class="badge font-monospace"
                    :style="fmt.ratingChipStyle(r.rating)"
                    style="font-size: 11px; padding: 3px 7px"
                  ><span style="font-weight:700">{{ r.rating }}</span> <span style="font-weight:400;opacity:0.7;font-family:var(--bs-body-font-family)">{{ r.agency }}</span></span>
                </template>
                <span v-else class="badge" style="font-size: 10px; background: rgba(108,117,125,0.08); color: var(--nla-text-muted)">Рейтинг не присвоен</span>
              </div>
            </div>
          </div>
          <!-- Right: price card -->
          <div class="col-lg-5">
            <div class="card p-3 h-100" style="border-left: 3px solid var(--nla-primary)">
              <div class="d-flex justify-content-between align-items-start">
                <div>
                  <div class="small text-muted mb-1">Текущая цена</div>
                  <div class="h4 fw-bold font-monospace mb-0">{{ data.last != null ? fmt.percent(data.last) : '—' }}</div>
                  <div class="small font-monospace text-muted">{{ fmt.priceRub(data.price_rub) }}</div>
                  <div v-if="data.last_change_prcnt != null" :class="data.last_change_prcnt >= 0 ? 'text-success' : 'text-danger'" class="small fw-semibold mt-1">
                    {{ data.last_change_prcnt >= 0 ? '+' : '' }}{{ data.last_change != null ? data.last_change.toFixed(2) : '0' }} ({{ data.last_change_prcnt >= 0 ? '+' : '' }}{{ data.last_change_prcnt.toFixed(2) }}%)
                  </div>
                </div>
                <div class="text-end">
                  <div class="small text-muted mb-1">Доходность</div>
                  <div class="h5 fw-bold text-positive font-monospace mb-0">{{ fmt.percent(data.yield) }}</div>
                  <div class="small text-muted mt-1">до {{ fmt.date(data.matdate) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

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
        <BondBasicTab v-if="activeTab === 'basic'" :bond="data" />
        <BondTradingTab v-else-if="activeTab === 'trading'" :bond="data" />
        <BondCouponsTab v-else-if="activeTab === 'coupons'" :coupons="coupons ?? []" :bond="data" />
        <BondYieldsTab v-else-if="activeTab === 'yields'" :bond="data" />
        <BondHistoryTab v-else-if="activeTab === 'history'" :history="history ?? []" :bond="data" />
        <BondDetailsTab v-else-if="activeTab === 'details'" :bond="data" :dohod="dohod" :dohod-loading="dohodLoading" />
        <BondAiTab v-else-if="activeTab === 'ai'" :secid="secid" :bond="data" :dohod="dohod" :analyses="analyses" :job-id="currentJobId" @analysis-complete="onAnalysisComplete" @analysis-deleted="refreshAnalyses" @start-analysis="startAnalysis" />
        <BondExternalTab v-else-if="activeTab === 'external'" :secid="secid" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { IssuerRatingResponse, DohodBondData } from '~/composables/useApi'

const route = useRoute()
const api = useApi()
const fmt = useFormat()
const auth = useAuth()
const favorites = useFavorites()

const secid = route.params.secid as string
const activeTab = ref('basic')
const analyzing = ref(false)
const currentJobId = ref<string | null>(null)

const tabs = [
  { id: 'basic', label: 'Основное', icon: 'bi-info-circle' },
  { id: 'trading', label: 'Торговля', icon: 'bi-bar-chart' },
  { id: 'history', label: 'История', icon: 'bi-calendar3' },
  { id: 'yields', label: 'Доходности', icon: 'bi-graph-up' },
  { id: 'coupons', label: 'Купоны', icon: 'bi-cash-stack' },
  { id: 'details', label: 'Детали', icon: 'bi-list-columns-reverse' },
  { id: 'ai', label: 'AI Анализ', icon: 'bi-stars' },
  { id: 'external', label: 'Внешние', icon: 'bi-box-arrow-up-right' },
]

const { data, pending, error } = useAsyncData(`bond-${secid}`, () => api.getBond(secid))

const { data: coupons } = useAsyncData(`coupons-${secid}`, () => api.getBondCoupons(secid), { lazy: true })
const { data: history } = useAsyncData(`history-${secid}`, () => api.getBondHistory(secid), { lazy: true })

const { data: analyses, refresh: refreshAnalyses } = useAsyncData(`analyses-${secid}`, () => api.getAnalyses(secid))
const { data: analysisStats } = useAsyncData(`stats-${secid}`, () => api.getAnalysisStats(secid))

// Fetch issuer credit rating by emitter_id (from all ratings map)
const issuerRating = ref<IssuerRatingResponse | null>(null)
const allRatings = ref<Record<string, IssuerRatingResponse>>({})
watch(data, async (bond) => {
  if (!bond) return
  try {
    const ratings = await api.getRatings()
    allRatings.value = ratings
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
  analyzing.value = true
  try {
    const res = await api.startAnalysis(secid)
    currentJobId.value = res.job_id
    activeTab.value = 'ai'
  } catch (e: any) {
    alert('Ошибка: ' + (e.message || 'Не удалось запустить анализ'))
  } finally {
    analyzing.value = false
  }
}

function onAnalysisComplete() {
  refreshAnalyses()
  currentJobId.value = null
}





useHead({ title: computed(() => data.value ? `${data.value.shortname} — NLA` : 'Загрузка…') })
</script>
