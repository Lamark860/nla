<template>
  <div class="max-w-md mx-auto mt-12">
    <div class="card p-8">
      <h1 class="text-xl font-bold text-slate-900 dark:text-white mb-6 text-center">
        {{ isRegister ? 'Регистрация' : 'Вход' }}
      </h1>

      <form @submit.prevent="submit" class="space-y-4">
        <div v-if="isRegister">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Имя</label>
          <input v-model="name" type="text" class="input w-full" placeholder="Ваше имя" />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Email</label>
          <input v-model="email" type="email" class="input w-full" placeholder="email@example.com" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Пароль</label>
          <input v-model="password" type="password" class="input w-full" placeholder="Минимум 6 символов" required />
        </div>

        <div v-if="error" class="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-500/10 px-3 py-2 rounded-lg">
          {{ error }}
        </div>

        <button
          type="submit"
          :disabled="submitting"
          class="btn-primary w-full text-sm"
        >
          {{ submitting ? 'Подождите...' : isRegister ? 'Зарегистрироваться' : 'Войти' }}
        </button>
      </form>

      <div class="mt-5 text-center">
        <button
          @click="isRegister = !isRegister; error = ''"
          class="text-sm text-primary-600 dark:text-primary-400 hover:underline"
        >
          {{ isRegister ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Регистрация' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const auth = useAuth()
const router = useRouter()

const isRegister = ref(false)
const email = ref('')
const password = ref('')
const name = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    let res: { ok: boolean; error?: string }
    if (isRegister.value) {
      res = await auth.register(email.value, password.value, name.value)
    } else {
      res = await auth.login(email.value, password.value)
    }
    if (res.ok) {
      router.push('/')
    } else {
      error.value = res.error || 'Произошла ошибка'
    }
  } finally {
    submitting.value = false
  }
}

// Redirect if already logged in
watch(() => auth.isLoggedIn.value, (loggedIn) => {
  if (loggedIn) router.push('/')
}, { immediate: true })
</script>
