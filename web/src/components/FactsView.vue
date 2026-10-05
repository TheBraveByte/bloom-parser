<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  PhCheck, PhCopy, PhDownloadSimple, PhMagnifyingGlass,
} from '@phosphor-icons/vue'
import { CSV_COLS, type TableExtractResult } from '../lib/api'

const props = defineProps<{
  result: TableExtractResult
  matrix: string[][]
  singleName?: string
  busy?: boolean
}>()

const flagIdx = CSV_COLS.indexOf('flag')
const valueIdx = CSV_COLS.indexOf('value')
const body = computed(() => props.matrix)
const failed = computed(() => props.result.files.filter(f => f.error))
const clean = computed(() => Math.max(0, props.result.rows - props.result.flagged))
const cleanPct = computed(() => (props.result.rows ? (clean.value / props.result.rows) * 100 : 0))

const query = ref('')
const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return body.value
  return body.value.filter(r => r.some(c => c.toLowerCase().includes(q)))
})

// Windowed rendering: mount a slice and grow it as the sentinel scrolls into
// view, so a 100k-row run doesn't put 100k <tr>s in the DOM at once.
const PAGE = 300
const visible = ref(PAGE)
const shown = computed(() => filtered.value.slice(0, visible.value))
const sentinel = ref<HTMLElement>()
let io: IntersectionObserver | undefined

watch(filtered, () => { visible.value = PAGE })

onMounted(() => {
  io = new IntersectionObserver((entries) => {
    if (entries.some(e => e.isIntersecting)) {
      visible.value = Math.min(visible.value + PAGE, filtered.value.length)
    }
  }, { rootMargin: '600px' })
  if (sentinel.value) io.observe(sentinel.value)
})
onUnmounted(() => io?.disconnect())

type Tile = { label: string; value: number; tone?: 'warn' | 'err' }
const tiles = computed<Tile[]>(() => [
  { label: 'Files', value: props.result.files.length },
  { label: 'Facts', value: props.result.rows },
  { label: 'Flagged', value: props.result.flagged, tone: props.result.flagged ? 'warn' : undefined },
  { label: 'Failed', value: failed.value.length, tone: failed.value.length ? 'err' : undefined },
])

const copied = ref(false)
async function copy() {
  await navigator.clipboard.writeText(props.result.csv)
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}
function download() {
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob([props.result.csv], { type: 'text/csv' }))
  a.download = `${props.singleName?.replace(/\.[^.]*$/, '') ?? 'tables'}.normalized.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}
</script>

<template>
  <div class="grid gap-4">
    <!-- KPI stat tiles -->
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <div v-for="t in tiles" :key="t.label" class="card px-4 py-3.5">
        <div class="text-xl font-semibold tracking-tight" :class="{ 'text-warn': t.tone === 'warn', 'text-err': t.tone === 'err' }">
          {{ t.value.toLocaleString() }}
        </div>
        <div class="mt-0.5 text-xs uppercase tracking-wide text-muted">{{ t.label }}</div>
      </div>
    </div>

    <!-- Quality bar: clean vs flagged, labelled (never colour alone) -->
    <div v-if="result.rows" class="card px-4 py-3.5">
      <div class="flex h-2 overflow-hidden rounded-full bg-well">
        <div class="h-full bg-ok" :style="{ width: cleanPct + '%' }" />
        <div class="ml-0.5 h-full flex-1 bg-warn" />
      </div>
      <div class="mt-2 flex justify-between text-xs text-muted">
        <span><span class="mr-1.5 inline-block size-2 rounded-full bg-ok align-middle" />{{ clean.toLocaleString() }} clean</span>
        <span>{{ result.flagged.toLocaleString() }} flagged<span class="ml-1.5 inline-block size-2 rounded-full bg-warn align-middle" /></span>
      </div>
    </div>

    <!-- Per-file status chips -->
    <div v-if="result.files.length > 1 || failed.length" class="card px-4 py-3">
      <div class="flex flex-wrap gap-1.5">
        <span
          v-for="f in result.files" :key="f.file"
          class="pill font-mono"
          :class="f.error ? 'pill-err' : ''"
          :title="f.error || `${f.rows} facts`"
        >{{ f.file }} <b>{{ f.error ? 'failed' : f.rows }}</b></span>
      </div>
    </div>

    <!-- Data card: toolbar + windowed table -->
    <div class="card overflow-hidden">
      <div class="card-head">
        Extracted facts
        <span class="ml-auto flex items-center gap-2 font-normal">
          <span class="relative">
            <PhMagnifyingGlass :size="14" class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted" />
            <input
              v-model="query" type="search" placeholder="Filter rows…"
              class="w-52 rounded-sm border border-line bg-panel py-1.5 pl-8 pr-2 text-sm focus-visible:outline-2 focus-visible:outline-accent"
            >
          </span>
          <span v-if="query" class="text-xs text-muted">{{ filtered.length.toLocaleString() }} / {{ body.length.toLocaleString() }}</span>
          <button class="btn-ghost py-1.5" :disabled="!body.length" @click="copy">
            <PhCheck v-if="copied" :size="14" class="text-ok" /><PhCopy v-else :size="14" />
            {{ copied ? 'Copied' : 'Copy CSV' }}
          </button>
          <button class="btn-primary py-1.5" :disabled="!body.length" @click="download">
            <PhDownloadSimple :size="14" /> Download CSV
          </button>
          <span v-if="busy" class="text-xs text-accent">streaming…</span>
        </span>
      </div>

      <div v-if="!body.length" class="p-12 text-center text-sm text-muted">No table rows extracted.</div>
      <div v-else class="max-h-[62vh] overflow-auto">
        <table class="w-full border-collapse text-sm">
          <thead class="sticky top-0 z-10">
            <tr>
              <th v-for="h in CSV_COLS" :key="h" class="border-b border-line bg-well px-3 py-2 text-left font-medium text-muted">{{ h }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(r, i) in shown" :key="i"
              class="odd:bg-well/40 hover:bg-well"
              :class="r[flagIdx] ? 'text-warn' : ''"
            >
              <td
                v-for="j in CSV_COLS.length" :key="j"
                class="border-b border-line/60 px-3 py-1.5 font-mono text-xs"
                :class="j - 1 === valueIdx ? 'text-right tabular-nums' : ''"
              >{{ r[j - 1] }}</td>
            </tr>
          </tbody>
        </table>
        <div ref="sentinel" class="h-px" />
        <div v-if="shown.length < filtered.length" class="p-3 text-center text-xs text-muted">
          {{ shown.length.toLocaleString() }} of {{ filtered.length.toLocaleString() }} rows — scroll for more
        </div>
      </div>
    </div>
  </div>
</template>
