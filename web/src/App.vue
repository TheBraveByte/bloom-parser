<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import {
  PhFileText, PhGridFour, PhLightning, PhMoon, PhSparkle, PhSun, PhTable,
} from '@phosphor-icons/vue'
import DropZone from './components/DropZone.vue'
import FactsView from './components/FactsView.vue'
import DocView from './components/DocView.vue'
import {
  CSV_HEADER, csvBodyRows, grpcExtract, pool, tableExtractFile,
  type AccResult, type ParsedDocument,
} from './lib/api'
import { theme, toggleTheme } from './lib/theme'

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
const facts = ref<AccResult>()
const matrix = ref<string[][]>([])
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

const totalSize = computed(() => {
  const n = files.value.reduce((s, f) => s + f.size, 0)
  return n > 1e6 ? `${(n / 1e6).toFixed(1)} MB` : `${(n / 1e3).toFixed(0)} KB`
})
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
  error.value = ''; facts.value = undefined; matrix.value = []; doc.value = undefined
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

// runTable processes files with bounded concurrency (one request per file, so
// results stream in, one slow file never blocks the rest, and work spreads
// across CPU cores). OCR dominates runtime, so parallelism is the main lever.
const concurrency = Math.min(navigator.hardwareConcurrency || 4, 8)

async function runTable() {
  const acc: AccResult = { csv: CSV_HEADER + '\r\n', rows: 0, flagged: 0, files: [], matrix: [] }
  facts.value = acc
  await pool(files.value, concurrency, async (f) => {
    try {
      const r = await tableExtractFile(f, refine.value)
      const body = r.csv.split(/\r?\n/).slice(1).filter(Boolean)
      acc.csv += body.length ? body.join('\r\n') + '\r\n' : ''
      acc.rows += r.rows; acc.flagged += r.flagged
      acc.files.push(...r.files)
      // Parse each file's rows once here; the table never re-parses the CSV.
      matrix.value = matrix.value.concat(csvBodyRows(r.csv))
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
  <header class="sticky top-0 z-20 border-b border-line bg-panel/85 backdrop-blur">
    <div class="mx-auto flex max-w-7xl items-center gap-3 px-5 py-3">
      <span class="grid size-8 place-items-center rounded-md bg-accent text-white">
        <PhGridFour :size="18" weight="bold" />
      </span>
      <div class="leading-tight">
        <div class="text-sm font-semibold tracking-tight">bloom parser</div>
        <div class="text-xs text-muted">scanned tables → normalized CSV</div>
      </div>
      <span class="pill ml-auto hidden sm:inline-flex">
        <PhLightning :size="12" /> concurrent extraction
      </span>
      <button
        class="icon-btn" :aria-label="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'"
        @click="toggleTheme"
      >
        <PhSun v-if="theme === 'dark'" :size="17" />
        <PhMoon v-else :size="17" />
      </button>
    </div>
  </header>

  <main class="mx-auto max-w-7xl px-5 pb-16">
    <!-- Hero strip -->
    <section class="pb-6 pt-8">
      <h1 class="text-xl font-semibold tracking-tight sm:text-2xl">
        Extract tables from scanned documents.
      </h1>
      <p class="mt-1.5 max-w-2xl text-sm text-muted">
        Drop statement or balance-sheet scans, pick a pipeline, and get a clean
        normalized CSV. OpenCV finds the grid, Tesseract reads the cells, and an
        optional vision model proofreads the low-confidence ones.
      </p>
      <div class="mt-3 flex flex-wrap gap-2">
        <span class="pill"><b>OpenCV</b> grid detection</span>
        <span class="pill"><b>Tesseract</b> OCR</span>
        <span class="pill"><b>Vision-LLM</b> refine</span>
        <span class="pill"><b>CSV</b> export</span>
      </div>
    </section>

    <!-- Intake -->
    <section class="grid items-start gap-4 lg:grid-cols-[1.1fr_1fr]">
      <div class="card">
        <div class="card-head"><span class="step">1</span>Documents
          <span v-if="files.length" class="ml-auto text-xs font-normal text-muted">
            {{ files.length }} file{{ files.length === 1 ? '' : 's' }} · {{ totalSize }}
          </span>
        </div>
        <div class="card-body">
          <DropZone :files="files" @update="files = $event" />
        </div>
      </div>

      <div class="grid gap-4">
        <div class="card">
          <div class="card-head"><span class="step">2</span>Pipeline</div>
          <div class="card-body space-y-2">
            <label class="opt" :class="pipeline === 'table' ? 'opt-on' : ''">
              <span class="flex items-center gap-2.5">
                <input v-model="pipeline" type="radio" value="table" class="accent-accent">
                <PhTable :size="16" class="text-muted" />
                <b class="text-sm">Table → normalized CSV</b>
                <span class="pill ml-auto">recommended</span>
              </span>
              <span class="mt-1 block pl-6 text-xs text-muted">
                scanned balance sheets · streaming per-file results · downloadable
              </span>
            </label>
            <label class="opt" :class="pipeline === 'document' ? 'opt-on' : ''">
              <span class="flex items-center gap-2.5">
                <input v-model="pipeline" type="radio" value="document" class="accent-accent">
                <PhFileText :size="16" class="text-muted" />
                <b class="text-sm">Raw extract</b>
              </span>
              <span class="mt-1 block pl-6 text-xs text-muted">
                gRPC · plain text &amp; metadata · no CSV
              </span>
            </label>
          </div>
        </div>

        <div class="card">
          <div class="card-head"><span class="step">3</span>Options</div>
          <div class="card-body">
            <template v-if="pipeline === 'table'">
              <label class="flex cursor-pointer items-start gap-2.5 text-sm">
                <input v-model="refine" type="checkbox" class="mt-0.5 accent-accent">
                <span>
                  <span class="flex items-center gap-1.5">
                    <PhSparkle :size="14" class="text-accent" /> Vision-LLM refine
                  </span>
                  <span class="mt-0.5 block text-xs text-muted">
                    Cleans low-confidence cells with a vision model. Much slower —
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
          </div>
        </div>

        <div class="card">
          <div class="card-head"><span class="step">4</span>Run</div>
          <div class="card-body">
            <button
              class="btn-primary w-full"
              :disabled="!files.length || busy"
              @click="run"
            >
              {{ busy ? `Extracting ${done}/${total}…` : files.length ? `Extract ${files.length} ${files.length === 1 ? 'file' : 'files'}` : 'Extract' }}
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
            <p v-else class="mt-2 text-center text-xs text-muted">
              {{ files.length ? estimate || 'ready' : 'add documents to begin' }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Results -->
    <section class="mt-6">
      <div v-if="error" class="card border-err/50 p-4 text-sm text-err">{{ error }}</div>
      <FactsView
        v-else-if="facts && facts.files.length" :result="facts" :matrix="matrix" :busy="busy"
        :single-name="files.length === 1 ? files[0].name : undefined"
      />
      <DocView v-else-if="doc" :doc="doc" />
      <div v-else-if="busy" class="card p-10 text-center text-sm text-muted">Working…</div>
      <div v-else class="card mx-auto max-w-lg p-10 text-center">
        <PhTable :size="28" class="mx-auto text-muted" />
        <p class="mt-3 font-medium">No extraction yet</p>
        <p class="mt-1.5 text-sm text-muted">
          Drop images above, keep <b>Table → normalized CSV</b> selected, press
          <b>Extract</b>, then <b>Download CSV</b>.
        </p>
      </div>
    </section>
  </main>
</template>
