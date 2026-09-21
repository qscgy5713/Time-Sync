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

function ownerKey(eventId) {
  return `timesync:owner:${eventId}`
}

// The organizer's secret for an event; only the browser that created the
// event holds it, and it is required to finalize/unfinalize.
export function getOwnerToken(eventId) {
  try {
    return localStorage.getItem(ownerKey(eventId))
  } catch {
    return null
  }
}

export function storeOwnerToken(eventId, token) {
  try {
    localStorage.setItem(ownerKey(eventId), token)
  } catch {
    // localStorage unavailable; silently ignore
  }
}
