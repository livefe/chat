import { ref } from 'vue'

const LONG_PRESS_MS = 450

const MOVE_CANCEL_PX = 24

export function useLongPress(onLongPress: () => void, delay = LONG_PRESS_MS) {
  const timer = ref<ReturnType<typeof setTimeout> | null>(null)
  let startX = 0
  let startY = 0
  let fired = false

  function clear() {
    if (timer.value !== null) {
      clearTimeout(timer.value)
      timer.value = null
    }
  }

  function onTouchStart(e: TouchEvent) {
    clear()
    fired = false
    const t = e.touches[0]
    startX = t?.clientX ?? 0
    startY = t?.clientY ?? 0
    timer.value = setTimeout(() => {
      timer.value = null
      fired = true
      onLongPress()
    }, delay)
  }

  function onTouchEnd() {
    clear()
  }

  function onTouchMove(e: TouchEvent) {
    if (fired) {
      e.preventDefault()
      return
    }
    const t = e.touches[0]
    if (!t) {
      clear()
      return
    }
    const dx = t.clientX - startX
    const dy = t.clientY - startY
    if (dx * dx + dy * dy > MOVE_CANCEL_PX * MOVE_CANCEL_PX) clear()
  }

  function consumeClickIfLongPressed(e?: Event) {
    if (!fired) return false
    fired = false
    e?.preventDefault()
    e?.stopPropagation()
    return true
  }

  return { onTouchStart, onTouchEnd, onTouchMove, consumeClickIfLongPressed }
}
