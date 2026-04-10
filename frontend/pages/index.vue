<template>
  <div>
    <!-- Page header -->
    <div class="d-flex flex-column flex-sm-row align-items-sm-center justify-content-sm-between gap-2 mb-4">
      <h1 class="h4 fw-bold mb-0">Облигации Московской биржи</h1>
      <div class="d-flex align-items-center gap-2">
        <NuxtLink to="/bonds/by-issuer" class="btn btn-outline-secondary btn-sm">
          <i class="bi bi-grid me-1"></i>По эмитентам
        </NuxtLink>
        <NuxtLink to="/bonds/monthly" class="btn btn-outline-secondary btn-sm">
          <i class="bi bi-calendar3 me-1"></i>Месячные купоны
        </NuxtLink>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="pending" class="card p-5 text-center">
      <div class="spinner-border" role="status"><span class="visually-hidden">Загрузка…</span></div>
      <p class="mt-3 small text-muted">Загрузка облигаций…</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="card p-5 text-center">
      <p class="text-danger small">{{ error.message || 'Ошибка загрузки' }}</p>
      <button class="btn btn-primary btn-sm mt-3" @click="refresh()">Повторить</button>
    </div>

    <!-- Data -->
    <BondTable v-else-if="data" :bonds="data.data" :meta="data.meta" :sort="currentSort" @sort="onSort" @page="onPage" />
  </div>
</template>

<script setup lang="ts">
const api = useApi()
const currentSort = ref('best')
const currentPage = ref(1)

const { data, pending, error, refresh } = useAsyncData(
  'bonds',
  () => api.getBonds(currentPage.value, 20, currentSort.value),
  { watch: [currentSort, currentPage] }
)

function onSort(sort: string) { currentSort.value = sort; currentPage.value = 1 }
function onPage(page: number) { currentPage.value = page; window.scrollTo({ top: 0, behavior: 'smooth' }) }

useHead({ title: 'Облигации — NLA' })
</script>
