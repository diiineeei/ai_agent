import { defineStore } from 'pinia'
import { ref } from 'vue'
import { mcpServersAPI } from '@/services/api'

export const useMcpServersStore = defineStore('mcp_servers', () => {
  const servers = ref([])
  const loading = ref(false)
  const error   = ref(null)

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      const { data } = await mcpServersAPI.list()
      servers.value = data || []
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function create(payload) {
    error.value = null
    try {
      const { data } = await mcpServersAPI.create(payload)
      servers.value.push(data)
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  async function update(id, payload) {
    error.value = null
    try {
      const { data } = await mcpServersAPI.update(id, payload)
      const idx = servers.value.findIndex((s) => s.id === id)
      if (idx !== -1) servers.value[idx] = data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  async function remove(id) {
    error.value = null
    try {
      await mcpServersAPI.delete(id)
      servers.value = servers.value.filter((s) => s.id !== id)
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  async function toggle(id) {
    error.value = null
    try {
      const { data } = await mcpServersAPI.toggle(id)
      const idx = servers.value.findIndex((s) => s.id === id)
      if (idx !== -1) servers.value[idx] = data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    }
  }

  return { servers, loading, error, fetchAll, create, update, remove, toggle }
})
