<template>
  <div class="row justify-content-center mt-5">
    <div class="col-12 col-sm-8 col-md-6 col-lg-4">
      <div class="card p-4">
        <h1 class="h5 fw-bold text-center mb-4">
          {{ isRegister ? 'Регистрация' : 'Вход' }}
        </h1>

        <form @submit.prevent="submit" class="d-flex flex-column gap-3">
          <div v-if="isRegister">
            <label class="form-label small">Имя</label>
            <input v-model="name" type="text" class="form-control" placeholder="Ваше имя" />
          </div>
          <div>
            <label class="form-label small">Email</label>
            <input v-model="email" type="email" class="form-control" placeholder="email@example.com" required />
          </div>
          <div>
            <label class="form-label small">Пароль</label>
            <input v-model="password" type="password" class="form-control" placeholder="Минимум 6 символов" required />
          </div>

          <div v-if="error" class="alert alert-danger small py-2 mb-0">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="submitting"
            class="btn btn-primary w-100"
          >
            {{ submitting ? 'Подождите...' : isRegister ? 'Зарегистрироваться' : 'Войти' }}
          </button>
        </form>

        <div class="mt-4 text-center">
          <button
            @click="isRegister = !isRegister; error = ''"
            class="btn btn-link btn-sm text-decoration-none"
            style="color: var(--nla-text-secondary)"
          >
            {{ isRegister ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Регистрация' }}
          </button>
        </div>
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

watch(() => auth.isLoggedIn.value, (loggedIn) => {
  if (loggedIn) router.push('/')
}, { immediate: true })
</script>
