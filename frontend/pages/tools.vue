<template>
  <div class="tools-page" style="margin: 0 -12px">
    <h1 class="h4 fw-bold mb-4">Инструменты</h1>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-4">
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'markdown' }" @click="activeTab = 'markdown'">
          <i class="bi bi-markdown me-1"></i> Markdown
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'json' }" @click="activeTab = 'json'">
          <i class="bi bi-braces me-1"></i> JSON Decode
        </button>
      </li>
    </ul>

    <!-- Markdown Tab -->
    <div v-if="activeTab === 'markdown'">
      <div class="row g-2">
        <div class="col-lg-3">
          <div class="card h-100">
            <div class="card-header d-flex align-items-center justify-content-between py-2">
              <span class="small fw-semibold">Markdown</span>
              <div class="d-flex gap-2">
                <button class="btn btn-outline-secondary btn-sm" @click="mdInput = ''" title="Очистить">
                  <i class="bi bi-trash"></i>
                </button>
                <button class="btn btn-outline-secondary btn-sm" @click="copyToClipboard(mdRendered)" title="Копировать HTML">
                  <i class="bi bi-clipboard"></i>
                </button>
              </div>
            </div>
            <div class="card-body p-0">
              <textarea
                v-model="mdInput"
                class="form-control border-0 rounded-0"
                style="min-height: 400px; resize: vertical; font-family: monospace; font-size: 13px"
                placeholder="Введите Markdown..."
              ></textarea>
            </div>
          </div>
        </div>
        <div class="col-lg-9">
          <div class="card h-100">
            <div class="card-header py-2">
              <span class="small fw-semibold">Превью</span>
            </div>
            <div class="card-body" style="min-height: 400px; overflow: auto">
              <div v-html="mdRendered" class="markdown-body"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- JSON Decode Tab -->
    <div v-if="activeTab === 'json'">
      <div class="row g-2">
        <div class="col-lg-3">
          <div class="card h-100">
            <div class="card-header d-flex align-items-center justify-content-between py-2">
              <span class="small fw-semibold">Encoded JSON</span>
              <div class="d-flex gap-2">
                <button class="btn btn-outline-secondary btn-sm" @click="jsonInput = ''" title="Очистить">
                  <i class="bi bi-trash"></i>
                </button>
                <button class="btn btn-primary btn-sm" @click="decodeJson" title="Декодировать">
                  Decode
                </button>
              </div>
            </div>
            <div class="card-body p-0">
              <textarea
                v-model="jsonInput"
                class="form-control border-0 rounded-0"
                style="min-height: 400px; resize: vertical; font-family: monospace; font-size: 13px"
                placeholder='Вставьте escaped/encoded JSON...'
              ></textarea>
            </div>
          </div>
        </div>
        <div class="col-lg-9">
          <div class="card h-100">
            <div class="card-header d-flex align-items-center justify-content-between py-2">
              <span class="small fw-semibold">Результат</span>
              <button v-if="jsonOutput" class="btn btn-outline-secondary btn-sm" @click="copyToClipboard(jsonOutput)" title="Копировать">
                <i class="bi bi-clipboard"></i>
              </button>
            </div>
            <div class="card-body p-0">
              <pre v-if="jsonOutput" class="m-0 p-3" style="min-height: 400px; overflow: auto; font-size: 13px; background: transparent"><code>{{ jsonOutput }}</code></pre>
              <div v-else-if="jsonError" class="p-3 text-danger small">{{ jsonError }}</div>
              <div v-else class="p-3 text-muted small">Результат декодирования появится здесь</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const activeTab = ref<'markdown' | 'json'>('markdown')

// --- Markdown ---
const mdInput = ref('# Заголовок\n\nОбычный текст с **жирным** и *курсивом*.\n\n- Элемент списка 1\n- Элемент списка 2\n\n| Облигация | Доходность | Рейтинг |\n| --- | --- | --- |\n| ОФЗ 26238 | 14.02% | AAA |\n| Газпром БО-001 | 12.5% | AA+ |\n\n```\ncode block\n```')

const mdRendered = computed(() => {
  return renderMarkdown(mdInput.value)
})

function renderMarkdown(text: string): string {
  if (!text) return ''
  let html = text
  // Code blocks
  html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')
  // Headers
  html = html.replace(/^### (.+)$/gm, '<h5>$1</h5>')
  html = html.replace(/^## (.+)$/gm, '<h4>$1</h4>')
  html = html.replace(/^# (.+)$/gm, '<h3>$1</h3>')
  // Bold & Italic
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')
  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
  // Images
  html = html.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" style="max-width:100%">')
  // Blockquotes
  html = html.replace(/^> (.+)$/gm, '<blockquote class="border-start border-3 ps-3 text-muted">$1</blockquote>')
  // Lists
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
  // Tables (| col | col |)
  html = html.replace(/((?:^\|.+\|$\n?)+)/gm, (block: string) => {
    const rows = block.trim().split('\n').filter(r => r.trim())
    if (rows.length < 2) return block
    // Check if 2nd row is separator (| --- | --- |)
    const isSep = /^\|[\s:-]+\|$/.test(rows[1].trim())
    const dataRows = isSep ? [rows[0], ...rows.slice(2)] : rows
    const headerRow = isSep ? rows[0] : null
    const parseRow = (r: string) => r.split('|').slice(1, -1).map(c => c.trim())
    let table = '<table class="table table-sm table-bordered">'
    if (headerRow) {
      table += '<thead><tr>' + parseRow(headerRow).map(c => `<th>${c}</th>`).join('') + '</tr></thead>'
      table += '<tbody>' + dataRows.slice(1).map(r => '<tr>' + parseRow(r).map(c => `<td>${c}</td>`).join('') + '</tr>').join('') + '</tbody>'
    } else {
      table += '<tbody>' + dataRows.map(r => '<tr>' + parseRow(r).map(c => `<td>${c}</td>`).join('') + '</tr>').join('') + '</tbody>'
    }
    return table + '</table>'
  })
  // Paragraphs (double newline)
  html = html.replace(/\n\n/g, '</p><p>')
  html = '<p>' + html + '</p>'
  // Clean empty paragraphs
  html = html.replace(/<p>\s*<\/p>/g, '')
  return html
}

// --- JSON Decode ---
const jsonInput = ref('')
const jsonOutput = ref('')
const jsonError = ref('')

function decodeJson() {
  jsonOutput.value = ''
  jsonError.value = ''
  let text = jsonInput.value.trim()
  if (!text) return

  // Try multiple decode strategies
  const strategies = [
    // 1. Direct JSON.parse
    () => JSON.parse(text),
    // 2. Unicode unescape (\uXXXX)
    () => JSON.parse(text.replace(/\\u([\dA-Fa-f]{4})/g, (_, g) => String.fromCharCode(parseInt(g, 16)))),
    // 3. URL decode
    () => JSON.parse(decodeURIComponent(text)),
    // 4. Base64 decode
    () => JSON.parse(atob(text)),
    // 5. Double-escaped (string within string)
    () => JSON.parse(JSON.parse(text)),
    // 6. Strip surrounding quotes and try
    () => {
      if ((text.startsWith('"') && text.endsWith('"')) || (text.startsWith("'") && text.endsWith("'"))) {
        return JSON.parse(text.slice(1, -1).replace(/\\"/g, '"').replace(/\\\\/g, '\\'))
      }
      throw new Error('not quoted')
    },
  ]

  for (const strategy of strategies) {
    try {
      const result = strategy()
      jsonOutput.value = JSON.stringify(result, null, 2)
      return
    } catch {}
  }

  jsonError.value = 'Не удалось декодировать JSON. Проверьте формат.'
}

// --- Utils ---
async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch { /* ignore */ }
}

useHead({ title: 'Инструменты — NLA' })
</script>
