<template>
  <div class="animate-fade-in">
    <p class="small text-muted mb-4">
      Внешние аналитические ресурсы для <span class="font-monospace fw-medium">{{ secid }}</span>
    </p>

    <div class="row g-3">
      <!-- T-Bank -->
      <div class="col-md-6 col-lg-4">
        <div class="card h-100 p-4 text-center">
          <div class="mb-3">
            <div class="mx-auto d-flex align-items-center justify-content-center rounded-3" style="width: 56px; height: 56px; background: var(--nla-primary-subtle)">
              <i class="bi bi-bank2 fs-4" style="color: var(--nla-primary)"></i>
            </div>
          </div>
          <h6 class="fw-semibold mb-1">Т-Банк Инвестиции</h6>
          <p class="small text-muted mb-3">Торговля, аналитика, прогнозы</p>
          <a :href="tbankUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline-primary btn-sm">
            Открыть <i class="bi bi-box-arrow-up-right ms-1"></i>
          </a>
        </div>
      </div>

      <!-- Dohod.ru -->
      <div class="col-md-6 col-lg-4">
        <div class="card h-100 p-4 text-center">
          <div class="mb-3">
            <div class="mx-auto d-flex align-items-center justify-content-center rounded-3" style="width: 56px; height: 56px; background: var(--nla-success-light)">
              <i class="bi bi-graph-up-arrow fs-4" style="color: var(--nla-success)"></i>
            </div>
          </div>
          <h6 class="fw-semibold mb-1">Dohod.ru</h6>
          <p class="small text-muted mb-3">Фундаментальный анализ эмитента</p>
          <a :href="dohodUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline-primary btn-sm">
            Открыть <i class="bi bi-box-arrow-up-right ms-1"></i>
          </a>
        </div>
      </div>

      <!-- MOEX -->
      <div class="col-md-6 col-lg-4">
        <div class="card h-100 p-4 text-center">
          <div class="mb-3">
            <div class="mx-auto d-flex align-items-center justify-content-center rounded-3" style="width: 56px; height: 56px; background: var(--nla-warning-light)">
              <i class="bi bi-building fs-4" style="color: var(--nla-warning)"></i>
            </div>
          </div>
          <h6 class="fw-semibold mb-1">Московская Биржа</h6>
          <p class="small text-muted mb-3">Официальные данные MOEX</p>
          <a :href="moexUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline-primary btn-sm">
            Открыть <i class="bi bi-box-arrow-up-right ms-1"></i>
          </a>
        </div>
      </div>

      <!-- Smart-lab -->
      <div class="col-md-6 col-lg-4">
        <div class="card h-100 p-4 text-center">
          <div class="mb-3">
            <div class="mx-auto d-flex align-items-center justify-content-center rounded-3" style="width: 56px; height: 56px; background: var(--nla-danger-light)">
              <i class="bi bi-people fs-4" style="color: var(--nla-danger)"></i>
            </div>
          </div>
          <h6 class="fw-semibold mb-1">Smart-lab</h6>
          <p class="small text-muted mb-3">Сообщество инвесторов</p>
          <a :href="smartlabUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline-primary btn-sm">
            Открыть <i class="bi bi-box-arrow-up-right ms-1"></i>
          </a>
        </div>
      </div>

      <!-- Rusbonds -->
      <div class="col-md-6 col-lg-4">
        <div class="card h-100 p-4 text-center">
          <div class="mb-3">
            <div class="mx-auto d-flex align-items-center justify-content-center rounded-3" style="width: 56px; height: 56px; background: var(--nla-info, rgba(6,182,212,0.15))">
              <i class="bi bi-journal-richtext fs-4" style="color: var(--nla-info)"></i>
            </div>
          </div>
          <h6 class="fw-semibold mb-1">Rusbonds</h6>
          <p class="small text-muted mb-3">Параметры выпуска облигации</p>
          <a :href="rusbondsUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline-primary btn-sm">
            Открыть <i class="bi bi-box-arrow-up-right ms-1"></i>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ secid: string }>()
const tbankUrl = computed(() => `https://www.tbank.ru/invest/bonds/${props.secid}/`)
const dohodUrl = computed(() => `https://analytics.dohod.ru/bond/${props.secid}`)
const moexUrl = computed(() => `https://www.moex.com/ru/issue.aspx?board=TQCB&code=${props.secid}`)
const smartlabUrl = computed(() => `https://smart-lab.ru/q/bonds/${props.secid}/`)
const rusbondsUrl = computed(() => `https://www.rusbonds.ru/srch_simple.asp?go=1&nick=${props.secid}`)
</script>
