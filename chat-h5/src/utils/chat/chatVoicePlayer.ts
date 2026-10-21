let shared: HTMLAudioElement | null = null
let currentKey = ''
const listeners = new Set<() => void>()

function ensureAudio(): HTMLAudioElement {
  if (!shared) {
    shared = new Audio()
    shared.preload = 'auto'
    shared.addEventListener('ended', () => {
      currentKey = ''
      notify()
    })
    shared.addEventListener('pause', () => {
      if (shared?.ended) return
    })
  }
  return shared
}

function notify() {
  listeners.forEach((fn) => fn())
}

export function subscribeVoicePlayer(fn: () => void): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

export function isVoicePlaying(key: string): boolean {
  return !!key && currentKey === key && !!shared && !shared.paused && !shared.ended
}

export function stopVoicePlayback() {
  if (!shared) return
  shared.pause()
  shared.removeAttribute('src')
  shared.load()
  currentKey = ''
  notify()
}

export async function toggleVoicePlayback(key: string, url: string): Promise<void> {
  const audio = ensureAudio()
  if (currentKey === key && !audio.paused) {
    stopVoicePlayback()
    return
  }
  stopVoicePlayback()
  currentKey = key
  audio.src = url
  notify()
  try {
    await audio.play()
    notify()
  } catch {
    currentKey = ''
    notify()
    throw new Error('play failed')
  }
}
