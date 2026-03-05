<template>
  <v-container class="pa-6" style="max-width: 960px">
    <div class="text-h5 font-weight-bold mb-6 d-flex align-center gap-2">
      <v-icon color="primary">mdi-file-multiple</v-icon>
      Arquivos
    </div>

    <!-- Upload card -->
    <v-card rounded="lg" variant="outlined" class="mb-6">
      <v-card-text class="pa-4">
        <div
          class="upload-zone d-flex flex-column align-center justify-center"
          :class="{ 'drag-over': isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onDrop"
          @click="fileInputEl?.click()"
        >
          <v-icon
            size="52"
            :color="isDragging ? 'primary' : 'medium-emphasis'"
            class="mb-3"
          >
            {{ isDragging ? 'mdi-cloud-download-outline' : 'mdi-cloud-upload-outline' }}
          </v-icon>
          <p class="text-body-1 font-weight-medium mb-1">
            {{ isDragging ? 'Solte para enviar' : 'Arraste um arquivo ou clique para selecionar' }}
          </p>
          <p class="text-caption text-medium-emphasis mb-0">
            Suportado: PDF, TXT, Markdown
          </p>
          <input
            ref="fileInputEl"
            type="file"
            accept=".pdf,.txt,.md,text/plain,application/pdf,text/markdown"
            style="display: none"
            @change="onFileSelected"
          />
        </div>

        <v-progress-linear
          v-if="store.uploading"
          indeterminate
          color="primary"
          class="mt-3"
          rounded
        />
      </v-card-text>
    </v-card>

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

    <!-- Files table -->
    <v-card rounded="lg" variant="outlined">
      <div class="d-flex align-center px-4 py-3">
        <span class="text-subtitle-1 font-weight-medium">Documentos enviados</span>
        <v-spacer />
        <v-btn
          icon="mdi-refresh"
          variant="text"
          size="small"
          :loading="store.loading"
          @click="store.fetchFiles()"
        />
      </div>

      <v-divider />

      <v-data-table
        :headers="headers"
        :items="store.files"
        :loading="store.loading"
        no-data-text="Nenhum arquivo enviado ainda."
        loading-text="Carregando arquivos…"
        items-per-page="10"
        density="comfortable"
      >
        <template #item.content_type="{ value }">
          <v-chip size="small" :color="mimeColor(value)" variant="tonal">
            <v-icon start size="14">{{ mimeIcon(value) }}</v-icon>
            {{ mimeLabel(value) }}
          </v-chip>
        </template>

        <template #item.size="{ value }">
          {{ formatSize(value) }}
        </template>

        <template #item.uploaded_at="{ value }">
          {{ formatDate(value) }}
        </template>

        <template #item.actions="{ item }">
          <v-btn
            icon="mdi-delete-outline"
            size="small"
            variant="text"
            color="error"
            @click="confirmDelete(item)"
          />
        </template>
      </v-data-table>
    </v-card>
  </v-container>

  <!-- Delete confirm dialog -->
  <v-dialog v-model="deleteDialog" max-width="400">
    <v-card rounded="lg">
      <v-card-title class="pt-4">Remover arquivo</v-card-title>
      <v-card-text>
        Tem certeza que deseja remover
        <strong>{{ fileToDelete?.name }}</strong>?
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="deleteDialog = false">Cancelar</v-btn>
        <v-btn color="error" @click="doDelete">Remover</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Snackbar -->
  <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000">
    {{ snackbarMsg }}
  </v-snackbar>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFilesStore } from '@/stores/files'

const store = useFilesStore()

const fileInputEl = ref(null)
const isDragging = ref(false)
const deleteDialog = ref(false)
const fileToDelete = ref(null)
const snackbar = ref(false)
const snackbarMsg = ref('')
const snackbarColor = ref('success')

const headers = [
  { title: 'Nome', key: 'name', sortable: true },
  { title: 'Tipo', key: 'content_type', sortable: false },
  { title: 'Tamanho', key: 'size', sortable: true },
  { title: 'Enviado em', key: 'uploaded_at', sortable: true },
  { title: '', key: 'actions', sortable: false, align: 'end' },
]

onMounted(() => store.fetchFiles())

function onFileSelected(e) {
  const file = e.target.files[0]
  if (file) upload(file)
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files[0]
  if (file) upload(file)
}

async function upload(file) {
  try {
    await store.uploadFile(file)
    showSnackbar(`"${file.name}" enviado com sucesso!`, 'success')
  } catch {
    showSnackbar(store.error || 'Erro ao enviar arquivo.', 'error')
  }
}

function confirmDelete(item) {
  fileToDelete.value = item
  deleteDialog.value = true
}

async function doDelete() {
  try {
    await store.deleteFile(fileToDelete.value.id)
    showSnackbar('Arquivo removido.', 'success')
  } catch {
    showSnackbar(store.error || 'Erro ao remover arquivo.', 'error')
  }
  deleteDialog.value = false
}

function showSnackbar(msg, color = 'success') {
  snackbarMsg.value = msg
  snackbarColor.value = color
  snackbar.value = true
}

function formatSize(bytes) {
  if (!bytes) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(v) {
  if (!v) return '—'
  return new Date(v).toLocaleString('pt-BR')
}

function mimeColor(type) {
  if (type === 'application/pdf') return 'red'
  if (type?.startsWith('text/')) return 'blue'
  return 'grey'
}

function mimeIcon(type) {
  if (type === 'application/pdf') return 'mdi-file-pdf-box'
  if (type?.startsWith('text/')) return 'mdi-file-document-outline'
  return 'mdi-file-outline'
}

function mimeLabel(type) {
  const labels = {
    'application/pdf': 'PDF',
    'text/plain': 'TXT',
    'text/markdown': 'MD',
  }
  return labels[type] || type
}
</script>

<style scoped>
.upload-zone {
  border: 2px dashed rgba(var(--v-border-color), 0.4);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  min-height: 160px;
  padding: 32px 16px;
}

.upload-zone:hover {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.04);
}

.upload-zone.drag-over {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.08);
}
</style>
