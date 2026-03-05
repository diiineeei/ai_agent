import { defineStore } from 'pinia'
import { ref } from 'vue'
import { filesAPI } from '@/services/api'

export const useFilesStore = defineStore('files', () => {
  const files = ref([])
  const loading = ref(false)
  const uploading = ref(false)
  const error = ref(null)

  async function fetchFiles() {
    loading.value = true
    error.value = null
    try {
      const { data } = await filesAPI.list()
      files.value = data || []
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function uploadFile(file) {
    uploading.value = true
    error.value = null
    try {
      const { data } = await filesAPI.upload(file)
      await fetchFiles()
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    } finally {
      uploading.value = false
    }
  }

  async function deleteFile(id) {
    error.value = null
    try {
      await filesAPI.delete(id)
      files.value = files.value.filter((f) => f.id !== id)
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  return { files, loading, uploading, error, fetchFiles, uploadFile, deleteFile }
})
