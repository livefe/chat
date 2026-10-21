import QRCode from 'qrcode'

import { resolveMediaUrl } from '@/utils/media'

const CARD_W = 750
const COVER_H = 562
const FOOTER_H = 220
const PAD = 36
const QR_SIZE = 148
const APP_LABEL = '聚点'

function loadImage(src: string): Promise<HTMLImageElement | null> {
  return new Promise((resolve) => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => resolve(img)
    img.onerror = () => resolve(null)
    img.src = src
  })
}

function drawCover(ctx: CanvasRenderingContext2D, img: HTMLImageElement) {
  const scale = Math.max(CARD_W / img.width, COVER_H / img.height)
  const w = img.width * scale
  const h = img.height * scale
  const x = (CARD_W - w) / 2
  const y = (COVER_H - h) / 2
  ctx.drawImage(img, x, y, w, h)
}

function drawTextCover(ctx: CanvasRenderingContext2D, content: string) {
  const gradient = ctx.createLinearGradient(0, 0, CARD_W, COVER_H)
  gradient.addColorStop(0, '#f5f5f5')
  gradient.addColorStop(1, '#e5e5e5')
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, CARD_W, COVER_H)

  const body = content.trim().replace(/\s+/g, ' ')
  if (!body) return

  const maxWidth = CARD_W - PAD * 2
  ctx.fillStyle = '#555555'
  ctx.font = '400 36px system-ui, sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  const lines = wrapLines(ctx, body, maxWidth, 6)
  const lineH = 54
  const blockH = lines.length * lineH
  let y = COVER_H / 2 - blockH / 2 + lineH / 2
  for (const line of lines) {
    ctx.fillText(line, CARD_W / 2, y)
    y += lineH
  }
  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'
}

function wrapLines(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  maxLines: number,
): string[] {
  const chars = Array.from(text)
  const lines: string[] = []
  let line = ''
  for (const ch of chars) {
    const next = line + ch
    if (ctx.measureText(next).width > maxWidth && line) {
      lines.push(line)
      line = ch
      if (lines.length >= maxLines) break
    } else {
      line = next
    }
  }
  if (lines.length < maxLines && line) lines.push(line)
  if (lines.length === maxLines) {
    const last = lines[maxLines - 1]
    if (last && last.length > 1 && ctx.measureText(last).width >= maxWidth - 12) {
      lines[maxLines - 1] = `${last.slice(0, Math.max(1, last.length - 1))}…`
    }
  }
  return lines
}

function wrapTitle(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  maxLines: number,
) {
  return wrapLines(ctx, text.trim() || '动态', maxWidth, maxLines)
}

function canvasToJpegBlob(canvas: HTMLCanvasElement): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) resolve(blob)
        else reject(new Error('分享图生成失败'))
      },
      'image/jpeg',
      0.92,
    )
  })
}

export async function buildPostShareCardBlob(options: {
  title: string
  url: string
  content?: string | null
  coverUrl?: string | null
}): Promise<Blob> {
  const canvas = document.createElement('canvas')
  canvas.width = CARD_W
  canvas.height = COVER_H + FOOTER_H
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('无法创建画布')

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  const coverSrc = resolveMediaUrl(options.coverUrl)
  const cover = coverSrc ? await loadImage(coverSrc) : null
  if (cover) {
    drawCover(ctx, cover)
    const shade = ctx.createLinearGradient(0, COVER_H - 100, 0, COVER_H)
    shade.addColorStop(0, 'rgba(0,0,0,0)')
    shade.addColorStop(1, 'rgba(0,0,0,0.2)')
    ctx.fillStyle = shade
    ctx.fillRect(0, COVER_H - 100, CARD_W, 100)
  } else {
    drawTextCover(ctx, options.content || '')
  }

  const qrDataUrl = await QRCode.toDataURL(options.url, {
    width: QR_SIZE * 2,
    margin: 1,
    errorCorrectionLevel: 'M',
    color: { dark: '#0f0f0f', light: '#ffffff' },
  })
  const qrImg = await loadImage(qrDataUrl)
  if (!qrImg) throw new Error('二维码生成失败')

  const textRight = CARD_W - PAD - QR_SIZE - 28
  const textWidth = Math.max(1, textRight - PAD)
  ctx.fillStyle = '#0f0f0f'
  ctx.font = '700 34px system-ui, sans-serif'
  const lines = wrapTitle(ctx, options.title, textWidth, 2)
  let titleY = COVER_H + 58
  for (const line of lines) {
    ctx.fillText(line, PAD, titleY)
    titleY += 44
  }

  ctx.fillStyle = '#4c8dff'
  ctx.font = '700 24px system-ui, sans-serif'
  const brandY = canvas.height - PAD - 4
  ctx.fillText(APP_LABEL, PAD, brandY)
  const brandW = ctx.measureText(APP_LABEL).width
  ctx.fillStyle = '#888888'
  ctx.font = '400 22px system-ui, sans-serif'
  ctx.fillText('扫码查看详情', PAD + brandW + 12, brandY)

  const qrLeft = CARD_W - PAD - QR_SIZE
  const qrTop = COVER_H + (FOOTER_H - QR_SIZE) / 2
  const framePad = 10
  ctx.fillStyle = '#ffffff'
  roundRect(
    ctx,
    qrLeft - framePad,
    qrTop - framePad,
    QR_SIZE + framePad * 2,
    QR_SIZE + framePad * 2,
    12,
  )
  ctx.fill()
  ctx.strokeStyle = 'rgba(0, 0, 0,0.1)'
  ctx.lineWidth = 1
  roundRect(
    ctx,
    qrLeft - framePad,
    qrTop - framePad,
    QR_SIZE + framePad * 2,
    QR_SIZE + framePad * 2,
    12,
  )
  ctx.stroke()
  ctx.drawImage(qrImg, qrLeft, qrTop, QR_SIZE, QR_SIZE)

  return canvasToJpegBlob(canvas)
}

function roundRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number,
) {
  const radius = Math.min(r, w / 2, h / 2)
  ctx.beginPath()
  ctx.moveTo(x + radius, y)
  ctx.arcTo(x + w, y, x + w, y + h, radius)
  ctx.arcTo(x + w, y + h, x, y + h, radius)
  ctx.arcTo(x, y + h, x, y, radius)
  ctx.arcTo(x, y, x + w, y, radius)
  ctx.closePath()
}
