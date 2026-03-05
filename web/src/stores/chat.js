import { defineStore } from 'pinia'
import { ref } from 'vue'
import { chatAPI } from '@/services/api'

function genSessionId() {
  return 'session-' + Math.random().toString(36).substring(2, 10)
}

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const sessionId = ref(genSessionId())
  const loading = ref(false)
  const error = ref(null)

  async function send(prompt) {
    error.value = null
    messages.value.push({ role: 'user', text: prompt })
    loading.value = true
    try {
      const { data } = await chatAPI.sendPrompt(sessionId.value, prompt)
      messages.value.push({ role: 'model', text: data.response })
    } catch (e) {
      messages.value.pop()
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function loadHistory() {
    error.value = null
    try {
      const { data } = await chatAPI.getHistory(sessionId.value)
      if (!data || data.length === 0) {
        messages.value = []
        return
      }
      messages.value = data
        .map((c) => ({
          role: c.role,
          text: c.parts?.find((p) => p.text)?.text || '',
        }))
        .filter((m) => m.text)
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    }
  }

  async function clearHistory() {
    await chatAPI.deleteHistory(sessionId.value)
    messages.value = []
    error.value = null
  }

  function newSession() {
    sessionId.value = genSessionId()
    messages.value = []
    error.value = null
  }

  function setSession(id) {
    if (!id.trim()) return
    sessionId.value = id.trim()
    messages.value = []
    error.value = null
  }

  return { messages, sessionId, loading, error, send, loadHistory, clearHistory, newSession, setSession }
})
