const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

async function request(path, options = {}) {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    throw new Error(data?.error || `請求失敗 (HTTP ${res.status})`)
  }
  return data
}

export const api = {
  createEvent(payload) {
    return request('/api/events', { method: 'POST', body: JSON.stringify(payload) })
  },
  getEvent(id) {
    return request(`/api/events/${id}`)
  },
  addParticipant(id, payload) {
    return request(`/api/events/${id}/participants`, { method: 'POST', body: JSON.stringify(payload) })
  },
  updateParticipant(id, participantId, payload) {
    return request(`/api/events/${id}/participants/${participantId}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
}
