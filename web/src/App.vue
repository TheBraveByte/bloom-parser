<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import DropZone from './components/DropZone.vue'
import FactsView from './components/FactsView.vue'
import DocView from './components/DocView.vue'
import {
  CSV_HEADER, grpcExtract, pool, tableExtractFile,
  type ParsedDocument, type TableExtractResult,
} from './lib/api'

type Pipeline = 'table' | 'document'

const files = ref<File[]>([])
const pipeline = ref<Pipeline>('table')
const refine = ref(false)
const ocr = ref(true)
const langs = ref('eng')
const format = ref('DOCUMENT_FORMAT_UNSPECIFIED')
const maxPages = ref(0)

const busy = ref(false)
const error = ref('')
const facts = ref<TableExtractResult>()
const doc = ref<ParsedDocument>()
const done = ref(0)
const total = ref(0)
const elapsed = ref(0)
let timer: number | undefined

const formats = [
  ['DOCUMENT_FORMAT_UNSPECIFIED', 'Auto-detect'],
  ['DOCUMENT_FORMAT_PNG', 'PNG'], ['DOCUMENT_FORMAT_JPEG', 'JPEG'],
  ['DOCUMENT_FORMAT_WEBP', 'WEBP'], ['DOCUMENT_FORMAT_TIFF', 'TIFF'],
  ['DOCUMENT_FORMAT_PDF', 'PDF'], ['DOCUMENT_FORMAT_XLSX', 'XLSX'],
  ['DOCUMENT_FORMAT_CSV', 'CSV / delimited'], ['DOCUMENT_FORMAT_JSON', 'JSON'],
  ['DOCUMENT_FORMAT_MARKDOWN_TABLE', 'Markdown table'],
]

const pct = computed(() => (total.value ? Math.round((done.value / total.value) * 100) : 0))
const secs = computed(() => (elapsed.value / 1000).toFixed(1))
// Rough estimate: refine adds ~1 min/file; plain extract ~2s/file.
const estimate = computed(() => {
  if (pipeline.value !== 'table') return ''
  const per = refine.value ? 60 : 2
  const t = files.value.length * per
  return t >= 60 ? `~${Math.ceil(t / 60)} min` : `~${t}s`
})

onUnmounted(() => clearInterval(timer))

async function run() {
  if (!files.value.length || busy.value) return
  busy.value = true
  error.value = ''; facts.value = undefined; doc.value = undefined
  done.value = 0; total.value = files.value.length; elapsed.value = 0
  const t0 = performance.now()
  timer = window.setInterval(() => { elapsed.value = performance.now() - t0 }, 100)
  try {
    if (pipeline.value === 'table') await runTable()
    else await runDocument()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    clearInterval(timer)
    elapsed.value = performance.now() - t0
    busy.value = false
  }
}

// runTable processes files with bounded concurrency so results stream in and
// one slow file never blocks the rest.
async function runTable() {
  const acc: TableExtractResult = { csv: CSV_HEADER + '\r\n', rows: 0, flagged: 0, files: [] }
  facts.value = acc
  await pool(files.value, 4, async (f) => {
    try {
      const r = await tableExtractFile(f, refine.value)
      const body = r.csv.split(/\r?\n/).slice(1).filter(Boolean)
      acc.csv += body.length ? body.join('\r\n') + '\r\n' : ''
      acc.rows += r.rows; acc.flagged += r.flagged
      acc.files.push(...r.files)
    } catch (e) {
      acc.files.push({ file: f.name, rows: 0, flagged: 0, error: e instanceof Error ? e.message : String(e) })
    }
    done.value++
    facts.value = { ...acc } // trigger reactive re-render as rows arrive
  })
}

async function runDocument() {
  doc.value = await grpcExtract(files.value[0], {
    format: format.value, ocr: ocr.value, ocrLanguages: langs.value, maxPages: maxPages.value,
  })
  done.value = 1
}
</script>

<template>
  <header class="glass flex items-baseline gap-3 border-x-0 border-t-0 px-6 py-4">
    <h1 class="text-lg font-semibold">Bloom Parser</h1>
    <span class="text-sm text-muted">scanned tables → normalized CSV</span>
  </header>

  <main class="grid min-h-[calc(100vh-61px)] grid-cols-[360px_1fr]">
    <section class="glass border-y-0 border-l-0 p-6">
      <label class="label">1 · Documents</label>
      <DropZone :files="files" @update="files = $event" />

      <label class="label">2 · What to produce</label>
      <div class="space-y-2">
        <label class="block cursor-pointer rounded-md border p-2.5 text-sm"
          :class="pipeline === 'table' ? 'border-accent bg-accent/10' : 'border-line'">
          <span class="flex items-center gap-2">
            <input v-model="pipeline" type="radio" value="table" class="accent-accent">
            <b>Table → normalized CSV</b>
          </span>
          <span class="mt-0.5 block pl-6 text-xs text-muted">scanned balance sheets · downloadable · recommended</span>
        </label>
        <label class="block cursor-pointer rounded-md border p-2.5 text-sm"
          :class="pipeline === 'document' ? 'border-accent bg-accent/10' : 'border-line'">
          <span class="flex items-center gap-2">
            <input v-model="pipeline" type="radio" value="document" class="accent-accent">
            <b>Raw extract</b>
          </span>
          <span class="mt-0.5 block pl-6 text-xs text-muted">gRPC · plain text &amp; metadata · no CSV</span>
        </label>
      </div>

      <template v-if="pipeline === 'table'">
        <label class="label">3 · Options</label>
        <label class="flex cursor-pointer items-start gap-2 text-sm">
          <input v-model="refine" type="checkbox" class="mt-1 accent-accent">
          <span>
            Vision-LLM refine
            <span class="block text-xs text-muted">
              Cleans low-confidence cells with a free vision model. Much slower —
              use it for a few files, not a large batch.
            </span>
          </span>
        </label>
      </template>
      <template v-else>
        <label class="label">Format</label>
        <select v-model="format" class="field">
          <option v-for="[v, l] in formats" :key="v" :value="v">{{ l }}</option>
        </select>
        <div class="mt-3 grid grid-cols-2 gap-3">
          <div>
            <label class="label">OCR languages</label>
            <input v-model="langs" type="text" class="field">
          </div>
          <div>
            <label class="label">Max pages</label>
            <input v-model.number="maxPages" type="number" min="0" class="field">
          </div>
        </div>
        <label class="mt-4 flex cursor-pointer items-center gap-2 text-sm">
          <input v-model="ocr" type="checkbox" class="accent-accent">
          Attempt OCR on image pages
        </label>
        <p v-if="files.length > 1" class="mt-2 text-xs text-warn">
          Raw extract reads only the first file.
        </p>
      </template>

      <button
        class="mt-5 w-full rounded-md bg-accent py-2.5 font-semibold text-white transition duration-200
               shadow-[inset_0_1px_0_rgb(255_255_255/0.15),0_4px_14px_rgb(79_140_255/0.25)]
               hover:-translate-y-px hover:brightness-110 active:translate-y-0 active:brightness-90
               focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent
               disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
        :disabled="!files.length || busy"
        @click="run"
      >
        {{ busy ? `Extracting ${done}/${total}…` : `Extract ${files.length || ''}` }}
      </button>

      <div v-if="busy || elapsed" class="mt-3">
        <div v-if="total > 1" class="h-1.5 overflow-hidden rounded-full bg-well">
          <div class="h-full rounded-full bg-accent transition-[width] duration-200" :style="{ width: pct + '%' }" />
        </div>
        <div class="mt-1.5 flex justify-between font-mono text-xs text-muted">
          <span>{{ busy ? `${done}/${total} · ${secs}s` : `done in ${secs}s` }}</span>
          <span v-if="busy && estimate">{{ estimate }}</span>
        </div>
      </div>
    </section>

    <section class="overflow-auto p-6">
      <div v-if="error" class="rounded-lg border border-err bg-err/10 p-4 text-err">{{ error }}</div>
      <FactsView
        v-else-if="facts && facts.files.length" :result="facts" :busy="busy"
        :single-name="files.length === 1 ? files[0].name : undefined"
      />
      <DocView v-else-if="doc" :doc="doc" />
      <div v-else-if="busy" class="mt-10 text-center text-muted">Working…</div>
      <div v-else class="mx-auto mt-16 max-w-sm text-center text-muted">
        <p class="text-fg">Extract tables from scanned documents.</p>
        <p class="mt-2 text-sm">
          Drop images on the left, keep <b>Table → normalized CSV</b> selected, press
          <b>Extract</b>, then <b>Download CSV</b>.
        </p>
      </div>
    </section>
  </main>
</template>

<style scoped>
.label {
  display: block;
  margin: 16px 0 6px;
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-muted);
}
.field {
  width: 100%;
  padding: 8px 10px;
  background: var(--color-panel);
  border: 1px solid var(--color-line);
  border-radius: var(--radius-md);
  color: var(--color-fg);
}
.field:focus-visible { outline: 2px solid var(--color-accent); }
</style>
