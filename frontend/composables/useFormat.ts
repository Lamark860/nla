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

    /** Rating color class based on score */
    ratingColor(rating: number | null | undefined): string {
      if (rating == null) return 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400'
      if (rating >= 75) return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
      if (rating >= 55) return 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-400'
      if (rating >= 35) return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-400'
      return 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400'
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
  }
}
