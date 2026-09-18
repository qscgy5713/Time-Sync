export const SLOT_MINUTES = 30

function pad2(n) {
  return String(n).padStart(2, '0')
}

export function minutesToHHmm(totalMinutes) {
  const h = Math.floor(totalMinutes / 60)
  const m = totalMinutes % 60
  return `${pad2(h)}:${pad2(m)}`
}
