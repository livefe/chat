import { onUnmounted, reactive, readonly, ref, watch } from 'vue'

/** 新消息提醒：提示音、页面隐藏时的标题计数、桌面通知。偏好保存在 localStorage。 */

const PREFS_KEY = 'admin_attention_prefs'
const SOUND_GAP_MS = 1500
const TITLE_PREFIX = /^\(\d+\+?\) /

interface AttentionPrefs {
  sound: boolean
  desktop: boolean
}

function loadPrefs(): AttentionPrefs {
  try {
    const raw = JSON.parse(localStorage.getItem(PREFS_KEY) || '{}') as Partial<AttentionPrefs>
    return { sound: raw.sound !== false, desktop: raw.desktop === true }
  } catch {
    return { sound: true, desktop: false }
  }
}

const prefs = reactive<AttentionPrefs>(loadPrefs())

function savePrefs() {
  localStorage.setItem(PREFS_KEY, JSON.stringify(prefs))
}

export const attentionPrefs = readonly(prefs)

export function desktopNotifySupported() {
  return typeof window !== 'undefined' && 'Notification' in window
}

export function setAttentionSound(on: boolean) {
  prefs.sound = on
  savePrefs()
}

/** 开启桌面通知需在用户点击时调用，以便浏览器弹出授权。返回最终是否开启。 */
export async function setAttentionDesktop(on: boolean) {
  if (!on) {
    prefs.desktop = false
    savePrefs()
    return false
  }
  if (!desktopNotifySupported()) return false
  let permission = Notification.permission
  if (permission === 'default') permission = await Notification.requestPermission()
  prefs.desktop = permission === 'granted'
  savePrefs()
  return prefs.desktop
}

const visible = ref(typeof document === 'undefined' || document.visibilityState === 'visible')
const titleCounts = new Map<string, number>()

function applyTitle() {
  const base = document.title.replace(TITLE_PREFIX, '')
  let total = 0
  titleCounts.forEach((n) => (total += n))
  document.title = total > 0 && !visible.value ? `(${total > 99 ? '99+' : total}) ${base}` : base
}

if (typeof document !== 'undefined') {
  document.addEventListener('visibilitychange', () => {
    visible.value = document.visibilityState === 'visible'
    applyTitle()
  })
}

/** 当前标签页是否可见（共享状态）。 */
export function usePageVisible() {
  return readonly(visible)
}

/** 页面隐藏时在标签标题前显示 "(n) "；组件卸载时自动清除。 */
export function useTitleCount(key: string, count: () => number) {
  watch(
    count,
    (n) => {
      if (n > 0) titleCounts.set(key, n)
      else titleCounts.delete(key)
      applyTitle()
    },
    { immediate: true },
  )
  onUnmounted(() => {
    titleCounts.delete(key)
    applyTitle()
  })
}

let audioCtx: AudioContext | null = null
let lastSoundAt = 0

function playSound() {
  const now = Date.now()
  if (!prefs.sound || now - lastSoundAt < SOUND_GAP_MS) return
  lastSoundAt = now
  try {
    audioCtx ??= new AudioContext()
    const ctx = audioCtx
    if (ctx.state === 'suspended') void ctx.resume()
    const start = ctx.currentTime
    ;[880, 1320].forEach((freq, i) => {
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      const t = start + i * 0.14
      osc.type = 'sine'
      osc.frequency.value = freq
      gain.gain.setValueAtTime(0.0001, t)
      gain.gain.exponentialRampToValueAtTime(0.18, t + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.0001, t + 0.12)
      osc.connect(gain).connect(ctx.destination)
      osc.start(t)
      osc.stop(t + 0.13)
    })
  } catch {
    audioCtx = null
  }
}

function notifyDesktop(title: string, body: string, tag?: string) {
  if (!prefs.desktop || !desktopNotifySupported() || Notification.permission !== 'granted') return
  if (visible.value && document.hasFocus()) return
  try {
    const n = new Notification(title, { body, tag })
    n.onclick = () => {
      window.focus()
      n.close()
    }
  } catch {
    prefs.desktop = false
  }
}

/** 播放提示音并在页面不在前台时弹出桌面通知。 */
export function alertAttention(title: string, body: string, tag?: string) {
  playSound()
  notifyDesktop(title, body, tag)
}
