import { onBeforeUnmount, ref } from 'vue'
import { showToast } from 'vant'

import {
  pickAudioMimeType,
  VOICE_MAX_DURATION_MS,
  VOICE_MIN_DURATION_MS,
} from '@/utils/chat/chatVoice'

export function useVoiceRecorder(opts: {
  canStart: () => boolean
  onStart: () => void
  onRecorded: (blob: Blob, mime: string, durationMs: number) => void | Promise<void>
}) {
  const recording = ref(false)
  const recordCanceling = ref(false)

  let mediaStream: MediaStream | null = null
  let mediaRecorder: MediaRecorder | null = null
  let recordChunks: Blob[] = []
  let recordStartedAt = 0
  let recordMime = ''
  let maxTimer: ReturnType<typeof setTimeout> | null = null
  let recordPointerId: number | null = null

  function clearMaxTimer() {
    if (maxTimer) {
      clearTimeout(maxTimer)
      maxTimer = null
    }
  }

  function stopMediaTracks() {
    mediaStream?.getTracks().forEach((t) => t.stop())
    mediaStream = null
  }

  function cancelRecording(silent = false) {
    clearMaxTimer()
    const recorder = mediaRecorder
    mediaRecorder = null
    recordChunks = []
    recording.value = false
    recordCanceling.value = false
    recordPointerId = null
    if (recorder && recorder.state !== 'inactive') {
      recorder.onstop = null
      try {
        recorder.stop()
      } catch {}
    }
    stopMediaTracks()
    if (!silent) showToast('已取消')
  }

  async function finishRecording() {
    clearMaxTimer()
    const recorder = mediaRecorder
    const startedAt = recordStartedAt
    const mime = recordMime
    const chunks = recordChunks
    mediaRecorder = null
    recording.value = false
    recordCanceling.value = false
    recordPointerId = null

    if (!recorder) {
      stopMediaTracks()
      return
    }

    const blob = await new Promise<Blob | null>((resolve) => {
      const done = () => {
        stopMediaTracks()
        if (!chunks.length) {
          resolve(null)
          return
        }
        resolve(new Blob(chunks, { type: mime || 'audio/webm' }))
      }
      if (recorder.state === 'inactive') {
        done()
        return
      }
      recorder.onstop = done
      try {
        recorder.stop()
      } catch {
        done()
      }
    })

    if (!blob || blob.size <= 0) {
      showToast('录音失败')
      return
    }
    const durationMs = Date.now() - startedAt
    if (durationMs < VOICE_MIN_DURATION_MS) {
      showToast('说话时间太短')
      return
    }
    await opts.onRecorded(blob, mime, durationMs)
  }

  async function startRecording(ev: PointerEvent) {
    if (recording.value || !opts.canStart()) return
    if (typeof MediaRecorder === 'undefined' || !navigator.mediaDevices?.getUserMedia) {
      showToast('当前环境不支持录音')
      return
    }
    const mime = pickAudioMimeType()
    if (!mime) {
      showToast('当前环境不支持录音')
      return
    }

    opts.onStart()

    try {
      mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true })
    } catch {
      showToast('请允许使用麦克风')
      return
    }

    try {
      recordChunks = []
      recordMime = mime
      mediaRecorder = new MediaRecorder(mediaStream, { mimeType: mime })
      mediaRecorder.ondataavailable = (e) => {
        if (e.data && e.data.size > 0) recordChunks.push(e.data)
      }
      mediaRecorder.start(200)
      recordStartedAt = Date.now()
      recording.value = true
      recordCanceling.value = false
      recordPointerId = ev.pointerId
      ;(ev.currentTarget as HTMLElement | null)?.setPointerCapture?.(ev.pointerId)
      clearMaxTimer()
      maxTimer = setTimeout(() => {
        void finishRecording()
      }, VOICE_MAX_DURATION_MS)
    } catch {
      stopMediaTracks()
      mediaRecorder = null
      recording.value = false
      showToast('录音失败')
    }
  }

  function onHoldPointerMove(ev: PointerEvent) {
    if (!recording.value || recordPointerId !== ev.pointerId) return
    const rect = (ev.currentTarget as HTMLElement).getBoundingClientRect()
    recordCanceling.value = ev.clientY < rect.top - 24
  }

  function onHoldPointerUp(ev: PointerEvent) {
    if (!recording.value || (recordPointerId != null && recordPointerId !== ev.pointerId)) return
    if (recordCanceling.value) cancelRecording()
    else void finishRecording()
  }

  function onHoldPointerCancel(ev: PointerEvent) {
    if (!recording.value || (recordPointerId != null && recordPointerId !== ev.pointerId)) return
    cancelRecording()
  }

  onBeforeUnmount(() => {
    cancelRecording(true)
  })

  return {
    recording,
    recordCanceling,
    startRecording,
    cancelRecording,
    onHoldPointerMove,
    onHoldPointerUp,
    onHoldPointerCancel,
  }
}
