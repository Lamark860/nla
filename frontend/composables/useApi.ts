// Typed API client for NLA Go backend

// --- Types ---

export interface Bond {
  secid: string
  shortname: string
  secname: string
  isin: string
  facevalue: number
  matdate: string
  coupon_period: number
  coupon_value: number
  coupon_percent: number
  next_coupon: string
  accrued_int: number
  bond_type: string
  // Extended securities fields
  boardid: string
  boardname: string
  latname: string
  regnumber: string
  currencyid: string
  faceunit: string
  issuesize: number
  issuesize_placed: number
  lotsize: number
  lotvalue: number
  minstep: number
  decimals: number
  listlevel: number
  sectype: string
  bondtype_name: string
  bondsubtype: string
  sectorid: string
  marketcode: string
  instrid: string
  settledate: string
  offerdate: string
  buybackdate: string
  buybackprice: number | null
  calloptiondate: string
  putoptiondate: string
  prevwaprice: number | null
  yieldatprevwaprice: number | null
  prevlegalcloseprice: number | null
  prevdate: string
  facevalue_on_settle: number
  // Marketdata
  last: number | null
  bid: number | null
  offer: number | null
  yield: number | null
  duration: number | null
  vol_today: number
  // Extended marketdata
  open: number | null
  low: number | null
  high: number | null
  waprice: number | null
  numtrades: number | null
  valtoday: number | null
  biddeptht: number | null
  offerdeptht: number | null
  numbids: number | null
  numoffers: number | null
  updatetime: string
  tradetime: string
  systime: string
  prevprice: number | null
  // Extended marketdata yield/spread
  spread: number | null
  yieldatwaprice: number | null
  yieldtoprevyield: number | null
  closeyield: number | null
  effectiveyield: number | null
  lastcngtolastwaprice: number | null
  waptoprevwapriceprcnt: number | null
  waptoprevwaprice: number | null
  lasttoprevprice: number | null
  priceminusprevwaprice: number | null
  marketprice: number | null
  marketpricetoday: number | null
  lcurrentprice: number | null
  lcloseprice: number | null
  change: number | null
  yieldtooffer: number | null
  yieldlastcoupon: number | null
  zspread: number | null
  zspreadatwaprice: number | null
  iricpiclose: number | null
  beiclose: number | null
  cbrclose: number | null
  calloptionyield: number | null
  calloptionduration: number | null
  // Price changes
  last_change: number | null
  last_change_prcnt: number | null
  trading_status: string
  // Calculated
  price_rub: number | null
  value_today_rub: number | null
  days_to_maturity: number | null
  is_float: boolean
  is_indexed: boolean
  bond_category: string
  coupon_display: number | null
}

export interface BondListResponse {
  data: Bond[]
  meta: { page: number; per_page: number; total: number }
}

export interface Coupon {
  coupon_date: string
  record_date: string
  start_date: string
  value: number
  value_percent: number
  value_rub: number
}

export interface OHLC {
  date: string
  open: number
  close: number
  high: number
  low: number
  volume: number
  value: number
}

export interface BondAnalysis {
  id: string
  secid: string
  response: string
  rating: number | null
  timestamp: string
}

export interface AnalysisStats {
  total: number
  avg_rating: number
  last_analysis: string | null
}

export interface JobStatus {
  job_id: string
  type: string
  status: 'pending' | 'running' | 'done' | 'error'
  result: any
  error: string
  created_at: string
  finished_at: string | null
}

export interface IssuerRating {
  issuer: string
  agency: string
  rating: string
  score: number
  updated_at: string
}

export interface IssuerRatingResponse {
  issuer: string
  ratings: IssuerRating[]
  score: number
}

export interface ChatSession {
  session_id: string
  title: string
  agent_type: string
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  session_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
}

export interface ChatAgent {
  type: string
  name: string
  description: string
  icon: string
}

export interface SendMessageResponse {
  user_message: ChatMessage
  assistant_message: ChatMessage
}

export interface ChatSessionDetail {
  session: ChatSession
  messages: ChatMessage[]
}

export interface IssuerGroup {
  emitter_id: number
  emitter_name: string
  bond_count: number
  bonds: Bond[]
}

export interface IssuerGroupResponse {
  groups: IssuerGroup[]
  total_issuers: number
  total_bonds: number
}

// --- API functions ---

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = import.meta.server
    ? (config.apiBaseServer as string)
    : config.public.apiBase as string

  function apiFetch<T>(path: string, opts?: any): Promise<T> {
    return $fetch<T>(path, { baseURL, ...opts })
  }

  return {
    // Bonds
    getBonds(page = 1, perPage = 20, sort = 'best') {
      return apiFetch<BondListResponse>(`/bonds?page=${page}&per_page=${perPage}&sort=${sort}`)
    },

    getBond(secid: string) {
      return apiFetch<Bond>(`/bonds/${secid}`)
    },

    getBondCoupons(secid: string) {
      return apiFetch<Coupon[]>(`/bonds/${secid}/coupons`)
    },

    getBondHistory(secid: string) {
      return apiFetch<OHLC[]>(`/bonds/${secid}/history`)
    },

    getMonthlyBonds() {
      return apiFetch<any>('/bonds/monthly')
    },

    getBondsGrouped(monthly = false) {
      const q = monthly ? '?monthly=true' : ''
      return apiFetch<IssuerGroupResponse>(`/bonds/grouped${q}`)
    },

    // AI Analysis
    startAnalysis(secid: string, data?: any) {
      return apiFetch<{ job_id: string; status: string }>(`/bonds/${secid}/analyze`, {
        method: 'POST',
        body: data ?? undefined,
      })
    },

    getAnalyses(secid: string) {
      return apiFetch<BondAnalysis[]>(`/bonds/${secid}/analyses`)
    },

    getAnalysis(id: string) {
      return apiFetch<BondAnalysis>(`/analyses/${id}`)
    },

    getAnalysisStats(secid: string) {
      return apiFetch<AnalysisStats>(`/bonds/${secid}/analysis-stats`)
    },

    getJobStatus(jobId: string) {
      return apiFetch<JobStatus>(`/jobs/${jobId}`)
    },

    getQueueStats() {
      return apiFetch<Record<string, number>>('/queue/stats')
    },

    // Ratings
    getRatings() {
      return apiFetch<Record<string, IssuerRatingResponse>>('/ratings')
    },

    getIssuerRating(issuer: string) {
      return apiFetch<IssuerRatingResponse>(`/ratings/search?issuer=${encodeURIComponent(issuer)}`)
    },

    // Chat
    getChatAgents() {
      return apiFetch<ChatAgent[]>('/chat/agents')
    },

    getChatSessions() {
      return apiFetch<ChatSession[]>('/chat/sessions')
    },

    createChatSession(agentType: string, title?: string) {
      return apiFetch<ChatSession>('/chat/sessions', {
        method: 'POST',
        body: { agent_type: agentType, title: title || '' },
      })
    },

    getChatSession(id: string) {
      return apiFetch<ChatSessionDetail>(`/chat/sessions/${id}`)
    },

    deleteChatSession(id: string) {
      return apiFetch<any>(`/chat/sessions/${id}`, { method: 'DELETE' })
    },

    sendChatMessage(sessionId: string, content: string) {
      return apiFetch<SendMessageResponse>(`/chat/sessions/${sessionId}/messages`, {
        method: 'POST',
        body: { content },
      })
    },

    // Favorites
    getFavorites(headers: Record<string, string>) {
      return apiFetch<{ secids: string[]; count: number }>('/favorites', { headers })
    },

    toggleFavorite(secid: string, headers: Record<string, string>) {
      return apiFetch<{ secid: string; is_favorite: boolean }>('/favorites/toggle', {
        method: 'POST',
        headers,
        body: { secid },
      })
    },
  }
}
