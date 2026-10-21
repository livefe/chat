import * as XLSX from 'xlsx'

function triggerDownload(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.rel = 'noopener'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => URL.revokeObjectURL(url), 2000)
}

export function downloadExcelTable(
  filename: string,
  headers: string[],
  rows: string[][],
  options?: { sheetName?: string; minWidth?: number; maxWidth?: number },
) {
  const sheetName = (options?.sheetName || 'Sheet1').slice(0, 31)
  const minWidth = options?.minWidth ?? 10
  const maxWidth = options?.maxWidth ?? 48
  const aoa = [headers, ...rows]
  const ws = XLSX.utils.aoa_to_sheet(aoa)

  ws['!cols'] = headers.map((header, i) => ({
    wch: columnWidthCh(
      header,
      rows.map((r) => r[i] || ''),
      minWidth,
      maxWidth,
    ),
  }))

  const rowCount = aoa.length
  ws['!rows'] = Array.from({ length: rowCount }, (_, i) => ({
    hpt: i === 0 ? 24 : 22,
  }))

  const range = XLSX.utils.decode_range(ws['!ref'] || 'A1')
  for (let R = range.s.r; R <= range.e.r; R += 1) {
    for (let C = range.s.c; C <= range.e.c; C += 1) {
      const addr = XLSX.utils.encode_cell({ r: R, c: C })
      const cell = ws[addr]
      if (!cell) continue
      cell.t = 's'
      cell.v = cell.v == null ? '' : String(cell.v)
      delete cell.w
      delete cell.z
    }
  }

  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, sheetName)

  const buffer = XLSX.write(wb, { bookType: 'xlsx', type: 'array' }) as ArrayBuffer
  const blob = new Blob([buffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
  const name = filename.replace(/\.(csv|xlsx?)$/i, '') + '.xlsx'
  triggerDownload(blob, name)
}

function displayWidth(text: string): number {
  let w = 0
  for (const ch of text) {
    w += (ch.codePointAt(0) || 0) > 0xff ? 2 : 1
  }
  return w
}

function columnWidthCh(
  header: string,
  values: string[],
  minWidth: number,
  maxWidth: number,
): number {
  let maxChars = displayWidth(header)
  for (const v of values) {
    maxChars = Math.max(maxChars, displayWidth(v))
  }
  return Math.min(maxWidth, Math.max(minWidth, maxChars + 2))
}
