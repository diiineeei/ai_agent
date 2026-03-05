<template>
  <v-container class="pa-6" style="max-width: 900px">
    <div class="text-h5 font-weight-bold mb-1 d-flex align-center gap-2">
      <v-icon color="primary">mdi-puzzle</v-icon>
      Skills
    </div>
    <p class="text-body-2 text-medium-emphasis mb-6">
      Ative ou desative skills do agente. As alterações têm efeito a partir da próxima mensagem.
    </p>

    <!-- Error -->
    <v-alert
      v-if="store.error"
      type="error"
      variant="tonal"
      density="compact"
      class="mb-4"
      closable
      @click:close="store.error = null"
    >
      {{ store.error }}
    </v-alert>

    <!-- Loading skeletons -->
    <v-row v-if="store.loading">
      <v-col v-for="n in 2" :key="n" cols="12" sm="6">
        <v-skeleton-loader type="card" rounded="lg" />
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
        <v-card
          rounded="lg"
          :variant="skill.enabled ? 'elevated' : 'outlined'"
          :elevation="skill.enabled ? 2 : 0"
        >
          <v-card-item>
            <template #prepend>
              <v-avatar
                :color="skill.enabled ? 'primary' : undefined"
                :variant="skill.enabled ? 'tonal' : 'outlined'"
                size="48"
              >
                <v-icon>{{ skillIcon(skill.name) }}</v-icon>
              </v-avatar>
            </template>

            <v-card-title class="text-body-1 font-weight-bold">
              {{ skillLabel(skill.name) }}
            </v-card-title>
            <v-card-subtitle>{{ skill.name }}</v-card-subtitle>

            <template #append>
              <v-switch
                :model-value="skill.enabled"
                color="primary"
                hide-details
                density="compact"
                :loading="toggling === skill.name"
                @update:model-value="toggle(skill.name)"
              />
            </template>
          </v-card-item>

          <v-card-text class="pt-0">
            <p class="text-body-2 text-medium-emphasis mb-2">{{ skill.description }}</p>
            <v-chip
              size="small"
              :color="skill.enabled ? 'success' : 'default'"
              variant="tonal"
            >
              <v-icon start size="12">
                {{ skill.enabled ? 'mdi-check-circle-outline' : 'mdi-close-circle-outline' }}
              </v-icon>
              {{ skill.enabled ? 'Ativa' : 'Inativa' }}
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col v-if="!store.loading && store.skills.length === 0" cols="12">
        <v-alert type="info" variant="tonal">
          Nenhuma skill cadastrada.
        </v-alert>
      </v-col>
    </v-row>
  </v-container>

  <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="2500">
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
  weather: {
    label: 'Clima',
    icon: 'mdi-weather-partly-cloudy',
  },
  search_documents: {
    label: 'Busca em Documentos',
    icon: 'mdi-text-search',
  },
}

function skillLabel(name) {
  return SKILL_META[name]?.label ?? name
}

function skillIcon(name) {
  return SKILL_META[name]?.icon ?? 'mdi-puzzle-outline'
}
</script>
