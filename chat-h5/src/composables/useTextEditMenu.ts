import { nextTick, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { showToast } from 'vant'

import { copyText } from '@/utils/share/shareCore'
import { caretRect, isTextField, type TextField } from '@/utils/textCaret'

export interface TextEditAction {
  name: string
  run: () => void | Promise<void>
}

const LONG_PRESS_MS = 450
const MOVE_CANCEL_PX = 24

function viewBox() {
  const vv = window.visualViewport
  if (!vv) {
    return { left: 0, top: 0, width: window.innerWidth, height: window.innerHeight }
  }
  return { left: vv.offsetLeft, top: vv.offsetTop, width: vv.width, height: vv.height }
}

function selectedText(el: TextField): string {
  return el.value.slice(el.selectionStart ?? 0, el.selectionEnd ?? 0)
}

function replaceSelection(el: TextField, text: string) {
  const start = el.selectionStart ?? 0
  const end = el.selectionEnd ?? 0
  el.focus()
  el.setRangeText(text, start, end, 'end')
  el.dispatchEvent(new Event('input', { bubbles: true }))
}

export function useTextEditMenu(menuRef: Ref<HTMLElement | null>) {
  const route = useRoute()
  const open = ref(false)
  const menuStyle = ref<Record<string, string>>({})
  const actions = ref<TextEditAction[]>([])

  let field: TextField | null = null
  let canDismiss = false
  const heldPointers = new Set<number>()
  let pressFrame = 0
  let pressStartedAt = 0
  let pressX = 0
  let pressY = 0
  let startX = 0
  let startY = 0
  let lastX = 0
  let lastY = 0
  let longPressed = false
  let holdField: TextField | null = null
  let holdFocused = false
  let holdMoved = false
  let holdReadOnly = false

  function clearPressTimer() {
    if (!pressFrame) return
    cancelAnimationFrame(pressFrame)
    pressFrame = 0
  }

  function close() {
    open.value = false
    field = null
    canDismiss = false
    actions.value = []
  }

  function placeMenu() {
    const el = field
    if (!el) return
    const a =
      document.activeElement === el
        ? caretRect(el, pressX, pressY)
        : { left: pressX, right: pressX, top: pressY - 11, bottom: pressY + 11 }
    const view = viewBox()
    const margin = 8
    const gap = 8
    const menu = menuRef.value
    const menuW = menu?.offsetWidth || 160
    const menuH = menu?.offsetHeight || 40
    const anchorX = (a.left + a.right) / 2

    const minLeft = view.left + margin
    const maxLeft = view.left + view.width - menuW - margin
    const left =
      maxLeft < minLeft ? minLeft : Math.min(Math.max(anchorX - menuW / 2, minLeft), maxLeft)

    const yAbove = a.top - menuH - gap
    const yBelow = a.bottom + gap
    const placeBelow =
      yAbove < view.top + margin && yBelow + menuH <= view.top + view.height - margin
    const minTop = view.top + margin
    const maxTop = view.top + view.height - menuH - margin
    const top =
      maxTop < minTop ? minTop : Math.min(Math.max(placeBelow ? yBelow : yAbove, minTop), maxTop)

    menuStyle.value = {
      left: `${Math.round(left)}px`,
      top: `${Math.round(top)}px`,
    }
  }

  async function doCopy() {
    const el = field
    if (!el) return
    const text = selectedText(el)
    if (!text) return
    try {
      await copyText(text)
      showToast('已复制')
    } catch {
      showToast('复制失败')
    }
  }

  async function doCut() {
    const el = field
    if (!el || el.readOnly) return
    const text = selectedText(el)
    if (!text) return
    try {
      await copyText(text)
      replaceSelection(el, '')
    } catch {
      showToast('剪切失败')
    }
  }

  async function doPaste() {
    const el = field
    if (!el || el.readOnly) return
    try {
      const text = navigator.clipboard?.readText ? await navigator.clipboard.readText() : ''
      if (!text) {
        showToast('剪贴板为空')
        return
      }
      replaceSelection(el, text)
    } catch {
      showToast('无法粘贴')
    }
  }

  function doSelectAll() {
    const el = field
    if (!el) return
    el.focus()
    el.setSelectionRange(0, el.value.length)
  }

  function openMenu(el: TextField, clientX: number, clientY: number, holding = false) {
    field = el
    const hasSel = (el.selectionStart ?? 0) !== (el.selectionEnd ?? 0)
    const readonly = el.readOnly
    const password = el instanceof HTMLInputElement && el.type === 'password'
    const next: TextEditAction[] = []
    if (!readonly) next.push({ name: '粘贴', run: doPaste })
    if (hasSel && !password) {
      next.push({ name: '复制', run: doCopy })
      if (!readonly) next.push({ name: '剪切', run: doCut })
    }
    if (el.value.length > 0) next.push({ name: '全选', run: doSelectAll })
    if (!next.length) {
      close()
      return
    }
    actions.value = next
    pressX = clientX
    pressY = clientY
    open.value = true
    canDismiss = !holding && heldPointers.size === 0
    placeMenu()
    void nextTick(() => placeMenu())
  }

  async function runAction(action: TextEditAction) {
    open.value = false
    canDismiss = false
    actions.value = []
    try {
      await action.run()
    } finally {
      field = null
    }
  }

  function onContextMenu(e: Event) {
    const el = e.target
    if (!isTextField(el)) return
    e.preventDefault()
    e.stopPropagation()
    openMenu(el, (e as MouseEvent).clientX, (e as MouseEvent).clientY)
  }

  function onPointerDown(e: PointerEvent) {
    heldPointers.add(e.pointerId)
    if (!open.value || !canDismiss) return
    const node = e.target as Node | null
    if (menuRef.value && node && menuRef.value.contains(node)) return
    close()
  }

  function onPointerUp(e: PointerEvent) {
    heldPointers.delete(e.pointerId)
    if (open.value && heldPointers.size === 0) canDismiss = true
  }

  function releaseHoldReadOnly(keepFocus: boolean) {
    const el = holdField
    if (!holdReadOnly || !el) return
    if (!keepFocus && document.activeElement === el) el.blur()
    el.readOnly = false
    holdReadOnly = false
  }

  function finishHold(focus: boolean) {
    const el = holdField
    const pressedLong = longPressed
    const shouldFocus =
      focus && !!el && !holdFocused && !pressedLong && !holdMoved && (!el.readOnly || holdReadOnly)
    longPressed = false
    releaseHoldReadOnly(shouldFocus)
    clearPressTimer()
    holdField = null
    if (open.value) canDismiss = heldPointers.size === 0
    if (shouldFocus) el.focus()
  }

  function onTouchStart(e: TouchEvent) {
    clearPressTimer()
    const el = e.target
    if (!isTextField(el)) return
    longPressed = false
    holdMoved = false
    holdField = el
    holdFocused = document.activeElement === el
    holdReadOnly = false
    if (!holdFocused && !el.readOnly) {
      e.preventDefault()
      el.readOnly = true
      holdReadOnly = true
    }
    const t = e.touches[0]
    startX = lastX = t?.clientX ?? 0
    startY = lastY = t?.clientY ?? 0
    pressStartedAt = performance.now()
    const tick = (now: number) => {
      if (holdField !== el || holdMoved) {
        pressFrame = 0
        return
      }
      if (now - pressStartedAt >= LONG_PRESS_MS) {
        pressFrame = 0
        longPressed = true
        releaseHoldReadOnly(false)
        openMenu(el, lastX, lastY, true)
        return
      }
      pressFrame = requestAnimationFrame(tick)
    }
    pressFrame = requestAnimationFrame(tick)
  }

  function onTouchMove(e: TouchEvent) {
    if (longPressed) {
      e.preventDefault()
      return
    }
    const t = e.touches[0]
    if (!t) {
      clearPressTimer()
      return
    }
    lastX = t.clientX
    lastY = t.clientY
    const dx = lastX - startX
    const dy = lastY - startY
    if (dx * dx + dy * dy > MOVE_CANCEL_PX * MOVE_CANCEL_PX) {
      holdMoved = true
      clearPressTimer()
    }
  }

  function onTouchEnd() {
    finishHold(true)
  }

  function onTouchCancel() {
    finishHold(false)
  }

  function onScroll(e: Event) {
    if (!open.value || holdField) return
    const node = e.target
    if (node instanceof Node && field && (node === field || field.contains(node))) return
    close()
  }

  function onFieldInput() {
    if (open.value) close()
  }

  function onFocusIn(e: Event) {
    if (!open.value || holdField) return
    const node = e.target as Node | null
    if (node === field) return
    if (menuRef.value && node && menuRef.value.contains(node)) return
    close()
  }

  watch(() => route.fullPath, close)

  watch(open, (isOpen) => {
    if (!isOpen) return
    const update = () => placeMenu()
    window.visualViewport?.addEventListener('resize', update)
    window.visualViewport?.addEventListener('scroll', update)
    window.addEventListener('resize', update)
    void nextTick(update)
    return () => {
      window.visualViewport?.removeEventListener('resize', update)
      window.visualViewport?.removeEventListener('scroll', update)
      window.removeEventListener('resize', update)
    }
  })

  onMounted(() => {
    document.addEventListener('contextmenu', onContextMenu, true)
    document.addEventListener('pointerdown', onPointerDown, true)
    document.addEventListener('pointerup', onPointerUp, true)
    document.addEventListener('pointercancel', onPointerUp, true)
    document.addEventListener('touchstart', onTouchStart, { capture: true, passive: false })
    document.addEventListener('touchmove', onTouchMove, { capture: true, passive: false })
    document.addEventListener('touchend', onTouchEnd, true)
    document.addEventListener('touchcancel', onTouchCancel, true)
    window.addEventListener('scroll', onScroll, true)
    document.addEventListener('input', onFieldInput, true)
    document.addEventListener('focusin', onFocusIn, true)
    document.addEventListener('visibilitychange', close)
  })

  onUnmounted(() => {
    clearPressTimer()
    document.removeEventListener('contextmenu', onContextMenu, true)
    document.removeEventListener('pointerdown', onPointerDown, true)
    document.removeEventListener('pointerup', onPointerUp, true)
    document.removeEventListener('pointercancel', onPointerUp, true)
    document.removeEventListener('touchstart', onTouchStart, true)
    document.removeEventListener('touchmove', onTouchMove, true)
    document.removeEventListener('touchend', onTouchEnd, true)
    document.removeEventListener('touchcancel', onTouchCancel, true)
    window.removeEventListener('scroll', onScroll, true)
    document.removeEventListener('input', onFieldInput, true)
    document.removeEventListener('focusin', onFocusIn, true)
    document.removeEventListener('visibilitychange', close)
    close()
  })

  return { open, menuStyle, actions, runAction }
}
