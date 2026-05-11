<!--
  BondInfoBasic.vue — заменяет BondBasicTab.vue.
  Та же сетка 2×2, но через Panel + InfoRow с :mono.
  Прогресс-бар жизни облигации стал «штатным» компонентом — видно ярче.
-->
<template>
  <div class="grid-2x2">
    <Panel title="Параметры выпуска" icon="info-circle">
      <InfoRow label="Полное название" :value="bond.secname" />
      <InfoRow label="Англ. название" :value="bond.latname || '—'" />
      <InfoRow label="ISIN" :value="bond.isin || '—'" mono />
      <InfoRow label="Гос. рег. номер" :value="bond.regnumber || '—'" mono />
      <InfoRow label="Тип облигации" :value="bond.bondtype_name || bond.bond_type || '—'" />
      <InfoRow label="Подтип" :value="bond.bondsubtype || '—'" />
      <InfoRow label="Категория" :value="bond.bond_category || '—'" />
      <InfoRow label="Статус" :value="tradingStatusText" :tone="bond.trading_status === 'T' ? 'success' : undefined" />
      <InfoRow label="Категория риска">
        <template #value>
          <span class="risk-pill" :class="`risk-pill--${bond.risk_category || 'unknown'}`">{{ riskLabel }}</span>
        </template>
      </InfoRow>
      <InfoRow label="Тип ЦБ" :value="secTypeText" />
      <InfoRow label="Рыночный код" :value="bond.marketcode || '—'" mono />
      <InfoRow label="Площадка торгов" :value="bond.boardname || bond.boardid || '—'" />
      <InfoRow label="Код инструмента" :value="bond.instrid || '—'" mono />
      <InfoRow label="Сектор" :value="bond.sectorid || '—'" />
    </Panel>

    <Panel title="Финансовые параметры" icon="cash-coin">
      <InfoRow label="Номинал" :value="facevalueText" mono />
      <InfoRow label="Объём эмиссии" :value="bond.issuesize ? fmt.num(bond.issuesize) + ' шт.' : '—'" mono />
      <InfoRow label="Размещено" :value="bond.issuesize_placed ? fmt.num(bond.issuesize_placed) + ' шт.' : '—'" mono />
      <InfoRow label="Лот" :value="lotText" mono />
      <InfoRow label="Мин. шаг цены" :value="bond.minstep ? bond.minstep.toString() : '—'" mono />
      <InfoRow label="Кол-во знаков" :value="bond.decimals != null ? String(bond.decimals) : '—'" mono />
      <InfoRow label="Уровень листинга" :value="bond.listlevel ? String(bond.listlevel) : '—'" />
      <InfoRow label="Номинал на дату" :value="bond.facevalue_on_settle ? fmt.priceRub(bond.facevalue_on_settle) : '—'" mono />
      <InfoRow label="Валюта расчётов" :value="currencyName(bond.currencyid) || '—'" />
    </Panel>

    <Panel title="Даты" icon="calendar3">
      <!-- Life progress -->
      <div v-if="bond.days_to_maturity != null" class="bond-life">
        <div class="bond-life__head">
          <span class="bond-life__lbl">Размещение</span>
          <span class="bond-life__lbl bond-life__lbl--right">
            Погашение · <strong>{{ fmt.dateShort(bond.matdate) }}</strong>
          </span>
        </div>
        <div class="bond-life__bar">
          <div class="bond-life__fill" :style="{ width: lifeProgress + '%' }"></div>
          <div class="bond-life__marker" :style="{ left: lifeProgress + '%' }" :title="`${lifeProgress}% пройдено`"></div>
        </div>
        <div class="bond-life__foot">
          <span><strong>{{ lifeProgress }}%</strong> пройдено</span>
          <span>осталось <strong>{{ fmt.daysToMaturity(bond.days_to_maturity) }}</strong></span>
        </div>
      </div>

      <InfoRow label="Погашение" :value="fmt.date(bond.matdate)" mono />
      <InfoRow label="До погашения" :value="fmt.daysToMaturity(bond.days_to_maturity)" />
      <InfoRow v-if="hasDate(bond.offerdate)" label="Оферта (PUT)" :value="fmt.date(bond.offerdate)" mono />
      <InfoRow v-if="hasDate(bond.putoptiondate)" label="PUT-опцион" :value="fmt.date(bond.putoptiondate)" mono />
      <InfoRow v-if="hasDate(bond.calloptiondate)" label="CALL-опцион" :value="fmt.date(bond.calloptiondate)" mono />
      <InfoRow v-if="hasDate(bond.buybackdate, true)" label="Buyback" :value="buybackText" mono />
      <InfoRow label="Следующий купон" :value="fmt.date(bond.next_coupon)" mono />
      <InfoRow v-if="bond.settledate" label="Дата расчётов" :value="fmt.date(bond.settledate)" mono />
      <InfoRow v-if="bond.prevdate" label="Пред. торговый день" :value="fmt.date(bond.prevdate)" mono />
    </Panel>

    <Panel title="Купонные параметры" icon="cash-stack">
      <InfoRow label="Ставка купона" :value="fmt.percent(bond.coupon_percent)" mono />
      <InfoRow label="Сумма купона" :value="fmt.priceRub(bond.coupon_value)" mono />
      <InfoRow label="Период купона" :value="periodText" />
      <InfoRow label="Следующий купон" :value="fmt.date(bond.next_coupon)" mono />
      <InfoRow label="НКД" :value="fmt.priceRub(bond.accrued_int)" mono />
      <InfoRow label="НКД от номинала" :value="nkdPctOfNominal" mono />
      <InfoRow label="Текущая доходность" :value="currentYield" mono tone="primary" />
      <InfoRow label="Пред. WAP" :value="bond.prevwaprice != null ? fmt.percent(bond.prevwaprice) : '—'" mono />
      <InfoRow label="Доходность по WAP" :value="bond.yieldatprevwaprice != null ? fmt.percent(bond.yieldatprevwaprice) : '—'" mono />
      <InfoRow label="Тип купона" :value="couponTypeText" />
    </Panel>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'
import Panel from './Panel.vue'
import InfoRow from './InfoRow.vue'

const props = defineProps<{ bond: Bond }>()
const fmt = useFormat()

const lifeProgress = computed(() => {
  if (!props.bond.days_to_maturity || !props.bond.matdate) return 50
  const totalDays = Math.max(props.bond.days_to_maturity + 365, 1825)
  const elapsed = totalDays - props.bond.days_to_maturity
  return Math.min(99, Math.max(1, Math.round((elapsed / totalDays) * 100)))
})
const tradingStatusText = computed(() => ({
  T: 'Торгуется', N: 'Не торгуется', S: 'Приостановлено'
}[props.bond.trading_status as string] || props.bond.trading_status || '—'))
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
const facevalueText = computed(() =>
  fmt.priceRub(props.bond.facevalue) + (props.bond.faceunit ? ' ' + currencyName(props.bond.faceunit) : '')
)
const lotText = computed(() => {
  if (!props.bond.lotsize) return '—'
  const lotRub = props.bond.lotvalue ? ` (${fmt.priceRub(props.bond.lotvalue)})` : ''
  return `${props.bond.lotsize} шт.${lotRub}`
})
const periodText = computed(() => {
  const p = props.bond.coupon_period
  if (!p) return '—'
  let name = ''
  if (p >= 27 && p <= 33) name = ' · ежемесячный'
  else if (p >= 85 && p <= 95) name = ' · ежеквартальный'
  else if (p >= 175 && p <= 190) name = ' · полугодовой'
  else if (p >= 355 && p <= 370) name = ' · годовой'
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
const buybackText = computed(() =>
  fmt.date(props.bond.buybackdate) + (props.bond.buybackprice ? ' · по ' + fmt.percent(props.bond.buybackprice) : '')
)

const riskLabel = computed(() => {
  const map: Record<string, string> = {
    safe: 'Низкий',
    stable: 'Стабильный',
    moderate: 'Умеренный',
    speculative: 'Спекулятивный',
    risky: 'Высокий',
    toxic: 'Токсичный',
    junk: 'Мусорный',
  }
  return map[props.bond.risk_category] || props.bond.risk_category || '—'
})

function hasDate(v?: string | null, allowZero = false): boolean {
  if (!v || v === 'None') return false
  if (!allowZero && v === '0000-00-00') return false
  return true
}
function currencyName(code: string) {
  const map: Record<string, string> = { SUR: 'RUB', USD: 'USD', EUR: 'EUR', CNY: 'CNY' }
  return map[code] || code || ''
}
</script>

<style scoped>
.grid-2x2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
@media (max-width: 992px) {
  .grid-2x2 { grid-template-columns: 1fr; }
}

/* Bond life progress */
.bond-life {
  padding: 16px 18px 14px;
  border-bottom: 1px solid var(--nla-border-light);
}
.bond-life__head,
.bond-life__foot {
  display: flex;
  justify-content: space-between;
  font: 500 11.5px / 1.3 var(--nla-font);
  color: var(--nla-text-muted);
}
.bond-life__head { margin-bottom: 8px; }
.bond-life__foot { margin-top: 8px; }
.bond-life__foot strong { color: var(--nla-text); font-weight: 600; }
.bond-life__lbl strong { color: var(--nla-text); font-weight: 600; font-family: var(--nla-font-mono); }
.bond-life__lbl--right { text-align: right; }

.bond-life__bar {
  position: relative;
  height: 8px;
  background: var(--nla-bg-subtle);
  border-radius: var(--nla-radius-pill);
  overflow: visible;
}
.bond-life__fill {
  position: absolute;
  inset: 0 auto 0 0;
  height: 100%;
  border-radius: var(--nla-radius-pill);
  background: linear-gradient(90deg,
    color-mix(in oklab, var(--nla-primary) 50%, transparent),
    var(--nla-primary)
  );
}
.bond-life__marker {
  position: absolute;
  top: 50%;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--nla-bg-card);
  border: 3px solid var(--nla-primary);
  transform: translate(-50%, -50%);
  box-shadow: 0 2px 6px rgba(91, 58, 168, 0.35);
}

/* Risk pill */
.risk-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: var(--nla-radius-sm);
  font: 600 11px/1.4 var(--nla-font);
  letter-spacing: 0.02em;
}
.risk-pill::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}
.risk-pill--safe,
.risk-pill--stable     { background: var(--nla-success-light); color: var(--nla-success); }
.risk-pill--moderate   { background: var(--nla-primary-light); color: var(--nla-primary-ink); }
.risk-pill--speculative,
.risk-pill--risky      { background: var(--nla-warning-light); color: var(--nla-warning); }
.risk-pill--toxic,
.risk-pill--junk       { background: var(--nla-danger-light); color: var(--nla-danger); }
.risk-pill--unknown    { background: var(--nla-bg-subtle); color: var(--nla-text-muted); }
[data-theme="dark"] .risk-pill--moderate { color: var(--nla-primary); }
</style>
