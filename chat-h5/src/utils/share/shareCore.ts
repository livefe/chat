export async function copyText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch {}
  }
  const ta = document.createElement('textarea')
  ta.value = text
  ta.setAttribute('readonly', '')
  ta.style.position = 'fixed'
  ta.style.left = '-9999px'
  document.body.appendChild(ta)
  ta.select()
  const ok = document.execCommand('copy')
  document.body.removeChild(ta)
  if (!ok) throw new Error('复制失败')
}

export async function shareImageFile(options: {
  file: File
  title: string
  text: string
}): Promise<boolean> {
  const data: ShareData = {
    files: [options.file],
    title: options.title,
    text: options.text,
  }
  if (!navigator.share || typeof navigator.canShare !== 'function' || !navigator.canShare(data)) {
    return false
  }
  try {
    await navigator.share(data)
    return true
  } catch (e: unknown) {
    if (e instanceof Error && e.name === 'AbortError') throw e
    return false
  }
}

export function saveBlobAsFile(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.setTimeout(() => URL.revokeObjectURL(url), 1500)
}
