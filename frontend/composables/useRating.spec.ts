import { describe, it, expect } from 'vitest'
import {
  normalizeRating,
  legacyScore10,
  ratingTierStyle,
  maxRatingOrd,
  ordMatchesBucket,
} from './useRating'

// The cases below mirror internal/service/rating_score_test.go. If you add a
// case in either file, mirror it here so Go and TS stay in lockstep.
describe('normalizeRating', () => {
  const cases: Array<[string, string, number, string]> = [
    // АКРА — "(RU)" suffix
    ['АКРА AAA', 'AAA(RU)', 22, 'AAA'],
    ['АКРА AA+', 'AA+(RU)', 21, 'AA'],
    ['АКРА AA', 'AA(RU)', 20, 'AA'],
    ['АКРА AA-', 'AA-(RU)', 19, 'AA'],
    ['АКРА A+', 'A+(RU)', 18, 'A'],
    ['АКРА A', 'A(RU)', 17, 'A'],
    ['АКРА A-', 'A-(RU)', 16, 'A'],
    ['АКРА BBB+', 'BBB+(RU)', 15, 'BBB'],
    ['АКРА BBB', 'BBB(RU)', 14, 'BBB'],
    ['АКРА BBB-', 'BBB-(RU)', 13, 'BBB'],
    ['АКРА BB+', 'BB+(RU)', 12, 'BB'],
    ['АКРА BB', 'BB(RU)', 11, 'BB'],
    ['АКРА BB-', 'BB-(RU)', 10, 'BB'],
    ['АКРА B+', 'B+(RU)', 9, 'B'],
    ['АКРА B', 'B(RU)', 8, 'B'],
    ['АКРА D', 'D(RU)', 1, 'D'],

    // Эксперт РА — "ru" prefix
    ['Эксперт ruAAA', 'ruAAA', 22, 'AAA'],
    ['Эксперт ruAA+', 'ruAA+', 21, 'AA'],
    ['Эксперт ruAA-', 'ruAA-', 19, 'AA'],
    ['Эксперт ruA+', 'ruA+', 18, 'A'],
    ['Эксперт ruBBB-', 'ruBBB-', 13, 'BBB'],
    ['Эксперт ruBB+', 'ruBB+', 12, 'BB'],
    ['Эксперт ruB', 'ruB', 8, 'B'],
    ['Эксперт ruC', 'ruC', 2, 'C'],
    ['Эксперт ruD', 'ruD', 1, 'D'],

    // НКР — ".ru" suffix
    ['НКР AAA.ru', 'AAA.ru', 22, 'AAA'],
    ['НКР BBB+.ru', 'BBB+.ru', 15, 'BBB'],
    ['НКР BB-.ru', 'BB-.ru', 10, 'BB'],
    ['НКР B.ru', 'B.ru', 8, 'B'],

    // НРА — "|ru|" suffix
    ['НРА AAA|ru|', 'AAA|ru|', 22, 'AAA'],
    ['НРА AA+|ru|', 'AA+|ru|', 21, 'AA'],
    ['НРА BBB|ru|', 'BBB|ru|', 14, 'BBB'],
    ['НРА BB|ru|', 'BB|ru|', 11, 'BB'],

    // S&P / Fitch — bare letters
    ['S&P AAA', 'AAA', 22, 'AAA'],
    ['Fitch AA-', 'AA-', 19, 'AA'],
    ['Fitch BBB+', 'BBB+', 15, 'BBB'],
    ['S&P BB', 'BB', 11, 'BB'],
    ['S&P D', 'D', 1, 'D'],
    ['Fitch RD', 'RD', 1, 'D'],

    // Moody's — case-sensitive notation
    ["Moody's Aaa", 'Aaa', 22, 'AAA'],
    ["Moody's Aa1", 'Aa1', 21, 'AA'],
    ["Moody's Aa2", 'Aa2', 20, 'AA'],
    ["Moody's Aa3", 'Aa3', 19, 'AA'],
    ["Moody's A1", 'A1', 18, 'A'],
    ["Moody's A2", 'A2', 17, 'A'],
    ["Moody's A3", 'A3', 16, 'A'],
    ["Moody's Baa1", 'Baa1', 15, 'BBB'],
    ["Moody's Baa2", 'Baa2', 14, 'BBB'],
    ["Moody's Baa3", 'Baa3', 13, 'BBB'],
    ["Moody's Ba1", 'Ba1', 12, 'BB'],
    ["Moody's Ba3", 'Ba3', 10, 'BB'],
    ["Moody's B2", 'B2', 8, 'B'],
    ["Moody's Caa1", 'Caa1', 6, 'CCC'],
    ["Moody's Ca", 'Ca', 3, 'CC'],
    ["Moody's C", 'C', 2, 'C'],

    // ДОХОДЪ — numeric 1-10 and "X/10" forms
    ['ДОХОДЪ 10', '10', 22, 'AAA'],
    ['ДОХОДЪ 10/10', '10/10', 22, 'AAA'],
    ['ДОХОДЪ 8', '8', 20, 'AA'],
    ['ДОХОДЪ 7/10', '7/10', 19, 'AA'],
    ['ДОХОДЪ 5', '5', 17, 'A'],
    ['ДОХОДЪ 4', '4', 15, 'BBB'],
    ['ДОХОДЪ 1', '1', 9, 'B'],
    ['ДОХОДЪ 0 invalid', '0', 0, ''],
    ['ДОХОДЪ 11 invalid', '11', 0, ''],

    // Outlook stripping
    ['АКРА with outlook', 'AAA(RU) Стабильный', 22, 'AAA'],
    ['АКРА AA Negative parens', 'AA(RU) (Negative)', 20, 'AA'],
    ['Fitch BBB+ Stable', 'BBB+ Stable', 15, 'BBB'],
    ['S&P AA- Positive', 'AA- Positive', 19, 'AA'],
    ['Эксперт ruA+ развивающийся', 'ruA+ развивающийся', 18, 'A'],

    // Whitespace / case noise
    ['Whitespace', '  AAA  ', 22, 'AAA'],
    ['Lower-case acronyms', 'aaa(ru)', 22, 'AAA'],
    ['Mixed case NKR', 'BBB+.RU', 15, 'BBB'],

    // Withdrawn / unrated
    ['Withdrawn ru', 'Отозван', 0, ''],
    ['Withdrawn lat', 'WD', 0, ''],
    ['Empty', '', 0, ''],
    ['Dash', '—', 0, ''],

    // Garbage in
    ['Bogus', 'XYZ', 0, ''],
    ['Number out of range', '42', 0, ''],
  ]

  for (const [name, input, wantOrd, wantTier] of cases) {
    it(name, () => {
      const { ord, tier } = normalizeRating(input)
      expect(ord, `${name}: ord`).toBe(wantOrd)
      expect(tier, `${name}: tier`).toBe(wantTier)
    })
  }
})

describe('normalizeRating — investment-vs-speculative boundary', () => {
  // The whole point of the 1-22 scale: BBB- (lowest investment grade) must
  // rank strictly above BB+ (top speculative). Legacy 1-10 collapsed both to 3.
  it('BBB- ranks above BB+', () => {
    const bbbMinus = normalizeRating('BBB-(RU)').ord
    const bbPlus = normalizeRating('BB+(RU)').ord
    expect(bbbMinus).toBeGreaterThan(bbPlus)
  })
})

describe('normalizeRating — cross-agency ordering', () => {
  const pairs: Array<[string, string]> = [
    ['AAA(RU)', 'ruA'],
    ['AA+.ru', 'BBB-(RU)'],
    ['Aaa', 'ruBB+'],
    ['BBB+|ru|', 'ruBB-'],
    ['AA-', 'B+(RU)'],
    ['Baa1', 'Caa1'],
  ]
  for (const [stronger, weaker] of pairs) {
    it(`${stronger} > ${weaker}`, () => {
      expect(normalizeRating(stronger).ord).toBeGreaterThan(normalizeRating(weaker).ord)
    })
  }
})

describe('legacyScore10', () => {
  const cases: Array<[number, number]> = [
    [22, 10], [21, 9], [20, 8], [19, 7], [18, 6],
    [17, 5], [16, 5], [15, 4], [14, 4], [13, 3],
    [12, 3], [11, 2], [10, 2], [9, 1], [8, 1],
    [7, 1], [6, 0], [0, 0],
  ]
  for (const [ord, want] of cases) {
    it(`ord ${ord} → legacy ${want}`, () => {
      expect(legacyScore10(ord)).toBe(want)
    })
  }
})

describe('ratingTierStyle', () => {
  it('Moody Baa1 is yellow (BBB tier), not red', () => {
    // The bug we fixed: legacy ratingChipStyle hit `r.startsWith('b')` first
    // and coloured Baa1 red. Now it normalises to BBB tier → yellow.
    expect(ratingTierStyle('Baa1').color).toBe('#997404')
  })

  it('Moody Ba1 is orange (BB tier)', () => {
    expect(ratingTierStyle('Ba1').color).toBe('#e8590c')
  })

  it('АКРА AAA is green', () => {
    expect(ratingTierStyle('AAA(RU)').color).toBe('#198754')
  })

  it('withdrawn renders muted', () => {
    expect(ratingTierStyle('WD').color).toBe('var(--nla-text-muted)')
  })

  it('empty renders default', () => {
    expect(ratingTierStyle('').color).toBe('var(--nla-text)')
  })
})

describe('maxRatingOrd', () => {
  it('returns 0 for empty input', () => {
    expect(maxRatingOrd([])).toBe(0)
    expect(maxRatingOrd(null)).toBe(0)
    expect(maxRatingOrd(undefined)).toBe(0)
  })

  it('returns highest ord across agencies', () => {
    const ratings = [
      { rating: 'BBB-(RU)', score_ord: 13 },
      { rating: 'BB+(RU)', score_ord: 12 },
      { rating: 'A(RU)',   score_ord: 17 },
    ]
    expect(maxRatingOrd(ratings)).toBe(17)
  })

  it('falls back to normalizeRating when score_ord is missing', () => {
    const ratings = [{ rating: 'AA(RU)' }, { rating: 'BBB-(RU)' }]
    expect(maxRatingOrd(ratings)).toBe(20)
  })

  it('skips NULL ratings', () => {
    const ratings = [
      { rating: 'NULL', score_ord: 0 },
      { rating: 'BBB(RU)', score_ord: 14 },
    ]
    expect(maxRatingOrd(ratings)).toBe(14)
  })
})

describe('ordMatchesBucket', () => {
  // The critical regression: BBB- (ord 13) and BB+ (ord 12) now fall in
  // distinct buckets. Legacy filters collapsed them into one.
  it('BBB- in bbb bucket, BB+ in bb bucket', () => {
    expect(ordMatchesBucket(13, 'bbb')).toBe(true)
    expect(ordMatchesBucket(13, 'bb')).toBe(false)
    expect(ordMatchesBucket(12, 'bb')).toBe(true)
    expect(ordMatchesBucket(12, 'bbb')).toBe(false)
  })

  it('AAA only matches aaa bucket', () => {
    expect(ordMatchesBucket(22, 'aaa')).toBe(true)
    expect(ordMatchesBucket(21, 'aaa')).toBe(false)
    expect(ordMatchesBucket(21, 'aa')).toBe(true)
  })

  it('B/CCC/CC/C/D all fall into b_below', () => {
    for (const ord of [9, 8, 7, 6, 5, 4, 3, 2, 1]) {
      expect(ordMatchesBucket(ord, 'b_below'), `ord ${ord}`).toBe(true)
    }
  })

  it('none bucket matches unrated', () => {
    expect(ordMatchesBucket(0, 'none')).toBe(true)
    expect(ordMatchesBucket(14, 'none')).toBe(false)
  })

  it('empty bucket matches everything', () => {
    expect(ordMatchesBucket(0, '')).toBe(true)
    expect(ordMatchesBucket(22, '')).toBe(true)
  })
})
