<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 animate-fade-in">
    <!-- Параметры выпуска -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z"/></svg>
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

    <!-- Финансовые параметры -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
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

    <!-- Даты -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"/></svg>
        Даты
      </div>
      <div class="p-5">
        <!-- Life progress bar -->
        <div v-if="bond.days_to_maturity != null" class="mb-5">
          <div class="flex justify-between text-xs mb-2" style="color: var(--nla-text-muted)">
            <span>Размещение</span>
            <span>Погашение: {{ fmt.date(bond.matdate) }}</span>
          </div>
          <div class="h-2 rounded-full overflow-hidden" style="background: var(--nla-border)">
            <div class="h-full bg-primary-500 dark:bg-primary-400 rounded-full transition-all" :style="{ width: lifeProgress + '%' }"></div>
          </div>
          <div class="flex justify-between text-xs mt-1.5" style="color: var(--nla-text-muted)">
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

    <!-- Купонные параметры -->
    <div class="card overflow-hidden">
      <div class="panel-header">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"/></svg>
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
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{
  bond: Bond
}>()

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
  // MOEX sectype codes
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
