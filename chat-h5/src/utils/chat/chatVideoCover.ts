export function formatVideoDuration(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return ''
  const total = Math.floor(seconds)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

export function fitChatVideoEl(el: HTMLVideoElement, maxW = 220, maxH = 220) {
  const w = el.videoWidth
  const h = el.videoHeight
  if (!(w > 0 && h > 0)) return
  const widthCap = Math.min(maxW, Math.floor(window.innerWidth * 0.6))
  const scale = Math.min(widthCap / w, maxH / h, 1)
  el.style.width = `${Math.max(1, Math.round(w * scale))}px`
  el.style.height = `${Math.max(1, Math.round(h * scale))}px`
}

function waitEvent(el: HTMLMediaElement, type: string, timeoutMs = 2500): Promise<void> {
  return new Promise((resolve) => {
    let done = false
    const finish = () => {
      if (done) return
      done = true
      el.removeEventListener(type, finish)
      window.clearTimeout(timer)
      resolve()
    }
    const timer = window.setTimeout(finish, timeoutMs)
    el.addEventListener(type, finish)
  })
}

export async function unlockVideoCoverFrame(el: HTMLVideoElement): Promise<string> {
  el.muted = true
  el.defaultMuted = true
  el.playsInline = true
  el.setAttribute('muted', '')
  el.setAttribute('playsinline', '')
  el.setAttribute('webkit-playsinline', '')
  el.setAttribute('x5-playsinline', 'true')
  el.setAttribute('x5-video-player-type', 'h5')
  el.preload = 'auto'

  if (el.readyState < 2) {
    try {
      el.load()
    } catch {}
    await Promise.race([waitEvent(el, 'loadeddata'), waitEvent(el, 'canplay')])
  }

  try {
    const playResult = el.play()
    if (playResult && typeof playResult.then === 'function') {
      await Promise.race([playResult, waitEvent(el, 'playing', 1500)])
    }
  } catch {}
  try {
    el.pause()
  } catch {}

  const dur = Number.isFinite(el.duration) && el.duration > 0 ? el.duration : 0
  const t = dur >= 1 ? 0.2 : Math.max(0.01, dur > 0 ? dur * 0.05 : 0.01)
  try {
    if (Math.abs(el.currentTime - t) > 0.05) {
      const seeked = waitEvent(el, 'seeked', 1500)
      el.currentTime = t
      await seeked
    }
  } catch {}

  fitChatVideoEl(el)

  try {
    if (el.videoWidth > 0 && el.videoHeight > 0) {
      const canvas = document.createElement('canvas')
      canvas.width = el.videoWidth
      canvas.height = el.videoHeight
      const ctx = canvas.getContext('2d')
      if (ctx) {
        ctx.drawImage(el, 0, 0, canvas.width, canvas.height)
        return canvas.toDataURL('image/jpeg', 0.72)
      }
    }
  } catch {}
  return ''
}
