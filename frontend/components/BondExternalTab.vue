<template>
  <div class="space-y-6 animate-fade-in">
    <p class="text-xs text-slate-400 dark:text-slate-500">
      Внешние аналитические ресурсы для <span class="font-mono font-medium">{{ secid }}</span>
    </p>

    <!-- T-Bank -->
    <div class="card overflow-hidden">
      <div class="px-5 py-3 border-b border-slate-200/80 dark:border-white/[0.06] flex items-center justify-between">
        <h3 class="section-title">Т-Банк Инвестиции</h3>
        <a
          :href="tbankUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs text-primary-600 hover:text-primary-500 dark:text-primary-400 font-medium"
        >
          Открыть ↗
        </a>
      </div>
      <div class="relative" style="padding-bottom: 60%; min-height: 400px;">
        <iframe
          :src="tbankUrl"
          class="absolute inset-0 w-full h-full border-0"
          sandbox="allow-scripts allow-same-origin"
          loading="lazy"
          @error="tbankError = true"
        />
        <div v-if="tbankError" class="absolute inset-0 flex items-center justify-center bg-slate-50 dark:bg-surface-900">
          <div class="text-center">
            <p class="text-slate-400 text-sm mb-3">Не удалось загрузить</p>
            <a :href="tbankUrl" target="_blank" rel="noopener noreferrer" class="btn-primary text-sm">
              Открыть на сайте ↗
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Dohod.ru -->
    <div class="card overflow-hidden">
      <div class="px-5 py-3 border-b border-slate-200/80 dark:border-white/[0.06] flex items-center justify-between">
        <h3 class="section-title">Доход (dohod.ru)</h3>
        <a
          :href="dohodUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs text-primary-600 hover:text-primary-500 dark:text-primary-400 font-medium"
        >
          Открыть ↗
        </a>
      </div>
      <div class="relative" style="padding-bottom: 60%; min-height: 400px;">
        <iframe
          :src="dohodUrl"
          class="absolute inset-0 w-full h-full border-0"
          sandbox="allow-scripts allow-same-origin"
          loading="lazy"
          @error="dohodError = true"
        />
        <div v-if="dohodError" class="absolute inset-0 flex items-center justify-center bg-slate-50 dark:bg-surface-900">
          <div class="text-center">
            <p class="text-slate-400 text-sm mb-3">Не удалось загрузить</p>
            <a :href="dohodUrl" target="_blank" rel="noopener noreferrer" class="btn-primary text-sm">
              Открыть на dohod.ru ↗
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ secid: string }>()

const tbankUrl = computed(() => `https://www.tbank.ru/invest/bonds/${props.secid}/`)
const dohodUrl = computed(() => `https://analytics.dohod.ru/bond/${props.secid}`)
const tbankError = ref(false)
const dohodError = ref(false)
</script>
