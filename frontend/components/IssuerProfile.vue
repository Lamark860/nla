<!--
  IssuerProfile.vue — карточка эмитента (страница /issuers/:id).
  Структура из handoff/preview/bond-detail.html → IssuerTab:
    1. Header: логотип-буква + имя + sub-line (ID, сектор, теги) + кнопка "Все эмитенты"
    2. 4-cell stats: облигаций в обращении / общий объём / средняя YTM / аналитический индекс
    3. Bonds table с подсветкой current
    4. Ratings cards (agency / grade / updated_at)
    5. Quality bars (из dohod): кредитное качество / стабильность / баланс / прибыль + сводный балл
-->
<template>
  <div class="ip-stack">
    <!-- 1+2. Header + stats in one Panel -->
    <Panel flush>
      <div class="ip-hdr">
        <div class="ip-logo" :style="logoStyle">{{ logoLetter }}</div>
        <div class="ip-meta">
          <h1 class="ip-name">{{ name || `Эмитент #${emitterId}` }}</h1>
          <div class="ip-sub">
            <span>ID <strong class="num">{{ emitterId }}</strong></span>
            <template v-if="sector">
              <span class="ip-sub__dot"></span>
              <span>{{ sector }}</span>
            </template>
            <template v-if="inn">
              <span class="ip-sub__dot"></span>
              <span>ИНН <strong class="num">{{ inn }}</strong></span>
            </template>
            <template v-if="ogrn">
              <span class="ip-sub__dot"></span>
              <span>ОГРН <strong class="num">{{ ogrn }}</strong></span>
            </template>
          </div>
          <div v-if="tags?.length" class="ip-tags">
            <Tag v-for="t in tags" :key="t">{{ t }}</Tag>
          </div>
        </div>
        <NuxtLink to="/bonds/by-issuer" class="ip-back-btn">
          <i class="bi bi-box-arrow-up-right" aria-hidden="true"></i>
          <span>Все эмитенты</span>
        </NuxtLink>
      </div>

      <div class="ip-stats">
        <div class="ip-cell">
          <div class="ip-lbl">Облигаций в обращении</div>
          <div class="ip-val">{{ bondCount }}</div>
          <div v-if="floatCount > 0 || fixedCount > 0" class="ip-sub-val">
            <template v-if="floatCount > 0">{{ floatCount }} {{ pluralFloat(floatCount) }}</template>
            <template v-if="floatCount > 0 && fixedCount > 0">, </template>
            <template v-if="fixedCount > 0">{{ fixedCount }} {{ pluralFix(fixedCount) }}</template>
          </div>
        </div>
        <div class="ip-cell">
          <div class="ip-lbl">Общий объём</div>
          <div class="ip-val">{{ totalDebt != null ? formatBigRub(totalDebt) : '—' }}</div>
          <div class="ip-sub-val">по номиналу</div>
        </div>
        <div class="ip-cell">
          <div class="ip-lbl">Средняя YTM</div>
          <div class="ip-val ip-val--up">{{ avgYield != null ? fmt.percent(avgYield) : '—' }}</div>
          <div class="ip-sub-val">взвеш. по объёму</div>
        </div>
        <div class="ip-cell">
          <div class="ip-lbl">Индекс эмитента</div>
          <div class="ip-val">
            <AiScore v-if="avgAi != null" :value="avgAi" />
            <span v-else>—</span>
          </div>
          <div class="ip-sub-val">{{ aiTotalCount > 0 ? `по ${aiTotalCount} ${pluralAnaliz(aiTotalCount)}` : 'нет анализов' }}</div>
        </div>
      </div>
    </Panel>

    <!-- 3. Bonds table -->
    <Panel
      v-if="bonds && bonds.length"
      title="Облигации эмитента"
      icon="list-ul"
      :meta="`${bonds.length} ${pluralBond(bonds.length)}`"
      flush
    >
      <div class="table-responsive">
        <table class="ip-bonds-tbl">
          <thead>
            <tr>
              <th>SECID</th>
              <th>Название</th>
              <th>Погашение</th>
              <th class="right">Купон</th>
              <th class="right">Цена</th>
              <th class="right">YTM</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="b in displayedTableBonds"
              :key="b.secid"
              :class="{ 'is-current': currentSecid && b.secid === currentSecid }"
            >
              <td class="ip-bonds__secid">{{ b.secid }}</td>
              <td class="ip-bonds__name">
                <NuxtLink :to="`/bonds/${b.secid}`">{{ b.shortname }}</NuxtLink>
                <span v-if="currentSecid && b.secid === currentSecid" class="ip-bonds__here">текущая</span>
                <Tag v-if="b.is_float" tone="primary" class="ip-bonds__tag">Флоатер</Tag>
              </td>
              <td class="ip-bonds__date">{{ fmt.dateShort(b.matdate) }}</td>
              <td class="right">{{ b.coupon_display != null ? fmt.percent(b.coupon_display) : '—' }}</td>
              <td class="right">{{ b.last != null ? fmt.percent(b.last) : '—' }}</td>
              <td class="right ip-bonds__ytm">{{ b.yield != null ? fmt.percent(b.yield) : '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <button v-if="hasHiddenTableBonds" class="ip-bonds__toggle" @click="showAllTableBonds = !showAllTableBonds">
        <i :class="showAllTableBonds ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
        <span v-if="!showAllTableBonds">
          Показать все {{ sortedBonds.length }} {{ pluralBond(sortedBonds.length) }} · скрыто {{ sortedBonds.length - TABLE_VISIBLE_LIMIT }}
        </span>
        <span v-else>Свернуть до {{ TABLE_VISIBLE_LIMIT }}</span>
      </button>
    </Panel>

    <!-- 4+5. Ratings + Quality side by side -->
    <div v-if="hasRatings || hasQuality" class="ip-grid">
      <Panel
        v-if="hasRatings"
        title="Кредитные рейтинги"
        icon="shield-check"
        :meta="`${ratings!.length} ${pluralAgency(ratings!.length)}`"
      >
        <div class="ip-ratings">
          <div v-for="r in ratings" :key="r.agency" class="ip-rating">
            <div class="ip-rating__agency">{{ r.agency }}</div>
            <div class="ip-rating__grade" :class="`ip-rating__grade--${r.tier ?? ''}`">{{ r.rating }}</div>
            <div v-if="r.updated_at" class="ip-rating__date">обновлено {{ fmt.dateShort(r.updated_at) }}</div>
          </div>
        </div>
      </Panel>

      <Panel
        v-if="hasQuality"
        title="Качество эмитента"
        icon="bar-chart-line"
        meta="по dohod.ru"
      >
        <div class="ip-quality">
          <div v-for="q in qualityRows" :key="q.label" class="ip-qbar">
            <span class="ip-qbar__lbl">{{ q.label }}</span>
            <span class="ip-qbar__track">
              <span class="ip-qbar__fill" :style="{ width: (q.value * 10) + '%' }"></span>
            </span>
            <span class="ip-qbar__val">{{ q.value.toFixed(1) }}</span>
          </div>
          <div v-if="dohod?.best_score != null" class="ip-qbar__total">
            <span>Сводный балл</span>
            <span class="ip-qbar__total-val">{{ dohod.best_score.toFixed(2) }} <small>/ 10</small></span>
          </div>
        </div>
        <p v-if="dohod?.description" class="ip-quality__desc">{{ dohod.description }}</p>
      </Panel>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond, AnalysisStats, DohodBondData } from '~/composables/useApi'

interface RatingRow {
  agency: string
  rating: string
  updated_at?: string
  tier?: string  // class suffix: aaa/aa/a/bbb/bb/b/ccc/d
}

const props = defineProps<{
  emitterId: number | string
  name?: string
  inn?: string
  ogrn?: string
  sector?: string
  tags?: string[]
  bonds?: Bond[]
  ratings?: RatingRow[]
  aiStats?: Record<string, AnalysisStats>
  dohod?: DohodBondData | null
  currentSecid?: string
}>()

const fmt = useFormat()

const bondCount = computed(() => props.bonds?.length ?? 0)

const floatCount = computed(() => props.bonds?.filter(b => b.is_float).length ?? 0)
const fixedCount = computed(() => bondCount.value - floatCount.value - (props.bonds?.filter(b => b.is_indexed).length ?? 0))

// Total debt = sum(facevalue × placed)
const totalDebt = computed(() => {
  if (!props.bonds?.length) return null
  let total = 0
  let counted = 0
  for (const b of props.bonds) {
    if (b.facevalue && b.issuesize_placed) {
      total += b.facevalue * b.issuesize_placed
      counted++
    }
  }
  return counted > 0 ? total : null
})

// Weighted YTM by placed volume
const avgYield = computed(() => {
  if (!props.bonds?.length) return null
  let sumWeighted = 0
  let sumWeight = 0
  for (const b of props.bonds) {
    if (b.yield != null && Number.isFinite(b.yield) && Math.abs(b.yield) < 100 && b.issuesize_placed) {
      sumWeighted += b.yield * b.issuesize_placed
      sumWeight += b.issuesize_placed
    }
  }
  if (sumWeight === 0) {
    // fallback на простое среднее, фильтруя выбросы
    const ys = props.bonds.map(b => b.yield).filter((v): v is number => v != null && Number.isFinite(v) && Math.abs(v) < 100)
    return ys.length ? ys.reduce((a, b) => a + b, 0) / ys.length : null
  }
  return sumWeighted / sumWeight
})

const avgAi = computed(() => {
  const stats = props.aiStats
  if (!stats || !props.bonds?.length) return null
  let sum = 0, total = 0
  for (const b of props.bonds) {
    const s = stats[b.secid]
    if (s && s.avg_rating > 0) {
      sum += s.avg_rating * s.total
      total += s.total
    }
  }
  return total > 0 ? sum / total : null
})

const aiTotalCount = computed(() => {
  const stats = props.aiStats
  if (!stats || !props.bonds?.length) return 0
  let total = 0
  for (const b of props.bonds) {
    const s = stats[b.secid]
    if (s) total += s.total
  }
  return total
})

const sortedBonds = computed(() => {
  if (!props.bonds) return []
  // current first, then by mat date asc
  return [...props.bonds].sort((a, b) => {
    if (props.currentSecid) {
      if (a.secid === props.currentSecid) return -1
      if (b.secid === props.currentSecid) return 1
    }
    return (a.matdate || '').localeCompare(b.matdate || '')
  })
})

const TABLE_VISIBLE_LIMIT = 10
const showAllTableBonds = ref(false)
const displayedTableBonds = computed(() => {
  if (showAllTableBonds.value || sortedBonds.value.length <= TABLE_VISIBLE_LIMIT) return sortedBonds.value
  return sortedBonds.value.slice(0, TABLE_VISIBLE_LIMIT)
})
const hasHiddenTableBonds = computed(() => sortedBonds.value.length > TABLE_VISIBLE_LIMIT)

const hasRatings = computed(() => (props.ratings?.length ?? 0) > 0)

// Quality rows from dohod — skip null/zero values
interface QualityRow { label: string; value: number }
const qualityRows = computed<QualityRow[]>(() => {
  const d = props.dohod
  if (!d) return []
  const rows: QualityRow[] = []
  if (d.quality != null && d.quality > 0)             rows.push({ label: 'Кредитное качество',     value: d.quality })
  if (d.stability != null && d.stability > 0)         rows.push({ label: 'Финансовая стабильность', value: d.stability })
  if (d.quality_balance != null && d.quality_balance > 0) rows.push({ label: 'Баланс',              value: d.quality_balance })
  if (d.quality_earnings != null && d.quality_earnings > 0) rows.push({ label: 'Прибыльность',       value: d.quality_earnings })
  return rows
})
const hasQuality = computed(() => qualityRows.value.length > 0)

// Logo letter and color (deterministic by name)
const logoLetter = computed(() => {
  const n = props.name || `Эмитент ${props.emitterId}`
  // Skip quotes and 'АО'/'ООО'/'ПАО' prefix
  const cleaned = n.replace(/^(АО|ООО|ПАО|ЗАО|АО\s*НПФ|ООО\s*НПФ)\s+/i, '').replace(/^["«»']/, '')
  return (cleaned[0] || 'N').toUpperCase()
})
const logoStyle = computed(() => {
  // hue based on string hash
  const s = String(props.name || props.emitterId)
  let hash = 0
  for (let i = 0; i < s.length; i++) hash = ((hash << 5) - hash + s.charCodeAt(i)) | 0
  const hue = Math.abs(hash) % 360
  return {
    background: `linear-gradient(135deg, hsl(${hue}, 55%, 45%), hsl(${(hue + 30) % 360}, 60%, 35%))`,
  }
})

function formatBigRub(v: number): string {
  if (v >= 1e12) return (v / 1e12).toFixed(2).replace('.', ',') + ' трлн ₽'
  if (v >= 1e9)  return (v / 1e9).toFixed(2).replace('.', ',') + ' млрд ₽'
  if (v >= 1e6)  return (v / 1e6).toFixed(1).replace('.', ',') + ' млн ₽'
  if (v >= 1e3)  return (v / 1e3).toFixed(0) + ' тыс ₽'
  return fmt.priceRub(v)
}

function pluralBond(n: number) { return plural(n, ['облигация', 'облигации', 'облигаций']) }
function pluralFloat(n: number) { return plural(n, ['флоатер', 'флоатера', 'флоатеров']) }
function pluralFix(n: number) { return plural(n, ['с фикс. купоном', 'с фикс. купоном', 'с фикс. купоном']) }
function pluralAnaliz(n: number) { return plural(n, ['анализу', 'анализам', 'анализам']) }
function pluralAgency(n: number) { return plural(n, ['агентство', 'агентства', 'агентств']) }
function plural(n: number, forms: [string, string, string]) {
  const m10 = n % 10, m100 = n % 100
  if (m100 >= 11 && m100 <= 19) return forms[2]
  if (m10 === 1) return forms[0]
  if (m10 >= 2 && m10 <= 4) return forms[1]
  return forms[2]
}
</script>

<style scoped>
.ip-stack { display: flex; flex-direction: column; gap: 16px; }

/* Header */
.ip-hdr {
  display: grid;
  grid-template-columns: 64px 1fr auto;
  gap: 16px;
  align-items: center;
  padding: 18px 22px;
  background: var(--nla-bg-elevated);
  border-bottom: 1px solid var(--nla-border);
}
.ip-logo {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  color: #fff;
  font: 700 24px/1 var(--nla-font);
  letter-spacing: -0.02em;
  box-shadow: var(--nla-shadow-sm);
}
.ip-meta { min-width: 0; }
.ip-name {
  margin: 0 0 4px 0;
  font: 700 20px/1.2 var(--nla-font);
  letter-spacing: -0.02em;
  color: var(--nla-text);
}
.ip-sub {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  font: 500 12.5px/1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.ip-sub .num {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  font-weight: 600;
}
.ip-sub__dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: var(--nla-text-muted);
  opacity: 0.5;
}
.ip-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 8px;
}
.ip-back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 12px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 12.5px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  text-decoration: none;
  flex-shrink: 0;
}
.ip-back-btn:hover {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  border-color: color-mix(in oklab, var(--nla-primary) 25%, var(--nla-border));
}

/* Stats 4-cell */
.ip-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
}
.ip-cell {
  padding: 16px 22px;
  border-right: 1px solid var(--nla-border-light);
}
.ip-cell:last-child { border-right: 0; }
.ip-lbl {
  font: 600 10.5px/1.2 var(--nla-font);
  color: var(--nla-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 6px;
}
.ip-val {
  font: 600 22px/1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  color: var(--nla-text);
}
.ip-val--up { color: var(--nla-success); }
.ip-sub-val {
  font: 500 11.5px/1.3 var(--nla-font);
  color: var(--nla-text-muted);
  margin-top: 4px;
  font-feature-settings: 'tnum';
}

@media (max-width: 768px) {
  .ip-hdr { grid-template-columns: 48px 1fr; }
  .ip-back-btn { grid-column: 1 / -1; justify-content: center; }
  .ip-logo { width: 48px; height: 48px; font-size: 20px; }
  .ip-stats { grid-template-columns: repeat(2, 1fr); }
  .ip-cell:nth-child(2) { border-right: 0; }
  .ip-cell:nth-child(n+3) { border-top: 1px solid var(--nla-border-light); }
}

/* Bonds table */
.ip-bonds-tbl {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 13px;
}
.ip-bonds-tbl th {
  background: var(--nla-bg-elevated);
  text-align: left;
  padding: 9px 14px;
  font: 600 10.5px/1 var(--nla-font);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--nla-text-muted);
  border-bottom: 1px solid var(--nla-border);
  white-space: nowrap;
}
.ip-bonds-tbl th.right { text-align: right; }
.ip-bonds-tbl td {
  padding: 11px 14px;
  border-bottom: 1px solid var(--nla-border-light);
  color: var(--nla-text-secondary);
  vertical-align: middle;
}
.ip-bonds-tbl td.right {
  text-align: right;
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
}
.ip-bonds-tbl tr:last-child td { border-bottom: 0; }

.ip-bonds__toggle {
  appearance: none;
  width: 100%;
  border: 0;
  border-top: 1px solid var(--nla-border-light);
  background: var(--nla-bg-elevated);
  padding: 11px 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font: 500 12.5px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
}
.ip-bonds__toggle:hover {
  background: var(--nla-bg-subtle);
  color: var(--nla-primary);
}
.ip-bonds__toggle .bi { font-size: 13px; }
.ip-bonds-tbl tr.is-current { background: var(--nla-primary-light); }
.ip-bonds-tbl tr.is-current td { color: var(--nla-primary-ink); font-weight: 600; }
[data-theme="dark"] .ip-bonds-tbl tr.is-current td { color: var(--nla-primary); }

.ip-bonds__secid {
  font: 500 11.5px/1 var(--nla-font-mono);
  color: var(--nla-text-muted);
}
.ip-bonds__name { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.ip-bonds__name a { color: inherit; text-decoration: none; font-weight: 500; }
.ip-bonds__name a:hover { color: var(--nla-primary); }
.ip-bonds__here {
  font: 600 9.5px/1 var(--nla-font);
  color: var(--nla-primary);
  background: var(--nla-primary-light);
  padding: 2px 6px;
  border-radius: var(--nla-radius-sm);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.ip-bonds__tag { font-size: 9.5px !important; padding: 2px 6px !important; }
.ip-bonds__date { color: var(--nla-text-muted); font-family: var(--nla-font-mono); }
.ip-bonds__ytm { color: var(--nla-success); }
[data-theme="dark"] .ip-bonds__ytm { color: var(--nla-success); }

/* Ratings + Quality grid */
.ip-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
@media (max-width: 992px) {
  .ip-grid { grid-template-columns: 1fr; }
}

/* Ratings cards */
.ip-ratings {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 16px 18px;
}
@media (max-width: 480px) {
  .ip-ratings { grid-template-columns: 1fr; }
}
.ip-rating {
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border-light);
  border-radius: var(--nla-radius);
  padding: 10px 14px;
}
.ip-rating__agency {
  font: 600 10.5px/1 var(--nla-font);
  color: var(--nla-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 6px;
}
.ip-rating__grade {
  font: 700 16px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  letter-spacing: -0.01em;
}
.ip-rating__date {
  font: 500 11px/1 var(--nla-font);
  color: var(--nla-text-muted);
  margin-top: 6px;
  font-feature-settings: 'tnum';
}

/* Quality bars */
.ip-quality {
  padding: 16px 18px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.ip-qbar {
  display: grid;
  grid-template-columns: minmax(120px, 1fr) 2fr 36px;
  align-items: center;
  gap: 12px;
}
.ip-qbar__lbl {
  font: 500 12.5px/1.3 var(--nla-font);
  color: var(--nla-text-secondary);
}
.ip-qbar__track {
  height: 6px;
  background: var(--nla-bg-subtle);
  border-radius: var(--nla-radius-sm);
  overflow: hidden;
}
.ip-qbar__fill {
  display: block;
  height: 100%;
  background: var(--nla-success);
  border-radius: var(--nla-radius-sm);
  transition: width 320ms ease;
}
.ip-qbar__val {
  font: 600 13px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  text-align: right;
}
.ip-qbar__total {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-top: 6px;
  padding-top: 12px;
  border-top: 1px solid var(--nla-border-light);
  font: 500 12px/1 var(--nla-font);
  color: var(--nla-text-muted);
}
.ip-qbar__total-val {
  font: 700 22px/1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-success);
}
.ip-qbar__total-val small {
  font: 500 12px/1 var(--nla-font);
  color: var(--nla-text-muted);
}
.ip-quality__desc {
  margin: 8px 18px 16px;
  padding-top: 12px;
  border-top: 1px solid var(--nla-border-light);
  font: 500 12.5px/1.5 var(--nla-font);
  color: var(--nla-text-secondary);
}
</style>
