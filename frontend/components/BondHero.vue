<!--
  BondHero.vue — шапка карточки облигации.
  Заменяет визуальный hero-блок в pages/bonds/[secid].vue.

  Содержит:
   - имя бумаги + secid + ISIN
   - кредитные рейтинги эмитента (RatingBadge)
   - Аналитический индекс (AiScore)
   - теги: категория, "Флоатер", период купона, статус торгов (live)
   - 5 KPI: цена, доходность, купон, погашение, изменение
   - actions: избранное, share, запуск анализа
-->
<template>
  <section class="bond-hero">
    <!-- Top: name + ratings + actions -->
    <div class="bond-hero__top">
      <div class="bond-hero__title">
        <div class="bond-hero__name">{{ bond.shortname || bond.secname }}</div>
        <div class="bond-hero__ids">
          <span class="bond-hero__secid">{{ bond.secid }}</span>
          <span v-if="bond.isin" class="bond-hero__isin">{{ bond.isin }}</span>
          <NuxtLink v-if="bond.emitter_id" :to="`/issuers/${bond.emitter_id}?current=${bond.secid}`" class="bond-hero__issuer">
            <i class="bi bi-building" aria-hidden="true"></i>
            {{ issuerName || `Эмитент #${bond.emitter_id}` }}
          </NuxtLink>
        </div>
      </div>

      <div class="bond-hero__actions">
        <button v-if="auth?.isLoggedIn?.value"
                class="hero-action"
                :class="{ 'is-fav': isFavorite }"
                :title="isFavorite ? 'Убрать из избранного' : 'В избранное'"
                @click="$emit('toggle-favorite')">
          <i :class="isFavorite ? 'bi-star-fill' : 'bi-star'" class="bi" aria-hidden="true"></i>
        </button>
        <button class="hero-action" title="Скопировать ссылку" @click="$emit('share')">
          <i class="bi bi-link-45deg" aria-hidden="true"></i>
        </button>
        <button class="hero-action hero-action--primary" @click="$emit('analyze')">
          <i class="bi bi-stars" aria-hidden="true"></i>
          <span>Анализ</span>
        </button>
      </div>
    </div>

    <!-- Tags row: rating badges, three scoring profiles, category/status/period/risk -->
    <div class="bond-hero__tags">
      <RatingBadge v-for="r in ratings" :key="r.agency" :rating="r.rating" :agency="r.agency" />

      <!-- Phase 3 — three deterministic scores, active profile highlighted.
           Click → switch the global profile + jump to the scoring tab. -->
      <template v-if="scores && scores.length">
        <button
          v-for="s in scoresOrdered"
          :key="s.profile_code"
          class="hero-score"
          :class="{ 'is-active': s.profile_code === scoring.profile.value }"
          :title="profileTitle(s)"
          @click="onScoreClick(s.profile_code)"
        >
          <span class="hero-score__icon" aria-hidden="true">{{ scoring.metaMap[s.profile_code as any]?.icon || '•' }}</span>
          <span class="hero-score__val" :style="fmt.aiRatingStyleSoft(s.score)">{{ Math.round(s.score) }}</span>
        </button>
      </template>

      <!-- Legacy LLM-analysis aggregate score; kept while the old tab lives. -->
      <AiScore v-else-if="aiScore != null" :value="aiScore" />

      <span class="bond-hero__divider" v-if="ratings.length || aiScore != null || (scores && scores.length)"></span>

      <Tag v-if="categoryLabel" :tone="bond.is_float ? 'primary' : 'default'">{{ categoryLabel }}</Tag>
      <Tag v-if="periodLabel">{{ periodLabel }}</Tag>
      <Pill v-if="bond.trading_status === 'T'" tone="live">Торги идут</Pill>
      <Pill v-else-if="bond.trading_status === 'S'" tone="warning">Приостановлены</Pill>
      <Pill v-else tone="default">Не торгуется</Pill>

      <Pill v-if="bond.is_near_offer" tone="warning">
        <i class="bi bi-flag-fill" aria-hidden="true"></i>Скоро оферта
      </Pill>

      <span v-if="bond.risk_category" class="bond-hero__risk" :class="`bond-hero__risk--${bond.risk_category}`">
        Риск: {{ riskLabel }}
      </span>
    </div>

    <!-- KPI row -->
    <div class="bond-hero__kpis">
      <KPI label="Цена" :value="bond.last != null ? fmt.percent(bond.last) : '—'">
        <template #sub v-if="bond.price_rub">
          {{ fmt.priceRub(bond.price_rub) }}<template v-if="bond.mid_price_pct != null"> · mid {{ fmt.percent(bond.mid_price_pct) }}</template>
        </template>
      </KPI>

      <KPI label="Доходность"
           :value="bond.yield != null ? fmt.percent(bond.yield) : '—'"
           :tone="yieldTone">
        <template #sub v-if="bond.duration">дюрация {{ bond.duration }}</template>
      </KPI>

      <KPI label="Купон" :value="fmt.percent(bond.coupon_display)">
        <template #sub v-if="bond.next_coupon || bond.coupon_value">
          <template v-if="bond.next_coupon">след. {{ fmt.dateShort(bond.next_coupon) }}</template>
          <template v-if="bond.coupon_value && bond.next_coupon"> · </template>
          <template v-if="bond.coupon_value">{{ fmt.priceRub(bond.coupon_value) }}</template>
        </template>
      </KPI>

      <KPI label="Погашение"
           :value="fmt.dateShort(bond.matdate)"
           tone="muted">
        <template #sub v-if="bond.days_to_maturity != null"
                  :sub-tone="bond.days_to_maturity < 365 ? 'danger' : 'muted'">
          {{ fmt.daysToMaturity(bond.days_to_maturity) }}
        </template>
      </KPI>

      <KPI label="Изменение"
           :value="formatChange(bond.last_change_prcnt)"
           :tone="changeTone">
        <template #sub>
          <template v-if="bond.last_change != null">{{ formatChangePoints(bond.last_change) }}</template>
          <template v-if="bond.last_change != null && bond.value_today_rub"> · </template>
          <template v-if="bond.value_today_rub">оборот {{ fmt.volume(bond.value_today_rub) }}</template>
        </template>
      </KPI>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Bond, ScoreResponse } from '~/composables/useApi'
import type { ProfileCode } from '~/composables/useScoringProfile'
import RatingBadge from './RatingBadge.vue'
import AiScore from './AiScore.vue'
import Tag from './Tag.vue'
import Pill from './Pill.vue'
import KPI from './KPI.vue'

const props = defineProps<{
  bond: Bond
  ratings?: Array<{ agency: string; rating: string }>
  aiScore?: number | null
  scores?: ScoreResponse[] | null
  issuerName?: string
  isFavorite?: boolean
}>()
const emit = defineEmits<{
  'toggle-favorite': []
  'share': []
  'analyze': []
  'open-score': []
}>()

const fmt = useFormat()
const auth = useAuth()
const scoring = useScoringProfile()

const profileOrder: ProfileCode[] = ['low', 'mid', 'high']
const scoresOrdered = computed(() => {
  if (!props.scores) return []
  return [...props.scores].sort((a, b) =>
    profileOrder.indexOf(a.profile_code as ProfileCode) - profileOrder.indexOf(b.profile_code as ProfileCode))
})

function profileTitle(s: ScoreResponse): string {
  const name = s.profile_name || scoring.metaMap[s.profile_code as ProfileCode]?.label || s.profile_code
  return `${name}: ${Math.round(s.score)}/100`
}

// Click on a hero score: pick that profile as the global active one and
// jump to the dedicated breakdown tab. The page owns the tab state, so
// we hand control over via an event.
function onScoreClick(code: string) {
  if (profileOrder.includes(code as ProfileCode)) scoring.set(code as ProfileCode)
  emit('open-score')
}

const ratings = computed(() => (props.ratings || []).filter(r => r.rating && r.rating !== 'NULL'))

const categoryLabel = computed(() => {
  if (props.bond.is_float) return 'Флоатер'
  if (props.bond.is_indexed) return 'Индексируемая'
  return props.bond.bond_category || 'Корпоративная'
})

const periodLabel = computed(() => {
  const p = props.bond.coupon_period
  if (!p) return ''
  if (p >= 27 && p <= 33)   return 'Ежемесячный купон'
  if (p >= 85 && p <= 95)   return 'Квартальный купон'
  if (p >= 175 && p <= 190) return 'Полугодовой купон'
  if (p >= 355 && p <= 370) return 'Годовой купон'
  return `Купон ${p} дн.`
})

const yieldTone = computed(() => {
  const y = props.bond.yield
  if (y == null) return 'muted'
  if (y >= 12) return 'success'
  if (y >= 8)  return 'primary'
  return 'default'
})

const changeTone = computed(() => {
  const v = props.bond.last_change_prcnt
  if (v == null || v === 0) return 'muted'
  return v > 0 ? 'success' : 'danger'
})

function formatChange(v: number | null | undefined): string {
  if (v == null) return '—'
  if (v === 0)   return '0,00%'
  return (v > 0 ? '+' : '') + v.toFixed(2).replace('.', ',') + '%'
}

function formatChangePoints(v: number | null | undefined): string {
  if (v == null) return '—'
  if (v === 0)   return '0,00 п.п.'
  return (v > 0 ? '+' : '') + v.toFixed(2).replace('.', ',') + ' п.п.'
}

const riskLabel = computed(() => {
  const map: Record<string, string> = {
    safe: 'низкий',
    stable: 'стабильный',
    moderate: 'умеренный',
    speculative: 'спекулятивный',
    risky: 'высокий',
    toxic: 'токсичный',
    junk: 'мусорный',
  }
  return map[props.bond.risk_category] || props.bond.risk_category
})
</script>

<style scoped>
.bond-hero {
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-2xl);
  padding: 28px 32px 24px;
  box-shadow: var(--nla-shadow-sm);
  position: relative;
  overflow: hidden;
}
/* лёгкий violet-glow в углу — узнаваемая деталь */
.bond-hero::before {
  content: '';
  position: absolute;
  inset: -40% -20% auto auto;
  width: 380px;
  height: 380px;
  background: radial-gradient(circle, color-mix(in oklab, var(--nla-primary) 14%, transparent) 0%, transparent 65%);
  pointer-events: none;
}

.bond-hero__top {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;
  position: relative;
}
.bond-hero__name {
  font: 700 24px / 1.15 var(--nla-font);
  letter-spacing: -0.02em;
  color: var(--nla-text);
  word-break: break-word;
}
.bond-hero__ids {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
  font: 500 13px / 1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.bond-hero__secid,
.bond-hero__isin {
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  letter-spacing: 0;
}
.bond-hero__secid { color: var(--nla-text); font-weight: 600; }
.bond-hero__issuer {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--nla-primary);
  text-decoration: none;
  font-weight: 500;
}
.bond-hero__issuer:hover { color: var(--nla-primary-hover); text-decoration: underline; }
.bond-hero__issuer i { font-size: 13px; }

.bond-hero__actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.hero-action {
  appearance: none;
  border: 1px solid var(--nla-border);
  background: var(--nla-bg-card);
  color: var(--nla-text-secondary);
  width: 36px;
  height: 36px;
  border-radius: var(--nla-radius);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease, border-color 120ms ease;
}
.hero-action:hover {
  background: var(--nla-bg-hover);
  color: var(--nla-text);
  border-color: var(--nla-border-strong);
}
.hero-action.is-fav { color: var(--nla-warning); border-color: color-mix(in oklab, var(--nla-warning) 30%, var(--nla-border)); }
.hero-action--primary {
  width: auto;
  padding: 0 14px;
  background: var(--nla-primary);
  border-color: var(--nla-primary);
  color: var(--nla-text-inverse);
  font: 600 13px / 1 var(--nla-font);
  gap: 6px;
}
.hero-action--primary:hover {
  background: var(--nla-primary-hover);
  border-color: var(--nla-primary-hover);
  color: var(--nla-text-inverse);
}

.bond-hero__tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 24px;
  position: relative;
}
.bond-hero__divider {
  width: 1px;
  height: 16px;
  background: var(--nla-border);
  margin: 0 4px;
}

/* Three-score badge cluster on the hero. Active profile gets the
   highlighted outline so the user always sees which lens is current. */
.hero-score {
  appearance: none;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 4px 2px 6px;
  border-radius: var(--nla-radius-pill);
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  font: 600 11.5px/1.4 var(--nla-font);
  color: var(--nla-text-secondary);
  transition: border-color 120ms ease, background 120ms ease;
}
.hero-score:hover {
  border-color: var(--nla-border-strong);
  background: var(--nla-bg-subtle);
}
.hero-score.is-active {
  border-color: color-mix(in oklab, var(--nla-primary) 45%, transparent);
  background: var(--nla-primary-light);
  color: var(--nla-primary-ink);
}
[data-theme="dark"] .hero-score.is-active { color: var(--nla-primary); }
.hero-score__icon { font-size: 13px; line-height: 1; }
.hero-score__val {
  padding: 2px 8px;
  border-radius: var(--nla-radius-pill);
  font-family: var(--nla-font-mono);
  font-feature-settings: 'tnum';
  font-weight: 700;
  font-size: 12px;
}

.bond-hero__risk {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: var(--nla-radius-pill);
  font: 600 11px/1 var(--nla-font);
  letter-spacing: 0.02em;
  white-space: nowrap;
}
.bond-hero__risk::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}
.bond-hero__risk--safe,
.bond-hero__risk--stable      { background: var(--nla-success-light); color: var(--nla-success); }
.bond-hero__risk--moderate    { background: var(--nla-primary-light); color: var(--nla-primary-ink); }
.bond-hero__risk--speculative,
.bond-hero__risk--risky       { background: var(--nla-warning-light); color: var(--nla-warning); }
.bond-hero__risk--toxic,
.bond-hero__risk--junk        { background: var(--nla-danger-light); color: var(--nla-danger); }
[data-theme="dark"] .bond-hero__risk--moderate { color: var(--nla-primary); }

.bond-hero__kpis {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 28px;
  position: relative;
}

@media (max-width: 992px) {
  .bond-hero { padding: 22px 20px 18px; }
  .bond-hero__kpis { grid-template-columns: repeat(2, 1fr); gap: 18px 24px; }
  .bond-hero__name { font-size: 20px; }
}
@media (max-width: 480px) {
  .bond-hero__top { flex-direction: column; }
  .bond-hero__actions { align-self: flex-end; }
  .bond-hero__kpis { grid-template-columns: 1fr 1fr; }
}
</style>
