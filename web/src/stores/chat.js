import { defineStore } from 'pinia'
import { ref } from 'vue'
import { chatAPI } from '@/services/api'

function genSessionId() {
  return 'session-' + Math.random().toString(36).substring(2, 10)
}

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const sessionId = ref(genSessionId())
  const agentConfigId = ref(null)
  const agentName = ref(null)
  const sessionName = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // name to be saved on the first message (set when creating a new named session)
  const _pendingName = ref(null)

  async function send(prompt) {
    error.value = null
    messages.value.push({ role: 'user', text: prompt })
    loading.value = true
    try {
      const { data } = await chatAPI.sendPrompt(sessionId.value, prompt, agentConfigId.value)
      messages.value.push({ role: 'model', text: data.response, tokenUsage: data.token_usage ?? null })
      if (data.agent_name) agentName.value = data.agent_name

      // Persist pending name after first message creates the session doc
      if (_pendingName.value) {
        const name = _pendingName.value
        _pendingName.value = null
        sessionName.value = name
        chatAPI.renameSession(sessionId.value, name).catch(() => {})
      }
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

  function newSession(configId, name = null) {
    sessionId.value = genSessionId()
    messages.value = []
    error.value = null
    agentConfigId.value = configId ?? null
    agentName.value = null
    sessionName.value = name || null
    _pendingName.value = name || null
  }

  function setSession(id, name = null) {
    if (!id.trim()) return
    sessionId.value = id.trim()
    messages.value = []
    error.value = null
    sessionName.value = name || null
    _pendingName.value = null
  }

  async function renameCurrentSession(name) {
    await chatAPI.renameSession(sessionId.value, name)
    sessionName.value = name || null
  }

  return {
    messages,
    sessionId,
    agentConfigId,
    agentName,
    sessionName,
    loading,
    error,
    send,
    loadHistory,
    clearHistory,
    newSession,
    setSession,
    renameCurrentSession,
  }
})
