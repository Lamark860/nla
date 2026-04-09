// Favorites composable — manage watchlist

export function useFavorites() {
  const auth = useAuth()
  const config = useRuntimeConfig()
  const baseURL = import.meta.server
    ? (config.apiBaseServer as string)
    : config.public.apiBase as string

  const favoriteSecids = useState<Set<string>>('favorite_secids', () => new Set())
  const loading = useState<boolean>('favorites_loading', () => false)

  async function load() {
    if (!auth.isLoggedIn.value) {
      favoriteSecids.value = new Set()
      return
    }
    try {
      loading.value = true
      const res = await $fetch<{ secids: string[]; count: number }>('/favorites', {
        baseURL,
        headers: auth.authHeaders(),
      })
      favoriteSecids.value = new Set(res.secids)
    } catch {
      favoriteSecids.value = new Set()
    } finally {
      loading.value = false
    }
  }

  async function toggle(secid: string): Promise<boolean> {
    if (!auth.isLoggedIn.value) return false
    try {
      const res = await $fetch<{ secid: string; is_favorite: boolean }>('/favorites/toggle', {
        baseURL,
        method: 'POST',
        headers: auth.authHeaders(),
        body: { secid },
      })
      if (res.is_favorite) {
        favoriteSecids.value.add(secid)
      } else {
        favoriteSecids.value.delete(secid)
      }
      // Trigger reactivity
      favoriteSecids.value = new Set(favoriteSecids.value)
      return res.is_favorite
    } catch {
      return false
    }
  }

  function isFavorite(secid: string): boolean {
    return favoriteSecids.value.has(secid)
  }

  const count = computed(() => favoriteSecids.value.size)

  return {
    favoriteSecids: readonly(favoriteSecids),
    loading: readonly(loading),
    count,
    load,
    toggle,
    isFavorite,
  }
}
