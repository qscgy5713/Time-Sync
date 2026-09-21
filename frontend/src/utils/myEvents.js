const KEY = 'timesync:my_events'
const MAX_ENTRIES = 50

export function getMyEvents() {
  try {
    const raw = localStorage.getItem(KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

// role: 'organizer' | 'participant'. An event you organized keeps that
// label even if you later also fill in your own availability on it.
export function recordMyEvent(id, title, role) {
  try {
    const list = getMyEvents()
    const existing = list.find((e) => e.id === id)
    const finalRole = existing?.role === 'organizer' ? 'organizer' : role
    const rest = list.filter((e) => e.id !== id)
    rest.unshift({ id, title, role: finalRole, visitedAt: Date.now() })
    localStorage.setItem(KEY, JSON.stringify(rest.slice(0, MAX_ENTRIES)))
  } catch {
    // localStorage unavailable; silently ignore
  }
}

export function removeMyEvent(id) {
  try {
    localStorage.setItem(KEY, JSON.stringify(getMyEvents().filter((e) => e.id !== id)))
  } catch {
    // ignore
  }
}
