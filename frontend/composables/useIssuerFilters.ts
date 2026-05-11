// Filter logic for the redesigned IssuerFilters panel. Shared between
// pages/bonds/by-issuer.vue and pages/bonds/monthly.vue so both pages
// stay in sync with one source of truth.
//
// Field semantics mirror the IssuerFilters.vue emit shape:
//   search       — case-insensitive substring match against bond name/ISIN/SECID/issuer name
//   category     — 'corporate' | 'ofz' | 'sub' | 'mun' | ''
//   sector       — value from sectorOptions; not currently mapped against bond data (TODO)
//   rating       — 'AAA' | 'AA' | 'A' | 'BBB' | 'BB' | 'B_BELOW' | 'NONE' | ''
//   aiBucket     — '80+' | '60-80' | '40-60' | '<40' | ''
//   yield/coupon/duration — { min: string; max: string } numeric ranges (empty strings ignored)
//   tradeable    — bond.trading_status === 'T'
//   hasRating    — only issuers with at least one parsed rating
//   isFloat      — bond.is_float
//   hideMatured  — drop bonds where days_to_maturity ≤ 0

import type { Bond, AnalysisStats, IssuerRating } from './useApi'
import { maxRatingOrd, ordMatchesBucket, type RatingBucket } from './useRating'

export interface IssuerFilterState {
  search: string
  category: string
  sector: string
  rating: string
  aiBucket: string
  yield: { min: string; max: string }
  coupon: { min: string; max: string }
  duration: { min: string; max: string }
  tradeable: boolean
  hasRating: boolean
  isFloat: boolean
  hideMatured: boolean
}

export function emptyIssuerFilterState(): IssuerFilterState {
  return {
    search: '', category: '', sector: '', rating: '', aiBucket: '',
    yield:    { min: '', max: '' },
    coupon:   { min: '', max: '' },
    duration: { min: '', max: '' },
    tradeable: false, hasRating: false, isFloat: false, hideMatured: true,
  }
}

const categoryToBondCategory: Record<string, string> = {
  corporate: 'Корпоративная',
  ofz: 'ОФЗ',
  sub: 'Субфедеральная',
  mun: 'Муниципальная',
}

function num(s: string): number | null {
  if (s === '' || s == null) return null
  const n = Number(s)
  return Number.isFinite(n) ? n : null
}

function inRange(v: number | null | undefined, min: string, max: string): boolean {
  const lo = num(min), hi = num(max)
  if (lo == null && hi == null) return true
  if (v == null) return false
  if (lo != null && v < lo) return false
  if (hi != null && v > hi) return false
  return true
}

function matchesAiBucket(avg: number, bucket: string): boolean {
  switch (bucket) {
    case '80+':   return avg >= 80
    case '60-80': return avg >= 60 && avg < 80
    case '40-60': return avg >= 40 && avg < 60
    case '<40':   return avg < 40
    default:      return true
  }
}

const ratingBucketMap: Record<string, RatingBucket> = {
  AAA: 'aaa', AA: 'aa', A: 'a', BBB: 'bbb', BB: 'bb',
  B_BELOW: 'b_below', NONE: 'none',
}

export function matchesBond(bond: Bond, f: IssuerFilterState): boolean {
  if (f.search) {
    const q = f.search.toLowerCase()
    const fields = [bond.shortname, bond.secname, bond.isin, bond.secid].filter(Boolean)
    if (!fields.some(s => s.toLowerCase().includes(q))) return false
  }
  if (f.category) {
    const want = categoryToBondCategory[f.category]
    if (want && bond.bond_category !== want) return false
  }
  if (f.isFloat && !bond.is_float) return false
  if (f.tradeable && bond.trading_status !== 'T') return false
  if (f.hideMatured && bond.days_to_maturity != null && bond.days_to_maturity <= 0) return false
  if (!inRange(bond.yield, f.yield.min, f.yield.max)) return false
  if (!inRange(bond.coupon_percent, f.coupon.min, f.coupon.max)) return false
  if (!inRange(bond.days_to_maturity, f.duration.min, f.duration.max)) return false
  // sector — backend doesn't expose comparable codes yet; intentional no-op
  return true
}

export function matchesIssuerRating(
  ratings: IssuerRating[] | undefined,
  rating: string,
  hasRatingChip: boolean,
): boolean {
  const ord = maxRatingOrd(ratings ?? [])
  if (hasRatingChip && ord <= 0) return false
  if (!rating) return true
  const bucket = ratingBucketMap[rating]
  if (!bucket) return true
  return ordMatchesBucket(ord, bucket)
}

export function matchesIssuerAi(
  bonds: Bond[],
  aiStats: Record<string, AnalysisStats> | undefined,
  bucket: string,
): boolean {
  if (!bucket) return true
  if (!aiStats) return false
  let total = 0, sum = 0
  for (const b of bonds) {
    const s = aiStats[b.secid]
    if (s && s.avg_rating > 0) {
      sum += s.avg_rating * s.total
      total += s.total
    }
  }
  if (total === 0) return false
  return matchesAiBucket(sum / total, bucket)
}
