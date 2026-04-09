<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-fade-in">
    <!-- Left: main content -->
    <div class="lg:col-span-2 space-y-6">
      <!-- Header -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456z"/></svg>
          Анализ облигации через ИИ
        </div>
        <div class="p-5 space-y-5">
          <!-- Issuer ratings JSON -->
          <div v-if="issuerRatingsJson">
            <label class="filter-label">Данные эмитента</label>
            <div class="code-block">{{ issuerRatingsJson }}</div>
          </div>

          <!-- Bond JSON data -->
          <div>
            <label class="filter-label">JSON данные для анализа</label>
            <div class="code-block">{{ bondJson }}</div>
          </div>

          <!-- Additional JSON (optional) -->
          <div>
            <label class="filter-label">Дополнительный JSON (опционально)</label>
            <textarea
              v-model="additionalJson"
              class="input font-mono text-xs min-h-[80px] resize-y"
              :placeholder='`{\"note\":\"Дополнительные поля для анализа\"}`'
            />
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1.5">
              Поля будут объединены с основными данными. Должен содержать корректный JSON.
            </p>
          </div>

          <!-- Submit button -->
          <button class="btn-primary" @click="$emit('startAnalysis')" :disabled="!!jobId">
            <svg v-if="jobId" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25" /><path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="4" stroke-linecap="round" class="opacity-75" /></svg>
            <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" /></svg>
            {{ jobId ? 'Анализ идёт…' : 'Отправить в ИИ' }}
          </button>
        </div>
      </div>

      <!-- Job polling status -->
      <div v-if="jobId && jobStatus && jobStatus.status !== 'done'" class="card p-6">
        <div class="flex items-center gap-4">
          <div class="w-11 h-11 rounded-xl bg-primary-50 dark:bg-primary-500/10 flex items-center justify-center">
            <div v-if="jobStatus.status === 'error'" class="text-red-500">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
            </div>
            <div v-else class="w-5 h-5 border-2 border-primary-200 border-t-primary-600 dark:border-primary-800 dark:border-t-primary-400 rounded-full animate-spin" />
          </div>
          <div>
            <p class="font-semibold text-slate-900 dark:text-white">
              {{ jobStatus.status === 'pending' ? 'В очереди…' : jobStatus.status === 'running' ? 'AI анализирует…' : 'Ошибка анализа' }}
            </p>
            <p class="text-sm text-slate-400 dark:text-slate-500 mt-0.5">
              {{ jobStatus.status === 'error' ? jobStatus.error : 'Обычно занимает 20–60 секунд' }}
            </p>
          </div>
        </div>
      </div>

      <!-- Expanded analysis (when viewing one) -->
      <div v-if="expandedAnalysis" class="card overflow-hidden">
        <div class="panel-header justify-between">
          <div class="flex items-center gap-2.5">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"/></svg>
            Результат анализа
          </div>
          <button class="btn-ghost text-xs py-1 px-2" @click="expandedId = null">Закрыть</button>
        </div>
        <div class="p-5">
          <div class="flex items-center gap-3 mb-4">
            <RatingBadge :rating="expandedAnalysis.rating" />
            <span class="text-sm text-slate-400 dark:text-slate-500 font-mono">{{ fmt.date(expandedAnalysis.timestamp) }}</span>
          </div>
          <div
            class="prose prose-sm dark:prose-invert max-w-none text-slate-700 dark:text-slate-300 leading-relaxed"
            v-html="renderMarkdown(expandedAnalysis.response)"
          />
        </div>
      </div>
    </div>

    <!-- Right sidebar -->
    <div class="space-y-6">
      <!-- History -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          История анализов
        </div>
        <div v-if="!analyses || analyses.length === 0" class="p-6 text-center">
          <div class="flex items-center justify-center gap-2 text-sm text-slate-400 dark:text-slate-500">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><circle cx="12" cy="12" r="10" stroke-width="1.5"/><path stroke-linecap="round" stroke-width="1.5" d="M12 8v4m0 4h.01"/></svg>
            Нет сохранённых анализов
          </div>
        </div>
        <div v-else class="divide-y divide-slate-100/80 dark:divide-slate-700/20">
          <button
            v-for="a in analyses"
            :key="a.id"
            class="w-full text-left px-5 py-3 hover:bg-slate-50 dark:hover:bg-white/[0.02] transition-colors"
            @click="expandedId = expandedId === a.id ? null : a.id"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <RatingBadge :rating="a.rating" size="sm" />
                <span class="text-xs text-slate-400 dark:text-slate-500 font-mono">{{ fmt.date(a.timestamp) }}</span>
              </div>
              <svg class="w-3.5 h-3.5 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1 line-clamp-1">{{ plainPreview(a.response) }}</p>
          </button>
        </div>
      </div>

      <!-- Stats -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"/></svg>
          Статистика
        </div>
        <div class="p-5 grid grid-cols-2 gap-4 text-center">
          <div>
            <div class="text-2xl font-bold text-slate-900 dark:text-white font-mono">{{ analyses?.length ?? 0 }}</div>
            <div class="text-xs text-slate-400 dark:text-slate-500 mt-1">Всего анализов</div>
          </div>
          <div>
            <div class="text-2xl font-bold text-slate-900 dark:text-white font-mono">{{ avgRating }}</div>
            <div class="text-xs text-slate-400 dark:text-slate-500 mt-1">Средний рейтинг</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond, BondAnalysis, JobStatus } from '~/composables/useApi'

const props = defineProps<{
  secid: string
  bond?: Bond
  analyses: BondAnalysis[] | null
  jobId: string | null
}>()

const emit = defineEmits<{
  analysisComplete: []
  startAnalysis: []
}>()

const api = useApi()
const fmt = useFormat()
const expandedId = ref<string | null>(null)
const additionalJson = ref('')

// Build bond JSON for display (like original)
const bondJson = computed(() => {
  if (!props.bond) return '{}'
  const b = props.bond
  const obj = {
    bond: {
      security: {
        SECID: b.secid,
        SHORTNAME: b.shortname,
        SECNAME: b.secname,
        ISIN: b.isin,
      },
      marketdata: {
        LAST: b.last,
        BID: b.bid,
        OFFER: b.offer,
        YIELD: b.yield,
        DURATION: b.duration,
        ACCRUEDINT: b.accrued_int,
      },
      description: {
        FACEVALUE: b.facevalue,
        MATDATE: b.matdate,
        COUPONPERIOD: b.coupon_period,
        COUPONVALUE: b.coupon_value,
        COUPONPERCENT: b.coupon_percent,
        NEXTCOUPON: b.next_coupon,
        TYPE: b.bond_type,
        CATEGORY: b.bond_category,
      },
      calculated: {
        price_rub: b.price_rub,
        value_today_rub: b.value_today_rub,
        days_to_maturity: b.days_to_maturity,
        is_float: b.is_float,
        is_indexed: b.is_indexed,
        coupon_display: b.coupon_display,
      },
    },
  }
  return JSON.stringify(obj, null, 2)
})

// Issuer ratings JSON
const issuerRatingsJson = computed(() => {
  // We don't have ratings in the bond prop, show placeholder
  return null
})

// Avg rating
const avgRating = computed(() => {
  if (!props.analyses?.length) return '—'
  const sum = props.analyses.reduce((acc, a) => acc + (a.rating || 0), 0)
  return Math.round(sum / props.analyses.length)
})

// Find expanded analysis
const expandedAnalysis = computed(() => {
  if (!expandedId.value || !props.analyses) return null
  return props.analyses.find(a => a.id === expandedId.value) ?? null
})

// Job polling
const jobStatus = ref<JobStatus | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

watch(() => props.jobId, (newId) => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (!newId) return

  jobStatus.value = { job_id: newId, type: 'ai_analysis', status: 'pending', result: null, error: '', created_at: '', finished_at: null }

  pollTimer = setInterval(async () => {
    try {
      const status = await api.getJobStatus(newId)
      jobStatus.value = status
      if (status.status === 'done' || status.status === 'error') {
        clearInterval(pollTimer!)
        pollTimer = null
        if (status.status === 'done') {
          emit('analysisComplete')
        }
      }
    } catch {
      // keep polling
    }
  }, 2000)
}, { immediate: true })

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

function plainPreview(text: string): string {
  return text.replace(/[#*_`\[\]|]/g, '').trim()
}

function renderMarkdown(md: string): string {
  return md
    .replace(/^### (.+)$/gm, '<h4 class="font-semibold mt-4 mb-2">$1</h4>')
    .replace(/^## (.+)$/gm, '<h3 class="font-semibold text-lg mt-5 mb-2">$1</h3>')
    .replace(/^# (.+)$/gm, '<h2 class="font-bold text-xl mt-6 mb-3">$1</h2>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/^\|(.+)\|$/gm, (_, row) => {
      const cells = row.split('|').map((c: string) => c.trim())
      return '<tr>' + cells.map((c: string) => `<td class="px-3 py-1 border border-slate-200 dark:border-slate-600">${c}</td>`).join('') + '</tr>'
    })
    .replace(/(<tr>[\s\S]*?<\/tr>)/g, '<table class="w-full text-sm border-collapse my-3">$1</table>')
    .replace(/<td[^>]*>\s*[-:]+\s*<\/td>/g, '')
    .replace(/\n/g, '<br>')
    .replace(/\[RATING:(\d+)\]/g, '<span class="badge bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-400">Оценка: $1/100</span>')
}
</script>
