import { onMounted, onUnmounted } from 'vue'

// Closes an open popover/dropdown when the user clicks outside `rootEl` or presses Escape.
export function useDismissable(rootEl, onDismiss) {
  function onClickOutside(e) {
    if (rootEl.value && !rootEl.value.contains(e.target)) onDismiss()
  }
  function onKeydown(e) {
    if (e.key === 'Escape') onDismiss()
  }

  onMounted(() => {
    window.addEventListener('mousedown', onClickOutside)
    window.addEventListener('keydown', onKeydown)
  })
  onUnmounted(() => {
    window.removeEventListener('mousedown', onClickOutside)
    window.removeEventListener('keydown', onKeydown)
  })
}
