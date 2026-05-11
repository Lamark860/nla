<template>
  <div class="ai-grid animate-fade-in">
    <!-- Left column: form + result -->
    <div class="ai-col">
      <Panel title="Аналитический индекс" icon="stars">
        <div class="ai-form">
          <div class="ai-field">
            <label class="ai-field__lbl">JSON данные для анализа · превью</label>
            <pre class="ai-code">{{ bondJson }}</pre>
            <div v-if="!dohod" class="ai-hint ai-hint--warn">
              <i class="bi bi-hourglass-split"></i>
              Dohod.ru данные ещё загружаются — будут учтены в расчёте автоматически
            </div>
            <div v-else class="ai-hint ai-hint--ok">
              <i class="bi bi-check-circle"></i>
              Dohod.ru данные включены · кредитный рейтинг, качество эмитента
            </div>
          </div>

          <div class="ai-field">
            <label class="ai-field__lbl">Дополнительный JSON · опционально</label>
            <textarea
              v-model="additionalJson"
              class="ai-textarea"
              rows="3"
              :placeholder='`{&quot;note&quot;:&quot;Дополнительные поля для анализа&quot;}`'
            ></textarea>
            <div class="ai-hint">Поля будут объединены с основными данными. Должен содержать корректный JSON.</div>
          </div>

          <button class="ai-submit" :disabled="!!jobId" @click="$emit('startAnalysis')">
            <i class="bi bi-send-fill"></i>
            <span>{{ jobId ? 'Идёт анализ…' : 'Рассчитать индекс' }}</span>
          </button>
        </div>
      </Panel>

      <!-- Job polling status -->
      <Panel v-if="jobId && jobStatus && jobStatus.status !== 'done'">
        <div class="ai-job">
          <div v-if="jobStatus.status === 'error'" class="ai-job__icon ai-job__icon--err">
            <i class="bi bi-x-circle-fill"></i>
          </div>
          <div v-else class="ai-job__spinner"><div class="ai-job__spinner-ring"></div></div>
          <div class="ai-job__text">
            <div class="ai-job__title">
              {{ jobStatus.status === 'pending' ? 'В очереди…' : jobStatus.status === 'running' ? 'Считаем индекс…' : 'Ошибка расчёта' }}
            </div>
            <div class="ai-job__sub">
              {{ jobStatus.status === 'error' ? jobStatus.error : 'Обычно занимает 20–60 секунд' }}
            </div>
          </div>
        </div>
      </Panel>

      <!-- Expanded analysis result -->
      <Panel v-if="expandedAnalysis" flush>
        <template #head>
          <div class="ai-result-head">
            <AiScore v-if="expandedAnalysis.rating != null" :value="expandedAnalysis.rating" />
            <span v-else class="ai-no-score">— / 100</span>
            <span class="ai-result-time">{{ fmt.date(expandedAnalysis.timestamp) }}</span>
            <div class="ai-result-actions">
              <button class="ai-icon-btn" title="Копировать" @click="copyResult(expandedAnalysis)">
                <i class="bi bi-clipboard"></i>
              </button>
              <button class="ai-icon-btn" title="Скачать .md" @click="downloadResult(expandedAnalysis)">
                <i class="bi bi-download"></i>
              </button>
              <button class="ai-icon-btn" title="Закрыть" @click="expandedId = null">
                <i class="bi bi-x-lg"></i>
              </button>
            </div>
          </div>
        </template>
        <div class="ai-result-body" v-html="renderMarkdown(expandedAnalysis.response)"></div>
      </Panel>
    </div>

    <!-- Right column: history + stats -->
    <div class="ai-col">
      <Panel title="История анализов" icon="clock-history" :meta="String(analyses?.length ?? 0)" flush>
        <div v-if="!analyses || analyses.length === 0" class="ai-empty">
          <i class="bi bi-info-circle"></i>
          <span>Нет сохранённых анализов</span>
        </div>
        <ul v-else class="ai-hist">
          <li
            v-for="a in analyses"
            :key="a.id"
            class="ai-hist-item"
            :class="{ active: expandedId === a.id }"
            @click="expandedId = expandedId === a.id ? null : a.id"
          >
            <div class="ai-hist__row">
              <div class="ai-hist__meta">
                <AiScore v-if="a.rating != null" :value="a.rating" compact />
                <span v-else class="ai-no-score-sm">— / 100</span>
                <span class="ai-hist__time">{{ fmt.date(a.timestamp) }}</span>
              </div>
              <div class="ai-hist__actions">
                <i
                  class="bi bi-trash3"
                  title="Удалить"
                  @click.stop="deleteAnalysis(a.id)"
                ></i>
                <i class="bi bi-chevron-right"></i>
              </div>
            </div>
            <p class="ai-hist__preview">{{ plainPreview(a.response) }}</p>
          </li>
        </ul>
      </Panel>

      <Panel title="Статистика" icon="bar-chart">
        <div class="ai-stats">
          <div class="ai-stat-cell">
            <div class="ai-stat-val">{{ analyses?.length ?? 0 }}</div>
            <div class="ai-stat-lbl">Всего анализов</div>
          </div>
          <div class="ai-stat-cell">
            <div class="ai-stat-val">
              <AiScore v-if="typeof avgRating === 'number'" :value="avgRating" />
              <span v-else>—</span>
            </div>
            <div class="ai-stat-lbl">Средний рейтинг</div>
          </div>
        </div>
      </Panel>
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

const avgRating = computed<number | '—'>(() => {
  if (!props.analyses?.length) return '—'
  const rated = props.analyses.filter(a => a.rating != null) as Array<{ rating: number }>
  if (!rated.length) return '—'
  const sum = rated.reduce((acc, a) => acc + a.rating, 0)
  return Math.round(sum / rated.length)
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

function plainPreview(text: string): string {
  return text.replace(/[#*_`\[\]|]/g, '').replace(/\s+/g, ' ').trim().slice(0, 160)
}

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

async function copyResult(a: BondAnalysis) {
  try {
    await navigator.clipboard.writeText(a.response)
  } catch {
    /* no-op — clipboard API not available */
  }
}

function downloadResult(a: BondAnalysis) {
  const blob = new Blob([a.response], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  const date = a.timestamp ? new Date(a.timestamp).toISOString().slice(0, 10) : 'analysis'
  link.download = `${props.secid}-${date}.md`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function renderMarkdown(md: string): string {
  // Process tables first (before \n→<br>)
  const lines = md.split('\n')
  const result: string[] = []
  let i = 0

  while (i < lines.length) {
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

<style scoped>
.ai-grid {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 16px;
  align-items: start;
}
.ai-col { display: flex; flex-direction: column; gap: 16px; min-width: 0; }

@media (max-width: 992px) {
  .ai-grid { grid-template-columns: 1fr; }
}

/* Form */
.ai-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 16px 18px;
}
.ai-field { display: flex; flex-direction: column; gap: 6px; }
.ai-field__lbl {
  font: 600 11px/1.2 var(--nla-font);
  color: var(--nla-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.ai-code {
  margin: 0;
  padding: 12px 14px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border-light);
  border-radius: var(--nla-radius);
  font: 500 11.5px/1.55 var(--nla-font-mono);
  color: var(--nla-text-secondary);
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
.ai-textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 12.5px/1.5 var(--nla-font-mono);
  color: var(--nla-text);
  resize: vertical;
  min-height: 60px;
}
.ai-textarea:focus { outline: none; border-color: var(--nla-primary); box-shadow: 0 0 0 2px var(--nla-primary-light); }
.ai-hint {
  font: 500 11.5px/1.4 var(--nla-font);
  color: var(--nla-text-muted);
  display: flex;
  align-items: center;
  gap: 6px;
}
.ai-hint--ok { color: var(--nla-success); }
.ai-hint--warn { color: var(--nla-warning); }
.ai-hint .bi { font-size: 12px; }

.ai-submit {
  appearance: none;
  border: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 18px;
  background: var(--nla-primary);
  color: #fff;
  border-radius: var(--nla-radius);
  font: 600 13px/1 var(--nla-font);
  letter-spacing: 0.01em;
  cursor: pointer;
  transition: background 120ms ease;
  align-self: flex-start;
}
.ai-submit:hover { background: var(--nla-primary-ink); }
.ai-submit:disabled { background: var(--nla-text-muted); cursor: not-allowed; opacity: 0.6; }

/* Job polling */
.ai-job {
  padding: 16px 18px;
  display: flex;
  align-items: center;
  gap: 14px;
}
.ai-job__icon { font-size: 24px; }
.ai-job__icon--err { color: var(--nla-danger); }
.ai-job__spinner {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
}
.ai-job__spinner-ring {
  width: 24px;
  height: 24px;
  border: 2.5px solid var(--nla-border);
  border-top-color: var(--nla-primary);
  border-radius: 50%;
  animation: aiSpin 0.8s linear infinite;
}
@keyframes aiSpin { to { transform: rotate(360deg); } }
.ai-job__title { font: 600 13px/1.4 var(--nla-font); color: var(--nla-text); }
.ai-job__sub { font: 500 11.5px/1.4 var(--nla-font); color: var(--nla-text-muted); margin-top: 2px; }

/* Result panel */
.ai-result-head {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}
.ai-result-time {
  font: 500 11.5px/1 var(--nla-font-mono);
  color: var(--nla-text-muted);
}
.ai-no-score {
  font: 600 14px/1 var(--nla-font-mono);
  color: var(--nla-text-muted);
  padding: 4px 10px;
  background: var(--nla-bg-subtle);
  border-radius: var(--nla-radius-sm);
}
.ai-result-actions { margin-left: auto; display: flex; gap: 4px; }
.ai-icon-btn {
  appearance: none;
  border: 0;
  background: transparent;
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  border-radius: var(--nla-radius-sm);
  color: var(--nla-text-muted);
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
}
.ai-icon-btn:hover { background: var(--nla-bg-subtle); color: var(--nla-text); }
.ai-result-body {
  padding: 18px 22px 22px;
  font: 400 13.5px/1.65 var(--nla-font);
  color: var(--nla-text);
}
.ai-result-body :deep(h2),
.ai-result-body :deep(h3),
.ai-result-body :deep(h4) { letter-spacing: -0.01em; }

/* History list */
.ai-empty {
  padding: 40px 18px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: var(--nla-text-muted);
  font: 500 12.5px/1.4 var(--nla-font);
}
.ai-empty .bi { font-size: 22px; opacity: 0.5; }

.ai-hist {
  list-style: none;
  margin: 0;
  padding: 0;
}
.ai-hist-item {
  padding: 12px 18px;
  border-top: 1px solid var(--nla-border-light);
  cursor: pointer;
  transition: background 120ms ease;
}
.ai-hist-item:first-child { border-top: 0; }
.ai-hist-item:hover { background: var(--nla-bg-subtle); }
.ai-hist-item.active { background: var(--nla-primary-light); }

.ai-hist__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.ai-hist__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.ai-hist__time {
  font: 500 11px/1 var(--nla-font-mono);
  color: var(--nla-text-muted);
  white-space: nowrap;
}
.ai-hist__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--nla-text-muted);
  font-size: 13px;
  flex-shrink: 0;
}
.ai-hist__actions .bi-trash3 { cursor: pointer; transition: color 120ms ease; }
.ai-hist__actions .bi-trash3:hover { color: var(--nla-danger); }
.ai-hist__preview {
  font: 500 11.5px/1.4 var(--nla-font);
  color: var(--nla-text-muted);
  margin: 6px 0 0 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ai-no-score-sm {
  font: 600 11px/1 var(--nla-font-mono);
  color: var(--nla-text-muted);
  padding: 2px 6px;
  background: var(--nla-bg-subtle);
  border-radius: 4px;
}

/* Stats */
.ai-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
}
.ai-stat-cell {
  padding: 18px 14px;
  text-align: center;
  border-right: 1px solid var(--nla-border-light);
}
.ai-stat-cell:last-child { border-right: 0; }
.ai-stat-val {
  font: 700 24px/1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 32px;
}
.ai-stat-lbl {
  font: 600 10.5px/1.2 var(--nla-font);
  color: var(--nla-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-top: 6px;
}
</style>
