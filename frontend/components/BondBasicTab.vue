<template>
  <div class="row g-4 animate-fade-in">
    <!-- Параметры выпуска -->
    <div class="col-lg-6">
      <div class="card overflow-hidden">
        <div class="panel-header">
          <i class="bi bi-info-circle"></i>
          Параметры выпуска
        </div>
        <div>
          <InfoRow label="Полное название" :value="bond.secname" />
          <InfoRow label="Англ. название" :value="bond.latname || '—'" />
          <InfoRow label="Тип облигации" :value="bond.bondtype_name || bond.bond_type || '—'" />
          <InfoRow label="Подтип" :value="bond.bondsubtype || '—'" />
          <InfoRow label="Категория" :value="bond.bond_category || '—'" />
          <InfoRow label="Статус" :value="tradingStatusText" />
          <InfoRow label="Тип ЦБ" :value="secTypeText" />
          <InfoRow label="Рыночный код" :value="bond.marketcode || '—'" />
          <InfoRow label="Код инструмента" :value="bond.instrid || '—'" />
          <InfoRow label="Сектор" :value="bond.sectorid || '—'" />
        </div>
      </div>
    </div>

    <!-- Финансовые параметры -->
    <div class="col-lg-6">
      <div class="card overflow-hidden">
        <div class="panel-header">
          <i class="bi bi-currency-dollar"></i>
          Финансовые параметры
        </div>
        <div>
          <InfoRow label="Номинал" :value="fmt.priceRub(bond.facevalue) + (bond.faceunit ? ' ' + currencyName(bond.faceunit) : '')" />
          <InfoRow label="Объём эмиссии" :value="bond.issuesize ? fmt.num(bond.issuesize) + ' шт.' : '—'" />
          <InfoRow label="Размещено" :value="bond.issuesize_placed ? fmt.num(bond.issuesize_placed) + ' шт.' : '—'" />
          <InfoRow label="Лот" :value="lotText" />
          <InfoRow label="Мин. шаг цены" :value="bond.minstep ? bond.minstep.toString() : '—'" />
          <InfoRow label="Кол-во знаков" :value="bond.decimals != null ? String(bond.decimals) : '—'" />
          <InfoRow label="Уровень листинга" :value="bond.listlevel ? String(bond.listlevel) : '—'" />
          <InfoRow label="Номинал на дату" :value="bond.facevalue_on_settle ? fmt.priceRub(bond.facevalue_on_settle) : '—'" />
          <InfoRow label="Валюта расчётов" :value="currencyName(bond.currencyid) || '—'" />
        </div>
      </div>
    </div>

    <!-- Даты -->
    <div class="col-lg-6">
      <div class="card overflow-hidden">
        <div class="panel-header">
          <i class="bi bi-calendar3"></i>
          Даты
        </div>
        <div class="p-4">
          <!-- Life progress bar -->
          <div v-if="bond.days_to_maturity != null" class="mb-4">
            <div class="d-flex justify-content-between small text-muted mb-2">
              <span>Размещение</span>
              <span>Погашение: {{ fmt.date(bond.matdate) }}</span>
            </div>
            <div class="progress" style="height: 8px">
              <div class="progress-bar bg-primary" role="progressbar" :style="{ width: lifeProgress + '%' }"></div>
            </div>
            <div class="d-flex justify-content-between small text-muted mt-1">
              <span>{{ lifeProgress }}% пройдено</span>
              <span>Осталось {{ fmt.daysToMaturity(bond.days_to_maturity) }}</span>
            </div>
          </div>
          <div>
            <InfoRow label="Погашение" :value="fmt.date(bond.matdate)" />
            <InfoRow label="До погашения" :value="fmt.daysToMaturity(bond.days_to_maturity)" />
            <InfoRow v-if="bond.offerdate && bond.offerdate !== 'None'" label="Оферта (PUT)" :value="fmt.date(bond.offerdate)" />
            <InfoRow v-if="bond.putoptiondate && bond.putoptiondate !== 'None'" label="PUT-опцион" :value="fmt.date(bond.putoptiondate)" />
            <InfoRow v-if="bond.calloptiondate && bond.calloptiondate !== 'None'" label="CALL-опцион" :value="fmt.date(bond.calloptiondate)" />
            <InfoRow v-if="bond.buybackdate && bond.buybackdate !== '0000-00-00' && bond.buybackdate !== 'None'" label="Buyback" :value="fmt.date(bond.buybackdate) + (bond.buybackprice ? ' по ' + fmt.percent(bond.buybackprice) : '')" />
            <InfoRow label="Следующий купон" :value="fmt.date(bond.next_coupon)" />
            <InfoRow label="Дата расчётов" :value="bond.settledate ? fmt.date(bond.settledate) : '—'" />
            <InfoRow label="Пред. торговый день" :value="bond.prevdate ? fmt.date(bond.prevdate) : '—'" />
          </div>
        </div>
      </div>
    </div>

    <!-- Купонные параметры -->
    <div class="col-lg-6">
      <div class="card overflow-hidden">
        <div class="panel-header">
          <i class="bi bi-cash-stack"></i>
          Купонные параметры
        </div>
        <div>
          <InfoRow label="Ставка купона" :value="fmt.percent(bond.coupon_percent)" />
          <InfoRow label="Сумма купона" :value="fmt.priceRub(bond.coupon_value)" />
          <InfoRow label="Период купона" :value="periodText" />
          <InfoRow label="Следующий купон" :value="fmt.date(bond.next_coupon)" />
          <InfoRow label="НКД" :value="fmt.priceRub(bond.accrued_int)" />
          <InfoRow label="НКД от номинала" :value="nkdPctOfNominal" />
          <InfoRow label="Текущая доходность" :value="currentYield" />
          <InfoRow label="Пред. WAP" :value="bond.prevwaprice != null ? fmt.percent(bond.prevwaprice) : '—'" />
          <InfoRow label="Доходность по WAP" :value="bond.yieldatprevwaprice != null ? fmt.percent(bond.yieldatprevwaprice) : '—'" />
          <InfoRow label="Тип" :value="couponTypeText" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{ bond: Bond }>()
const fmt = useFormat()

const lifeProgress = computed(() => {
  if (!props.bond.days_to_maturity || !props.bond.matdate) return 50
  const totalDays = Math.max(props.bond.days_to_maturity + 365, 1825)
  const elapsed = totalDays - props.bond.days_to_maturity
  return Math.min(99, Math.max(1, Math.round((elapsed / totalDays) * 100)))
})

const tradingStatusText = computed(() => {
  switch (props.bond.trading_status) {
    case 'T': return 'Торгуется'
    case 'N': return 'Не торгуется'
    case 'S': return 'Приостановлено'
    default: return props.bond.trading_status || '—'
  }
})

const secTypeText = computed(() => {
  const t = props.bond.sectype || props.bond.bond_type
  if (!t) return '—'
  const map: Record<string, string> = {
    '3': 'ОФЗ', '4': 'Субфедеральная', '5': 'Муниципальная',
    '6': 'Корпоративная (индексируемая)', '7': 'Корпоративная (ипотечная)',
    '8': 'Корпоративная', 'C': 'ЦБ РФ', 'D': 'Еврооблигация'
  }
  return map[t] || t
})

const lotText = computed(() => {
  if (!props.bond.lotsize) return '—'
  const lotRub = props.bond.lotvalue ? ` (${fmt.priceRub(props.bond.lotvalue)})` : ''
  return `${props.bond.lotsize} шт.${lotRub}`
})

const periodText = computed(() => {
  const p = props.bond.coupon_period
  if (!p) return '—'
  let name = ''
  if (p >= 27 && p <= 33) name = ' (ежемесячный)'
  else if (p >= 85 && p <= 95) name = ' (ежеквартальный)'
  else if (p >= 175 && p <= 190) name = ' (полугодовой)'
  else if (p >= 355 && p <= 370) name = ' (годовой)'
  return `${p} дн.${name}`
})

const nkdPctOfNominal = computed(() => {
  if (!props.bond.accrued_int || !props.bond.facevalue) return '—'
  return ((props.bond.accrued_int / props.bond.facevalue) * 100).toFixed(3) + '%'
})

const currentYield = computed(() => {
  if (!props.bond.coupon_value || !props.bond.price_rub || props.bond.price_rub <= 0) return '—'
  const cpy = props.bond.coupon_period > 0 ? 365 / props.bond.coupon_period : 2
  return ((props.bond.coupon_value * cpy / props.bond.price_rub) * 100).toFixed(2) + '%'
})

const couponTypeText = computed(() => {
  if (props.bond.is_float) return 'Плавающий (флоатер)'
  if (props.bond.is_indexed) return 'Индексируемый'
  return 'Фиксированный'
})

function currencyName(code: string): string {
  if (!code) return ''
  const map: Record<string, string> = { SUR: 'RUB', USD: 'USD', EUR: 'EUR', CNY: 'CNY' }
  return map[code] || code
}
</script>
