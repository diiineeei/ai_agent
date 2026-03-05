<template>
  <v-container max-width="900">
    <div class="d-flex align-center mb-4">
      <h1 class="text-h5 font-weight-bold">Agentes</h1>
      <v-spacer />
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">
        Novo agente
      </v-btn>
    </div>

    <v-alert v-if="store.error" type="error" variant="tonal" class="mb-4">
      {{ store.error }}
    </v-alert>

    <v-progress-linear v-if="store.loading" indeterminate class="mb-4" />

    <v-row>
      <v-col
        v-for="cfg in store.configs"
        :key="cfg.id"
        cols="12"
        sm="6"
        md="4"
      >
        <v-card rounded="lg" height="100%">
          <v-card-title class="pt-4">{{ cfg.name }}</v-card-title>
          <v-card-subtitle>{{ cfg.model }}</v-card-subtitle>
          <v-card-text>
            <v-chip size="small" class="mr-1">
              {{ cfg.enabled_skills?.length ?? 0 }} skills
            </v-chip>
            <p
              v-if="cfg.system_instruction"
              class="text-body-2 text-medium-emphasis mt-2 mb-0"
              style="
                overflow: hidden;
                display: -webkit-box;
                -webkit-line-clamp: 2;
                -webkit-box-orient: vertical;
              "
            >
              {{ cfg.system_instruction }}
            </p>
          </v-card-text>
          <v-card-actions>
            <v-btn size="small" variant="text" prepend-icon="mdi-pencil" @click="openEdit(cfg)">
              Editar
            </v-btn>
            <v-btn
              size="small"
              variant="text"
              color="error"
              prepend-icon="mdi-delete"
              @click="openDelete(cfg)"
            >
              Excluir
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <div
      v-if="!store.loading && store.configs.length === 0"
      class="text-center text-medium-emphasis py-12"
    >
      <v-icon size="64" style="opacity: 0.3">mdi-robot-happy</v-icon>
      <p class="mt-4">Nenhum agente cadastrado.</p>
    </div>
  </v-container>

  <!-- Create / Edit dialog -->
  <v-dialog v-model="formDialog" max-width="560" persistent>
    <v-card rounded="lg">
      <v-card-title class="pt-4">
        {{ editTarget ? 'Editar agente' : 'Novo agente' }}
      </v-card-title>
      <v-card-text>
        <v-text-field
          v-model="form.name"
          label="Nome *"
          variant="outlined"
          density="comfortable"
          class="mb-3"
          hide-details="auto"
          :rules="[(v) => !!v || 'Obrigatório']"
        />
        <v-combobox
          v-model="form.model"
          label="Modelo *"
          :items="modelSuggestions"
          variant="outlined"
          density="comfortable"
          class="mb-3"
          hide-details="auto"
        />
        <v-textarea
          v-model="form.system_instruction"
          label="Instrução de sistema"
          variant="outlined"
          density="comfortable"
          auto-grow
          rows="3"
          class="mb-1"
          hide-details
        />
        <div class="d-flex justify-end mb-3">
          <v-btn
            size="small"
            variant="tonal"
            color="secondary"
            prepend-icon="mdi-auto-fix"
            :loading="improving"
            :disabled="!form.model || !form.system_instruction.trim()"
            @click="improveInstruction"
          >
            Melhorar com IA
          </v-btn>
        </div>
        <div class="text-body-2 mb-2">Skills habilitadas</div>
        <v-checkbox
          v-for="skill in skillsStore.skills"
          :key="skill.name"
          v-model="form.enabled_skills"
          :label="skill.name"
          :value="skill.name"
          density="compact"
          hide-details
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="closeForm">Cancelar</v-btn>
        <v-btn color="primary" :loading="saving" @click="submitForm">Salvar</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Delete confirm dialog -->
  <v-dialog v-model="deleteDialog" max-width="380">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Excluir agente</v-card-title>
      <v-card-text>
        Tem certeza que deseja excluir <strong>{{ deleteTarget?.name }}</strong>?
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="deleteDialog = false">Cancelar</v-btn>
        <v-btn color="error" :loading="deleting" @click="confirmDelete">Excluir</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAgentConfigsStore } from '@/stores/agent_configs'
import { useSkillsStore } from '@/stores/skills'
import { agentConfigsAPI } from '@/services/api'

const store = useAgentConfigsStore()
const skillsStore = useSkillsStore()

const modelSuggestions = ['gemini-2.5-flash', 'gemini-2.5-pro']

const formDialog = ref(false)
const deleteDialog = ref(false)
const saving = ref(false)
const deleting = ref(false)
const improving = ref(false)
const editTarget = ref(null)
const deleteTarget = ref(null)

const emptyForm = () => ({
  name: '',
  model: 'gemini-2.5-flash',
  system_instruction: '',
  enabled_skills: [],
})
const form = ref(emptyForm())

onMounted(() => {
  store.fetchAll()
  skillsStore.fetchSkills()
})

function openCreate() {
  editTarget.value = null
  form.value = emptyForm()
  formDialog.value = true
}

function openEdit(cfg) {
  editTarget.value = cfg
  form.value = {
    name: cfg.name,
    model: cfg.model,
    system_instruction: cfg.system_instruction ?? '',
    enabled_skills: [...(cfg.enabled_skills ?? [])],
  }
  formDialog.value = true
}

function closeForm() {
  formDialog.value = false
}

async function improveInstruction() {
  improving.value = true
  try {
    const { data } = await agentConfigsAPI.improveInstruction(
      form.value.model,
      form.value.system_instruction,
    )
    form.value.system_instruction = data.instruction
  } finally {
    improving.value = false
  }
}

async function submitForm() {
  if (!form.value.name || !form.value.model) return
  saving.value = true
  try {
    if (editTarget.value) {
      await store.update(editTarget.value.id, form.value)
    } else {
      await store.create(form.value)
    }
    formDialog.value = false
  } finally {
    saving.value = false
  }
}

function openDelete(cfg) {
  deleteTarget.value = cfg
  deleteDialog.value = true
}

async function confirmDelete() {
  deleting.value = true
  try {
    await store.remove(deleteTarget.value.id)
    deleteDialog.value = false
  } finally {
    deleting.value = false
  }
}
</script>
