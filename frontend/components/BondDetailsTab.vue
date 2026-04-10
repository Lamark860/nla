<template>
  <div v-if="bond">
    <div class="row g-4">
      <!-- Дополнительные цены -->
      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-tags me-2"></i>Дополнительные цены</div>
        <InfoRow label="Рыночная цена" :value="fmt.percent(bond.marketpricetoday)" />
        <InfoRow label="Рыночная цена (вчера)" :value="fmt.percent(bond.marketprice)" />
        <InfoRow label="Предыд. юр. цена закрытия" :value="fmt.percent(bond.prevlegalcloseprice)" />
        <InfoRow label="Цена выкупа" :value="fmt.percent(bond.buybackprice)" />
        <InfoRow label="Средневзвешенная цена" :value="fmt.percent(bond.waprice)" />
        <InfoRow label="Средняя bid/ask (mid)" :value="fmt.percent(bond.mid_price_pct)" />
        <InfoRow label="Mid в ₽" :value="fmt.priceRub(bond.mid_price_rub)" />
        <InfoRow label="Цена закрытия" :value="fmt.percent(bond.lcloseprice)" />
        <InfoRow label="Текущая цена" :value="fmt.percent(bond.lcurrentprice)" />
        <InfoRow label="Bid в ₽" :value="fmt.priceRub(bond.bid_rub)" />
        <InfoRow label="Offer в ₽" :value="fmt.priceRub(bond.offer_rub)" />
      </div>


      <!-- Расчётные показатели -->
      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-calculator me-2"></i>Расчётные показатели</div>
        <InfoRow label="Спред (абс.)" :value="fmt.num(bond.spread_absolute, 4)" />
        <InfoRow label="Спред (%)" :value="fmt.percent(bond.spread_percent)" />
        <InfoRow label="Текущая доходность" :value="fmt.percent(bond.current_yield)" />
        <InfoRow label="Модифицированная дюрация" :value="fmt.num(bond.modified_duration, 0)" />
        <InfoRow label="Дюрация по WAP" :value="bond.durationwaprice != null ? fmt.num(bond.durationwaprice) + ' дн.' : '—'" />
        <InfoRow label="Средний размер сделки" :value="fmt.priceRub(bond.avg_trade_size)" />
        <InfoRow label="Общая глубина стакана" :value="fmt.num(bond.total_depth)" />
        <InfoRow label="Соотношение bid/offer" :value="fmt.num(bond.bid_offer_ratio, 2)" />
        <InfoRow label="НКД в % номинала" :value="fmt.percent(bond.accrued_int_pct)" />
        <InfoRow label="Прогресс жизни" :value="bond.life_progress != null ? fmt.num(bond.life_progress, 1) + '%' : '—'" />
      </div>
    </div>

    <!-- Даты и оферты -->
    <div class="row g-4 mt-1">
      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-calendar-event me-2"></i>Даты и оферты</div>
        <InfoRow label="Дата погашения" :value="fmt.date(bond.matdate)" />
        <InfoRow label="Лет до погашения" :value="bond.years_to_maturity != null ? fmt.num(bond.years_to_maturity, 2) : '—'" />
        <InfoRow label="Дата размещения" :value="fmt.date(bond.settledate)" />
        <InfoRow label="PUT-оферта" :value="bond.putoptiondate ? `${fmt.date(bond.putoptiondate)} (${bond.days_to_put ?? '?'} дн.)` : '—'" />
        <InfoRow label="CALL-оферта" :value="bond.calloptiondate ? `${fmt.date(bond.calloptiondate)} (${bond.days_to_call ?? '?'} дн.)` : '—'" />
        <InfoRow label="CALL доходность" :value="fmt.percent(bond.calloptionyield)" />
        <InfoRow label="CALL дюрация" :value="bond.calloptionduration != null ? fmt.num(bond.calloptionduration) + ' дн.' : '—'" />
        <InfoRow label="Ближняя оферта (<90 дн.)" :value="bond.is_near_offer ? '⚠️ Да' : 'Нет'" />
        <InfoRow label="Дата выкупа" :value="fmt.date(bond.buybackdate)" />
        <InfoRow label="Дата оферты" :value="fmt.date(bond.offerdate)" />
      </div>

      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-shield-exclamation me-2"></i>Оценка рисков</div>
        <InfoRow label="Категория риска">
          <template #value>
            <span :class="riskBadgeClass(bond.risk_category)" class="badge">{{ riskLabel(bond.risk_category) }}</span>
          </template>
        </InfoRow>
        <InfoRow label="Тип купона">
          <template #value>
            <span v-if="bond.is_float" class="badge bg-info">Флоатер</span>
            <span v-else-if="bond.is_indexed" class="badge bg-secondary">Индексируемая</span>
            <span v-else class="badge bg-light text-dark border">Фиксированный</span>
          </template>
        </InfoRow>
        <InfoRow label="Статус торгов" :value="bond.trading_status_text" />
        <InfoRow label="Уровень листинга" :value="bond.listlevel ? String(bond.listlevel) : '—'" />
        <InfoRow label="Сектор" :value="bond.sectorid || '—'" />
        <InfoRow label="Категория" :value="bond.bond_category" />
        <InfoRow label="Подтип" :value="bond.bondsubtype || '—'" />
        <InfoRow label="Инструмент" :value="bond.instrid || '—'" />
      </div>
    </div>

    <!-- Объёмы торгов -->
    <div class="row g-4 mt-1">
      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-bar-chart me-2"></i>Объёмы торгов</div>
        <InfoRow label="Кол-во сделок" :value="fmt.num(bond.numtrades)" />
        <InfoRow label="Объём (шт.)" :value="fmt.num(bond.vol_today)" />
        <InfoRow label="Объём (₽)" :value="fmt.volume(bond.value_today_rub)" />
        <InfoRow label="Средний размер сделки" :value="fmt.priceRub(bond.avg_trade_size)" />
        <InfoRow label="Глубина bid" :value="fmt.num(bond.biddeptht)" />
        <InfoRow label="Глубина offer" :value="fmt.num(bond.offerdeptht)" />
        <InfoRow label="Заявок на покупку" :value="fmt.num(bond.numbids)" />
        <InfoRow label="Заявок на продажу" :value="fmt.num(bond.numoffers)" />
      </div>

      <div class="col-lg-6">
        <div class="panel-header mb-3"><i class="bi bi-info-circle me-2"></i>Параметры выпуска</div>
        <InfoRow label="SECID" :value="bond.secid" />
        <InfoRow label="ISIN" :value="bond.isin" />
        <InfoRow label="Рег. номер" :value="bond.regnumber || '—'" />
        <InfoRow label="Объём выпуска" :value="fmt.num(bond.issuesize)" />
        <InfoRow label="Размещено" :value="fmt.num(bond.issuesize_placed)" />
        <InfoRow label="Номинал" :value="`${fmt.num(bond.facevalue)} ${bond.faceunit || ''}`" />
        <InfoRow label="Лот" :value="String(bond.lotsize || '—')" />
        <InfoRow label="Шаг цены" :value="fmt.num(bond.minstep, 4)" />
        <InfoRow label="Валюта" :value="bond.currencyid || '—'" />
      </div>
    </div>

    <!-- Dohod.ru analytics -->
    <div class="mt-4">
      <div class="d-flex align-items-center gap-2 mb-3">
        <div class="panel-header mb-0"><i class="bi bi-graph-up-arrow me-2"></i>Аналитика Dohod.ru</div>
        <div v-if="dohodLoading" class="spinner-border spinner-border-sm text-primary" role="status">
          <span class="visually-hidden">Загрузка…</span>
        </div>
        <small v-if="!dohod && !dohodLoading" class="text-muted">Данные загружаются автоматически…</small>
      </div>

      <template v-if="dohod">
        <div class="row g-4">
          <!-- Кредитные рейтинги -->
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-shield-check me-2"></i>Кредитные рейтинги</div>
            <InfoRow label="Общий рейтинг">
              <template #value>
                <span class="badge font-monospace fw-bold" :style="{ backgroundColor: qualityColor(dohod.credit_rating, 10), color: '#fff' }">
                  {{ dohod.credit_rating_text || '—' }} ({{ dohod.credit_rating }}/10)
                </span>
              </template>
            </InfoRow>
            <InfoRow label="Оценка Dohod.ru">
              <template #value>
                <span v-if="dohod.estimation_rating_text" class="badge bg-secondary font-monospace">{{ dohod.estimation_rating_text }}</span>
                <span v-else>—</span>
              </template>
            </InfoRow>
            <InfoRow label="АКРА" :value="dohod.akra || '—'" />
            <InfoRow label="Эксперт РА" :value="dohod.expert_ra || '—'" />
            <InfoRow label="Fitch" :value="dohod.fitch || '—'" />
            <InfoRow label="Moody's" :value="dohod.moody || '—'" />
            <InfoRow label="S&P" :value="dohod.sp || '—'" />
          </div>

          <!-- Качество эмитента -->
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-speedometer2 me-2"></i>Качество эмитента</div>
            <InfoRow label="Общее качество">
              <template #value>
                <QualityBadge :value="dohod.quality" />
              </template>
            </InfoRow>
            <InfoRow label="Внешнее (рейтинг)">
              <template #value><QualityBadge :value="dohod.quality_outside" /></template>
            </InfoRow>
            <InfoRow label="Внутреннее (фин.)">
              <template #value><QualityBadge :value="dohod.quality_inside" /></template>
            </InfoRow>
            <InfoRow label="Баланс">
              <template #value><QualityBadge :value="dohod.quality_balance" /></template>
            </InfoRow>
            <InfoRow label="Прибыльность">
              <template #value><QualityBadge :value="dohod.quality_earnings" /></template>
            </InfoRow>
            <InfoRow label="Изменение прибыли" :value="dohod.quality_profit_change != null ? fmt.num(dohod.quality_profit_change, 1) : '—'" />
          </div>
        </div>

        <div class="row g-4 mt-1">
          <!-- Рентабельность и оборачиваемость -->
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-cash-coin me-2"></i>Рентабельность</div>
            <InfoRow label="ROS (рейтинг)">
              <template #value><QualityBadge :value="dohod.profit_ros" /></template>
            </InfoRow>
            <InfoRow label="ROS (значение)" :value="dohod.profit_ros_value != null ? fmt.num(dohod.profit_ros_value, 1) + '%' : '—'" />
            <InfoRow label="Опер. прибыль (рейтинг)">
              <template #value><QualityBadge :value="dohod.profit_oper" /></template>
            </InfoRow>
            <InfoRow label="Опер. прибыль (значение)" :value="dohod.profit_oper_value != null ? fmt.num(dohod.profit_oper_value, 1) + '%' : '—'" />
            <div class="panel-header mb-3 mt-3"><i class="bi bi-arrow-repeat me-2"></i>Оборачиваемость</div>
            <InfoRow label="Запасы">
              <template #value><QualityBadge :value="dohod.turnover_inventory" /></template>
            </InfoRow>
            <InfoRow label="Оборотные активы">
              <template #value><QualityBadge :value="dohod.turnover_cur_asset" /></template>
            </InfoRow>
            <InfoRow label="Дебиторская задолж.">
              <template #value><QualityBadge :value="dohod.turnover_receivable" /></template>
            </InfoRow>
          </div>

          <!-- Ликвидность и стабильность -->
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-droplet me-2"></i>Ликвидность</div>
            <InfoRow label="Балансовая">
              <template #value><QualityBadge :value="dohod.liq_balance" /></template>
            </InfoRow>
            <InfoRow label="Текущая">
              <template #value>
                <QualityBadge :value="dohod.liq_current" />
                <small v-if="dohod.liq_current_value != null" class="text-muted ms-1">({{ fmt.num(dohod.liq_current_value, 2) }})</small>
              </template>
            </InfoRow>
            <InfoRow label="Быстрая">
              <template #value>
                <QualityBadge :value="dohod.liq_quick" />
                <small v-if="dohod.liq_quick_value != null" class="text-muted ms-1">({{ fmt.num(dohod.liq_quick_value, 2) }})</small>
              </template>
            </InfoRow>
            <InfoRow label="Денежная">
              <template #value>
                <QualityBadge :value="dohod.liq_cash_ratio" />
                <small v-if="dohod.liq_cash_value != null" class="text-muted ms-1">({{ fmt.num(dohod.liq_cash_value, 2) }})</small>
              </template>
            </InfoRow>
            <div class="panel-header mb-3 mt-3"><i class="bi bi-building me-2"></i>Стабильность</div>
            <InfoRow label="Стабильность">
              <template #value><QualityBadge :value="dohod.stability" /></template>
            </InfoRow>
            <InfoRow label="Краткоср. долг">
              <template #value>
                <QualityBadge :value="dohod.stability_short_debt" />
                <small v-if="dohod.stability_debt_value != null" class="text-muted ms-1">({{ fmt.num(dohod.stability_debt_value, 2) }})</small>
              </template>
            </InfoRow>
          </div>
        </div>

        <!-- Ключевые метрики -->
        <div class="row g-4 mt-1">
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-trophy me-2"></i>Ключевые метрики</div>
            <InfoRow label="Лучший результат">
              <template #value><QualityBadge :value="dohod.best_score" /></template>
            </InfoRow>
            <InfoRow label="Риск просадки" :value="dohod.down_risk != null ? fmt.num(dohod.down_risk, 1) + '%' : '—'" />
            <InfoRow label="Ликвидность" :value="dohod.liquidity_score != null ? fmt.num(dohod.liquidity_score, 1) : '—'" />
            <InfoRow label="Полная доходность" :value="dohod.total_return != null ? fmt.num(dohod.total_return, 2) + '%' : '—'" />
            <InfoRow label="Текущая доходность" :value="dohod.current_yield != null ? fmt.num(dohod.current_yield, 2) + '%' : '—'" />
            <InfoRow label="Размер" :value="dohod.size != null ? fmt.num(dohod.size, 1) : '—'" />
            <InfoRow label="Сложность" :value="dohod.complexity != null ? fmt.num(dohod.complexity, 1) : '—'" />
          </div>

          <!-- Штрафы/бонусы -->
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-plus-slash-minus me-2"></i>Корректировки (DP)</div>
            <InfoRow label="DP1 (рентабельность)" :value="dohod.dp1 != null ? (dohod.dp1 >= 0 ? '+' : '') + fmt.num(dohod.dp1, 1) : '—'" />
            <InfoRow label="DP2 (прибыль)" :value="dohod.dp2 != null ? (dohod.dp2 >= 0 ? '+' : '') + fmt.num(dohod.dp2, 1) : '—'" />
            <InfoRow label="DP3 (баланс)" :value="dohod.dp3 != null ? (dohod.dp3 >= 0 ? '+' : '') + fmt.num(dohod.dp3, 1) : '—'" />
          </div>
        </div>

        <!-- Данные облигации (dohod.ru) -->
        <div class="row g-4 mt-1">
          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-journal-text me-2"></i>Параметры облигации (Dohod.ru)</div>
            <InfoRow v-if="dohod.description" label="Описание" :value="dohod.description" />
            <InfoRow label="Ближайшее событие">
              <template #value>
                <span v-if="dohod.event" class="badge" :class="eventBadgeClass(dohod.event)">{{ dohod.event }}</span>
                <span v-else>—</span>
              </template>
            </InfoRow>
            <InfoRow label="Купон (dohod)" :value="dohod.coupon_rate != null ? fmt.num(dohod.coupon_rate, 2) + '%' : '—'" />
            <InfoRow v-if="dohod.coupon_rate_after_put != null" label="Ставка после оферты" :value="fmt.num(dohod.coupon_rate_after_put, 2) + '%'" />
            <InfoRow v-if="dohod.coupon_size != null" label="Размер купона" :value="fmt.num(dohod.coupon_size, 2) + ' ₽'" />
            <InfoRow v-if="dohod.early_redemption_call" label="Досрочное погашение (CALL)" :value="dohod.early_redemption_call" />
            <InfoRow v-if="dohod.simple_yield != null" label="Простая доходность" :value="fmt.num(dohod.simple_yield, 2) + '%'" />
            <InfoRow v-if="dohod.dohod_duration != null" label="Дюрация (dohod)" :value="fmt.num(dohod.dohod_duration, 2) + ' лет'" />
            <InfoRow v-if="dohod.years_to_maturity != null" label="Лет до погашения (dohod)" :value="fmt.num(dohod.years_to_maturity, 2)" />
          </div>

          <div class="col-lg-6">
            <div class="panel-header mb-3"><i class="bi bi-flag me-2"></i>Особенности</div>
            <InfoRow v-if="dohod.frn_formula_text" label="Формула флоатера" :value="dohod.frn_formula_text" />
            <InfoRow v-if="dohod.frn_index" label="Базовая ставка" :value="dohod.frn_index + (dohod.frn_index_add != null ? ' + ' + fmt.num(dohod.frn_index_add, 2) + '%' : '')" />
            <InfoRow v-if="dohod.min_lot != null" label="Минимальный лот (dohod)" :value="fmt.num(dohod.min_lot, 0) + ' шт.'" />
            <InfoRow v-if="dohod.sector_text" label="Сектор (dohod)" :value="dohod.sector_text" />
            <InfoRow label="Для квал. инвесторов">
              <template #value>
                <span v-if="dohod.for_qualified_only" class="badge bg-warning text-dark">Да</span>
                <span v-else class="badge bg-light text-dark border">Нет</span>
              </template>
            </InfoRow>
            <InfoRow label="Налоговые льготы">
              <template #value>
                <div class="d-flex flex-wrap gap-1">
                  <span v-if="dohod.tax_longterm_free" class="badge bg-success">ИИС/ДВ</span>
                  <span v-if="dohod.tax_free" class="badge bg-success">Без НДФЛ</span>
                  <span v-if="dohod.tax_currency_free" class="badge bg-success">Валютн.</span>
                  <span v-if="!dohod.tax_longterm_free && !dohod.tax_free && !dohod.tax_currency_free" class="text-muted">Нет</span>
                </div>
              </template>
            </InfoRow>
            <div class="panel-header mb-3 mt-3"><i class="bi bi-person-badge me-2"></i>Эмитент</div>
            <InfoRow label="Эмитент" :value="dohod.issuer_name || '—'" />
            <InfoRow v-if="dohod.borrower_name && dohod.borrower_name !== dohod.issuer_name" label="Заёмщик" :value="dohod.borrower_name" />
            <InfoRow label="Страна" :value="dohod.country || '—'" />
            <div class="text-muted small mt-2">
              <i class="bi bi-clock me-1"></i>Данные от {{ formatDohodDate(dohod.fetched_at) }}
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond, DohodBondData } from '~/composables/useApi'

const props = defineProps<{
  bond: Bond | null
  dohod?: DohodBondData | null
  dohodLoading?: boolean
}>()
const fmt = useFormat()

function qualityColor(val: number | null, max = 10): string {
  if (val == null) return '#6c757d'
  const pct = val / max
  if (pct >= 0.8) return '#198754'
  if (pct >= 0.6) return '#0dcaf0'
  if (pct >= 0.4) return '#0d6efd'
  if (pct >= 0.2) return '#fd7e14'
  return '#dc3545'
}

function formatDohodDate(dateStr: string): string {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function riskBadgeClass(risk: string): string {
  switch (risk) {
    case 'low': return 'bg-success'
    case 'medium': return 'bg-primary'
    case 'medium-high': return 'bg-warning text-dark'
    case 'high': return 'bg-danger'
    case 'toxic': return 'bg-dark'
    default: return 'bg-secondary'
  }
}

function riskLabel(risk: string): string {
  switch (risk) {
    case 'low': return 'Низкий'
    case 'medium': return 'Средний'
    case 'medium-high': return 'Выше среднего'
    case 'high': return 'Высокий'
    case 'toxic': return 'Токсичный'
    default: return risk || '—'
  }
}

function eventBadgeClass(event: string): string {
  if (event.includes('put')) return 'bg-warning text-dark'
  if (event.includes('call') || event.includes('досрочн')) return 'bg-danger'
  if (event.includes('погашение')) return 'bg-primary'
  return 'bg-secondary'
}
</script>
