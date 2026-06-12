/**
 * Stable deep equality comparison.
 *
 * Unlike JSON.stringify(objA) !== JSON.stringify(objB), this function
 * sorts object keys before comparison, so insertion-order differences
 * do not cause false negatives.
 *
 * @param {*} a
 * @param {*} b
 * @returns {boolean} true if a and b are deeply equal (key-order independent)
 */
export function deepEqual(a, b) {
  // Same reference or both null/undefined
  if (a === b) return true

  // One is null/undefined but not both
  if (a == null || b == null) return false

  // Both must be objects (not primitives)
  if (typeof a !== 'object' || typeof b !== 'object') return false

  // Both must be arrays or both plain objects
  const aIsArray = Array.isArray(a)
  const bIsArray = Array.isArray(b)
  if (aIsArray !== bIsArray) return false

  if (aIsArray) {
    if (a.length !== b.length) return false
    for (let i = 0; i < a.length; i++) {
      if (!deepEqual(a[i], b[i])) return false
    }
    return true
  }

  // Plain objects: compare sorted keys
  const keysA = Object.keys(a).sort()
  const keysB = Object.keys(b).sort()
  if (keysA.length !== keysB.length) return false
  for (let i = 0; i < keysA.length; i++) {
    if (keysA[i] !== keysB[i]) return false
    if (!deepEqual(a[keysA[i]], b[keysB[i]])) return false
  }
  return true
}

/**
 * Check if a config group has unpublished draft changes.
 * @param {object} config — must have draftData and publishedData properties
 * @returns {boolean}
 */
export function hasDraft(config) {
  if (!config) return false
  return !deepEqual(config.draftData || {}, config.publishedData || {})
}

/**
 * Count the number of diff entries between draft and published.
 * Returns { added, modified, deleted, total }
 * @param {object} config — must have draftData and publishedData properties
 * @returns {{ added: number, modified: number, deleted: number, total: number }}
 */
export function countDiff(config) {
  if (!config) return { added: 0, modified: 0, deleted: 0, total: 0 }
  const pub = config.publishedData || {}
  const draft = config.draftData || {}
  const allKeys = new Set([...Object.keys(pub), ...Object.keys(draft)])

  let added = 0, modified = 0, deleted = 0
  for (const key of allKeys) {
    const inPub = key in pub
    const inDraft = key in draft
    if (!inPub && inDraft) added++
    else if (inPub && !inDraft) deleted++
    else if (pub[key] !== draft[key]) modified++
  }
  return { added, modified, deleted, total: added + modified + deleted }
}
