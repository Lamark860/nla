// Global selected scoring profile (Phase 2 / Phase 3). Persists to
// localStorage so the choice survives reloads. Lives as a useState ref
// so any component composing this composable gets the same reactive value.
//
// Profile codes mirror the Go side: scoring.ProfileLow/Mid/High and the
// scoring_profiles.code column. Keep names in sync with the migration
// seed; the engine only knows these three at runtime.

export type ProfileCode = 'low' | 'mid' | 'high'

export interface ProfileMeta {
  code: ProfileCode
  shortLabel: string  // sidebar chip
  label: string       // full RU label
  icon: string        // emoji used in the sidebar switcher
  ariaLabel: string
}

export const PROFILE_META: Record<ProfileCode, ProfileMeta> = {
  low:  { code: 'low',  shortLabel: 'Низкий',     label: 'Низкий риск',     icon: '🛡️', ariaLabel: 'Профиль «Низкий риск»' },
  mid:  { code: 'mid',  shortLabel: 'Средний',    label: 'Средний риск',    icon: '⚖️', ariaLabel: 'Профиль «Средний риск»' },
  high: { code: 'high', shortLabel: 'Повышенный', label: 'Повышенный риск', icon: '🚀', ariaLabel: 'Профиль «Повышенный риск»' },
}

export const PROFILE_ORDER: ProfileCode[] = ['low', 'mid', 'high']

const storageKey = 'nla_scoring_profile'
const defaultProfile: ProfileCode = 'mid'

function isProfileCode(v: unknown): v is ProfileCode {
  return v === 'low' || v === 'mid' || v === 'high'
}

export function useScoringProfile() {
  // useState is Nuxt's SSR/SPA-safe shared ref. With ssr: false the SSR
  // branch never runs, but using useState (instead of plain ref) keeps
  // the value shared across components without a Pinia store.
  const profile = useState<ProfileCode>('scoring_profile', () => defaultProfile)

  // Hydrate from localStorage on client. Safe to call multiple times —
  // the first call wins, subsequent calls noop because the value is
  // already non-default after init.
  function init() {
    if (import.meta.server) return
    const saved = localStorage.getItem(storageKey)
    if (isProfileCode(saved)) profile.value = saved
  }

  function set(code: ProfileCode) {
    profile.value = code
    if (!import.meta.server) localStorage.setItem(storageKey, code)
  }

  const meta = computed(() => PROFILE_META[profile.value])

  return {
    profile,            // readonly-ish ref, but kept as ref so v-model="profile.value = …" works
    set,
    init,
    meta,
    order: PROFILE_ORDER,
    metaMap: PROFILE_META,
  }
}
