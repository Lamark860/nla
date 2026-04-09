<template>
  <div>
    <!-- Back link -->
    <NuxtLink to="/" class="inline-flex items-center gap-1.5 text-sm text-slate-400 hover:text-primary-600 dark:hover:text-primary-400 mb-5 transition-colors">
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
      Все облигации
    </NuxtLink>

    <!-- Loading -->
    <div v-if="pending" class="card p-16 text-center">
      <div class="inline-block w-6 h-6 border-2 border-primary-200 border-t-primary-600 dark:border-primary-800 dark:border-t-primary-400 rounded-full animate-spin"></div>
      <p class="mt-4 text-xs text-slate-400 dark:text-slate-500">Загрузка данных…</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="card p-10 text-center">
      <p class="text-red-600 dark:text-red-400 text-sm">{{ error.message || 'Облигация не найдена' }}</p>
      <NuxtLink to="/" class="btn-primary mt-4 inline-block text-sm">Назад к списку</NuxtLink>
    </div>

    <!-- Bond data -->
    <template v-else-if="data">
      <!-- Header card -->
      <div class="mb-6">
        <div class="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-6">
          <!-- Left: name + badges + info row -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-3 flex-wrap">
              <h1 class="text-2xl font-bold text-slate-900 dark:text-white">{{ data.shortname }}</h1>
              <!-- Favorite toggle -->
              <button
                v-if="auth.isLoggedIn.value"
                class="p-1.5 rounded-lg transition-colors"
                :class="favorites.isFavorite(data.secid)
                  ? 'text-amber-500 hover:text-amber-600'
                  : 'text-slate-300 hover:text-amber-400 dark:text-slate-600 dark:hover:text-amber-400'"
                :title="favorites.isFavorite(data.secid) ? 'Убрать из избранного' : 'В избранное'"
                @click="favorites.toggle(data.secid)"
              >
                <svg class="w-5 h-5" :fill="favorites.isFavorite(data.secid) ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                </svg>
              </button>
              <span v-if="data.trading_status === 'T'" class="badge bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">{{ data.boardname || 'Торги идут' }}</span>
              <span v-if="data.trading_status === 'N'" class="badge bg-slate-500/15 text-slate-500 dark:text-slate-400">Торги не ведутся</span>
              <span class="badge bg-slate-500/15 text-slate-500 dark:text-slate-400 font-mono">{{ data.boardid || 'TQCB' }}</span>
            </div>
            <!-- Info row: ISIN | Код | Тип | Валюта -->
            <div class="flex items-center gap-6 mt-3 flex-wrap text-sm">
              <div>
                <span style="color: var(--nla-text-muted)">ISIN</span>
                <span class="ml-1.5 font-semibold font-mono" style="color: var(--nla-text-secondary)">{{ data.isin }}</span>
              </div>
              <div>
                <span style="color: var(--nla-text-muted)">Код</span>
                <span class="ml-1.5 font-semibold font-mono" style="color: var(--nla-text-secondary)">{{ data.secid }}</span>
              </div>
              <div v-if="data.regnumber">
                <span style="color: var(--nla-text-muted)">Рег. №</span>
                <span class="ml-1.5 font-semibold font-mono" style="color: var(--nla-text-secondary)">{{ data.regnumber }}</span>
              </div>
              <div>
                <span style="color: var(--nla-text-muted)">Тип</span>
                <span class="ml-1.5 font-semibold" style="color: var(--nla-text-secondary)">{{ data.bond_category }}</span>
              </div>
              <div>
                <span style="color: var(--nla-text-muted)">Валюта</span>
                <span class="ml-1.5 font-semibold" style="color: var(--nla-text-secondary)">{{ data.currencyid === 'SUR' ? 'RUB' : data.currencyid || 'RUB' }}</span>
              </div>
              <span
                v-if="issuerRating"
                :class="creditScoreClass(issuerRating.score)"
                class="badge font-mono font-bold"
                :title="issuerRating.ratings.map(r => `${r.agency}: ${r.rating}`).join(', ')"
              >{{ issuerRating.ratings[0]?.rating }}</span>
            </div>
          </div>
          <!-- Right: price card -->
          <div class="card p-5 shrink-0 lg:min-w-[320px] border-l-[3px] border-l-primary-500">
            <div class="flex justify-between items-start">
              <div>
                <div class="text-xs mb-1" style="color: var(--nla-text-muted)">Текущая цена</div>
                <div class="text-3xl font-bold font-mono" style="color: var(--nla-text)">{{ data.last != null ? fmt.percent(data.last) : '—' }}</div>
                <div class="text-sm font-mono mt-0.5" style="color: var(--nla-text-muted)">{{ fmt.priceRub(data.price_rub) }}</div>
                <div v-if="data.last_change_prcnt != null" :class="data.last_change_prcnt >= 0 ? 'text-emerald-500' : 'text-red-500'" class="text-sm font-semibold mt-1">
                  {{ data.last_change_prcnt >= 0 ? '+' : '' }}{{ data.last_change != null ? data.last_change.toFixed(2) : '0' }} ({{ data.last_change_prcnt >= 0 ? '+' : '' }}{{ data.last_change_prcnt.toFixed(2) }}%)
                </div>
              </div>
              <div class="text-right">
                <div class="text-xs mb-1" style="color: var(--nla-text-muted)">Доходность</div>
                <div class="text-2xl font-bold text-positive font-mono">{{ fmt.percent(data.yield) }}</div>
                <div class="text-xs mt-1" style="color: var(--nla-text-muted)">до {{ fmt.date(data.matdate) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tabs (pill-style, ASH-matched) -->
      <nav class="bond-tabs mb-6">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          :class="activeTab === tab.id ? 'bond-tab bond-tab--active' : 'bond-tab'"
          @click="activeTab = tab.id"
        >
          <span class="bond-tab__icon" v-html="tab.icon" />
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
        <BondAiTab v-else-if="activeTab === 'ai'" :secid="secid" :bond="data" :analyses="analyses" :job-id="currentJobId" @analysis-complete="onAnalysisComplete" @start-analysis="startAnalysis" />
        <BondExternalTab v-else-if="activeTab === 'external'" :secid="secid" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { IssuerRatingResponse } from '~/composables/useApi'

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
  { id: 'basic', label: 'Основное', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><circle cx="12" cy="12" r="10" stroke-width="1.5"/><path stroke-linecap="round" stroke-width="1.5" d="M12 8v4m0 4h.01"/></svg>' },
  { id: 'trading', label: 'Торговля', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"/></svg>' },
  { id: 'history', label: 'История', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"/></svg>' },
  { id: 'yields', label: 'Доходности', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>' },
  { id: 'coupons', label: 'Купоны', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"/></svg>' },
  { id: 'ai', label: 'AI Анализ', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z"/></svg>' },
  { id: 'external', label: 'Внешние', icon: '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25"/></svg>' },
]

const { data, pending, error } = useAsyncData(`bond-${secid}`, () => api.getBond(secid))

const { data: coupons } = useAsyncData(`coupons-${secid}`, () => api.getBondCoupons(secid), { lazy: true })
const { data: history } = useAsyncData(`history-${secid}`, () => api.getBondHistory(secid), { lazy: true })

const { data: analyses, refresh: refreshAnalyses } = useAsyncData(`analyses-${secid}`, () => api.getAnalyses(secid))
const { data: analysisStats } = useAsyncData(`stats-${secid}`, () => api.getAnalysisStats(secid))

// Fetch issuer credit rating based on bond name
const issuerRating = ref<IssuerRatingResponse | null>(null)
watch(data, async (bond) => {
  if (!bond) return
  // Extract issuer name from secname
  const issuerName = (bond.secname || bond.shortname).split(/\s+(БО|ПБО|ПАО|ООО|АО|НАО|001P|002P)/)[0].trim()
  try {
    const r = await api.getIssuerRating(issuerName)
    if (r && r.ratings?.length) issuerRating.value = r
  } catch { /* no rating found */ }
}, { immediate: true })

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

function categoryBadge(cat: string): string {
  switch (cat) {
    case 'ОФЗ': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400'
    case 'Корпоративная': return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
    case 'Субфедеральная': return 'bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-400'
    case 'Муниципальная': return 'bg-orange-50 text-orange-700 dark:bg-orange-500/10 dark:text-orange-400'
    default: return 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400'
  }
}

function creditScoreClass(score: number): string {
  if (score >= 9) return 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 ring-1 ring-emerald-500/20'
  if (score >= 7) return 'bg-blue-500/15 text-blue-700 dark:text-blue-400 ring-1 ring-blue-500/20'
  if (score >= 5) return 'bg-amber-500/15 text-amber-700 dark:text-amber-400 ring-1 ring-amber-500/20'
  if (score >= 3) return 'bg-orange-500/15 text-orange-700 dark:text-orange-400 ring-1 ring-orange-500/20'
  return 'bg-red-500/15 text-red-700 dark:text-red-400 ring-1 ring-red-500/20'
}

useHead({ title: computed(() => data.value ? `${data.value.shortname} — NLA` : 'Загрузка…') })
</script>
