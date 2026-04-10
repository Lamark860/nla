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
  durationwaprice: number | null
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
  // Extended calculated fields
  years_to_maturity: number | null
  days_to_call: number | null
  days_to_put: number | null
  days_to_next_coupon: number | null
  current_yield: number | null
  modified_duration: number | null
  spread_absolute: number | null
  spread_percent: number | null
  mid_price_pct: number | null
  mid_price_rub: number | null
  bid_rub: number | null
  offer_rub: number | null
  avg_trade_size: number | null
  total_depth: number | null
  bid_offer_ratio: number | null
  accrued_int_pct: number | null
  life_progress: number | null
  is_near_offer: boolean
  has_no_fixed_coupon: boolean
  trading_status_text: string
  risk_category: string
  // Issuer info (resolved from bond_issuers)
  emitter_id?: number
  emitter_name?: string
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
  emitter_id: number
  issuer: string
  agency: string
  rating: string
  score: number
  updated_at: string
}

export interface IssuerRatingResponse {
  emitter_id: number
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

export interface DohodBondData {
  isin: string
  secid: string
  issuer_name: string
  issuer_sector: string
  borrower_name: string
  country: string
  credit_rating: number
  credit_rating_text: string
  akra: string
  expert_ra: string
  fitch: string
  moody: string
  sp: string
  estimation_rating: number
  estimation_rating_text: string
  quality: number | null
  quality_outside: number | null
  quality_inside: number | null
  quality_balance: number | null
  quality_earnings: number | null
  quality_roe_score: number | null
  quality_roe_value: number | null
  quality_net_debt: number | null
  quality_net_debt_value: number | null
  quality_profit_change: number | null
  dp1: number | null
  dp2: number | null
  dp3: number | null
  profit_ros: number | null
  profit_ros_value: number | null
  profit_oper: number | null
  profit_oper_value: number | null
  turnover_inventory: number | null
  turnover_cur_asset: number | null
  turnover_receivable: number | null
  liq_balance: number | null
  liq_current: number | null
  liq_quick: number | null
  liq_cash_ratio: number | null
  liq_current_value: number | null
  liq_quick_value: number | null
  liq_cash_value: number | null
  stability: number | null
  stability_short_debt: number | null
  stability_debt_value: number | null
  best_score: number | null
  down_risk: number | null
  liquidity_score: number | null
  total_return: number | null
  current_yield: number | null
  size: number | null
  complexity: number | null
  // Bond description
  description: string
  event: string
  coupon_rate: number | null
  coupon_rate_after_put: number | null
  coupon_size: number | null
  early_redemption_call: string
  years_to_maturity: number | null
  dohod_duration: number | null
  dohod_duration_md: number | null
  simple_yield: number | null
  for_qualified_only: boolean
  tax_longterm_free: boolean
  tax_free: boolean
  tax_currency_free: boolean
  sector_text: string
  min_lot: number | null
  frn_index: string
  frn_index_add: number | null
  frn_formula_text: string
  fetched_at: string
  updated_at: string
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

    deleteAnalysis(id: string) {
      return apiFetch<{ status: string }>(`/analyses/${id}`, { method: 'DELETE' })
    },

    getAnalysisStats(secid: string) {
      return apiFetch<AnalysisStats>(`/bonds/${secid}/analysis-stats`)
    },

    getBulkAnalysisStats() {
      return apiFetch<Record<string, AnalysisStats>>('/analyses/bulk-stats')
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

    // Admin
    clearCache() {
      return apiFetch<{ deleted: number }>('/bonds/clear-cache', { method: 'POST' })
    },

    toggleIssuer(id: number) {
      return apiFetch<{ emitter_id: number; hidden: boolean }>(`/issuers/${id}/toggle`, { method: 'POST' })
    },

    // Dohod.ru details
    getDohodDetails(secid: string) {
      return apiFetch<DohodBondData | { job_id: string; status: string }>(`/bonds/${secid}/dohod`)
    },
  }
}
