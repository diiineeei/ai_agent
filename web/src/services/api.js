import axios from 'axios'

const http = axios.create({ baseURL: '/api' })

export const chatAPI = {
  sendPrompt: (sessionId, prompt) =>
    http.post('/prompt', { session_id: sessionId, prompt }),
  getHistory: (sessionId) =>
    http.get('/history', { params: { session_id: sessionId } }),
  deleteHistory: (sessionId) =>
    http.delete('/history', { params: { session_id: sessionId } }),
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

export const settingsAPI = {
  get: () => http.get('/settings/system-instruction'),
  set: (value) => http.put('/settings/system-instruction', { value }),
}
