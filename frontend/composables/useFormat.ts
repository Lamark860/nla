// Formatting helpers for bond data display

export function useFormat() {
  return {
    /** Price in RUB — e.g. "982.50 ₽" */
    priceRub(val: number | null | undefined): string {
      if (val == null) return '—'
      return val.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' ₽'
    },

    /** Percent — e.g. "12.35%" */
    percent(val: number | null | undefined, decimals = 2): string {
      if (val == null) return '—'
      // Cap absurd values from MOEX data quality issues
      if (Math.abs(val) > 999) return val > 0 ? '>999%' : '<-999%'
      return val.toFixed(decimals) + '%'
    },

    /** Number with locale — e.g. "1 234 567" */
    num(val: number | null | undefined, decimals = 0): string {
      if (val == null) return '—'
      return val.toLocaleString('ru-RU', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
    },

    /** Date — e.g. "15 мар 2026" */
    date(val: string | null | undefined): string {
      if (!val) return '—'
      try {
        return new Date(val).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
      } catch {
        return val
      }
    },

    /** Short date — "15.03.26" */
    dateShort(val: string | null | undefined): string {
      if (!val) return '—'
      try {
        return new Date(val).toLocaleDateString('ru-RU')
      } catch {
        return val
      }
    },

    /** Days to maturity — "1 год 3 мес" / "45 дн" */
    daysToMaturity(days: number | null | undefined): string {
      if (days == null) return '—'
      if (days < 0) return 'Погашена'
      if (days <= 30) return `${days} дн`
      if (days < 365) return `${Math.round(days / 30)} мес`
      const years = Math.floor(days / 365)
      const months = Math.round((days % 365) / 30)
      if (months === 0) return `${years} г`
      return `${years} г ${months} мес`
    },

    /** Volume in human readable — "12.3 млн" */
    volume(val: number | null | undefined): string {
      if (val == null || val === 0) return '—'
      if (val >= 1_000_000_000) return (val / 1_000_000_000).toFixed(1) + ' млрд'
      if (val >= 1_000_000) return (val / 1_000_000).toFixed(1) + ' млн'
      if (val >= 1_000) return (val / 1_000).toFixed(0) + ' тыс'
      return val.toLocaleString('ru-RU')
    },

    /**
     * Rating color style for AI scores (0-100 scale).
     * Returns { background, color } for inline styles.
     */
    aiRatingStyle(rating: number | null | undefined): { background: string; color: string } {
      if (rating == null) return { background: '#6c757d', color: '#fff' }
      if (rating >= 80) return { background: '#198754', color: '#fff' }
      if (rating >= 60) return { background: '#0d6efd', color: '#fff' }
      if (rating >= 40) return { background: '#ffc107', color: '#000' }
      if (rating >= 20) return { background: '#fd7e14', color: '#fff' }
      return { background: '#dc3545', color: '#fff' }
    },

    /**
     * Semi-transparent AI rating style (colored text on tinted background).
     */
    aiRatingStyleSoft(rating: number | null | undefined): { background: string; color: string } {
      if (rating == null) return { background: 'rgba(108,117,125,0.15)', color: '#6c757d' }
      if (rating >= 80) return { background: 'rgba(25,135,84,0.15)', color: '#198754' }
      if (rating >= 60) return { background: 'rgba(13,110,253,0.15)', color: '#0d6efd' }
      if (rating >= 40) return { background: 'rgba(255,193,7,0.15)', color: '#997404' }
      if (rating >= 20) return { background: 'rgba(253,126,20,0.15)', color: '#fd7e14' }
      return { background: 'rgba(220,53,69,0.15)', color: '#dc3545' }
    },

    /**
     * Issuer/emitter rating color (1-10 dohod.ru scale).
     * Returns background color string.
     */
    issuerRatingBg(score: number): string {
      const c: Record<number, string> = {
        10: '#198754', 9: '#198754', 8: '#17a2b8', 7: '#5bc0de',
        6: '#0d6efd', 5: '#6ea8fe', 4: '#ffc107', 3: '#fd7e14',
        2: '#dc3545', 1: '#a70820', 0: '#000000',
      }
      return c[score] ?? '#6c757d'
    },

    /** Date+time — "09.04.2026 19:43:58" */
    dateTime(val: string | null | undefined): string {
      if (!val) return '—'
      try {
        const d = new Date(val)
        return d.toLocaleDateString('ru-RU') + ' ' + d.toLocaleTimeString('ru-RU')
      } catch {
        return val
      }
    },

    /** Time only — "19:43:58" */
    time(val: string | null | undefined): string {
      if (!val) return '—'
      // If already HH:MM:SS format, return as-is
      if (/^\d{2}:\d{2}(:\d{2})?$/.test(val)) return val
      try {
        return new Date(val).toLocaleTimeString('ru-RU')
      } catch {
        return val
      }
    },

    /**
     * Credit rating chip style by rating text (agency-specific formats).
     * Returns { background, color } for inline styles.
     */
    ratingChipStyle(rating: string): { background: string; color: string } {
      const r = (rating || '').toLowerCase()
      if (r === 'отозван' || r === 'отозвано' || r.startsWith('wd'))
        return { background: 'rgba(108,117,125,0.12)', color: 'var(--nla-text-muted)' }
      if (r.startsWith('aaa') || r.startsWith('ruaaa'))
        return { background: 'rgba(25,135,84,0.14)', color: '#198754' }
      if (r.startsWith('aa') || r.startsWith('ruaa'))
        return { background: 'rgba(13,110,253,0.14)', color: '#0d6efd' }
      if (r.startsWith('a') || r.startsWith('rua'))
        return { background: 'rgba(32,201,151,0.14)', color: '#0d9488' }
      if (r.startsWith('bbb') || r.startsWith('rubbb'))
        return { background: 'rgba(255,193,7,0.14)', color: '#997404' }
      if (r.startsWith('bb') || r.startsWith('rubb'))
        return { background: 'rgba(253,126,20,0.14)', color: '#e8590c' }
      if (r.startsWith('b') || r.startsWith('rub') || r.startsWith('c') || r.startsWith('d'))
        return { background: 'rgba(220,53,69,0.14)', color: '#dc3545' }
      return { background: 'rgba(108,117,125,0.12)', color: 'var(--nla-text)' }
    },
  }
}
