export type PermSet = Set<string>

const KEY = 'perms-cache'

export function loadPerms(): PermSet {
  try {
    const raw = localStorage.getItem(KEY)
    if (!raw) return new Set()
    const arr = JSON.parse(raw)
    if (Array.isArray(arr)) return new Set(arr)
    return new Set()
  } catch {
    return new Set()
  }
}

export function savePerms(s: PermSet) {
  localStorage.setItem(KEY, JSON.stringify(Array.from(s)))
}

export function hasPerm(s: PermSet, required?: string) {
  if (!required) return true
  if (!s || s.size === 0) return false
  return s.has(required)
}
