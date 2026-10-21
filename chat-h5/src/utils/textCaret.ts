export type TextField = HTMLInputElement | HTMLTextAreaElement

export interface CaretRect {
  left: number
  right: number
  top: number
  bottom: number
}

const SKIP_INPUT_TYPES = new Set([
  'button',
  'submit',
  'reset',
  'checkbox',
  'radio',
  'file',
  'image',
  'hidden',
  'range',
  'color',
  'date',
  'datetime-local',
  'month',
  'time',
  'week',
])

const MIRROR_STYLE = [
  'box-sizing',
  'width',
  'border-top-width',
  'border-right-width',
  'border-bottom-width',
  'border-left-width',
  'padding-top',
  'padding-right',
  'padding-bottom',
  'padding-left',
  'font',
  'letter-spacing',
  'text-indent',
  'text-transform',
  'word-spacing',
  'text-align',
  'line-height',
] as const

export function isTextField(el: EventTarget | null): el is TextField {
  if (el instanceof HTMLTextAreaElement) return !el.disabled
  if (!(el instanceof HTMLInputElement) || el.disabled) return false
  return !SKIP_INPUT_TYPES.has((el.type || 'text').toLowerCase())
}

export function caretRect(el: TextField, pressX: number, pressY: number): CaretRect {
  const box = el.getBoundingClientRect()
  const line = Math.min(22, Math.max(box.height, 16))
  const fallback = () => {
    const x = Math.min(Math.max(pressX, box.left), box.right)
    const y = Math.min(Math.max(pressY, box.top), box.bottom)
    return { left: x, right: x, top: y - line / 2, bottom: y + line / 2 }
  }
  if (el instanceof HTMLInputElement && el.type === 'password') return fallback()
  const start = el.selectionStart
  const end = el.selectionEnd
  if (start == null || end == null) return fallback()

  const cs = getComputedStyle(el)
  const mirror = document.createElement('div')
  const s = mirror.style
  s.position = 'fixed'
  s.left = `${box.left}px`
  s.top = `${box.top}px`
  s.height = `${box.height}px`
  s.visibility = 'hidden'
  s.pointerEvents = 'none'
  s.overflow = 'hidden'
  s.whiteSpace = el instanceof HTMLTextAreaElement ? 'pre-wrap' : 'pre'
  s.wordWrap = 'break-word'
  for (const name of MIRROR_STYLE) s.setProperty(name, cs.getPropertyValue(name))
  const mark = document.createElement('span')
  mark.textContent = start === end ? '\u200b' : el.value.slice(start, end) || '\u200b'
  mirror.append(el.value.slice(0, start), mark, el.value.slice(end))
  document.body.append(mirror)
  mirror.scrollTop = el.scrollTop
  mirror.scrollLeft = el.scrollLeft
  const m = mark.getBoundingClientRect()
  mirror.remove()

  const top = Math.min(Math.max(m.top, box.top), box.bottom)
  const bottom = Math.min(Math.max(m.bottom, box.top), box.bottom)
  const left = Math.min(Math.max(m.left, box.left), box.right)
  const right = Math.min(Math.max(m.right, box.left), box.right)
  if (bottom - top < 2) return fallback()
  return { left, right, top, bottom }
}
