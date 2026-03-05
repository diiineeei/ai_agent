<template>
  <v-container class="pa-6" style="max-width: 900px">
    <div class="text-h5 font-weight-bold mb-1 d-flex align-center gap-2">
      <v-icon color="primary">mdi-cog</v-icon>
      Configurações
    </div>
    <p class="text-body-2 text-medium-emphasis mb-6">
      Edite a instrução de sistema do agente. As alterações têm efeito a partir da próxima mensagem.
    </p>

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

    <v-card rounded="lg">
      <v-card-text>
        <v-skeleton-loader v-if="store.loading" type="paragraph" />
        <v-textarea
          v-else
          v-model="draft"
          label="Instrução de sistema"
          auto-grow
          rows="10"
          variant="outlined"
          :disabled="store.saving"
        />
      </v-card-text>

      <v-card-actions class="px-4 pb-4 gap-2">
        <v-btn
          color="primary"
          variant="flat"
          :loading="store.saving"
          :disabled="store.loading"
          @click="handleSave"
        >
          Salvar
        </v-btn>
        <v-btn
          variant="text"
          :disabled="store.loading || store.saving"
          @click="handleCancel"
        >
          Cancelar
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>

  <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="2500">
    {{ snackbarMsg }}
  </v-snackbar>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'

const store = useSettingsStore()
const draft = ref('')
const snackbar = ref(false)
const snackbarMsg = ref('')
const snackbarColor = ref('success')

onMounted(async () => {
  await store.fetch()
  draft.value = store.instruction
})

async function handleSave() {
  try {
    await store.save(draft.value)
    snackbarMsg.value = 'Instrução salva com sucesso.'
    snackbarColor.value = 'success'
    snackbar.value = true
  } catch {
    snackbarMsg.value = store.error || 'Erro ao salvar instrução.'
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}

function handleCancel() {
  draft.value = store.instruction
}
</script>
