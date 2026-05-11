<!--
  IssuerProfile.vue — шапка профиля эмитента (страница /issuers/:id).
  Имя + ИНН/ОГРН + рейтинги + AI-балл + KPI (выпуск, среднее YTM, дюрация, сумма долга).
-->
<template>
  <section class="ip">
    <div class="ip__head">
      <div class="ip__title">
        <h1 class="ip__name">{{ name || `Эмитент #${emitterId}` }}</h1>
        <div class="ip__ids">
          <span v-if="inn">ИНН <strong class="num">{{ inn }}</strong></span>
          <span v-if="ogrn">ОГРН <strong class="num">{{ ogrn }}</strong></span>
          <span v-if="sector" class="ip__sector">{{ sector }}</span>
        </div>
      </div>
      <a v-if="website" :href="website" target="_blank" rel="noopener" class="ip__link">
        <i class="bi bi-box-arrow-up-right" aria-hidden="true"></i>
        {{ websiteLabel }}
      </a>
    </div>

    <div class="ip__chips">
      <RatingBadge v-for="r in ratings" :key="r.agency" :rating="r.rating" :agency="r.agency" />
      <AiScore v-if="aiScore != null" :value="aiScore" />
      <Tag v-if="quality">Качество · {{ quality }}/10</Tag>
    </div>

    <div class="ip__kpis">
      <KPI label="Бумаг в обращении" :value="bondCount" />
      <KPI label="Средняя доходность" :value="avgYield != null ? fmt.percent(avgYield) : '—'" tone="success" />
      <KPI label="Средняя дюрация" :value="avgDuration ? `${avgDuration} дн` : '—'" />
      <KPI label="Объём долга" :value="totalDebt ? fmt.priceRub(totalDebt) : '—'" tone="primary" />
    </div>
  </section>
</template>

<script setup lang="ts">
import RatingBadge from './RatingBadge.vue'
import AiScore from './AiScore.vue'
import Tag from './Tag.vue'
import KPI from './KPI.vue'

const props = defineProps<{
  emitterId: number | string
  name?: string
  inn?: string
  ogrn?: string
  sector?: string
  website?: string
  ratings?: Array<{ agency: string; rating: string }>
  aiScore?: number | null
  quality?: number | null
  bondCount?: number | string
  avgYield?: number | null
  avgDuration?: number | null
  totalDebt?: number | null
}>()
const fmt = useFormat()

const websiteLabel = computed(() => {
  if (!props.website) return ''
  try { return new URL(props.website).host } catch { return props.website }
})
</script>

<style scoped>
.ip {
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius-2xl);
  padding: 24px 28px 22px;
  box-shadow: var(--nla-shadow-sm);
  position: relative;
  overflow: hidden;
}
.ip::before {
  content: '';
  position: absolute;
  inset: -50% -10% auto auto;
  width: 360px;
  height: 360px;
  background: radial-gradient(circle, color-mix(in oklab, var(--nla-primary) 12%, transparent) 0%, transparent 60%);
  pointer-events: none;
}

.ip__head {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
  position: relative;
}
.ip__title { flex: 1 1 auto; min-width: 0; }
.ip__name {
  margin: 0;
  font: 700 22px / 1.2 var(--nla-font);
  letter-spacing: -0.02em;
  color: var(--nla-text);
}
.ip__ids {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 6px;
  font: 500 12.5px / 1.4 var(--nla-font);
  color: var(--nla-text-muted);
}
.ip__ids .num { font-family: var(--nla-font-mono); color: var(--nla-text); font-weight: 600; }
.ip__sector {
  padding: 1px 8px;
  background: var(--nla-bg-subtle);
  border-radius: var(--nla-radius-sm);
  color: var(--nla-text-secondary);
}
.ip__link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--nla-bg-subtle);
  border: 1px solid var(--nla-border);
  border-radius: var(--nla-radius);
  font: 500 12.5px / 1 var(--nla-font);
  color: var(--nla-text-secondary);
  text-decoration: none;
  flex-shrink: 0;
}
.ip__link:hover { background: var(--nla-primary-light); color: var(--nla-primary-ink); border-color: color-mix(in oklab, var(--nla-primary) 25%, var(--nla-border)); }

.ip__chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 22px;
  position: relative;
}

.ip__kpis {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
  position: relative;
}
@media (max-width: 768px) {
  .ip { padding: 20px 18px; }
  .ip__kpis { grid-template-columns: repeat(2, 1fr); gap: 18px; }
}
</style>
