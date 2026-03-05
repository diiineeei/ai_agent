<template>
  <v-container class="pa-6" style="max-width: 960px">

    <!-- Page header -->
    <div class="d-flex align-center mb-6">
      <v-avatar color="tertiary" variant="tonal" size="48" rounded="lg" class="mr-3">
        <v-icon size="26">mdi-chart-bar</v-icon>
      </v-avatar>
      <div class="flex-grow-1">
        <h1 class="text-h5 font-weight-bold mb-0">Métricas</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Avaliações das respostas por agente.
        </p>
      </div>
      <v-btn icon variant="text" :loading="loading" @click="load">
        <v-icon>mdi-refresh</v-icon>
        <v-tooltip activator="parent" location="bottom">Atualizar</v-tooltip>
      </v-btn>
    </div>

    <v-alert v-if="error" type="error" variant="tonal" rounded="lg" class="mb-4">{{ error }}</v-alert>

    <!-- Loading -->
    <v-row v-if="loading">
      <v-col v-for="n in 3" :key="n" cols="12" sm="6" md="4">
        <v-skeleton-loader type="card" rounded="xl" />
      </v-col>
    </v-row>

    <!-- Stats cards -->
    <v-row v-else-if="rows.length > 0">
      <v-col v-for="row in rows" :key="row.agent_config_id" cols="12" sm="6" md="4">
        <v-card rounded="xl" height="100%">
          <!-- Agent header -->
          <div class="pa-4 d-flex align-center">
            <v-avatar color="primary" variant="tonal" size="44" class="mr-3 flex-shrink-0">
              <span class="text-body-1 font-weight-bold">{{ agentInitial(row.agent_config_id) }}</span>
            </v-avatar>
            <div class="overflow-hidden flex-grow-1">
              <div class="text-body-1 font-weight-bold text-truncate">{{ agentName(row.agent_config_id) }}</div>
              <div class="text-caption text-medium-emphasis">{{ row.thumbs_up + row.thumbs_down }} avaliações</div>
            </div>
          </div>

          <v-divider />

          <v-card-text class="pa-4">
            <!-- Satisfaction bar -->
            <div class="d-flex justify-space-between mb-1">
              <span class="text-caption text-medium-emphasis">Satisfação</span>
              <span class="text-caption font-weight-bold" :class="satisfactionColor(row)">
                {{ satisfactionPct(row) }}%
              </span>
            </div>
            <v-progress-linear
              :model-value="satisfactionPct(row)"
              :color="satisfactionColor(row)"
              bg-color="error-lighten-4"
              rounded="pill"
              height="8"
              class="mb-4"
            />

            <!-- Counts -->
            <div class="d-flex gap-3">
              <div class="d-flex align-center gap-1">
                <v-icon color="success" size="18">mdi-thumb-up</v-icon>
                <span class="text-body-2 font-weight-medium">{{ row.thumbs_up }}</span>
              </div>
              <div class="d-flex align-center gap-1">
                <v-icon color="error" size="18">mdi-thumb-down</v-icon>
                <span class="text-body-2 font-weight-medium">{{ row.thumbs_down }}</span>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Empty state -->
    <div v-else class="text-center py-16 text-medium-emphasis">
      <v-icon size="72" style="opacity:.15">mdi-chart-bar</v-icon>
      <p class="text-h6 mt-4 mb-1 font-weight-regular">Nenhuma avaliação ainda</p>
      <p class="text-body-2">Use os botões 👍 👎 no chat para avaliar as respostas.</p>
    </div>

  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { feedbackAPI } from '@/services/api'
import { useAgentConfigsStore } from '@/stores/agent_configs'

const agentConfigsStore = useAgentConfigsStore()

const rows    = ref([])
const loading = ref(false)
const error   = ref(null)

onMounted(async () => {
  await agentConfigsStore.fetchAll()
  await load()
})

async function load() {
  loading.value = true
  error.value = null
  try {
    const { data } = await feedbackAPI.stats()
    rows.value = data ?? []
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  } finally {
    loading.value = false
  }
}

function agentName(id) {
  if (!id) return 'Agente desconhecido'
  return agentConfigsStore.configs.find((c) => c.id === id)?.name ?? id
}

function agentInitial(id) {
  return agentName(id)[0]?.toUpperCase() ?? '?'
}

function satisfactionPct(row) {
  const total = row.thumbs_up + row.thumbs_down
  if (total === 0) return 0
  return Math.round((row.thumbs_up / total) * 100)
}

function satisfactionColor(row) {
  const pct = satisfactionPct(row)
  if (pct >= 70) return 'success'
  if (pct >= 40) return 'warning'
  return 'error'
}
</script>
