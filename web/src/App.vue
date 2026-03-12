<template>
  <v-app :theme="theme">
    <v-navigation-drawer
      v-model="drawer"
      :permanent="!mobile"
      :temporary="mobile"
      :rail="!mobile && rail"
    >
      <!-- Header -->
      <div
        class="d-flex align-center px-2 bg-primary"
        style="height: 56px; gap: 4px;"
        :style="rail && !mobile ? 'cursor:pointer; justify-content:center' : ''"
        @click="rail && !mobile ? (rail = false) : undefined"
      >
        <v-icon color="white" size="22">mdi-robot</v-icon>
        <span v-if="!rail" class="text-body-1 font-weight-medium ml-1 flex-grow-1 text-white">Playground AI</span>
        <v-btn
          v-if="!mobile && !rail"
          icon="mdi-chevron-left"
          variant="text"
          color="white"
          size="small"
          @click.stop="rail = true"
        />
      </div>

      <v-divider />

      <v-list density="compact" nav class="mt-1">
        <v-list-item
          v-for="item in navItems"
          :key="item.to"
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          exact
          rounded="lg"
        />
      </v-list>

      <template #append>

        <!-- Últimas sessões (só quando expandido) -->
        <template v-if="!rail">
          <v-divider class="mb-1" />
          <div class="d-flex align-center px-3 pt-2 pb-1">
            <span class="text-caption text-medium-emphasis flex-grow-1">Recentes</span>
            <v-btn
              icon size="x-small" variant="text"
              :loading="loadingSessions"
              @click="fetchSessions"
            >
              <v-icon size="14">mdi-refresh</v-icon>
            </v-btn>
          </div>
          <v-list density="compact" nav class="pt-0 pb-1">
            <v-list-item
              v-for="s in recentSessions"
              :key="s.session_id"
              :active="s.session_id === chatStore.sessionId"
              color="primary"
              rounded="lg"
              class="session-item"
              @click="openSession(s)"
            >
              <template #prepend>
                <v-avatar
                  :color="agentOf(s) ? 'primary' : 'grey'"
                  variant="tonal"
                  size="28"
                  class="mr-2 flex-shrink-0"
                >
                  <v-img v-if="agentOf(s)?.avatar" :src="agentOf(s).avatar" cover />
                  <span v-else-if="agentOf(s)" style="font-size:11px;font-weight:700">
                    {{ agentOf(s).name[0].toUpperCase() }}
                  </span>
                  <v-icon v-else size="14">mdi-robot-outline</v-icon>
                </v-avatar>
              </template>
              <v-list-item-title class="text-caption font-weight-medium">
                {{ s.name || s.session_id.slice(0, 14) + '…' }}
              </v-list-item-title>
              <v-list-item-subtitle style="font-size:10px!important;opacity:.7">
                {{ agentOf(s)?.name ?? '—' }} · {{ s.message_count }} msg · {{ formatDate(s.updated_at) }}
              </v-list-item-subtitle>
            </v-list-item>
            <v-list-item v-if="!loadingSessions && recentSessions.length === 0" disabled>
              <v-list-item-title class="text-caption text-disabled">Nenhuma sessão</v-list-item-title>
            </v-list-item>
          </v-list>
        </template>

        <v-divider />
        <v-list density="compact" nav>
          <v-list-item
            :prepend-icon="theme === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night'"
            :title="theme === 'dark' ? 'Modo Claro' : 'Modo Escuro'"
            rounded="lg"
            @click="toggleTheme"
          />
        </v-list>
      </template>
    </v-navigation-drawer>

    <v-app-bar v-if="mobile" flat color="primary">
      <v-app-bar-nav-icon @click="drawer = !drawer" />
      <v-app-bar-title>
        <v-icon color="primary" class="mr-1">mdi-robot</v-icon>
        Playground AI
      </v-app-bar-title>
      <template #append>
        <v-btn
          :icon="theme === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night'"
          variant="text"
          @click="toggleTheme"
        />
      </template>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useDisplay } from 'vuetify'
import { useRouter } from 'vue-router'
import { chatAPI } from '@/services/api'
import { useChatStore } from '@/stores/chat'
import { useAgentConfigsStore } from '@/stores/agent_configs'

const { mobile } = useDisplay()
const router = useRouter()
const chatStore = useChatStore()
const agentConfigsStore = useAgentConfigsStore()

const drawer = ref(true)
const rail   = ref(false)
const theme  = ref('light')

const navItems = [
  { title: 'Chat',        icon: 'mdi-chat',              to: '/' },
  { title: 'Arquivos',    icon: 'mdi-file-multiple',     to: '/files' },
  { title: 'Skills',      icon: 'mdi-puzzle',            to: '/skills' },
  { title: 'Agentes',     icon: 'mdi-robot-happy',       to: '/agents' },
  { title: 'Servidores MCP', icon: 'mdi-connection',     to: '/mcp' },
  { title: 'Métricas',    icon: 'mdi-chart-bar',         to: '/metrics' },
]

function toggleTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}

// ── Sessões recentes ────────────────────────────────────────
const recentSessions  = ref([])
const loadingSessions = ref(false)

async function fetchSessions() {
  loadingSessions.value = true
  try {
    const { data } = await chatAPI.listSessions()
    recentSessions.value = (data ?? []).slice(0, 5)
  } catch { /* silencioso */ }
  finally { loadingSessions.value = false }
}

function agentOf(s) {
  return agentConfigsStore.configs.find(c => c.id === s.agent_config_id) ?? null
}

function openSession(s) {
  chatStore.setSession(s.session_id, s.name || null)
  chatStore.agentConfigId = s.agent_config_id || null
  chatStore.agentName     = agentConfigsStore.configs.find(c => c.id === s.agent_config_id)?.name ?? null
  chatStore.loadHistory()
  router.push('/')
  if (mobile.value) drawer.value = false
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return ''
  const now = new Date()
  const diffH = (now - d) / 3600000
  if (diffH < 24) return d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
}

// Atualiza a lista quando uma mensagem nova é enviada
watch(() => chatStore.messages.length, (len, prev) => {
  if (len > prev) fetchSessions()
})

// Expande o drawer para mostrar sessões ao abrir
watch(rail, (isRail) => {
  if (!isRail) fetchSessions()
})

onMounted(async () => {
  await agentConfigsStore.fetchAll()
  fetchSessions()
})
</script>

<style scoped>
.session-item :deep(.v-list-item-title) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
