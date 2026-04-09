<template>
  <div class="flex gap-6 h-[calc(100vh-theme(spacing.14)-theme(spacing.12)-theme(spacing.16))]">
    <!-- Sidebar: Sessions -->
    <div class="w-72 shrink-0 flex flex-col card overflow-hidden">
      <div class="p-4 border-b border-slate-200/80 dark:border-white/[0.06]">
        <h2 class="text-sm font-semibold text-slate-900 dark:text-white">Чаты</h2>
        <div class="flex gap-2 mt-3">
          <button
            v-for="agent in agents"
            :key="agent.type"
            class="flex-1 btn-secondary text-xs py-2"
            @click="createSession(agent.type, agent.name)"
          >
            + {{ agent.name }}
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto">
        <div
          v-for="s in sessions"
          :key="s.session_id"
          :class="[
            'px-4 py-3 cursor-pointer border-b border-slate-100/80 dark:border-white/[0.03] transition-colors group',
            activeSessionId === s.session_id
              ? 'bg-primary-50/50 dark:bg-primary-500/5'
              : 'hover:bg-slate-50 dark:hover:bg-white/[0.02]'
          ]"
          @click="selectSession(s.session_id)"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <p class="text-sm font-medium text-slate-900 dark:text-white truncate">{{ s.title || 'Новый чат' }}</p>
              <p class="text-[11px] text-slate-400 dark:text-slate-500 mt-0.5">
                {{ agentName(s.agent_type) }} · {{ formatTime(s.updated_at) }}
              </p>
            </div>
            <button
              class="opacity-0 group-hover:opacity-100 p-1 text-slate-300 hover:text-red-500 dark:text-slate-600 dark:hover:text-red-400 transition-all shrink-0"
              @click.stop="deleteSession(s.session_id)"
              title="Удалить"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
            </button>
          </div>
        </div>

        <div v-if="sessions.length === 0" class="p-4 text-center text-xs text-slate-400 dark:text-slate-500">
          Нет активных чатов
        </div>
      </div>
    </div>

    <!-- Main: Chat -->
    <div class="flex-1 flex flex-col card overflow-hidden">
      <template v-if="activeSessionId">
        <!-- Header -->
        <div class="px-5 py-3 border-b border-slate-200/80 dark:border-white/[0.06] flex items-center justify-between">
          <div>
            <h3 class="text-sm font-semibold text-slate-900 dark:text-white">{{ activeSession?.title }}</h3>
            <p class="text-[11px] text-slate-400 dark:text-slate-500">{{ agentName(activeSession?.agent_type || '') }}</p>
          </div>
        </div>

        <!-- Messages -->
        <div ref="messagesContainer" class="flex-1 overflow-y-auto p-5 space-y-4">
          <div
            v-for="(msg, i) in messages"
            :key="i"
            :class="[
              'max-w-[85%] rounded-2xl px-4 py-3 text-sm leading-relaxed',
              msg.role === 'user'
                ? 'ml-auto bg-primary-500 text-white'
                : 'mr-auto bg-slate-100 dark:bg-surface-850 text-slate-800 dark:text-slate-200'
            ]"
          >
            <div class="whitespace-pre-wrap">{{ msg.content }}</div>
            <div
              :class="[
                'text-[10px] mt-1.5',
                msg.role === 'user' ? 'text-white/60' : 'text-slate-400 dark:text-slate-500'
              ]"
            >{{ formatTime(msg.created_at) }}</div>
          </div>

          <div v-if="sending" class="mr-auto max-w-[85%] rounded-2xl px-4 py-3 bg-slate-100 dark:bg-surface-850">
            <div class="flex items-center gap-2 text-sm text-slate-400 dark:text-slate-500">
              <div class="flex gap-1">
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 dark:bg-slate-600 animate-bounce" style="animation-delay: 0ms"></span>
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 dark:bg-slate-600 animate-bounce" style="animation-delay: 150ms"></span>
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 dark:bg-slate-600 animate-bounce" style="animation-delay: 300ms"></span>
              </div>
              Думаю…
            </div>
          </div>
        </div>

        <!-- Input -->
        <div class="p-4 border-t border-slate-200/80 dark:border-white/[0.06]">
          <form @submit.prevent="sendMessage" class="flex gap-2">
            <input
              v-model="inputText"
              type="text"
              placeholder="Введите сообщение…"
              class="input flex-1 text-sm"
              :disabled="sending"
              ref="inputRef"
            />
            <button
              type="submit"
              class="btn-primary text-sm px-4"
              :disabled="sending || !inputText.trim()"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
            </button>
          </form>
        </div>
      </template>

      <!-- Empty state -->
      <div v-else class="flex-1 flex items-center justify-center">
        <div class="text-center">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-primary-500/10 dark:bg-primary-500/5 flex items-center justify-center">
            <svg class="w-8 h-8 text-primary-500/60" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p class="text-sm text-slate-500 dark:text-slate-400">Выберите чат или создайте новый</p>
          <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">AI-ассистент по облигациям</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ChatSession, ChatMessage, ChatAgent } from '~/composables/useApi'

const api = useApi()

const agents = ref<ChatAgent[]>([])
const sessions = ref<ChatSession[]>([])
const messages = ref<ChatMessage[]>([])
const activeSessionId = ref<string | null>(null)
const inputText = ref('')
const sending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

const activeSession = computed(() => sessions.value.find(s => s.session_id === activeSessionId.value))

// Load agents and sessions on mount
onMounted(async () => {
  try {
    const [a, s] = await Promise.all([
      api.getChatAgents(),
      api.getChatSessions(),
    ])
    agents.value = a
    sessions.value = s
  } catch (e) {
    console.error('Failed to load chat data:', e)
  }
})

async function createSession(agentType: string, agentName: string) {
  try {
    const session = await api.createChatSession(agentType, `Чат: ${agentName}`)
    sessions.value.unshift(session)
    await selectSession(session.session_id)
  } catch (e: any) {
    console.error('Create session error:', e)
  }
}

async function selectSession(id: string) {
  activeSessionId.value = id
  messages.value = []
  try {
    const detail = await api.getChatSession(id)
    messages.value = detail.messages || []
    await nextTick()
    scrollToBottom()
    inputRef.value?.focus()
  } catch (e) {
    console.error('Load session error:', e)
  }
}

async function deleteSession(id: string) {
  try {
    await api.deleteChatSession(id)
    sessions.value = sessions.value.filter(s => s.session_id !== id)
    if (activeSessionId.value === id) {
      activeSessionId.value = null
      messages.value = []
    }
  } catch (e) {
    console.error('Delete session error:', e)
  }
}

async function sendMessage() {
  const content = inputText.value.trim()
  if (!content || !activeSessionId.value || sending.value) return

  inputText.value = ''
  sending.value = true

  // Optimistic user message
  messages.value.push({
    session_id: activeSessionId.value,
    role: 'user',
    content,
    created_at: new Date().toISOString(),
  })
  await nextTick()
  scrollToBottom()

  try {
    const resp = await api.sendChatMessage(activeSessionId.value, content)
    // Replace optimistic message and add assistant response
    messages.value[messages.value.length - 1] = resp.user_message
    messages.value.push(resp.assistant_message)

    // Update session in sidebar
    const idx = sessions.value.findIndex(s => s.session_id === activeSessionId.value)
    if (idx !== -1) {
      sessions.value[idx].updated_at = new Date().toISOString()
    }
  } catch (e: any) {
    messages.value.push({
      session_id: activeSessionId.value!,
      role: 'assistant',
      content: 'Ошибка: ' + (e.message || 'Не удалось получить ответ'),
      created_at: new Date().toISOString(),
    })
  } finally {
    sending.value = false
    await nextTick()
    scrollToBottom()
    inputRef.value?.focus()
  }
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

function agentName(type: string): string {
  const agent = agents.value.find(a => a.type === type)
  return agent?.name || type
}

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

useHead({ title: 'Чат — NLA' })
</script>
