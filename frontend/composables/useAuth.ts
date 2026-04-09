// Auth composable — JWT token management + user state

interface User {
  id: number
  email: string
  name: string
}

const tokenKey = 'nla_token'

export function useAuth() {
  const user = useState<User | null>('auth_user', () => null)
  const token = useState<string | null>('auth_token', () => null)

  // Init from localStorage (client only)
  function init() {
    if (import.meta.server) return
    const saved = localStorage.getItem(tokenKey)
    if (saved) {
      token.value = saved
      fetchMe()
    }
  }

  function authHeaders(): Record<string, string> {
    if (!token.value) return {}
    return { Authorization: `Bearer ${token.value}` }
  }

  const config = useRuntimeConfig()
  const baseURL = import.meta.server
    ? (config.apiBaseServer as string)
    : config.public.apiBase as string

  async function fetchMe() {
    if (!token.value) return
    try {
      const data = await $fetch<User>('/auth/me', {
        baseURL,
        headers: authHeaders(),
      })
      user.value = data
    } catch {
      logout()
    }
  }

  async function login(email: string, password: string): Promise<{ ok: boolean; error?: string }> {
    try {
      const res = await $fetch<{ token: string; user: User }>('/auth/login', {
        baseURL,
        method: 'POST',
        body: { email, password },
      })
      token.value = res.token
      user.value = res.user
      if (!import.meta.server) localStorage.setItem(tokenKey, res.token)
      return { ok: true }
    } catch (e: any) {
      const msg = e?.data?.error || e?.message || 'Ошибка входа'
      return { ok: false, error: msg }
    }
  }

  async function register(email: string, password: string, name: string): Promise<{ ok: boolean; error?: string }> {
    try {
      const res = await $fetch<{ token: string; user: User }>('/auth/register', {
        baseURL,
        method: 'POST',
        body: { email, password, name },
      })
      token.value = res.token
      user.value = res.user
      if (!import.meta.server) localStorage.setItem(tokenKey, res.token)
      return { ok: true }
    } catch (e: any) {
      const msg = e?.data?.error || e?.message || 'Ошибка регистрации'
      return { ok: false, error: msg }
    }
  }

  function logout() {
    token.value = null
    user.value = null
    if (!import.meta.server) localStorage.removeItem(tokenKey)
  }

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  return {
    user: readonly(user),
    token: readonly(token),
    isLoggedIn,
    init,
    login,
    register,
    logout,
    authHeaders,
  }
}
