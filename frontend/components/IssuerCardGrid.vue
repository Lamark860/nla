<template>
  <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
    <div
      v-for="issuer in issuers"
      :key="issuer.emitter_id"
      class="issuer-card"
    >
      <!-- Header -->
      <button
        class="issuer-card__header"
        @click="toggle(issuer.emitter_id)"
      >
        <div class="flex-1 min-w-0">
          <div class="issuer-card__name">{{ issuer.emitter_name }}</div>
          <div class="issuer-card__meta">
            <span class="issuer-card__id">ID {{ issuer.emitter_id }}</span>
            <span
              v-if="getIssuerRating(issuer.emitter_name)"
              :style="{ backgroundColor: scoreBadgeColor(getIssuerRating(issuer.emitter_name)!.score) }"
              class="issuer-card__rating"
              :title="getIssuerRating(issuer.emitter_name)!.ratings.map((r: any) => `${r.agency}: ${r.rating}`).join(', ')"
            >{{ getIssuerRating(issuer.emitter_name)!.ratings[0]?.rating }}-{{ getIssuerRating(issuer.emitter_name)!.score }}</span>
            <span
              v-for="cat in getCategories(issuer)"
              :key="cat"
              :class="categoryBadge(cat)"
              class="badge-sm"
            >{{ categoryShort(cat) }}</span>
            <span v-if="issuer.bonds.some((b: any) => b.is_float)" class="badge-sm bg-cyan-50 text-cyan-700 dark:bg-cyan-500/10 dark:text-cyan-400">Плав</span>
            <span class="issuer-card__count">{{ issuer.bond_count }} {{ bondWord(issuer.bond_count) }}</span>
          </div>
        </div>
        <svg
          :class="expanded.has(issuer.emitter_id) ? 'rotate-180' : ''"
          class="w-5 h-5 shrink-0 transition-transform" style="color: var(--nla-text-muted)"
          fill="none" viewBox="0 0 24 24" stroke="currentColor"
        ><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
      </button>

      <!-- Preview (collapsed) -->
      <div v-if="!expanded.has(issuer.emitter_id)" class="issuer-card__preview">
        {{ issuer.bonds[0]?.shortname }}<span v-if="issuer.bonds.length > 1"> и ещё {{ issuer.bonds.length - 1 }}</span>
      </div>

      <!-- Expanded bonds -->
      <div v-if="expanded.has(issuer.emitter_id)" class="issuer-card__bonds">
        <div
          v-for="bond in issuer.bonds"
          :key="bond.secid"
          class="issuer-card__bond"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-semibold text-sm" style="color: var(--nla-text)">{{ bond.shortname }}</span>
                <span class="font-mono text-[11px]" style="color: var(--nla-text-muted)">{{ bond.secid }}</span>
              </div>
              <div class="flex items-center gap-1.5 mt-1.5 flex-wrap">
                <span v-if="bond.bond_category" :class="categoryBadge(bond.bond_category)" class="badge-sm">{{ categoryShort(bond.bond_category) }}</span>
                <span class="badge-sm bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400">{{ bond.is_float ? 'Флоатер' : bond.is_indexed ? 'Индексируемая' : 'Фиксированный' }}</span>
                <span v-if="bond.coupon_period" class="badge-sm bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-400">{{ formatPeriod(bond.coupon_period) }}</span>
              </div>
              <div class="grid grid-cols-4 gap-x-4 gap-y-1 mt-3">
                <div>
                  <div class="text-[10px] uppercase" style="color: var(--nla-text-muted)">Цена:</div>
                  <div class="text-sm font-semibold font-mono" style="color: var(--nla-text)">{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</div>
                </div>
                <div>
                  <div class="text-[10px] uppercase" style="color: var(--nla-text-muted)">Дох:</div>
                  <div :class="yieldColor(bond.yield)" class="text-sm font-semibold font-mono">{{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}</div>
                </div>
                <div>
                  <div class="text-[10px] uppercase" style="color: var(--nla-text-muted)">НКД:</div>
                  <div class="text-sm font-semibold font-mono" style="color: var(--nla-text)">{{ fmt.priceRub(bond.accrued_int) }}</div>
                </div>
                <div>
                  <div class="text-[10px] uppercase" style="color: var(--nla-text-muted)">Купон:</div>
                  <div class="text-sm font-semibold font-mono" style="color: var(--nla-text)">{{ fmt.percent(bond.coupon_display) }}</div>
                </div>
              </div>
              <div class="mt-2 text-[11px]" style="color: var(--nla-text-muted)">
                Погашение: <span class="font-mono" style="color: var(--nla-text-secondary)">{{ fmt.date(bond.matdate) }}</span> ({{ fmt.daysToMaturity(bond.days_to_maturity) }})
              </div>
            </div>
            <NuxtLink :to="`/bonds/${bond.secid}`" class="shrink-0 w-8 h-8 rounded-lg flex items-center justify-center hover:bg-primary-50 dark:hover:bg-primary-500/10 transition-all" style="color: var(--nla-text-muted)">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><path stroke-linecap="round" d="M12 8v4m0 4h.01"/></svg>
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { IssuerGroup, IssuerRatingResponse } from '~/composables/useApi'

const fmt = useFormat()

const props = defineProps<{
  issuers: IssuerGroup[]
  ratings: Record<string, IssuerRatingResponse>
}>()

const expanded = ref(new Set<number>())

function toggle(id: number) {
  if (expanded.value.has(id)) {
    expanded.value.delete(id)
  } else {
    expanded.value.add(id)
  }
}

function getIssuerRating(name: string): IssuerRatingResponse | null {
  if (props.ratings[name]) return props.ratings[name]
  const lower = name.toLowerCase()
  for (const [key, val] of Object.entries(props.ratings)) {
    if (lower.includes(key.toLowerCase()) || key.toLowerCase().includes(lower)) return val
  }
  return null
}

function getCategories(issuer: IssuerGroup): string[] {
  return [...new Set(issuer.bonds.map(b => b.bond_category).filter(Boolean))]
}

function bondWord(n: number): string {
  const mod10 = n % 10, mod100 = n % 100
  if (mod100 >= 11 && mod100 <= 19) return 'облигаций'
  if (mod10 === 1) return 'облигация'
  if (mod10 >= 2 && mod10 <= 4) return 'облигации'
  return 'облигаций'
}

function scoreBadgeColor(score: number): string {
  const c: Record<number, string> = { 10: '#198754', 9: '#198754', 8: '#0dcaf0', 7: '#6edff6', 6: '#0d6efd', 5: '#6ea8fe', 4: '#ffc107', 3: '#fd7e14', 2: '#dc3545', 1: '#a70820', 0: '#000000' }
  return c[score] ?? '#6c757d'
}

function categoryBadge(cat: string): string {
  switch (cat) {
    case 'ОФЗ': return 'bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400'
    case 'Корпоративная': return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
    case 'Субфедеральная': return 'bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-400'
    case 'Муниципальная': return 'bg-orange-50 text-orange-700 dark:bg-orange-500/10 dark:text-orange-400'
    default: return 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400'
  }
}

function categoryShort(cat: string): string {
  switch (cat) {
    case 'ОФЗ': return 'ОФЗ'
    case 'Корпоративная': return 'Корп'
    case 'Субфедеральная': return 'Субф'
    case 'Муниципальная': return 'Мун'
    default: return cat
  }
}

function yieldColor(y: number | null | undefined): string {
  if (y == null) return ''
  if (y >= 15) return 'text-positive'
  if (y >= 10) return 'text-positive'
  return ''
}

function formatPeriod(days: number): string {
  if (days >= 27 && days <= 33) return '1 мес.'
  if (days >= 85 && days <= 95) return '3 мес.'
  if (days >= 175 && days <= 190) return '6 мес.'
  if (days >= 355 && days <= 370) return '12 мес.'
  return days + ' дн.'
}
</script>
