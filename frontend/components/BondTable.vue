<template>
  <div class="card overflow-hidden">
    <!-- Header with sort -->
    <div class="card-header d-flex align-items-center justify-content-between gap-3 flex-wrap py-3">
      <div class="d-flex align-items-center gap-3">
        <select v-model="currentSort" class="form-select form-select-sm" style="width: auto" @change="$emit('sort', currentSort)">
          <option value="best">Лучшие</option>
          <option value="coupon_desc">Купон ↓</option>
          <option value="coupon_asc">Купон ↑</option>
          <option value="yield_desc">Доходность ↓</option>
          <option value="yield_asc">Доходность ↑</option>
          <option value="maturity_asc">Погашение ↑</option>
          <option value="maturity_desc">Погашение ↓</option>
          <option value="volume_desc">Объём ↓</option>
          <option value="duration_asc">Дюрация ↑</option>
          <option value="duration_desc">Дюрация ↓</option>
        </select>
        <span class="small text-muted">Найдено: <span class="fw-semibold">{{ meta.total }}</span></span>
      </div>
    </div>

    <!-- Table -->
    <div class="table-responsive">
      <table class="data-table" style="min-width: 1100px">
        <thead>
          <tr>
            <th class="text-start" style="width: 200px">НАЗВАНИЕ</th>
            <th class="text-end" style="width: 85px">ДОХ.</th>
            <th class="text-end" style="width: 95px">ЦЕНА</th>
            <th class="text-end d-none d-md-table-cell" style="width: 100px">ИЗМЕНЕНИЕ</th>
            <th class="text-end d-none d-lg-table-cell" style="width: 110px">ОБЪЕМ</th>
            <th class="text-end d-none d-lg-table-cell" style="width: 60px">НКД</th>
            <th class="text-end d-none d-md-table-cell" style="width: 100px">КУПОН</th>
            <th class="text-end d-none d-lg-table-cell" style="width: 120px">ПОГАШЕНИЕ</th>
            <th class="text-center d-none d-xl-table-cell" style="width: 120px">СТАТУС</th>
            <th class="text-center d-none d-xl-table-cell" style="width: 60px">★</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="bond in bonds" :key="bond.secid" class="cursor-pointer" style="cursor: pointer" @click="$router.push(`/bonds/${bond.secid}`)">
            <!-- НАЗВАНИЕ -->
            <td>
              <div>
                <span :class="categoryBadge(bond.bond_category)" class="badge-sm mb-1 d-inline-block">
                  {{ categoryLabel(bond.bond_category, bond.is_float, bond.is_indexed) }}
                </span>
                <div class="fw-medium">{{ bond.shortname }}</div>
                <div class="small text-muted font-monospace">{{ bond.isin }}</div>
              </div>
            </td>
            <!-- ДОХ. -->
            <td class="text-end font-monospace">
              <span :class="yieldColor(bond.yield)" class="fw-semibold">{{ bond.yield != null ? fmt.percent(bond.yield) : '—' }}</span>
            </td>
            <!-- ЦЕНА -->
            <td class="text-end font-monospace">
              <div>{{ bond.last != null ? fmt.percent(bond.last) : '—' }}</div>
              <div v-if="bond.price_rub" class="small text-success">{{ fmt.priceRub(bond.price_rub) }}</div>
            </td>
            <!-- ИЗМЕНЕНИЕ -->
            <td class="text-end d-none d-md-table-cell font-monospace">
              <span :class="changeColor(bond.last_change_prcnt)">{{ formatChange(bond.last_change_prcnt) }}</span>
            </td>
            <!-- ОБЪЕМ -->
            <td class="text-end d-none d-lg-table-cell font-monospace">
              <div>{{ fmt.volume(bond.value_today_rub) }}</div>
              <div v-if="bond.vol_today > 0" class="small text-muted">штук: {{ fmt.num(bond.vol_today) }}</div>
            </td>
            <!-- НКД -->
            <td class="text-end d-none d-lg-table-cell font-monospace text-muted">
              {{ bond.accrued_int > 0 ? fmt.num(bond.accrued_int, 2) : '—' }}
            </td>
            <!-- КУПОН -->
            <td class="text-end d-none d-md-table-cell font-monospace">
              <div class="d-flex align-items-center justify-content-end gap-1">
                <span>{{ fmt.percent(bond.coupon_display) }}</span>
                <span v-if="bond.is_float" class="badge bg-info" style="font-size: 9px">F</span>
                <span v-if="bond.is_indexed" class="badge bg-secondary" style="font-size: 9px">I</span>
              </div>
              <div class="small text-muted">{{ bond.next_coupon ? fmt.dateShort(bond.next_coupon) : '' }}</div>
            </td>
            <!-- ПОГАШЕНИЕ -->
            <td class="text-end d-none d-lg-table-cell">
              <div class="font-monospace small">{{ fmt.date(bond.matdate) }}</div>
              <div :class="bond.days_to_maturity != null && bond.days_to_maturity < 365 ? 'text-warning' : 'text-muted'" class="small">
                {{ fmt.daysToMaturity(bond.days_to_maturity) }}
              </div>
            </td>
            <!-- СТАТУС -->
            <td class="text-center d-none d-xl-table-cell">
              <span :class="tradingStatusBadge(bond.trading_status)" class="badge">{{ tradingStatusText(bond.trading_status) }}</span>
            </td>
            <!-- ДЕЙСТВИЯ -->
            <td class="text-center d-none d-xl-table-cell">
              <button
                v-if="auth.isLoggedIn.value"
                class="btn btn-link p-0"
                :class="favorites.isFavorite(bond.secid) ? 'text-warning' : 'text-muted'"
                :title="favorites.isFavorite(bond.secid) ? 'Убрать из избранного' : 'В избранное'"
                @click.stop="favorites.toggle(bond.secid)"
              >
                <i :class="favorites.isFavorite(bond.secid) ? 'bi-star-fill' : 'bi-star'" class="bi"></i>
              </button>
              <NuxtLink v-else :to="`/bonds/${bond.secid}`" class="btn btn-sm btn-outline-secondary rounded-circle" @click.stop>
                <i class="bi bi-info-circle"></i>
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="card-footer d-flex align-items-center justify-content-between py-2">
      <button :disabled="meta.page <= 1" class="btn btn-outline-secondary btn-sm" @click="$emit('page', meta.page - 1)">← Назад</button>
      <span class="small text-muted font-monospace">{{ (meta.page - 1) * meta.per_page + 1 }}–{{ Math.min(meta.page * meta.per_page, meta.total) }} из {{ meta.total }}</span>
      <button :disabled="meta.page * meta.per_page >= meta.total" class="btn btn-outline-secondary btn-sm" @click="$emit('page', meta.page + 1)">Вперёд →</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bond } from '~/composables/useApi'

const props = defineProps<{ bonds: Bond[]; meta: { page: number; per_page: number; total: number }; sort: string }>()
defineEmits<{ sort: [value: string]; page: [value: number] }>()

const fmt = useFormat()
const auth = useAuth()
const favorites = useFavorites()
const currentSort = ref(props.sort ?? 'best')

watch(() => props.sort, (val) => { if (val) currentSort.value = val })

function yieldColor(y: number | null | undefined): string {
  if (y == null) return 'text-muted'
  if (y >= 12) return 'text-success'
  if (y >= 8) return 'text-primary'
  return ''
}

function categoryBadge(cat: string | undefined): string {
  switch (cat) {
    case 'ОФЗ': return 'bg-primary text-white'
    default: return 'bg-secondary text-white'
  }
}

function categoryLabel(cat: string | undefined, isFloat: boolean, isIndexed: boolean): string {
  if (isFloat) return 'Флоатер'
  if (isIndexed) return 'Индексируемая'
  return 'Фикс с известным купоном'
}

function changeColor(val: number | null | undefined): string {
  if (val == null || val === 0) return 'text-muted'
  return val > 0 ? 'text-success' : 'text-danger'
}

function formatChange(val: number | null | undefined): string {
  if (val == null) return '—'
  return (val > 0 ? '+' : '') + val.toFixed(2) + '%'
}

function tradingStatusBadge(status: string | undefined): string {
  switch (status) {
    case 'T': return 'bg-success'
    case 'S': return 'bg-warning text-dark'
    default: return 'bg-secondary'
  }
}

function tradingStatusText(status: string | undefined): string {
  switch (status) {
    case 'T': return 'Торги идут'
    case 'S': return 'Приостановлены'
    default: return 'Не ведутся'
  }
}
</script>
