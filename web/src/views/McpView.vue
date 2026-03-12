<template>
  <v-container class="pa-6" style="max-width: 960px">

    <!-- Page header -->
    <div class="d-flex align-center mb-6">
      <v-avatar color="primary" variant="tonal" size="48" rounded="lg" class="mr-3">
        <v-icon size="26">mdi-connection</v-icon>
      </v-avatar>
      <div class="flex-grow-1">
        <h1 class="text-h5 font-weight-bold mb-0">Servidores MCP</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Gerencie conexões com servidores Model Context Protocol externos.
        </p>
      </div>
      <v-btn color="primary" variant="flat" rounded="lg" prepend-icon="mdi-plus" @click="openCreate">
        Novo servidor
      </v-btn>
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

    <!-- Server cards -->
    <v-row v-else>
      <v-col
        v-for="srv in store.servers"
        :key="srv.id"
        cols="12"
        sm="6"
      >
        <v-card rounded="xl" height="100%" class="d-flex flex-column">

          <!-- Card header -->
          <div class="pa-4 d-flex align-center">
            <v-avatar
              :color="srv.enabled ? 'primary' : undefined"
              :variant="srv.enabled ? 'tonal' : 'outlined'"
              size="44"
              class="flex-shrink-0 mr-3"
            >
              <v-icon :color="srv.enabled ? 'primary' : 'medium-emphasis'">
                {{ transportIcon(srv.transport) }}
              </v-icon>
            </v-avatar>

            <div class="flex-grow-1 overflow-hidden">
              <div class="text-body-1 font-weight-bold text-truncate">{{ srv.name }}</div>
              <div class="d-flex align-center gap-1 mt-1">
                <v-chip size="x-small" variant="tonal" color="secondary">
                  {{ srv.transport }}
                </v-chip>
                <v-chip
                  size="x-small"
                  :color="srv.enabled ? 'success' : 'default'"
                  variant="tonal"
                >
                  {{ srv.enabled ? 'Ativo' : 'Inativo' }}
                </v-chip>
              </div>
            </div>

            <v-switch
              :model-value="srv.enabled"
              color="primary"
              hide-details
              density="compact"
              :loading="toggling === srv.id"
              @update:model-value="toggle(srv)"
            />
          </div>

          <v-divider />

          <v-card-text class="flex-grow-1 pa-4">
            <p v-if="srv.description" class="text-body-2 text-medium-emphasis mb-3">
              {{ srv.description }}
            </p>

            <!-- stdio details -->
            <div v-if="srv.transport === 'stdio'" class="command-block rounded-lg pa-2 mb-2">
              <span class="text-caption text-medium-emphasis">$ </span>
              <span class="text-caption font-weight-medium">{{ srv.command }}</span>
              <span v-if="srv.args?.length" class="text-caption text-medium-emphasis">
                {{ ' ' + srv.args.join(' ') }}
              </span>
            </div>

            <!-- http details -->
            <div v-if="srv.transport === 'http'" class="command-block rounded-lg pa-2 mb-2">
              <span class="text-caption text-medium-emphasis">URL: </span>
              <span class="text-caption font-weight-medium">{{ srv.url }}</span>
            </div>

            <!-- Env vars count -->
            <div v-if="envCount(srv) > 0" class="d-flex align-center gap-1">
              <v-icon size="13" color="medium-emphasis">mdi-variable</v-icon>
              <span class="text-caption text-medium-emphasis">{{ envCount(srv) }} variável(is) de ambiente</span>
            </div>
          </v-card-text>

          <v-divider />

          <v-card-actions class="pa-2">
            <v-btn size="small" variant="text" prepend-icon="mdi-pencil-outline" @click="openEdit(srv)">
              Editar
            </v-btn>
            <v-spacer />
            <v-btn size="small" variant="text" color="error" icon @click="openDelete(srv)">
              <v-icon size="18">mdi-delete-outline</v-icon>
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>

      <!-- Empty state -->
      <v-col v-if="store.servers.length === 0" cols="12">
        <div class="text-center py-16 text-medium-emphasis">
          <v-icon size="72" style="opacity:.15">mdi-connection</v-icon>
          <p class="text-h6 mt-4 mb-1 font-weight-regular">Nenhum servidor MCP cadastrado</p>
          <p class="text-body-2">
            Clique em <strong>Novo servidor</strong> para conectar um servidor MCP.
          </p>
        </div>
      </v-col>
    </v-row>
  </v-container>

  <!-- ── Create / Edit dialog ─────────────────────────── -->
  <v-dialog v-model="formDialog" max-width="640" persistent scrollable>
    <v-card rounded="xl">

      <!-- Dialog header -->
      <div class="pa-5 d-flex align-center">
        <v-avatar color="primary" variant="tonal" size="48" rounded="lg" class="mr-4">
          <v-icon size="26">mdi-connection</v-icon>
        </v-avatar>
        <div>
          <div class="text-h6 font-weight-bold">
            {{ editTarget ? 'Editar servidor MCP' : 'Novo servidor MCP' }}
          </div>
          <div class="text-body-2 text-medium-emphasis">
            Configure a conexão com um servidor MCP externo
          </div>
        </div>
      </div>

      <v-divider />

      <v-card-text class="pa-5">

        <!-- Name + enabled -->
        <v-row dense>
          <v-col cols="12" sm="8">
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
          <v-col cols="12" sm="4" class="d-flex align-center">
            <v-switch
              v-model="form.enabled"
              label="Habilitado"
              color="primary"
              hide-details
              density="comfortable"
            />
          </v-col>
        </v-row>

        <!-- Description -->
        <v-row dense class="mt-2">
          <v-col cols="12">
            <v-text-field
              v-model="form.description"
              label="Descrição"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-text-short"
              hide-details
            />
          </v-col>
        </v-row>

        <!-- Transport -->
        <v-row dense class="mt-2">
          <v-col cols="12">
            <v-btn-toggle
              v-model="form.transport"
              color="primary"
              variant="outlined"
              rounded="lg"
              mandatory
              density="comfortable"
            >
              <v-btn value="stdio" prepend-icon="mdi-console">stdio</v-btn>
              <v-btn value="http" prepend-icon="mdi-web">HTTP / SSE</v-btn>
            </v-btn-toggle>
          </v-col>
        </v-row>

        <!-- stdio fields -->
        <template v-if="form.transport === 'stdio'">
          <v-row dense class="mt-3">
            <v-col cols="12">
              <v-text-field
                v-model="form.command"
                label="Comando *"
                placeholder="Ex: /usr/local/bin/mcp-weather"
                variant="outlined"
                density="comfortable"
                prepend-inner-icon="mdi-console-line"
                hide-details="auto"
                :rules="[(v) => !!v || 'Obrigatório para stdio']"
              />
            </v-col>
          </v-row>
          <v-row dense class="mt-2">
            <v-col cols="12">
              <v-combobox
                v-model="form.args"
                label="Argumentos"
                placeholder="Pressione Enter para adicionar cada argumento"
                variant="outlined"
                density="comfortable"
                prepend-inner-icon="mdi-format-list-bulleted"
                chips
                closable-chips
                multiple
                hide-details
              />
            </v-col>
          </v-row>
        </template>

        <!-- http fields -->
        <template v-if="form.transport === 'http'">
          <v-row dense class="mt-3">
            <v-col cols="12">
              <v-text-field
                v-model="form.url"
                label="URL *"
                placeholder="Ex: http://localhost:3000/mcp"
                variant="outlined"
                density="comfortable"
                prepend-inner-icon="mdi-link-variant"
                hide-details="auto"
                :rules="[(v) => !!v || 'Obrigatório para HTTP']"
              />
            </v-col>
          </v-row>
        </template>

        <!-- Environment variables -->
        <div class="d-flex align-center gap-1 mt-5 mb-2">
          <v-icon size="16" color="medium-emphasis">mdi-variable</v-icon>
          <span class="text-body-2 font-weight-medium">Variáveis de ambiente</span>
          <v-spacer />
          <v-btn size="x-small" variant="tonal" color="primary" prepend-icon="mdi-plus" @click="addEnvVar">
            Adicionar
          </v-btn>
        </div>

        <div v-for="(pair, i) in envPairs" :key="i" class="d-flex gap-2 mb-2">
          <v-text-field
            v-model="pair.key"
            label="Chave"
            variant="outlined"
            density="compact"
            hide-details
            style="flex:1"
          />
          <v-text-field
            v-model="pair.value"
            label="Valor"
            variant="outlined"
            density="compact"
            hide-details
            style="flex:2"
          />
          <v-btn icon variant="text" color="error" size="small" @click="removeEnvVar(i)">
            <v-icon size="18">mdi-close</v-icon>
          </v-btn>
        </div>

        <p v-if="envPairs.length === 0" class="text-caption text-disabled mt-1">
          Nenhuma variável de ambiente configurada.
        </p>

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
          :disabled="!canSave"
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
        <p class="text-h6 font-weight-bold mb-1">Excluir servidor MCP</p>
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

  <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="2500" rounded="lg">
    {{ snackbarMsg }}
  </v-snackbar>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMcpServersStore } from '@/stores/mcp_servers'

const store = useMcpServersStore()

const formDialog   = ref(false)
const deleteDialog = ref(false)
const saving       = ref(false)
const deleting     = ref(false)
const toggling     = ref(null)
const editTarget   = ref(null)
const deleteTarget = ref(null)
const snackbar     = ref(false)
const snackbarMsg  = ref('')
const snackbarColor = ref('success')

// Env vars são gerenciadas como array de pares { key, value } para facilitar edição
const envPairs = ref([])

const emptyForm = () => ({
  name: '',
  description: '',
  transport: 'stdio',
  command: '',
  args: [],
  url: '',
  enabled: true,
})
const form = ref(emptyForm())

const canSave = computed(() => {
  if (!form.value.name) return false
  if (form.value.transport === 'stdio' && !form.value.command) return false
  if (form.value.transport === 'http' && !form.value.url) return false
  return true
})

onMounted(() => store.fetchAll())

function transportIcon(t) {
  return t === 'http' ? 'mdi-web' : 'mdi-console'
}

function envCount(srv) {
  return Object.keys(srv.env || {}).length
}

function openCreate() {
  editTarget.value = null
  form.value = emptyForm()
  envPairs.value = []
  formDialog.value = true
}

function openEdit(srv) {
  editTarget.value = srv
  form.value = {
    name:        srv.name,
    description: srv.description ?? '',
    transport:   srv.transport ?? 'stdio',
    command:     srv.command ?? '',
    args:        [...(srv.args ?? [])],
    url:         srv.url ?? '',
    enabled:     srv.enabled,
  }
  envPairs.value = Object.entries(srv.env ?? {}).map(([key, value]) => ({ key, value }))
  formDialog.value = true
}

function closeForm() {
  formDialog.value = false
}

function addEnvVar() {
  envPairs.value.push({ key: '', value: '' })
}

function removeEnvVar(i) {
  envPairs.value.splice(i, 1)
}

function buildEnvMap() {
  return Object.fromEntries(
    envPairs.value.filter((p) => p.key.trim()).map((p) => [p.key.trim(), p.value])
  )
}

async function submitForm() {
  saving.value = true
  try {
    const payload = {
      ...form.value,
      args: form.value.args ?? [],
      env: buildEnvMap(),
    }
    if (editTarget.value) {
      await store.update(editTarget.value.id, payload)
      notify('Servidor atualizado com sucesso.', 'success')
    } else {
      await store.create(payload)
      notify('Servidor criado com sucesso.', 'success')
    }
    formDialog.value = false
  } catch {
    notify(store.error || 'Erro ao salvar.', 'error')
  } finally {
    saving.value = false
  }
}

function openDelete(srv) {
  deleteTarget.value = srv
  deleteDialog.value = true
}

async function confirmDelete() {
  deleting.value = true
  try {
    await store.remove(deleteTarget.value.id)
    deleteDialog.value = false
    notify('Servidor removido.', 'info')
  } catch {
    notify(store.error || 'Erro ao excluir.', 'error')
  } finally {
    deleting.value = false
  }
}

async function toggle(srv) {
  toggling.value = srv.id
  try {
    await store.toggle(srv.id)
    const updated = store.servers.find((s) => s.id === srv.id)
    notify(`Servidor "${srv.name}" ${updated?.enabled ? 'ativado' : 'desativado'}.`, updated?.enabled ? 'success' : 'info')
  } catch {
    notify(store.error || 'Erro ao alterar status.', 'error')
  } finally {
    toggling.value = null
  }
}

function notify(msg, color = 'success') {
  snackbarMsg.value = msg
  snackbarColor.value = color
  snackbar.value = true
}
</script>

<style scoped>
.command-block {
  background: rgba(var(--v-theme-on-surface), 0.05);
  font-family: monospace;
  word-break: break-all;
}
</style>
