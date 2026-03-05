import { defineStore } from 'pinia'
import { ref } from 'vue'
import { settingsAPI } from '@/services/api'

export const useSettingsStore = defineStore('settings', () => {
  const instruction = ref('')
  const loading = ref(false)
  const saving = ref(false)
  const error = ref(null)

  async function fetch() {
    loading.value = true
    error.value = null
    try {
      const res = await settingsAPI.get()
      instruction.value = res.data.value
    } catch (e) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function save(value) {
    saving.value = true
    error.value = null
    try {
      const res = await settingsAPI.set(value)
      instruction.value = res.data.value
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      saving.value = false
    }
  }

  return { instruction, loading, saving, error, fetch, save }
})
