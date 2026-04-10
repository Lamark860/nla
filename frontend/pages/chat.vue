<template>
  <div class="d-flex gap-3" style="height: calc(100vh - 180px)">
    <!-- Sidebar: Sessions -->
    <div class="card d-flex flex-column flex-shrink-0 overflow-hidden" style="width: 280px">
      <div class="card-header py-3">
        <h6 class="fw-semibold mb-2">Чаты</h6>
        <div class="d-flex flex-column gap-1">
          <button
            v-for="agent in agents"
            :key="agent.type"
            class="btn btn-outline-secondary btn-sm text-start d-flex align-items-center gap-2"
            @click="createSession(agent.type, agent.name)"
          >
            <i :class="agent.icon" class="bi"></i>
            <span>{{ agent.name }}</span>
          </button>
        </div>
      </div>

      <div class="flex-grow-1 overflow-auto">
        <div
          v-for="s in sessions"
          :key="s.session_id"
          :class="[
            'px-3 py-3 border-bottom cursor-pointer',
            activeSessionId === s.session_id ? 'bg-primary bg-opacity-10' : ''
          ]"
          style="cursor: pointer"
          @click="selectSession(s.session_id)"
        >
          <div class="d-flex align-items-start justify-content-between gap-2">
            <div class="text-truncate">
              <p class="small fw-medium mb-0 text-truncate">{{ s.title || 'Новый чат' }}</p>
              <p class="text-muted mb-0" style="font-size: 11px">
                {{ agentName(s.agent_type) }} · {{ formatTime(s.updated_at) }}
              </p>
            </div>
            <button
              class="btn btn-link btn-sm text-muted p-0 flex-shrink-0"
              @click.stop="deleteSession(s.session_id)"
              title="Удалить"
            >
              <i class="bi bi-trash" style="font-size: 12px"></i>
            </button>
          </div>
        </div>

        <div v-if="sessions.length === 0" class="p-4 text-center small text-muted">
          Нет активных чатов
        </div>
      </div>
    </div>

    <!-- Main: Chat -->
    <div class="flex-grow-1 card d-flex flex-column overflow-hidden">
      <template v-if="activeSessionId">
        <!-- Header -->
        <div class="card-header d-flex align-items-center justify-content-between py-2">
          <div>
            <h6 class="fw-semibold mb-0 small">{{ activeSession?.title }}</h6>
            <span class="text-muted" style="font-size: 11px">{{ agentName(activeSession?.agent_type || '') }}</span>
          </div>
        </div>

        <!-- Messages -->
        <div ref="messagesContainer" class="flex-grow-1 overflow-auto p-4 d-flex flex-column gap-3">
          <div
            v-for="(msg, i) in messages"
            :key="i"
            :class="[
              'rounded-3 px-3 py-2 small',
              msg.role === 'user'
                ? 'ms-auto chat-bubble-user'
                : 'me-auto chat-bubble-assistant'
            ]"
            style="max-width: 85%"
          >
            <div v-if="msg.role === 'assistant'" v-html="renderMarkdown(msg.content)" class="markdown-body"></div>
            <div v-else style="white-space: pre-wrap">{{ msg.content }}</div>
            <div
              :class="msg.role === 'user' ? 'opacity-50' : 'text-muted'"
              style="font-size: 10px; margin-top: 4px"
            >{{ formatTime(msg.created_at) }}</div>
          </div>

          <div v-if="sending" class="me-auto bg-body-secondary rounded-3 px-3 py-2" style="max-width: 85%">
            <div class="d-flex align-items-center gap-2 small text-muted">
              <div class="spinner-grow spinner-grow-sm" role="status"></div>
              Думаю…
            </div>
          </div>
        </div>

        <!-- Input -->
        <div class="card-footer py-3">
          <div class="d-flex gap-2">
            <textarea
              v-model="inputText"
              placeholder="Введите сообщение… (Shift+Enter — новая строка)"
              class="form-control form-control-sm"
              style="resize: none; min-height: 38px; max-height: 120px; overflow-y: auto"
              rows="1"
              :disabled="sending"
              ref="inputRef"
              @keydown.enter.exact.prevent="sendMessage"
              @input="autoResize"
            ></textarea>
            <button
              class="btn btn-primary btn-sm px-3 align-self-end"
              :disabled="sending || !inputText.trim()"
              @click="sendMessage"
            >
              <i class="bi bi-send"></i>
            </button>
          </div>
        </div>
      </template>

      <!-- Empty state -->
      <div v-else class="flex-grow-1 d-flex align-items-center justify-content-center">
        <div class="text-center">
          <div class="mx-auto mb-3 d-flex align-items-center justify-content-center rounded-3 bg-primary bg-opacity-10" style="width: 64px; height: 64px">
            <i class="bi bi-chat-dots text-primary" style="font-size: 28px"></i>
          </div>
          <p class="small text-muted mb-1">Выберите чат или создайте новый</p>
          <p class="text-muted" style="font-size: 12px">AI-ассистент по облигациям</p>
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
const inputRef = ref<HTMLTextAreaElement | null>(null)

const activeSession = computed(() => sessions.value.find(s => s.session_id === activeSessionId.value))

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
    messages.value[messages.value.length - 1] = resp.user_message
    messages.value.push(resp.assistant_message)

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

function autoResize(e: Event) {
  const el = e.target as HTMLTextAreaElement
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

function renderMarkdown(text: string): string {
  if (!text) return ''
  let html = text
  html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')
  html = html.replace(/^### (.+)$/gm, '<h6 class="fw-semibold mt-2 mb-1">$1</h6>')
  html = html.replace(/^## (.+)$/gm, '<h6 class="fw-bold mt-2 mb-1">$1</h6>')
  html = html.replace(/^# (.+)$/gm, '<h5 class="fw-bold mt-2 mb-1">$1</h5>')
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
  html = html.replace(/^> (.+)$/gm, '<blockquote class="border-start border-3 ps-3 text-muted mb-1">$1</blockquote>')
  html = html.replace(/^---$/gm, '<hr class="my-2">')
  html = html.replace(/^\d+\.\s+(.+)$/gm, '<li>$1</li>')
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul class="mb-1 ps-3">$&</ul>')
  html = html.replace(/((?:^\|.+\|$\n?)+)/gm, (block: string) => {
    const rows = block.trim().split('\n').filter(r => r.trim())
    if (rows.length < 2) return block
    const isSep = /^\|[\s:-]+\|$/.test(rows[1].trim())
    const parseRow = (r: string) => r.split('|').slice(1, -1).map(c => c.trim())
    let table = '<table class="table table-sm table-bordered mb-1" style="color: var(--nla-text)">'
    if (isSep && rows.length > 2) {
      table += '<thead><tr>' + parseRow(rows[0]).map(c => `<th>${c}</th>`).join('') + '</tr></thead>'
      table += '<tbody>' + rows.slice(2).map(r => '<tr>' + parseRow(r).map(c => `<td>${c}</td>`).join('') + '</tr>').join('') + '</tbody>'
    } else {
      table += '<tbody>' + rows.map(r => '<tr>' + parseRow(r).map(c => `<td>${c}</td>`).join('') + '</tr>').join('') + '</tbody>'
    }
    return table + '</table>'
  })
  html = html.replace(/\n\n/g, '</p><p>')
  html = '<p>' + html + '</p>'
  html = html.replace(/<p>\s*<\/p>/g, '')
  return html
}

useHead({ title: 'Чат — NLA' })
</script>

<style scoped>
.chat-bubble-user {
  background: var(--nla-primary);
  color: #fff;
}
.chat-bubble-assistant {
  background: var(--nla-bg-card);
  border: 1px solid var(--nla-border);
  color: var(--nla-text);
}
.markdown-body h5, .markdown-body h6 { font-size: 0.85rem; color: var(--nla-text); }
.markdown-body p { margin-bottom: 0.4rem; color: var(--nla-text); }
.markdown-body ul { margin-bottom: 0.3rem; }
.markdown-body li { color: var(--nla-text); }
.markdown-body pre { background: var(--nla-bg-code); padding: 0.5rem; border-radius: 6px; font-size: 0.8rem; }
.markdown-body code { font-size: 0.8rem; }
.markdown-body blockquote { font-size: 0.85rem; }
.markdown-body hr { border-color: var(--nla-border); }
.markdown-body table { font-size: 0.8rem; }
.markdown-body strong { color: var(--nla-text); }
</style>
