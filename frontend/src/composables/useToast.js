import { reactive, readonly } from 'vue'

/**
 * Module-level reactive state — shared singleton across all callers.
 * Import useToast() anywhere to read or mutate the same queue.
 */
const _state = reactive({ toasts: [] })
let _nextId  = 0

/**
 * Add a toast to the queue. Returns the assigned numeric id.
 * @param {{ type: string, message?: string, title?: string, duration?: number }} opts
 */
const _add = (opts) => {
  const id       = ++_nextId
  const duration = opts.duration ?? 5000
  _state.toasts.push({ id, ...opts, duration })
  if (duration > 0) setTimeout(() => _remove(id), duration)
  return id
}

const _remove = (id) => {
  const idx = _state.toasts.findIndex(t => t.id === id)
  if (idx !== -1) _state.toasts.splice(idx, 1)
}

/**
 * Composable — call anywhere in a component or another composable.
 *
 * @example
 * const toast = useToast()
 * toast.success('Salon créé !')
 * toast.error('Erreur réseau', { title: 'Connexion impossible' })
 * toast.xp({ xpGained: 50, newXP: 350, newLevel: 4, levelUp: false })
 */
export const useToast = () => ({
  /** Read-only list of active toasts */
  toasts: readonly(_state.toasts),

  success: (message, opts = {}) => _add({ type: 'success', message, ...opts }),
  error:   (message, opts = {}) => _add({ type: 'error',   message, ...opts }),
  info:    (message, opts = {}) => _add({ type: 'info',    message, ...opts }),
  warning: (message, opts = {}) => _add({ type: 'warning', message, ...opts }),

  /**
   * XP gain toast (maps to the existing XP_GAINED WebSocket event).
   * @param {{ xpGained: number, newXP: number, newLevel: number, levelUp: boolean }} data
   */
  xp: (data) => _add({ type: 'xp', ...data, duration: 5000 }),

  /** Manually dismiss a toast before its timer expires */
  dismiss: _remove,
})
