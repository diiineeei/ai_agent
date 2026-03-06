<template>
  <v-container class="pa-6" style="max-width: 960px">

    <!-- Page header -->
    <div class="d-flex align-center mb-6">
      <v-avatar color="primary" variant="tonal" size="48" rounded="lg" class="mr-3">
        <v-icon size="26">mdi-robot-happy</v-icon>
      </v-avatar>
      <div class="flex-grow-1">
        <h1 class="text-h5 font-weight-bold mb-0">Agentes</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Configure agentes com modelos, instruções e skills diferentes.
        </p>
      </div>
      <v-btn color="primary" variant="flat" rounded="lg" prepend-icon="mdi-plus" @click="openCreate">
        Novo agente
      </v-btn>
    </div>

    <v-alert v-if="store.error" type="error" variant="tonal" rounded="lg" class="mb-4" closable>
      {{ store.error }}
    </v-alert>

    <!-- Loading -->
    <v-row v-if="store.loading">
      <v-col v-for="n in 3" :key="n" cols="12" sm="6" md="4">
        <v-skeleton-loader type="card" rounded="xl" />
      </v-col>
    </v-row>

    <!-- Agent cards -->
    <v-row v-else>
      <v-col
        v-for="cfg in store.configs"
        :key="cfg.id"
        cols="12" sm="6" md="4"
      >
        <v-card rounded="xl" height="100%" class="d-flex flex-column">
          <!-- Card header -->
          <div class="pa-4 d-flex align-center">
            <v-avatar color="primary" variant="tonal" size="44" class="flex-shrink-0 mr-3">
              <span class="text-body-1 font-weight-bold">{{ cfg.name[0].toUpperCase() }}</span>
            </v-avatar>
            <div class="overflow-hidden flex-grow-1">
              <div class="text-body-1 font-weight-bold text-truncate">{{ cfg.name }}</div>
              <div class="d-flex align-center gap-1 mt-1 flex-wrap">
                <v-chip size="x-small" variant="tonal" color="secondary">
                  <v-icon start size="10">mdi-chip</v-icon>
                  {{ cfg.model }}
                </v-chip>
                <v-chip
                  v-if="cfg.provider === 'ollama'"
                  size="x-small"
                  variant="tonal"
                  color="orange"
                >
                  <v-icon start size="10">mdi-server-outline</v-icon>
                  Ollama
                </v-chip>
              </div>
            </div>
          </div>

          <v-divider />

          <v-card-text class="flex-grow-1 pa-4">
            <!-- System instruction preview -->
            <p
              v-if="cfg.system_instruction"
              class="text-body-2 text-medium-emphasis mb-3"
              style="overflow:hidden;display:-webkit-box;-webkit-line-clamp:3;-webkit-box-orient:vertical"
            >
              {{ cfg.system_instruction }}
            </p>
            <p v-else class="text-caption text-disabled font-italic mb-3">Sem instrução de sistema.</p>

            <!-- Skills chips -->
            <div class="d-flex flex-wrap gap-1">
              <v-chip
                v-for="s in cfg.enabled_skills"
                :key="s"
                size="x-small"
                variant="tonal"
                color="primary"
                :prepend-icon="skillIcon(s)"
              >
                {{ skillLabel(s) }}
              </v-chip>
              <span v-if="!cfg.enabled_skills?.length" class="text-caption text-disabled">
                Nenhuma skill
              </span>
            </div>
          </v-card-text>

          <v-divider />

          <v-card-actions class="pa-2">
            <v-btn size="small" variant="text" prepend-icon="mdi-pencil-outline" @click="openEdit(cfg)">
              Editar
            </v-btn>
            <v-spacer />
            <v-btn size="small" variant="text" color="error" icon @click="openDelete(cfg)">
              <v-icon size="18">mdi-delete-outline</v-icon>
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>

      <!-- Empty state -->
      <v-col v-if="store.configs.length === 0" cols="12">
        <div class="text-center py-16 text-medium-emphasis">
          <v-icon size="72" style="opacity:.15">mdi-robot-happy</v-icon>
          <p class="text-h6 mt-4 mb-1 font-weight-regular">Nenhum agente cadastrado</p>
          <p class="text-body-2">Clique em <strong>Novo agente</strong> para começar.</p>
        </div>
      </v-col>
    </v-row>
  </v-container>

  <!-- ── Create / Edit dialog ─────────────────────────── -->
  <v-dialog v-model="formDialog" max-width="620" persistent scrollable>
    <v-card rounded="xl">

      <!-- Dialog header -->
      <div class="pa-5 d-flex align-center">
        <v-avatar color="primary" variant="tonal" size="52" rounded="lg" class="mr-4">
          <v-icon size="28">mdi-robot-happy</v-icon>
        </v-avatar>
        <div>
          <div class="text-h6 font-weight-bold">
            {{ editTarget ? 'Editar agente' : 'Novo agente' }}
          </div>
          <div class="text-body-2 text-medium-emphasis">
            Configure o comportamento do agente de IA
          </div>
        </div>
      </div>

      <v-divider />

      <v-card-text class="pa-5">

        <!-- Name row -->
        <v-row dense>
          <v-col cols="12">
            <v-text-field
              v-model="form.name"
              label="Nome *"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-tag-outline"
              hide-details="auto"
              :rules="[(v) => !!v || 'Obrigatório']"
            />
          </v-col>
        </v-row>

        <!-- Provider + Model row -->
        <v-row dense class="mt-2">
          <v-col cols="12" sm="4">
            <v-select
              v-model="form.provider"
              label="Provedor"
              :items="[{ title: 'Gemini', value: 'gemini' }, { title: 'Ollama', value: 'ollama' }]"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-server-outline"
              hide-details
            />
          </v-col>
          <v-col cols="12" sm="8">
            <v-combobox
              v-model="form.model"
              label="Modelo *"
              :items="modelSuggestions"
              item-title="title"
              item-value="title"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-chip"
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <!-- Ollama info -->
        <v-alert
          v-if="form.provider === 'ollama'"
          type="info"
          variant="tonal"
          density="compact"
          rounded="lg"
          class="mt-3"
          icon="mdi-information-outline"
        >
          Modelos menores como <strong>llama3.2:1b</strong> e <strong>3b</strong> têm comportamento instável com skills.
          Para uso com skills, prefira <strong>llama3.1</strong>, <strong>llama3.2</strong> (8b) ou <strong>qwen2.5</strong>.
        </v-alert>

        <!-- Base URL (Ollama only) -->
        <v-row v-if="form.provider === 'ollama'" dense class="mt-2">
          <v-col cols="12">
            <v-text-field
              v-model="form.base_url"
              label="URL do Ollama"
              placeholder="http://localhost:11434"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-link-variant"
              hide-details
            />
          </v-col>
        </v-row>

        <!-- System instruction -->
        <div class="d-flex align-center gap-1 mt-5 mb-2">
          <v-icon size="16" color="medium-emphasis">mdi-text-box-outline</v-icon>
          <span class="text-body-2 font-weight-medium">Instrução de sistema</span>
        </div>

        <v-textarea
          v-model="form.system_instruction"
          placeholder="Ex: Você é um assistente especializado em análise de dados. Responda sempre em português de forma objetiva e técnica."
          variant="outlined"
          density="comfortable"
          auto-grow
          rows="4"
          hide-details
        />

        <div class="d-flex align-center justify-space-between mt-2 mb-1">
          <span class="text-caption text-disabled">Define como o agente se comporta</span>
          <v-btn
            size="small"
            variant="tonal"
            color="secondary"
            rounded="lg"
            prepend-icon="mdi-auto-fix"
            :loading="improving"
            :disabled="!form.model || !form.system_instruction.trim()"
            @click="improveInstruction"
          >
            Melhorar descrição
          </v-btn>
        </div>

        <!-- Skills -->
        <div class="d-flex align-center gap-1 mt-5 mb-3">
          <v-icon size="16" color="medium-emphasis">mdi-puzzle-outline</v-icon>
          <span class="text-body-2 font-weight-medium">Skills habilitadas</span>
        </div>

        <div class="d-flex flex-wrap gap-2">
          <v-card
            v-for="skill in skillsStore.skills"
            :key="skill.name"
            :variant="form.enabled_skills.includes(skill.name) ? 'flat' : 'outlined'"
            :color="form.enabled_skills.includes(skill.name) ? 'primary' : undefined"
            rounded="lg"
            class="skill-chip cursor-pointer"
            @click="toggleSkill(skill.name)"
          >
            <div class="d-flex align-center gap-2 px-3 py-2">
              <v-icon
                :color="form.enabled_skills.includes(skill.name) ? 'on-primary' : 'medium-emphasis'"
                size="18"
              >
                {{ skillIcon(skill.name) }}
              </v-icon>
              <span
                class="text-body-2 font-weight-medium"
                :class="form.enabled_skills.includes(skill.name) ? 'text-on-primary' : 'text-medium-emphasis'"
              >
                {{ skillLabel(skill.name) }}
              </span>
              <v-icon
                v-if="form.enabled_skills.includes(skill.name)"
                size="14"
                color="on-primary"
              >
                mdi-check
              </v-icon>
            </div>
          </v-card>
          <p v-if="!skillsStore.skills.length" class="text-caption text-disabled">
            Nenhuma skill disponível.
          </p>
        </div>

      </v-card-text>

      <v-divider />

      <v-card-actions class="pa-4">
        <v-btn variant="text" rounded="lg" @click="closeForm">Cancelar</v-btn>
        <v-spacer />
        <v-btn
          color="primary"
          variant="flat"
          rounded="lg"
          min-width="120"
          :loading="saving"
          :disabled="!form.name || !form.model"
          @click="submitForm"
        >
          Salvar
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- ── Delete confirm dialog ─────────────────────────── -->
  <v-dialog v-model="deleteDialog" max-width="400">
    <v-card rounded="xl">
      <v-card-text class="pa-6 text-center">
        <v-avatar color="error" variant="tonal" size="56" class="mb-4">
          <v-icon size="28">mdi-delete-outline</v-icon>
        </v-avatar>
        <p class="text-h6 font-weight-bold mb-1">Excluir agente</p>
        <p class="text-body-2 text-medium-emphasis">
          Tem certeza que deseja excluir <strong>{{ deleteTarget?.name }}</strong>?
          Esta ação não pode ser desfeita.
        </p>
      </v-card-text>
      <v-card-actions class="pa-4 pt-0">
        <v-btn variant="text" rounded="lg" class="flex-grow-1" @click="deleteDialog = false">
          Cancelar
        </v-btn>
        <v-btn
          color="error"
          variant="flat"
          rounded="lg"
          class="flex-grow-1"
          :loading="deleting"
          @click="confirmDelete"
        >
          Excluir
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgentConfigsStore } from '@/stores/agent_configs'
import { useSkillsStore } from '@/stores/skills'
import { agentConfigsAPI } from '@/services/api'

const route = useRoute()
const store = useAgentConfigsStore()
const skillsStore = useSkillsStore()

const GEMINI_MODELS = [
  { title: 'gemini-2.5-flash', subtitle: 'Rápido e eficiente' },
  { title: 'gemini-2.5-pro',   subtitle: 'Mais capaz' },
]
const OLLAMA_MODELS = [
  { title: 'llama3.2:1b',  subtitle: '~1.3 GB · mais leve' },
  { title: 'llama3.2:3b',  subtitle: '~2 GB' },
  { title: 'qwen2.5:3b',   subtitle: '~2 GB' },
  { title: 'phi4-mini',    subtitle: '~2.5 GB' },
  { title: 'llama3.2',     subtitle: '~5 GB' },
  { title: 'llama3.1',     subtitle: '~5 GB' },
  { title: 'mistral',      subtitle: '~5 GB' },
  { title: 'qwen2.5',      subtitle: '~5 GB' },
  { title: 'phi4',         subtitle: '~9 GB' },
  { title: 'gemma3',       subtitle: '~9 GB' },
]

const modelSuggestions = computed(() =>
  form.value.provider === 'ollama' ? OLLAMA_MODELS : GEMINI_MODELS
)

const SKILL_META = {
  weather:           { label: 'Clima',              icon: 'mdi-weather-partly-cloudy' },
  search_documents:  { label: 'Busca em Documentos', icon: 'mdi-text-search' },
  suggest_questions: { label: 'Sugestões',           icon: 'mdi-help-circle-outline' },
}
const skillLabel = (name) => SKILL_META[name]?.label ?? name
const skillIcon  = (name) => SKILL_META[name]?.icon  ?? 'mdi-puzzle-outline'

const formDialog   = ref(false)
const deleteDialog = ref(false)
const saving       = ref(false)
const deleting     = ref(false)
const improving    = ref(false)
const editTarget   = ref(null)
const deleteTarget = ref(null)

const emptyForm = () => ({ name: '', model: 'gemini-2.5-flash', provider: 'gemini', base_url: '', system_instruction: '', enabled_skills: [] })
const form = ref(emptyForm())

watch(() => form.value.provider, (provider, prev) => {
  if (provider === prev) return
  form.value.model = provider === 'ollama' ? OLLAMA_MODELS[0].title : GEMINI_MODELS[0].title
  if (provider !== 'ollama') form.value.base_url = ''
})

onMounted(async () => {
  await Promise.all([store.fetchAll(), skillsStore.fetchSkills()])
  if (route.query.edit) {
    const cfg = store.configs.find((c) => c.id === route.query.edit)
    if (cfg) openEdit(cfg)
  }
})

function toggleSkill(name) {
  const idx = form.value.enabled_skills.indexOf(name)
  if (idx === -1) form.value.enabled_skills.push(name)
  else form.value.enabled_skills.splice(idx, 1)
}

function openCreate() { editTarget.value = null; form.value = emptyForm(); formDialog.value = true }

function openEdit(cfg) {
  editTarget.value = cfg
  form.value = { name: cfg.name, model: cfg.model, provider: cfg.provider ?? 'gemini', base_url: cfg.base_url ?? '', system_instruction: cfg.system_instruction ?? '', enabled_skills: [...(cfg.enabled_skills ?? [])] }
  formDialog.value = true
}

function closeForm() { formDialog.value = false }

async function improveInstruction() {
  improving.value = true
  try {
    const { data } = await agentConfigsAPI.improveInstruction(form.value.model, form.value.system_instruction)
    form.value.system_instruction = data.instruction
  } finally { improving.value = false }
}

async function submitForm() {
  if (!form.value.name || !form.value.model) return
  saving.value = true
  try {
    const payload = {
      ...form.value,
      model: typeof form.value.model === 'object' ? form.value.model.title : form.value.model,
    }
    editTarget.value ? await store.update(editTarget.value.id, payload) : await store.create(payload)
    formDialog.value = false
  } finally { saving.value = false }
}

function openDelete(cfg) { deleteTarget.value = cfg; deleteDialog.value = true }

async function confirmDelete() {
  deleting.value = true
  try { await store.remove(deleteTarget.value.id); deleteDialog.value = false }
  finally { deleting.value = false }
}
</script>

<style scoped>
.skill-chip { transition: all .15s ease; user-select: none; }
.skill-chip:hover { transform: translateY(-1px); }
</style>
