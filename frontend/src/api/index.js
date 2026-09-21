// Nullish coalescing (not ||): an explicitly empty string means "same origin,
// use relative paths" (production behind a single reverse proxy), which must
// not fall back to the localhost default.
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

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
  getStats() {
    return request('/api/stats')
  },
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
  finalizeEvent(id, optionId, ownerToken) {
    return request(`/api/events/${id}/finalize`, {
      method: 'PUT',
      body: JSON.stringify({ option_id: optionId, owner_token: ownerToken }),
    })
  },
}
