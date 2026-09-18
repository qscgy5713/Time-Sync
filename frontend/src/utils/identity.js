function storageKey(eventId) {
  return `timesync:participant:${eventId}`
}

export function getStoredParticipant(eventId) {
  try {
    const raw = localStorage.getItem(storageKey(eventId))
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export function storeParticipant(eventId, id, editToken) {
  try {
    localStorage.setItem(storageKey(eventId), JSON.stringify({ id, editToken }))
  } catch {
    // localStorage unavailable (private mode, quota, etc.); silently ignore
  }
}

export function clearStoredParticipant(eventId) {
  try {
    localStorage.removeItem(storageKey(eventId))
  } catch {
    // ignore
  }
}
