import { defineStore } from 'pinia'
import { ref } from 'vue'
import { agentConfigsAPI } from '@/services/api'

export const useAgentConfigsStore = defineStore('agent_configs', () => {
  const configs = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      const { data } = await agentConfigsAPI.list()
      configs.value = data ?? []
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function create(data) {
    const { data: created } = await agentConfigsAPI.create(data)
    configs.value.push(created)
    return created
  }

  async function update(id, data) {
    const { data: updated } = await agentConfigsAPI.update(id, data)
    const idx = configs.value.findIndex((c) => c.id === id)
    if (idx !== -1) configs.value[idx] = updated
    return updated
  }

  async function remove(id) {
    await agentConfigsAPI.delete(id)
    configs.value = configs.value.filter((c) => c.id !== id)
  }

  return { configs, loading, error, fetchAll, create, update, remove }
})
