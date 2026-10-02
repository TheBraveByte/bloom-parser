export interface FileStatus {
  file: string
  rows: number
  flagged: number
  error?: string
}

export interface TableExtractResult {
  csv: string
  rows: number
  flagged: number
  files: FileStatus[]
}

export interface DocValue {
  stringValue?: string
  intValue?: string
  doubleValue?: number
  boolValue?: boolean
  datetimeValue?: string
}

export interface DocTable {
  columns: { name: string; type: string }[]
  rows: { values: DocValue[] }[]
}

export interface DocPage {
  number: number
  kind: string
  source?: string
  error?: { code: string; message: string }
  images?: { format: string; width: number; height: number; colorModel: string }[]
  confidence?: number
  text?: string
  tables?: DocTable[]
  warnings?: { code: string; message: string }[]
}

export interface ParsedDocument {
  name: string
  format: string
  pages: DocPage[]
}

function toBase64(buf: ArrayBuffer): string {
  const b = new Uint8Array(buf)
  let s = ''
  const chunk = 0x8000
  for (let i = 0; i < b.length; i += chunk) {
    s += String.fromCharCode.apply(null, Array.from(b.subarray(i, i + chunk)))
  }
  return btoa(s)
}

export async function tableExtract(files: File[], refine: boolean): Promise<TableExtractResult> {
  const fd = new FormData()
  for (const f of files) fd.append('file', f, f.name)
  fd.append('refine', refine ? 'true' : 'false')
  const res = await fetch('/v1/table-extract', { method: 'POST', body: fd })
  if (!res.ok) throw new Error(await res.text() || `HTTP ${res.status}`)
  return res.json()
}

export async function grpcExtract(
  file: File,
  opts: { format: string; ocr: boolean; ocrLanguages: string; maxPages: number },
): Promise<ParsedDocument> {
  const body = {
    document: { name: file.name, content: toBase64(await file.arrayBuffer()), format: opts.format },
    options: {
      ocr: opts.ocr,
      ocr_languages: opts.ocrLanguages,
      max_pages: opts.maxPages || 0,
    },
  }
  const res = await fetch('/v1/documents:extract', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const json = await res.json()
  if (!res.ok) throw new Error(json.message || `HTTP ${res.status}`)
  return json.document
}

/** Minimal RFC-4180 CSV parser: quotes, escaped quotes, CRLF. */
export function parseCSV(text: string): string[][] {
  const rows: string[][] = []
  let row: string[] = []
  let cur = ''
  let q = false
  for (let i = 0; i < text.length; i++) {
    const c = text[i]
    if (q) {
      if (c === '"') {
        if (text[i + 1] === '"') { cur += '"'; i++ } else q = false
      } else cur += c
    } else if (c === '"') q = true
    else if (c === ',') { row.push(cur); cur = '' }
    else if (c === '\n' || c === '\r') {
      if (c === '\r' && text[i + 1] === '\n') i++
      row.push(cur); cur = ''
      if (row.length > 1 || row[0] !== '') rows.push(row)
      row = []
    } else cur += c
  }
  if (cur !== '' || row.length) { row.push(cur); rows.push(row) }
  return rows
}
