<template>
  <div class="d-flex flex-column" style="height: calc(100vh - var(--v-layout-top, 0px))">

    <!-- ── Header ──────────────────────────────────────── -->
    <div class="d-flex align-center gap-1 px-4" style="height: 64px; flex-shrink: 0">
      <v-avatar color="primary" size="40" class="mr-1">
        <v-icon>mdi-robot-happy</v-icon>
      </v-avatar>

      <div class="flex-grow-1 overflow-hidden">
        <div class="text-subtitle-2 font-weight-bold text-truncate" style="line-height:1.2">
          {{ store.agentName ?? 'AI Agent' }}
        </div>
        <div class="text-caption text-medium-emphasis text-truncate" style="font-family:monospace">
          {{ store.sessionId }}
        </div>
      </div>

      <v-btn icon size="small" variant="text" :loading="loadingHistory" @click="loadHistory">
        <v-icon>mdi-history</v-icon>
        <v-tooltip activator="parent" location="bottom">Carregar histórico</v-tooltip>
      </v-btn>
      <v-btn icon size="small" variant="text" @click="openSessionsDialog">
        <v-icon>mdi-folder-open-outline</v-icon>
        <v-tooltip activator="parent" location="bottom">Sessões</v-tooltip>
      </v-btn>
      <v-btn icon size="small" variant="text" color="primary" @click="newSessionDialog = true">
        <v-icon>mdi-plus-circle-outline</v-icon>
        <v-tooltip activator="parent" location="bottom">Nova sessão</v-tooltip>
      </v-btn>
      <v-btn
        v-if="store.messages.length > 0"
        icon size="small" variant="text" color="error"
        @click="clearDialog = true"
      >
        <v-icon>mdi-delete-outline</v-icon>
        <v-tooltip activator="parent" location="bottom">Limpar histórico</v-tooltip>
      </v-btn>
    </div>

    <v-divider />

    <!-- ── Messages ─────────────────────────────────────── -->
    <div ref="messagesEl" class="flex-grow-1 overflow-y-auto messages-area">

      <!-- Empty state -->
      <div
        v-if="store.messages.length === 0 && !store.loading"
        class="d-flex flex-column align-center justify-center text-center h-100"
      >
        <v-avatar color="primary" size="80" class="mb-5" style="opacity:.15">
          <v-icon size="48">mdi-robot-happy</v-icon>
        </v-avatar>
        <p class="text-h6 mb-1 font-weight-regular">
          {{ store.agentName ? `Olá! Sou ${store.agentName}` : 'Olá!' }}
        </p>
        <p class="text-body-2 text-medium-emphasis">Como posso ajudar você hoje?</p>
      </div>

      <template v-for="(msg, i) in store.messages" :key="i">

        <!-- System notification -->
        <div v-if="msg.role === 'system'" class="d-flex justify-center my-3">
          <v-chip size="x-small" variant="tonal" prepend-icon="mdi-paperclip">
            {{ msg.text }}
          </v-chip>
        </div>

        <!-- Chat bubbles -->
        <div
          v-else
          class="d-flex mb-2"
          :class="msg.role === 'user' ? 'justify-end' : 'align-end gap-2'"
        >
          <v-avatar v-if="msg.role !== 'user'" color="primary" size="28">
            <v-icon size="14">mdi-robot</v-icon>
          </v-avatar>

          <!-- User -->
          <v-sheet
            v-if="msg.role === 'user'"
            color="primary"
            rounded="xl"
            class="px-4 py-3 bubble-user"
          >
            <p class="text-body-1 text-on-primary mb-0" style="white-space:pre-wrap;line-height:1.55">{{ msg.text }}</p>
          </v-sheet>

          <!-- Model -->
          <v-sheet
            v-else
            rounded="xl"
            class="border px-4 py-3 bubble-model"
          >
            <div class="markdown-body text-body-1" v-html="renderMarkdown(msg.text)" />
          </v-sheet>
        </div>

      </template>

      <!-- Typing indicator -->
      <div v-if="store.loading" class="d-flex align-end gap-2 mb-2">
        <v-avatar color="primary" size="28">
          <v-icon size="14">mdi-robot</v-icon>
        </v-avatar>
        <v-sheet rounded="xl" class="border px-4 py-3" style="border-bottom-left-radius:4px">
          <div class="typing-dots"><span /><span /><span /></div>
        </v-sheet>
      </div>

    </div>

    <!-- Error -->
    <v-alert
      v-if="store.error"
      type="error" variant="tonal" density="compact"
      class="mx-4 mb-2" closable
      @click:close="store.error = null"
    >{{ store.error }}</v-alert>

    <v-divider />

    <!-- ── Input ──────────────────────────────────────── -->
    <div class="pa-4" style="flex-shrink:0">

      <!-- Upload status -->
      <v-slide-y-reverse-transition>
        <div v-if="uploadStatus" class="mb-2">
          <v-chip
            size="small" variant="tonal"
            :color="uploadStatus.state === 'success' ? 'success' : uploadStatus.state === 'error' ? 'error' : undefined"
            :prepend-icon="uploadStatus.state === 'success' ? 'mdi-check-circle' : uploadStatus.state === 'error' ? 'mdi-alert-circle' : undefined"
          >
            <template v-if="uploadStatus.state === 'uploading'" #prepend>
              <v-progress-circular size="13" width="2" indeterminate class="mr-1" />
            </template>
            {{ uploadStatus.name }}
          </v-chip>
        </div>
      </v-slide-y-reverse-transition>

      <!-- Input card -->
      <input ref="fileInput" type="file" accept=".txt,.pdf" style="display:none" @change="onFileSelected" />
      <v-card
        variant="outlined"
        rounded="xl"
        :color="inputFocused ? 'primary' : undefined"
        flat
      >
        <v-textarea
          v-model="input"
          placeholder="Digite sua mensagem…"
          variant="plain"
          rows="1"
          auto-grow
          max-rows="6"
          hide-details
          density="compact"
          class="px-2 pt-2 pb-0"
          @keydown.enter.exact.prevent="sendMessage"
          @keydown.shift.enter.exact="input += '\n'"
          @focus="inputFocused = true"
          @blur="inputFocused = false"
        />

        <div class="d-flex align-center px-2 pb-2 pt-1">
          <v-btn
            icon size="small" variant="text"
            :disabled="uploading"
            @click="triggerFileInput"
          >
            <v-icon size="20">mdi-paperclip</v-icon>
            <v-tooltip activator="parent" location="top">Anexar arquivo (.txt, .pdf)</v-tooltip>
          </v-btn>
          <v-spacer />
          <span class="text-caption text-disabled mr-3 d-none d-sm-inline">Shift+Enter = nova linha</span>
          <v-btn
            icon color="primary" size="small" variant="flat" rounded="lg"
            :loading="store.loading"
            :disabled="!input.trim()"
            @click="sendMessage"
          >
            <v-icon size="18">mdi-send</v-icon>
          </v-btn>
        </div>
      </v-card>
    </div>

  </div>

  <!-- ── Sessions dialog ─────────────────────────────── -->
  <v-dialog v-model="sessionDialog" max-width="560" scrollable>
    <v-card rounded="lg">
      <v-card-title class="pt-4 d-flex align-center">
        Sessões
        <v-spacer />
        <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus"
          @click="sessionDialog = false; newSessionDialog = true">
          Nova sessão
        </v-btn>
      </v-card-title>
      <v-divider />
      <v-card-text class="pa-0" style="max-height:420px">
        <v-progress-linear v-if="loadingSessions" indeterminate />
        <v-list v-if="sessions.length > 0" lines="two">
          <v-list-item
            v-for="s in sessions" :key="s.session_id"
            :active="s.session_id === store.sessionId"
            color="primary" rounded="lg" class="mx-2 my-1"
            @click="selectSession(s)"
          >
            <template #title>
              <span class="text-body-2 font-weight-medium" style="font-family:monospace">{{ s.session_id }}</span>
            </template>
            <template #subtitle>
              {{ getAgentName(s.agent_config_id) }}
              <span class="mx-1 text-disabled">·</span>{{ s.message_count }} msg
              <template v-if="s.updated_at">
                <span class="mx-1 text-disabled">·</span>{{ formatDate(s.updated_at) }}
              </template>
            </template>
            <template #append>
              <v-icon v-if="s.session_id === store.sessionId" size="16" color="primary">mdi-check-circle</v-icon>
            </template>
          </v-list-item>
        </v-list>
        <div v-else-if="!loadingSessions" class="text-center text-medium-emphasis py-10">
          <v-icon size="40" style="opacity:.25">mdi-chat-outline</v-icon>
          <p class="mt-2 text-body-2">Nenhuma sessão anterior.</p>
        </div>
      </v-card-text>
      <v-divider />
      <v-card-text class="pt-3 pb-2">
        <v-row no-gutters align="center" class="gap-2">
          <v-col>
            <v-text-field v-model="manualSessionId" label="Ou digitar um ID de sessão"
              variant="outlined" density="compact" hide-details @keydown.enter="applyManualSession" />
          </v-col>
          <v-col cols="auto">
            <v-btn color="primary" variant="tonal" :disabled="!manualSessionId.trim()" @click="applyManualSession">Ir</v-btn>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="pt-0">
        <v-spacer />
        <v-btn @click="sessionDialog = false">Fechar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ── New session dialog ──────────────────────────── -->
  <v-dialog v-model="newSessionDialog" max-width="440">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Nova sessão</v-card-title>
      <v-card-text>
        <v-select v-model="selectedConfigId" :items="agentConfigsStore.configs"
          item-title="name" item-value="id" label="Selecione o agente"
          variant="outlined" density="comfortable" hide-details />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="newSessionDialog = false">Cancelar</v-btn>
        <v-btn color="primary" :disabled="!selectedConfigId" @click="startNewSession">Iniciar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ── Clear confirm dialog ────────────────────────── -->
  <v-dialog v-model="clearDialog" max-width="380">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Limpar histórico</v-card-title>
      <v-card-text>Apagará todo o histórico desta sessão no servidor. Tem certeza?</v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="clearDialog = false">Cancelar</v-btn>
        <v-btn color="error" @click="doClear">Limpar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useAgentConfigsStore } from '@/stores/agent_configs'
import { chatAPI, filesAPI } from '@/services/api'
import { renderMarkdown } from '@/utils/markdown'

const store = useChatStore()
const agentConfigsStore = useAgentConfigsStore()

const input        = ref('')
const inputFocused = ref(false)
const messagesEl   = ref(null)
const fileInput    = ref(null)
const loadingHistory = ref(false)
const uploading    = ref(false)
const uploadStatus = ref(null)

const clearDialog      = ref(false)
const sessionDialog    = ref(false)
const newSessionDialog = ref(false)
const manualSessionId  = ref('')
const selectedConfigId = ref(null)
const sessions         = ref([])
const loadingSessions  = ref(false)

onMounted(async () => {
  await agentConfigsStore.fetchAll()
  if (!store.agentConfigId && agentConfigsStore.configs.length > 0) {
    store.agentConfigId = agentConfigsStore.configs[0].id
    store.agentName     = agentConfigsStore.configs[0].name
  }
  selectedConfigId.value = store.agentConfigId
  scrollToBottom()
})

watch(() => store.messages.length, scrollToBottom)

// ── Messaging ──────────────────────────────────────────
async function sendMessage() {
  const text = input.value.trim()
  if (!text || store.loading) return
  input.value = ''
  await store.send(text)
}

async function loadHistory() {
  loadingHistory.value = true
  await store.loadHistory()
  scrollToBottom()
  loadingHistory.value = false
}

async function doClear() {
  await store.clearHistory()
  clearDialog.value = false
}

// ── File upload ────────────────────────────────────────
function triggerFileInput() { fileInput.value?.click() }

async function onFileSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  e.target.value = ''
  uploading.value  = true
  uploadStatus.value = { state: 'uploading', name: file.name }
  try {
    await filesAPI.upload(file)
    uploadStatus.value = { state: 'success', name: file.name }
    store.messages.push({ role: 'system', text: `${file.name} enviado com sucesso` })
    setTimeout(() => { uploadStatus.value = null }, 3000)
  } catch {
    uploadStatus.value = { state: 'error', name: `Erro ao enviar ${file.name}` }
    setTimeout(() => { uploadStatus.value = null }, 4000)
  } finally {
    uploading.value = false
  }
}

// ── Sessions ───────────────────────────────────────────
async function openSessionsDialog() {
  sessionDialog.value = true
  manualSessionId.value = ''
  loadingSessions.value = true
  try {
    const { data } = await chatAPI.listSessions()
    sessions.value = data ?? []
  } catch { sessions.value = [] }
  finally { loadingSessions.value = false }
}

function getAgentName(id) {
  if (!id) return '—'
  return agentConfigsStore.configs.find((c) => c.id === id)?.name ?? id
}

function formatDate(dateStr) {
  const d = new Date(dateStr)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return ''
  return d.toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
}

function selectSession(s) {
  store.setSession(s.session_id)
  store.agentConfigId = s.agent_config_id || null
  store.agentName     = getAgentName(s.agent_config_id)
  sessionDialog.value = false
  loadHistory()
}

function applyManualSession() {
  const id = manualSessionId.value.trim()
  if (!id) return
  store.setSession(id)
  store.agentConfigId = null
  store.agentName     = null
  sessionDialog.value = false
}

function startNewSession() {
  const cfg = agentConfigsStore.configs.find((c) => c.id === selectedConfigId.value)
  store.newSession(selectedConfigId.value)
  if (cfg) store.agentName = cfg.name
  newSessionDialog.value = false
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}
</script>

<style scoped>
.messages-area {
  padding: 24px 20px 12px;
  scroll-behavior: smooth;
}

/* Bubbles */
.bubble-user  { border-bottom-right-radius: 4px !important; max-width: 74%; }
.bubble-model { border-bottom-left-radius:  4px !important; max-width: 74%; }

/* Typing animation */
.typing-dots { display: flex; gap: 5px; align-items: center; height: 18px; }
.typing-dots span {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: currentColor;
  opacity: .5;
  animation: bounce 1.4s infinite;
}
.typing-dots span:nth-child(2) { animation-delay: .2s; }
.typing-dots span:nth-child(3) { animation-delay: .4s; }
@keyframes bounce {
  0%, 80%, 100% { transform: scale(.8); opacity: .4; }
  40%           { transform: scale(1.2); opacity: 1;  }
}

/* Markdown inside model bubble */
:deep(.markdown-body) {
  line-height: 1.6;
  > *:first-child { margin-top: 0; }
  > *:last-child  { margin-bottom: 0; }
  p  { margin: .45em 0; }
  h1,h2,h3,h4,h5,h6 { margin: .8em 0 .3em; font-weight: 600; line-height: 1.3; }
  h1 { font-size: 1.35em; } h2 { font-size: 1.15em; } h3 { font-size: 1.02em; }
  ul, ol { margin: .4em 0; padding-left: 1.5em; }
  li { margin: .15em 0; }
  code {
    font-family: 'Roboto Mono', monospace;
    font-size: .875em;
    padding: .15em .4em;
    border-radius: 4px;
    background: rgba(var(--v-theme-on-surface), 0.08);
  }
  pre {
    margin: .6em 0; padding: .75em 1em;
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    overflow-x: auto;
    code { padding: 0; background: none; font-size: .85em; }
  }
  blockquote {
    margin: .5em 0; padding: .3em .75em;
    border-left: 3px solid currentColor; opacity: .75;
  }
  table {
    border-collapse: collapse; margin: .5em 0; width: 100%; font-size: .9em;
    th, td {
      border: 1px solid rgba(var(--v-theme-on-surface), 0.15);
      padding: .3em .6em;
    }
    th { font-weight: 600; background: rgba(var(--v-theme-on-surface), 0.06); }
  }
  hr { border: none; border-top: 1px solid rgba(var(--v-theme-on-surface), 0.15); margin: .8em 0; }
  a  { color: inherit; text-decoration: underline; opacity: .8; }
  a:hover { opacity: 1; }
}
</style>
