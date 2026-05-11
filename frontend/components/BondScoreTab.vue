<!--
  BondScoreTab.vue — таб «Индекс» в карточке бумаги.

  Источник: GET /api/v1/bonds/{secid}/score → массив из трёх ScoreResponse
  (low/mid/high). Один пресет помечен «активным» через useScoringProfile;
  активный балл крупно сверху, два других — компактно для сравнения.
  Ниже разворачиваемая таблица 12 факторов активного профиля.

  «Получить разбор» → POST /score/explain → polling /jobs/{id} → подтягивает
  обновлённый ScoreResponse, в котором уже лежит explanation.text.
-->
<template>
  <section class="bond-score">
    <!-- Top: three scores side by side -->
    <Panel class="bond-score__head" flush>
      <template #head>
        <i class="bi bi-shield-shaded panel-head__icon"></i>
        <span class="panel-head__title">Аналитический индекс</span>
        <span class="panel-head__meta" v-if="active">обновлено {{ fmt.dateTime(active.computed_at) }}</span>
      </template>

      <div class="bond-score__profiles">
        <button
          v-for="r in sortedResults"
          :key="r.profile_code"
          class="profile-card"
          :class="{ 'is-active': r.profile_code === scoring.profile.value }"
          @click="scoring.set(r.profile_code as ProfileCode)"
        >
          <span class="profile-card__icon" aria-hidden="true">{{ scoring.metaMap[r.profile_code as ProfileCode]?.icon || '•' }}</span>
          <div class="profile-card__body">
            <div class="profile-card__lbl">{{ r.profile_name || scoring.metaMap[r.profile_code as ProfileCode]?.label || r.profile_code }}</div>
            <div class="profile-card__val" :style="badgeStyle(r.score)">{{ Math.round(r.score) }}</div>
          </div>
        </button>
      </div>

      <div v-if="active?.missing_factors?.length" class="bond-score__missing">
        <i class="bi bi-info-circle" aria-hidden="true"></i>
        Данных не хватило по факторам: <strong>{{ missingLabels.join(', ') }}</strong>.
        Этим факторам присвоено нейтральное значение, остальные посчитаны как обычно.
      </div>
    </Panel>

    <!-- Breakdown: 12 factor rows for the active profile -->
    <Panel v-if="active" class="bond-score__breakdown" flush>
      <template #head>
        <i class="bi bi-list-ol panel-head__icon"></i>
        <span class="panel-head__title">Разбор по факторам · {{ active.profile_name || active.profile_code }}</span>
        <button class="bond-score__toggle" @click="breakdownOpen = !breakdownOpen">
          {{ breakdownOpen ? 'Свернуть' : 'Развернуть' }}
          <i :class="breakdownOpen ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
        </button>
      </template>

      <table v-show="breakdownOpen" class="breakdown-table">
        <thead>
          <tr>
            <th class="t-factor">Фактор</th>
            <th class="t-raw">Сырое</th>
            <th class="t-norm">0..100</th>
            <th class="t-weight">Вес</th>
            <th class="t-contrib">Вклад</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in active.breakdown" :key="row.factor" :class="{ 'is-missing': !row.has_data }">
            <td class="t-factor">
              <span class="factor-name">{{ row.name }}</span>
              <span v-if="!row.has_data" class="factor-flag">нет данных</span>
            </td>
            <td class="t-raw">{{ formatRaw(row) }}</td>
            <td class="t-norm">
              <div class="norm-bar">
                <div class="norm-bar__fill" :style="{ width: clampPct(row.normalized) + '%' }"></div>
                <span class="norm-bar__val">{{ Math.round(row.normalized) }}</span>
              </div>
            </td>
            <td class="t-weight" :class="weightClass(row.weight)">{{ formatWeight(row.weight) }}</td>
            <td class="t-contrib" :class="contribClass(row.contribution)">{{ formatContrib(row.contribution) }}</td>
          </tr>
        </tbody>
        <tfoot>
          <tr>
            <td colspan="4" class="t-foot">Итого</td>
            <td class="t-contrib t-foot">{{ active.score.toFixed(1) }}</td>
          </tr>
        </tfoot>
      </table>
    </Panel>

    <!-- Explanation: cached or async-generated LLM text -->
    <Panel v-if="active" class="bond-score__explain" flush>
      <template #head>
        <i class="bi bi-chat-square-text panel-head__icon"></i>
        <span class="panel-head__title">Разбор простым языком</span>
        <button
          class="bond-score__explain-btn"
          :disabled="explainLoading"
          @click="requestExplain"
        >
          <i v-if="explainLoading" class="bi bi-arrow-clockwise spin"></i>
          <i v-else class="bi bi-stars"></i>
          {{ explainBtnLabel }}
        </button>
      </template>

      <div class="explain-body">
        <div v-if="active.explanation" class="explain-text">{{ active.explanation.text }}</div>
        <div v-else-if="explainLoading" class="explain-empty">
          <i class="bi bi-arrow-clockwise spin"></i>
          Запрос отправлен. LLM обычно отвечает 5–15 секунд.
        </div>
        <div v-else class="explain-empty">
          Нажмите «Получить разбор», чтобы получить текстовое объяснение балла.
          <br>Источник — детерминированный скоринг-движок (12 факторов), LLM только пересказывает.
        </div>
        <div v-if="explainError" class="explain-error">
          <i class="bi bi-exclamation-triangle"></i>
          {{ explainError }}
        </div>
      </div>
    </Panel>
  </section>
</template>

<script setup lang="ts">
import type { ScoreResponse, BreakdownItem } from '~/composables/useApi'
import type { ProfileCode } from '~/composables/useScoringProfile'
import Panel from './Panel.vue'

const props = defineProps<{
  secid: string
  scores: ScoreResponse[] | null
}>()
const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const fmt = useFormat()
const api = useApi()
const scoring = useScoringProfile()

const breakdownOpen = ref(true)
const explainLoading = ref(false)
const explainError = ref<string | null>(null)

// Stable display order regardless of what the API returns (defensive — the
// backend already orders low/mid/high, but tests + future user profiles
// could reorder).
const profileOrder: ProfileCode[] = ['low', 'mid', 'high']
const sortedResults = computed(() => {
  if (!props.scores) return []
  return [...props.scores].sort((a, b) =>
    profileOrder.indexOf(a.profile_code as ProfileCode) - profileOrder.indexOf(b.profile_code as ProfileCode))
})

const active = computed<ScoreResponse | null>(() => {
  if (!props.scores) return null
  return props.scores.find(r => r.profile_code === scoring.profile.value)
    || props.scores[0]
    || null
})

const missingLabels = computed(() => {
  if (!active.value?.missing_factors || !active.value.breakdown) return []
  const byCode = new Map(active.value.breakdown.map(b => [b.factor, b.name]))
  return active.value.missing_factors.map(c => byCode.get(c) || c)
})

const explainBtnLabel = computed(() => {
  if (explainLoading.value) return 'Готовится…'
  if (active.value?.explanation) return 'Обновить разбор'
  return 'Получить разбор'
})

function badgeStyle(score: number) {
  // Reuse the AI-score colour ramp from useFormat — same 0..100 semantics.
  return fmt.aiRatingStyleSoft(score)
}

function clampPct(v: number): number {
  if (!Number.isFinite(v)) return 0
  if (v < 0) return 0
  if (v > 100) return 100
  return v
}

function formatRaw(row: BreakdownItem): string {
  if (!row.has_data || row.raw == null) return '—'
  // Units differ per factor — keep it short, the table is a quick glance.
  // YTM/coupon/price-style factors fall in the ones-to-tens range; size /
  // liquidity hit hundreds of millions. Heuristic: huge values → SI suffix.
  if (Math.abs(row.raw) >= 1_000_000) {
    return (row.raw / 1_000_000).toFixed(row.raw >= 100_000_000 ? 0 : 1) + ' млн'
  }
  if (Math.abs(row.raw) >= 1_000) {
    return (row.raw / 1_000).toFixed(0) + ' тыс'
  }
  // Small numbers: 2 decimals for fractional, integer otherwise.
  return Number.isInteger(row.raw) ? String(row.raw) : row.raw.toFixed(2)
}

function formatWeight(w: number): string {
  if (w === 0) return '—'
  return (w > 0 ? '+' : '') + w.toFixed(2)
}

function formatContrib(c: number): string {
  return (c >= 0 ? '+' : '') + c.toFixed(2)
}

function weightClass(w: number): string {
  if (w > 0) return 'tone-pos'
  if (w < 0) return 'tone-neg'
  return 'tone-zero'
}

function contribClass(c: number): string {
  if (c > 0.5) return 'tone-pos'
  if (c < -0.5) return 'tone-neg'
  return 'tone-zero'
}

async function requestExplain() {
  if (!active.value) return
  explainLoading.value = true
  explainError.value = null
  try {
    const enq = await api.requestScoreExplanation(props.secid, active.value.profile_code)
    await pollJob(enq.job_id)
    emit('refresh')
  } catch (e: any) {
    explainError.value = e?.data?.error || e?.message || 'Не удалось сгенерировать разбор'
  } finally {
    explainLoading.value = false
  }
}

// Polls /jobs/{id} every 2s up to a 60s ceiling — same cadence as the
// existing AI-analysis pipeline. Worker latency is dominated by the
// OpenAI call (~5-15s), so the loop usually exits in 1-2 ticks.
async function pollJob(jobId: string): Promise<void> {
  const deadline = Date.now() + 60_000
  while (Date.now() < deadline) {
    const j = await api.getJobStatus(jobId)
    if (j.status === 'done')  return
    if (j.status === 'error') throw new Error(j.error || 'job failed')
    await new Promise(r => setTimeout(r, 2000))
  }
  throw new Error('timeout: разбор не успел сгенерироваться за минуту')
}
</script>

<style scoped>
.bond-score {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* head panel */
.bond-score__profiles {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1px;
  background: var(--nla-border-light);
}
.profile-card {
  appearance: none;
  border: 0;
  background: var(--nla-bg-card);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 20px;
  cursor: pointer;
  text-align: left;
  transition: background 120ms ease;
}
.profile-card:hover { background: var(--nla-bg-subtle); }
.profile-card.is-active {
  background: var(--nla-primary-light);
  box-shadow: inset 3px 0 0 var(--nla-primary);
}
.profile-card__icon { font-size: 22px; line-height: 1; }
.profile-card__body { display: flex; flex-direction: column; gap: 4px; }
.profile-card__lbl {
  font: 500 11px / 1 var(--nla-font);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
.profile-card__val {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: var(--nla-radius-pill);
  font: 700 18px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.bond-score__missing {
  padding: 10px 18px;
  background: var(--nla-warning-light, var(--nla-bg-subtle));
  color: var(--nla-warning, var(--nla-text-secondary));
  font: 500 12px/1.5 var(--nla-font);
  display: flex;
  gap: 8px;
  align-items: flex-start;
  border-top: 1px solid var(--nla-border-light);
}
.bond-score__missing i { font-size: 14px; flex-shrink: 0; margin-top: 2px; }
.bond-score__missing strong { color: var(--nla-text); font-weight: 600; }

/* breakdown panel */
.bond-score__toggle {
  margin-left: auto;
  appearance: none;
  border: 1px solid var(--nla-border);
  background: var(--nla-bg-card);
  color: var(--nla-text-secondary);
  font: 500 12px/1 var(--nla-font);
  padding: 5px 10px;
  border-radius: var(--nla-radius);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.bond-score__toggle:hover { background: var(--nla-bg-subtle); color: var(--nla-text); }

.breakdown-table {
  width: 100%;
  border-collapse: collapse;
  font: 500 12.5px/1.4 var(--nla-font);
}
.breakdown-table th,
.breakdown-table td {
  padding: 9px 14px;
  text-align: left;
  border-bottom: 1px solid var(--nla-border-light);
}
.breakdown-table th {
  font: 600 10px/1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  background: var(--nla-bg-elevated);
}
.breakdown-table tbody tr:hover { background: var(--nla-bg-subtle); }
.breakdown-table tr.is-missing td { color: var(--nla-text-muted); }
.factor-name { color: var(--nla-text); font-weight: 500; }
.factor-flag {
  margin-left: 8px;
  font: 600 9.5px/1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  background: var(--nla-bg-subtle);
  padding: 2px 6px;
  border-radius: 4px;
}
.t-raw, .t-norm, .t-weight, .t-contrib {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.t-raw     { width: 100px; }
.t-norm    { width: 160px; }
.t-weight  { width: 80px;  text-align: right; }
.t-contrib { width: 90px;  text-align: right; }
.t-foot { font-weight: 700; background: var(--nla-bg-elevated); }

.tone-pos  { color: var(--nla-success); }
.tone-neg  { color: var(--nla-danger); }
.tone-zero { color: var(--nla-text-muted); }

.norm-bar {
  position: relative;
  height: 18px;
  background: var(--nla-bg-subtle);
  border-radius: 4px;
  overflow: hidden;
}
.norm-bar__fill {
  position: absolute;
  inset: 0 auto 0 0;
  background: linear-gradient(90deg,
              color-mix(in oklab, var(--nla-primary) 18%, transparent) 0%,
              color-mix(in oklab, var(--nla-primary) 60%, transparent) 100%);
  transition: width 200ms ease;
}
.norm-bar__val {
  position: relative;
  padding: 0 8px;
  display: inline-block;
  line-height: 18px;
  font-weight: 600;
  font-size: 11px;
  color: var(--nla-text);
}

/* explain panel */
.bond-score__explain-btn {
  margin-left: auto;
  appearance: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--nla-primary);
  border: 1px solid var(--nla-primary);
  color: var(--nla-text-inverse);
  font: 600 12.5px/1 var(--nla-font);
  border-radius: var(--nla-radius);
  cursor: pointer;
}
.bond-score__explain-btn:hover:not(:disabled) {
  background: var(--nla-primary-hover);
  border-color: var(--nla-primary-hover);
}
.bond-score__explain-btn:disabled { opacity: 0.6; cursor: progress; }
.bond-score__explain-btn i { font-size: 13px; }

.explain-body { padding: 16px 18px; }
.explain-text {
  font: 500 14px/1.65 var(--nla-font);
  color: var(--nla-text);
  white-space: pre-wrap;
}
.explain-empty {
  font: 500 13px/1.6 var(--nla-font);
  color: var(--nla-text-muted);
}
.explain-empty i { margin-right: 6px; }
.explain-error {
  margin-top: 10px;
  padding: 10px 12px;
  background: var(--nla-danger-light, var(--nla-bg-subtle));
  color: var(--nla-danger);
  font: 500 12px/1.5 var(--nla-font);
  border-radius: var(--nla-radius);
  display: flex;
  gap: 8px;
  align-items: flex-start;
}
.explain-error i { font-size: 14px; flex-shrink: 0; margin-top: 2px; }

.spin { animation: spin 900ms linear infinite; display: inline-block; }
@keyframes spin { from { transform: rotate(0); } to { transform: rotate(360deg); } }

@media (max-width: 768px) {
  .bond-score__profiles { grid-template-columns: 1fr; }
  .profile-card.is-active { box-shadow: inset 0 3px 0 var(--nla-primary); }
  .t-raw, .t-norm  { width: auto; }
  .breakdown-table th, .breakdown-table td { padding: 8px 10px; }
}
</style>
