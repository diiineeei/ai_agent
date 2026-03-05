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
        @click="sessionDialog = true"
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
          <p class="text-body-1 mb-0" style="white-space: pre-wrap">{{ msg.text }}</p>
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

  <!-- Change session dialog -->
  <v-dialog v-model="sessionDialog" max-width="440">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Sessão</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="sessionInput"
          label="ID da sessão"
          variant="outlined"
          density="comfortable"
          hide-details
          @keydown.enter="applySession"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="sessionDialog = false">Cancelar</v-btn>
        <v-btn color="primary" @click="applySession">Aplicar</v-btn>
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

const store = useChatStore()
const agentConfigsStore = useAgentConfigsStore()

const input = ref('')
const messagesEl = ref(null)
const loadingHistory = ref(false)
const clearDialog = ref(false)
const sessionDialog = ref(false)
const newSessionDialog = ref(false)
const sessionInput = ref(store.sessionId)
const selectedConfigId = ref(null)

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

function applySession() {
  store.setSession(sessionInput.value)
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
