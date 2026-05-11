// Credit-rating normaliser — TS port of internal/service/rating_score.go.
// Keep the two implementations behaviourally identical: the Go test table
// (rating_score_test.go) and the TS test table (useRating.spec.ts) cover the
// same agency formats so behavioural drift is caught immediately.

export type RatingTier = 'AAA' | 'AA' | 'A' | 'BBB' | 'BB' | 'B' | 'CCC' | 'CC' | 'C' | 'D' | ''

const ordToTier: Record<number, RatingTier> = {
  22: 'AAA',
  21: 'AA', 20: 'AA', 19: 'AA',
  18: 'A', 17: 'A', 16: 'A',
  15: 'BBB', 14: 'BBB', 13: 'BBB',
  12: 'BB', 11: 'BB', 10: 'BB',
  9: 'B', 8: 'B', 7: 'B',
  6: 'CCC', 5: 'CCC', 4: 'CCC',
  3: 'CC',
  2: 'C',
  1: 'D',
}

const moodysMap: Record<string, number> = {
  Aaa: 22,
  Aa1: 21, Aa2: 20, Aa3: 19,
  A1: 18, A2: 17, A3: 16,
  Baa1: 15, Baa2: 14, Baa3: 13,
  Ba1: 12, Ba2: 11, Ba3: 10,
  B1: 9, B2: 8, B3: 7,
  Caa1: 6, Caa2: 5, Caa3: 4,
  Ca: 3, C: 2,
}

const letterMap: Record<string, number> = {
  AAA: 22,
  'AA+': 21, AA: 20, 'AA-': 19,
  'A+': 18, A: 17, 'A-': 16,
  'BBB+': 15, BBB: 14, 'BBB-': 13,
  'BB+': 12, BB: 11, 'BB-': 10,
  'B+': 9, B: 8, 'B-': 7,
  'CCC+': 6, CCC: 5, 'CCC-': 4,
  CC: 3, C: 2,
  D: 1, RD: 1,
}

const moodysRe = /^(Aaa|Aa[1-3]|A[1-3]|Baa[1-3]|Ba[1-3]|B[1-3]|Caa[1-3]|Ca|C)$/
const letterRe = /^(AAA|AA[+-]?|A[+-]?|BBB[+-]?|BB[+-]?|B[+-]?|CCC[+-]?|CC|C|D|RD)$/

const dohodRoundTrip: Record<number, number> = {
  10: 22, 9: 21, 8: 20, 7: 19, 6: 18, 5: 17, 4: 15, 3: 13, 2: 11, 1: 9,
}

const outlookWords = [
  'стабильный', 'позитивный', 'негативный', 'развивающийся',
  'стаб.', 'позит.', 'негат.', 'разв.',
  'stable', 'positive', 'negative', 'developing',
]

export interface NormalizedRating {
  ord: number
  tier: RatingTier
}

export function normalizeRating(text: string | null | undefined): NormalizedRating {
  let s = (text ?? '').trim()
  if (!s) return { ord: 0, tier: '' }

  const low = s.toLowerCase()
  if (low === 'отозван' || low === 'отозвано' || low.startsWith('wd') || s === '—') {
    return { ord: 0, tier: '' }
  }

  const dohod = parseDohodNumeric(s)
  if (dohod > 0) return { ord: dohod, tier: ordToTier[dohod] ?? '' }

  s = stripOutlook(s)
  s = stripNationalSuffix(s)

  // "ru"/"Ru"/"RU" prefix used by Эксперт РА and some MOEX CCI variants
  if (s.length > 2 && s.slice(0, 2).toLowerCase() === 'ru') {
    s = s.slice(2)
  }
  s = s.trim()

  // Try Moody's (case-sensitive: only "Aaa" / "Baa1" patterns) first
  if (moodysRe.test(s)) {
    const ord = moodysMap[s] ?? 0
    if (ord > 0) return { ord, tier: ordToTier[ord] ?? '' }
  }

  const upper = s.toUpperCase()
  if (letterRe.test(upper)) {
    const ord = letterMap[upper] ?? 0
    if (ord > 0) return { ord, tier: ordToTier[ord] ?? '' }
  }

  return { ord: 0, tier: '' }
}

// LegacyScore10 maps the 22-level ordinal back to the 1-10 scale used by
// older API/frontend code paths. Mirrors internal/service/rating_score.go.
export function legacyScore10(ord: number): number {
  if (ord >= 22) return 10
  if (ord === 21) return 9
  if (ord === 20) return 8
  if (ord === 19) return 7
  if (ord === 18) return 6
  if (ord === 17 || ord === 16) return 5
  if (ord === 15 || ord === 14) return 4
  if (ord === 13 || ord === 12) return 3
  if (ord === 11 || ord === 10) return 2
  if (ord >= 7) return 1
  return 0
}

function parseDohodNumeric(s: string): number {
  let clean = s.trim()
  const slash = clean.indexOf('/')
  if (slash > 0) clean = clean.slice(0, slash).trim()
  if (!/^\d+$/.test(clean)) return 0
  const n = parseInt(clean, 10)
  if (n < 1 || n > 10) return 0
  return dohodRoundTrip[n] ?? 0
}

function stripNationalSuffix(s: string): string {
  const paren = s.indexOf('(')
  if (paren > 0) s = s.slice(0, paren)
  for (const suf of ['.ru', '.RU', '|ru|', '|RU|']) {
    if (s.endsWith(suf)) s = s.slice(0, -suf.length)
  }
  return s.trim()
}

function stripOutlook(s: string): string {
  const low = s.toLowerCase()
  for (const w of outlookWords) {
    const i = low.indexOf(w)
    if (i > 0) return s.slice(0, i).trim()
  }
  return s
}

// Inline-style colour for a rating chip. Uses the canonical tier from
// normalizeRating so Moody's `Baa1` and S&P `BBB+` get the same colour.
export function ratingTierStyle(rating: string | null | undefined): { background: string; color: string } {
  const r = (rating ?? '').trim()
  if (!r) return { background: 'rgba(108,117,125,0.12)', color: 'var(--nla-text)' }

  const low = r.toLowerCase()
  if (low === 'отозван' || low === 'отозвано' || low.startsWith('wd')) {
    return { background: 'rgba(108,117,125,0.12)', color: 'var(--nla-text-muted)' }
  }

  const { tier } = normalizeRating(r)
  switch (tier) {
    case 'AAA': return { background: 'rgba(25,135,84,0.14)', color: '#198754' }
    case 'AA':  return { background: 'rgba(13,110,253,0.14)', color: '#0d6efd' }
    case 'A':   return { background: 'rgba(32,201,151,0.14)', color: '#0d9488' }
    case 'BBB': return { background: 'rgba(255,193,7,0.14)', color: '#997404' }
    case 'BB':  return { background: 'rgba(253,126,20,0.14)', color: '#e8590c' }
    case 'B':   return { background: 'rgba(220,53,69,0.14)', color: '#dc3545' }
    case 'CCC':
    case 'CC':
    case 'C':
    case 'D':   return { background: 'rgba(165,29,42,0.18)', color: '#a51d2a' }
    default:    return { background: 'rgba(108,117,125,0.12)', color: 'var(--nla-text)' }
  }
}

// Highest 22-level ordinal across a list of ratings. Frontend filters use this
// so AA+ vs AA- and BBB- vs BB+ stay distinguishable — the legacy 1-10 score
// collapses those pairs, the ordinal does not.
//
// Records lacking score_ord are normalised on the fly so the helper still works
// against responses from older API versions.
export function maxRatingOrd(ratings: Array<{ rating?: string; score_ord?: number }> | null | undefined): number {
  if (!ratings || ratings.length === 0) return 0
  let max = 0
  for (const r of ratings) {
    if (!r.rating || r.rating === 'NULL') continue
    const ord = (r.score_ord && r.score_ord > 0)
      ? r.score_ord
      : normalizeRating(r.rating).ord
    if (ord > max) max = ord
  }
  return max
}

// Filter buckets used by IssuerFilters.vue. `ord` is the canonical 1-22 scale.
export type RatingBucket = 'aaa' | 'aa' | 'a' | 'bbb' | 'bb' | 'b_below' | 'none' | ''

export function ordMatchesBucket(ord: number, bucket: RatingBucket): boolean {
  if (!bucket) return true
  switch (bucket) {
    case 'aaa':     return ord === 22
    case 'aa':      return ord >= 19 && ord <= 21
    case 'a':       return ord >= 16 && ord <= 18
    case 'bbb':     return ord >= 13 && ord <= 15
    case 'bb':      return ord >= 10 && ord <= 12
    case 'b_below': return ord >= 1 && ord <= 9
    case 'none':    return ord <= 0
    default:        return true
  }
}

// useRating exposes the helpers as a Nuxt composable for ergonomic imports
// (parallels useFormat()). The pure functions are also re-exported above.
export function useRating() {
  return { normalizeRating, legacyScore10, ratingTierStyle, maxRatingOrd, ordMatchesBucket }
}
