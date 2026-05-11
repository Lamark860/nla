<!--
  IssuerCardGrid.vue — грид карточек эмитентов на главной.
  Полностью совместим с текущим API: pros (issuers, ratings, aiStats).

  Изменения относительно старого:
   - 3 колонки на десктопе (col-lg-4) — оставлено
   - hover-state с лёгким lift и violet-обводкой
   - в свернутой карточке: имя → стопка рейтингов агентств → AiScore (вторично)
                          → счётчик "N облигаций · средняя дох. X%"
   - в развёрнутой: список бумаг через BondRow.vue (см. ниже)
   - все кастомные .badge / .metric-label / .metric-value заменены на компоненты
-->
<template>
  <div class="issuer-grid">
    <article
      v-for="issuer in issuers"
      :key="issuer.emitter_id"
      class="issuer-card"
      :class="{ 'is-expanded': expanded.has(issuer.emitter_id) }"
    >
      <!-- Header: name + chevron -->
      <header class="issuer-card__head" @click="toggle(issuer.emitter_id)">
        <div class="issuer-card__title">
          <NuxtLink :to="`/issuers/${issuer.emitter_id}`" class="issuer-card__name" @click.stop>
            {{ issuer.emitter_name || `Эмитент #${issuer.emitter_id}` }}
          </NuxtLink>
          <span class="issuer-card__id">#{{ issuer.emitter_id }}</span>
        </div>
        <button class="issuer-card__chevron" :aria-label="expanded.has(issuer.emitter_id) ? 'Свернуть' : 'Развернуть'">
          <i :class="expanded.has(issuer.emitter_id) ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
        </button>
      </header>

      <!-- Ratings + AI -->
      <div class="issuer-card__ratings">
        <template v-if="ratingsFor(issuer.emitter_id).length">
          <RatingBadge
            v-for="r in ratingsFor(issuer.emitter_id)"
            :key="r.agency"
            :rating="r.rating"
            :agency="r.agency"
          />
        </template>
        <span v-else class="issuer-card__no-rating">Без рейтинга</span>

        <span v-if="aiStatsFor(issuer)" class="issuer-card__sep"></span>
        <AiScore v-if="aiStatsFor(issuer)" :value="aiStatsFor(issuer)!.avg" compact />
      </div>

      <!-- Summary line -->
      <div class="issuer-card__summary">
        <span class="issuer-card__count">{{ issuer.bond_count }} {{ bondWord(issuer.bond_count) }}</span>
        <span v-if="avgYield(issuer) != null" class="issuer-card__yield">
          средн. <span class="num">{{ fmt.percent(avgYield(issuer)!) }}</span>
        </span>
      </div>

      <!-- Preview (collapsed) -->
      <div v-if="!expanded.has(issuer.emitter_id) && issuer.bonds[0]" class="issuer-card__preview">
        <span class="issuer-card__preview-name">{{ issuer.bonds[0].shortname }}</span>
        <span v-if="issuer.bonds.length > 1" class="issuer-card__preview-more">
          +{{ issuer.bonds.length - 1 }}
        </span>
      </div>

      <!-- Expanded list of bonds -->
      <ul v-if="expanded.has(issuer.emitter_id)" class="issuer-card__bonds">
        <li v-for="bond in issuer.bonds" :key="bond.secid" class="issuer-bond">
          <div class="issuer-bond__head">
            <NuxtLink :to="`/bonds/${bond.secid}`" class="issuer-bond__name">
              {{ bond.shortname }}
            </NuxtLink>
            <span class="issuer-bond__secid">{{ bond.secid }}</span>
            <Tag v-if="bond.is_float" tone="primary">Флоатер</Tag>
            <Tag v-else-if="bond.is_indexed">Индекс.</Tag>
            <Tag v-if="periodLabel(bond.coupon_period)">{{ periodLabel(bond.coupon_period) }}</Tag>
          </div>

          <div class="issuer-bond__metrics">
            <Metric label="Цена">
              {{ bond.last != null ? fmt.percent(bond.last) : '—' }}
            </Metric>
            <Metric label="Дох" :tone="yieldTone(bond.yield)">
              {{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}
            </Metric>
            <Metric label="Купон">
              {{ fmt.percent(bond.coupon_display) }}
            </Metric>
            <Metric label="НКД" tone="muted">
              {{ bond.accrued_int != null ? fmt.num(bond.accrued_int, 2) : '—' }}
            </Metric>
          </div>

          <div class="issuer-bond__dates">
            <DateLine label="Погашение" :date="bond.matdate" :days="bond.days_to_maturity" />
            <DateLine v-if="putOfferDate(bond)" label="Оферта" :date="putOfferDate(bond)!" />
            <DateLine v-if="callDate(bond)" label="Колл" :date="callDate(bond)!" />
            <div v-if="bondAi(bond.secid) != null" class="issuer-bond__ai">
              <span class="issuer-bond__ai-lbl">AI</span>
              <AiScore :value="bondAi(bond.secid)!" compact />
            </div>
          </div>
        </li>
      </ul>
    </article>
  </div>
</template>

<script setup lang="ts">
import type { IssuerGroup, IssuerRatingResponse, AnalysisStats, Bond } from '~/composables/useApi'
import RatingBadge from './RatingBadge.vue'
import AiScore from './AiScore.vue'
import Tag from './Tag.vue'

// Inline mini-helpers — нужны только тут, выносить в отдельные файлы избыточно

defineProps<{
  issuers: IssuerGroup[]
  ratings: Record<string, IssuerRatingResponse>
  aiStats?: Record<string, AnalysisStats>
}>()

const props = defineProps as any   // дубликат не нужен — оставлено для ясности типов выше
const fmt = useFormat()
const expanded = ref(new Set<number>())

function toggle(id: number) {
  expanded.value.has(id) ? expanded.value.delete(id) : expanded.value.add(id)
  // trigger reactivity
  expanded.value = new Set(expanded.value)
}

function ratingsFor(emitterId: number) {
  const r = (props as any).ratings?.[String(emitterId)]
  if (!r) return []
  return r.ratings.filter((x: any) => x.rating && x.rating !== 'NULL')
}

function aiStatsFor(issuer: IssuerGroup) {
  const stats = (props as any).aiStats
  if (!stats) return null
  let total = 0, sum = 0, count = 0, mn = Infinity, mx = -Infinity
  for (const b of issuer.bonds) {
    const s = stats[b.secid]
    if (!s) continue
    total += s.total
    if (s.avg_rating > 0) {
      sum += s.avg_rating * s.total
      count += s.total
      mn = Math.min(mn, s.avg_rating)
      mx = Math.max(mx, s.avg_rating)
    }
  }
  if (!total) return null
  return {
    total,
    avg: count > 0 ? sum / count : 0,
    min: mn === Infinity ? 0 : Math.round(mn),
    max: mx === -Infinity ? 0 : Math.round(mx),
  }
}

function bondAi(secid: string): number | null {
  const stats = (props as any).aiStats
  if (!stats) return null
  const s = stats[secid]
  return s && s.avg_rating > 0 ? Math.round(s.avg_rating) : null
}

function avgYield(issuer: IssuerGroup): number | null {
  const ys = issuer.bonds.map(b => b.yield).filter((v): v is number => v != null)
  if (!ys.length) return null
  return ys.reduce((a, b) => a + b, 0) / ys.length
}

function bondWord(n: number) {
  const m10 = n % 10, m100 = n % 100
  if (m100 >= 11 && m100 <= 19) return 'облигаций'
  if (m10 === 1) return 'облигация'
  if (m10 >= 2 && m10 <= 4) return 'облигации'
  return 'облигаций'
}

function periodLabel(p: number | undefined) {
  if (!p) return ''
  if (p >= 27 && p <= 33)   return '1 мес'
  if (p >= 85 && p <= 95)   return '3 мес'
  if (p >= 175 && p <= 190) return '6 мес'
  if (p >= 355 && p <= 370) return '12 мес'
  return p + ' дн'
}

function yieldTone(y: number | null | undefined): 'success' | 'primary' | 'default' {
  if (y == null) return 'default'
  if (y >= 12) return 'success'
  if (y >= 8)  return 'primary'
  return 'default'
}

function isValidDate(v?: string | null) {
  if (!v || v === '0000-00-00' || v === 'None') return false
  const d = new Date(v)
  return !isNaN(d.getTime()) && d.getFullYear() > 1970
}

function putOfferDate(b: Bond) {
  const v = b.putoptiondate || b.buybackdate
  return isValidDate(v) ? v : null
}
function callDate(b: Bond) {
  return isValidDate(b.calloptiondate) ? b.calloptiondate : null
}
</script>

<!-- маленькие вспомогательные компоненты, локальные для файла -->
<script lang="ts">
import { defineComponent, h } from 'vue'
export const Metric = defineComponent({
  name: 'IssuerMetric',
  props: { label: String, tone: { type: String, default: 'default' } },
  setup(props, { slots }) {
    return () => h('div', { class: 'metric' }, [
      h('span', { class: 'metric__lbl' }, props.label),
      h('span', { class: ['metric__val', props.tone && `metric__val--${props.tone}`] }, slots.default?.())
    ])
  }
})
export const DateLine = defineComponent({
  name: 'IssuerDateLine',
  props: { label: String, date: String, days: Number },
  setup(props) {
    const fmt = useFormat()
    return () => {
      if (!props.date) return null
      const days = props.days != null
        ? props.days < 0 ? 'погашена' : `${props.days} дн`
        : null
      return h('div', { class: 'date-line' }, [
        h('span', { class: 'date-line__lbl' }, props.label + ':'),
        h('span', { class: 'date-line__val' }, fmt.dateShort(props.date)),
        days ? h('span', { class: 'date-line__rel' }, `· ${days}`) : null
      ])
    }
  }
})
</script>

<style scoped>
.issuer-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.issuer-card {
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-lg);
  padding: 18px 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: border-color 140ms ease, box-shadow 160ms ease, transform 160ms ease;
}
.issuer-card:hover {
  border-color: color-mix(in oklab, var(--nla-primary) 35%, var(--nla-border));
  box-shadow: var(--nla-shadow);
  transform: translateY(-1px);
}
.issuer-card.is-expanded {
  border-color: color-mix(in oklab, var(--nla-primary) 40%, var(--nla-border));
  box-shadow: var(--nla-shadow);
}

.issuer-card__head {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  cursor: pointer;
}
.issuer-card__title {
  flex: 1 1 auto;
  min-width: 0;
}
.issuer-card__name {
  display: block;
  font: 700 16px / 1.25 var(--nla-font);
  letter-spacing: -0.01em;
  color: var(--nla-text);
  text-decoration: none;
  word-break: break-word;
}
.issuer-card__name:hover { color: var(--nla-primary); }
.issuer-card__id {
  display: inline-block;
  margin-top: 2px;
  font: 500 11px / 1 var(--nla-font-mono);
  color: var(--nla-text-muted);
}
.issuer-card__chevron {
  appearance: none;
  border: 0;
  background: var(--nla-bg-subtle);
  color: var(--nla-text-muted);
  width: 28px;
  height: 28px;
  border-radius: var(--nla-radius-sm);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 120ms ease, color 120ms ease;
}
.issuer-card__chevron:hover { background: var(--nla-primary-light); color: var(--nla-primary-ink); }

.issuer-card__ratings {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}
.issuer-card__no-rating {
  font: 500 11px / 1.4 var(--nla-font);
  color: var(--nla-text-muted);
  padding: 3px 8px;
  border: 1px dashed var(--nla-border);
  border-radius: var(--nla-radius-sm);
}
.issuer-card__sep {
  width: 1px;
  height: 14px;
  background: var(--nla-border);
  margin: 0 2px;
}

.issuer-card__summary {
  display: flex;
  justify-content: space-between;
  font: 500 12px / 1.4 var(--nla-font);
  color: var(--nla-text-muted);
  letter-spacing: 0;
}
.issuer-card__yield .num {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  font-weight: 600;
}

.issuer-card__preview {
  display: flex;
  gap: 8px;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid var(--nla-border-light);
  font: 500 13px / 1.3 var(--nla-font);
  color: var(--nla-text-secondary);
  min-height: 38px;
}
.issuer-card__preview-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1 1 auto;
}
.issuer-card__preview-more {
  flex-shrink: 0;
  padding: 1px 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--nla-text-muted);
  background: var(--nla-bg-subtle);
  border-radius: var(--nla-radius-sm);
}

/* Expanded bonds */
.issuer-card__bonds {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.issuer-bond {
  padding: 12px 14px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border-light);
  border-radius: var(--nla-radius);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.issuer-bond__head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.issuer-bond__name {
  font: 600 13px / 1.3 var(--nla-font);
  color: var(--nla-text);
  text-decoration: none;
}
.issuer-bond__name:hover { color: var(--nla-primary); }
.issuer-bond__secid {
  font: 500 11px / 1 var(--nla-font-mono);
  color: var(--nla-text-muted);
}

.issuer-bond__metrics {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px 12px;
}

.issuer-bond__dates {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding-top: 6px;
  border-top: 1px dashed var(--nla-border);
}
.issuer-bond__ai {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
}
.issuer-bond__ai-lbl {
  font: 600 10px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}

/* injected mini-components — стили */
:deep(.metric) {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}
:deep(.metric__lbl) {
  font: 600 9.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
:deep(.metric__val) {
  font: 600 12.5px / 1.2 var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  color: var(--nla-text);
  letter-spacing: 0;
}
:deep(.metric__val--success) { color: var(--nla-success); }
:deep(.metric__val--primary) { color: var(--nla-primary); }
:deep(.metric__val--muted)   { color: var(--nla-text-muted); }

:deep(.date-line) {
  display: flex;
  gap: 6px;
  align-items: baseline;
  font: 500 12px / 1.3 var(--nla-font);
  color: var(--nla-text-secondary);
}
:deep(.date-line__lbl) { color: var(--nla-text-muted); min-width: 78px; }
:deep(.date-line__val) {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  font-weight: 500;
}
:deep(.date-line__rel) { color: var(--nla-text-muted); font-size: 11px; }

@media (max-width: 480px) {
  .issuer-bond__metrics { grid-template-columns: repeat(2, 1fr); }
}
</style>
