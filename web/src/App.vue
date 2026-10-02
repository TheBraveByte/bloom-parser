<script setup lang="ts">
import { ref } from 'vue'
import DropZone from './components/DropZone.vue'
import FactsView from './components/FactsView.vue'
import DocView from './components/DocView.vue'
import { grpcExtract, tableExtract, type ParsedDocument, type TableExtractResult } from './lib/api'

type Pipeline = 'table' | 'document'

const files = ref<File[]>([])
const pipeline = ref<Pipeline>('table')
const refine = ref(false)
const ocr = ref(true)
const langs = ref('eng')
const format = ref('DOCUMENT_FORMAT_UNSPECIFIED')
const maxPages = ref(0)
const busy = ref(false)
const status = ref('')
const error = ref('')
const facts = ref<TableExtractResult>()
const doc = ref<ParsedDocument>()

const formats = [
  ['DOCUMENT_FORMAT_UNSPECIFIED', 'Auto-detect'],
  ['DOCUMENT_FORMAT_PNG', 'PNG'], ['DOCUMENT_FORMAT_JPEG', 'JPEG'],
  ['DOCUMENT_FORMAT_WEBP', 'WEBP'], ['DOCUMENT_FORMAT_TIFF', 'TIFF'],
  ['DOCUMENT_FORMAT_PDF', 'PDF'], ['DOCUMENT_FORMAT_XLSX', 'XLSX'],
  ['DOCUMENT_FORMAT_CSV', 'CSV / delimited'], ['DOCUMENT_FORMAT_JSON', 'JSON'],
  ['DOCUMENT_FORMAT_MARKDOWN_TABLE', 'Markdown table'],
]

async function run() {
  if (!files.value.length || busy.value) return
  busy.value = true
  error.value = ''; facts.value = undefined; doc.value = undefined
  const t0 = performance.now()
  try {
    if (pipeline.value === 'table') {
      status.value = `Extracting ${files.value.length} file(s)…`
      facts.value = await tableExtract(files.value, refine.value)
    } else {
      status.value = 'Extracting…'
      doc.value = await grpcExtract(files.value[0], {
        format: format.value, ocr: ocr.value,
        ocrLanguages: langs.value, maxPages: maxPages.value,
      })
    }
    status.value = `${(performance.now() - t0).toFixed(0)} ms`
  } catch (e) {
    status.value = 'failed'
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <header class="glass flex items-baseline gap-3 border-x-0 border-t-0 px-6 py-4">
    <h1 class="text-lg font-semibold">Bloom Parser</h1>
    <span class="text-sm text-muted">document extraction console</span>
  </header>

  <main class="grid min-h-[calc(100vh-61px)] grid-cols-[360px_1fr]">
    <section class="glass border-y-0 border-l-0 p-6">
      <label class="label">Documents</label>
      <DropZone :files="files" @update="files = $event" />

      <label class="label">Pipeline</label>
      <div class="space-y-1.5">
        <label class="flex cursor-pointer items-center gap-2 text-sm">
          <input v-model="pipeline" type="radio" value="table" class="accent-accent">
          OpenCV table pipeline <span class="text-xs text-muted">scanned balance sheets</span>
        </label>
        <label class="flex cursor-pointer items-center gap-2 text-sm">
          <input v-model="pipeline" type="radio" value="document" class="accent-accent">
          Document extract <span class="text-xs text-muted">gRPC · any format</span>
        </label>
      </div>

      <template v-if="pipeline === 'table'">
        <label class="mt-4 flex cursor-pointer items-center gap-2 text-sm">
          <input v-model="refine" type="checkbox" class="accent-accent">
          Vision-LLM refine <span class="text-xs text-muted">slow, needs API key</span>
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
        {{ busy ? 'Extracting…' : `Extract ${files.length || ''}` }}
      </button>
      <div class="mt-2 break-all font-mono text-xs text-muted">{{ status }}</div>
    </section>

    <section class="overflow-auto p-6">
      <div v-if="error" class="rounded-lg border border-err bg-err/10 p-4 text-err">{{ error }}</div>
      <FactsView
        v-else-if="facts" :result="facts"
        :single-name="files.length === 1 ? files[0].name : undefined"
      />
      <DocView v-else-if="doc" :doc="doc" />
      <div v-else class="mt-10 text-center text-muted">
        Choose files and press <b>Extract</b>.
      </div>
    </section>
  </main>
</template>

<style scoped>
.label {
  display: block;
  margin: 14px 0 4px;
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
