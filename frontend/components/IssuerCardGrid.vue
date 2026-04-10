<template>
  <div class="row g-3">
    <div
      v-for="issuer in issuers"
      :key="issuer.emitter_id"
      class="col-md-6 col-lg-4"
    >
      <div class="card">
        <!-- Card header: issuer name + badges -->
        <div class="card-header" style="cursor: pointer" @click="toggle(issuer.emitter_id)">
          <div class="d-flex justify-content-between align-items-start mb-2" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
            <strong style="overflow: hidden; text-overflow: ellipsis">{{ issuer.emitter_name || `Эмитент #${issuer.emitter_id}` }}</strong>
          </div>
          <div class="d-flex flex-wrap align-items-center gap-1">
            <span class="badge bg-secondary font-monospace" style="font-size: 11px; padding: 3px 6px">{{ issuer.emitter_id }}</span>
            <template v-if="getIssuerRating(issuer.emitter_id)">
              <span
                v-for="r in getIssuerRating(issuer.emitter_id)!.ratings.filter(x => x.rating && x.rating !== 'NULL')"
                :key="r.agency"
                class="badge font-monospace fw-bold"
                :style="fmt.ratingChipStyle(r.rating)"
                :title="r.agency"
                style="font-size: 11px; padding: 3px 6px"
              >{{ r.rating }}</span>
            </template>
            <span v-else class="badge" style="font-size: 11px; padding: 3px 6px; background: rgba(108,117,125,0.08); color: var(--nla-text-muted)">—</span>
            <template v-if="getIssuerAiStats(issuer)">
              <span class="badge fw-semibold" :style="fmt.aiRatingStyleSoft(getIssuerAiStats(issuer)!.avg)" :title="`AI: средний балл ${getIssuerAiStats(issuer)!.avg.toFixed(0)}, анализов ${getIssuerAiStats(issuer)!.total}`">
                🤖 {{ getIssuerAiStats(issuer)!.min === getIssuerAiStats(issuer)!.max ? getIssuerAiStats(issuer)!.avg.toFixed(0) : `${getIssuerAiStats(issuer)!.min}–${getIssuerAiStats(issuer)!.max}` }}
              </span>
            </template>
            <span class="ms-auto" style="color: var(--nla-text)">
              <i :class="expanded.has(issuer.emitter_id) ? 'bi-chevron-up' : 'bi-chevron-down'" class="bi"></i>
            </span>
          </div>
          <small class="text-muted d-block mt-1">{{ issuer.bond_count }} {{ bondWord(issuer.bond_count) }}</small>
        </div>

        <!-- Preview (collapsed) -->
        <div v-if="!expanded.has(issuer.emitter_id) && issuer.bonds[0]" class="card-body py-2 border-bottom">
          <small class="text-muted">{{ issuer.bonds[0].shortname }}<span v-if="issuer.bonds.length > 1"> и ещё {{ issuer.bonds.length - 1 }}</span></small>
        </div>

        <!-- Expanded bonds -->
        <template v-if="expanded.has(issuer.emitter_id)">
          <div
            v-for="bond in issuer.bonds"
            :key="bond.secid"
            class="card-body py-2 border-top"
            style="background: var(--nla-bg-card)"
          >
            <div class="d-flex justify-content-between align-items-center">
              <div class="flex-grow-1 me-2" style="min-width: 0">
                <!-- Name + SECID + badges -->
                <div class="d-flex align-items-center flex-wrap gap-1 mb-1">
                  <strong class="me-1 text-truncate">{{ bond.shortname }}</strong>
                  <small class="text-muted flex-shrink-0 me-1">{{ bond.secid }}</small>
                  <span v-if="bond.coupon_period" class="badge border text-muted" style="font-size: 10px">{{ formatPeriod(bond.coupon_period) }}</span>
                  <span v-if="bond.is_float" class="badge bg-info" style="font-size: 10px">Флоатер</span>
                  <span v-if="bond.is_indexed" class="badge bg-secondary" style="font-size: 10px">Индексируемая</span>
                </div>
                <!-- Financial metrics -->
                <div class="issuer-bond-metrics mb-1">
                  <div>
                    <span class="metric-label">Цена</span>
                    <span class="metric-value">{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</span>
                  </div>
                  <div>
                    <span class="metric-label">Дох</span>
                    <span class="metric-value" :class="bond.yield != null && bond.yield > 12 ? 'text-success' : 'text-primary'">{{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}</span>
                  </div>
                  <div>
                    <span class="metric-label">НКД</span>
                    <span class="metric-value">{{ bond.accrued_int != null ? fmt.num(bond.accrued_int, 2) : '—' }}</span>
                  </div>
                  <div>
                    <span class="metric-label">Купон</span>
                    <span class="metric-value">{{ fmt.percent(bond.coupon_display) }}</span>
                  </div>
                </div>
                <!-- Dates -->
                <div class="issuer-bond-dates">
                  <div>
                    <span class="date-label">Оферта:</span>
                    <template v-if="isValidDate(bond.putoptiondate || bond.buybackdate)">
                      <strong>{{ fmt.dateShort(bond.putoptiondate || bond.buybackdate) }}</strong>
                      <small class="text-muted ms-1">({{ daysUntil(bond.putoptiondate || bond.buybackdate) }})</small>
                    </template>
                    <span v-else class="text-muted">—</span>
                  </div>
                  <div>
                    <span class="date-label">Колл-опцион:</span>
                    <template v-if="isValidDate(bond.calloptiondate)">
                      <strong>{{ fmt.dateShort(bond.calloptiondate) }}</strong>
                      <small class="text-muted ms-1">({{ daysUntil(bond.calloptiondate) }})</small>
                    </template>
                    <span v-else class="text-muted">—</span>
                  </div>
                  <div>
                    <span class="date-label">Погашение:</span>
                    <strong>{{ fmt.dateShort(bond.matdate) }}</strong>
                    <small class="text-muted ms-1">({{ daysLabel(bond.days_to_maturity) }})</small>
                  </div>
                  <div v-if="getBondAiRating(bond.secid)">
                    <span class="date-label">AI рейтинг:</span>
                    <span class="badge fw-semibold" style="font-size: 11px" :style="fmt.aiRatingStyleSoft(getBondAiRating(bond.secid))">🤖 {{ getBondAiRating(bond.secid) }}</span>
                  </div>
                </div>
              </div>
              <!-- Detail link -->
              <NuxtLink :to="`/bonds/${bond.secid}`" class="btn btn-sm btn-outline-secondary" title="Подробнее">
                <i class="bi bi-info-circle"></i>
              </NuxtLink>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { IssuerGroup, IssuerRatingResponse, AnalysisStats } from '~/composables/useApi'

const fmt = useFormat()

const props = defineProps<{
  issuers: IssuerGroup[]
  ratings: Record<string, IssuerRatingResponse>
  aiStats?: Record<string, AnalysisStats>
}>()

const expanded = ref(new Set<number>())

function toggle(id: number) {
  if (expanded.value.has(id)) expanded.value.delete(id)
  else expanded.value.add(id)
}

function getIssuerRating(emitterId: number): IssuerRatingResponse | null {
  const key = String(emitterId)
  return props.ratings[key] ?? null
}

function getIssuerAiStats(issuer: IssuerGroup): { total: number; avg: number; min: number; max: number } | null {
  if (!props.aiStats) return null
  let total = 0
  let ratingMin = Infinity
  let ratingMax = -Infinity
  let ratingSum = 0
  let ratingCount = 0
  for (const bond of issuer.bonds) {
    const stats = props.aiStats[bond.secid]
    if (stats) {
      total += stats.total
      if (stats.avg_rating > 0) {
        ratingSum += stats.avg_rating * stats.total
        ratingCount += stats.total
        if (stats.avg_rating < ratingMin) ratingMin = stats.avg_rating
        if (stats.avg_rating > ratingMax) ratingMax = stats.avg_rating
      }
    }
  }
  if (total === 0) return null
  const avg = ratingCount > 0 ? ratingSum / ratingCount : 0
  return { total, avg, min: ratingMin === Infinity ? 0 : Math.round(ratingMin), max: ratingMax === -Infinity ? 0 : Math.round(ratingMax) }
}

function bondWord(n: number): string {
  const mod10 = n % 10, mod100 = n % 100
  if (mod100 >= 11 && mod100 <= 19) return 'облигаций'
  if (mod10 === 1) return 'облигация'
  if (mod10 >= 2 && mod10 <= 4) return 'облигации'
  return 'облигаций'
}



function daysUntil(dateStr: string): string {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return '—'
  const now = new Date()
  const diff = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return 'прошло'
  return `${diff} дн.`
}

function daysLabel(days: number | null | undefined): string {
  if (days == null) return '—'
  if (days < 0) return 'погашена'
  return `${days} дн.`
}

function isValidDate(val: string | null | undefined): boolean {
  if (!val || val === '0000-00-00' || val === 'None') return false
  const d = new Date(val)
  return !isNaN(d.getTime()) && d.getFullYear() > 1970
}

function formatPeriod(days: number): string {
  if (days >= 27 && days <= 33) return '1 мес.'
  if (days >= 85 && days <= 95) return '3 мес.'
  if (days >= 175 && days <= 190) return '6 мес.'
  if (days >= 355 && days <= 370) return '12 мес.'
  return days + ' дн.'
}

function getBondAiRating(secid: string): number | null {
  if (!props.aiStats) return null
  const stats = props.aiStats[secid]
  if (!stats || stats.avg_rating <= 0) return null
  return Math.round(stats.avg_rating)
}




</script>
