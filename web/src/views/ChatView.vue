<template>
  <div class="chat-root">
    <!-- Session bar -->
    <div class="chat-header px-4 py-2 d-flex align-center flex-wrap gap-1">
      <v-chip
        color="primary"
        variant="outlined"
        label
        size="small"
        prepend-icon="mdi-identifier"
        class="cursor-pointer"
        @click="openSessionsDialog"
      >
        {{ store.sessionId }}
      </v-chip>

      <v-chip
        v-if="store.agentName"
        color="secondary"
        variant="tonal"
        label
        size="small"
        prepend-icon="mdi-robot-happy"
      >
        {{ store.agentName }}
      </v-chip>

      <v-btn
        size="small"
        variant="text"
        prepend-icon="mdi-history"
        :loading="loadingHistory"
        @click="loadHistory"
      >
        Carregar histórico
      </v-btn>

      <v-btn
        size="small"
        variant="text"
        prepend-icon="mdi-plus-circle-outline"
        @click="newSessionDialog = true"
      >
        Nova sessão
      </v-btn>

      <v-spacer />

      <v-btn
        v-if="store.messages.length > 0"
        size="small"
        variant="text"
        color="error"
        prepend-icon="mdi-delete-outline"
        @click="clearDialog = true"
      >
        Limpar
      </v-btn>
    </div>

    <v-divider />

    <!-- Messages -->
    <div ref="messagesEl" class="chat-messages">
      <!-- Empty state -->
      <div
        v-if="store.messages.length === 0 && !store.loading"
        class="d-flex flex-column align-center justify-center h-100 text-medium-emphasis"
      >
        <v-icon size="72" color="primary" class="mb-4" style="opacity: 0.3">mdi-robot</v-icon>
        <p class="text-h6 mb-1">Olá! Como posso ajudar?</p>
        <p class="text-body-2">Envie uma mensagem para começar.</p>
      </div>

      <!-- Message bubbles -->
      <div
        v-for="(msg, i) in store.messages"
        :key="i"
        class="mb-3"
        :class="msg.role === 'user' ? 'd-flex justify-end' : 'd-flex justify-start align-start'"
      >
        <!-- Agent avatar -->
        <v-avatar
          v-if="msg.role !== 'user'"
          color="primary"
          size="32"
          class="mr-2 mt-1 flex-shrink-0"
        >
          <v-icon size="18">mdi-robot</v-icon>
        </v-avatar>

        <!-- User bubble -->
        <v-card
          v-if="msg.role === 'user'"
          color="primary"
          rounded="xl"
          class="px-4 py-2"
          max-width="75%"
        >
          <p class="text-body-1 text-white mb-0" style="white-space: pre-wrap">{{ msg.text }}</p>
        </v-card>

        <!-- Model bubble -->
        <v-card
          v-else
          variant="tonal"
          rounded="xl"
          class="px-4 py-2"
          max-width="75%"
        >
          <div class="markdown-body text-body-1" v-html="renderMarkdown(msg.text)" />
        </v-card>
      </div>

      <!-- Typing indicator -->
      <div v-if="store.loading" class="d-flex justify-start align-start mb-3">
        <v-avatar color="primary" size="32" class="mr-2 mt-1 flex-shrink-0">
          <v-icon size="18">mdi-robot</v-icon>
        </v-avatar>
        <v-card variant="tonal" rounded="xl" class="px-4 py-3">
          <div class="typing-indicator">
            <span /><span /><span />
          </div>
        </v-card>
      </div>
    </div>

    <!-- Error -->
    <v-alert
      v-if="store.error"
      type="error"
      variant="tonal"
      density="compact"
      class="mx-4 mb-2"
      closable
      @click:close="store.error = null"
    >
      {{ store.error }}
    </v-alert>

    <v-divider />

    <!-- Input -->
    <div class="chat-input px-4 py-3">
      <v-row no-gutters align="end">
        <v-col>
          <v-textarea
            v-model="input"
            placeholder="Digite sua mensagem… (Enter para enviar)"
            variant="outlined"
            rows="1"
            auto-grow
            max-rows="6"
            hide-details
            density="comfortable"
            rounded="lg"
            @keydown.enter.exact.prevent="sendMessage"
            @keydown.shift.enter.exact="input += '\n'"
          />
        </v-col>
        <v-col cols="auto" class="pl-2">
          <v-btn
            icon="mdi-send"
            color="primary"
            size="large"
            :loading="store.loading"
            :disabled="!input.trim()"
            @click="sendMessage"
          />
        </v-col>
      </v-row>
      <p class="text-caption text-medium-emphasis mt-1 mb-0 pl-1">
        Enter para enviar · Shift+Enter para nova linha
      </p>
    </div>
  </div>

  <!-- Sessions list dialog -->
  <v-dialog v-model="sessionDialog" max-width="560" scrollable>
    <v-card rounded="lg">
      <v-card-title class="pt-4 d-flex align-center gap-2">
        Sessões
        <v-spacer />
        <v-btn
          size="small"
          variant="tonal"
          color="primary"
          prepend-icon="mdi-plus"
          @click="sessionDialog = false; newSessionDialog = true"
        >
          Nova sessão
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-0" style="max-height: 420px">
        <v-progress-linear v-if="loadingSessions" indeterminate />

        <v-list v-if="sessions.length > 0" lines="two">
          <v-list-item
            v-for="s in sessions"
            :key="s.session_id"
            :active="s.session_id === store.sessionId"
            color="primary"
            rounded="lg"
            class="mx-2 my-1"
            @click="selectSession(s)"
          >
            <template #title>
              <span class="text-body-2 font-weight-medium font-monospace">{{ s.session_id }}</span>
            </template>
            <template #subtitle>
              <span>{{ getAgentName(s.agent_config_id) }}</span>
              <span class="mx-1 text-disabled">·</span>
              <span>{{ s.message_count }} msg</span>
              <template v-if="s.updated_at">
                <span class="mx-1 text-disabled">·</span>
                <span>{{ formatDate(s.updated_at) }}</span>
              </template>
            </template>
            <template #append>
              <v-icon v-if="s.session_id === store.sessionId" size="16" color="primary">
                mdi-check-circle
              </v-icon>
            </template>
          </v-list-item>
        </v-list>

        <div
          v-else-if="!loadingSessions"
          class="text-center text-medium-emphasis py-8"
        >
          <v-icon size="40" style="opacity: 0.3">mdi-chat-outline</v-icon>
          <p class="mt-2 text-body-2">Nenhuma sessão anterior.</p>
        </div>
      </v-card-text>

      <v-divider />

      <v-card-text class="pt-3 pb-3">
        <v-row no-gutters align="center" class="gap-2">
          <v-col>
            <v-text-field
              v-model="manualSessionId"
              label="Ou digitar um ID de sessão"
              variant="outlined"
              density="compact"
              hide-details
              @keydown.enter="applyManualSession"
            />
          </v-col>
          <v-col cols="auto">
            <v-btn
              color="primary"
              variant="tonal"
              :disabled="!manualSessionId.trim()"
              @click="applyManualSession"
            >
              Ir
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>

      <v-card-actions class="pt-0">
        <v-spacer />
        <v-btn @click="sessionDialog = false">Fechar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- New session dialog (agent selector) -->
  <v-dialog v-model="newSessionDialog" max-width="440">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Nova sessão</v-card-title>
      <v-card-text>
        <v-select
          v-model="selectedConfigId"
          :items="agentConfigsStore.configs"
          item-title="name"
          item-value="id"
          label="Selecione o agente"
          variant="outlined"
          density="comfortable"
          hide-details
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="newSessionDialog = false">Cancelar</v-btn>
        <v-btn color="primary" :disabled="!selectedConfigId" @click="startNewSession">
          Iniciar
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Clear confirm dialog -->
  <v-dialog v-model="clearDialog" max-width="380">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Limpar histórico</v-card-title>
      <v-card-text>
        Apagará todo o histórico da sessão atual no servidor. Tem certeza?
      </v-card-text>
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
import { chatAPI } from '@/services/api'
import { renderMarkdown } from '@/utils/markdown'

const store = useChatStore()
const agentConfigsStore = useAgentConfigsStore()

const input = ref('')
const messagesEl = ref(null)
const loadingHistory = ref(false)
const clearDialog = ref(false)
const sessionDialog = ref(false)
const newSessionDialog = ref(false)
const manualSessionId = ref('')
const selectedConfigId = ref(null)
const sessions = ref([])
const loadingSessions = ref(false)

onMounted(async () => {
  await agentConfigsStore.fetchAll()
  if (!store.agentConfigId && agentConfigsStore.configs.length > 0) {
    store.agentConfigId = agentConfigsStore.configs[0].id
    store.agentName = agentConfigsStore.configs[0].name
  }
  selectedConfigId.value = store.agentConfigId
  scrollToBottom()
})

watch(
  () => store.messages.length,
  () => scrollToBottom(),
)

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

async function openSessionsDialog() {
  sessionDialog.value = true
  manualSessionId.value = ''
  loadingSessions.value = true
  try {
    const { data } = await chatAPI.listSessions()
    sessions.value = data ?? []
  } catch {
    sessions.value = []
  } finally {
    loadingSessions.value = false
  }
}

function getAgentName(agentConfigId) {
  if (!agentConfigId) return '—'
  const cfg = agentConfigsStore.configs.find((c) => c.id === agentConfigId)
  return cfg?.name ?? agentConfigId
}

function formatDate(dateStr) {
  const d = new Date(dateStr)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return ''
  return d.toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
}

function selectSession(s) {
  store.setSession(s.session_id)
  store.agentConfigId = s.agent_config_id || null
  store.agentName = getAgentName(s.agent_config_id)
  sessionDialog.value = false
  loadHistory()
}

function applyManualSession() {
  const id = manualSessionId.value.trim()
  if (!id) return
  store.setSession(id)
  store.agentConfigId = null
  store.agentName = null
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
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}
</script>

<style scoped>
.chat-root {
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--v-layout-top, 0px));
}

.chat-header {
  flex-shrink: 0;
  min-height: 52px;
}

.chat-messages {
  flex: 1 1 0;
  min-height: 0;
  overflow-y: auto;
  padding: 16px 16px 8px;
}

.chat-input {
  flex-shrink: 0;
}

.font-monospace {
  font-family: monospace;
}

:deep(.markdown-body) {
  line-height: 1.6;

  > *:first-child { margin-top: 0; }
  > *:last-child  { margin-bottom: 0; }

  p { margin: 0.5em 0; }

  h1, h2, h3, h4, h5, h6 {
    margin: 0.8em 0 0.4em;
    font-weight: 600;
    line-height: 1.3;
  }
  h1 { font-size: 1.4em; }
  h2 { font-size: 1.2em; }
  h3 { font-size: 1.05em; }

  ul, ol {
    margin: 0.4em 0;
    padding-left: 1.5em;
  }
  li { margin: 0.2em 0; }

  code {
    font-family: 'Roboto Mono', monospace;
    font-size: 0.875em;
    padding: 0.15em 0.4em;
    border-radius: 4px;
    background: rgba(128, 128, 128, 0.15);
  }

  pre {
    margin: 0.6em 0;
    padding: 0.75em 1em;
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.06);
    overflow-x: auto;

    code {
      padding: 0;
      background: none;
      font-size: 0.85em;
    }
  }

  blockquote {
    margin: 0.5em 0;
    padding: 0.3em 0.8em;
    border-left: 3px solid currentColor;
    opacity: 0.75;
  }

  table {
    border-collapse: collapse;
    margin: 0.5em 0;
    width: 100%;
    font-size: 0.9em;

    th, td {
      border: 1px solid rgba(128, 128, 128, 0.3);
      padding: 0.35em 0.6em;
      text-align: left;
    }
    th { font-weight: 600; background: rgba(128, 128, 128, 0.1); }
  }

  hr {
    border: none;
    border-top: 1px solid rgba(128, 128, 128, 0.3);
    margin: 0.8em 0;
  }

  a {
    color: inherit;
    text-decoration: underline;
    opacity: 0.85;
    &:hover { opacity: 1; }
  }
}

.typing-indicator {
  display: flex;
  gap: 5px;
  align-items: center;
  height: 20px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.5;
  animation: typing-bounce 1.4s infinite;
}

.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-indicator span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing-bounce {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.4;
  }
  40% {
    transform: scale(1.2);
    opacity: 1;
  }
}
</style>
