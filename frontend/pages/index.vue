<template>
  <div>
    <!-- Page header -->
    <div class="mb-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Облигации Московской биржи</h1>
        </div>
        <div class="flex items-center gap-2">
          <NuxtLink to="/bonds/by-issuer" class="btn-secondary text-xs">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
            По эмитентам
          </NuxtLink>
          <NuxtLink to="/bonds/monthly" class="btn-secondary text-xs">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
            Месячные купоны
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="pending" class="card p-16 text-center">
      <div class="inline-block w-6 h-6 border-2 border-primary-200 border-t-primary-600 dark:border-primary-800 dark:border-t-primary-400 rounded-full animate-spin"></div>
      <p class="mt-4 text-xs text-slate-400 dark:text-slate-500">Загрузка облигаций…</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="card p-10 text-center">
      <p class="text-red-600 dark:text-red-400 text-sm">{{ error.message || 'Ошибка загрузки' }}</p>
      <button class="btn-primary mt-4 text-sm" @click="refresh()">Повторить</button>
    </div>

    <!-- Data -->
    <BondTable
      v-else-if="data"
      :bonds="data.data"
      :meta="data.meta"
      :sort="currentSort"
      @sort="onSort"
      @page="onPage"
    />
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

function onSort(sort: string) {
  currentSort.value = sort
  currentPage.value = 1
}

function onPage(page: number) {
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

useHead({ title: 'Облигации — NLA' })
</script>
