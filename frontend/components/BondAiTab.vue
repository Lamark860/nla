<template>
  <div class="row g-4 animate-fade-in">
    <!-- Left: main content -->
    <div class="col-lg-8">
      <!-- Header -->
      <div class="card overflow-hidden mb-4">
        <div class="panel-header">
          <i class="bi bi-stars"></i>
          Анализ облигации через ИИ
        </div>
        <div class="p-4 d-flex flex-column gap-4">
          <!-- Bond JSON data -->
          <div>
            <label class="filter-label">JSON данные для анализа (превью)</label>
            <div class="code-block">{{ bondJson }}</div>
            <div v-if="!dohod" class="form-text small text-warning">
              <i class="bi bi-hourglass-split me-1"></i>Dohod.ru данные ещё загружаются — ИИ получит их автоматически при анализе
            </div>
            <div v-else class="form-text small text-success">
              <i class="bi bi-check-circle me-1"></i>Dohod.ru данные включены (кредитный рейтинг, качество эмитента)
            </div>
          </div>

          <!-- Additional JSON -->
          <div>
            <label class="filter-label">Дополнительный JSON (опционально)</label>
            <textarea
              v-model="additionalJson"
              class="form-control font-monospace small"
              rows="3"
              :placeholder='`{"note":"Дополнительные поля для анализа"}`'
            ></textarea>
            <div class="form-text small">Поля будут объединены с основными данными. Должен содержать корректный JSON.</div>
          </div>

          <!-- Submit button -->
          <button class="btn btn-primary d-inline-flex align-items-center gap-2" @click="$emit('startAnalysis')" :disabled="!!jobId">
            <i class="bi bi-send"></i>
            Отправить в ИИ
          </button>
        </div>
      </div>

      <!-- Job polling status -->
      <div v-if="jobId && jobStatus && jobStatus.status !== 'done'" class="card p-4 mb-4" style="background: var(--nla-bg-card); color: var(--nla-text)">
        <div class="d-flex align-items-center gap-3">
          <div v-if="jobStatus.status === 'error'" class="text-danger"><i class="bi bi-x-circle fs-4"></i></div>
          <div v-else class="spinner-border" style="color: var(--nla-primary)" role="status"></div>
          <div>
            <p class="fw-semibold mb-0">
              {{ jobStatus.status === 'pending' ? 'В очереди…' : jobStatus.status === 'running' ? 'AI анализирует…' : 'Ошибка анализа' }}
            </p>
            <p class="small text-muted mb-0 mt-1">
              {{ jobStatus.status === 'error' ? jobStatus.error : 'Обычно занимает 20–60 секунд' }}
            </p>
          </div>
        </div>
      </div>

      <!-- Expanded analysis -->
      <div v-if="expandedAnalysis" class="card overflow-hidden mb-4">
        <div class="panel-header justify-content-between">
          <div class="d-flex align-items-center gap-2"><i class="bi bi-file-text"></i> Результат анализа</div>
          <button class="btn btn-sm btn-outline-secondary" @click="expandedId = null">Закрыть</button>
        </div>
        <div class="p-4">
          <div class="d-flex align-items-center gap-3 mb-3">
            <RatingBadge :rating="expandedAnalysis.rating" />
            <span class="small text-muted font-monospace">{{ fmt.date(expandedAnalysis.timestamp) }}</span>
          </div>
          <div class="ai-response" v-html="renderMarkdown(expandedAnalysis.response)"></div>
        </div>
      </div>
    </div>

    <!-- Right sidebar -->
    <div class="col-lg-4">
      <!-- History -->
      <div class="card overflow-hidden mb-4">
        <div class="panel-header">
          <i class="bi bi-clock-history"></i>
          История анализов
        </div>
        <div v-if="!analyses || analyses.length === 0" class="p-4 text-center text-muted small">
          <i class="bi bi-info-circle me-1"></i> Нет сохранённых анализов
        </div>
        <div v-else class="list-group list-group-flush">
          <button
            v-for="a in analyses"
            :key="a.id"
            class="list-group-item list-group-item-action px-4 py-3"
            @click="expandedId = expandedId === a.id ? null : a.id"
          >
            <div class="d-flex align-items-center justify-content-between">
              <div class="d-flex align-items-center gap-2">
                <RatingBadge :rating="a.rating" size="sm" />
                <span class="small text-muted font-monospace">{{ fmt.date(a.timestamp) }}</span>
              </div>
              <div class="d-flex align-items-center gap-2">
                <i class="bi bi-trash small text-muted" style="cursor: pointer" title="Удалить анализ" @click.stop="deleteAnalysis(a.id)"></i>
                <i class="bi bi-chevron-right small text-muted"></i>
              </div>
            </div>
            <p class="small text-muted mt-1 mb-0 text-truncate">{{ plainPreview(a.response) }}</p>
          </button>
        </div>
      </div>

      <!-- Stats -->
      <div class="card overflow-hidden">
        <div class="panel-header">
          <i class="bi bi-bar-chart"></i>
          Статистика
        </div>
        <div class="p-4 row g-3 text-center">
          <div class="col-6">
            <div class="h3 fw-bold font-monospace mb-0">{{ analyses?.length ?? 0 }}</div>
            <div class="small text-muted mt-1">Всего анализов</div>
          </div>
          <div class="col-6">
            <div class="h3 fw-bold font-monospace mb-0">
              <span v-if="avgRating !== '—'" class="badge" :style="avgRatingStyle">{{ avgRating }}</span>
              <span v-else>—</span>
            </div>
            <div class="small text-muted mt-1">Средний рейтинг</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond, BondAnalysis, JobStatus, DohodBondData } from '~/composables/useApi'

const props = defineProps<{
  secid: string
  bond?: Bond
  dohod?: DohodBondData | null
  analyses: BondAnalysis[] | null
  jobId: string | null
}>()

const emit = defineEmits<{ analysisComplete: []; startAnalysis: []; analysisDeleted: [] }>()

const api = useApi()
const fmt = useFormat()
const expandedId = ref<string | null>(null)
const additionalJson = ref('')

const bondJson = computed(() => {
  if (!props.bond) return '{}'
  const b = props.bond
  const data: any = {
    bond: {
      security: { SECID: b.secid, SHORTNAME: b.shortname, SECNAME: b.secname, ISIN: b.isin },
      marketdata: { LAST: b.last, BID: b.bid, OFFER: b.offer, YIELD: b.yield, DURATION: b.duration, ACCRUEDINT: b.accrued_int, SPREAD: b.spread, OPEN: b.open, HIGH: b.high, LOW: b.low, WAPRICE: b.waprice, NUMTRADES: b.numtrades, VOLUME: b.vol_today, VALTODAY: b.valtoday, BIDDEPTH: b.biddeptht, OFFERDEPTH: b.offerdeptht },
      description: { FACEVALUE: b.facevalue, MATDATE: b.matdate, COUPONPERIOD: b.coupon_period, COUPONVALUE: b.coupon_value, COUPONPERCENT: b.coupon_percent, NEXTCOUPON: b.next_coupon, TYPE: b.bond_type, CATEGORY: b.bond_category, LISTLEVEL: b.listlevel, SECTOR: b.sectorid },
      calculated: { price_rub: b.price_rub, value_today_rub: b.value_today_rub, days_to_maturity: b.days_to_maturity, is_float: b.is_float, is_indexed: b.is_indexed, coupon_display: b.coupon_display, spread_absolute: b.spread_absolute, spread_percent: b.spread_percent, current_yield: b.current_yield, modified_duration: b.modified_duration, years_to_maturity: b.years_to_maturity, risk_category: b.risk_category },
    },
  }
  if (props.dohod) {
    data.dohodDetails = {
      credit_rating: props.dohod.credit_rating,
      credit_rating_text: props.dohod.credit_rating_text,
      quality: props.dohod.quality,
      quality_outside: props.dohod.quality_outside,
      quality_inside: props.dohod.quality_inside,
      stability: props.dohod.stability,
      issuer_name: props.dohod.issuer_name,
      borrower_name: props.dohod.borrower_name,
      coupon_rate: props.dohod.coupon_rate,
      simple_yield: props.dohod.simple_yield,
      event: props.dohod.event,
      sector_text: props.dohod.sector_text,
    }
  }
  return JSON.stringify(data, null, 2)
})

const avgRating = computed(() => {
  if (!props.analyses?.length) return '—'
  const sum = props.analyses.reduce((acc, a) => acc + (a.rating || 0), 0)
  return Math.round(sum / props.analyses.length)
})

const avgRatingStyle = computed(() => {
  const base = fmt.aiRatingStyle(typeof avgRating.value === 'number' ? avgRating.value : null)
  return { ...base, fontSize: '1.2rem' }
})

const expandedAnalysis = computed(() => {
  if (!expandedId.value || !props.analyses) return null
  return props.analyses.find(a => a.id === expandedId.value) ?? null
})

const jobStatus = ref<JobStatus | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

watch(() => props.jobId, (newId) => {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (!newId) return
  jobStatus.value = { job_id: newId, type: 'ai_analysis', status: 'pending', result: null, error: '', created_at: '', finished_at: null }
  pollTimer = setInterval(async () => {
    try {
      const status = await api.getJobStatus(newId)
      jobStatus.value = status
      if (status.status === 'done' || status.status === 'error') {
        clearInterval(pollTimer!); pollTimer = null
        if (status.status === 'done') emit('analysisComplete')
      }
    } catch { /* keep polling */ }
  }, 2000)
}, { immediate: true })

onUnmounted(() => { if (pollTimer) clearInterval(pollTimer) })

function plainPreview(text: string): string { return text.replace(/[#*_`\[\]|]/g, '').trim() }

async function deleteAnalysis(id: string) {
  if (!confirm('Удалить этот анализ?')) return
  try {
    await api.deleteAnalysis(id)
    if (expandedId.value === id) expandedId.value = null
    emit('analysisDeleted')
  } catch (e: any) {
    alert('Ошибка: ' + (e.message || 'Не удалось удалить'))
  }
}

function renderMarkdown(md: string): string {
  // Process tables first (before \n→<br>)
  const lines = md.split('\n')
  const result: string[] = []
  let i = 0

  while (i < lines.length) {
    // Detect table: line with pipes
    if (/^\|(.+)\|$/.test(lines[i].trim())) {
      const tableLines: string[] = []
      while (i < lines.length && /^\|(.+)\|$/.test(lines[i].trim())) {
        tableLines.push(lines[i].trim())
        i++
      }
      if (tableLines.length >= 2) {
        result.push(renderTable(tableLines))
      } else {
        result.push(...tableLines)
      }
    } else {
      result.push(lines[i])
      i++
    }
  }

  return result.join('\n')
    .replace(/^### (.+)$/gm, '<h4 class="fw-semibold mt-3 mb-2">$1</h4>')
    .replace(/^## (.+)$/gm, '<h3 class="fw-semibold mt-4 mb-2">$1</h3>')
    .replace(/^# (.+)$/gm, '<h2 class="fw-bold mt-4 mb-3">$1</h2>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br>')
    .replace(/\[RATING:(\d+\.?\d*)\]/g, '<span class="badge bg-primary">Оценка: $1/100</span>')
}

function renderTable(lines: string[]): string {
  const parseRow = (line: string) =>
    line.split('|').slice(1, -1).map(c => c.trim())

  const headers = parseRow(lines[0])
  // Skip separator line (|---|---|)
  const startRow = /^[\s|:-]+$/.test(lines[1]) ? 2 : 1

  let html = '<div class="table-responsive my-3"><table class="table table-sm table-bordered mb-0" style="color: var(--nla-text)"><thead><tr>'
  for (const h of headers) {
    html += `<th class="small fw-semibold" style="background: var(--nla-bg); white-space: nowrap">${h}</th>`
  }
  html += '</tr></thead><tbody>'
  for (let r = startRow; r < lines.length; r++) {
    const cells = parseRow(lines[r])
    html += '<tr>'
    for (const c of cells) {
      html += `<td class="small">${c}</td>`
    }
    html += '</tr>'
  }
  html += '</tbody></table></div>'
  return html
}
</script>
