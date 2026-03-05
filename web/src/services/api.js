import axios from 'axios'

const http = axios.create({ baseURL: '/api' })

export const chatAPI = {
  sendPrompt: (sessionId, prompt, agentConfigId = null) => {
    const body = { session_id: sessionId, prompt }
    if (agentConfigId) body.agent_config_id = agentConfigId
    return http.post('/prompt', body)
  },
  getHistory: (sessionId) =>
    http.get('/history', { params: { session_id: sessionId } }),
  deleteHistory: (sessionId) =>
    http.delete('/history', { params: { session_id: sessionId } }),
  listSessions: () => http.get('/sessions'),
  renameSession: (sessionId, name) =>
    http.put(`/sessions/${sessionId}/name`, { name }),
}

export const filesAPI = {
  upload: (file) => {
    const form = new FormData()
    form.append('file', file)
    return http.post('/files', form)
  },
  list: () => http.get('/files'),
  delete: (id) => http.delete(`/files/${id}`),
}

export const skillsAPI = {
  list: () => http.get('/skills'),
  toggle: (name) => http.put(`/skills/${name}/toggle`),
}

export const agentConfigsAPI = {
  list: () => http.get('/agent-configs'),
  getById: (id) => http.get(`/agent-configs/${id}`),
  create: (data) => http.post('/agent-configs', data),
  update: (id, data) => http.put(`/agent-configs/${id}`, data),
  delete: (id) => http.delete(`/agent-configs/${id}`),
  improveInstruction: (model, instruction) =>
    http.post('/agent-configs/improve-instruction', { model, instruction }),
}
