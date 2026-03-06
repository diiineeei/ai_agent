<template>
  <v-container class="pa-6" style="max-width: 960px">

    <!-- Page header -->
    <div class="d-flex align-center mb-6">
      <v-avatar color="primary" variant="tonal" size="48" rounded="lg" class="mr-3">
        <v-icon size="26">mdi-puzzle-outline</v-icon>
      </v-avatar>
      <div class="flex-grow-1">
        <h1 class="text-h5 font-weight-bold mb-0">Skills</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Ative ou desative capacidades globais dos agentes.
        </p>
      </div>
    </div>

    <v-alert v-if="store.error" type="error" variant="tonal" rounded="lg" class="mb-4" closable
      @click:close="store.error = null">
      {{ store.error }}
    </v-alert>

    <!-- Loading -->
    <v-row v-if="store.loading">
      <v-col v-for="n in 2" :key="n" cols="12" sm="6">
        <v-skeleton-loader type="card" rounded="xl" />
      </v-col>
    </v-row>

    <!-- Skill cards -->
    <v-row v-else>
      <v-col
        v-for="skill in store.skills"
        :key="skill.name"
        cols="12"
        sm="6"
      >
        <v-card rounded="xl" height="100%">
          <!-- Card header -->
          <div class="pa-4 d-flex align-center">
            <v-avatar
              :color="skill.enabled ? 'primary' : undefined"
              :variant="skill.enabled ? 'tonal' : 'outlined'"
              size="48"
              class="flex-shrink-0 mr-3"
            >
              <v-icon :color="skill.enabled ? 'primary' : 'medium-emphasis'">
                {{ skillIcon(skill.name) }}
              </v-icon>
            </v-avatar>

            <div class="flex-grow-1 overflow-hidden">
              <div class="text-body-1 font-weight-bold text-truncate">
                {{ skillLabel(skill.name) }}
              </div>
              <div class="text-caption text-medium-emphasis text-truncate">
                {{ skill.name }}
              </div>
            </div>

            <v-switch
              :model-value="skill.enabled"
              color="primary"
              hide-details
              density="compact"
              :loading="toggling === skill.name"
              @update:model-value="toggle(skill.name)"
            />
          </div>

          <v-divider />

          <v-card-text class="pa-4">
            <p class="text-body-2 text-medium-emphasis mb-3">
              {{ skill.description }}
            </p>
            <v-chip
              size="small"
              :color="skill.enabled ? 'success' : 'default'"
              variant="tonal"
            >
              <v-icon start size="13">
                {{ skill.enabled ? 'mdi-check-circle-outline' : 'mdi-close-circle-outline' }}
              </v-icon>
              {{ skill.enabled ? 'Ativa' : 'Inativa' }}
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Empty state -->
      <v-col v-if="store.skills.length === 0" cols="12">
        <div class="text-center py-16 text-medium-emphasis">
          <v-icon size="72" style="opacity:.15">mdi-puzzle-outline</v-icon>
          <p class="text-h6 mt-4 mb-1 font-weight-regular">Nenhuma skill disponível</p>
          <p class="text-body-2">As skills são registradas automaticamente pelo servidor.</p>
        </div>
      </v-col>
    </v-row>
  </v-container>

  <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="2500" rounded="lg">
    {{ snackbarMsg }}
  </v-snackbar>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSkillsStore } from '@/stores/skills'

const store = useSkillsStore()
const toggling = ref(null)
const snackbar = ref(false)
const snackbarMsg = ref('')
const snackbarColor = ref('success')

onMounted(() => store.fetchSkills())

async function toggle(name) {
  toggling.value = name
  try {
    await store.toggleSkill(name)
    const skill = store.skills.find((s) => s.name === name)
    snackbarMsg.value = `Skill "${skillLabel(name)}" ${skill?.enabled ? 'ativada' : 'desativada'}.`
    snackbarColor.value = skill?.enabled ? 'success' : 'info'
    snackbar.value = true
  } catch {
    snackbarMsg.value = store.error || 'Erro ao alterar skill.'
    snackbarColor.value = 'error'
    snackbar.value = true
  } finally {
    toggling.value = null
  }
}

const SKILL_META = {
  weather:           { label: 'Clima',              icon: 'mdi-weather-partly-cloudy' },
  search_documents:  { label: 'Busca em Documentos', icon: 'mdi-text-search' },
  suggest_questions: { label: 'Sugestões',           icon: 'mdi-help-circle-outline' },
}

const skillLabel = (name) => SKILL_META[name]?.label ?? name
const skillIcon  = (name) => SKILL_META[name]?.icon  ?? 'mdi-puzzle-outline'
</script>
