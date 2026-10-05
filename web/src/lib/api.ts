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

/** AccResult is the console's accumulator: the merged CSV plus its parsed
 *  matrix, built incrementally as per-file results stream in so the table
 *  never re-parses the whole document on each update. */
export interface AccResult extends TableExtractResult {
  matrix: string[][]
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

/** Native base64 via FileReader: runs in C++, no call-stack chunking, and
 *  avoids the intermediate binary-string the manual loop builds. */
function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const fr = new FileReader()
    fr.onload = () => resolve(String(fr.result).split(',', 2)[1] ?? '')
    fr.onerror = () => reject(fr.error)
    fr.readAsDataURL(file)
  })
}

export async function tableExtract(files: File[], refine: boolean): Promise<TableExtractResult> {
  const fd = new FormData()
  for (const f of files) fd.append('file', f, f.name)
  fd.append('refine', refine ? 'true' : 'false')
  const res = await fetch('/v1/table-extract', { method: 'POST', body: fd })
  if (!res.ok) throw new Error(await res.text() || `HTTP ${res.status}`)
  return res.json()
}

/** Extract one file; used by the console to show live per-file progress. */
export function tableExtractFile(file: File, refine: boolean): Promise<TableExtractResult> {
  return tableExtract([file], refine)
}

export const CSV_HEADER = 'file,row,section,line_item,period,value,raw,flag'
export const CSV_COLS = CSV_HEADER.split(',')

/** Body rows of a per-file CSV (header skipped) — parse once at ingest. */
export function csvBodyRows(csv: string): string[][] {
  return parseCSV(csv).slice(1)
}

/** Run fn over items with bounded concurrency, awaiting all. */
export async function pool<T>(items: T[], limit: number, fn: (item: T) => Promise<void>): Promise<void> {
  const queue = items.slice()
  const workers = Array.from({ length: Math.min(limit, items.length) }, async () => {
    for (let next = queue.shift(); next !== undefined; next = queue.shift()) await fn(next)
  })
  await Promise.all(workers)
}

export async function grpcExtract(
  file: File,
  opts: { format: string; ocr: boolean; ocrLanguages: string; maxPages: number },
): Promise<ParsedDocument> {
  const body = {
    document: { name: file.name, content: await fileToBase64(file), format: opts.format },
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
