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
          <i :class="expanded.has(issuer.emitter_id) ? 'bi-dash-lg' : 'bi-plus-lg'" class="bi"></i>
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
        <span class="issuer-card__count">{{ issuer.bond_count }} {{ bondWord(issuer.bond_count) }} в обращении</span>
        <span v-if="avgYield(issuer) != null" class="issuer-card__yield">
          средн. <span class="num">{{ fmt.percent(avgYield(issuer)!) }}</span>
        </span>
      </div>

      <!-- Preview (collapsed) -->
      <div v-if="!expanded.has(issuer.emitter_id) && issuer.bonds[0]" class="issuer-card__preview">
        <span class="issuer-card__preview-name">{{ issuer.bonds[0].shortname }}</span>
        <span v-if="firstBondYield(issuer) != null" class="issuer-card__preview-yield">
          {{ fmt.percent(firstBondYield(issuer)!) }}
        </span>
        <span v-if="issuer.bonds.length > 1" class="issuer-card__preview-more">
          +{{ issuer.bonds.length - 1 }}
        </span>
      </div>

      <!-- Expanded list of bonds -->
      <ul v-if="expanded.has(issuer.emitter_id)" class="issuer-card__bonds">
        <NuxtLink
          v-for="bond in displayedBondsFor(issuer)"
          :key="bond.secid"
          :to="`/bonds/${bond.secid}`"
          class="issuer-bond"
        >
          <div class="issuer-bond__head">
            <div class="issuer-bond__name-block">
              <strong class="issuer-bond__name">{{ bond.shortname }}</strong>
              <span class="issuer-bond__secid">{{ bond.secid }}</span>
              <Tag v-if="periodLabel(bond.coupon_period)">{{ periodLabel(bond.coupon_period) }}</Tag>
              <Tag v-if="durationLabel(bond.duration)">{{ durationLabel(bond.duration) }}</Tag>
              <Tag v-if="bond.is_float" tone="primary">Флоатер</Tag>
              <Tag v-else-if="bond.is_indexed">Индекс.</Tag>
            </div>
            <span class="issuer-bond__link" title="Подробнее">
              <i class="bi bi-arrow-up-right" aria-hidden="true"></i>
            </span>
          </div>

          <div class="issuer-bond__metrics">
            <Metric label="Цена">
              {{ bond.last != null ? fmt.percent(bond.last) : '—' }}
            </Metric>
            <Metric label="Доходн." :tone="yieldTone(bond.yield)">
              {{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}
            </Metric>
            <Metric label="НКД" tone="muted">
              {{ bond.accrued_int != null ? fmt.num(bond.accrued_int, 2) : '—' }}
            </Metric>
            <Metric label="Купон" :sub="bond.coupon_value ? fmt.priceRub(bond.coupon_value) : null">
              {{ fmt.percent(bond.coupon_display) }}
            </Metric>
          </div>

          <div class="issuer-bond__dates">
            <DateLine
              v-if="bond.next_coupon"
              label="След. купон"
              :date="bond.next_coupon"
              :days="bond.days_to_next_coupon"
            />
            <DateLine
              label="Оферта"
              :date="putOfferDate(bond)"
              :days="putOfferDays(bond)"
            />
            <DateLine
              label="Колл-опцион"
              :date="callDate(bond)"
              :days="callDays(bond)"
            />
            <DateLine
              label="Погашение"
              :date="bond.matdate"
              :days="bond.days_to_maturity"
            />
            <div v-if="bondAi(bond.secid) != null" class="issuer-bond__ai-row">
              <span class="issuer-bond__ai-lbl">Индекс</span>
              <span class="issuer-bond__ai-pill">
                <span class="issuer-bond__ai-dot"></span>
                <span class="issuer-bond__ai-num">{{ bondAi(bond.secid) }}<span class="issuer-bond__ai-max">/100</span></span>
              </span>
            </div>
          </div>
        </NuxtLink>
        <button
          v-if="hasHiddenBondsFor(issuer)"
          class="issuer-bonds__toggle"
          @click="toggleShowAllFor(issuer.emitter_id)"
        >
          <i :class="showAllBonds.has(issuer.emitter_id) ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
          <span v-if="!showAllBonds.has(issuer.emitter_id)">
            Показать все {{ issuer.bonds.length }} бумаг · скрыто {{ hiddenBondsCountFor(issuer) }}
          </span>
          <span v-else>Свернуть до {{ BONDS_VISIBLE_LIMIT }}</span>
        </button>
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

const props = defineProps<{
  issuers: IssuerGroup[]
  ratings: Record<string, IssuerRatingResponse>
  aiStats?: Record<string, AnalysisStats>
}>()

const fmt = useFormat()
const expanded = ref(new Set<number>())
const showAllBonds = ref(new Set<number>())

const BONDS_VISIBLE_LIMIT = 10

function displayedBondsFor(issuer: IssuerGroup): Bond[] {
  if (showAllBonds.value.has(issuer.emitter_id)) return issuer.bonds
  if (issuer.bonds.length <= BONDS_VISIBLE_LIMIT) return issuer.bonds
  return issuer.bonds.slice(0, BONDS_VISIBLE_LIMIT)
}
function hasHiddenBondsFor(issuer: IssuerGroup): boolean {
  return issuer.bonds.length > BONDS_VISIBLE_LIMIT
}
function hiddenBondsCountFor(issuer: IssuerGroup): number {
  return Math.max(0, issuer.bonds.length - BONDS_VISIBLE_LIMIT)
}
function toggleShowAllFor(id: number) {
  if (showAllBonds.value.has(id)) showAllBonds.value.delete(id)
  else showAllBonds.value.add(id)
  showAllBonds.value = new Set(showAllBonds.value)
}

function toggle(id: number) {
  expanded.value.has(id) ? expanded.value.delete(id) : expanded.value.add(id)
  // trigger reactivity
  expanded.value = new Set(expanded.value)
}

function ratingsFor(emitterId: number) {
  const r = props.ratings?.[String(emitterId)]
  if (!r) return []
  return r.ratings.filter(x => x.rating && x.rating !== 'NULL')
}

function aiStatsFor(issuer: IssuerGroup) {
  const stats = props.aiStats
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
  const stats = props.aiStats
  if (!stats) return null
  const s = stats[secid]
  return s && s.avg_rating > 0 ? Math.round(s.avg_rating) : null
}

function avgYield(issuer: IssuerGroup): number | null {
  const ys = issuer.bonds
    .map(b => b.yield)
    .filter((v): v is number => v != null && Number.isFinite(v) && Math.abs(v) < 100)
  if (!ys.length) return null
  return ys.reduce((a, b) => a + b, 0) / ys.length
}

function firstBondYield(issuer: IssuerGroup): number | null {
  const y = issuer.bonds[0]?.yield
  if (y == null || !Number.isFinite(y) || Math.abs(y) > 999) return null
  return y
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

function daysFromTodayTo(date?: string | null): number | null {
  if (!date || !isValidDate(date)) return null
  const t = new Date(date).getTime()
  if (!Number.isFinite(t)) return null
  return Math.round((t - Date.now()) / 86400_000)
}

function putOfferDays(b: Bond): number | null {
  return daysFromTodayTo(putOfferDate(b))
}
function callDays(b: Bond): number | null {
  return daysFromTodayTo(callDate(b))
}

function durationLabel(days: number | null | undefined): string {
  if (days == null || days <= 0) return ''
  // дюрация компактно: 12 мес / 3,2 г
  if (days < 365) return Math.round(days / 30) + ' мес'
  return (days / 365).toFixed(1).replace('.', ',') + ' г'
}
</script>

<!-- маленькие вспомогательные компоненты, локальные для файла -->
<script lang="ts">
import { defineComponent, h, type PropType } from 'vue'
export const Metric = defineComponent({
  name: 'IssuerMetric',
  props: {
    label: String,
    tone: { type: String, default: 'default' },
    sub: { type: String as PropType<string | null>, default: null },
  },
  setup(props, { slots }) {
    return () => h('div', { class: 'metric' }, [
      h('span', { class: 'metric__lbl' }, props.label),
      h('span', { class: ['metric__val', props.tone && `metric__val--${props.tone}`] }, slots.default?.()),
      props.sub ? h('span', { class: 'metric__sub' }, props.sub) : null,
    ])
  }
})
export const DateLine = defineComponent({
  name: 'IssuerDateLine',
  props: {
    label: String,
    date: { type: String as PropType<string | null>, default: null },
    days: { type: Number as PropType<number | null>, default: null },
  },
  setup(props) {
    const fmt = useFormat()
    return () => {
      // handoff: даты Оферта/Колл/Погашение рендерятся всегда — если нет, в значении пустая черта.
      if (!props.date) {
        return h('div', { class: 'date-line' }, [
          h('span', { class: 'date-line__lbl' }, props.label),
          h('span', { class: 'date-line__val date-line__val--empty' }, '—'),
          h('span'),
        ])
      }
      const d = props.days
      let relText: string | null = null
      let relClass = 'date-line__rel'
      if (d != null) {
        if (d < 0) { relText = 'погашена' }
        else if (d === 0) { relText = 'сегодня'; relClass += ' date-line__rel--soon' }
        else {
          relText = fmt.daysToMaturity(d)
          if (d < 200) relClass += ' date-line__rel--soon'
        }
      }
      return h('div', { class: 'date-line' }, [
        h('span', { class: 'date-line__lbl' }, props.label),
        h('span', { class: 'date-line__val' }, fmt.dateShort(props.date)),
        relText ? h('span', { class: relClass }, relText) : h('span'),
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
.issuer-card__preview-yield {
  flex-shrink: 0;
  font: 600 12.5px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-success);
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
  background: var(--nla-bg-subtle);
  border-top: 1px solid var(--nla-border);
  margin: 12px -20px -16px;
  border-radius: 0 0 var(--nla-radius-lg) var(--nla-radius-lg);
}
.issuer-bond {
  padding: 12px 16px;
  border-bottom: 1px solid var(--nla-border-light);
  display: flex;
  flex-direction: column;
  gap: 10px;
  text-decoration: none;
  color: inherit;
  transition: background 120ms ease;
}
.issuer-bond:last-child { border-bottom: 0; }
.issuer-bond:hover { background: var(--nla-bg-card); }
.issuer-bond:hover .issuer-bond__link {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  border-color: color-mix(in oklab, var(--nla-primary) 25%, var(--nla-border));
}

.issuer-bond__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.issuer-bond__name-block {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px 8px;
  row-gap: 5px;
  min-width: 0;
  flex: 1 1 auto;
}
.issuer-bond__name {
  font: 600 13px / 1.3 var(--nla-font);
  color: var(--nla-text);
  letter-spacing: -0.005em;
  white-space: nowrap;
}
.issuer-bond__secid {
  font: 500 10.5px / 1 var(--nla-font-mono);
  color: var(--nla-text-muted);
  white-space: nowrap;
}
.issuer-bond__link {
  flex-shrink: 0;
  width: 26px;
  height: 26px;
  border-radius: 7px;
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  color: var(--nla-text-muted);
  display: grid;
  place-items: center;
  font-size: 13px;
  transition: background 120ms ease, color 120ms ease, border-color 120ms ease;
}

.issuer-bond__metrics {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0;
  border: 1px solid var(--nla-border);
  border-radius: 8px;
  background: var(--nla-bg-card);
  overflow: hidden;
}

.issuer-bond__dates {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 11.5px;
}

.issuer-bond__ai-row {
  display: grid;
  grid-template-columns: 84px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}
.issuer-bond__ai-lbl {
  font: 500 10px / 1 var(--nla-font);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  white-space: nowrap;
}
.issuer-bond__ai-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px 3px 6px;
  border-radius: 5px;
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
  justify-self: start;
}
[data-theme="dark"] .issuer-bond__ai-pill { color: var(--nla-primary); }
.issuer-bond__ai-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}
.issuer-bond__ai-num {
  font: 600 11px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
}
.issuer-bond__ai-max { opacity: 0.55; }

.issuer-bonds__toggle {
  appearance: none;
  width: 100%;
  border: 0;
  border-top: 1px solid var(--nla-border-light);
  background: var(--nla-bg-card);
  padding: 10px 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font: 500 12px/1 var(--nla-font);
  color: var(--nla-text-secondary);
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
}
.issuer-bonds__toggle:hover {
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
}
[data-theme="dark"] .issuer-bonds__toggle:hover { color: var(--nla-primary); }
.issuer-bonds__toggle .bi { font-size: 13px; }

/* injected mini-components — стили */
:deep(.metric) {
  padding: 7px 10px;
  border-right: 1px solid var(--nla-border-light);
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
:deep(.metric:last-child) { border-right: 0; }
:deep(.metric__lbl) {
  font: 500 9.5px / 1 var(--nla-font);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
}
:deep(.metric__val) {
  font: 600 13px / 1.1 var(--nla-font-mono);
  font-feature-settings: 'tnum', 'zero';
  color: var(--nla-text);
  letter-spacing: -0.01em;
}
:deep(.metric__val--success) { color: var(--nla-success); }
:deep(.metric__val--primary) { color: var(--nla-primary); }
:deep(.metric__val--muted)   { color: var(--nla-text-muted); }
:deep(.metric__sub) {
  font: 500 10px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text-muted);
  margin-top: 2px;
  letter-spacing: 0;
}

:deep(.date-line) {
  display: grid;
  grid-template-columns: 84px minmax(0, 1fr) auto;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}
:deep(.date-line__lbl) {
  font: 500 10px / 1 var(--nla-font);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--nla-text-muted);
  white-space: nowrap;
}
:deep(.date-line__val) {
  font: 500 12px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}
:deep(.date-line__val--empty) {
  color: var(--nla-text-muted);
  opacity: 0.55;
}
:deep(.date-line__rel) {
  font: 500 11px / 1 var(--nla-font-mono);
  font-feature-settings: 'tnum';
  color: var(--nla-text-muted);
  white-space: nowrap;
  justify-self: end;
}
:deep(.date-line__rel--soon) {
  color: var(--nla-warning);
  font-weight: 600;
}

@media (max-width: 480px) {
  .issuer-bond__metrics { grid-template-columns: repeat(2, 1fr); }
  :deep(.metric:nth-child(2)) { border-right: 0; }
  :deep(.metric:nth-child(n+3)) { border-top: 1px solid var(--nla-border-light); }
}
</style>
